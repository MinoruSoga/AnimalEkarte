package medicalrecord

// vaccine_repository_test.go — VaccineRepository の統合テスト（実 Postgres テスト DB）。
// happy path・species フィルタ・not-found・clinic_id 隔離・ソフトデリート除外・
// Update/Delete/Reorder・CountUsageByVaccineID（vaccinations 経由）・
// CountChildrenByParentID を対象とする。
//
// makeVaccineMaster / makeVaccinationRecord は vaccination_master_preload_clinic_isolation_test.go
// に定義済みのものを再利用する（同一パッケージ内での重複宣言を避けるため）。
// makeTestOwner は prescription_repository_test.go の repotest ラッパー、makeSpeciesAndPet は
// diagnosis_name_repository_test.go の共有ローカルコピーを再利用する。

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

// setupVaccineRepositoryTestDB は vaccines テーブルと、CountUsageByVaccineID が参照する
// vaccinations / owners / pets / animal_species を用意する。
func setupVaccineRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Vaccine{}, &model.AnimalSpecies{}, &model.Owner{}, &model.Pet{}, &model.Vaccination{},
	))
	db.Exec("TRUNCATE TABLE vaccinations CASCADE")
	db.Exec("TRUNCATE TABLE vaccines CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE owners CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func TestVaccineRepository_Create_And_FindByID(t *testing.T) {
	db := setupVaccineRepositoryTestDB(t)
	repo := NewVaccineRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	t.Run("happy path: Create してから FindByID で取得できる", func(t *testing.T) {
		price := int64(3000)
		species := model.VaccineSpeciesDog
		v := &model.Vaccine{ClinicID: clinicID, Name: "混合ワクチン", Price: &price, Species: &species}
		require.NoError(t, repo.Create(ctx, v))
		require.NotZero(t, v.ID)

		got, err := repo.FindByID(ctx, clinicID, v.ID)
		require.NoError(t, err)
		assert.Equal(t, "混合ワクチン", got.Name)
		require.NotNil(t, got.Species)
		assert.Equal(t, model.VaccineSpeciesDog, *got.Species)
	})

	t.Run("存在しない ID は NotFound エラー", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicID, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("別クリニックからは FindByID できない（clinic_id 隔離）", func(t *testing.T) {
		v := makeVaccineMaster(t, db, clinicID, "医院1限定ワクチン")
		_, err := repo.FindByID(ctx, uint64(999), v.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックからは NotFound であるべき: %v", err)
	})
}

func TestVaccineRepository_FindAll_SpeciesFilterAndClinicIsolation(t *testing.T) {
	db := setupVaccineRepositoryTestDB(t)
	repo := NewVaccineRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	dog := model.VaccineSpeciesDog
	cat := model.VaccineSpeciesCat
	require.NoError(t, db.WithContext(ctx).Create(&model.Vaccine{ClinicID: clinicA, Name: "犬用ワクチン1", Species: &dog}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Vaccine{ClinicID: clinicA, Name: "犬用ワクチン2", Species: &dog}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Vaccine{ClinicID: clinicA, Name: "猫用ワクチン", Species: &cat}).Error)
	makeVaccineMaster(t, db, clinicB, "他院ワクチン")

	t.Run("clinicA 全件は3件", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA, nil)
		require.NoError(t, err)
		assert.Len(t, got, 3)
	})

	t.Run("species フィルタで犬用のみ2件", func(t *testing.T) {
		dogStr := "dog"
		got, err := repo.FindAll(ctx, clinicA, &dogStr)
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("clinicB は自院の1件のみ（混入なし）", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB, nil)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "他院ワクチン", got[0].Name)
	})
}

func TestVaccineRepository_FindAll_ExcludesSoftDeleted(t *testing.T) {
	db := setupVaccineRepositoryTestDB(t)
	repo := NewVaccineRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	active := makeVaccineMaster(t, db, clinicID, "現役ワクチン")
	deleted := makeVaccineMaster(t, db, clinicID, "削除済みワクチン")
	require.NoError(t, repo.Delete(ctx, clinicID, deleted.ID))

	got, err := repo.FindAll(ctx, clinicID, nil)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, active.ID, got[0].ID)

	var rawCount int64
	db.Unscoped().Model(&model.Vaccine{}).Where("id = ?", deleted.ID).Count(&rawCount)
	assert.Equal(t, int64(1), rawCount, "ソフトデリートされた行はDBにまだ存在する")
}

