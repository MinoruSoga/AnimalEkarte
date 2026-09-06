package medicalrecord

// diagnosis_type_repository_test.go — DiagnosisTypeRepository の統合テスト。
//
// 移動元: internal/repository/diagnosistype/repository_test.go（BE8-4 batch27）→ BE9-2C roll-up。
//
// makeDiagnosisTypeMaster/makeDiagnosisNameRec は diagnosis_name_repository_test.go からも
// 使う（旧 diagnosistype/diagnosisname 両サブパッケージが import cycle 回避のため独立に複製
// していたのを、BE9-2C で同一 medicalrecord パッケージへ統合したことで重複定義がコンパイル
// エラーになる — ここに一本化し、diagnosis_name_repository_test.go 側の重複定義は削除する）。
//
// TestDiagnosisTypeRepository_CountChildrenByParentID は internal/repository/
// diagnosis_repository_test.go に意図的に残置されている（旧 accepted deviation の経緯は
// そちらのコメント参照）。
//
// 保護する不変条件:
//   - FindAll/FindByID/Update/Delete/Reorder は clinic_id で正しく分離される。
//   - FindByID は Names を clinic_id 述語付きで Preload する。

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

// setupDiagnosisTypeTestDB は diagnosis_types / diagnosis_names 用に DB を整備する。
func setupDiagnosisTypeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.DiagnosisType{}, &model.DiagnosisName{}))
	db.Exec("TRUNCATE TABLE diagnosis_names CASCADE")
	db.Exec("TRUNCATE TABLE diagnosis_types CASCADE")
	return db
}

func makeDiagnosisTypeMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.DiagnosisType {
	t.Helper()
	dt := &model.DiagnosisType{ClinicID: clinicID, Name: name}
	require.NoError(t, db.Create(dt).Error)
	return dt
}

func makeDiagnosisNameRec(t *testing.T, db *gorm.DB, clinicID, typeID uint64, name string) *model.DiagnosisName {
	t.Helper()
	dn := &model.DiagnosisName{ClinicID: clinicID, DiagnosisTypeID: typeID, Name: name}
	require.NoError(t, db.Create(dn).Error)
	return dn
}

func TestDiagnosisTypeRepository_FindAll(t *testing.T) {
	db := setupDiagnosisTypeTestDB(t)
	repo := NewDiagnosisTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	typeA1 := makeDiagnosisTypeMaster(t, db, clinicA, "内科")
	typeA2 := makeDiagnosisTypeMaster(t, db, clinicA, "外科")
	_ = makeDiagnosisTypeMaster(t, db, clinicB, "医院Bの分類")

	t.Run("同一クリニックの分類のみ取得する", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, got, 2)
		ids := []uint64{got[0].ID, got[1].ID}
		assert.ElementsMatch(t, []uint64{typeA1.ID, typeA2.ID}, ids)
	})

	t.Run("limitに応じて取得件数が絞られるがtotalは影響を受けない", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, 1, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, got, 1)
	})

	t.Run("別クリニックからはクリニックAの分類が見えない", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicB, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.NotEqual(t, typeA1.ID, got[0].ID)
		assert.NotEqual(t, typeA2.ID, got[0].ID)
	})
}

func TestDiagnosisTypeRepository_FindByID(t *testing.T) {
	db := setupDiagnosisTypeTestDB(t)
	repo := NewDiagnosisTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "皮膚科")
	nameA := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "アトピー性皮膚炎")

	t.Run("同一クリニックで取得できNamesがPreloadされる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, typeA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Len(t, got.Names, 1)
		assert.Equal(t, nameA.ID, got.Names[0].ID)
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, 99999999)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックからは取得できない", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, typeA.ID)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestDiagnosisTypeRepository_Create(t *testing.T) {
	db := setupDiagnosisTypeTestDB(t)
	repo := NewDiagnosisTypeRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	dt := &model.DiagnosisType{ClinicID: clinicA, Name: "眼科"}
	require.NoError(t, repo.Create(ctx, dt))
	assert.NotZero(t, dt.ID)

	got, err := repo.FindByID(ctx, clinicA, dt.ID)
	require.NoError(t, err)
	assert.Equal(t, "眼科", got.Name)
}

func TestDiagnosisTypeRepository_Update(t *testing.T) {
	db := setupDiagnosisTypeTestDB(t)
	repo := NewDiagnosisTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "更新前")

	t.Run("同一クリニックからの更新は成功する", func(t *testing.T) {
		name := "更新後"
		got, err := repo.Update(ctx, clinicA, typeA.ID, UpdateDiagnosisTypeInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "更新後", got.Name)
	})

	t.Run("別クリニックからの更新はNotFoundで変更されない", func(t *testing.T) {
		name := "不正更新"
		got, err := repo.Update(ctx, clinicB, typeA.ID, UpdateDiagnosisTypeInput{Name: &name})
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		reread, err := repo.FindByID(ctx, clinicA, typeA.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後", reread.Name)
	})

	t.Run("存在しないIDの更新はNotFound", func(t *testing.T) {
		name := "x"
		got, err := repo.Update(ctx, clinicA, 99999999, UpdateDiagnosisTypeInput{Name: &name})
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestDiagnosisTypeRepository_Delete(t *testing.T) {
	db := setupDiagnosisTypeTestDB(t)
	repo := NewDiagnosisTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "削除対象")

	t.Run("別クリニックからの削除はNotFoundで実際には削除されない", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, typeA.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, typeA.ID)
		require.NoError(t, err)
		assert.NotNil(t, got)
	})

	t.Run("同一クリニックからの削除は成功しソフトデリートされる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, typeA.ID))

		got, err := repo.FindByID(ctx, clinicA, typeA.ID)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))

		var raw model.DiagnosisType
		require.NoError(t, db.Unscoped().Where("id = ?", typeA.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "deleted_at がセットされているべき")
	})

	t.Run("存在しないIDの削除はNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 99999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestDiagnosisTypeRepository_Reorder(t *testing.T) {
	db := setupDiagnosisTypeTestDB(t)
	repo := NewDiagnosisTypeRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	t1 := makeDiagnosisTypeMaster(t, db, clinicA, "順序A")
	t2 := makeDiagnosisTypeMaster(t, db, clinicA, "順序B")
	t3 := makeDiagnosisTypeMaster(t, db, clinicA, "順序C")

	t.Run("指定順にsort_orderが振り直される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{t3.ID, t1.ID, t2.ID}))

		got, _, err := repo.FindAll(ctx, clinicA, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, t3.ID, got[0].ID)
		assert.Equal(t, t1.ID, got[1].ID)
		assert.Equal(t, t2.ID, got[2].ID)
	})

	t.Run("存在しないIDを含む場合はエラーになりロールバックされる", func(t *testing.T) {
		err := repo.Reorder(ctx, clinicA, []uint64{t1.ID, 99999999, t2.ID})
		require.Error(t, err)

		got, _, err := repo.FindAll(ctx, clinicA, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, t3.ID, got[0].ID, "直前のReorder結果のまま（ロールバック）")
		assert.Equal(t, t1.ID, got[1].ID)
		assert.Equal(t, t2.ID, got[2].ID)
	})
}
