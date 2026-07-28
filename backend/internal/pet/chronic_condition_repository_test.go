package pet

// pet_chronic_condition_repository_test.go — PetChronicConditionRepository の統合テスト（BE-012）。
//
// 注意: Update/Delete は updateScopedByID/deleteScopedByID（helpers.go）を使い RowsAffected==0 を
// NotFound として検査する（F3）。存在しない ID / 別クリニックの ID を渡すと apperrors.IsNotFound
// な err を返す。

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

// setupPetChronicConditionTestDB は pet_chronic_condition_repository のテスト用に DB を整備する。
func setupPetChronicConditionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.PetChronicCondition{}))
	db.Exec("TRUNCATE TABLE pet_chronic_conditions CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func makeChronicCondition(t *testing.T, db *gorm.DB, clinicID, petID uint64, code, name string, diagnosedAt time.Time, isActive bool) *model.PetChronicCondition {
	t.Helper()
	c := &model.PetChronicCondition{
		ClinicID:      clinicID,
		PetID:         petID,
		ConditionCode: code,
		ConditionName: name,
		DiagnosedAt:   diagnosedAt,
		IsActive:      isActive,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c
}

func TestPetChronicConditionRepository_FindByPetID(t *testing.T) {
	db := setupPetChronicConditionTestDB(t)
	repo := NewChronicConditionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "慢性疾患飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "慢性疾患ペット")

	old := makeChronicCondition(t, db, clinicA, pet.ID, "DM", "糖尿病", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), true)
	recent := makeChronicCondition(t, db, clinicA, pet.ID, "CKD", "慢性腎臓病", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), true)
	deleted := makeChronicCondition(t, db, clinicA, pet.ID, "OLD", "削除済み疾患", time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), false)
	require.NoError(t, db.WithContext(ctx).Delete(deleted).Error)

	// 別クリニックの同一ペットID疾患（実際は起こり得ないが clinic_id 述語の検証として作成）
	otherOwner := makeTestOwner(t, db, clinicB, "別クリニック慢性疾患飼主")
	otherPet := makeSpeciesAndPet(t, db, clinicB, otherOwner.ID, "別クリニックペット")
	makeChronicCondition(t, db, clinicB, otherPet.ID, "OTH", "別クリニック疾患", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), true)

	got, err := repo.FindByPetID(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	require.Len(t, got, 2, "ソフト削除を除外して2件")
	assert.Equal(t, recent.ID, got[0].ID, "diagnosed_at DESC で最新が先頭")
	assert.Equal(t, old.ID, got[1].ID)
}

