package medicalrecord

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// makeMedicalRecordForPet はテスト用カルテ 1 件を作成する。deleted=true なら論理削除する。
// pet_id は medical_records → pets の FK 制約があるため、呼び出し側で実在ペットの ID を渡すこと。
// 既存ヘルパー（makeTestOwner / makeSpeciesAndPet）に合わせ context は内部生成する。
func makeMedicalRecordForPet(t *testing.T, db *gorm.DB, clinicID, petID uint64, date time.Time, deleted bool) {
	t.Helper()
	ctx := context.Background()
	pid := petID
	mr := &model.MedicalRecord{ClinicID: clinicID, PetID: &pid, Date: date}
	require.NoError(t, db.WithContext(ctx).Create(mr).Error)
	if deleted {
		// gorm.DeletedAt により deleted_at がセットされる（論理削除）。
		require.NoError(t, db.WithContext(ctx).Delete(mr).Error)
	}
}

// TestMedicalRecordRepository_FindFirstVisitDateByPetID は #158 飼主レポートの初診日導出を検証する。
// 期待: 同一ペット・同一医院の有効カルテのうち最古の date を返し、
// 論理削除カルテ・別医院カルテ・別ペットのカルテは集計から除外する。捏造はしない（無ければ nil）。
func TestMedicalRecordRepository_FindFirstVisitDateByPetID(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}))
	// 共有テスト DB の他テスト残骸と混ざらないよう、FK 連鎖ごと初期化する。
	db.Exec("TRUNCATE TABLE medical_records, pets, animal_species CASCADE")

	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	d := func(s string) time.Time {
		tm, err := time.Parse("2006-01-02", s)
		require.NoError(t, err)
		return tm
	}

	owner := makeTestOwner(t, db, clinicA, "山田太郎")
	petA := makeSpeciesAndPet(t, db, clinicA, owner.ID, "ポチ")
	// otherPet の所属医院（clinicB）はテスト対象外。pet スコープ除外の検証では
	// otherPet.ID を使った clinicA のカルテを作り、petA への問い合わせから除外されることを見る。
	otherPet := makeSpeciesAndPet(t, db, clinicB, owner.ID, "タマ")
	deletedOnlyPet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "ミケ")

	// clinicA / petA: 有効カルテの最古は 2023-03-01。
	makeMedicalRecordForPet(t, db, clinicA, petA.ID, d("2024-01-15"), false)
	makeMedicalRecordForPet(t, db, clinicA, petA.ID, d("2023-03-01"), false) // 最古の有効カルテ
	// より古いが論理削除済み → 除外される。
	makeMedicalRecordForPet(t, db, clinicA, petA.ID, d("2022-01-01"), true)
	// 同じ pet だが別医院 (clinicB) のより古いカルテ → clinic スコープで除外される。
	makeMedicalRecordForPet(t, db, clinicB, petA.ID, d("2020-01-01"), false)
	// clinicA の別ペットのより古いカルテ → pet スコープで除外される。
	makeMedicalRecordForPet(t, db, clinicA, otherPet.ID, d("2019-01-01"), false)
	// 全カルテ論理削除のペット。
	makeMedicalRecordForPet(t, db, clinicA, deletedOnlyPet.ID, d("2021-05-05"), true)

	t.Run("最古の有効カルテ date を返し、論理削除・別医院・別ペットを除外する", func(t *testing.T) {
		got, err := repo.FindFirstVisitDateByPetID(ctx, clinicA, petA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "2023-03-01", got.Format("2006-01-02"))
	})

	t.Run("カルテが無いペットは nil を返す（捏造しない）", func(t *testing.T) {
		got, err := repo.FindFirstVisitDateByPetID(ctx, clinicA, uint64(99999))
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("全カルテが論理削除されたペットは nil を返す", func(t *testing.T) {
		got, err := repo.FindFirstVisitDateByPetID(ctx, clinicA, deletedOnlyPet.ID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

// ---- FindAll（B-1: server-side pagination / search / filter 拡張） ----

// setupMedicalRecordListTestDB は FindAll のテストに必要な関連テーブル（pets/animal_species/staffs/inquiries）
// を整備した上で、B-1 で追加した search/filter 用の JOIN 対象データが他テストと混ざらないよう TRUNCATE する。
func setupMedicalRecordListTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.Inquiry{},
		&model.Billing{},
	))
	db.Exec("TRUNCATE TABLE billings, inquiries, staff_clinic_assignments, medical_records, pets, animal_species, staffs CASCADE")
	return db
}

// makeFullMedicalRecord は search/filter テスト用にフィールドを一通り指定してカルテを1件作成する。
func makeFullMedicalRecord(t *testing.T, db *gorm.DB, mr *model.MedicalRecord) *model.MedicalRecord {
	t.Helper()
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

type currentOwnerTransferFixture struct {
	PreviousOwner *model.Owner
	CurrentOwner  *model.Owner
	Pet           *model.Pet
	Record        *model.MedicalRecord
}

func makeCurrentOwnerTransferFixture(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	recordNo string,
	recordDate time.Time,
) currentOwnerTransferFixture {
	t.Helper()
	ctx := context.Background()
	previousOwner := makeTestOwner(t, db, clinicID, recordNo+" 譲渡前飼主")
	currentOwner := makeTestOwner(t, db, clinicID, recordNo+" 譲渡後飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, previousOwner.ID, recordNo+" 譲渡ペット")
	record := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     recordDate,
		OwnerID:  &previousOwner.ID,
		PetID:    &pet.ID,
		Status:   model.MedicalRecordStatusFinalized,
	})
	require.NoError(t, db.WithContext(ctx).Model(pet).Update("owner_id", currentOwner.ID).Error)
	return currentOwnerTransferFixture{
		PreviousOwner: previousOwner,
		CurrentOwner:  currentOwner,
		Pet:           pet,
		Record:        record,
	}
}

// makeInquiryForRecord は指定カルテに主訴（chief_complaint）付きの問診を1件作成する。
func makeInquiryForRecord(t *testing.T, db *gorm.DB, medicalRecordID uint64, chiefComplaint string) {
	t.Helper()
	inq := &model.Inquiry{MedicalRecordID: medicalRecordID, ChiefComplaint: chiefComplaint}
	require.NoError(t, db.WithContext(context.Background()).Create(inq).Error)
}

// TestMedicalRecordRepository_FindAll_ClinicIsolation は clinic_id 隔離が
// search/filter/pagination の全経路で維持されることを検証する（B-1）。
func TestMedicalRecordRepository_FindAll_ClinicIsolation(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "医院A飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "医院Aペット")
	ownerB := makeTestOwner(t, db, clinicB, "医院B飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "医院Bペット")

	recA := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "A-001", Date: time.Now(), OwnerID: &ownerA.ID, PetID: &petA.ID})
	makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicB, RecordNo: "B-001", Date: time.Now(), OwnerID: &ownerB.ID, PetID: &petB.ID})

	t.Run("指定した clinicIDs 以外のカルテは含まれない", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, recA.ID, got[0].ID)
	})

	t.Run("clinicIDs が空はフェイルセーフで空配列を返す", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{}, MedicalRecordListFilters{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, got)
	})

	t.Run("検索語が他院データにマッチしても混入しない", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: "医院B"}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, got)
	})
}

