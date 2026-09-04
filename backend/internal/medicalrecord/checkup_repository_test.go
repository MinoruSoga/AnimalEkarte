package medicalrecord

// repository_test.go — Repository の統合テスト（カバレッジ向上）。
//
// 移動元: checkup_repository_test.go（BE8-4 batch24）。
//
// 対象: FindByClinicID / FindByMedicalRecordID / FindByOwnerID / FindByID /
//       Create / Update / Delete
// 検証観点: 正常系、フィルタ、clinic_id 隔離、ソフトデリート除外、NotFound ラップ。
//
// CheckupType Preload の clinic 隔離は repository/master_preload_clinic_isolation_test.go の
// TestCheckupRepository_FindByID_CheckupTypePreloadClinicIsolation で別途検証済みのため、
// 本ファイルでは重複させない（同テストはフラット package の facade 経由で動作確認する）。
//
// makeSpeciesAndPet/makeHistoryMedicalRecord は medicalrecord テストパッケージ共有の
// ローカルコピー（diagnosis_name_repository_test.go 定義）をそのまま使う。makeCheckupTypeMaster
// は checkup_field 系テストも参照する共有ヘルパーとして本ファイルに 1 箇所だけ定義する。
// makeTestOwner はフラット側と同様 testdb.MakeTestOwner に委譲する（prescription 側の
// makeTestOwner ラッパー経由でも解決可能）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupCheckupRepoTestDB は checkups / checkup_types / medical_records 周りを整備する。
func setupCheckupRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.MedicalRecord{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.CheckupType{},
		&model.Checkup{},
	))
	db.Exec("TRUNCATE TABLE checkups CASCADE")
	db.Exec("TRUNCATE TABLE checkup_types CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	return db
}

func makeCheckupTypeMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.CheckupType {
	t.Helper()
	ct := &model.CheckupType{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(ct).Error)
	return ct
}

