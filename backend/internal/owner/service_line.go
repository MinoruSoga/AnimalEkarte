package owner

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// lineUserIDPattern: LINE user IDs are opaque tokens; allow alnum + _- only, max 64 (INF-03).
var lineUserIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func validateLineUserID(id string) error {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return apperrors.WrapInvalidInput("LINE User ID is required")
	}
	if utf8.RuneCountInString(trimmed) > 64 {
		return apperrors.WrapInvalidInput("LINE User ID is too long")
	}
	if !lineUserIDPattern.MatchString(trimmed) {
		return apperrors.WrapInvalidInput("LINE User ID contains invalid characters")
	}
	if strings.Contains(trimmed, "..") || strings.ContainsAny(trimmed, "?#/") {
		return apperrors.WrapInvalidInput("LINE User ID contains invalid characters")
	}
	return nil
}

func (s *ownerService) LinkLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string, actorUserID *uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find owner")
	}
	// Q22 Guard 1: 同一クリニック内で LINE User ID が重複していないか確認する。
	if lineUserID != nil {
		// INF-03: reject empty / path-injection-prone IDs before persistence or external API use.
		if err := validateLineUserID(*lineUserID); err != nil {
			return err
		}
		existing, err := s.repo.FindByLineUserID(ctx, clinicID, *lineUserID)
		if err != nil && !apperrors.IsNotFound(err) {
			return apperrors.Wrap(err, "failed to check line user id uniqueness")
		}
		if err == nil && existing != nil && existing.ID != id {
			return apperrors.WrapConflict("この LINE User ID はすでに別の飼主に紐付けられています")
		}
	}
	if err := s.repo.UpdateLineUserID(ctx, clinicID, id, lineUserID); err != nil {
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
