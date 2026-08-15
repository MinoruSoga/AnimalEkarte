package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type logoutAuditIdentity struct {
	staffID  uint64
	clinicID uint64
	valid    bool
}

type refreshTokenRevocation struct {
	familyID string
}

const maxRefreshTokenCookieAggregateBytes = 32 << 10

func (h *HTTPHandler) authService() AuthService {
	if h.deps.Auth != nil {
		return h.deps.Auth
	}
	return NewAuthService(h.deps.Accounts, h.deps.Staff, h.deps.EffectivePermissions)
}

// ResolveClinicInfo returns the main clinic and all assigned clinic IDs.
func ResolveClinicInfo(
	assignments []model.StaffClinicAssignment,
) (mainClinicID string, clinicIDs []uint64) {
	return NewAuthService(nil, nil, nil).ResolveClinicInfo(assignments)
}

// ResolveSystemAdminMainClinicID keeps an active preferred clinic or falls
// back to the first active, nonzero clinic for a system administrator.
func ResolveSystemAdminMainClinicID(
	mainClinicID string,
	isSystemAdmin bool,
	allClinics []model.Clinic,
) string {
	return NewAuthService(nil, nil, nil).
		ResolveSystemAdminMainClinicID(mainClinicID, isSystemAdmin, allClinics)
}

// AuditClinicIDFromAssignments chooses the main clinic, falling back to the first assignment.
func AuditClinicIDFromAssignments(assignments []model.StaffClinicAssignment) (uint64, bool) {
	if len(assignments) == 0 {
		return 0, false
	}
	fallback := assignments[0].ClinicID
	for i := range assignments {
		if assignments[i].IsMain {
			return assignments[i].ClinicID, true
		}
	}
	return fallback, true
}

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

// RefreshToken verifies, rotates, revokes, and reissues the refresh token.
func (h *HTTPHandler) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()
	token, err := c.Cookie(RefreshTokenCookieName)
	if err != nil || token == "" {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("refresh token not found"))
		return
	}

	claims, err := h.deps.Tokens.VerifyRefreshToken(ctx, token)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	staffID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("invalid token"))
		return
	}
	staff, err := h.deps.Staff.GetByID(ctx, staffID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			httpapi.RespondError(c, apperrors.Wrap(err, "failed to get refresh staff"))
			return
		}
		httpapi.RespondError(c, apperrors.WrapUnauthorized("user not found"))
		return
	}
	if staff == nil ||
		!staff.IsActive ||
		staff.DeletedAt.Valid ||
		staff.AccountID == nil {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("invalid refresh identity"))
		return
	}

	account, err := h.deps.Accounts.GetByID(ctx, *staff.AccountID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			httpapi.RespondError(c, apperrors.Wrap(err, "failed to get refresh account"))
			return
		}
		httpapi.RespondError(c, apperrors.WrapUnauthorized("invalid refresh identity"))
		return
	}
	if account == nil ||
		account.ID != *staff.AccountID ||
		!account.IsActive ||
		account.DeletedAt.Valid {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("invalid refresh identity"))
		return
	}
	accountEpoch := account.UpdatedAt.UnixNano()
	if !TokenMatchesAccountEpoch(claims, accountEpoch) {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("refresh session has expired"))
		return
	}

	assignments, err := h.deps.StaffAssignments.FindAllByStaffID(ctx, staff.ID)
	if err != nil {
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to get clinic assignments"))
		return
	}
	mainClinicID, clinicIDs := h.authService().ResolveClinicInfo(assignments)
	if account.IsSystemAdmin {
		allClinics, listErr := h.deps.Clinics.ListClinics(ctx)
		if listErr != nil {
			httpapi.RespondError(c, apperrors.Wrap(listErr, "failed to get clinics"))
			return
		}
		mainClinicID = h.authService().ResolveSystemAdminMainClinicID(
			mainClinicID,
			account.IsSystemAdmin,
			allClinics,
		)
		clinicIDs = activeSystemAdminClinicIDs(allClinics)
	}
	if mainClinicID == "" {
		httpapi.RespondError(c, apperrors.WrapForbidden("no clinic access is available"))
		return
	}

	familyID, validFamily := refreshTokenFamilyID(claims)
	if !validFamily {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("invalid refresh token claims"))
		return
	}
	if h.deps.TokenBlacklist == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("token validation unavailable"))
		return
	}
	if err := h.deps.TokenBlacklist.RevokeToken(
		ctx,
		claims.ID,
		claims.ExpiresAt.Time,
	); err != nil {
		if apperrors.IsAlreadyExists(err) {
			if familyErr := revokeRefreshTokenFamily(
				ctx,
				h.deps.TokenBlacklist,
				familyID,
			); familyErr != nil {
				slog.ErrorContext(ctx, "concurrent refresh family revocation failed")
				httpapi.RespondError(c, familyErr)
				return
			}
			httpapi.RespondError(c, apperrors.WrapUnauthorized("refresh token reuse detected"))
			return
		}
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to revoke old refresh token"))
		return
	}

	if err := h.issueAuthCookies(
		c,
		staff.ID,
		mainClinicID,
		account.IsSystemAdmin,
		clinicIDs,
		accountEpoch,
		familyID,
		claims.ExpiresAt.Time,
	); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "token refreshed"})
}

