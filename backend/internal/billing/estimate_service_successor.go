package billing

import (
	"context"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// Estimate successor draft creation (TASK-012).
// Split from estimate_service.go (ARCH-A4-billing S3) — behavior unchanged.

// CreateSuccessor は承認済み/却下済み見積の後継ドラフトを新規作成する（TASK-012 FINAL B）。
// - 原見積は変更しない（unlock 経路は存在しない）
// - 確定カルテに紐付く原見積でも LockDraftMedicalRecord を呼ばず後継を許可する（明示訂正）
// - 監査は fail-closed（auditTx nil または LogEntryTx 失敗で TX 全体ロールバック）
func (s *estimateService) CreateSuccessor(
	ctx context.Context,
	clinicID, originalID uint64,
	input *CreateSuccessorInput,
) (*model.Estimate, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input is required")
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return nil, apperrors.WrapInvalidInput("reason is required")
	}
	if len([]rune(reason)) > 500 {
		return nil, apperrors.WrapInvalidInput("reason must be at most 500 characters")
	}
	if input.ActorID == 0 {
		return nil, apperrors.WrapInvalidInput("actor is required")
	}
	if s.auditTx == nil {
		return nil, apperrors.WrapInternalServerError("estimate audit dependency is required")
	}

	original, err := s.repo.FindByID(ctx, clinicID, originalID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find estimate for successor", "error", err)
		return nil, apperrors.Wrap(err, "failed to find estimate")
	}
	if !isEstimateLocked(original.Status) {
		return nil, apperrors.WrapConflict("承認済みまたは却下済みの見積書のみ後継ドラフトを作成できます")
	}

	title := original.Title
	if input.Title != nil {
		title = *input.Title
	}
	comment := original.Comment
	if input.Comment != nil {
		comment = *input.Comment
	}
	notes := original.Notes
	if input.Notes != nil {
		notes = *input.Notes
	}
	actorID := input.ActorID

	var successor *model.Estimate
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		created, err := s.createSuccessorInTx(txCtx, clinicID, actorID, original, title, comment, notes, reason)
		if err != nil {
			return err
		}
		successor = created
		return nil
	}); err != nil {
		return nil, err
	}
	if successor == nil {
		return nil, apperrors.WrapInternalServerError("estimate successor create returned empty record")
	}

	slog.InfoContext(ctx, "estimate successor created",
		slog.Uint64("original_id", original.ID),
		slog.Uint64("successor_id", successor.ID),
		slog.Uint64("clinic_id", clinicID))
	return successor, nil
}
