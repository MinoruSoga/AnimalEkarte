package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/authjwt"
	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

const testTokenJWTSecret = "test-secret-for-token-service"

func TestTokenService_IssueAndVerify(t *testing.T) {
	ctx := context.Background()

	t.Run("roundtrip: 発行した refresh token を検証できる", func(t *testing.T) {
		blacklist := &mockTokenBlacklistRepository{}
		svc := NewTokenService(testTokenJWTSecret, NewTokenBlacklistService(blacklist))

		issued, err := svc.IssueRefreshToken(42, "7", true, []uint64{7, 8})
		require.NoError(t, err)
		require.NotEmpty(t, issued.Token)

		claims, err := svc.VerifyRefreshToken(ctx, issued.Token)
		require.NoError(t, err)
		assert.Equal(t, "42", claims.UserID)
		assert.Equal(t, "7", claims.ClinicID)
		assert.True(t, claims.IsSystemAdmin)
		assert.Equal(t, []uint64{7, 8}, claims.ClinicIDs)
		assert.Equal(t, refreshTokenSubject, claims.Subject)
		assert.NotEmpty(t, claims.ID)

		access, err := svc.IssueAccessToken(42, "7", true, []uint64{7, 8})
		require.NoError(t, err)
		parsed := &authjwt.Claims{}
		_, err = jwt.ParseWithClaims(access.Token, parsed, func(_ *jwt.Token) (any, error) {
			return []byte(testTokenJWTSecret), nil
		})
		require.NoError(t, err)
		assert.Equal(t, "42", parsed.UserID)
		assert.Empty(t, parsed.Subject)
	})

	t.Run("tamper detect: 改竄した token は拒否する", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)

		issued, err := svc.IssueRefreshToken(1, "1", false, []uint64{1})
		require.NoError(t, err)

		parts := strings.Split(issued.Token, ".")
		require.Len(t, parts, 3)
		sig := []byte(parts[2])
		if sig[0] == 'A' {
			sig[0] = 'B'
		} else {
			sig[0] = 'A'
		}
		parts[2] = string(sig)
		tampered := strings.Join(parts, ".")

		_, err = svc.VerifyRefreshToken(ctx, tampered)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "invalid or expired refresh token")
	})

	t.Run("blacklist reject: 失効済み JTI の token は拒否する", func(t *testing.T) {
		revokedJTIs := map[string]bool{}
		repo := &mockTokenBlacklistRepository{
			createFn: func(_ context.Context, entry *model.TokenBlacklist) error {
				revokedJTIs[entry.JTI] = true
				return nil
			},
			existsByJTIFn: func(_ context.Context, jti string) (bool, error) {
				return revokedJTIs[jti], nil
			},
		}
		blacklistSvc := NewTokenBlacklistService(repo)
		svc := NewTokenService(testTokenJWTSecret, blacklistSvc)

		issued, err := svc.IssueRefreshToken(5, "3", false, []uint64{3})
		require.NoError(t, err)

		parsed := &authjwt.Claims{}
		_, err = jwt.ParseWithClaims(issued.Token, parsed, func(_ *jwt.Token) (any, error) {
			return []byte(testTokenJWTSecret), nil
		})
		require.NoError(t, err)
		require.NotEmpty(t, parsed.ID)
		require.NotNil(t, parsed.ExpiresAt)

		err = blacklistSvc.RevokeToken(ctx, parsed.ID, parsed.ExpiresAt.Time)
		require.NoError(t, err)

		_, err = svc.VerifyRefreshToken(ctx, issued.Token)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.True(t, strings.Contains(err.Error(), "token has been revoked"))
	})

	t.Run("invalid token type: access token を refresh 検証に流用すると拒否する", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)

		access, err := svc.IssueAccessToken(9, "2", false, []uint64{2})
		require.NoError(t, err)

		_, err = svc.VerifyRefreshToken(ctx, access.Token)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "invalid token type")
	})
}

