package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/config"
	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	accessTokenCookieName  = "access_token"
	refreshTokenCookieName = "refresh_token"
	legacyCookieName       = "auth_token"
)

// buildMeResponse はスタッフデータと補助情報からMeResponseを構築する。
// effectivePerms は事前に計算された実効権限マップ。nil の場合はデフォルト（全権限なし）。
func buildMeResponse(staff *model.Staff, account *model.Account, mainClinicID string, clinicNameMap map[string]string, allClinics []model.Clinic, effectivePerms EffectivePermissions) *MeResponse {
	meClinicList := make([]MeClinicMembership, 0)
	isSystemAdmin := account != nil && account.IsSystemAdmin
	if staff != nil && len(staff.ClinicAssignments) > 0 {
		for _, asg := range staff.ClinicAssignments {
			clIDStr := strconv.FormatUint(asg.ClinicID, 10)
			meClinicList = append(meClinicList, MeClinicMembership{
				ClinicID:   clIDStr,
				ClinicName: clinicNameMap[clIDStr],
				IsMain:     clIDStr == mainClinicID,
			})
		}
	}

	permMap := effectivePerms
	if permMap == nil {
		permMap = make(EffectivePermissions)
	}

	var meClinic *MeClinicInfo
	for i := range allClinics {
		if strconv.FormatUint(allClinics[i].ID, 10) != mainClinicID {
			continue
		}
		cl := &allClinics[i]
		var logoURL *string
		if cl.LogoURL != "" {
			logoURL = &cl.LogoURL
		}
		meClinic = &MeClinicInfo{
			ID:                 strconv.FormatUint(cl.ID, 10),
			Name:               cl.Name,
			PostalCode:         cl.PostalCode,
			Address:            cl.Address,
			PhoneNumber:        cl.PhoneNumber,
			FaxNumber:          cl.FaxNumber,
			RegistrationNumber: cl.RegistrationNumber,
			DirectorName:       cl.DirectorName,
			Email:              cl.Email,
			Website:            cl.Website,
			LogoURL:            logoURL,
			StandardTaxRate:    cl.StandardTaxRate,
			ReducedTaxRate:     cl.ReducedTaxRate,
		}
		break
	}

	var occupation *string
	if staff != nil && staff.Occupation != nil {
		occ := staff.Occupation.Name
		occupation = &occ
	}

	staffID := uint64(0)
	if staff != nil {
		staffID = staff.ID
	}

	return &MeResponse{
		ID:            strconv.FormatUint(staffID, 10),
		Email:         account.Email,
		DisplayName:   staff.Name,
		IsSystemAdmin: isSystemAdmin,
		Occupation:    occupation,
		MainClinicID:  mainClinicID,
		Clinic:        meClinic,
		Clinics:       meClinicList,
		Permissions:   permMap,
	}
}

// authenticateUser はメール/パスワードを検証してアカウントとスタッフを返す。
// 認証失敗・アカウント無効・スタッフ無効の場合は apperrors.ErrUnauthorized ラップエラーを返す。
func (h *Handler) authenticateUser(ctx context.Context, email, password, clientIP, userAgent string) (*model.Account, *model.Staff, error) {
	account, err := h.svc.Account.FindByEmail(ctx, email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			// 監査ログ: ログイン失敗（アカウント不存在）
			if auditErr := h.svc.Audit.LogAuthLogin(ctx, nil, nil, model.AuditActionAuthLoginFailure, clientIP, userAgent); auditErr != nil {
				slog.ErrorContext(ctx, "failed to log login failure", slog.String("error", auditErr.Error()))
			}
			return nil, nil, apperrors.WrapUnauthorized("メールアドレスまたはパスワードが正しくありません")
		}
		return nil, nil, err
	}
	if !account.IsActive {
		return nil, nil, apperrors.WrapUnauthorized("アカウントが無効です")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
		// 監査ログ: ログイン失敗（パスワード不正）
		if auditErr := h.svc.Audit.LogAuthLogin(ctx, nil, &account.ID, model.AuditActionAuthLoginFailure, clientIP, userAgent); auditErr != nil {
			slog.ErrorContext(ctx, "failed to log login failure", slog.String("error", auditErr.Error()))
		}
		return nil, nil, apperrors.WrapUnauthorized("メールアドレスまたはパスワードが正しくありません")
	}
	staff, err := h.svc.Staff.FindByAccountID(ctx, account.ID)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "スタッフ情報の取得に失敗しました")
	}
	// BUG-134: スタッフの有効性チェック（退職者・一時停止アカウント防止）
	if !staff.IsActive {
		return nil, nil, apperrors.WrapUnauthorized("このアカウントは無効です")
	}
	return account, staff, nil
}

