package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/authjwt"
)

const (
	accessTokenTTL       = 15 * time.Minute
	refreshTokenTTL      = 7 * 24 * time.Hour
	refreshTokenSubject  = "refresh"
	legacyTokenClockSkew = time.Minute
	maxJWTBytes          = 4096
)

// IssuedToken contains a signed token and its expiry.
type IssuedToken struct {
	Token     string
	ExpiresAt time.Time
}

// TokenService issues and verifies access and refresh JWTs.
type TokenService interface {
	IssueAccessToken(staffID uint64, mainClinicID string, isSystemAdmin bool, clinicIDs []uint64, accountEpoch int64) (*IssuedToken, error)
	IssueRefreshToken(staffID uint64, mainClinicID string, isSystemAdmin bool, clinicIDs []uint64, accountEpoch int64) (*IssuedToken, error)
	IssueRefreshTokenInFamily(staffID uint64, mainClinicID string, isSystemAdmin bool, clinicIDs []uint64, accountEpoch int64, familyID string, familyExpiresAt time.Time) (*IssuedToken, error)
	VerifyAccessToken(tokenStr string) (*authjwt.Claims, error)
	VerifyRefreshToken(ctx context.Context, tokenStr string) (*authjwt.Claims, error)
	ParseRefreshTokenClaims(tokenStr string) (*authjwt.Claims, error)
}

type tokenService struct {
	jwtSecret      []byte
	tokenBlacklist TokenBlacklistService
	randomRead     func([]byte) (int, error)
}

// NewTokenService constructs JWT issuing and verification.
func NewTokenService(jwtSecret string, tokenBlacklist TokenBlacklistService) TokenService {
	return &tokenService{
		jwtSecret:      []byte(jwtSecret),
		tokenBlacklist: tokenBlacklist,
		randomRead:     rand.Read,
	}
}

func newTokenJTI(randomRead func([]byte) (int, error)) (string, error) {
	if randomRead == nil {
		return "", errors.New("jwt entropy source is not configured")
	}
	var tokenBytes [16]byte
	readBytes, err := randomRead(tokenBytes[:])
	if err != nil {
		return "", fmt.Errorf("read jwt entropy: %w", err)
	}
	if readBytes != len(tokenBytes) {
		return "", fmt.Errorf(
			"read jwt entropy: %w: got %d bytes",
			io.ErrUnexpectedEOF,
			readBytes,
		)
	}
	return hex.EncodeToString(tokenBytes[:]), nil
}

func (s *tokenService) jwtKeyFunc(token *jwt.Token) (any, error) {
	if token.Method != jwt.SigningMethodHS256 {
		return nil, jwt.ErrSignatureInvalid
	}
	return s.jwtSecret, nil
}

func (s *tokenService) parseClaims(tokenStr string) (*authjwt.Claims, error) {
	if tokenStr == "" || len(tokenStr) > maxJWTBytes {
		return nil, fmt.Errorf("%w: jwt input length is invalid", jwt.ErrTokenMalformed)
	}
	claims := &authjwt.Claims{}
	if _, err := jwt.ParseWithClaims(tokenStr, claims, s.jwtKeyFunc); err != nil {
		return nil, apperrors.Wrap(err, "failed to parse jwt claims")
	}
	return claims, nil
}

func (s *tokenService) IssueAccessToken(
	staffID uint64,
	mainClinicID string,
	isSystemAdmin bool,
	clinicIDs []uint64,
	accountEpoch int64,
) (*IssuedToken, error) {
	if accountEpoch <= 0 {
		return nil, apperrors.WrapInternalServerError("account session epoch is unavailable")
	}
	jti, err := newTokenJTI(s.randomRead)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to generate jwt identifier")
	}
	now := time.Now()
	expiresAt := now.Add(accessTokenTTL)
	claims := &authjwt.Claims{
		UserID:        strconv.FormatUint(staffID, 10),
		ClinicID:      mainClinicID,
		IsSystemAdmin: isSystemAdmin,
		ClinicIDs:     append([]uint64(nil), clinicIDs...),
		AccountEpoch:  accountEpoch,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to sign jwt")
	}
	return &IssuedToken{Token: tokenString, ExpiresAt: expiresAt}, nil
}

func (s *tokenService) IssueRefreshToken(
	staffID uint64,
	mainClinicID string,
	isSystemAdmin bool,
	clinicIDs []uint64,
	accountEpoch int64,
) (*IssuedToken, error) {
	return s.issueRefreshToken(
		staffID,
		mainClinicID,
		isSystemAdmin,
		clinicIDs,
		accountEpoch,
		"",
		time.Time{},
	)
}

func (s *tokenService) IssueRefreshTokenInFamily(
	staffID uint64,
	mainClinicID string,
	isSystemAdmin bool,
	clinicIDs []uint64,
	accountEpoch int64,
	familyID string,
	familyExpiresAt time.Time,
) (*IssuedToken, error) {
	if familyID == "" || len(familyID) > maxRefreshFamilyIDBytes {
		return nil, apperrors.WrapInternalServerError("refresh token family is unavailable")
	}
	if familyExpiresAt.IsZero() {
		return nil, apperrors.WrapInternalServerError("refresh token family expiry is unavailable")
	}
	return s.issueRefreshToken(
		staffID,
		mainClinicID,
		isSystemAdmin,
		clinicIDs,
		accountEpoch,
		familyID,
		familyExpiresAt,
	)
}