func TestVaccineRepository_Update(t *testing.T) {
	db := setupVaccineRepositoryTestDB(t)
	repo := NewVaccineRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	v := makeVaccineMaster(t, db, clinicA, "更新前ワクチン")

	t.Run("同一クリニックでは Update が反映される", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, v.ID, map[string]any{"name": "更新後ワクチン"})
		require.NoError(t, err)
		assert.Equal(t, "更新後ワクチン", got.Name)
	})

	t.Run("別クリニックからの Update は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, v.ID, map[string]any{"name": "改ざん試行"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, v.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後ワクチン", got.Name, "別クリニックからの Update で名称が変わってはならない")
	})

	t.Run("存在しない ID の Update は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"name": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestVaccineRepository_Delete(t *testing.T) {
	db := setupVaccineRepositoryTestDB(t)
	repo := NewVaccineRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	v := makeVaccineMaster(t, db, clinicA, "削除対象ワクチン")

	t.Run("別クリニックからの Delete は NotFound で行が残る", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, v.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, v.ID)
		require.NoError(t, err)
		assert.Equal(t, v.ID, got.ID, "別クリニックからの Delete で行が消えてはならない")
	})

	t.Run("同一クリニックでは Delete が成功し以後 FindByID は NotFound", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, v.ID))
		_, err := repo.FindByID(ctx, clinicA, v.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の Delete は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestVaccineRepository_Reorder(t *testing.T) {
	db := setupVaccineRepositoryTestDB(t)
	repo := NewVaccineRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	v1 := makeVaccineMaster(t, db, clinicA, "ワクチンX")
	v2 := makeVaccineMaster(t, db, clinicA, "ワクチンY")
	v3 := makeVaccineMaster(t, db, clinicA, "ワクチンZ")

	t.Run("並び順が指定順に更新される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{v3.ID, v1.ID, v2.ID}))

		got, err := repo.FindAll(ctx, clinicA, nil)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, v3.ID, got[0].ID)
		assert.Equal(t, v1.ID, got[1].ID)
		assert.Equal(t, v2.ID, got[2].ID)
	})

	t.Run("別クリニックの ID を含む Reorder はエラーで中断する", func(t *testing.T) {
		other := makeVaccineMaster(t, db, clinicB, "他院ワクチン")
		err := repo.Reorder(ctx, clinicA, []uint64{v1.ID, other.ID})
		require.Error(t, err, "clinicA スコープに存在しない ID を含む Reorder は失敗すべき")
	})
}

func TestVaccineRepository_CountUsageByVaccineID(t *testing.T) {
	db := setupVaccineRepositoryTestDB(t)
	repo := NewVaccineRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	vaccine := makeVaccineMaster(t, db, clinicA, "使用中ワクチン")
	owner := makeTestOwner(t, db, clinicA, "飼主A")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "タマ")
	makeVaccinationRecord(t, db, clinicA, pet.ID, vaccine.ID)

	t.Run("同一クリニックでは1件カウントされる", func(t *testing.T) {
		count, err := repo.CountUsageByVaccineID(ctx, clinicA, vaccine.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックからは0件（clinic_id 隔離）", func(t *testing.T) {
		count, err := repo.CountUsageByVaccineID(ctx, clinicB, vaccine.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("未使用ワクチンは0件", func(t *testing.T) {
		unused := makeVaccineMaster(t, db, clinicA, "未使用ワクチン")
		count, err := repo.CountUsageByVaccineID(ctx, clinicA, unused.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリート済み接種記録は数えない（P2）", func(t *testing.T) {
		vaccine2 := makeVaccineMaster(t, db, clinicA, "削除記録用ワクチン")
		rec := makeVaccinationRecord(t, db, clinicA, pet.ID, vaccine2.ID)
		require.NoError(t, db.WithContext(ctx).Delete(rec).Error)

		count, err := repo.CountUsageByVaccineID(ctx, clinicA, vaccine2.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "ソフトデリート済みの接種記録は使用件数に含まれない")
	})
}

// TestVaccineRepository_CountChildrenByParentID は BUG-390 の子ワクチンカウントを検証する。
func TestVaccineRepository_CountChildrenByParentID(t *testing.T) {
	db := setupVaccineRepositoryTestDB(t)
	repo := NewVaccineRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	parent := makeVaccineMaster(t, db, clinicA, "親ワクチン")
	child1 := &model.Vaccine{ClinicID: clinicA, Name: "子ワクチン1", ParentID: &parent.ID}
	child2 := &model.Vaccine{ClinicID: clinicA, Name: "子ワクチン2", ParentID: &parent.ID}
	require.NoError(t, db.WithContext(ctx).Create(child1).Error)
	require.NoError(t, db.WithContext(ctx).Create(child2).Error)

	deletedChild := &model.Vaccine{ClinicID: clinicA, Name: "削除済み子ワクチン", ParentID: &parent.ID}
	require.NoError(t, db.WithContext(ctx).Create(deletedChild).Error)
	require.NoError(t, repo.Delete(ctx, clinicA, deletedChild.ID))

	t.Run("clinicA では2件（削除済み除外）", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicA, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("clinicB からは0件（clinic_id 隔離）", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicB, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("子を持たないワクチンは0件", func(t *testing.T) {
		leaf := makeVaccineMaster(t, db, clinicA, "子なしワクチン")
		count, err := repo.CountChildrenByParentID(ctx, clinicA, leaf.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
