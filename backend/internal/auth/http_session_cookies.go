package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ClearCookie immediately expires one cookie with the configured security attributes.
func ClearCookie(c *gin.Context, name, path string, config CookieConfig) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   config.Secure,
		SameSite: config.SameSite,
	})
}

// IssueAuthCookies issues the hardened 15-minute access and 7-day rotating refresh cookies.
func (h *HTTPHandler) IssueAuthCookies(
	c *gin.Context,
	staffID uint64,
	mainClinicID string,
	isSystemAdmin bool,
	clinicIDs []uint64,
	accountEpoch int64,
) error {
	return h.issueAuthCookies(
		c,
		staffID,
		mainClinicID,
		isSystemAdmin,
		clinicIDs,
		accountEpoch,
		"",
		time.Time{},
	)
}

func (h *HTTPHandler) issueAuthCookies(
	c *gin.Context,
	staffID uint64,
	mainClinicID string,
	isSystemAdmin bool,
	clinicIDs []uint64,
	accountEpoch int64,
	refreshFamilyID string,
	refreshFamilyExpiresAt time.Time,
) error {
	accessIssued, err := h.deps.Tokens.IssueAccessToken(
		staffID,
		mainClinicID,
		isSystemAdmin,
		clinicIDs,
		accountEpoch,
	)
	if err != nil {
		return err
	}
	if accessIssued == nil {
		return apperrors.WrapInternalServerError("access token issuer returned no token")
	}

	var refreshIssued *IssuedToken
	if refreshFamilyID == "" {
		refreshIssued, err = h.deps.Tokens.IssueRefreshToken(
			staffID,
			mainClinicID,
			isSystemAdmin,
			clinicIDs,
			accountEpoch,
		)
	} else {
		refreshIssued, err = h.deps.Tokens.IssueRefreshTokenInFamily(
			staffID,
			mainClinicID,
			isSystemAdmin,
			clinicIDs,
			accountEpoch,
			refreshFamilyID,
			refreshFamilyExpiresAt,
		)
	}
	if err != nil {
		return err
	}
	if refreshIssued == nil {
		return apperrors.WrapInternalServerError("refresh token issuer returned no token")
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     AccessTokenCookieName,
		Value:    accessIssued.Token,
		Path:     "/",
		MaxAge:   int(time.Until(accessIssued.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   h.cookies.Secure,
		SameSite: h.cookies.SameSite,
	})
	ClearCookie(c, RefreshTokenCookieName, LegacyRefreshTokenCookiePath, h.cookies)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    refreshIssued.Token,
		Path:     RefreshTokenCookiePath,
		MaxAge:   int(time.Until(refreshIssued.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   h.cookies.Secure,
		SameSite: h.cookies.SameSite,
	})
	return nil
}
