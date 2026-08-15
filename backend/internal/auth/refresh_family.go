package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/authjwt"
)

const (
	refreshFamilyBlacklistPrefix = "refresh-family:"
	maxRefreshFamilyIDBytes      = 1024
)

func refreshTokenFamilyID(claims *authjwt.Claims) (string, bool) {
	if claims == nil || claims.ID == "" {
		return "", false
	}
	familyID := claims.RefreshFamilyID
	if familyID == "" {
		// Tokens issued before refresh families were introduced become the root
		// of a family at their first rotation.
		familyID = claims.ID
	}
	return familyID, len(familyID) <= maxRefreshFamilyIDBytes
}

func refreshFamilyBlacklistKey(familyID string) string {
	digest := sha256.Sum256([]byte(familyID))
	return refreshFamilyBlacklistPrefix + hex.EncodeToString(digest[:])
}

func revokeRefreshTokenFamily(
	ctx context.Context,
	blacklist TokenBlacklistService,
	familyID string,
) error {
	if blacklist == nil {
		return apperrors.WrapInternalServerError("token validation unavailable")
	}
	if familyID == "" || len(familyID) > maxRefreshFamilyIDBytes {
		return apperrors.WrapUnauthorized("invalid refresh token claims")
	}
	err := blacklist.RevokeToken(
		ctx,
		refreshFamilyBlacklistKey(familyID),
		time.Now().Add(refreshTokenTTL),
	)
	if err != nil && !apperrors.IsAlreadyExists(err) {
		return apperrors.Wrap(err, "failed to revoke refresh token family")
	}
	return nil
}
