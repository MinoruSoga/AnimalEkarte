package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/authjwt"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	testTokenJWTSecret = "test-secret-for-token-service"
	testAccountEpoch   = int64(1_727_123_456_789_012_345)
)

func signTestJWT(t *testing.T, method jwt.SigningMethod, claims *authjwt.Claims) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	signed, err := token.SignedString([]byte(testTokenJWTSecret))
	require.NoError(t, err)
	return signed
}

func accessClaimsWithSubject(subject string) *authjwt.Claims {
	return &authjwt.Claims{
		UserID:   "42",
		ClinicID: "7",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}

func refreshClaimsForAlg() *authjwt.Claims {
	return &authjwt.Claims{
		UserID:   "42",
		ClinicID: "7",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   refreshTokenSubject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}

func legacyClaims(subject string, lifetime time.Duration) *authjwt.Claims {
	issuedAt := time.Now().Truncate(time.Second)
	return legacyClaimsAt(subject, lifetime, issuedAt)
}

func legacyClaimsAt(
	subject string,
	lifetime time.Duration,
	issuedAt time.Time,
) *authjwt.Claims {
	return &authjwt.Claims{
		UserID:   "42",
		ClinicID: "7",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "legacy-token-jti",
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(lifetime)),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
		},
	}
}

func TestTokenService_IssueAndVerify(t *testing.T) {
	ctx := context.Background()

	t.Run("roundtrip: 発行した refresh token を検証できる", func(t *testing.T) {
		blacklist := &mockTokenBlacklistRepository{}
		svc := NewTokenService(testTokenJWTSecret, NewTokenBlacklistService(blacklist))

		issued, err := svc.IssueRefreshToken(42, "7", true, []uint64{7, 8}, testAccountEpoch)
		require.NoError(t, err)
		require.NotEmpty(t, issued.Token)

		claims, err := svc.VerifyRefreshToken(ctx, issued.Token)
		require.NoError(t, err)
		assert.Equal(t, "42", claims.UserID)
		assert.Equal(t, "7", claims.ClinicID)
		assert.True(t, claims.IsSystemAdmin)
		assert.Equal(t, []uint64{7, 8}, claims.ClinicIDs)
		assert.Equal(t, testAccountEpoch, claims.AccountEpoch)
		assert.Equal(t, refreshTokenSubject, claims.Subject)
		assert.NotEmpty(t, claims.ID)

		access, err := svc.IssueAccessToken(42, "7", true, []uint64{7, 8}, testAccountEpoch)
		require.NoError(t, err)
		parsed := &authjwt.Claims{}
		_, err = jwt.ParseWithClaims(access.Token, parsed, func(_ *jwt.Token) (any, error) {
			return []byte(testTokenJWTSecret), nil
		})
		require.NoError(t, err)
		assert.Equal(t, "42", parsed.UserID)
		assert.Equal(t, testAccountEpoch, parsed.AccountEpoch)
		assert.Empty(t, parsed.Subject)
	})

	t.Run("tamper detect: 改竄した token は拒否する", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)

		issued, err := svc.IssueRefreshToken(1, "1", false, []uint64{1}, testAccountEpoch)
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

		issued, err := svc.IssueRefreshToken(5, "3", false, []uint64{3}, testAccountEpoch)
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

	t.Run("missing JTI: rotationできない refresh token は拒否する", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, NewTokenBlacklistService(&mockTokenBlacklistRepository{}))
		token := signTestJWT(t, jwt.SigningMethodHS256, refreshClaimsForAlg())

		_, err := svc.VerifyRefreshToken(ctx, token)

		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "invalid refresh token claims")
	})

	t.Run("current refresh missing issued-at is rejected", func(t *testing.T) {
		svc := NewTokenService(
			testTokenJWTSecret,
			NewTokenBlacklistService(&mockTokenBlacklistRepository{}),
		)
		claims := refreshClaimsForAlg()
		claims.ID = "current-refresh-jti"
		claims.AccountEpoch = testAccountEpoch
		claims.IssuedAt = nil
		token := signTestJWT(t, jwt.SigningMethodHS256, claims)

		_, err := svc.VerifyRefreshToken(ctx, token)

		require.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
		assert.Contains(t, err.Error(), "invalid refresh token claims")
	})

	t.Run("current refresh over seven-day lifetime is rejected", func(t *testing.T) {
		svc := NewTokenService(
			testTokenJWTSecret,
			NewTokenBlacklistService(&mockTokenBlacklistRepository{}),
		)
		issuedAt := time.Now().Truncate(time.Second)
		claims := legacyClaimsAt(
			refreshTokenSubject,
			refreshTokenTTL+time.Second,
			issuedAt,
		)
		claims.AccountEpoch = testAccountEpoch
		token := signTestJWT(t, jwt.SigningMethodHS256, claims)

		_, err := svc.VerifyRefreshToken(ctx, token)

		require.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
		assert.Contains(t, err.Error(), "invalid refresh token claims")
	})

	t.Run("legacy refresh without account epoch is accepted within seven-day window", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, NewTokenBlacklistService(&mockTokenBlacklistRepository{}))
		claims := legacyClaims(refreshTokenSubject, refreshTokenTTL)
		token := signTestJWT(t, jwt.SigningMethodHS256, claims)

		verified, err := svc.VerifyRefreshToken(ctx, token)

		require.NoError(t, err)
		assert.Zero(t, verified.AccountEpoch)
	})

	t.Run("missing blacklist: dependency failure is not mislabeled as unauthorized", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)
		issued, err := svc.IssueRefreshToken(5, "3", false, []uint64{3}, testAccountEpoch)
		require.NoError(t, err)

		_, err = svc.VerifyRefreshToken(ctx, issued.Token)

		require.Error(t, err)
		assert.False(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "token validation unavailable")
	})

	t.Run("blacklist read error: dependency cause is preserved for 5xx mapping", func(t *testing.T) {
		blacklistUnavailable := errors.New("blacklist unavailable")
		blacklist := NewTokenBlacklistService(&mockTokenBlacklistRepository{
			existsByJTIFn: func(_ context.Context, _ string) (bool, error) {
				return false, blacklistUnavailable
			},
		})
		svc := NewTokenService(testTokenJWTSecret, blacklist)
		issued, err := svc.IssueRefreshToken(5, "3", false, []uint64{3}, testAccountEpoch)
		require.NoError(t, err)

		claims, err := svc.VerifyRefreshToken(ctx, issued.Token)

		assert.Nil(t, claims)
		require.Error(t, err)
		assert.ErrorIs(t, err, blacklistUnavailable)
		assert.False(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "token validation failed")
	})

	t.Run("invalid token type: access token を refresh 検証に流用すると拒否する", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)

		access, err := svc.IssueAccessToken(9, "2", false, []uint64{2}, testAccountEpoch)
		require.NoError(t, err)

		_, err = svc.VerifyRefreshToken(ctx, access.Token)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "invalid token type")
	})

	t.Run("alg reject: HS384 は拒否する", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)
		token := signTestJWT(t, jwt.SigningMethodHS384, refreshClaimsForAlg())

		_, err := svc.VerifyRefreshToken(ctx, token)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "invalid or expired refresh token")
	})
}

