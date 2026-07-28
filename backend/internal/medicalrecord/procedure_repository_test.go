package medicalrecord

// repository_test.go — Repository の統合テスト（実 Postgres テスト DB）。
// happy path・not-found・clinic_id 隔離・ソフトデリート除外・Update/Delete/Reorder・
// CountChildrenByParentID・CountUsageByProcedureID を対象とする。
//
// 移動元: procedure_repository_test.go（BE8-4 batch21）。makeProcedure ヘルパーは
// 元は treatment_repository_test.go のものを同一パッケージ内で再利用していたが、
// パッケージ分割に伴い本ファイルへローカル複製する。

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

// setupProcedureRepositoryTestDB は procedures テーブルと、CountUsageByProcedureID が
// JOIN する medical_records / treatments を用意する。
func setupProcedureRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Procedure{}, &model.MedicalRecord{}, &model.Treatment{}))
	db.Exec("TRUNCATE TABLE treatments CASCADE")
	db.Exec("TRUNCATE TABLE procedures CASCADE")
	return db
}

func TestRepository_Create_And_FindByID(t *testing.T) {
	db := setupProcedureRepositoryTestDB(t)
	repo := NewProcedureRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	t.Run("happy path: Create してから FindByID で取得できる", func(t *testing.T) {
		price := int64(5000)
		p := &model.Procedure{ClinicID: clinicID, Name: "去勢手術", Price: &price, Anesthesia: model.AnesthesiaTypeGeneral, IsSurgery: true}
		require.NoError(t, repo.Create(ctx, p))
		require.NotZero(t, p.ID)

		got, err := repo.FindByID(ctx, clinicID, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "去勢手術", got.Name)
		assert.True(t, got.IsSurgery)
		assert.Equal(t, model.AnesthesiaTypeGeneral, got.Anesthesia)
	})

	t.Run("存在しない ID は NotFound エラー", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicID, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("別クリニックからは FindByID できない（clinic_id 隔離）", func(t *testing.T) {
		p := makeProcedure(t, db, clinicID, "医院1限定手技", model.AnesthesiaTypeNone, false)
		_, err := repo.FindByID(ctx, uint64(999), p.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックからは NotFound であるべき: %v", err)
	})
}

func TestRepository_FindAll_ClinicIsolation(t *testing.T) {
	db := setupProcedureRepositoryTestDB(t)
	repo := NewProcedureRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	makeProcedure(t, db, clinicA, "手技A-1", model.AnesthesiaTypeNone, false)
	makeProcedure(t, db, clinicA, "手技A-2", model.AnesthesiaTypeLocal, false)
	makeProcedure(t, db, clinicB, "手技B-1", model.AnesthesiaTypeNone, false)

	t.Run("clinicA は自院の2件のみ", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("clinicB は自院の1件のみ（混入なし）", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "手技B-1", got[0].Name)
	})
}

func TestRepository_FindAll_ExcludesSoftDeleted(t *testing.T) {
	db := setupProcedureRepositoryTestDB(t)
	repo := NewProcedureRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	active := makeProcedure(t, db, clinicID, "現役手技", model.AnesthesiaTypeNone, false)
	deleted := makeProcedure(t, db, clinicID, "削除済み手技", model.AnesthesiaTypeNone, false)
	require.NoError(t, repo.Delete(ctx, clinicID, deleted.ID))

	got, err := repo.FindAll(ctx, clinicID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, active.ID, got[0].ID)

	var rawCount int64
	db.Unscoped().Model(&model.Procedure{}).Where("id = ?", deleted.ID).Count(&rawCount)
	assert.Equal(t, int64(1), rawCount, "ソフトデリートされた行はDBにまだ存在する")
}

func TestRepository_Update(t *testing.T) {
	db := setupProcedureRepositoryTestDB(t)
	repo := NewProcedureRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	p := makeProcedure(t, db, clinicA, "更新前手技", model.AnesthesiaTypeNone, false)

	t.Run("同一クリニックでは Update が反映される", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, p.ID, map[string]any{"name": "更新後手技", "is_surgery": true})
		require.NoError(t, err)
		assert.Equal(t, "更新後手技", got.Name)
		assert.True(t, got.IsSurgery)
	})

	t.Run("別クリニックからの Update は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, p.ID, map[string]any{"name": "改ざん試行"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後手技", got.Name, "別クリニックからの Update で名称が変わってはならない")
	})

	t.Run("存在しない ID の Update は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"name": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestRepository_Delete(t *testing.T) {
	db := setupProcedureRepositoryTestDB(t)
	repo := NewProcedureRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	p := makeProcedure(t, db, clinicA, "削除対象手技", model.AnesthesiaTypeNone, false)

	t.Run("別クリニックからの Delete は NotFound で行が残る", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, p.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, p.ID)
		require.NoError(t, err)
		assert.Equal(t, p.ID, got.ID, "別クリニックからの Delete で行が消えてはならない")
	})

	t.Run("同一クリニックでは Delete が成功し以後 FindByID は NotFound", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, p.ID))
		_, err := repo.FindByID(ctx, clinicA, p.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の Delete は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestRepository_Reorder(t *testing.T) {
	db := setupProcedureRepositoryTestDB(t)
	repo := NewProcedureRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	p1 := makeProcedure(t, db, clinicA, "手技X", model.AnesthesiaTypeNone, false)
	p2 := makeProcedure(t, db, clinicA, "手技Y", model.AnesthesiaTypeNone, false)
	p3 := makeProcedure(t, db, clinicA, "手技Z", model.AnesthesiaTypeNone, false)

	t.Run("並び順が指定順に更新される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{p3.ID, p1.ID, p2.ID}))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, p3.ID, got[0].ID)
		assert.Equal(t, p1.ID, got[1].ID)
		assert.Equal(t, p2.ID, got[2].ID)
	})

	t.Run("別クリニックの ID を含む Reorder はエラーで中断する", func(t *testing.T) {
		other := makeProcedure(t, db, clinicB, "他院手技", model.AnesthesiaTypeNone, false)
		err := repo.Reorder(ctx, clinicA, []uint64{p1.ID, other.ID})
		require.Error(t, err, "clinicA スコープに存在しない ID を含む Reorder は失敗すべき")
	})
}

