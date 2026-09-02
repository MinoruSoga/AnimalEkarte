package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// Login authenticates a staff account and issues access/refresh cookies.
func (h *HTTPHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var input LoginInput
	if err := bindAuthJSON(c, &input); err != nil {
		httpapi.RespondError(c, err)
		return
	}

	account, staff, err := h.AuthenticateUser(
		ctx,
		input.Email,
		input.Password,
		c.ClientIP(),
		c.Request.Header.Get("User-Agent"),
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	assignments, err := h.deps.StaffAssignments.FindAllByStaffID(ctx, staff.ID)
	if err != nil {
		httpapi.RespondError(c, apperrors.Wrap(err, "所属クリニック情報の取得に失敗しました"))
		return
	}
	staff = withClinicAssignments(staff, assignments)

	mainClinicID, clinicIDs := h.authService().ResolveClinicInfo(assignments)
	allClinics, err := h.deps.Clinics.ListClinics(ctx)
	if err != nil {
		if account.IsSystemAdmin {
			httpapi.RespondError(
				c,
				apperrors.Wrap(err, "failed to resolve system administrator clinic"),
			)
			return
		}
		allClinics = nil
	}
	mainClinicID = h.authService().
		ResolveSystemAdminMainClinicID(mainClinicID, account.IsSystemAdmin, allClinics)
	if account.IsSystemAdmin {
		clinicIDs = activeSystemAdminClinicIDs(allClinics)
	}
	if mainClinicID == "" {
		httpapi.RespondError(c, apperrors.WrapForbidden("no clinic access is available"))
		return
	}

	if err := h.IssueAuthCookies(
		c,
		staff.ID,
		mainClinicID,
		account.IsSystemAdmin,
		clinicIDs,
		account.UpdatedAt.UnixNano(),
	); err != nil {
		httpapi.RespondError(c, err)
		return
	}

	c.Set("clinic_id", mainClinicID)
	c.Set("user_id", strconv.FormatUint(staff.ID, 10))
	meResponse := h.BuildMeResponse(
		c,
		staff,
		account,
		mainClinicID,
		account.IsSystemAdmin,
		allClinics,
	)

	if auditClinicID, parseErr := strconv.ParseUint(mainClinicID, 10, 64); parseErr == nil &&
		auditClinicID != 0 &&
		h.deps.Audit != nil {
		auditCtx, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			credentialAuditTimeout,
		)
		defer cancel()
		if logErr := h.deps.Audit.LogAuthLogin(
			auditCtx,
			&auditClinicID,
			&staff.ID,
			model.AuditActionAuthLoginSuccess,
			c.ClientIP(),
			c.Request.Header.Get("User-Agent"),
		); logErr != nil {
			slog.ErrorContext(
				ctx,
				"audit log failed for login success",
				"staff_id", staff.ID,
				"clinic_id", auditClinicID,
				"error_type", fmt.Sprintf("%T", logErr),
			)
		}
	}

	c.JSON(http.StatusOK, LoginResponse{
		IsSystemAdmin: account.IsSystemAdmin,
		User:          meResponse,
	})
}

// BuildMeResponse assembles /me from already loaded domain models.
func (h *HTTPHandler) BuildMeResponse(
	c *gin.Context,
	staff *model.Staff,
	account *model.Account,
	mainClinicID string,
	isSystemAdmin bool,
	allClinics []model.Clinic,
) *MeResponse {
	clinicNameMap := make(map[string]string, len(allClinics))
	for i := range allClinics {
		clinic := &allClinics[i]
		clinicNameMap[strconv.FormatUint(clinic.ID, 10)] = clinic.Name
	}
	mainClinicID = h.authService().
		ResolveSystemAdminMainClinicID(mainClinicID, isSystemAdmin, allClinics)
	permissions := h.CalculateEffectivePermissions(c, isSystemAdmin, staff.ID)
	return ToMeResponse(
		staff,
		account,
		mainClinicID,
		clinicNameMap,
		allClinics,
		permissions,
	)
}

// Logout revokes every session family represented by a valid refresh cookie
// and clears all current/legacy cookies. Independent login families remain active.
func (h *HTTPHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	auditIdentity, err := h.revokeRefreshTokenCookies(ctx, c.Request.Cookies())
	if err != nil {
		slog.ErrorContext(ctx, "logout refresh cookie processing failed")
		httpapi.RespondError(c, err)
		return
	}

	ClearCookie(c, AccessTokenCookieName, "/", h.cookies)
	ClearCookie(c, LegacyTokenCookieName, "/", h.cookies)
	ClearCookie(c, RefreshTokenCookieName, RefreshTokenCookiePath, h.cookies)
	ClearCookie(c, RefreshTokenCookieName, LegacyRefreshTokenCookiePath, h.cookies)
	ClearCookie(c, "prev_clinic_id", "/", h.cookies)

	h.auditLogoutBestEffort(c, auditIdentity)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// RedirectLogout preserves POST and CSRF headers for cached legacy clients.
func (h *HTTPHandler) RedirectLogout(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, "/api/v1/auth/refresh/logout")
}

