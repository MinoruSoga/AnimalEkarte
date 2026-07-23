package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	accessTokenCookieName  = "access_token"
	refreshTokenCookieName = "refresh_token"
	legacyCookieName       = "auth_token"

	refreshTokenCookiePath       = "/api/v1/auth"
	legacyRefreshTokenCookiePath = "/api/v1/auth/refresh"

	maxRefreshTokenCookieCount = 2
	maxRefreshTokenCookieBytes = 4096
)

type logoutAuditIdentity struct {
	staffID  uint64
	clinicID uint64
	valid    bool
}

type refreshTokenRevocation struct {
	jti       string
	expiresAt time.Time
}

// Login godoc
// Login はメール/パスワードで認証してJWTトークンを返す。
// accounts.password_hash を bcrypt で検証する。
func (h *Handler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	account, staff, err := h.authenticateUser(ctx, input.Email, input.Password, c.ClientIP(), c.Request.Header.Get("User-Agent"))
	if err != nil {
		RespondError(c, err)
		return
	}

	assignments, err := h.svc.StaffClinicAssignment.FindAllByStaffID(ctx, staff.ID)
	if err != nil {
		RespondError(c, apperrors.Wrap(err, "所属クリニック情報の取得に失敗しました"))
		return
	}
	staff.ClinicAssignments = assignments

	mainClinicID, clinicIDs := h.authSvc().ResolveClinicInfo(assignments)

	// クリニック一覧取得 (フォールバック適用に必要なため JWT 発行前に取得)
	allClinics, err := h.svc.Clinic.ListClinics(ctx)
	if err != nil {
		allClinics = nil
	}

	// system_admin で assignments なしの場合、allClinics[0] を main にフォールバック (JWT 発行前に解決)
	mainClinicID = h.authSvc().ResolveSystemAdminMainClinicID(mainClinicID, account.IsSystemAdmin, allClinics)

	if err := h.issueAuthCookies(c, staff.ID, mainClinicID, account.IsSystemAdmin, clinicIDs); err != nil {
		RespondError(c, err)
		return
	}

	// Set context values that extractClinicID expects.
	// calculateEffectivePermissions calls extractClinicID which reads c.Get("clinic_id").
	// Without these, extractClinicID writes a 401 "missing clinic context" side-effect
	// (RespondError without c.Abort) and then Login continues, producing a double-write
	// response with HTTP 401 + user payload concatenated.
	c.Set("clinic_id", mainClinicID)
	c.Set("user_id", strconv.FormatUint(staff.ID, 10))

	meResponse := h.buildMeResponse(c, staff, account, mainClinicID, account.IsSystemAdmin, allClinics)

	// 監査ログ: ログイン成功
	if len(clinicIDs) > 0 {
		mainCID := clinicIDs[0]
		if logErr := h.svc.Audit.LogAuthLogin(ctx, &mainCID, &staff.ID, model.AuditActionAuthLoginSuccess, c.ClientIP(), c.Request.Header.Get("User-Agent")); logErr != nil {
			slog.ErrorContext(ctx, "audit log failed for login success", "staff_id", staff.ID, "error", logErr)
		}
	}

	c.JSON(http.StatusOK, LoginResponse{
		IsSystemAdmin: account.IsSystemAdmin,
		User:          meResponse,
	})
}

// buildMeResponse は取得済みの staff/account/clinics から /me 応答を組み立てる（E-4）。
// clinics の取得とその失敗時ポリシーは呼び出し元の責務（Login=nil継続 / GetMe=エラー応答）。
// mainClinicID は呼び出し元で resolveSystemAdminMainClinicID 済みでも未解決でもよい
// （本関数内でも呼ぶため冪等 — Login は issueAuthCookies 前に解決済みの値を渡す）。
func (h *Handler) buildMeResponse(c *gin.Context, staff *model.Staff, account *model.Account,
	mainClinicID string, isSystemAdmin bool, allClinics []model.Clinic) *MeResponse {
	clinicNameMap := make(map[string]string, len(allClinics))
	for i := range allClinics {
		cl := &allClinics[i]
		clinicNameMap[strconv.FormatUint(cl.ID, 10)] = cl.Name
	}
	mainClinicID = h.authSvc().ResolveSystemAdminMainClinicID(mainClinicID, isSystemAdmin, allClinics)
	permMap := h.calculateEffectivePermissions(c, isSystemAdmin, staff.ID)
	return toMeResponse(staff, account, mainClinicID, clinicNameMap, allClinics, permMap)
}

