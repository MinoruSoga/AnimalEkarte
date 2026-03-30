package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

const (
	accessTokenCookieName  = "access_token"
	refreshTokenCookieName = "refresh_token"
	// 後方互換のため旧Cookie名も読み込む
	legacyCookieName = "auth_token"
)

// buildMeResponse はユーザーデータと補助情報からMeResponseを構築する。
// clinicNameMap はクリニックID（string）→クリニック名のマップ。
// mainClinicID はJWTクレームまたはログイン時のメインクリニックID（string）。
func buildMeResponse(data *repository.UserAccountWithMemberships, mainClinicID string, clinicNameMap map[string]string, allClinics []model.Clinic) *MeResponse {
	meClinicList := make([]MeClinicMembership, 0, len(data.Memberships))
	for _, m := range data.Memberships {
		clIDStr := strconv.FormatUint(m.ClinicID, 10)
		meClinicList = append(meClinicList, MeClinicMembership{
			ClinicID:   clIDStr,
			ClinicName: clinicNameMap[clIDStr],
			IsMain:     clIDStr == mainClinicID,
		})
	}

	// 実効権限マップを構築する（company単位のフラット構造）
	// system_admin / clinic_admin は全リソース全CRUD true（DB問い合わせ不要）
	permMap := make(EffectivePermissions)
	if data.UserAccount.UserType == model.UserTypeSystemAdmin || data.UserAccount.UserType == model.UserTypeClinicAdmin {
		permMap = buildAllPermissions()
	} else {
		// staff はグループのUNIONで実効権限を計算（DBから取得済み）
		for _, row := range data.EffectivePermRows {
			permMap[row.Resource] = ResourcePermission{
				View:   row.CanView,
				Create: row.CanCreate,
				Edit:   row.CanEdit,
				Delete: row.CanDelete,
			}
		}
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
		}
		break
	}

	var jobTitle *string
	if data.UserAccount.JobTitle != nil {
		jt := data.UserAccount.JobTitle.Name
		jobTitle = &jt
	}
	var staffRole *string
	if data.UserAccount.Staff != nil {
		sr := string(data.UserAccount.Staff.StaffRole)
		staffRole = &sr
	}
	var avatarURL *string
	if data.UserAccount.AvatarURL != "" {
		av := data.UserAccount.AvatarURL
		avatarURL = &av
	}

	return &MeResponse{
		ID:           strconv.FormatUint(data.UserAccount.ID, 10),
		Email:        data.UserAccount.Email,
		DisplayName:  data.UserAccount.DisplayName,
		UserType:     string(data.UserAccount.UserType),
		StaffRole:    staffRole,
		JobTitle:     jobTitle,
		AvatarURL:    avatarURL,
		MainClinicID: mainClinicID,
		Clinic:       meClinic,
		Clinics:      meClinicList,
		Permissions:  permMap,
	}
}

// Login godoc
// Login はメール/パスワードで認証してJWTトークンを返す。
// user_accounts.password_hash を bcrypt で検証する。
func (h *Handler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// DB からユーザーを取得
	account, err := h.svc.UserAccount.FindByEmail(ctx, input.Email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが正しくありません"})
			return
		}
		RespondError(c, err)
		return
	}

	// アカウント状態チェック
	if account.Status != "active" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "アカウントが無効です"})
		return
	}

	// パスワード検証
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(input.Password)); err != nil {
		h.writeAuditLog(c, model.AuditActionAuthLoginFailure, "auth", nil,
			nil, repository.MarshalAuditJSON(map[string]string{"email": input.Email}))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが正しくありません"})
		return
	}

	// 所属クリニックを取得してメインクリニックIDを決定
	memberships, mErr := h.svc.UserAccount.GetMemberships(ctx, account.ID)
	mainClinicID := ""
	if mErr == nil {
		for _, m := range memberships {
			if m.IsMain {
				mainClinicID = strconv.FormatUint(m.ClinicID, 10)
				break
			}
		}
		// isMain がなければ先頭を使う
		if mainClinicID == "" && len(memberships) > 0 {
			mainClinicID = strconv.FormatUint(memberships[0].ClinicID, 10)
		}
	}

	// アクセストークン: 15分
	expiresAt := time.Now().Add(15 * time.Minute)
	claims := &middleware.JWTClaims{
		UserID:   strconv.FormatUint(account.ID, 10),
		ClinicID: mainClinicID,
		UserType: string(account.UserType), //nolint:unconvert // model.UserType is a named string type; explicit cast required
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessTokenStr, err := accessToken.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		RespondError(c, apperrors.Wrap(err, "failed to sign jwt"))
		return
	}

	// リフレッシュトークン: 7日
	var mainClinicIDUint uint64
	if mainClinicID != "" {
		if parsed, parseErr := strconv.ParseUint(mainClinicID, 10, 64); parseErr == nil {
			mainClinicIDUint = parsed
		}
	}
	rawRefreshToken, err := h.svc.Auth.CreateRefreshToken(ctx, account.ID, mainClinicIDUint)
	if err != nil {
		RespondError(c, err)
		return
	}

	// ユーザー詳細情報を取得（ログインレスポンスに含めて /me 呼び出しを不要にする）
	userData, err := h.svc.UserAccount.GetWithMemberships(ctx, strconv.FormatUint(account.ID, 10))
	if err != nil {
		RespondError(c, err)
		return
	}
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

	// httpOnly Cookie でトークンをセット（XSS によるトークン盗難を防ぐ）
	// net/http.SetCookie を直接使用して SameSite を含む全属性を確実に設定する
	// （Gin の SetCookie は SameSite 非対応のため使わない）
	isProduction := h.cfg.GinMode == "release"
	sameSite := http.SameSiteLaxMode
	if isProduction {
		sameSite = http.SameSiteStrictMode
	}
	// アクセストークン Cookie（15分）
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    accessTokenStr,
		Path:     "/",
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})
	// リフレッシュトークン Cookie（7日、/api/v1/auth/refresh のみ送信）
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    rawRefreshToken,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})

	h.writeAuditLog(c, model.AuditActionAuthLoginSuccess, "auth", nil, nil, nil)
	c.JSON(http.StatusOK, LoginResponse{
		Token:     accessTokenStr, // 後方互換のため残す（Cookie 移行完了後に削除可）
		ExpiresAt: expiresAt.Unix(),
		UserType:  claims.UserType,
		User:      buildMeResponse(userData, mainClinicID, clinicNameMap, allClinics),
	})
}

