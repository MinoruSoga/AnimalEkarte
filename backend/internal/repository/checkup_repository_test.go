package repository

// checkup_repository_test.go — CheckupRepository の統合テスト（カバレッジ向上）。
//
// 対象: FindByClinicID / FindByMedicalRecordID / FindByOwnerID / FindByID /
//       Create / Update / Delete
// 検証観点: 正常系、フィルタ、clinic_id 隔離、ソフトデリート除外、NotFound ラップ。
//
// CheckupType Preload の clinic 隔離は master_preload_clinic_isolation_test.go の
// TestCheckupRepository_FindByID_CheckupTypePreloadClinicIsolation で別途検証済みのため、
// 本ファイルでは重複させない。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupCheckupRepoTestDB は checkups / checkup_types / medical_records 周りを整備する。
func setupCheckupRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.CheckupType{}, &model.Checkup{}))
	db.Exec("TRUNCATE TABLE checkups CASCADE")
	db.Exec("TRUNCATE TABLE checkup_types CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
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

func TestCheckupRepository_FindByClinicID_FiltersAndClinicIsolation(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "健診一覧飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "健診一覧ポチA")
	mrA := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-LIST-A", time.Now())
	ctA := makeCheckupTypeMaster(t, db, clinicA, "医院Aの健診")

	ownerB := makeTestOwner(t, db, clinicB, "健診一覧飼主B")
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
}

func TestCheckupRepository_FindByMedicalRecordID(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "健診MR飼主")
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

// FindByOwnerID は ISSUE-004 タグ再同期用: medical_records 経由で owner_id を解決する。
func TestCheckupRepository_FindByOwnerID(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "健診飼主同期")
	otherOwner := makeTestOwner(t, db, clinicA, "別の飼主")
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
	require.Len(t, got, 1, "指定飼主に紐づく健診のみ返る")
	assert.Equal(t, own.ID, got[0].ID)
	for _, c := range got {
		assert.NotEqual(t, other.ID, c.ID)
	}
}

func TestCheckupRepository_FindByID(t *testing.T) {
	db := setupCheckupRepoTestDB(t)
	repo := NewCheckupRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "健診単体飼主")
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

	owner := makeTestOwner(t, db, clinicA, "新規健診飼主")
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

	owner := makeTestOwner(t, db, clinicA, "更新健診飼主")
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

	owner := makeTestOwner(t, db, clinicA, "削除健診飼主")
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
