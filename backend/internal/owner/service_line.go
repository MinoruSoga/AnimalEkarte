package owner

import (
	"context"
	"regexp"
	"strings"
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

// LinkLineUserID は EMR-504 / EMR-549 C1 で固定 422 にした。owners.line_user_id は DROP 済みで、置換先
// owner_line_binding_states の writer（EMR-469 の link / unlink intent）は未接続。入力検証と存在確認の後、
// 成功を装わず旧列へも書かずに拒否する（link / unlink どちらも同じ）。
func (s *ownerService) LinkLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string, actorUserID *uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find owner")
	}
	if lineUserID != nil {
		// INF-03: reject empty / path-injection-prone IDs before the pending-writer response.
		if err := validateLineUserID(*lineUserID); err != nil {
			return err
		}
	}
	return apperrors.WrapUnprocessableWithCode(
		apperrors.CodeLineBindingPending,
		"LINE User ID の紐付け変更は現在受け付けできません（紐付け機能の準備中です）",
	)
}

// ConfirmLineID は EMR-504 / EMR-549 C1 で固定 422 にした。line_id_confirmed_at/by は binding の
// verified_at/by に置き換わり、その writer（EMR-469 の verify intent）は未接続。成功を装わない。
func (s *ownerService) ConfirmLineID(ctx context.Context, clinicID, id uint64, actorUserID *uint64) (*model.Owner, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}
	return nil, apperrors.WrapUnprocessableWithCode(
		apperrors.CodeLineBindingPending,
		"LINE ID の確認は現在受け付けできません（紐付け機能の準備中です）",
	)
}