// makeCheckupWithDates は date / next_date を指定して Checkup を作成する。
func makeCheckupWithDates(t *testing.T, db *gorm.DB, clinicID, mrID, petID, checkupTypeID uint64, date time.Time, nextDate *time.Time) *model.Checkup {
	t.Helper()
	pid := petID
	c := &model.Checkup{
		ClinicID:        clinicID,
		MedicalRecordID: mrID,
		PetID:           &pid,
		CheckupTypeID:   checkupTypeID,
		Date:            date,
		NextDate:        nextDate,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c
}

func TestCheckupRepository_LockByIDForUpdateRequiresAmbientTransaction(t *testing.T) {
	repo := NewCheckupRepository(nil)

	got, err := repo.LockByIDForUpdate(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestCheckupRepository_FindByClinicID_FiltersAndClinicIsolation(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "健診一覧飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "健診一覧ポチA")
	mrA := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-LIST-A", time.Now())
	ctA := makeCheckupTypeMaster(t, db, clinicA, "医院Aの健診")

	ownerB := testdb.MakeTestOwner(t, db, clinicB, "健診一覧飼主B")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "健診一覧ポチB")
	mrB := makeHistoryMedicalRecord(t, db, clinicB, petB.ID, "MR-LIST-B", time.Now())
	ctB := makeCheckupTypeMaster(t, db, clinicB, "医院Bの健診")

	d1 := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	d3 := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	n1 := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)

	old := makeCheckupWithDates(t, db, clinicA, mrA.ID, petA.ID, ctA.ID, d1, nil)
	mid := makeCheckupWithDates(t, db, clinicA, mrA.ID, petA.ID, ctA.ID, d2, &n1)
	recent := makeCheckupWithDates(t, db, clinicA, mrA.ID, petA.ID, ctA.ID, d3, nil)
	_ = makeCheckupWithDates(t, db, clinicB, mrB.ID, petB.ID, ctB.ID, d2, nil) // 別クリニック

	t.Run("フィルタ無しで clinic A の健診が date DESC で返り total は全件数", func(t *testing.T) {
		got, total, err := repo.FindByClinicID(ctx, clinicA, CheckupFilters{}, 1, 20)
		require.NoError(t, err)
		require.Len(t, got, 3, "clinic A の健診のみ返る")
		assert.EqualValues(t, 3, total)
		assert.Equal(t, recent.ID, got[0].ID)
		assert.Equal(t, mid.ID, got[1].ID)
		assert.Equal(t, old.ID, got[2].ID)
		for _, c := range got {
			assert.Equal(t, clinicA, c.ClinicID)
		}
	})

	t.Run("StartDate/EndDate で絞り込み", func(t *testing.T) {
		start := "2026-02-01"
		end := "2026-02-28"
		got, total, err := repo.FindByClinicID(ctx, clinicA, CheckupFilters{StartDate: &start, EndDate: &end}, 1, 20)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.EqualValues(t, 1, total)
		assert.Equal(t, mid.ID, got[0].ID)
	})

	t.Run("NextStartDate/NextEndDate で絞り込み", func(t *testing.T) {
		nStart := "2026-04-01"
		nEnd := "2026-04-30"
		got, total, err := repo.FindByClinicID(ctx, clinicA, CheckupFilters{NextStartDate: &nStart, NextEndDate: &nEnd}, 1, 20)
		require.NoError(t, err)
		require.Len(t, got, 1, "next_date を持つ健診のみヒットする")
		assert.EqualValues(t, 1, total)
		assert.Equal(t, mid.ID, got[0].ID)
	})

	t.Run("page/limit で結果件数を絞り込みつつ total はフィルタ全件を維持", func(t *testing.T) {
		got, total, err := repo.FindByClinicID(ctx, clinicA, CheckupFilters{}, 1, 2)
		require.NoError(t, err)
		require.Len(t, got, 2, "limit=2 で 1 ページ目は 2 件のみ")
		assert.EqualValues(t, 3, total, "total は limit に関わらず全件数")
		assert.Equal(t, recent.ID, got[0].ID)
		assert.Equal(t, mid.ID, got[1].ID)

		got2, total2, err := repo.FindByClinicID(ctx, clinicA, CheckupFilters{}, 2, 2)
		require.NoError(t, err)
		require.Len(t, got2, 1, "2 ページ目は残り 1 件")
		assert.EqualValues(t, 3, total2)
		assert.Equal(t, old.ID, got2[0].ID)
	})

	t.Run("PetID で page 選択前に絞り込む", func(t *testing.T) {
		otherPet := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "健診一覧ポチA2")
		otherMR := makeHistoryMedicalRecord(t, db, clinicA, otherPet.ID, "MR-LIST-A2", time.Now())
		_ = makeCheckupWithDates(t, db, clinicA, otherMR.ID, otherPet.ID, ctA.ID, d3, nil)
		got, total, err := repo.FindByClinicID(ctx, clinicA, CheckupFilters{PetID: &petA.ID}, 1, 2)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.EqualValues(t, 3, total)
		for _, checkup := range got {
			require.NotNil(t, checkup.PetID)
			assert.Equal(t, petA.ID, *checkup.PetID)
		}
	})
}