// Logout godoc
func (h *Handler) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	// ParseRefreshTokenClaims は BL 照合なし（既に失効済みでも parse でき、冪等 revoke 可能）。
	auditIdentity, refreshRevokeErr := h.revokeRefreshTokenCookies(ctx, c.Request.Cookies())

	if refreshRevokeErr != nil {
		slog.ErrorContext(ctx, "logout refresh cookie processing failed", "error", refreshRevokeErr)
		RespondError(c, refreshRevokeErr)
		return
	}

	isProduction := h.cfg.GinMode == "release"
	sameSite := http.SameSiteLaxMode
	if isProduction {
		sameSite = http.SameSiteNoneMode
	}

	clearCookie(c, accessTokenCookieName, "/", isProduction, sameSite)
	clearCookie(c, legacyCookieName, "/", isProduction, sameSite)
	clearCookie(c, refreshTokenCookieName, refreshTokenCookiePath, isProduction, sameSite)
	clearCookie(c, refreshTokenCookieName, legacyRefreshTokenCookiePath, isProduction, sameSite)
	clearCookie(c, "prev_clinic_id", "/", isProduction, sameSite)

	h.auditLogoutBestEffort(c, auditIdentity)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// RedirectLogout keeps cached clients that still call /api/v1/logout safe while
// the refresh cookie is narrowed to /api/v1/auth. A 307 preserves POST and the
// CSRF header; the browser then attaches both legacy and current refresh cookies.
func (h *Handler) RedirectLogout(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, "/api/v1/auth/refresh/logout")
}

func (h *Handler) revokeRefreshTokenCookies(ctx context.Context, cookies []*http.Cookie) (logoutAuditIdentity, error) {
	tokens, err := uniqueRefreshTokenValues(cookies)
	if err != nil {
		return logoutAuditIdentity{}, err
	}

	revocations, auditIdentity, err := h.parseRefreshTokenRevocations(tokens)
	if err != nil {
		return logoutAuditIdentity{}, err
	}
	if len(revocations) > 0 && (h.svc == nil || h.svc.TokenBlacklist == nil) {
		return logoutAuditIdentity{}, errors.New("token blacklist service unavailable")
	}

	var result error
	for _, revocation := range revocations {
		if err := h.svc.TokenBlacklist.RevokeToken(ctx, revocation.jti, revocation.expiresAt); err != nil && !apperrors.IsAlreadyExists(err) {
			result = errors.Join(result, err)
		}
	}
	return auditIdentity, result
}

func uniqueRefreshTokenValues(cookies []*http.Cookie) ([]string, error) {
	seenTokens := make(map[string]struct{}, len(cookies))
	tokens := make([]string, 0, maxRefreshTokenCookieCount)
	for _, cookie := range cookies {
		if cookie.Name != refreshTokenCookieName || cookie.Value == "" {
			continue
		}
		if _, duplicate := seenTokens[cookie.Value]; duplicate {
			continue
		}
		seenTokens[cookie.Value] = struct{}{}
		if len(tokens) >= maxRefreshTokenCookieCount {
			return nil, apperrors.WrapInvalidInput("invalid refresh cookie set")
		}
		tokens = append(tokens, cookie.Value)
	}
	return tokens, nil
}

func (h *Handler) parseRefreshTokenRevocations(tokens []string) ([]refreshTokenRevocation, logoutAuditIdentity, error) {
	revocations := make([]refreshTokenRevocation, 0, len(tokens))
	var signedUserID string
	var signedClinicID string
	hasSignedIdentity := false
	for _, token := range tokens {
		if len(token) > maxRefreshTokenCookieBytes {
			continue
		}

		claims, err := h.tokenSvc().ParseRefreshTokenClaims(token)
		if err != nil || claims.ID == "" || claims.ExpiresAt == nil {
			continue
		}
		if hasSignedIdentity && (claims.UserID != signedUserID || claims.ClinicID != signedClinicID) {
			return nil, logoutAuditIdentity{}, apperrors.WrapInvalidInput("invalid refresh cookie identity")
		}
		if !hasSignedIdentity {
			signedUserID = claims.UserID
			signedClinicID = claims.ClinicID
			hasSignedIdentity = true
		}
		revocations = append(revocations, refreshTokenRevocation{jti: claims.ID, expiresAt: claims.ExpiresAt.Time})
	}

	staffID, staffErr := strconv.ParseUint(signedUserID, 10, 64)
	clinicID, clinicErr := strconv.ParseUint(signedClinicID, 10, 64)
	auditIdentity := logoutAuditIdentity{
		staffID: staffID, clinicID: clinicID,
		valid: hasSignedIdentity && staffErr == nil && clinicErr == nil,
	}
	return revocations, auditIdentity, nil
}

// auditLogoutBestEffort はログアウト監査ログをベストエフォートで記録する（E-2）。
// 保護グループ外の実ルートでは Auth middleware の context 値がないため、
// 署名検証済み refresh claims の staff/clinic を fallback として使う。
func (h *Handler) auditLogoutBestEffort(c *gin.Context, fallback logoutAuditIdentity) {
	if h.svc == nil || h.svc.Audit == nil {
		return
	}
	userIDVal, hasUser := c.Get("user_id")
	clinicIDVal, hasClinic := c.Get("clinic_id")
	if !hasUser || !hasClinic {
		if fallback.valid {
			h.logLogoutAudit(c, fallback.staffID, fallback.clinicID)
		}
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		return
	}
	clinicIDStr, ok := clinicIDVal.(string)
	if !ok {
		return
	}
	staffID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return
	}
	clinicID, err := strconv.ParseUint(clinicIDStr, 10, 64)
	if err != nil {
		return
	}

	h.logLogoutAudit(c, staffID, clinicID)
}

