package medicalrecord

// hospitalization_plan_repository_test.go — HospitalizationPlanRepository の統合テスト。
// 実 Postgres テスト DB (setupTestDB) に対して実行する。

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

// setupHospitalizationPlanRepoTestDB は hospitalization_plans と、使用状況集計で必要になる
// hospitalizations/care_plan_items 一式を整備する。Cage/Medicine/Procedure は Hospitalization /
// CarePlanItem の belongsTo 関連先として AutoMigrate 対象に含める
// （master_preload_clinic_isolation_test.go と同じ組み合わせ）。
func setupHospitalizationPlanRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{}, &model.Cage{},
		&model.Medicine{}, &model.Procedure{},
		&model.HospitalizationPlan{}, &model.Hospitalization{}, &model.CarePlanItem{},
	))
	db.Exec("TRUNCATE TABLE care_plan_items CASCADE")
	db.Exec("TRUNCATE TABLE hospitalizations CASCADE")
	db.Exec("TRUNCATE TABLE hospitalization_plans CASCADE")
	db.Exec("TRUNCATE TABLE medicines CASCADE")
	db.Exec("TRUNCATE TABLE procedures CASCADE")
	db.Exec("TRUNCATE TABLE cages CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func makeHospitalizationPlanFixture(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.HospitalizationPlan {
	t.Helper()
	p := &model.HospitalizationPlan{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(p).Error)
	return p
}

func TestHospitalizationPlanRepository_FindAll(t *testing.T) {
	db := setupHospitalizationPlanRepoTestDB(t)
	repo := NewHospitalizationPlanRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	planA1 := makeHospitalizationPlanFixture(t, db, clinicA, "スタンダードプラン")
	planA2 := makeHospitalizationPlanFixture(t, db, clinicA, "プレミアムプラン")
	_ = makeHospitalizationPlanFixture(t, db, clinicB, "医院Bのプラン")

	t.Run("同一クリニックのプランのみ取得する", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 2)
		ids := []uint64{got[0].ID, got[1].ID}
		assert.ElementsMatch(t, []uint64{planA1.ID, planA2.ID}, ids)
	})

	t.Run("別クリニックからは見えない", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB)
		require.NoError(t, err)
		require.Len(t, got, 1)
	})
}

func TestHospitalizationPlanRepository_FindByID(t *testing.T) {
	db := setupHospitalizationPlanRepoTestDB(t)
	repo := NewHospitalizationPlanRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	planA := makeHospitalizationPlanFixture(t, db, clinicA, "個別取得プラン")

	t.Run("同一クリニックで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, planA.ID)
		require.NoError(t, err)
		assert.Equal(t, planA.ID, got.ID)
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, 99999999)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックからは取得できない", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, planA.ID)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestHospitalizationPlanRepository_Create(t *testing.T) {
	db := setupHospitalizationPlanRepoTestDB(t)
	repo := NewHospitalizationPlanRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	price := int64(5000)
	p := &model.HospitalizationPlan{ClinicID: clinicA, Name: "新規プラン", Price: &price}
	require.NoError(t, repo.Create(ctx, p))
	assert.NotZero(t, p.ID)

	got, err := repo.FindByID(ctx, clinicA, p.ID)
	require.NoError(t, err)
	require.NotNil(t, got.Price)
	assert.Equal(t, price, *got.Price)
}

func TestHospitalizationPlanRepository_Update(t *testing.T) {
	db := setupHospitalizationPlanRepoTestDB(t)
	repo := NewHospitalizationPlanRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	planA := makeHospitalizationPlanFixture(t, db, clinicA, "更新前プラン")

	t.Run("同一クリニックからの更新は成功する", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, planA.ID, map[string]any{"name": "更新後プラン"})
		require.NoError(t, err)
		assert.Equal(t, "更新後プラン", got.Name)
	})

	t.Run("別クリニックからの更新はNotFound", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicB, planA.ID, map[string]any{"name": "不正更新"})
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの更新はNotFound", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, 99999999, map[string]any{"name": "x"})
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestHospitalizationPlanRepository_Delete(t *testing.T) {
	db := setupHospitalizationPlanRepoTestDB(t)
	repo := NewHospitalizationPlanRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	planA := makeHospitalizationPlanFixture(t, db, clinicA, "削除対象プラン")

	t.Run("別クリニックからの削除はNotFoundで実際には削除されない", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, planA.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, planA.ID)
		require.NoError(t, err)
		assert.NotNil(t, got)
	})

	t.Run("同一クリニックからの削除は成功しソフトデリートされる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, planA.ID))

		got, err := repo.FindByID(ctx, clinicA, planA.ID)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))

		var raw model.HospitalizationPlan
		require.NoError(t, db.Unscoped().Where("id = ?", planA.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid)
	})

	t.Run("存在しないIDの削除はNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 99999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestHospitalizationPlanRepository_Reorder(t *testing.T) {
	db := setupHospitalizationPlanRepoTestDB(t)
	repo := NewHospitalizationPlanRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	p1 := makeHospitalizationPlanFixture(t, db, clinicA, "順序1")
	p2 := makeHospitalizationPlanFixture(t, db, clinicA, "順序2")
	p3 := makeHospitalizationPlanFixture(t, db, clinicA, "順序3")

	t.Run("指定順にsort_orderが振り直される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{p3.ID, p1.ID, p2.ID}))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, p3.ID, got[0].ID)
		assert.Equal(t, p1.ID, got[1].ID)
		assert.Equal(t, p2.ID, got[2].ID)
	})

	t.Run("存在しないIDを含む場合はロールバックされる", func(t *testing.T) {
		err := repo.Reorder(ctx, clinicA, []uint64{p1.ID, 99999999, p2.ID})
		require.Error(t, err)

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, p3.ID, got[0].ID, "直前のReorder結果のまま（ロールバック）")
		assert.Equal(t, p1.ID, got[1].ID)
		assert.Equal(t, p2.ID, got[2].ID)
	})
}

// TestHospitalizationPlanRepository_CountUsageByHospitalizationPlanID は
// プランを参照するケアプラン項目数の集計を検証する。
func TestHospitalizationPlanRepository_CountUsageByHospitalizationPlanID(t *testing.T) {
	db := setupHospitalizationPlanRepoTestDB(t)
	repo := NewHospitalizationPlanRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ownerA := makeTestOwner(t, db, clinicA, "プラン使用状況飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "プラン使用状況ポチ")
	planA := makeHospitalizationPlanFixture(t, db, clinicA, "使用状況プラン")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)

	item := &model.CarePlanItem{
		HospitalizationID:     hospA.ID,
		Type:                  model.CarePlanTypeItem,
		Name:                  "プラン利用",
		HospitalizationPlanID: &planA.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(item).Error)

	count, err := repo.CountUsageByHospitalizationPlanID(ctx, clinicA, planA.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
