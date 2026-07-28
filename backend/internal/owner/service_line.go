package owner

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *ownerService) LinkLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string, actorUserID *uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find owner")
	}
	// Q22 Guard 1: 同一クリニック内で LINE User ID が重複していないか確認する。
	if lineUserID != nil {
		existing, err := s.repo.FindByLineUserID(ctx, clinicID, *lineUserID)
		if err != nil && !apperrors.IsNotFound(err) {
			slog.ErrorContext(ctx, "failed to check line user id uniqueness", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to check line user id uniqueness")
		}
		if err == nil && existing != nil && existing.ID != id {
			return apperrors.WrapConflict("この LINE User ID はすでに別の飼主に紐付けられています")
		}
	}
	if err := s.repo.UpdateLineUserID(ctx, clinicID, id, lineUserID); err != nil {
		slog.ErrorContext(ctx, "failed to link line user id", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to link line user id")
	}
	// Q22 Guard 3: 監査ログ（best-effort — 失敗してもリンク操作は続行）。
	if s.auditLogger != nil {
		action := "owner.line_id.link"
		if lineUserID == nil {
			action = "owner.line_id.unlink"
		}
		if logErr := s.auditLogger.LogLstepOperation(ctx, clinicID, actorUserID, action, "owner", &id); logErr != nil {
			slog.WarnContext(ctx, "audit log failed for line id link", "error", logErr, "owner_id", id)
		}
	}
	return nil
}
func (s *ownerService) ConfirmLineID(ctx context.Context, clinicID, id uint64, actorUserID *uint64) (*model.Owner, error) {
	owner, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}
	if owner.LineUserID == nil || strings.TrimSpace(*owner.LineUserID) == "" {
		return nil, apperrors.WrapInvalidInput("LINE User ID is not linked")
	}
	now := time.Now()
	// Q22 Guard 2: 確認者（actorUserID）を line_id_confirmed_by に記録する。
	fields := map[string]any{
		colLineIDConfirmedAt: now,
		colLineIDConfirmedBy: actorUserID,
	}
	updated, err := s.updateOwnerAndFind(
		ctx,
		clinicID,
		id,
		fields,
		"failed to confirm line id",
		"failed to confirm line id",
	)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "line id confirmed",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID))
	// Q22 Guard 3: 監査ログ（best-effort）。
	if s.auditLogger != nil {
		if logErr := s.auditLogger.LogLstepOperation(ctx, clinicID, actorUserID, "owner.line_id.confirm", "owner", &id); logErr != nil {
			slog.WarnContext(ctx, "audit log failed for line id confirm", "error", logErr, "owner_id", id)
		}
	}
	return updated, nil
}
