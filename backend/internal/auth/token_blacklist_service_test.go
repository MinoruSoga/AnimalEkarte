package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockTokenBlacklistRepository は TokenBlacklistRepository のテスト用モック実装
type mockTokenBlacklistRepository struct {
	createFn        func(ctx context.Context, entry *model.TokenBlacklist) error
	existsByJTIFn   func(ctx context.Context, jti string) (bool, error)
	deleteExpiredFn func(ctx context.Context) error
}

func (m *mockTokenBlacklistRepository) Create(ctx context.Context, entry *model.TokenBlacklist) error {
	if m.createFn != nil {
		return m.createFn(ctx, entry)
	}
	return nil
}

func (m *mockTokenBlacklistRepository) ExistsByJTI(ctx context.Context, jti string) (bool, error) {
	if m.existsByJTIFn != nil {
		return m.existsByJTIFn(ctx, jti)
	}
	return false, nil
}

func (m *mockTokenBlacklistRepository) DeleteExpired(ctx context.Context) error {
	if m.deleteExpiredFn != nil {
		return m.deleteExpiredFn(ctx)
	}
	return nil
}

func TestTokenBlacklistService_RevokeToken(t *testing.T) {
	ctx := context.Background()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	t.Run("正常: JTI をブラックリストに登録する", func(t *testing.T) {
		var captured *model.TokenBlacklist
		repo := &mockTokenBlacklistRepository{
			createFn: func(_ context.Context, entry *model.TokenBlacklist) error {
				captured = entry
				return nil
			},
		}
		svc := NewTokenBlacklistService(repo)

		err := svc.RevokeToken(ctx, "test-jti", expiresAt)

		require.NoError(t, err)
		require.NotNil(t, captured)
		assert.Equal(t, "test-jti", captured.JTI)
		assert.Equal(t, expiresAt, captured.ExpiresAt)
	})

	t.Run("異常: リポジトリエラーはラップして返す", func(t *testing.T) {
		rawErr := errors.New("db error")
		repo := &mockTokenBlacklistRepository{
			createFn: func(_ context.Context, _ *model.TokenBlacklist) error {
				return apperrors.Wrap(rawErr, "token_blacklist")
			},
		}
		svc := NewTokenBlacklistService(repo)

		err := svc.RevokeToken(ctx, "test-jti", expiresAt)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to revoke token")
	})
}

func TestTokenBlacklistService_IsRevoked(t *testing.T) {
	ctx := context.Background()

	t.Run("正常: ブラックリスト登録済みの JTI は true を返す", func(t *testing.T) {
		repo := &mockTokenBlacklistRepository{
			existsByJTIFn: func(_ context.Context, jti string) (bool, error) {
				return jti == "revoked-jti", nil
			},
		}
		svc := NewTokenBlacklistService(repo)

		revoked, err := svc.IsRevoked(ctx, "revoked-jti")

		require.NoError(t, err)
		assert.True(t, revoked)
	})

	t.Run("正常: 未登録の JTI は false を返す", func(t *testing.T) {
		repo := &mockTokenBlacklistRepository{
			existsByJTIFn: func(_ context.Context, _ string) (bool, error) {
				return false, nil
			},
		}
		svc := NewTokenBlacklistService(repo)

		revoked, err := svc.IsRevoked(ctx, "active-jti")

		require.NoError(t, err)
		assert.False(t, revoked)
	})

	t.Run("異常: リポジトリエラーはラップして返す", func(t *testing.T) {
		rawErr := errors.New("db error")
		repo := &mockTokenBlacklistRepository{
			existsByJTIFn: func(_ context.Context, _ string) (bool, error) {
				return false, apperrors.Wrap(rawErr, "token_blacklist")
			},
		}
		svc := NewTokenBlacklistService(repo)

		revoked, err := svc.IsRevoked(ctx, "any-jti")

		require.Error(t, err)
		assert.False(t, revoked)
		assert.Contains(t, err.Error(), "failed to check token blacklist")
	})
}

func TestTokenBlacklistService_DeleteExpired(t *testing.T) {
	ctx := context.Background()

	t.Run("正常: 期限切れエントリを削除する", func(t *testing.T) {
		called := false
		repo := &mockTokenBlacklistRepository{
			deleteExpiredFn: func(_ context.Context) error {
				called = true
				return nil
			},
		}
		svc := NewTokenBlacklistService(repo)

		err := svc.DeleteExpired(ctx)

		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("異常: リポジトリエラーはラップして返す", func(t *testing.T) {
		rawErr := errors.New("db error")
		repo := &mockTokenBlacklistRepository{
			deleteExpiredFn: func(_ context.Context) error {
				return apperrors.Wrap(rawErr, "token_blacklist")
			},
		}
		svc := NewTokenBlacklistService(repo)

		err := svc.DeleteExpired(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete expired blacklist entries")
	})
}