// resolveClinicInfo はスタッフのクリニック割り当てからメインクリニックIDと所属ID一覧を返す。
// メインクリニックが未設定の場合は最初の割り当てを使用する。
func resolveClinicInfo(assignments []model.StaffClinicAssignment) (mainClinicID string, clinicIDs []uint64) {
	clinicIDs = make([]uint64, 0, len(assignments))
	for _, asg := range assignments {
		if asg.IsMain && mainClinicID == "" {
			mainClinicID = strconv.FormatUint(asg.ClinicID, 10)
		}
		clinicIDs = append(clinicIDs, asg.ClinicID)
	}
	if mainClinicID == "" && len(assignments) > 0 {
		mainClinicID = strconv.FormatUint(assignments[0].ClinicID, 10)
	}
	return mainClinicID, clinicIDs
}

// issueAuthCookies は JWT アクセストークン（15分）とリフレッシュトークン（7日）を生成して Cookie にセットする。
// クロスオリジン対応のため SameSite=None + Secure=true を使用する。
func (h *Handler) issueAuthCookies(c *gin.Context, staffID uint64, mainClinicID string, isSystemAdmin bool, clinicIDs []uint64) error {
	// BUG-TEST-002: 開発環境（HTTP）での Cookie 受理を許可するため、モードに応じて Secure/SameSite を切り替える
	isProduction := h.cfg.GinMode == "release"
	sameSite := http.SameSiteLaxMode
	secure := false

	if isProduction {
		sameSite = http.SameSiteNoneMode
		secure = true
	}

	expiresAt := time.Now().Add(15 * time.Minute)
	claims := &middleware.JWTClaims{
		UserID:        strconv.FormatUint(staffID, 10),
		ClinicID:      mainClinicID,
		IsSystemAdmin: isSystemAdmin,
		ClinicIDs:     clinicIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessTokenStr, err := accessToken.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return apperrors.Wrap(err, "failed to sign jwt")
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    accessTokenStr,
		Path:     "/",
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	// Refresh Token 発行（7日間有効、token rotation で毎回更新）
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := &middleware.JWTClaims{
		UserID:        strconv.FormatUint(staffID, 10),
		ClinicID:      mainClinicID,
		IsSystemAdmin: isSystemAdmin,
		ClinicIDs:     clinicIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "refresh",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err := refreshToken.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return apperrors.Wrap(err, "failed to sign refresh token")
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshTokenStr,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   int(time.Until(refreshExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
	return nil
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

	assignments, err := h.svc.StaffClinicAssignment.FindByStaffID(ctx, staff.ID)
	if err != nil {
		RespondError(c, apperrors.Wrap(err, "所属クリニック情報の取得に失敗しました"))
		return
	}
	staff.ClinicAssignments = assignments

	mainClinicID, clinicIDs := resolveClinicInfo(assignments)

	if err := h.issueAuthCookies(c, staff.ID, mainClinicID, account.IsSystemAdmin, clinicIDs); err != nil {
		RespondError(c, err)
		return
	}

	// クリニック一覧を取得してレスポンス構築
	allClinics, err := h.svc.Clinic.ListClinics(ctx)
	if err != nil {
		allClinics = nil
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
		if auditErr := h.svc.Audit.LogAuthLogin(ctx, &mainCID, &staff.ID, model.AuditActionAuthLoginSuccess, c.ClientIP(), c.Request.Header.Get("User-Agent")); auditErr != nil {
			slog.ErrorContext(ctx, "failed to log login success", slog.String("error", auditErr.Error()))
		}
	}

	c.JSON(http.StatusOK, LoginResponse{
		IsSystemAdmin: account.IsSystemAdmin,
		User:          buildMeResponse(staff, account, mainClinicID, clinicNameMap, allClinics, permMap),
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
						if auditErr := h.svc.Audit.LogAuthLogin(ctx, &clinicID, &staffID, model.AuditActionAuthLogout, c.ClientIP(), c.Request.Header.Get("User-Agent")); auditErr != nil {
							slog.ErrorContext(ctx, "failed to log logout", slog.String("error", auditErr.Error()))
						}
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
	assignments, asgErr := h.svc.StaffClinicAssignment.FindByStaffID(ctx, staff.ID)
	if asgErr != nil {
		RespondError(c, apperrors.Wrap(asgErr, "failed to get clinic assignments"))
		return
	}
	clinicIDs := make([]uint64, 0, len(assignments))
	mainClinicID := claims.ClinicID
	for _, a := range assignments {
		clinicIDs = append(clinicIDs, a.ClinicID)
	}

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

// ChangeMyPassword は認証済みユーザーが自分のパスワードを変更する（BUG-148）
func (h *Handler) ChangeMyPassword(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password"     binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// 新しいパスワードの複雑性チェック
	if err := validatePassword(req.NewPassword); err != nil {
		RespondError(c, err)
		return
	}

	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}
	staff, err := h.svc.Staff.GetByID(ctx, staffID)
	if err != nil {
		RespondError(c, err)
		return
	}
	if staff.AccountID == nil {
		RespondError(c, apperrors.WrapInvalidInput("このスタッフにはアカウントが紐づいていません"))
		return
	}

	account, err := h.svc.Account.GetByID(ctx, *staff.AccountID)
	if err != nil {
		RespondError(c, err)
		return
	}

	// 現在のパスワード検証
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		RespondError(c, apperrors.WrapUnauthorized("現在のパスワードが正しくありません"))
		return
	}

	// 新しいパスワードをハッシュ化して更新
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), config.BcryptCost)
	if err != nil {
		RespondError(c, apperrors.Wrap(err, "failed to hash password"))
		return
	}
	if err := h.svc.Account.UpdatePasswordHash(ctx, *staff.AccountID, string(hashed)); err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "パスワードを変更しました"})
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
	assignments, err := h.svc.StaffClinicAssignment.FindByStaffID(ctx, staff.ID)
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
	permMap := h.calculateEffectivePermissions(ctx, isSystemAdmin, staff.ID)

	c.JSON(http.StatusOK, buildMeResponse(staff, account, mainClinicIDStr, clinicNameMap, allClinics, permMap))
}

// buildAllPermissions は全リソースに対して全CRUD true のマップを返す。
// is_system_admin=true 用。
func buildAllPermissions() EffectivePermissions {
	m := make(EffectivePermissions, len(model.AllResources))
	for _, res := range model.AllResources {
		m[string(res)] = ResourcePermission{View: true, Create: true, Edit: true, Delete: true}
	}
	return m
}

// calculateEffectivePermissions はユーザー種別に応じた実効権限を計算する。
// isSystemAdmin=true: 全リソース全権限バイパス
// isSystemAdmin=false: DB の staff_permission_groups → permission_group_rules から UNION 計算
func (h *Handler) calculateEffectivePermissions(ctx context.Context, isSystemAdmin bool, staffID uint64) EffectivePermissions {
	// system_admin は全権限バイパス
	if isSystemAdmin {
		return buildAllPermissions()
	}

	// staff: DB から実効権限を取得
	rules, err := h.repos.PermissionGroup.GetEffectivePermissionsByStaffID(ctx, staffID)
	if err != nil {
		// エラー時は空の権限（最小権限の原則）
		return make(EffectivePermissions)
	}

	permMap := make(EffectivePermissions, len(rules))
	for _, rule := range rules {
		permMap[rule.Resource] = ResourcePermission{
			View:   rule.CanView,
			Create: rule.CanCreate,
			Edit:   rule.CanEdit,
			Delete: rule.CanDelete,
		}
	}
	return permMap
}

// ---- パスワードリセット ----

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordRequest struct {
	Token    string `json:"token"    binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// ForgotPassword はパスワードリセットメールを送信する。
// アカウントが存在しない場合も 200 を返す（メール存在有無の漏洩防止）。
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	ctx := c.Request.Context()
	if err := h.svc.PasswordReset.ForgotPassword(ctx, req.Email); err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent."})
}

// ResetPassword は rawToken と新パスワードでパスワードを更新する。
func (h *Handler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	if err := validatePassword(req.Password); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	ctx := c.Request.Context()
	if err := h.svc.PasswordReset.ResetPassword(ctx, req.Token, req.Password); err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully."})
}
