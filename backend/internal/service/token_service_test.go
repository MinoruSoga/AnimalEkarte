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

		tampered := issued.Token
		if tampered[len(tampered)-1] == 'a' {
			tampered = tampered[:len(tampered)-1] + "b"
		} else {
			tampered = tampered[:len(tampered)-1] + "a"
		}

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
