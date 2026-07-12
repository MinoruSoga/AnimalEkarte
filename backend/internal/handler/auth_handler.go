package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	accessTokenCookieName  = "access_token"
	refreshTokenCookieName = "refresh_token"
	legacyCookieName       = "auth_token"

	// E-3: issueAuthCookies / RefreshToken で重複していた TTL・subject の直値を集約。
	accessTokenTTL      = 15 * time.Minute
	refreshTokenTTL     = 7 * 24 * time.Hour
	refreshTokenSubject = "refresh"
)

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

	mainClinicID, clinicIDs := resolveClinicInfo(assignments)

	// クリニック一覧取得 (フォールバック適用に必要なため JWT 発行前に取得)
	allClinics, err := h.svc.Clinic.ListClinics(ctx)
	if err != nil {
		allClinics = nil
	}

	// system_admin で assignments なしの場合、allClinics[0] を main にフォールバック (JWT 発行前に解決)
	mainClinicID = resolveSystemAdminMainClinicID(mainClinicID, account.IsSystemAdmin, allClinics)

	if err := h.issueAuthCookies(c, staff.ID, mainClinicID, account.IsSystemAdmin, clinicIDs); err != nil {
		RespondError(c, err)
		return
	}

	clinicNameMap := make(map[string]string)
	for i := range allClinics {
		cl := &allClinics[i]
		clinicNameMap[strconv.FormatUint(cl.ID, 10)] = cl.Name
	}

	// Set context values that extractClinicID expects.
	// calculateEffectivePermissions calls extractClinicID which reads c.Get("clinic_id").
	// Without these, extractClinicID writes a 401 "missing clinic context" side-effect
	// (RespondError without c.Abort) and then Login continues, producing a double-write
	// response with HTTP 401 + user payload concatenated.
	c.Set("clinic_id", mainClinicID)
	c.Set("user_id", strconv.FormatUint(staff.ID, 10))

	permMap := h.calculateEffectivePermissions(c, account.IsSystemAdmin, staff.ID)

	// 監査ログ: ログイン成功
	if len(clinicIDs) > 0 {
		mainCID := clinicIDs[0]
		if logErr := h.svc.Audit.LogAuthLogin(ctx, &mainCID, &staff.ID, model.AuditActionAuthLoginSuccess, c.ClientIP(), c.Request.Header.Get("User-Agent")); logErr != nil {
			slog.ErrorContext(ctx, "audit log failed for login success", "staff_id", staff.ID, "error", logErr)
		}
	}

	c.JSON(http.StatusOK, LoginResponse{
		IsSystemAdmin: account.IsSystemAdmin,
		User:          toMeResponse(staff, account, mainClinicID, clinicNameMap, allClinics, permMap),
	})
}

// Logout godoc
func (h *Handler) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	// refresh_token の jti をブラックリストに登録してサーバーサイド失効させる（ベストエフォート）。
	// 失効に失敗してもログアウト自体はブロックしない。
	if tokenStr, err := c.Cookie(refreshTokenCookieName); err == nil && tokenStr != "" {
		claims := &middleware.JWTClaims{}
		if _, parseErr := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(h.cfg.JWTSecret), nil
		}); parseErr == nil && claims.ID != "" && claims.ExpiresAt != nil {
			if revokeErr := h.svc.TokenBlacklist.RevokeToken(ctx, claims.ID, claims.ExpiresAt.Time); revokeErr != nil {
				slog.ErrorContext(ctx, "failed to revoke refresh token on logout (best-effort)", "jti", claims.ID, "error", revokeErr)
			}
		}
	}

	h.auditLogoutBestEffort(c)

	isProduction := h.cfg.GinMode == "release"
	sameSite := http.SameSiteLaxMode
	if isProduction {
		sameSite = http.SameSiteNoneMode
	}

	clearCookie(c, accessTokenCookieName, "/", isProduction, sameSite)
	clearCookie(c, legacyCookieName, "/", isProduction, sameSite)
	clearCookie(c, refreshTokenCookieName, "/api/v1/auth/refresh", isProduction, sameSite)
	clearCookie(c, "prev_clinic_id", "/", isProduction, sameSite)

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// auditLogoutBestEffort はログアウト監査ログをベストエフォートで記録する（E-2）。
// extractStaffID/extractClinicID は Auth middleware が設定する user_id/clinic_id を前提とし、
// 存在しない場合に 401 レスポンスを書き込む副作用がある。
// /logout は保護グループ外（Auth middleware なし）なのでこれらの関数は使用しない。
// 代わりに c.Get() で直接チェックし、存在する場合のみ監査ログを記録する。
func (h *Handler) auditLogoutBestEffort(c *gin.Context) {
	userIDVal, hasUser := c.Get("user_id")
	clinicIDVal, hasClinic := c.Get("clinic_id")
	if !hasUser || !hasClinic {
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

	// refresh_token を検証
	claims := &middleware.JWTClaims{}
	if _, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.cfg.JWTSecret), nil
	}); err != nil {
		RespondError(c, apperrors.WrapUnauthorized("invalid or expired refresh token"))
		return
	}

	// Subject が refreshTokenSubject であることを確認（access_token の流用を防止）
	if claims.Subject != refreshTokenSubject {
		RespondError(c, apperrors.WrapUnauthorized("invalid token type"))
		return
	}

	// JTI ブラックリスト照合（ログアウト済み・強制失効済みトークンを拒否）
	if claims.ID != "" {
		revoked, blacklistErr := h.svc.TokenBlacklist.IsRevoked(ctx, claims.ID)
		if blacklistErr != nil {
			// DB エラーはフェイルセーフ: 照合失敗はリフレッシュを拒否する
			slog.ErrorContext(ctx, "token blacklist check failed", "jti", claims.ID, "error", blacklistErr)
			RespondError(c, apperrors.WrapUnauthorized("token validation failed"))
			return
		}
		if revoked {
			RespondError(c, apperrors.WrapUnauthorized("token has been revoked"))
			return
		}
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
	// resolveClinicInfo で最新割り当てから mainClinicID を再計算（旧 claims の値を引き継がない）
	mainClinicID, clinicIDs := resolveClinicInfo(assignments)

	// Token rotation: 旧 JTI をブラックリストに登録して旧トークンを失効させる（ベストエフォート）。
	// 失効失敗は回帰させない（新トークン発行は続行）。
	if claims.ID != "" && claims.ExpiresAt != nil {
		if revokeErr := h.svc.TokenBlacklist.RevokeToken(ctx, claims.ID, claims.ExpiresAt.Time); revokeErr != nil {
			slog.ErrorContext(ctx, "failed to revoke old refresh token on rotation (best-effort)", "jti", claims.ID, "error", revokeErr)
		}
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
	clinicNameMap := make(map[string]string, len(allClinics))
	for i := range allClinics {
		cl := &allClinics[i]
		clinicNameMap[strconv.FormatUint(cl.ID, 10)] = cl.Name
	}

	// 実効権限を計算
	isSystemAdmin := false
	if account != nil {
		isSystemAdmin = account.IsSystemAdmin
	}

	mainClinicIDStr = resolveSystemAdminMainClinicID(mainClinicIDStr, isSystemAdmin, allClinics)

	permMap := h.calculateEffectivePermissions(c, isSystemAdmin, staff.ID)

	c.JSON(http.StatusOK, toMeResponse(staff, account, mainClinicIDStr, clinicNameMap, allClinics, permMap))
}