func TestMedicalRecordRepository_FindAllAndCount_CurrentOwnerAfterTransfer(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicID = uint64(70101)

	fixture := makeCurrentOwnerTransferFixture(
		t,
		db,
		clinicID,
		"MR-CURRENT-OWNER-TRANSFER",
		time.Now(),
	)

	got, total, err := repo.FindAll(ctx, []uint64{clinicID}, MedicalRecordListFilters{OwnerID: &fixture.CurrentOwner.ID}, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, got, 1)
	assert.Equal(t, fixture.Record.ID, got[0].ID)
	require.NotNil(t, got[0].OwnerID)
	assert.Equal(t, fixture.PreviousOwner.ID, *got[0].OwnerID, "returned owner_id remains the historical snapshot")

	oldRows, oldTotal, err := repo.FindAll(ctx, []uint64{clinicID}, MedicalRecordListFilters{OwnerID: &fixture.PreviousOwner.ID}, 1, 20)
	require.NoError(t, err)
	assert.Zero(t, oldTotal)
	assert.Empty(t, oldRows)

	currentCount, err := repo.CountByOwnerID(ctx, clinicID, fixture.CurrentOwner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), currentCount)
	previousCount, err := repo.CountByOwnerID(ctx, clinicID, fixture.PreviousOwner.ID)
	require.NoError(t, err)
	assert.Zero(t, previousCount)

	foreignOwner := makeTestOwner(t, db, clinicID+1, "別医院飼主")
	require.NoError(t, db.WithContext(ctx).Model(fixture.Pet).Update("owner_id", foreignOwner.ID).Error)
	foreignRows, foreignTotal, err := repo.FindAll(
		ctx,
		[]uint64{clinicID},
		MedicalRecordListFilters{OwnerID: &foreignOwner.ID},
		1,
		20,
	)
	require.NoError(t, err)
	assert.Zero(t, foreignTotal, "current owner must belong to the record clinic")
	assert.Empty(t, foreignRows)
	foreignCount, err := repo.CountByOwnerID(ctx, clinicID, foreignOwner.ID)
	require.NoError(t, err)
	assert.Zero(t, foreignCount)
}

