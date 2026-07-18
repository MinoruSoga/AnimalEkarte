package repository

// diagnosis_repository_test.go — DiagnosisTypeRepository/DiagnosisNameRepository の facade 経由テスト。
//
// TestDiagnosisTypeRepository_FindAll/FindByID/Create/Update/Delete/Reorder は
// repository/diagnosistype/repository_test.go へ移設済み（BE8-4 batch27）。
// TestDiagnosisNameRepository_* は repository/diagnosisname/repository_test.go へ移設済み（BE8-4 batch28）。
//
// TestDiagnosisTypeRepository_CountChildrenByParentID のみ、ソフトデリート検証サブテストが
// NewDiagnosisNameRepository(db) を直接呼ぶため意図的にここへ残置する（accepted deviation —
// diagnosistype パッケージに移すと repository → diagnosistype facade との import cycle になる。
// facade 経由で diagnosistype.Repository / diagnosisname.Repository の実装をテストしており、
// カバレッジの欠落はない）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// setupDiagnosisRepoTestDB は diagnosis_types / diagnosis_names 用に DB を整備する。
func setupDiagnosisRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db, &model.DiagnosisType{}, &model.DiagnosisName{}))
	db.Exec("TRUNCATE TABLE diagnosis_names CASCADE")
	db.Exec("TRUNCATE TABLE diagnosis_types CASCADE")
	return db
}

func TestDiagnosisTypeRepository_CountChildrenByParentID(t *testing.T) {
	db := setupDiagnosisRepoTestDB(t)
	repo := NewDiagnosisTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "分類A")
	_ = makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "名前1")
	nameToDelete := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "名前2")

	t.Run("子が0件", func(t *testing.T) {
		emptyType := makeDiagnosisTypeMaster(t, db, clinicA, "子なし分類")
		count, err := repo.CountChildrenByParentID(ctx, clinicA, emptyType.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("子が複数件", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicA, typeA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("ソフトデリートされた子は除外される", func(t *testing.T) {
		nameRepo := NewDiagnosisNameRepository(db)
		require.NoError(t, nameRepo.Delete(ctx, clinicA, nameToDelete.ID))

		count, err := repo.CountChildrenByParentID(ctx, clinicA, typeA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックからは0件", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicB, typeA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
