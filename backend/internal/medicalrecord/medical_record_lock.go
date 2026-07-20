package medicalrecord

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// lockDraftMedicalRecord is a documented, behavior-identical duplicate of
// internal/service.lockDraftMedicalRecord (medical_record_lock.go), kept local so this package
// does not import internal/service (ADR-006). It is a pure helper over the consumer-side
// medicalRecordLocker view (see service_deps.go). Follow-up: collapse both copies onto a shared
// package once a second migrated domain needs it.
//
// X-11 不変条件（行ロック + finalized ガード）を一箇所に集約する（BE-refactor.md E-5）。
// LockByIDForUpdate の行ロックで finalize（medical_record_repository.Update の draft-only WHERE）と
// 直列化し、確定と同時の子エンティティ書込が確定済みカルテに混入する競合を防ぐ。
// findErrMsg は LockByIDForUpdate 失敗時の slog.ErrorContext / apperrors.Wrap メッセージ、
// conflictMsg は確定済みカルテだった場合の apperrors.WrapConflict メッセージ。
// いずれも呼び出し元ごとの既存文言をそのまま渡す（テストが assert しているため一字も変えない）。
func lockDraftMedicalRecord(ctx context.Context, repo medicalRecordLocker, clinicID, recordID uint64, findErrMsg, conflictMsg string) error {
	parent, err := repo.LockByIDForUpdate(ctx, clinicID, recordID)
	if err != nil {
		slog.ErrorContext(ctx, findErrMsg, "error", err)
		return apperrors.Wrap(err, findErrMsg)
	}
	// fail-closed: parent が nil の場合はカルテ不在として NotFound を返す（BE-refactor.md A-5）。
	if parent == nil {
		return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", recordID))
	}
	if parent.Status == model.MedicalRecordStatusFinalized {
		return apperrors.WrapConflict(conflictMsg)
	}
	return nil
}