func TestDB_MedicalRecordRepositoryFindAllCorrelatesRelationsToEachParentClinic(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)

	ownerA := makeTestOwner(t, db, clinicA, "会計Preload隔離飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "会計Preload隔離ペット")
	doctorA := makeMedicalRecordListStaff(t, db, clinicB, "カルテ一覧担当医A", model.StaffTypeDoctor)
	enteredByA := makeMedicalRecordListStaff(t, db, clinicB, "カルテ一覧入力者A", model.StaffTypeNurse)
	for _, staffID := range []uint64{doctorA.ID, enteredByA.ID} {
		require.NoError(t, db.Create(&model.StaffClinicAssignment{
			StaffID: staffID, ClinicID: clinicA,
		}).Error)
	}
	validRecord := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "MR-BILLING-VALID", Date: time.Now(),
		OwnerID: &ownerA.ID, PetID: &petA.ID, DoctorID: &doctorA.ID, EnteredBy: &enteredByA.ID,
	})
	foreignOnlyRecord := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "MR-BILLING-FOREIGN-ONLY", Date: time.Now().Add(-time.Hour),
		OwnerID: &ownerA.ID, PetID: &petA.ID,
	})
	ownerB := makeTestOwner(t, db, clinicB, "会計Preload隔離飼主B")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "会計Preload隔離ペットB")
	doctorB := makeMedicalRecordListStaff(t, db, clinicA, "カルテ一覧担当医B", model.StaffTypeDoctor)
	enteredByB := makeMedicalRecordListStaff(t, db, clinicA, "カルテ一覧入力者B", model.StaffTypeNurse)
	for _, staffID := range []uint64{doctorB.ID, enteredByB.ID} {
		require.NoError(t, db.Create(&model.StaffClinicAssignment{
			StaffID: staffID, ClinicID: clinicB,
		}).Error)
	}
	validRecordB := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicB, RecordNo: "MR-BILLING-VALID-B", Date: time.Now().Add(-2 * time.Hour),
		OwnerID: &ownerB.ID, PetID: &petB.ID, DoctorID: &doctorB.ID, EnteredBy: &enteredByB.ID,
	})

	unassignedDoctor := makeMedicalRecordListStaff(t, db, clinicA, "カルテ一覧未所属医師", model.StaffTypeDoctor)
	foreignEnteredBy := makeMedicalRecordListStaff(t, db, clinicB, "カルテ一覧別院入力者", model.StaffTypeNurse)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: foreignEnteredBy.ID, ClinicID: clinicB,
	}).Error)
	inactiveDoctor := makeMedicalRecordListStaff(t, db, clinicA, "カルテ一覧無効医師", model.StaffTypeDoctor)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: inactiveDoctor.ID, ClinicID: clinicA,
	}).Error)
	require.NoError(t, db.Model(inactiveDoctor).UpdateColumn("is_active", false).Error)
	inactiveDoctorRecord := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "MR-INACTIVE-DOCTOR", Date: time.Now(),
		OwnerID: &ownerA.ID, PetID: &petA.ID, DoctorID: &inactiveDoctor.ID,
	})

	unassignedDoctorRecord := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "MR-UNASSIGNED-DOCTOR", Date: time.Now(),
		OwnerID: &ownerA.ID, PetID: &petA.ID, DoctorID: &unassignedDoctor.ID,
	})
	pollutedRecords := []*model.MedicalRecord{
		makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "MR-FOREIGN-OWNER", Date: time.Now(),
			OwnerID: &ownerB.ID,
		}),
		makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "MR-FOREIGN-PET", Date: time.Now(),
			OwnerID: &ownerA.ID, PetID: &petB.ID,
		}),
		makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "MR-FOREIGN-ENTERED-BY", Date: time.Now(),
			OwnerID: &ownerA.ID, PetID: &petA.ID, EnteredBy: &foreignEnteredBy.ID,
		}),
	}

	validBilling := &model.Billing{
		ClinicID: clinicA, MedicalRecordID: &validRecord.ID,
		Status: model.BillingStatusWaiting, ScheduledDate: time.Now(),
	}
	require.NoError(t, db.WithContext(ctx).Create(validBilling).Error)
	foreignForValid := &model.Billing{
		ClinicID: clinicB, MedicalRecordID: &validRecord.ID,
		Status: model.BillingStatusWaiting, ScheduledDate: time.Now(),
	}
	require.NoError(t, db.WithContext(ctx).Create(foreignForValid).Error)
	foreignOnly := &model.Billing{
		ClinicID: clinicB, MedicalRecordID: &foreignOnlyRecord.ID,
		Status: model.BillingStatusWaiting, ScheduledDate: time.Now(),
	}
	require.NoError(t, db.WithContext(ctx).Create(foreignOnly).Error)
	validBillingB := &model.Billing{
		ClinicID: clinicB, MedicalRecordID: &validRecordB.ID,
		Status: model.BillingStatusWaiting, ScheduledDate: time.Now(),
	}
	require.NoError(t, db.WithContext(ctx).Create(validBillingB).Error)

	// Both clinics are intentionally authorized. An IN-clause-only preload would
	// still leak clinic B's polluted billing into clinic A's parent record.
	got, total, err := repo.FindAll(ctx, []uint64{clinicA, clinicB}, MedicalRecordListFilters{}, 1, 100)
	require.NoError(t, err)
	require.EqualValues(t, 5, total)

	byID := make(map[uint64]model.MedicalRecord, len(got))
	for _, record := range got {
		byID[record.ID] = record
	}
	validResult, ok := byID[validRecord.ID]
	require.True(t, ok)
	require.NotNil(t, validResult.Owner)
	assert.Equal(t, ownerA.ID, validResult.Owner.ID)
	require.NotNil(t, validResult.Pet)
	assert.Equal(t, petA.ID, validResult.Pet.ID)
	require.NotNil(t, validResult.Doctor)
	assert.Equal(t, doctorA.ID, validResult.Doctor.ID)
	require.NotNil(t, validResult.EnteredByStaff)
	assert.Equal(t, enteredByA.ID, validResult.EnteredByStaff.ID)
	require.NotNil(t, validResult.Billing, "same-clinic billing must be retained")
	assert.Equal(t, validBilling.ID, validResult.Billing.ID)
	assert.NotEqual(t, foreignForValid.ID, validResult.Billing.ID)

	foreignOnlyResult, ok := byID[foreignOnlyRecord.ID]
	require.True(t, ok)
	assert.Nil(t, foreignOnlyResult.Billing, "foreign-clinic billing must not populate accounting data")

	validResultB, ok := byID[validRecordB.ID]
	require.True(t, ok)
	require.NotNil(t, validResultB.Billing, "authorized clinic B's matching billing must be retained")
	assert.Equal(t, validBillingB.ID, validResultB.Billing.ID)

	inactiveDoctorResult, ok := byID[inactiveDoctorRecord.ID]
	require.True(t, ok, "inactive same-clinic staff must not hide the medical-record history")
	assert.Nil(t, inactiveDoctorResult.Doctor, "inactive staff is hidden by the current-relation preload")

	_, ok = byID[unassignedDoctorRecord.ID]
	assert.True(t, ok, "same-clinic staff without assignment must appear in the list")

	for _, polluted := range pollutedRecords {
		_, ok := byID[polluted.ID]
		assert.False(t, ok, "polluted medical record %d must fail closed", polluted.ID)
	}
}

