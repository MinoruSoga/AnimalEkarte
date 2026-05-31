package handler

import (
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

	permMap := h.calculateEffectivePermissions(ctx, account.IsSystemAdmin, staff.ID)

	// 監査ログ: ログイン成功
	if len(clinicIDs) > 0 {
		mainCID := clinicIDs[0]
		_ = h.svc.Audit.LogAuthLogin(ctx, &mainCID, &staff.ID, model.AuditActionAuthLoginSuccess, c.ClientIP(), c.Request.Header.Get("User-Agent"))
	}

	c.JSON(http.StatusOK, LoginResponse{
		IsSystemAdmin: account.IsSystemAdmin,
		User:          toMeResponse(staff, account, mainClinicID, clinicNameMap, allClinics, permMap),
	})
}

// Logout godoc
func (h *Handler) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	// 監査ログ: ログアウト（ベストエフォート）
	// extractStaffID/extractClinicID は Auth middleware が設定する user_id/clinic_id を前提とし、
	// 存在しない場合に 401 レスポンスを書き込む副作用がある。
	// /logout は保護グループ外（Auth middleware なし）なのでこれらの関数は使用しない。
	// 代わりに c.Get() で直接チェックし、存在する場合のみ監査ログを記録する。
	userIDVal, hasUser := c.Get("user_id")
	clinicIDVal, hasClinic := c.Get("clinic_id")
	if hasUser && hasClinic {
		if userIDStr, ok := userIDVal.(string); ok {
			if clinicIDStr, ok := clinicIDVal.(string); ok {
				if staffID, err := strconv.ParseUint(userIDStr, 10, 64); err == nil {
					if clinicID, err := strconv.ParseUint(clinicIDStr, 10, 64); err == nil {
						_ = h.svc.Audit.LogAuthLogin(ctx, &clinicID, &staffID, model.AuditActionAuthLogout, c.ClientIP(), c.Request.Header.Get("User-Agent"))
					}
				}
			}
		}
	}

	isProduction := h.cfg.GinMode == "release"
	sameSite := http.SameSiteLaxMode
	if isProduction {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     legacyCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/api/v1/auth/refresh",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "prev_clinic_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
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

	// Subject が "refresh" であることを確認（access_token の流用を防止）
	if claims.Subject != "refresh" {
		RespondError(c, apperrors.WrapUnauthorized("invalid token type"))
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
	// resolveClinicInfo で最新割り当てから mainClinicID を再計算（旧 claims の値を引き継がない）
	mainClinicID, clinicIDs := resolveClinicInfo(assignments)

	// 新しい access_token（15分）
	expiresAt := time.Now().Add(15 * time.Minute)
	newAccessClaims := &middleware.JWTClaims{
		UserID:        claims.UserID,
		ClinicID:      mainClinicID,
		IsSystemAdmin: claims.IsSystemAdmin,
		ClinicIDs:     clinicIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	newAccess := jwt.NewWithClaims(jwt.SigningMethodHS256, newAccessClaims)
	newAccessStr, signErr := newAccess.SignedString([]byte(h.cfg.JWTSecret))
	if signErr != nil {
		RespondError(c, apperrors.Wrap(signErr, "failed to sign access token"))
		return
	}

	// 新しい refresh_token（7日、rotation）
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	newRefreshClaims := &middleware.JWTClaims{
		UserID:        claims.UserID,
		ClinicID:      mainClinicID,
		IsSystemAdmin: claims.IsSystemAdmin,
		ClinicIDs:     clinicIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "refresh",
		},
	}
	newRefresh := jwt.NewWithClaims(jwt.SigningMethodHS256, newRefreshClaims)
	newRefreshStr, rSignErr := newRefresh.SignedString([]byte(h.cfg.JWTSecret))
	if rSignErr != nil {
		RespondError(c, apperrors.Wrap(rSignErr, "failed to sign refresh token"))
		return
	}

	// Cookie 設定
	sameSite := http.SameSiteNoneMode
	secure := true

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    newAccessStr,
		Path:     "/",
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    newRefreshStr,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   int(time.Until(refreshExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

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
	if err == nil {
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

	permMap := h.calculateEffectivePermissions(ctx, isSystemAdmin, staff.ID)

	c.JSON(http.StatusOK, toMeResponse(staff, account, mainClinicIDStr, clinicNameMap, allClinics, permMap))
}
