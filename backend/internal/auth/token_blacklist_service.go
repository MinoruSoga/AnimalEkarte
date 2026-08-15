package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TokenBlacklistService manages revoked refresh-token JTIs and family markers.
type TokenBlacklistService interface {
	RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
	DeleteExpired(ctx context.Context) error
}

type tokenBlacklistService struct {
	repo TokenBlacklistRepository
}

// NewTokenBlacklistService constructs refresh-token revocation management.
func NewTokenBlacklistService(repo TokenBlacklistRepository) TokenBlacklistService {
	return &tokenBlacklistService{repo: repo}
}

func (s *tokenBlacklistService) RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error {
	entry := &model.TokenBlacklist{
		JTI:       jti,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.Create(ctx, entry); err != nil {
		if !apperrors.IsAlreadyExists(err) {
			slog.ErrorContext(ctx, "failed to revoke token identifier")
		}
		return apperrors.Wrap(err, "failed to revoke token")
	}
	return nil
}

func (s *tokenBlacklistService) IsRevoked(ctx context.Context, jti string) (bool, error) {
	revoked, err := s.repo.ExistsByJTI(ctx, jti)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check token blacklist")
		return false, apperrors.Wrap(err, "failed to check token blacklist")
	}
	return revoked, nil
}

func (s *tokenBlacklistService) DeleteExpired(ctx context.Context) error {
	if err := s.repo.DeleteExpired(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to delete expired blacklist entries", "error", err)
		return apperrors.Wrap(err, "failed to delete expired blacklist entries")
	}
	return nil
}