func (h *Handler) logLogoutAudit(c *gin.Context, staffID, clinicID uint64) {
	ctx := c.Request.Context()
	if logErr := h.svc.Audit.LogAuthLogin(ctx, &clinicID, &staffID, model.AuditActionAuthLogout, c.ClientIP(), c.Request.Header.Get("User-Agent")); logErr != nil {
		slog.ErrorContext(ctx, "audit log failed for logout", "staff_id", staffID, "clinic_id", clinicID, "error", logErr)
	}
}

// RefreshToken は refresh_token Cookie を検証し、新しい access_token + refresh_token を発行する。
// Token Rotation: 使用済み refresh_token は新しいものに置換される。
func (h *Handler) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()

	tokenStr, err := c.Cookie(refreshTokenCookieName)
	if err != nil || tokenStr == "" {
		RespondError(c, apperrors.WrapUnauthorized("refresh token not found"))
		return
	}

	claims, err := h.tokenSvc().VerifyRefreshToken(ctx, tokenStr)
	if err != nil {
		RespondError(c, err)
		return
	}

	// staff の有効性チェック
	staffID, parseErr := strconv.ParseUint(claims.UserID, 10, 64)
	if parseErr != nil {
		RespondError(c, apperrors.WrapUnauthorized("invalid token"))
		return
	}
	staff, findErr := h.svc.Staff.GetByID(ctx, staffID)
	if findErr != nil {
		RespondError(c, apperrors.WrapUnauthorized("user not found"))
		return
	}
	if !staff.IsActive {
		RespondError(c, apperrors.WrapUnauthorized("このアカウントは無効です"))
		return
	}

	// 所属クリニック再取得（assignment 変更があった場合に最新を反映）
	assignments, asgErr := h.svc.StaffClinicAssignment.FindAllByStaffID(ctx, staff.ID)
	if asgErr != nil {
		RespondError(c, apperrors.Wrap(asgErr, "failed to get clinic assignments"))
		return
	}
	// ResolveClinicInfo で最新割り当てから mainClinicID を再計算（旧 claims の値を引き継がない）
	mainClinicID, clinicIDs := h.authSvc().ResolveClinicInfo(assignments)

	// Token rotation: 旧 JTI の永続失効に成功した場合だけ新しいtokenを発行する。
	if revokeErr := h.svc.TokenBlacklist.RevokeToken(ctx, claims.ID, claims.ExpiresAt.Time); revokeErr != nil {
		RespondError(c, apperrors.Wrap(revokeErr, "failed to revoke old refresh token"))
		return
	}

	// 新しい access_token + refresh_token を発行して Cookie にセットする（E-3: ログイン時と同じ発行処理に委譲）。
	if err := h.issueAuthCookies(c, staff.ID, mainClinicID, claims.IsSystemAdmin, clinicIDs); err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token refreshed"})
}

// GetMe godoc
func (h *Handler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	userIDVal, exists := c.Get("user_id")
	if !exists {
		RespondError(c, apperrors.WrapUnauthorized("missing user context"))
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		RespondError(c, apperrors.WrapInternalServerError("invalid user context"))
		return
	}
	mainClinicIDVal, _ := c.Get("clinic_id")
	mainClinicIDStr, _ := mainClinicIDVal.(string)

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInternalServerError("invalid user id"))
		return
	}

	// staff を取得
	staff, err := h.svc.Staff.GetByID(ctx, userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	// account を取得
	var account *model.Account
	if staff.AccountID != nil {
		account, err = h.svc.Account.GetByID(ctx, *staff.AccountID)
		if err != nil {
			RespondError(c, err)
			return
		}
	}

	// clinic assignments を取得
	assignments, err := h.svc.StaffClinicAssignment.FindAllByStaffID(ctx, staff.ID)
	if err != nil {
		// F-1: 取得失敗は観測性のため slog 記録（P11）。挙動は従来通り — ClinicAssignments を
		// 未設定のまま続行する（error 伝播へのフリップは製品判断のため follow-up）。
		slog.ErrorContext(ctx, "failed to find clinic assignments", "error", err, "staff_id", staff.ID)
	} else {
		staff.ClinicAssignments = assignments
	}

	// クリニック一覧を取得
	allClinics, err := h.svc.Clinic.ListClinics(ctx)
	if err != nil {
		RespondError(c, err)
		return
	}

	isSystemAdmin := false
	if account != nil {
		isSystemAdmin = account.IsSystemAdmin
	}

	c.JSON(http.StatusOK, h.buildMeResponse(c, staff, account, mainClinicIDStr, isSystemAdmin, allClinics))
}