func TestTokenService_IssueFailsClosedWhenJTIEntropyIsUnavailable(t *testing.T) {
	entropyFailure := errors.New("entropy unavailable")
	tests := []struct {
		name  string
		read  func([]byte) (int, error)
		issue func(*tokenService) (*IssuedToken, error)
	}{
		{
			name: "reader error",
			read: func([]byte) (int, error) {
				return 0, entropyFailure
			},
			issue: func(service *tokenService) (*IssuedToken, error) {
				return service.IssueAccessToken(1, "1", false, []uint64{1}, testAccountEpoch)
			},
		},
		{
			name: "short read",
			read: func(destination []byte) (int, error) {
				return len(destination) - 1, nil
			},
			issue: func(service *tokenService) (*IssuedToken, error) {
				return service.IssueRefreshToken(1, "1", false, []uint64{1}, testAccountEpoch)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewTokenService(testTokenJWTSecret, nil).(*tokenService)
			service.randomRead = test.read

			issued, err := test.issue(service)

			assert.Nil(t, issued)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "generate jwt identifier")
		})
	}
}

func TestTokenService_RejectsJWTInputOverMaximumSize(t *testing.T) {
	service := NewTokenService(testTokenJWTSecret, nil)
	oversized := strings.Repeat("x", maxJWTBytes+1)

	accessClaims, accessErr := service.VerifyAccessToken(oversized)
	refreshClaims, refreshErr := service.ParseRefreshTokenClaims(oversized)

	assert.Nil(t, accessClaims)
	require.Error(t, accessErr)
	assert.ErrorIs(t, accessErr, apperrors.ErrUnauthorized)
	assert.Nil(t, refreshClaims)
	require.Error(t, refreshErr)
	assert.ErrorIs(t, refreshErr, apperrors.ErrUnauthorized)
}