func (s *tokenService) issueRefreshToken(
	staffID uint64,
	mainClinicID string,
	isSystemAdmin bool,
	clinicIDs []uint64,
	accountEpoch int64,
	familyID string,
	familyExpiresAt time.Time,
) (*IssuedToken, error) {
	if accountEpoch <= 0 {
		return nil, apperrors.WrapInternalServerError("account session epoch is unavailable")
	}
	jti, err := newTokenJTI(s.randomRead)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to generate jwt identifier")
	}
	isExistingFamily := familyID != ""
	if !isExistingFamily {
		familyID = jti
	}
	now := time.Now()
	expiresAt := now.Add(refreshTokenTTL)
	if isExistingFamily {
		if !familyExpiresAt.After(now) {
			return nil, apperrors.WrapUnauthorized("refresh token family has expired")
		}
		if familyExpiresAt.Before(expiresAt) {
			expiresAt = familyExpiresAt
		}
	}
	claims := &authjwt.Claims{
		UserID:          strconv.FormatUint(staffID, 10),
		ClinicID:        mainClinicID,
		IsSystemAdmin:   isSystemAdmin,
		ClinicIDs:       append([]uint64(nil), clinicIDs...),
		AccountEpoch:    accountEpoch,
		RefreshFamilyID: familyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   refreshTokenSubject,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to sign refresh token")
	}
	return &IssuedToken{Token: tokenString, ExpiresAt: expiresAt}, nil
}

func (s *tokenService) VerifyAccessToken(tokenStr string) (*authjwt.Claims, error) {
	claims, err := s.parseClaims(tokenStr)
	if err != nil {
		return nil, apperrors.WrapUnauthorized("invalid or expired token")
	}
	if claims.Subject != "" {
		return nil, apperrors.WrapUnauthorized("invalid token type")
	}
	if claims.ID == "" ||
		claims.AccountEpoch < 0 ||
		!legacyTokenWindowValid(claims, accessTokenTTL, time.Now()) {
		return nil, apperrors.WrapUnauthorized("invalid access token claims")
	}
	return claims, nil
}

func (s *tokenService) VerifyRefreshToken(
	ctx context.Context,
	tokenStr string,
) (*authjwt.Claims, error) {
	claims, err := s.parseClaims(tokenStr)
	if err != nil {
		return nil, apperrors.WrapUnauthorized("invalid or expired refresh token")
	}
	if claims.Subject != refreshTokenSubject {
		return nil, apperrors.WrapUnauthorized("invalid token type")
	}
	if claims.ID == "" ||
		claims.AccountEpoch < 0 ||
		!legacyTokenWindowValid(claims, refreshTokenTTL, time.Now()) {
		return nil, apperrors.WrapUnauthorized("invalid refresh token claims")
	}
	familyID, validFamily := refreshTokenFamilyID(claims)
	if !validFamily {
		return nil, apperrors.WrapUnauthorized("invalid refresh token claims")
	}
	if s.tokenBlacklist == nil {
		slog.ErrorContext(ctx, "refresh token validation unavailable: token blacklist is not configured")
		return nil, apperrors.WrapInternalServerError("token validation unavailable")
	}

	familyRevoked, err := s.tokenBlacklist.IsRevoked(
		ctx,
		refreshFamilyBlacklistKey(familyID),
	)
	if err != nil {
		slog.ErrorContext(ctx, "refresh token family blacklist check failed")
		return nil, apperrors.Wrap(err, "token validation failed")
	}
	if familyRevoked {
		return nil, apperrors.WrapUnauthorized("refresh token family has been revoked")
	}

	revoked, err := s.tokenBlacklist.IsRevoked(ctx, claims.ID)
	if err != nil {
		slog.ErrorContext(ctx, "token blacklist check failed")
		return nil, apperrors.Wrap(err, "token validation failed")
	}
	if revoked {
		if familyErr := revokeRefreshTokenFamily(
			ctx,
			s.tokenBlacklist,
			familyID,
		); familyErr != nil {
			slog.ErrorContext(ctx, "refresh token reuse family revocation failed")
			return nil, familyErr
		}
		return nil, apperrors.WrapUnauthorized("token has been revoked")
	}
	return claims, nil
}

// legacyTokenWindowValid validates the issuer's exact lifetime and clock
// bounds for both current and transitional pre-account-epoch tokens.
func legacyTokenWindowValid(
	claims *authjwt.Claims,
	maximumLifetime time.Duration,
	now time.Time,
) bool {
	if claims == nil || claims.IssuedAt == nil || claims.ExpiresAt == nil {
		return false
	}
	issuedAt := claims.IssuedAt.Time
	expiresAt := claims.ExpiresAt.Time
	lifetime := expiresAt.Sub(issuedAt)
	return lifetime > 0 &&
		lifetime <= maximumLifetime &&
		!issuedAt.After(now.Add(legacyTokenClockSkew))
}

// TokenMatchesAccountEpoch validates either an exact current account epoch or
// a bounded legacy token issued no earlier than the account's last mutation.
// JWT NumericDate has one-second precision. Legacy comparison is intentionally
// strict so a same-second account mutation fails closed; epoch-bearing tokens
// retain nanosecond equality.
func TokenMatchesAccountEpoch(claims *authjwt.Claims, accountEpoch int64) bool {
	if claims == nil || accountEpoch <= 0 {
		return false
	}
	if claims.AccountEpoch > 0 {
		return claims.AccountEpoch == accountEpoch
	}
	if claims.AccountEpoch < 0 || claims.IssuedAt == nil {
		return false
	}
	return claims.IssuedAt.Unix() > time.Unix(0, accountEpoch).Unix()
}

func (s *tokenService) ParseRefreshTokenClaims(tokenStr string) (*authjwt.Claims, error) {
	claims, err := s.parseClaims(tokenStr)
	if err != nil {
		return nil, apperrors.WrapUnauthorized("invalid or expired refresh token")
	}
	if claims.Subject != refreshTokenSubject {
		return nil, apperrors.WrapUnauthorized("invalid token type")
	}
	return claims, nil
}