func makeMedicalRecordListStaff(
	t *testing.T,
	db *gorm.DB,
	primaryClinicID uint64,
	name string,
	staffType model.StaffType,
) *model.Staff {
	t.Helper()
	staff := &model.Staff{
		ClinicID: primaryClinicID, Name: name, StaffType: staffType, IsActive: true,
	}
	require.NoError(t, db.Create(staff).Error)
	return staff
}

// TestMedicalRecordRepository_FindAll_Search は search が飼主名・ペット名・record_no・主訴を
// 部分一致で横断検索できることを検証する（B-1 AC-2）。
func TestMedicalRecordRepository_FindAll_Search(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ownerTarget := makeTestOwner(t, db, clinicA, "検索対象飼主タロウ")
	ownerOther := makeTestOwner(t, db, clinicA, "別の飼主ハナコ")
	// makeTestOwner/makeSpeciesAndPet は NameKana を設定しないため、カタカナを含む語で検索すると
	// NormalizeKana によりひらがな化されたパターンが生カタカナ列と文字種不一致でマッチしない
	// （owner_repository_test.go の同種コメント参照）。カナ正規化の影響を受けない漢字で検証する。
	petForOwnerTarget := makeSpeciesAndPet(t, db, clinicA, ownerTarget.ID, "飼主検索用ペット")
	petTarget := makeSpeciesAndPet(t, db, clinicA, ownerOther.ID, "検索対象犬")
	petOther := makeSpeciesAndPet(t, db, clinicA, ownerOther.ID, "別のペット")

	byOwnerName := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "R-OWNER", Date: time.Now(), OwnerID: &ownerTarget.ID, PetID: &petForOwnerTarget.ID})
	byPetName := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "R-PET", Date: time.Now(), OwnerID: &ownerOther.ID, PetID: &petTarget.ID})
	byRecordNo := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "R-検索対象-999", Date: time.Now(), OwnerID: &ownerOther.ID, PetID: &petOther.ID})
	byChiefComplaint := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "R-INQUIRY", Date: time.Now(), OwnerID: &ownerOther.ID, PetID: &petOther.ID})
	makeInquiryForRecord(t, db, byChiefComplaint.ID, "検索対象の主訴：嘔吐")
	makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "R-NOMATCH", Date: time.Now(), OwnerID: &ownerOther.ID, PetID: &petOther.ID})

	t.Run("飼主名で部分一致検索できる", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: "検索対象飼主"}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, byOwnerName.ID, got[0].ID)
	})

	t.Run("ペット名で部分一致検索できる", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: "検索対象犬"}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, byPetName.ID, got[0].ID)
	})

	t.Run("record_no で部分一致検索できる", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: "検索対象-999"}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, byRecordNo.ID, got[0].ID)
	})

	t.Run("主訴（chief_complaint）で部分一致検索できる", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: "検索対象の主訴"}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, byChiefComplaint.ID, got[0].ID)
	})

	t.Run("該当しない検索語は空を返す", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: "存在しない検索語XYZ"}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, got)
	})
}