func TestTokenService_VerifyAccessToken(t *testing.T) {
	t.Run("roundtrip: 発行した access token を検証できる", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)

		issued, err := svc.IssueAccessToken(42, "7", true, []uint64{7, 8}, testAccountEpoch)
		require.NoError(t, err)

		claims, err := svc.VerifyAccessToken(issued.Token)
		require.NoError(t, err)
		assert.Equal(t, "42", claims.UserID)
		assert.Equal(t, "7", claims.ClinicID)
		assert.True(t, claims.IsSystemAdmin)
		assert.Equal(t, []uint64{7, 8}, claims.ClinicIDs)
		assert.Equal(t, testAccountEpoch, claims.AccountEpoch)
		assert.Empty(t, claims.Subject)
		assert.NotEmpty(t, claims.ID)
	})

	t.Run("legacy access without account epoch is accepted within fifteen-minute window", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)
		token := signTestJWT(
			t,
			jwt.SigningMethodHS256,
			legacyClaims("", accessTokenTTL),
		)

		claims, err := svc.VerifyAccessToken(token)

		require.NoError(t, err)
		assert.Zero(t, claims.AccountEpoch)
	})

	t.Run("current access missing registered claims is rejected", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)
		claims := accessClaimsWithSubject("")
		claims.AccountEpoch = testAccountEpoch
		claims.ID = ""
		claims.IssuedAt = nil
		claims.ExpiresAt = nil
		token := signTestJWT(t, jwt.SigningMethodHS256, claims)

		_, err := svc.VerifyAccessToken(token)

		require.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
		assert.Contains(t, err.Error(), "invalid access token claims")
	})

	t.Run("current access over fifteen-minute lifetime is rejected", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)
		issuedAt := time.Now().Truncate(time.Second)
		claims := legacyClaimsAt(
			"",
			accessTokenTTL+time.Second,
			issuedAt,
		)
		claims.AccountEpoch = testAccountEpoch
		token := signTestJWT(t, jwt.SigningMethodHS256, claims)

		_, err := svc.VerifyAccessToken(token)

		require.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
		assert.Contains(t, err.Error(), "invalid access token claims")
	})

	t.Run("legacy access over fifteen-minute window is rejected", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)
		token := signTestJWT(
			t,
			jwt.SigningMethodHS256,
			accessClaimsWithSubject(""),
		)

		_, err := svc.VerifyAccessToken(token)

		require.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
		assert.Contains(t, err.Error(), "invalid access token claims")
	})

	t.Run("tamper detect: 改竄した token は拒否する", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)

		issued, err := svc.IssueAccessToken(1, "1", false, []uint64{1}, testAccountEpoch)
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

		refresh, err := svc.IssueRefreshToken(9, "2", false, []uint64{2}, testAccountEpoch)
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

		issued, err := svc.IssueAccessToken(5, "3", false, []uint64{3}, testAccountEpoch)
		require.NoError(t, err)

		claims, err := svc.VerifyAccessToken(issued.Token)
		require.NoError(t, err)
		assert.Equal(t, "5", claims.UserID)
		assert.False(t, existsCalled, "VerifyAccessToken must not call blacklist")
	})

	t.Run("non-empty subject: 任意の非空 Subject を拒否する", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)
		token := signTestJWT(t, jwt.SigningMethodHS256, accessClaimsWithSubject("user"))

		_, err := svc.VerifyAccessToken(token)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "invalid token type")
	})

	t.Run("alg reject: HS384 は拒否する", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)
		token := signTestJWT(t, jwt.SigningMethodHS384, accessClaimsWithSubject(""))

		_, err := svc.VerifyAccessToken(token)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "invalid or expired token")
	})
}