// GetMe returns the authenticated staff profile and effective permissions.
func (h *HTTPHandler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()
	userIDValue, exists := c.Get("user_id")
	if !exists {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("missing user context"))
		return
	}
	userID, ok := userIDValue.(string)
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("invalid user context"))
		return
	}
	mainClinicIDValue, _ := c.Get("clinic_id")
	mainClinicID, _ := mainClinicIDValue.(string)

	parsedUserID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("invalid user id"))
		return
	}
	staff, err := h.deps.Staff.GetByID(ctx, parsedUserID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	var account *model.Account
	if staff.AccountID != nil {
		account, err = h.deps.Accounts.GetByID(ctx, *staff.AccountID)
		if err != nil {
			httpapi.RespondError(c, err)
			return
		}
	}

	assignments, err := h.deps.StaffAssignments.FindAllByStaffID(ctx, staff.ID)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to find clinic assignments",
			"error", err,
			"staff_id", staff.ID,
		)
	} else {
		staff = withClinicAssignments(staff, assignments)
	}

	allClinics, err := h.deps.Clinics.ListClinics(ctx)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	isSystemAdmin := account != nil && account.IsSystemAdmin
	c.JSON(
		http.StatusOK,
		h.BuildMeResponse(c, staff, account, mainClinicID, isSystemAdmin, allClinics),
	)
}

// AuthenticateUser validates credentials and records attributable password failures.
func (h *HTTPHandler) AuthenticateUser(
	ctx context.Context,
	email, password, clientIP, userAgent string,
) (*model.Account, *model.Staff, error) {
	timing := h.loginFailureTiming.withDefaults()
	startedAt := timing.now()
	account, staff, err := h.authService().AuthenticateUser(ctx, email, password)
	if err != nil {
		if errors.Is(err, apperrors.ErrUnauthorized) {
			accountID, knownAccount := IsAuthenticateWrongPassword(err)
			h.enqueueLoginFailureAudit(
				accountID,
				knownAccount,
				clientIP,
				userAgent,
			)
			timing.wait(ctx, startedAt)
		}
		return nil, nil, err
	}
	return account, staff, nil
}

func withClinicAssignments(
	staff *model.Staff,
	assignments []model.StaffClinicAssignment,
) *model.Staff {
	if staff == nil {
		return nil
	}
	result := *staff
	result.ClinicAssignments = append(
		[]model.StaffClinicAssignment(nil),
		assignments...,
	)
	return &result
}

// AuditKnownAccountLoginFailure writes an attributable wrong-password audit event.
func (h *HTTPHandler) AuditKnownAccountLoginFailure(
	ctx context.Context,
	accountID uint64,
	clientIP, userAgent string,
) {
	if h.deps.Staff == nil || h.deps.StaffAssignments == nil || h.deps.Audit == nil {
		return
	}
	staff, err := h.deps.Staff.FindByAccountID(ctx, accountID)
	if err != nil {
		slog.WarnContext(
			ctx,
			"skip audit log for login failure: staff not resolved",
			"account_id", accountID,
			"error_type", fmt.Sprintf("%T", err),
		)
		return
	}
	assignments, err := h.deps.StaffAssignments.FindAllByStaffID(ctx, staff.ID)
	if err != nil {
		slog.WarnContext(
			ctx,
			"skip audit log for login failure: clinic assignments not resolved",
			"staff_id", staff.ID,
			"error_type", fmt.Sprintf("%T", err),
		)
		return
	}
	clinicID, ok := auditClinicIDFromStaffAssignments(staff.ID, assignments)
	if !ok {
		clinicID, ok = h.systemAdminAuditClinicID(ctx, accountID)
		if !ok {
			slog.WarnContext(
				ctx,
				"skip audit log for login failure: no attributable clinic",
				"staff_id", staff.ID,
				"account_id", accountID,
			)
			return
		}
	}
	if err := h.deps.Audit.LogAuthLogin(
		ctx,
		&clinicID,
		&staff.ID,
		model.AuditActionAuthLoginFailure,
		clientIP,
		userAgent,
	); err != nil {
		slog.ErrorContext(
			ctx,
			"audit log failed for login failure",
			"staff_id", staff.ID,
			"clinic_id", clinicID,
			"error_type", fmt.Sprintf("%T", err),
		)
	}
}

func (h *HTTPHandler) systemAdminAuditClinicID(
	ctx context.Context,
	accountID uint64,
) (uint64, bool) {
	if h.deps.Accounts == nil || h.deps.Clinics == nil {
		return 0, false
	}
	account, err := h.deps.Accounts.GetByID(ctx, accountID)
	if err != nil {
		slog.WarnContext(
			ctx,
			"skip system administrator audit clinic resolution: account unavailable",
			"account_id", accountID,
			"error_type", fmt.Sprintf("%T", err),
		)
		return 0, false
	}
	if account == nil ||
		account.ID != accountID ||
		!account.IsActive ||
		account.DeletedAt.Valid ||
		!account.IsSystemAdmin {
		return 0, false
	}
	clinics, err := h.deps.Clinics.ListClinics(ctx)
	if err != nil {
		slog.WarnContext(
			ctx,
			"skip system administrator audit clinic resolution: clinics unavailable",
			"account_id", accountID,
			"error_type", fmt.Sprintf("%T", err),
		)
		return 0, false
	}
	for i := range clinics {
		if clinics[i].ID != 0 && clinics[i].IsActive {
			return clinics[i].ID, true
		}
	}
	return 0, false
}

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
