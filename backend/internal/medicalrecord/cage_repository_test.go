package medicalrecord

// cage_repository_test.go — CageRepository の統合テスト（実 Postgres テスト DB）。
// happy path・not-found・clinic_id 隔離・ソフトデリート除外・Update/Delete/Reorder・
// CountUsageByCageID（hospitalizations 経由・P2 deleted_at 除外）を対象とする。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupCageRepositoryTestDB は cages テーブルと、CountUsageByCageID が参照する
// hospitalizations / owners / pets / animal_species を用意する。
func setupCageRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Cage{}, &model.AnimalSpecies{}, &model.Owner{}, &model.Pet{}, &model.Hospitalization{},
	))
	db.Exec("TRUNCATE TABLE hospitalizations CASCADE")
	db.Exec("TRUNCATE TABLE cages CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE owners CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func TestCageRepository_Create_And_FindByID(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := NewCageRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	t.Run("happy path: Create してから FindByID で取得できる", func(t *testing.T) {
		price := int64(2000)
		c := &model.Cage{ClinicID: clinicID, Name: "ケージ1", CageType: model.CageTypeDog, CageSize: model.CageSizeMedium, Price: &price}
		require.NoError(t, repo.Create(ctx, c))
		require.NotZero(t, c.ID)

		got, err := repo.FindByID(ctx, clinicID, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "ケージ1", got.Name)
		require.NotNil(t, got.Price)
		assert.Equal(t, int64(2000), *got.Price)
		assert.Equal(t, model.CageTypeDog, got.CageType)
	})

	t.Run("存在しない ID は NotFound エラー", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicID, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("別クリニックからは FindByID できない（clinic_id 隔離）", func(t *testing.T) {
		c := makeCageMaster(t, db, clinicID, "医院1限定ケージ")
		_, err := repo.FindByID(ctx, uint64(999), c.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックからは NotFound であるべき: %v", err)
	})
}

func TestCageRepository_FindAll_TypeFilterAndClinicIsolation(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := NewCageRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	require.NoError(t, db.WithContext(ctx).Create(&model.Cage{ClinicID: clinicA, Name: "犬ケージ1", CageType: model.CageTypeDog, CageSize: model.CageSizeSmall}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Cage{ClinicID: clinicA, Name: "犬ケージ2", CageType: model.CageTypeDog, CageSize: model.CageSizeSmall}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Cage{ClinicID: clinicA, Name: "猫ケージ1", CageType: model.CageTypeCat, CageSize: model.CageSizeSmall}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Cage{ClinicID: clinicB, Name: "他院ケージ", CageType: model.CageTypeDog, CageSize: model.CageSizeSmall}).Error)

	t.Run("clinicA 全件は3件", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA, nil)
		require.NoError(t, err)
		assert.Len(t, got, 3)
	})

	t.Run("cageType フィルタで犬ケージのみ2件", func(t *testing.T) {
		dogStr := "dog"
		got, err := repo.FindAll(ctx, clinicA, &dogStr)
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("clinicB は自院の1件のみ（混入なし）", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB, nil)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "他院ケージ", got[0].Name)
	})
}

func TestCageRepository_FindAll_ExcludesSoftDeleted(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := NewCageRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	active := makeCageMaster(t, db, clinicID, "現役ケージ")
	deleted := makeCageMaster(t, db, clinicID, "削除済みケージ")
	require.NoError(t, repo.Delete(ctx, clinicID, deleted.ID))

	got, err := repo.FindAll(ctx, clinicID, nil)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, active.ID, got[0].ID)

	// 行自体は DB に残っている（ソフトデリート = deleted_at セットのみ）
	var rawCount int64
	db.Unscoped().Model(&model.Cage{}).Where("id = ?", deleted.ID).Count(&rawCount)
	assert.Equal(t, int64(1), rawCount, "ソフトデリートされた行はDBにまだ存在する")
}

func TestCageRepository_Update(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := NewCageRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c := makeCageMaster(t, db, clinicA, "更新前ケージ")

	t.Run("同一クリニックでは Update が反映される", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, c.ID, map[string]any{"name": "更新後ケージ"})
		require.NoError(t, err)
		assert.Equal(t, "更新後ケージ", got.Name)
	})

	t.Run("別クリニックからの Update は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, c.ID, map[string]any{"name": "改ざん試行"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)

		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後ケージ", got.Name, "別クリニックからの Update で名称が変わってはならない")
	})

	t.Run("存在しない ID の Update は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"name": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestCageRepository_Delete(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := NewCageRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c := makeCageMaster(t, db, clinicA, "削除対象ケージ")

	t.Run("別クリニックからの Delete は NotFound で行が残る", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, c.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		assert.Equal(t, c.ID, got.ID, "別クリニックからの Delete で行が消えてはならない")
	})

	t.Run("同一クリニックでは Delete が成功し以後 FindByID は NotFound", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, c.ID))
		_, err := repo.FindByID(ctx, clinicA, c.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の Delete は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestCageRepository_Reorder(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := NewCageRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c1 := makeCageMaster(t, db, clinicA, "ケージX")
	c2 := makeCageMaster(t, db, clinicA, "ケージY")
	c3 := makeCageMaster(t, db, clinicA, "ケージZ")

	t.Run("並び順が指定順に更新される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{c3.ID, c1.ID, c2.ID}))

		got, err := repo.FindAll(ctx, clinicA, nil)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, c3.ID, got[0].ID)
		assert.Equal(t, c1.ID, got[1].ID)
		assert.Equal(t, c2.ID, got[2].ID)
	})

	t.Run("別クリニックの ID を含む Reorder はエラーで中断する", func(t *testing.T) {
		other := makeCageMaster(t, db, clinicB, "他院ケージ")
		err := repo.Reorder(ctx, clinicA, []uint64{c1.ID, other.ID})
		require.Error(t, err, "clinicA スコープに存在しない ID を含む Reorder は失敗すべき")
	})
}

func TestCageRepository_CountUsageByCageID(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := NewCageRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	cage := makeCageMaster(t, db, clinicA, "使用中ケージ")
	owner := makeTestOwner(t, db, clinicA, "飼主A")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "ポチ")
	_ = makeHospitalizationRec(t, db, clinicA, owner.ID, pet.ID, &cage.ID)

	t.Run("同一クリニックでは1件カウントされる", func(t *testing.T) {
		count, err := repo.CountUsageByCageID(ctx, clinicA, cage.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックからは0件（clinic_id 隔離）", func(t *testing.T) {
		count, err := repo.CountUsageByCageID(ctx, clinicB, cage.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("未使用ケージは0件", func(t *testing.T) {
		unused := makeCageMaster(t, db, clinicA, "未使用ケージ")
		count, err := repo.CountUsageByCageID(ctx, clinicA, unused.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリート済みの入院は数えない（P2）", func(t *testing.T) {
		cage2 := makeCageMaster(t, db, clinicA, "削除対象ケージ")
		h := makeHospitalizationRec(t, db, clinicA, owner.ID, pet.ID, &cage2.ID)
		require.NoError(t, db.WithContext(ctx).Delete(h).Error)

		count, err := repo.CountUsageByCageID(ctx, clinicA, cage2.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "ソフトデリート済みの入院は使用件数に含まれない")
	})
}
