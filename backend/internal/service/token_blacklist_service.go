package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// TokenBlacklistService は refresh_token の JTI ブラックリスト管理を定義する。
type TokenBlacklistService interface {
	// RevokeToken は指定 JTI をブラックリストに登録する（ログアウト・強制失効）。
	RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error
	// IsRevoked は指定 JTI が失効済みかを返す。
	// true の場合、呼び出し元は 401 を返すこと。
	IsRevoked(ctx context.Context, jti string) (bool, error)
	// DeleteExpired は期限切れエントリを物理削除する（バッチメンテ用）。
	DeleteExpired(ctx context.Context) error
}

type tokenBlacklistService struct {
	repo repository.TokenBlacklistRepository
}

// NewTokenBlacklistService は TokenBlacklistService の実装を返す。
func NewTokenBlacklistService(repo repository.TokenBlacklistRepository) TokenBlacklistService {
	return &tokenBlacklistService{repo: repo}
}

func (s *tokenBlacklistService) RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error {
	entry := &model.TokenBlacklist{
		JTI:       jti,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.Create(ctx, entry); err != nil {
		slog.ErrorContext(ctx, "failed to revoke token", "jti", jti, "error", err)
		return apperrors.Wrap(err, "failed to revoke token")
	}
	return nil
}

func (s *tokenBlacklistService) IsRevoked(ctx context.Context, jti string) (bool, error) {
	revoked, err := s.repo.ExistsByJTI(ctx, jti)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check token blacklist", "jti", jti, "error", err)
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