// TestRepository_CountChildrenByParentID は BUG-390 の子手技カウントを検証する。
func TestRepository_CountChildrenByParentID(t *testing.T) {
	db := setupProcedureRepositoryTestDB(t)
	repo := NewProcedureRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	parent := makeProcedure(t, db, clinicA, "親手技", model.AnesthesiaTypeNone, false)
	child1 := &model.Procedure{ClinicID: clinicA, Name: "子手技1", ParentID: &parent.ID}
	child2 := &model.Procedure{ClinicID: clinicA, Name: "子手技2", ParentID: &parent.ID}
	require.NoError(t, db.WithContext(ctx).Create(child1).Error)
	require.NoError(t, db.WithContext(ctx).Create(child2).Error)

	// ソフトデリート済みの子は数えない
	deletedChild := &model.Procedure{ClinicID: clinicA, Name: "削除済み子手技", ParentID: &parent.ID}
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

	t.Run("子を持たない手技は0件", func(t *testing.T) {
		leaf := makeProcedure(t, db, clinicA, "子なし手技", model.AnesthesiaTypeNone, false)
		count, err := repo.CountChildrenByParentID(ctx, clinicA, leaf.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

// TestRepository_CountUsageByProcedureID_KnownCarePlanItemsColumnBug は
// CountUsageByProcedureID（BUG-107）の既知の不具合（本テストで顕在化）を検証する。
//
// 既知の不具合: CountUsageByProcedureID の care_plan_items 集計クエリ（repository.go
// 内の第2クエリ）が `care_plan_items.deleted_at IS NULL` を参照するが、care_plan_items テーブルには
// deleted_at 列が存在しない（migrations/001_init.sql の CREATE TABLE care_plan_items 定義・
// model.CarePlanItem 構造体のいずれにも soft-delete 列がない）。medicine_repository.go の
// CountUsageByMedicineID と全く同型の不具合（medicine_repository_test.go の
// TestMedicineRepository_CountUsageByMedicineID_TreatmentUsage を参照）。
// そのため treatments 側の集計が成功していても、本メソッドは呼び出す度にエラーを返す。
func TestRepository_CountUsageByProcedureID_KnownCarePlanItemsColumnBug(t *testing.T) {
	db := setupProcedureRepositoryTestDB(t)
	repo := NewProcedureRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	proc := makeProcedure(t, db, clinicA, "使用中手技", model.AnesthesiaTypeNone, false)

	mr := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "P-0001", Date: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
	require.NoError(t, db.WithContext(ctx).Create(mr).Error)

	treatment := &model.Treatment{MedicalRecordID: mr.ID, ItemType: model.TreatmentItemTypeProcedure, ProcedureID: &proc.ID}
	require.NoError(t, db.WithContext(ctx).Create(treatment).Error)

	_, err := repo.CountUsageByProcedureID(ctx, clinicA, proc.ID)
	require.Error(t, err, "care_plan_items.deleted_at 列が存在しないため既知の不具合でエラーになる")
}