// TestMedicalRecordRepository_FindAll_Filters は status / doctor_id / animal_species_id /
// start_date・end_date フィルタが独立して機能することを検証する（B-1 AC-2）。
func TestMedicalRecordRepository_FindAll_Filters(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "フィルタ検証飼主")
	dogPet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "犬ペット")
	catSpecies := &model.AnimalSpecies{Name: "猫"}
	require.NoError(t, db.WithContext(ctx).Create(catSpecies).Error)
	catPet := &model.Pet{ClinicID: clinicA, OwnerID: owner.ID, AnimalSpeciesID: catSpecies.ID, Name: "猫ペット"}
	require.NoError(t, db.WithContext(ctx).Create(catPet).Error)

	doctorA := makeDoctor(t, db, clinicA, "担当医A")
	doctorB := makeDoctor(t, db, clinicA, "担当医B")
	ensureVaccinationTestClinics(t, db, clinicA)
	for _, doctorID := range []uint64{doctorA.ID, doctorB.ID} {
		require.NoError(t, db.Create(&model.StaffClinicAssignment{
			StaffID: doctorID, ClinicID: clinicA,
		}).Error)
	}

	draft := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "F-DRAFT", Date: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		OwnerID: &owner.ID, PetID: &dogPet.ID, DoctorID: &doctorA.ID, Status: model.MedicalRecordStatusDraft,
	})
	finalized := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "F-FINAL", Date: time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
		OwnerID: &owner.ID, PetID: &catPet.ID, DoctorID: &doctorB.ID, Status: model.MedicalRecordStatusFinalized,
	})

	t.Run("status で絞り込める", func(t *testing.T) {
		statusDraft := model.MedicalRecordStatusDraft
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Status: &statusDraft}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, draft.ID, got[0].ID)
	})

	t.Run("doctor_id で絞り込める", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{DoctorID: &doctorB.ID}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, finalized.ID, got[0].ID)
	})

	t.Run("animal_species_id で絞り込める", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{AnimalSpeciesID: &catSpecies.ID}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, finalized.ID, got[0].ID)
	})

	t.Run("start_date/end_date で絞り込める（既存挙動の維持）", func(t *testing.T) {
		start := "2026-02-01"
		end := "2026-02-28"
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{StartDate: &start, EndDate: &end}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, finalized.ID, got[0].ID)
	})

	t.Run("フィルタ未指定時は全件返す", func(t *testing.T) {
		_, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
	})
}