// Logout godoc
// httpOnly Cookie を MaxAge=-1 でクリアする。リフレッシュトークンも無効化する。
func (h *Handler) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	isProduction := h.cfg.GinMode == "release"
	sameSite := http.SameSiteLaxMode
	if isProduction {
		sameSite = http.SameSiteStrictMode
	}

	// リフレッシュトークンを無効化する
	if rawRefreshToken, err := c.Cookie(refreshTokenCookieName); err == nil && rawRefreshToken != "" {
		_ = h.svc.Auth.RevokeRefreshToken(ctx, rawRefreshToken)
	}

	// アクセストークン Cookie をクリア
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})
	// 旧Cookie名もクリア（後方互換）
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     legacyCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})
	// リフレッシュトークン Cookie をクリア
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/api/v1/auth/refresh",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})
	h.writeAuditLog(c, model.AuditActionAuthLogout, "auth", nil, nil, nil)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// RefreshToken godoc
// POST /api/v1/auth/refresh — リフレッシュトークンを検証して新しいアクセストークンを発行する
func (h *Handler) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()
	isProduction := h.cfg.GinMode == "release"
	sameSite := http.SameSiteLaxMode
	if isProduction {
		sameSite = http.SameSiteStrictMode
	}

	rawRefreshToken, err := c.Cookie(refreshTokenCookieName)
	if err != nil || rawRefreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}

	userID, clinicID, newRawRefreshToken, err := h.svc.Auth.RefreshToken(ctx, rawRefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	// ユーザー情報を取得して新しいアクセストークンを発行
	userData, err := h.svc.UserAccount.GetWithMemberships(ctx, strconv.FormatUint(userID, 10))
	if err != nil {
		RespondError(c, err)
		return
	}

	// アカウント状態を確認
	if userData.UserAccount.Status != model.AccountStatusActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "account is disabled"})
		return
	}

	mainClinicID := strconv.FormatUint(clinicID, 10)
	expiresAt := time.Now().Add(15 * time.Minute)
	claims := &middleware.JWTClaims{
		UserID:   strconv.FormatUint(userID, 10),
		ClinicID: mainClinicID,
		UserType: string(userData.UserAccount.UserType),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessTokenStr, err := accessToken.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		RespondError(c, apperrors.Wrap(err, "failed to sign jwt"))
		return
	}

	// アクセストークン Cookie を更新
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    accessTokenStr,
		Path:     "/",
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})
	// ローテーション済みリフレッシュトークン Cookie を更新
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    newRawRefreshToken,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})

	c.JSON(http.StatusOK, gin.H{
		"expires_at": expiresAt.Unix(),
	})
}

// GetMe godoc
// GetMe はJWTクレームからログインユーザー情報を返す。
func (h *Handler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}
	mainClinicIDVal, _ := c.Get("clinic_id")
	mainClinicIDStr, _ := mainClinicIDVal.(string)

	data, err := h.svc.UserAccount.GetWithMemberships(ctx, userIDStr)
	if err != nil {
		RespondError(c, err)
		return
	}

	// クリニック情報（全クリニック名を解決するため Clinic サービス経由）
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

	c.JSON(http.StatusOK, buildMeResponse(data, mainClinicIDStr, clinicNameMap, allClinics))
}

// buildAllPermissions は全リソースに対して全CRUD true のマップを返す。
// system_admin / clinic_admin はグループ設定に関係なく全権限を持つ。
func buildAllPermissions() EffectivePermissions {
	m := make(EffectivePermissions, len(model.AllResources))
	for _, res := range model.AllResources {
		m[string(res)] = ResourcePermission{View: true, Create: true, Edit: true, Delete: true}
	}
	return m
}

// ForgotPassword godoc
// POST /api/v1/auth/forgot-password — パスワードリセットトークンを発行する
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	// セキュリティ: アカウントが存在しなくても同じレスポンスを返す
	_ = h.svc.Auth.ForgotPassword(c.Request.Context(), req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link has been sent"})
}

// ResetPassword godoc
// POST /api/v1/auth/reset-password — パスワードリセットトークンを使ってパスワードを更新する
func (h *Handler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	if err := h.svc.Auth.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password has been reset successfully"})
}

// ChangeMyPassword godoc
// PUT /api/v1/users/me/password — 自分のパスワードを変更する（BUG-062）
func (h *Handler) ChangeMyPassword(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	if err := h.svc.Auth.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}