func TestCheckupRepository_PatientRelationsAreClinicScoped(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "健診関係スコープ飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "健診関係スコープペットA")
	recordA := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-CHECKUP-SCOPE-A", time.Now())
	require.NoError(t, db.Model(recordA).Update("owner_id", ownerA.ID).Error)
	recordA.OwnerID = &ownerA.ID
	typeA := makeCheckupTypeMaster(t, db, clinicA, "健診関係スコープ種別A")
	valid := makeCheckupWithDates(t, db, clinicA, recordA.ID, petA.ID, typeA.ID, time.Now(), nil)

	ensureVaccinationTestClinics(t, db, clinicA, clinicB)
	validDoctor := makeDoctor(t, db, clinicB, "健診関係スコープ有効医師")
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: validDoctor.ID, ClinicID: clinicA,
	}).Error)
	require.NoError(t, db.Model(valid).Update("doctor_id", validDoctor.ID).Error)
	valid.DoctorID = &validDoctor.ID

	unassignedDoctor := makeDoctor(t, db, clinicA, "健診関係スコープ未所属医師")
	inactiveDoctor := makeDoctor(t, db, clinicA, "健診関係スコープ無効医師")
	require.NoError(t, db.Model(inactiveDoctor).UpdateColumn("is_active", false).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: inactiveDoctor.ID, ClinicID: clinicA,
	}).Error)
	nurse := makeDoctor(t, db, clinicA, "健診関係スコープ看護師")
	require.NoError(t, db.Model(nurse).UpdateColumn("staff_type", model.StaffTypeNurse).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: nurse.ID, ClinicID: clinicA,
	}).Error)

	ownerB := testdb.MakeTestOwner(t, db, clinicB, "健診関係スコープ飼主B")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "健診関係スコープペットB")
	recordB := makeHistoryMedicalRecord(t, db, clinicB, petB.ID, "MR-CHECKUP-SCOPE-B", time.Now())
	require.NoError(t, db.Model(recordB).Update("owner_id", ownerB.ID).Error)
	recordB.OwnerID = &ownerB.ID
	typeB := makeCheckupTypeMaster(t, db, clinicB, "健診関係スコープ種別B")

	pollutedOwnerPet := makeSpeciesAndPet(
		t,
		db,
		clinicA,
		ownerB.ID,
		"健診関係スコープ別院飼主ペット",
	)
	polluted := []*model.Checkup{
		makeCheckupWithDates(t, db, clinicA, recordB.ID, petA.ID, typeA.ID, time.Now(), nil),
		makeCheckupWithDates(t, db, clinicA, recordA.ID, petB.ID, typeA.ID, time.Now(), nil),
		makeCheckupWithDates(t, db, clinicA, recordA.ID, pollutedOwnerPet.ID, typeA.ID, time.Now(), nil),
		makeCheckupWithDates(t, db, clinicA, recordA.ID, petA.ID, typeB.ID, time.Now(), nil),
		makeCheckupWithDates(t, db, clinicB, recordA.ID, petA.ID, typeA.ID, time.Now(), nil),
	}
	for _, doctorID := range []uint64{unassignedDoctor.ID, inactiveDoctor.ID, nurse.ID} {
		item := makeCheckupWithDates(t, db, clinicA, recordA.ID, petA.ID, typeA.ID, time.Now(), nil)
		require.NoError(t, db.Model(item).Update("doctor_id", doctorID).Error)
		polluted = append(polluted, item)
	}

	for _, item := range polluted {
		got, err := repo.FindByID(ctx, clinicA, item.ID)
		require.Error(t, err, "polluted checkup %d must fail closed", item.ID)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, got)
	}

	listed, total, err := repo.FindByClinicID(ctx, clinicA, CheckupFilters{}, 1, 100)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, listed, 1)
	assert.Equal(t, valid.ID, listed[0].ID)
	require.NotNil(t, listed[0].MedicalRecord)
	require.NotNil(t, listed[0].MedicalRecord.Pet)
	require.NotNil(t, listed[0].MedicalRecord.Pet.Owner)
	assert.Equal(t, ownerA.ID, listed[0].MedicalRecord.Pet.Owner.ID)
	require.NotNil(t, listed[0].Doctor)
	assert.Equal(t, validDoctor.ID, listed[0].Doctor.ID)

	byRecord, err := repo.FindByMedicalRecordID(ctx, clinicA, recordA.ID)
	require.NoError(t, err)
	require.Len(t, byRecord, 1)
	assert.Equal(t, valid.ID, byRecord[0].ID)

	byOwner, err := repo.FindByOwnerID(ctx, clinicA, ownerA.ID)
	require.NoError(t, err)
	require.Len(t, byOwner, 1)
	assert.Equal(t, valid.ID, byOwner[0].ID)
}

