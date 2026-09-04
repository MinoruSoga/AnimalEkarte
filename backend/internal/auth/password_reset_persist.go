package auth

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *passwordResetService) persistPasswordResetToken(
	txCtx context.Context,
	accountID uint64,
	resetToken *model.PasswordResetToken,
) (suppressed bool, err error) {
	lockedAccount, lockErr := s.accountRepo.FindByIDForUpdate(txCtx, accountID)
	if lockErr != nil {
		return false, apperrors.Wrap(lockErr, "failed to lock password reset account")
	}
	if lockedAccount == nil || lockedAccount.ID != accountID {
		return false, apperrors.WrapUnauthorized("password reset account is unavailable")
	}
	latestToken, latestErr := s.tokenRepo.FindLatestByAccountIDForUpdate(txCtx, accountID)
	if latestErr != nil && !apperrors.IsNotFound(latestErr) {
		return false, apperrors.Wrap(
			latestErr,
			"failed to inspect existing password reset token",
		)
	}
	now := s.currentTime()
	if latestErr == nil && activeRecentResetToken(latestToken, now) {
		return true, nil
	}
	resetToken.CreatedAt = nextAccountSessionEpoch(lockedAccount.UpdatedAt, now)
	if deleteErr := s.tokenRepo.DeleteByAccountID(txCtx, accountID); deleteErr != nil {
		return false, apperrors.Wrap(deleteErr, "failed to clean up existing tokens")
	}
	if createErr := s.tokenRepo.Create(txCtx, resetToken); createErr != nil {
		return false, apperrors.Wrap(createErr, "failed to create reset token")
	}
	return false, nil
}