func (h *HTTPHandler) revokeRefreshTokenCookies(
	ctx context.Context,
	cookies []*http.Cookie,
) (logoutAuditIdentity, error) {
	tokens, collectionErr := UniqueRefreshTokenValues(cookies)
	revocations, auditIdentity := h.parseRefreshTokenRevocations(tokens)
	if len(revocations) > 0 && h.deps.TokenBlacklist == nil {
		return logoutAuditIdentity{}, errors.Join(
			collectionErr,
			errors.New("token blacklist service unavailable"),
		)
	}

	result := collectionErr
	for _, revocation := range revocations {
		err := revokeRefreshTokenFamily(
			ctx,
			h.deps.TokenBlacklist,
			revocation.familyID,
		)
		if err != nil {
			result = errors.Join(result, err)
		}
	}
	return auditIdentity, result
}

// UniqueRefreshTokenValues deduplicates refresh values while bounding retained
// input to the production server's 32 KiB request-header limit.
func UniqueRefreshTokenValues(cookies []*http.Cookie) ([]string, error) {
	seenTokens := make(map[string]struct{})
	tokens := make([]string, 0)
	aggregateBytes := 0
	for _, cookie := range cookies {
		if cookie.Name != RefreshTokenCookieName || cookie.Value == "" {
			continue
		}
		cookieBytes := len(cookie.Name) + len(cookie.Value) + len("=") + len("; ")
		if cookieBytes > maxRefreshTokenCookieAggregateBytes-aggregateBytes {
			return tokens, apperrors.WrapInvalidInput("invalid refresh cookie set")
		}
		aggregateBytes += cookieBytes
		if _, duplicate := seenTokens[cookie.Value]; duplicate {
			continue
		}
		seenTokens[cookie.Value] = struct{}{}
		tokens = append(tokens, cookie.Value)
	}
	return tokens, nil
}

func (h *HTTPHandler) parseRefreshTokenRevocations(
	tokens []string,
) ([]refreshTokenRevocation, logoutAuditIdentity) {
	revocations := make([]refreshTokenRevocation, 0, len(tokens))
	seenFamilies := make(map[string]struct{}, len(tokens))
	var signedUserID string
	var signedClinicID string
	hasSignedIdentity := false
	ambiguousIdentity := false
	for _, token := range tokens {
		if len(token) > maxRefreshTokenCookieBytes {
			continue
		}

		claims, err := h.deps.Tokens.ParseRefreshTokenClaims(token)
		if err != nil || claims.ID == "" || claims.ExpiresAt == nil {
			continue
		}
		familyID, validFamily := refreshTokenFamilyID(claims)
		if !validFamily {
			continue
		}
		if hasSignedIdentity &&
			(claims.UserID != signedUserID || claims.ClinicID != signedClinicID) {
			ambiguousIdentity = true
		}
		if !hasSignedIdentity {
			signedUserID = claims.UserID
			signedClinicID = claims.ClinicID
			hasSignedIdentity = true
		}
		if _, duplicateFamily := seenFamilies[familyID]; duplicateFamily {
			continue
		}
		seenFamilies[familyID] = struct{}{}
		revocations = append(revocations, refreshTokenRevocation{
			familyID: familyID,
		})
	}

	staffID, staffErr := strconv.ParseUint(signedUserID, 10, 64)
	clinicID, clinicErr := strconv.ParseUint(signedClinicID, 10, 64)
	return revocations, logoutAuditIdentity{
		staffID:  staffID,
		clinicID: clinicID,
		valid: hasSignedIdentity &&
			!ambiguousIdentity &&
			staffErr == nil &&
			clinicErr == nil,
	}
}

func (h *HTTPHandler) auditLogoutBestEffort(
	c *gin.Context,
	fallback logoutAuditIdentity,
) {
	if h.deps.Audit == nil {
		return
	}
	userIDValue, hasUser := c.Get("user_id")
	clinicIDValue, hasClinic := c.Get("clinic_id")
	if !hasUser || !hasClinic {
		if fallback.valid {
			h.logLogoutAudit(c, fallback.staffID, fallback.clinicID)
		}
		return
	}
	userID, userOK := userIDValue.(string)
	clinicID, clinicOK := clinicIDValue.(string)
	if !userOK || !clinicOK {
		return
	}
	staffID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return
	}
	parsedClinicID, err := strconv.ParseUint(clinicID, 10, 64)
	if err != nil {
		return
	}
	h.logLogoutAudit(c, staffID, parsedClinicID)
}

func (h *HTTPHandler) logLogoutAudit(c *gin.Context, staffID, clinicID uint64) {
	ctx := c.Request.Context()
	if err := h.deps.Audit.LogAuthLogin(
		ctx,
		&clinicID,
		&staffID,
		model.AuditActionAuthLogout,
		c.ClientIP(),
		c.Request.Header.Get("User-Agent"),
	); err != nil {
		slog.ErrorContext(
			ctx,
			"audit log failed for logout",
			"staff_id", staffID,
			"clinic_id", clinicID,
			"error", err,
		)
	}
}
