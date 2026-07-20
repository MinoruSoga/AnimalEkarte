package medicalrecord

// clinical_plan_repository_optimistic_lock_test.go — BUG-416③: clinical_plan の楽観的ロック
// （version 列 + expectedVersion 述語）の統合テスト。移動元 internal/repository（BE9-2D sub-batch④a）。
//
// medical_record_repository.go の Update（BE-refactor.md X-10）と同型の disambiguation を検証する:
//   (a) expectedVersion 一致 → 更新成功、version は +1 される
//   (b) expectedVersion 不一致（親は draft のまま）→ Conflict「他のユーザーが...」
//   (c) expectedVersion == nil → 従来どおり無条件更新（照合スキップ、後方互換）
//   (d) 親カルテが確定済み（expectedVersion は一致していても）→ 従来の「確定済み...」Conflict
//       （(b) の新しい disambiguation 分岐が既存の確定済みガードを退行させていないことの証明）
//
// setupClinicalPlanTestDB / makeClinicalPlanMedicalRecord は clinical_plan_repository_test.go の
// ものを再利用する（重複定義しない）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestClinicalPlanRepository_Update_OptimisticLock(t *testing.T) {
	db := setupClinicalPlanTestDB(t)
	repo := NewClinicalPlanRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	t.Run("expectedVersion一致で更新成功しversionが+1される", func(t *testing.T) {
		mr := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-LOCK-001")
		plan := &model.ClinicalPlan{MedicalRecordID: mr.ID, PhysicalExam: "更新前"}
		require.NoError(t, repo.Create(ctx, plan))
		require.Equal(t, 1, plan.Version, "新規作成時の初期バージョンは1であるべき")

		expected := 1
		err := repo.Update(ctx, clinicA, plan.ID, map[string]any{
			"physical_exam": "更新後",
			"version":       2,
		}, &expected)
		require.NoError(t, err)

		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後", got.PhysicalExam)
		assert.Equal(t, 2, got.Version, "更新成功時はversionが+1されているべき")
	})

	t.Run("expectedVersionが古い(stale)場合はConflictで新しいメッセージを返す", func(t *testing.T) {
		mr := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-LOCK-002")
		plan := &model.ClinicalPlan{MedicalRecordID: mr.ID, PhysicalExam: "更新前"}
		require.NoError(t, repo.Create(ctx, plan))

		// 先に一度更新してversionを2に進める（他ユーザーの書込みをシミュレート）
		current := 1
		require.NoError(t, repo.Update(ctx, clinicA, plan.ID, map[string]any{
			"physical_exam": "他ユーザーによる更新",
			"version":       2,
		}, &current))

		// stale な expectedVersion（1のまま）で更新を試みる
		stale := 1
		err := repo.Update(ctx, clinicA, plan.ID, map[string]any{
			"physical_exam": "競合更新",
			"version":       2,
		}, &stale)

		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "version不一致はConflictであるべき: %v", err)
		assert.Contains(t, err.Error(), "他のユーザーがこの所見・診断を変更しました", "version不一致は専用メッセージであるべき")

		// 競合更新が反映されていないことを確認
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		assert.Equal(t, "他ユーザーによる更新", got.PhysicalExam, "stale版の更新は反映されてはならない")
		assert.Equal(t, 2, got.Version)
	})

	t.Run("expectedVersionがnilの場合は実際のversionを問わず無条件更新される(後方互換)", func(t *testing.T) {
		mr := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-LOCK-003")
		plan := &model.ClinicalPlan{MedicalRecordID: mr.ID, PhysicalExam: "更新前"}
		require.NoError(t, repo.Create(ctx, plan))

		// version を明示的に進めておく
		v1 := 1
		require.NoError(t, repo.Update(ctx, clinicA, plan.ID, map[string]any{
			"physical_exam": "中間更新",
			"version":       2,
		}, &v1))

		// expectedVersion=nil（medical_record_subrecords.go の best-effort 呼び出し等）は
		// 実際のversion(2)と無関係に成功するべき
		err := repo.Update(ctx, clinicA, plan.ID, map[string]any{
			"physical_exam": "照合スキップ更新",
		}, nil)
		require.NoError(t, err)

		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		assert.Equal(t, "照合スキップ更新", got.PhysicalExam)
	})

	t.Run("親カルテが確定済みの場合はexpectedVersion一致でも従来のConflictメッセージのまま", func(t *testing.T) {
		mrFinalized := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-LOCK-004")
		finalizedPlan := &model.ClinicalPlan{MedicalRecordID: mrFinalized.ID, PhysicalExam: "確定前"}
		require.NoError(t, repo.Create(ctx, finalizedPlan))
		require.NoError(t, db.WithContext(ctx).Model(&model.MedicalRecord{}).
			Where("id = ?", mrFinalized.ID).Update("status", model.MedicalRecordStatusFinalized).Error)

		// expectedVersion はDB上の実際の値と一致させる（バージョン不一致が理由でないことを担保）
		matching := finalizedPlan.Version
		err := repo.Update(ctx, clinicA, finalizedPlan.ID, map[string]any{
			"physical_exam": "確定後の書込",
			"version":       matching + 1,
		}, &matching)

		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "確定済みカルテへの更新はConflictであるべき: %v", err)
		assert.Contains(t, err.Error(), "確定済みカルテの所見・診断は編集できません",
			"親確定済みの場合はversion不一致用の新メッセージではなく従来のメッセージを返すべき（回帰防止）")
	})
}