// TestMedicalRecordRepository_FindAll_Sort は B-1 follow-up の列ソート server 化を検証する。
// 許可キー（date/owner_name/pet_name/status）ごとに ORDER BY が反映されること、
// 未許可キーは既定順（date DESC, created_at DESC）にフォールバックすることを確認する。
func TestMedicalRecordRepository_FindAll_Sort(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ownerA := makeTestOwner(t, db, clinicA, "Aソート飼主")
	ownerB := makeTestOwner(t, db, clinicA, "Bソート飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "Aソートペット")
	petB := makeSpeciesAndPet(t, db, clinicA, ownerB.ID, "Bソートペット")

	// 日付の昇順と record_no の対応が食い違うように意図的に作成順序をずらす。
	recNewer := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "S-NEW", Date: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerB.ID, PetID: &petB.ID, Status: model.MedicalRecordStatusFinalized,
	})
	recOlder := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "S-OLD", Date: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerA.ID, PetID: &petA.ID, Status: model.MedicalRecordStatusDraft,
	})

	t.Run("sort=date & order=asc は日付昇順になる", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Sort: "date", Order: "asc"}, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, recOlder.ID, got[0].ID)
		assert.Equal(t, recNewer.ID, got[1].ID)
	})

	t.Run("sort=date & order=desc は日付降順になる", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Sort: "date", Order: "desc"}, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, recNewer.ID, got[0].ID)
		assert.Equal(t, recOlder.ID, got[1].ID)
	})

	t.Run("sort=owner_name & order=asc は飼主名昇順になる（JOIN経由）", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Sort: "owner_name", Order: "asc"}, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, recOlder.ID, got[0].ID, "Aソート飼主 が先頭")
		assert.Equal(t, recNewer.ID, got[1].ID, "Bソート飼主 が次")
	})

	t.Run("sort=pet_name & order=desc はペット名降順になる（JOIN経由）", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Sort: "pet_name", Order: "desc"}, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, recNewer.ID, got[0].ID, "Bソートペット が先頭")
		assert.Equal(t, recOlder.ID, got[1].ID, "Aソートペット が次")
	})

	t.Run("sort=status & order=asc はステータス昇順になる", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Sort: "status", Order: "asc"}, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, string(model.MedicalRecordStatusDraft), string(got[0].Status))
	})

	t.Run("未許可の sort キーは既定順（date DESC）にフォールバックする", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Sort: "unknown_column", Order: "asc"}, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, recNewer.ID, got[0].ID, "既定順(date DESC)を維持する")
		assert.Equal(t, recOlder.ID, got[1].ID)
	})
}

// TestMedicalRecordRepository_FindAll_Pagination は total > limit のとき
// page1 と page2 が重複しないことを固定する回帰テスト（B-1 旧 failure mode 再発防止 / harness P1）。
func TestMedicalRecordRepository_FindAll_Pagination(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "ページング検証飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "ページング検証ペット")

	const totalRecords = 25
	const limit = 20
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ids := make([]uint64, 0, totalRecords)
	for i := 0; i < totalRecords; i++ {
		mr := makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA,
			RecordNo: fmt.Sprintf("P-%03d", i),
			// 日付をずらして採番する（BE デフォルトソート: date DESC, created_at DESC）
			Date:    base.AddDate(0, 0, i),
			OwnerID: &owner.ID,
			PetID:   &pet.ID,
		})
		ids = append(ids, mr.ID)
	}

	t.Run("total は limit を超えても正しい件数を返す（旧 failure mode: 常に20件表示の再発防止）", func(t *testing.T) {
		_, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{}, 1, limit)
		require.NoError(t, err)
		assert.Equal(t, int64(totalRecords), total, "total は limit=20 を超えた実件数(25件)を返すべき")
	})

	t.Run("page1 は limit 件、page2 は残り件数を返し重複しない", func(t *testing.T) {
		page1, total1, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{}, 1, limit)
		require.NoError(t, err)
		page2, total2, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{}, 2, limit)
		require.NoError(t, err)

		assert.Equal(t, int64(totalRecords), total1)
		assert.Equal(t, int64(totalRecords), total2)
		assert.Len(t, page1, limit, "page1 は limit 件（20件）返すべき")
		assert.Len(t, page2, totalRecords-limit, "page2 は残り件数（5件）返すべき")

		seen := map[uint64]bool{}
		for _, r := range page1 {
			seen[r.ID] = true
		}
		for _, r := range page2 {
			assert.False(t, seen[r.ID], "page2 のレコードが page1 と重複してはならない（B-1 旧 failure mode: 常に先頭20件のみ表示の再発防止）")
			seen[r.ID] = true
		}
		for _, id := range ids {
			assert.True(t, seen[id], "作成した全レコード(id=%d)が page1+page2 のいずれかに含まれるべき（欠落防止）", id)
		}
		assert.Len(t, seen, len(ids), "page1+page2 の合計件数は作成件数と一致すべき（重複/欠落なし）")
	})
}

// ---- H-7: CRUD 系メソッドの正常系・clinic_id 隔離カバレッジ拡充 ----
// FindByID/FindByIDForClinics/Create/Update/Delete/CountByPetID/CountByOwnerID は
// FindAll 系テストでは経路が通らず未カバーだったため、正常系と越境防止を中心に追加する。

