package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/animal-ekarte/backend/internal/authjwt"
	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

const (
	accessTokenTTL      = 15 * time.Minute
	refreshTokenTTL     = 7 * 24 * time.Hour
	refreshTokenSubject = "refresh"
)

// IssuedToken は JWT 発行結果（トークン文字列と有効期限）。
type IssuedToken struct {
	Token     string
	ExpiresAt time.Time
}

// TokenService は JWT アクセストークン/リフレッシュトークンの発行と検証を定義する。
type TokenService interface {
	IssueAccessToken(staffID uint64, mainClinicID string, isSystemAdmin bool, clinicIDs []uint64) (*IssuedToken, error)
	IssueRefreshToken(staffID uint64, mainClinicID string, isSystemAdmin bool, clinicIDs []uint64) (*IssuedToken, error)
	// VerifyAccessToken は access JWT を HS256 検証する（ブラックリスト照合なし）。
	// Subject が空でないトークンは拒否する（refresh および任意非空 subject の誤用防止）。
	VerifyAccessToken(tokenStr string) (*authjwt.Claims, error)
	VerifyRefreshToken(ctx context.Context, tokenStr string) (*authjwt.Claims, error)
	// ParseRefreshTokenClaims は Logout 用のベストエフォート parse。
	// HS256 + subject=refresh のみ検証し、ブラックリスト照合は行わない（VerifyRefreshToken とは別）。
	ParseRefreshTokenClaims(tokenStr string) (*authjwt.Claims, error)
}

type tokenService struct {
	jwtSecret      []byte
	tokenBlacklist TokenBlacklistService
}

// NewTokenService は TokenService の実装を返す。
// tokenBlacklist は VerifyRefreshToken の JTI 照合に使用する（nil 許容 — 照合スキップ）。
func NewTokenService(jwtSecret string, tokenBlacklist TokenBlacklistService) TokenService {
	return &tokenService{
		jwtSecret:      []byte(jwtSecret),
		tokenBlacklist: tokenBlacklist,
	}
}

func newTokenJti() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(bytes[:])
}

// jwtKeyFunc は HS256 のみ許可する keyFunc（Issue* と一貫）。
func (s *tokenService) jwtKeyFunc(t *jwt.Token) (any, error) {
	if t.Method != jwt.SigningMethodHS256 {
		return nil, jwt.ErrSignatureInvalid
	}
	return s.jwtSecret, nil
}

// parseClaims は JWT を HS256 で検証し claims を返す（subject / blacklist 判定は呼び出し側）。
func (s *tokenService) parseClaims(tokenStr string) (*authjwt.Claims, error) {
	claims := &authjwt.Claims{}
	if _, err := jwt.ParseWithClaims(tokenStr, claims, s.jwtKeyFunc); err != nil {
		return nil, apperrors.Wrap(err, "failed to parse jwt claims")
	}
	return claims, nil
}

func (s *tokenService) IssueAccessToken(staffID uint64, mainClinicID string, isSystemAdmin bool, clinicIDs []uint64) (*IssuedToken, error) {
	expiresAt := time.Now().Add(accessTokenTTL)
	claims := &authjwt.Claims{
		UserID:        strconv.FormatUint(staffID, 10),
		ClinicID:      mainClinicID,
		IsSystemAdmin: isSystemAdmin,
		ClinicIDs:     clinicIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        newTokenJti(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessTokenStr, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to sign jwt")
	}
	return &IssuedToken{Token: accessTokenStr, ExpiresAt: expiresAt}, nil
}

func (s *tokenService) IssueRefreshToken(staffID uint64, mainClinicID string, isSystemAdmin bool, clinicIDs []uint64) (*IssuedToken, error) {
	refreshExpiresAt := time.Now().Add(refreshTokenTTL)
	refreshClaims := &authjwt.Claims{
		UserID:        strconv.FormatUint(staffID, 10),
		ClinicID:      mainClinicID,
		IsSystemAdmin: isSystemAdmin,
		ClinicIDs:     clinicIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        newTokenJti(),
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   refreshTokenSubject,
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to sign refresh token")
	}
	return &IssuedToken{Token: refreshTokenStr, ExpiresAt: refreshExpiresAt}, nil
}

// VerifyAccessToken は access token を署名検証し claims を返す（ブラックリスト照合なし）。
// middleware.Auth と bit-compatible にするため BL は見ない。空でない Subject は拒否する。
func (s *tokenService) VerifyAccessToken(tokenStr string) (*authjwt.Claims, error) {
	claims, err := s.parseClaims(tokenStr)
	if err != nil {
		return nil, apperrors.WrapUnauthorized("invalid or expired token")
	}
	if claims.Subject != "" {
		return nil, apperrors.WrapUnauthorized("invalid token type")
	}
	return claims, nil
}

func (s *tokenService) VerifyRefreshToken(ctx context.Context, tokenStr string) (*authjwt.Claims, error) {
	claims, err := s.parseClaims(tokenStr)
	if err != nil {
		return nil, apperrors.WrapUnauthorized("invalid or expired refresh token")
	}

	if claims.Subject != refreshTokenSubject {
		return nil, apperrors.WrapUnauthorized("invalid token type")
	}

	if claims.ID != "" && s.tokenBlacklist != nil {
		revoked, blacklistErr := s.tokenBlacklist.IsRevoked(ctx, claims.ID)
		if blacklistErr != nil {
			slog.ErrorContext(ctx, "token blacklist check failed", "jti", claims.ID, "error", blacklistErr)
			return nil, apperrors.WrapUnauthorized("token validation failed")
		}
		if revoked {
			return nil, apperrors.WrapUnauthorized("token has been revoked")
		}
	}

	return claims, nil
}

// ParseRefreshTokenClaims は refresh token を署名検証し claims を返す（ブラックリスト照合なし）。
// Logout のベストエフォート失効向け。access token 誤失効防止のため subject=refresh は必須。
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
