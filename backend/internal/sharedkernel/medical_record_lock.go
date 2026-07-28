package sharedkernel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// MedicalRecordLocker は LockByIDForUpdate 1メソッドの narrow interface（BE-refactor.md E-5）。
// ctx-txKey 方式のリポジトリフィールドと repo-swap 方式の tx-bound リポジトリの両方がこれを満たす。
type MedicalRecordLocker interface {
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
}

// LockDraftMedicalRecord は X-11 不変条件（行ロック + finalized ガード）を一箇所に集約する
// （BE-refactor.md E-5・臨床安全ガード）。LockByIDForUpdate の行ロックで finalize
// （medical_record_repository.Update の draft-only WHERE）と直列化し、確定と同時の子エンティティ
// 書込が確定済みカルテに混入する競合を防ぐ。
// findErrMsg は LockByIDForUpdate 失敗時の slog.ErrorContext / apperrors.Wrap メッセージ、
// conflictMsg は確定済みカルテだった場合の apperrors.WrapConflict メッセージ。
// いずれも呼び出し元ごとの既存文言をそのまま渡す（テストが assert しているため一字も変えない）。
func LockDraftMedicalRecord(ctx context.Context, repo MedicalRecordLocker, clinicID, recordID uint64, findErrMsg, conflictMsg string) error {
	parent, err := repo.LockByIDForUpdate(ctx, clinicID, recordID)
	if err != nil {
		slog.ErrorContext(ctx, findErrMsg, "error", err)
		return apperrors.Wrap(err, findErrMsg)
	}
	// fail-closed: parent が nil の場合はカルテ不在として NotFound を返す（BE-refactor.md A-5）。
	// 旧実装は `parent != nil &&` ガードにより nil を「draft とみなして続行」する fail-open だった。
	if parent == nil {
		return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", recordID))
	}
	if parent.Status == model.MedicalRecordStatusFinalized {
		return apperrors.WrapConflict(conflictMsg)
	}
	return nil
}