func TestLegacyTokenWindowValid_ExactTTLBoundaries(t *testing.T) {
	now := time.Unix(1_727_123_456, 0)
	tests := []struct {
		name     string
		lifetime time.Duration
		maximum  time.Duration
		want     bool
	}{
		{
			name:     "access exact fifteen minutes",
			lifetime: accessTokenTTL,
			maximum:  accessTokenTTL,
			want:     true,
		},
		{
			name:     "access one second over",
			lifetime: accessTokenTTL + time.Second,
			maximum:  accessTokenTTL,
			want:     false,
		},
		{
			name:     "refresh exact seven days",
			lifetime: refreshTokenTTL,
			maximum:  refreshTokenTTL,
			want:     true,
		},
		{
			name:     "refresh one second over",
			lifetime: refreshTokenTTL + time.Second,
			maximum:  refreshTokenTTL,
			want:     false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := &authjwt.Claims{RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(test.lifetime)),
			}}

			assert.Equal(
				t,
				test.want,
				legacyTokenWindowValid(claims, test.maximum, now),
			)
		})
	}
}

func TestTokenMatchesAccountEpoch(t *testing.T) {
	issuedAt := time.Unix(1_727_123_456, 0)

	assert.True(t, TokenMatchesAccountEpoch(
		&authjwt.Claims{AccountEpoch: 900},
		900,
	))
	assert.False(t, TokenMatchesAccountEpoch(
		&authjwt.Claims{AccountEpoch: 899},
		900,
	))
	assert.True(t, TokenMatchesAccountEpoch(
		&authjwt.Claims{RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(issuedAt),
		}},
		issuedAt.Add(-time.Second).UnixNano(),
	))
	assert.False(t, TokenMatchesAccountEpoch(
		&authjwt.Claims{RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(issuedAt),
		}},
		issuedAt.Add(500*time.Millisecond).UnixNano(),
	))
	assert.False(t, TokenMatchesAccountEpoch(
		&authjwt.Claims{RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(issuedAt),
		}},
		issuedAt.Add(time.Second).UnixNano(),
	))
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

		issued, err := svc.IssueRefreshToken(5, "3", false, []uint64{3}, testAccountEpoch)
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
		access, err := svc.IssueAccessToken(9, "2", false, []uint64{2}, testAccountEpoch)
		require.NoError(t, err)

		_, err = svc.ParseRefreshTokenClaims(access.Token)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "invalid token type")
	})

	t.Run("tampered: 署名不正で拒否", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)
		issued, err := svc.IssueRefreshToken(1, "1", false, []uint64{1}, testAccountEpoch)
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

	t.Run("alg reject: HS384 は拒否する", func(t *testing.T) {
		svc := NewTokenService(testTokenJWTSecret, nil)
		token := signTestJWT(t, jwt.SigningMethodHS384, refreshClaimsForAlg())

		_, err := svc.ParseRefreshTokenClaims(token)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Contains(t, err.Error(), "invalid or expired refresh token")
	})
}