func TestTokenService_VerifyAccessToken(t *testing.T) {
	t.Run("roundtrip: 発行した access token を検証できる", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)

		issued, err := svc.IssueAccessToken(42, "7", true, []uint64{7, 8})
		require.NoError(t, err)

		claims, err := svc.VerifyAccessToken(issued.Token)
		require.NoError(t, err)
		assert.Equal(t, "42", claims.UserID)
		assert.Equal(t, "7", claims.ClinicID)
		assert.True(t, claims.IsSystemAdmin)
		assert.Equal(t, []uint64{7, 8}, claims.ClinicIDs)
		assert.Empty(t, claims.Subject)
		assert.NotEmpty(t, claims.ID)
	})

	t.Run("tamper detect: 改竄した token は拒否する", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)

		issued, err := svc.IssueAccessToken(1, "1", false, []uint64{1})
		require.NoError(t, err)

		parts := strings.Split(issued.Token, ".")
		require.Len(t, parts, 3)
		sig := []byte(parts[2])
		if sig[0] == 'A' {
			sig[0] = 'B'
		} else {
			sig[0] = 'A'
		}
		parts[2] = string(sig)
		tampered := strings.Join(parts, ".")

		_, err = svc.VerifyAccessToken(tampered)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "invalid or expired token")
	})

	t.Run("invalid token type: refresh token を access 検証に流用すると拒否する", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)

		refresh, err := svc.IssueRefreshToken(9, "2", false, []uint64{2})
		require.NoError(t, err)

		_, err = svc.VerifyAccessToken(refresh.Token)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "invalid token type")
	})

	t.Run("blacklist 非呼び出し: revoked JTI でも access は通る（BLなし）", func(t *testing.T) {
		existsCalled := false
		repo := &mockTokenBlacklistRepository{
			existsByJTIFn: func(_ context.Context, _ string) (bool, error) {
				existsCalled = true
				return true, nil
			},
		}
		blacklistSvc := NewTokenBlacklistService(repo)
		svc := NewTokenService(testTokenJWTSecret, blacklistSvc)

		issued, err := svc.IssueAccessToken(5, "3", false, []uint64{3})
		require.NoError(t, err)

		claims, err := svc.VerifyAccessToken(issued.Token)
		require.NoError(t, err)
		assert.Equal(t, "5", claims.UserID)
		assert.False(t, existsCalled, "VerifyAccessToken must not call blacklist")
	})
}

func TestTokenService_ParseRefreshTokenClaims(t *testing.T) {
	t.Run("valid refresh: claims を返しブラックリストを見ない", func(t *testing.T) {
		revokedJTIs := map[string]bool{}
		repo := &mockTokenBlacklistRepository{
			existsByJTIFn: func(_ context.Context, jti string) (bool, error) {
				return revokedJTIs[jti], nil
			},
		}
		blacklistSvc := NewTokenBlacklistService(repo)
		svc := NewTokenService(testTokenJWTSecret, blacklistSvc)

		issued, err := svc.IssueRefreshToken(5, "3", false, []uint64{3})
		require.NoError(t, err)

		parsed, err := svc.ParseRefreshTokenClaims(issued.Token)
		require.NoError(t, err)
		require.NotEmpty(t, parsed.ID)
		require.NotNil(t, parsed.ExpiresAt)

		// 失効済みでも Parse は成功する（Logout 用・BL照合なし）
		revokedJTIs[parsed.ID] = true
		again, err := svc.ParseRefreshTokenClaims(issued.Token)
		require.NoError(t, err)
		assert.Equal(t, parsed.ID, again.ID)
		assert.Equal(t, refreshTokenSubject, again.Subject)
	})

	t.Run("access token: subject 不一致で拒否", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)
		access, err := svc.IssueAccessToken(9, "2", false, []uint64{2})
		require.NoError(t, err)

		_, err = svc.ParseRefreshTokenClaims(access.Token)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "invalid token type")
	})

	t.Run("tampered: 署名不正で拒否", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)
		issued, err := svc.IssueRefreshToken(1, "1", false, []uint64{1})
		require.NoError(t, err)

		parts := strings.Split(issued.Token, ".")
		require.Len(t, parts, 3)
		sig := []byte(parts[2])
		if sig[0] == 'A' {
			sig[0] = 'B'
		} else {
			sig[0] = 'A'
		}
		parts[2] = string(sig)
		tampered := strings.Join(parts, ".")

		_, err = svc.ParseRefreshTokenClaims(tampered)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
	})
}
