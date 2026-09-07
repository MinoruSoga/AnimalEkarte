package medicalrecord

// medicine_repository_test.go — MedicineRepository の統合テスト（実 Postgres テスト DB）。
//
// medicine_dose_param_clinic_isolation_test.go が MedicineDoseParamRepository の clinic_id 隔離を
// 別途カバーしているため、本ファイルは MedicineRepository（薬剤マスタ本体）の happy path・
// ページネーション・ソフトデリート除外・not-found・clinic_id 隔離・Reorder を対象とする。

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

// setupMedicineRepositoryTestDB は medicines テーブルを用意する。
// setupTestDB は medicine_calculation_type 等の ENUM のみ作成するため、テーブル自体は AutoMigrate で用意する。
func setupMedicineRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Medicine{}))
	db.Exec("TRUNCATE TABLE medicines CASCADE")
	return db
}

// makeMedicine はテスト用の Medicine を作成して返す（全 NOT NULL 列は default 依存で問題ない）。
func makeMedicine(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Medicine {
	t.Helper()
	m := &model.Medicine{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(m).Error)
	return m
}

func TestMedicineRepository_Create_And_FindByID(t *testing.T) {
	db := setupMedicineRepositoryTestDB(t)
	repo := NewMedicineRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	t.Run("happy path: Create してから FindByID で取得できる", func(t *testing.T) {
		price := int64(1500)
		m := &model.Medicine{ClinicID: clinicID, Name: "アモキシシリン", Price: &price}
		require.NoError(t, repo.Create(ctx, m))
		require.NotZero(t, m.ID)

		got, err := repo.FindByID(ctx, clinicID, m.ID)
		require.NoError(t, err)
		assert.Equal(t, "アモキシシリン", got.Name)
		require.NotNil(t, got.Price)
		assert.Equal(t, int64(1500), *got.Price)
	})

	t.Run("存在しない ID は NotFound エラー", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicID, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("別クリニックからは FindByID できない（clinic_id 隔離）", func(t *testing.T) {
		m := makeMedicine(t, db, clinicID, "医院1限定薬")
		_, err := repo.FindByID(ctx, uint64(999), m.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックからは NotFound であるべき: %v", err)
	})
}

func TestMedicineRepository_FindAll_PaginationAndClinicIsolation(t *testing.T) {
	db := setupMedicineRepositoryTestDB(t)
	repo := NewMedicineRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	makeMedicine(t, db, clinicA, "薬A-1")
	makeMedicine(t, db, clinicA, "薬A-2")
	makeMedicine(t, db, clinicA, "薬A-3")
	makeMedicine(t, db, clinicB, "薬B-1")

	t.Run("clinicA は自院の3件のみ", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, got, 3)
	})

	t.Run("clinicB は自院の1件のみ（混入なし）", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicB, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, "薬B-1", got[0].Name)
	})

	t.Run("ページネーション: page/limit が正しく効く", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, 1, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total, "total は全件数")
		assert.Len(t, got, 2, "page1 limit2 → 2件")

		got2, total2, err := repo.FindAll(ctx, clinicA, 2, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total2)
		assert.Len(t, got2, 1, "page2 limit2 → 残り1件")
	})
}

func TestMedicineRepository_FindAll_ExcludesSoftDeleted(t *testing.T) {
	db := setupMedicineRepositoryTestDB(t)
	repo := NewMedicineRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	active := makeMedicine(t, db, clinicID, "現役薬")
	deleted := makeMedicine(t, db, clinicID, "削除済み薬")
	require.NoError(t, repo.Delete(ctx, clinicID, deleted.ID))

	got, total, err := repo.FindAll(ctx, clinicID, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total, "ソフトデリート済みは件数に含まれない")
	require.Len(t, got, 1)
	assert.Equal(t, active.ID, got[0].ID)

	// 行自体は DB に残っている（ソフトデリート = deleted_at セットのみ）
	var rawCount int64
	db.Unscoped().Model(&model.Medicine{}).Where("id = ?", deleted.ID).Count(&rawCount)
	assert.Equal(t, int64(1), rawCount, "ソフトデリートされた行はDBにまだ存在する")
}

