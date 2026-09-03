package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// logReplaceDeletionTx is the medicalrecord-local replace-audit tail
// (snapshot → Replace → deletion audit). Actor type comes from
// auditActorTypeFor, which delegates to sharedkernel.AuditActorTypeFor.
// The audit sink is the consumer-side AuditTxLogger + AuditEntry view
// in service_deps.go. The emitted audit record is field-for-field
// identical (tests assert its contents).

// logReplaceDeletionTx は replace系操作（スナップショット→Replace→削除監査）の監査テールを
// 一箇所に集約する（BE-refactor.md E-6）。deletedCount<=0（純粋な新規挿入のみで削除を伴わない置換）の
// 場合は何もせず nil を返す。logMsg/wrapMsg/idFieldName は各呼び出し元の既存文言・ログキーをそのまま渡す
// （テストが監査レコードの内容を field-for-field で assert しているため一字も変えない）。
func logReplaceDeletionTx(txCtx context.Context, auditTx AuditTxLogger, clinicID uint64, actorID *uint64,
	deletedCount int64, action, resource string, resourceID uint64,
	oldVal, newVal any, metadata map[string]any, logMsg, wrapMsg, idFieldName string) error {
	if deletedCount <= 0 {
		return nil
	}
	actorType := auditActorTypeFor(actorID)
	if err := auditTx.LogEntryTx(txCtx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  actorType,
		Action:     action,
		Resource:   resource,
		ResourceID: &resourceID,
		OldValue:   oldVal,
		NewValue:   newVal,
		Metadata:   metadata,
	}); err != nil {
		slog.ErrorContext(txCtx, logMsg, "error", err, idFieldName, resourceID, "clinic_id", clinicID)
		return apperrors.Wrap(err, wrapMsg)
	}
	return nil
}

// auditActorTypeFor は actorID の有無から監査ログの actor 種別を導出する。
func auditActorTypeFor(actorID *uint64) string {
	return sharedkernel.AuditActorTypeFor(actorID)
}