func TestPetChronicConditionRepository_FindByPetID_RejectsCorruptCrossClinicPetRelation(t *testing.T) {
	db := setupPetChronicConditionTestDB(t)
	repo := NewChronicConditionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "慢性疾患取得対象飼主")
	crossClinicPet := makeSpeciesAndPet(t, db, clinicB, ownerA.ID, "破損した別医院ペット")
	makeChronicCondition(
		t,
		db,
		clinicA,
		crossClinicPet.ID,
		"LEAK",
		"別医院ペット由来疾患",
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		true,
	)

	got, err := repo.FindByPetID(ctx, clinicA, crossClinicPet.ID)

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestPetChronicConditionRepository_FindByID(t *testing.T) {
	db := setupPetChronicConditionTestDB(t)
	repo := NewChronicConditionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "単件取得飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "単件取得ペット")
	cond := makeChronicCondition(t, db, clinicA, pet.ID, "ASTH", "喘息", time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), true)

	t.Run("同一クリニックでは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, pet.ID, cond.ID)
		require.NoError(t, err)
		assert.Equal(t, "喘息", got.ConditionName)
	})

	t.Run("別クリニックからは取得できない(clinic_id 隔離)", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, pet.ID, cond.ID)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, pet.ID, 999888006)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("同一クリニックの別ペットからは取得できない", func(t *testing.T) {
		otherPet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "別ペット")
		got, err := repo.FindByID(ctx, clinicA, otherPet.ID, cond.ID)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("ソフト削除済みはGORM標準スコープにより取得できない", func(t *testing.T) {
		deleted := makeChronicCondition(t, db, clinicA, pet.ID, "DEL", "削除予定疾患", time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), true)
		require.NoError(t, db.WithContext(ctx).Delete(deleted).Error)

		got, err := repo.FindByID(ctx, clinicA, pet.ID, deleted.ID)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestPetChronicConditionRepository_FindByID_RejectsCorruptCrossClinicPetRelation(t *testing.T) {
	db := setupPetChronicConditionTestDB(t)
	repo := NewChronicConditionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "慢性疾患単件取得対象飼主")
	crossClinicPet := makeSpeciesAndPet(t, db, clinicB, ownerA.ID, "破損した別医院ペット")
	cond := makeChronicCondition(
		t,
		db,
		clinicA,
		crossClinicPet.ID,
		"LEAK",
		"別医院ペット由来疾患",
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		true,
	)

	got, err := repo.FindByID(ctx, clinicA, crossClinicPet.ID, cond.ID)

	assert.Nil(t, got)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestPetChronicConditionRepository_Create(t *testing.T) {
	db := setupPetChronicConditionTestDB(t)
	repo := NewChronicConditionRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "新規作成飼主(慢性疾患)")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "新規作成ペット(慢性疾患)")

	record := &model.PetChronicCondition{
		ClinicID:      clinicA,
		PetID:         pet.ID,
		ConditionCode: "HD",
		ConditionName: "股関節形成不全",
		DiagnosedAt:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		IsActive:      true,
	}
	err := repo.Create(ctx, record)
	require.NoError(t, err)
	assert.NotZero(t, record.ID)

	got, err := repo.FindByID(ctx, clinicA, pet.ID, record.ID)
	require.NoError(t, err)
	assert.Equal(t, "股関節形成不全", got.ConditionName)
}

func TestPetChronicConditionRepository_Update(t *testing.T) {
	db := setupPetChronicConditionTestDB(t)
	repo := NewChronicConditionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "更新テスト飼主(慢性疾患)")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "更新テストペット(慢性疾患)")
	otherPet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "更新テスト別ペット(慢性疾患)")
	cond := makeChronicCondition(t, db, clinicA, pet.ID, "ALG", "アレルギー", time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC), true)

	t.Run("成功", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, pet.ID, cond.ID, map[string]any{"condition_name": "食物アレルギー"})
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, pet.ID, cond.ID)
		require.NoError(t, err)
		assert.Equal(t, "食物アレルギー", got.ConditionName)
	})

	t.Run("別クリニックからのUpdateはNotFoundになり行を変更しない", func(t *testing.T) {
		err := repo.Update(ctx, clinicB, pet.ID, cond.ID, map[string]any{"condition_name": "不正書き換え"})
		assert.True(t, apperrors.IsNotFound(err), "別クリニックからの Update は RowsAffected==0 で NotFound になるべき")

		got, err := repo.FindByID(ctx, clinicA, pet.ID, cond.ID)
		require.NoError(t, err)
		assert.Equal(t, "食物アレルギー", got.ConditionName, "別クリニックからの Update で値が変わってはならない")
	})

	t.Run("同一クリニックの別ペットからのUpdateはNotFoundになり行を変更しない", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, otherPet.ID, cond.ID, map[string]any{"condition_name": "不正書き換え"})
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, pet.ID, cond.ID)
		require.NoError(t, err)
		assert.Equal(t, "食物アレルギー", got.ConditionName)
	})
}