// TestMedicalRecordRepository_FindByID は正常取得と他院からのアクセス拒否を検証する。
func TestMedicalRecordRepository_FindByID(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "FindByID飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "FindByIDペット")
	rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "FBI-001", Date: time.Now(), OwnerID: &ownerA.ID, PetID: &petA.ID,
	})

	t.Run("同一医院からは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, rec.ID, got.ID)
		assert.Equal(t, "FBI-001", got.RecordNo)
	})

	t.Run("他院からは NotFound（clinic_id 越境防止）", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, uint64(999999))
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("BUG-406: 保存済み Inquiry が FindByID の結果に含まれる", func(t *testing.T) {
		makeInquiryForRecord(t, db, rec.ID, "再読込後も残るはずの主訴")

		got, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.NoError(t, err)
		require.NotNil(t, got.Inquiry, "Inquiry が Preload されていない（BUG-406 根本原因）")
		assert.Equal(t, "再読込後も残るはずの主訴", got.Inquiry.ChiefComplaint)
	})

	t.Run("所属未登録の担当医・入力者でも詳細は取得できる", func(t *testing.T) {
		// seed 検査の medical_record_id は存在するが、imported doctor に
		// staff_clinic_assignments が無い。一覧スコープで詳細まで隠すと
		// 検査管理からのカルテ検査タブが 404 になる。
		unassigned := makeMedicalRecordListStaff(t, db, clinicA, "所属未登録医師", model.StaffTypeDoctor)
		recUnassigned := makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "FBI-UNASSIGNED", Date: time.Now(),
			OwnerID: &ownerA.ID, PetID: &petA.ID, DoctorID: &unassigned.ID, EnteredBy: &unassigned.ID,
		})

		got, err := repo.FindByID(ctx, clinicA, recUnassigned.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, recUnassigned.ID, got.ID)
	})
}

// TestMedicalRecordRepository_FindByIDForClinics はマルチクリニック横断取得の
// clinic_id 隔離（許可リスト外は拒否）を検証する。
func TestMedicalRecordRepository_FindByIDForClinics(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB, clinicC = uint64(1), uint64(2), uint64(3)

	ownerA := makeTestOwner(t, db, clinicA, "FBIFC飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "FBIFCペット")
	rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "FBIFC-001", Date: time.Now(), OwnerID: &ownerA.ID, PetID: &petA.ID,
	})

	t.Run("許可リストに所属医院が含まれれば取得できる", func(t *testing.T) {
		got, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, rec.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, rec.ID, got.ID)
	})

	t.Run("許可リストに所属医院が含まれなければ NotFound", func(t *testing.T) {
		_, err := repo.FindByIDForClinics(ctx, []uint64{clinicB, clinicC}, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("許可リストが空はフェイルセーフで NotFound", func(t *testing.T) {
		_, err := repo.FindByIDForClinics(ctx, []uint64{}, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestMedicalRecordRepository_Create は作成の正常系と record_no 一意制約違反を検証する。
func TestMedicalRecordRepository_Create(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "Create飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "Createペット")

	t.Run("正常系: Create 後 FindByID で取得できる", func(t *testing.T) {
		rec := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "CR-001", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID}
		require.NoError(t, repo.Create(ctx, rec))
		require.NotZero(t, rec.ID)

		got, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, model.MedicalRecordStatusDraft, got.Status, "既定ステータスは draft であるべき")
	})
}

// TestMedicalRecordRepository_Delete は論理削除の正常系と clinic_id 越境削除の拒否を検証する。
func TestMedicalRecordRepository_Delete(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "Delete飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "Deleteペット")

	t.Run("同一医院からの削除は成功し以降 FindByID で NotFound になる", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "DEL-001", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
		require.NoError(t, repo.Delete(ctx, clinicA, rec.ID))

		_, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("他院からの削除は NotFound（越境削除の拒否）", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "DEL-XCLINIC", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
		err := repo.Delete(ctx, clinicB, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		// 実データが削除されていないことも確認する。
		got, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, "DEL-XCLINIC", got.RecordNo)
	})

	t.Run("存在しない ID の削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, uint64(999999))
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("確定済みカルテは削除できない", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "DEL-FINALIZED", Date: time.Now(),
			OwnerID: &owner.ID, PetID: &pet.ID, Status: model.MedicalRecordStatusFinalized,
		})

		err := repo.Delete(ctx, clinicA, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))

		var persisted model.MedicalRecord
		require.NoError(t, db.Unscoped().First(&persisted, rec.ID).Error)
		assert.Equal(t, model.MedicalRecordStatusFinalized, persisted.Status)
		assert.False(t, persisted.DeletedAt.Valid)
	})

	t.Run("確定commit待機後の削除はConflictになり確定記録を残す", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "DEL-FINALIZE-RACE", Date: time.Now(),
			OwnerID: &owner.ID, PetID: &pet.ID, Status: model.MedicalRecordStatusDraft,
		})
		tx := db.Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()
		txCtx := persistence.WithTxValue(ctx, tx)
		_, err := repo.Update(txCtx, clinicA, rec.ID, map[string]any{
			"status": model.MedicalRecordStatusFinalized,
		}, nil)
		require.NoError(t, err)

		deleteDone := make(chan error, 1)
		go func() { deleteDone <- repo.Delete(ctx, clinicA, rec.ID) }()
		select {
		case err := <-deleteDone:
			require.Failf(t, "delete did not wait for finalization", "err=%v", err)
		case <-time.After(100 * time.Millisecond):
		}
		require.NoError(t, tx.Commit().Error)

		err = <-deleteDone
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		var persisted model.MedicalRecord
		require.NoError(t, db.Unscoped().First(&persisted, rec.ID).Error)
		assert.Equal(t, model.MedicalRecordStatusFinalized, persisted.Status)
		assert.False(t, persisted.DeletedAt.Valid)
	})
}

