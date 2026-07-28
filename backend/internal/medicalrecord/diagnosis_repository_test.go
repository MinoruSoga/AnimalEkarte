package medicalrecord

// diagnosis_repository_test.go — DiagnosisTypeRepository/DiagnosisNameRepository の facade 経由テスト。
//
// TestDiagnosisTypeRepository_FindAll/FindByID/Create/Update/Delete/Reorder と
// TestDiagnosisNameRepository_* は internal/medicalrecord/diagnosis_type_repository_test.go /
// diagnosis_name_repository_test.go へ移設済み（BE9-2C roll-up。旧
// repository/diagnosistype・repository/diagnosisname サブパッケージは削除済み）。
//
// TestDiagnosisTypeRepository_CountChildrenByParentID のみ、ソフトデリート検証サブテストが
// NewDiagnosisNameRepository(db) を直接呼ぶため意図的にここへ残置する（accepted deviation —
// このテストのfixtureヘルパー makeDiagnosisTypeMaster/makeDiagnosisNameRec は
// master_preload_clinic_isolation_test.go 側で定義され他ドメインとも共有されているため、
// このテストだけを medicalrecord へ切り出すとfixtureの複製が必要になる。BE9-2C時点で
// DiagnosisType/DiagnosisNameが同一medicalrecordパッケージへ統合されたため、旧来のimport
// cycle制約は解消済みだが、fixture共有による残留を優先し軽微な変更に留める）。
// facade 経由で medicalrecord.DiagnosisTypeRepository / medicalrecord.DiagnosisNameRepository
// の実装をテストしており、カバレッジの欠落はない。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiagnosisTypeRepository_CountChildrenByParentID(t *testing.T) {
	db := setupDiagnosisTypeTestDB(t)
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
