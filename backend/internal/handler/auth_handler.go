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
		}
		break
	}

	var jobTitle *string
	if staff != nil && staff.JobTitle != nil {
		jt := staff.JobTitle.Name
		jobTitle = &jt
	}
	var staffRole *string
	if staff != nil {
		sr := string(staff.StaffRole)
		staffRole = &sr
	}

	staffID := uint64(0)
	if staff != nil {
		staffID = staff.ID
	}

	userType := "staff"
	if account != nil {
		userType = account.UserType
	}

	return &MeResponse{
		ID:           strconv.FormatUint(staffID, 10),
		Email:        account.Email,
		DisplayName:  staff.Name,
		UserType:     userType,
		StaffRole:    staffRole,
		JobTitle:     jobTitle,
		MainClinicID: mainClinicID,
		Clinic:       meClinic,
		Clinics:      meClinicList,
		Permissions:  permMap,
	}
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

	// accounts テーブルからアカウントを取得
	account, err := h.svc.Account.FindByEmail(ctx, input.Email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			RespondError(c, apperrors.WrapUnauthorized("メールアドレスまたはパスワードが正しくありません"))
			return
		}
		RespondError(c, err)
		return
	}

	// アカウント状態チェック
	if !account.IsActive {
		RespondError(c, apperrors.WrapUnauthorized("アカウントが無効です"))
		return
	}

	// パスワード検証
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(input.Password)); err != nil {
		slog.WarnContext(ctx, "login: password mismatch", slog.String("email", input.Email))
		RespondError(c, apperrors.WrapUnauthorized("メールアドレスまたはパスワードが正しくありません"))
		return
	}

	// staff を取得
	slog.InfoContext(ctx, "login: account found",
		slog.String("email", account.Email),
		slog.String("user_type", account.UserType),
		slog.Uint64("account_id", account.ID))

	staff, err := h.repos.Staff.FindByAccountID(ctx, account.ID)
	if err != nil {
		slog.ErrorContext(ctx, "login: failed to find staff", slog.String("error", err.Error()))
		RespondError(c, apperrors.Wrap(err, "スタッフ情報の取得に失敗しました"))
		return
	}

	// clinic assignments を取得
	assignments, err := h.svc.StaffClinicAssignment.FindByStaffID(ctx, staff.ID)
	if err != nil {
		slog.ErrorContext(ctx, "login: failed to find clinic assignments", slog.String("error", err.Error()))
		RespondError(c, apperrors.Wrap(err, "所属クリニック情報の取得に失敗しました"))
		return
	}
	staff.ClinicAssignments = assignments

	// メインクリニックを決定
	mainClinicID := ""
	for _, asg := range assignments {
		if asg.IsMain {
			mainClinicID = strconv.FormatUint(asg.ClinicID, 10)
			break
		}
	}
	if mainClinicID == "" && len(assignments) > 0 {
		mainClinicID = strconv.FormatUint(assignments[0].ClinicID, 10)
	}

	// アクセストークン生成（15分）
	expiresAt := time.Now().Add(15 * time.Minute)
	claims := &middleware.JWTClaims{
		UserID:   strconv.FormatUint(staff.ID, 10),
		ClinicID: mainClinicID,
		UserType: account.UserType,
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

	// Cookie 設定
	// クロスオリジン（localhost:3003 ← localhost:8080）でも Cookie が送信されるように
	// SameSite=None 使用時は Secure=true が必須（ブラウザ仕様）
	// localhost での Secure Cookie はブラウザが許可しているため、開発環境でも Secure=true を設定可
	sameSite := http.SameSiteNoneMode
	secure := true // SameSite=None の場合は常に true（localhost での Secure Cookie はブラウザが許可）

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    accessTokenStr,
		Path:     "/",
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	// クリニック一覧を取得してレスポンス構築
	allClinics, err := h.svc.Clinic.ListClinics(ctx)
	if err != nil {
		slog.WarnContext(ctx, "login: failed to list clinics", slog.String("error", err.Error()))
	}
	clinicNameMap := make(map[string]string)
	for i := range allClinics {
		cl := &allClinics[i]
		clinicNameMap[strconv.FormatUint(cl.ID, 10)] = cl.Name
	}

	// 実効権限を計算
	permMap := h.calculateEffectivePermissions(ctx, account.UserType, staff.ID)

	slog.InfoContext(ctx, "login successful", slog.String("email", account.Email), slog.Uint64("staff_id", staff.ID))

	c.JSON(http.StatusOK, LoginResponse{
		Token:     accessTokenStr,
		ExpiresAt: expiresAt.Unix(),
		UserType:  account.UserType,
		User:      buildMeResponse(staff, account, mainClinicID, clinicNameMap, allClinics, permMap),
	})
}

// Logout godoc
func (h *Handler) Logout(c *gin.Context) {
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

	slog.InfoContext(c.Request.Context(), "logout successful")
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
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
	staff, err := h.repos.Staff.FindByID(ctx, userID)
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
	userType := "staff"
	if account != nil {
		userType = account.UserType
	}
	permMap := h.calculateEffectivePermissions(ctx, userType, staff.ID)

	c.JSON(http.StatusOK, buildMeResponse(staff, account, mainClinicIDStr, clinicNameMap, allClinics, permMap))
}

// buildAllPermissions は全リソースに対して全CRUD true のマップを返す。
// system_admin / clinic_admin 用。
func buildAllPermissions() EffectivePermissions {
	m := make(EffectivePermissions, len(model.AllResources))
	for _, res := range model.AllResources {
		m[string(res)] = ResourcePermission{View: true, Create: true, Edit: true, Delete: true}
	}
	return m
}

// calculateEffectivePermissions はユーザー種別に応じた実効権限を計算する。
// system_admin / clinic_admin: 全リソース全権限
// staff: DB の staff_permission_groups → permission_group_rules から UNION 計算
func (h *Handler) calculateEffectivePermissions(ctx context.Context, userType string, staffID uint64) EffectivePermissions {
	// system_admin / clinic_admin は全権限バイパス
	if userType == string(model.UserTypeSystemAdmin) || userType == string(model.UserTypeClinicAdmin) {
		return buildAllPermissions()
	}

	// staff: DB から実効権限を取得
	rules, err := h.repos.PermissionGroup.GetEffectivePermissionsByStaffID(ctx, staffID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to calculate effective permissions",
			slog.String("error", err.Error()),
			slog.Uint64("staff_id", staffID))
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