// TestMedicalRecordRepository_CountByPetID_CountByOwnerID は集計系メソッドの
// clinic_id 隔離・論理削除除外を検証する。
func TestMedicalRecordRepository_CountByPetID_CountByOwnerID(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "Count飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "Countペット")
	otherOwner := makeTestOwner(t, db, clinicB, "他院Count飼主")
	otherPet := makeSpeciesAndPet(t, db, clinicB, otherOwner.ID, "他院Countペット")

	makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "CNT-001", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
	makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "CNT-002", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
	makeMedicalRecordForPet(t, db, clinicA, pet.ID, time.Now(), true) // 論理削除は除外
	makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicB, RecordNo: "CNT-OTHER", Date: time.Now(), OwnerID: &otherOwner.ID, PetID: &otherPet.ID})

	t.Run("CountByPetID は同一医院・有効カルテのみ数える", func(t *testing.T) {
		count, err := repo.CountByPetID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("CountByPetID は他院からは0を返す（clinic_id 隔離）", func(t *testing.T) {
		count, err := repo.CountByPetID(ctx, clinicB, pet.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CountByOwnerID は同一医院・有効カルテのみ数える", func(t *testing.T) {
		count, err := repo.CountByOwnerID(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("CountByOwnerID は他院からは0を返す（clinic_id 隔離）", func(t *testing.T) {
		count, err := repo.CountByOwnerID(ctx, clinicB, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestMedicalRecordRepository_CountByPetAndDate_IsScopedAndJoinsAmbientTransaction(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	date := time.Date(2026, time.August, 1, 0, 0, 0, 0, config.JST)

	ownerA := makeTestOwner(t, db, clinicA, "日付集計A飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "日付集計Aペット")
	otherPetA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "日付集計A別ペット")
	ownerB := makeTestOwner(t, db, clinicB, "日付集計B飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "日付集計Bペット")

	makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "DATE-A-MATCH", Date: date,
		OwnerID: &ownerA.ID, PetID: &petA.ID,
	})
	makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "DATE-A-OTHER-PET", Date: date,
		OwnerID: &ownerA.ID, PetID: &otherPetA.ID,
	})
	makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "DATE-A-OTHER-DAY", Date: date.AddDate(0, 0, 1),
		OwnerID: &ownerA.ID, PetID: &petA.ID,
	})
	makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicB, RecordNo: "DATE-B-MATCH", Date: date,
		OwnerID: &ownerB.ID, PetID: &petB.ID,
	})
	deleted := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "DATE-A-DELETED", Date: date,
		OwnerID: &ownerA.ID, PetID: &petA.ID,
	})
	require.NoError(t, db.Delete(deleted).Error)

	count, err := repo.CountByPetAndDate(ctx, clinicA, petA.ID, "2026-08-01")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	foreignCount, err := repo.CountByPetAndDate(ctx, clinicB, petA.ID, "2026-08-01")
	require.NoError(t, err)
	assert.Equal(t, int64(0), foreignCount)

	tx := db.Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()
	txCtx := persistence.WithTxValue(ctx, tx)
	require.NoError(t, repo.Create(txCtx, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "DATE-A-UNCOMMITTED", Date: date,
		OwnerID: &ownerA.ID, PetID: &petA.ID, Status: model.MedicalRecordStatusDraft,
	}))
	inTxCount, err := repo.CountByPetAndDate(txCtx, clinicA, petA.ID, "2026-08-01")
	require.NoError(t, err)
	assert.Equal(t, int64(2), inTxCount)
	require.NoError(t, tx.Rollback().Error)

	afterRollback, err := repo.CountByPetAndDate(ctx, clinicA, petA.ID, "2026-08-01")
	require.NoError(t, err)
	assert.Equal(t, int64(1), afterRollback)
}