func TestCheckupRepository_FindByMedicalRecordID(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicA, "健診MR飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "健診MRポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-DETAIL", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "健診種別")

	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	first := makeCheckupWithDates(t, db, clinicA, mr.ID, pet.ID, ct.ID, d1, nil)
	second := makeCheckupWithDates(t, db, clinicA, mr.ID, pet.ID, ct.ID, d2, nil)

	got, err := repo.FindByMedicalRecordID(ctx, clinicA, mr.ID)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, first.ID, got[0].ID, "date ASC 昇順で先頭に来る")
	assert.Equal(t, second.ID, got[1].ID)

	t.Run("存在しない medical_record_id は空", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, 999999)
		require.NoError(t, err)
		assert.Len(t, got, 0)
	})
}

// FindByOwnerID は ISSUE-004 タグ再同期用: medical_records.pet_id から現在の pets.owner_id を解決する。
func TestCheckupRepository_FindByOwnerID(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicA, "健診飼主同期")
	otherOwner := testdb.MakeTestOwner(t, db, clinicA, "別の飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "健診同期ポチ")
	ct := makeCheckupTypeMaster(t, db, clinicA, "同期健診種別")

	mr := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "MR-OWNERSYNC", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID, Status: model.MedicalRecordStatusFinalized}
	require.NoError(t, db.WithContext(ctx).Create(mr).Error)
	otherMR := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "MR-OTHEROWNER", Date: time.Now(), OwnerID: &otherOwner.ID, PetID: &pet.ID, Status: model.MedicalRecordStatusFinalized}
	require.NoError(t, db.WithContext(ctx).Create(otherMR).Error)

	own := makeCheckupWithDates(t, db, clinicA, mr.ID, pet.ID, ct.ID, time.Now(), nil)
	other := makeCheckupWithDates(t, db, clinicA, otherMR.ID, pet.ID, ct.ID, time.Now(), nil)

	got, err := repo.FindByOwnerID(ctx, clinicA, owner.ID)
	require.NoError(t, err)
	require.Len(t, got, 2, "現在飼主のペットに紐づく健診は snapshot owner に関係なく返る")
	assert.ElementsMatch(t, []uint64{own.ID, other.ID}, []uint64{got[0].ID, got[1].ID})
	otherOwnerRows, err := repo.FindByOwnerID(ctx, clinicA, otherOwner.ID)
	require.NoError(t, err)
	assert.Empty(t, otherOwnerRows)
}

func TestCheckupRepository_FindByOwnerID_CurrentOwnerAfterTransfer(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicID = uint64(70102)

	fixture := makeCurrentOwnerTransferFixture(
		t,
		db,
		clinicID,
		"MR-CHECKUP-CURRENT-OWNER",
		time.Now(),
	)
	checkupType := makeCheckupTypeMaster(t, db, clinicID, "譲渡健診")
	checkup := makeCheckupWithDates(t, db, clinicID, fixture.Record.ID, fixture.Pet.ID, checkupType.ID, time.Now(), nil)

	got, err := repo.FindByOwnerID(ctx, clinicID, fixture.CurrentOwner.ID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, checkup.ID, got[0].ID)
	var storedRecord model.MedicalRecord
	require.NoError(t, db.WithContext(ctx).First(&storedRecord, fixture.Record.ID).Error)
	require.NotNil(t, storedRecord.OwnerID)
	assert.Equal(t, fixture.PreviousOwner.ID, *storedRecord.OwnerID, "medical record keeps the historical owner snapshot")

	previous, err := repo.FindByOwnerID(ctx, clinicID, fixture.PreviousOwner.ID)
	require.NoError(t, err)
	assert.Empty(t, previous)
}