func TestMedicineRepository_CountChildrenByParentID(t *testing.T) {
	db := setupMedicineRepositoryTestDB(t)
	repo := NewMedicineRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	parent := makeMedicine(t, db, clinicA, "親薬剤")
	child1 := &model.Medicine{ClinicID: clinicA, Name: "子薬剤1", ParentID: &parent.ID}
	child2 := &model.Medicine{ClinicID: clinicA, Name: "子薬剤2", ParentID: &parent.ID}
	require.NoError(t, db.WithContext(ctx).Create(child1).Error)
	require.NoError(t, db.WithContext(ctx).Create(child2).Error)

	// ソフトデリート済みの子は数えない
	deletedChild := &model.Medicine{ClinicID: clinicA, Name: "削除済み子薬剤", ParentID: &parent.ID}
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

	t.Run("子を持たない薬剤は0件", func(t *testing.T) {
		leaf := makeMedicine(t, db, clinicA, "子なし薬剤")
		count, err := repo.CountChildrenByParentID(ctx, clinicA, leaf.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestMedicineRepository_Update(t *testing.T) {
	db := setupMedicineRepositoryTestDB(t)
	repo := NewMedicineRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	m := makeMedicine(t, db, clinicA, "更新前薬剤")

	t.Run("同一クリニックでは Update が反映される", func(t *testing.T) {
		name := "更新後薬剤"
		got, err := repo.Update(ctx, clinicA, m.ID, UpdateMedicineInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "更新後薬剤", got.Name)
	})

	t.Run("別クリニックからの Update は NotFound", func(t *testing.T) {
		name := "改ざん試行"
		_, err := repo.Update(ctx, clinicB, m.ID, UpdateMedicineInput{Name: &name})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)

		// 改ざんされていないことを確認
		got, err := repo.FindByID(ctx, clinicA, m.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後薬剤", got.Name, "別クリニックからの Update で名称が変わってはならない")
	})

	t.Run("存在しない ID の Update は NotFound", func(t *testing.T) {
		name := "x"
		_, err := repo.Update(ctx, clinicA, 999999, UpdateMedicineInput{Name: &name})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestMedicineRepository_Delete(t *testing.T) {
	db := setupMedicineRepositoryTestDB(t)
	repo := NewMedicineRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	m := makeMedicine(t, db, clinicA, "削除対象薬剤")

	t.Run("別クリニックからの Delete は NotFound で行が残る", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, m.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, m.ID)
		require.NoError(t, err)
		assert.Equal(t, m.ID, got.ID, "別クリニックからの Delete で行が消えてはならない")
	})

	t.Run("同一クリニックでは Delete が成功し以後 FindByID は NotFound", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, m.ID))
		_, err := repo.FindByID(ctx, clinicA, m.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の Delete は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestMedicineRepository_Reorder(t *testing.T) {
	db := setupMedicineRepositoryTestDB(t)
	repo := NewMedicineRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	m1 := makeMedicine(t, db, clinicA, "薬X")
	m2 := makeMedicine(t, db, clinicA, "薬Y")
	m3 := makeMedicine(t, db, clinicA, "薬Z")

	t.Run("並び順が指定順に更新される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{m3.ID, m1.ID, m2.ID}))

		got, _, err := repo.FindAll(ctx, clinicA, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, m3.ID, got[0].ID)
		assert.Equal(t, m1.ID, got[1].ID)
		assert.Equal(t, m2.ID, got[2].ID)
	})

	t.Run("別クリニックの ID を含む Reorder はエラーで中断する", func(t *testing.T) {
		mOther := makeMedicine(t, db, clinicB, "他院薬")
		err := repo.Reorder(ctx, clinicA, []uint64{m1.ID, mOther.ID})
		require.Error(t, err, "clinicA スコープに存在しない ID を含む Reorder は失敗すべき")
	})
}

// TestMedicineRepository_CountUsageByMedicineID_TreatmentUsage は treatments 経由の使用件数を検証する。
func TestMedicineRepository_CountUsageByMedicineID_TreatmentUsage(t *testing.T) {
	db := setupMedicineRepositoryTestDB(t)
	repo := NewMedicineRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.MedicalRecord{}, &model.Treatment{}))

	medA := makeMedicine(t, db, clinicA, "使用中薬剤")

	mr := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "R-0001", Date: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
	require.NoError(t, db.WithContext(ctx).Create(mr).Error)

	treatment := &model.Treatment{MedicalRecordID: mr.ID, ItemType: model.TreatmentItemTypeMedicine, MedicineID: &medA.ID}
	require.NoError(t, db.WithContext(ctx).Create(treatment).Error)

	t.Run("同一クリニックの treatment 使用が1件カウントされる", func(t *testing.T) {
		count, err := repo.CountUsageByMedicineID(ctx, clinicA, medA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックからは0件（clinic_id 隔離・JOIN スコープ）", func(t *testing.T) {
		count, err := repo.CountUsageByMedicineID(ctx, clinicB, medA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