func TestPetChronicConditionRepository_Delete(t *testing.T) {
	db := setupPetChronicConditionTestDB(t)
	repo := NewChronicConditionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "削除テスト飼主(慢性疾患)")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "削除テストペット(慢性疾患)")
	otherPet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "削除テスト別ペット(慢性疾患)")

	t.Run("成功", func(t *testing.T) {
		cond := makeChronicCondition(t, db, clinicA, pet.ID, "SKIN", "皮膚炎", time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC), true)
		err := repo.Delete(ctx, clinicA, pet.ID, cond.ID)
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, clinicA, pet.ID, cond.ID)
		assert.True(t, apperrors.IsNotFound(err), "削除後は NotFound になるべき")
	})

	t.Run("別クリニックからのDeleteはNotFoundになり対象行を削除しない", func(t *testing.T) {
		cond := makeChronicCondition(t, db, clinicA, pet.ID, "EYE", "白内障", time.Date(2026, 1, 25, 0, 0, 0, 0, time.UTC), true)
		err := repo.Delete(ctx, clinicB, pet.ID, cond.ID)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックからの Delete は RowsAffected==0 で NotFound になるべき")

		got, err := repo.FindByID(ctx, clinicA, pet.ID, cond.ID)
		require.NoError(t, err, "別クリニックからの Delete では削除されないはず")
		assert.Equal(t, cond.ID, got.ID)
	})

	t.Run("同一クリニックの別ペットからのDeleteはNotFoundになり対象行を削除しない", func(t *testing.T) {
		cond := makeChronicCondition(t, db, clinicA, pet.ID, "EAR", "外耳炎", time.Date(2026, 1, 26, 0, 0, 0, 0, time.UTC), true)
		err := repo.Delete(ctx, clinicA, otherPet.ID, cond.ID)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, pet.ID, cond.ID)
		require.NoError(t, err)
		assert.Equal(t, cond.ID, got.ID)
	})
}

func TestPetChronicConditionRepository_FindActiveConditionCodesByOwner(t *testing.T) {
	db := setupPetChronicConditionTestDB(t)
	repo := NewChronicConditionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "アクティブ疾患飼主")
	deceasedAt := time.Now().Add(-24 * time.Hour)
	aliveSpeciesID := makeSyncSpeciesID(t, db)

	petAlive1 := makePetDetailed(t, db, clinicA, owner.ID, aliveSpeciesID, "アクティブ疾患ペット1", nil, nil)
	petAlive2 := makePetDetailed(t, db, clinicA, owner.ID, aliveSpeciesID, "アクティブ疾患ペット2", nil, nil)
	petDeceased := makePetDetailed(t, db, clinicA, owner.ID, aliveSpeciesID, "死亡ペット", nil, &deceasedAt)

	makeChronicCondition(t, db, clinicA, petAlive1.ID, "DM", "糖尿病", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), true)
	makeChronicCondition(t, db, clinicA, petAlive1.ID, "DM", "糖尿病(重複コード)", time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), true)
	makeChronicCondition(t, db, clinicA, petAlive2.ID, "CKD", "慢性腎臓病", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), true)
	makeChronicCondition(t, db, clinicA, petAlive2.ID, "INACTIVE", "非アクティブ疾患", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), false)
	inactiveDeleted := makeChronicCondition(t, db, clinicA, petAlive1.ID, "DELETED", "削除済み疾患", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), true)
	require.NoError(t, db.WithContext(ctx).Delete(inactiveDeleted).Error)
	makeChronicCondition(t, db, clinicA, petDeceased.ID, "DEAD", "死亡ペットの疾患", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), true)

	// 別クリニックの疾患は混入しない
	otherOwner := makeTestOwner(t, db, clinicB, "別クリニックアクティブ疾患飼主")
	otherPet := makeSpeciesAndPet(t, db, clinicB, otherOwner.ID, "別クリニックペット(疾患)")
	makeChronicCondition(t, db, clinicB, otherPet.ID, "OTH", "別クリニック疾患", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), true)

	got, err := repo.FindActiveConditionCodesByOwner(ctx, clinicA, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"CKD", "DM"}, got, "DISTINCT + ORDER BY condition_code、非アクティブ/削除済み/死亡ペット/別クリニックは除外")
}

func TestPetChronicConditionRepository_FindActiveConditionCodesByOwner_RejectsCorruptCrossClinicPetRelation(t *testing.T) {
	db := setupPetChronicConditionTestDB(t)
	repo := NewChronicConditionRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "タグ同期対象飼主")
	crossClinicPet := makeSpeciesAndPet(t, db, clinicB, ownerA.ID, "破損した別医院ペット")
	makeChronicCondition(
		t,
		db,
		clinicA,
		crossClinicPet.ID,
		"LEAK",
		"別医院ペット由来疾患",
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		true,
	)

	got, err := repo.FindActiveConditionCodesByOwner(ctx, clinicA, ownerA.ID)

	require.NoError(t, err)
	assert.Empty(t, got)
}