// TestCheckupRepository_FindByOwnerIDs は G2F-02 page bulk: multi-owner index、clinic 隔離、空入力を検証する。
func TestCheckupRepository_FindByOwnerIDs(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db).(*checkupRepository)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner1 := testdb.MakeTestOwner(t, db, clinicA, "bulk飼主1")
	owner2 := testdb.MakeTestOwner(t, db, clinicA, "bulk飼主2")
	otherClinicOwner := testdb.MakeTestOwner(t, db, clinicB, "他院飼主")
	pet1 := makeSpeciesAndPet(t, db, clinicA, owner1.ID, "bulkポチ1")
	pet2 := makeSpeciesAndPet(t, db, clinicA, owner2.ID, "bulkポチ2")
	petB := makeSpeciesAndPet(t, db, clinicB, otherClinicOwner.ID, "他院ポチ")
	ctA := makeCheckupTypeMaster(t, db, clinicA, "bulk健診")
	ctB := makeCheckupTypeMaster(t, db, clinicB, "他院健診")

	mr1 := makeHistoryMedicalRecord(t, db, clinicA, pet1.ID, "MR-BULK-1", time.Now())
	mr2 := makeHistoryMedicalRecord(t, db, clinicA, pet2.ID, "MR-BULK-2", time.Now())
	mrB := makeHistoryMedicalRecord(t, db, clinicB, petB.ID, "MR-BULK-B", time.Now())

	c1 := makeCheckupWithDates(t, db, clinicA, mr1.ID, pet1.ID, ctA.ID, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), nil)
	c2 := makeCheckupWithDates(t, db, clinicA, mr2.ID, pet2.ID, ctA.ID, time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), nil)
	_ = makeCheckupWithDates(t, db, clinicB, mrB.ID, petB.ID, ctB.ID, time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC), nil)

	t.Run("indexes by owner and excludes other clinic", func(t *testing.T) {
		got, err := repo.FindByOwnerIDs(ctx, clinicA, []uint64{owner1.ID, owner2.ID, otherClinicOwner.ID})
		require.NoError(t, err)
		require.Len(t, got[owner1.ID], 1)
		assert.Equal(t, c1.ID, got[owner1.ID][0].ID)
		require.NotNil(t, got[owner1.ID][0].CheckupType)
		require.Len(t, got[owner2.ID], 1)
		assert.Equal(t, c2.ID, got[owner2.ID][0].ID)
		assert.Empty(t, got[otherClinicOwner.ID], "cross-clinic owner must not leak into clinicA bulk")
	})

	t.Run("empty ownerIDs returns empty map", func(t *testing.T) {
		got, err := repo.FindByOwnerIDs(ctx, clinicA, nil)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("history cap is applied per owner", func(t *testing.T) {
		// healthTagOwnerHistoryMax+1 rows for owner1; only the newest max rows return.
		for i := 0; i < healthTagOwnerHistoryMax+1; i++ {
			d := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
			makeCheckupWithDates(t, db, clinicA, mr1.ID, pet1.ID, ctA.ID, d, nil)
		}
		got, err := repo.FindByOwnerIDs(ctx, clinicA, []uint64{owner1.ID})
		require.NoError(t, err)
		assert.LessOrEqual(t, len(got[owner1.ID]), healthTagOwnerHistoryMax)
	})
}

// TestMedicalRecordRepository_FindOwnerVisitSummariesByOwnerIDs は page bulk visit-summary を検証する。
func TestMedicalRecordRepository_FindOwnerVisitSummariesByOwnerIDs(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewMedicalRecordRepository(db).(*medicalRecordRepository)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner1 := testdb.MakeTestOwner(t, db, clinicA, "visit bulk1")
	owner2 := testdb.MakeTestOwner(t, db, clinicA, "visit bulk2")
	other := testdb.MakeTestOwner(t, db, clinicB, "visit other clinic")
	pet1 := makeSpeciesAndPet(t, db, clinicA, owner1.ID, "visit pet1")
	pet2 := makeSpeciesAndPet(t, db, clinicA, owner2.ID, "visit pet2")
	petB := makeSpeciesAndPet(t, db, clinicB, other.ID, "visit petB")

	// owner1: 2 records in last year → AnnualCount 2
	makeHistoryMedicalRecord(t, db, clinicA, pet1.ID, "MR-VS-1A", time.Now().AddDate(0, -1, 0))
	makeHistoryMedicalRecord(t, db, clinicA, pet1.ID, "MR-VS-1B", time.Now().AddDate(0, -2, 0))
	// owner2: none
	// other clinic: should not appear under clinicA
	makeHistoryMedicalRecord(t, db, clinicB, petB.ID, "MR-VS-B", time.Now())

	got, err := repo.FindOwnerVisitSummariesByOwnerIDs(ctx, clinicA, []uint64{owner1.ID, owner2.ID, other.ID})
	require.NoError(t, err)
	require.NotNil(t, got[owner1.ID])
	assert.Equal(t, int64(2), got[owner1.ID].AnnualCount)
	assert.Equal(t, int64(2), got[owner1.ID].TotalCount)
	assert.Nil(t, got[owner2.ID], "owners with no visits are absent (caller treats as zero)")
	assert.Nil(t, got[other.ID], "cross-clinic owner must not appear")

	empty, err := repo.FindOwnerVisitSummariesByOwnerIDs(ctx, clinicA, nil)
	require.NoError(t, err)
	assert.Empty(t, empty)
	_ = pet2
}

func TestCheckupRepository_FindByID(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := testdb.MakeTestOwner(t, db, clinicA, "健診単体飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "健診単体ポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-SINGLE", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "単体健診種別")
	c := makeCheckupWithDates(t, db, clinicA, mr.ID, pet.ID, ct.ID, time.Now(), nil)

	t.Run("同一クリニックで取得しMedicalRecordがPreloadされる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		require.NotNil(t, got.MedicalRecord)
		assert.Equal(t, mr.ID, got.MedicalRecord.ID)
	})

	t.Run("別クリニックからは NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, c.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestCheckupRepository_Create(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicA, "新規健診飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "新規健診ポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-NEW", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "新規健診種別")

	c := &model.Checkup{ClinicID: clinicA, MedicalRecordID: mr.ID, PetID: &pet.ID, CheckupTypeID: ct.ID, Date: time.Now()}
	require.NoError(t, repo.Create(ctx, c))
	assert.NotZero(t, c.ID)

	got, err := repo.FindByID(ctx, clinicA, c.ID)
	require.NoError(t, err)
	assert.Equal(t, c.ID, got.ID)
}

func TestCheckupRepository_Update(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := testdb.MakeTestOwner(t, db, clinicA, "更新健診飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "更新健診ポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-UPD", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "更新健診種別")
	c := makeCheckupWithDates(t, db, clinicA, mr.ID, pet.ID, ct.ID, time.Now(), nil)

	t.Run("同一クリニックで更新できる", func(t *testing.T) {
		require.NoError(t, repo.Update(ctx, clinicA, c.ID, map[string]any{"result": "異常なし"}))
		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "異常なし", got.Result)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicB, c.ID, map[string]any{"result": "乗っ取り"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の更新は NotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, 999999, map[string]any{"result": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestCheckupRepository_Delete(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := testdb.MakeTestOwner(t, db, clinicA, "削除健診飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "削除健診ポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-DEL", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "削除健診種別")
	c := makeCheckupWithDates(t, db, clinicA, mr.ID, pet.ID, ct.ID, time.Now(), nil)

	t.Run("別クリニックからの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, c.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("同一クリニックで削除でき、ソフトデリートされる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, c.ID))

		_, err := repo.FindByID(ctx, clinicA, c.ID)
		assert.True(t, apperrors.IsNotFound(err))

		all, _, err := repo.FindByClinicID(ctx, clinicA, CheckupFilters{}, 1, 20)
		require.NoError(t, err)
		for _, x := range all {
			assert.NotEqual(t, c.ID, x.ID)
		}

		var raw model.Checkup
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", c.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "deleted_at が設定されているべき")
	})
}
