package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/middleware"
)

// MeClinicMembership は GET /me のクリニック所属情報
type MeClinicMembership struct {
	ClinicID   string `json:"clinic_id"`
	ClinicName string `json:"clinic_name"`
	IsMain     bool   `json:"is_main"`
}

// MeResponse は GET /me のレスポンス（フロントエンド AuthUser と対応）
type MeResponse struct {
	ID           string               `json:"id"`
	Email        string               `json:"email"`
	DisplayName  string               `json:"display_name"`
	UserType     string               `json:"user_type"`
	StaffRole    *string              `json:"staff_role"`
	JobTitle     *string              `json:"job_title"`
	AvatarURL    *string              `json:"avatar_url"`
	MainClinicID string               `json:"main_clinic_id"`
	Clinics      []MeClinicMembership `json:"clinics"`
	Permissions  map[string][]string  `json:"permissions"`
}

// LoginInput はログインリクエストのボディ
type LoginInput struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginResponse はログイン成功時のレスポンス
type LoginResponse struct {
	ExpiresAt int64       `json:"expires_at"`
	UserType  string      `json:"user_type"`
	User      *MeResponse `json:"user"`
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
		slog.ErrorContext(ctx, "failed to find user account",
			slog.String("email", input.Email),
			slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// アカウント状態チェック
	if account.Status != "active" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "アカウントが無効です"})
		return
	}

	// パスワード検証
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが正しくありません"})
		return
	}

	// 所属クリニックを取得してメインクリニックIDを決定
	memberships, mErr := h.svc.UserAccount.GetMemberships(ctx, account.ID)
	mainClinicID := ""
	if mErr != nil {
		slog.WarnContext(ctx, "failed to get memberships, token will have empty clinic_id",
			slog.Uint64("user_id", account.ID),
			slog.String("error", mErr.Error()))
	} else {
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

	expiresAt := time.Now().Add(24 * time.Hour)
	claims := &middleware.JWTClaims{
		UserID:   strconv.FormatUint(account.ID, 10),
		ClinicID: mainClinicID,
		UserType: string(account.UserType), //nolint:unconvert // model.UserType is a named string type; explicit cast required
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		slog.ErrorContext(ctx, "failed to sign JWT", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// ユーザー詳細情報を取得（ログインレスポンスに含めて /me 呼び出しを不要にする）
	userData, err := h.svc.UserAccount.GetWithMemberships(ctx, strconv.FormatUint(account.ID, 10))
	if err != nil {
		slog.ErrorContext(ctx, "failed to get user data for login response", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	allClinics, err := h.svc.Clinic.ListClinics(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list clinics for login response", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	clinicNameMap := make(map[string]string, len(allClinics))
	for _, cl := range allClinics {
		clinicNameMap[strconv.FormatUint(cl.ID, 10)] = cl.Name
	}
	meClinicList := make([]MeClinicMembership, 0, len(userData.Memberships))
	for _, m := range userData.Memberships {
		clIDStr := strconv.FormatUint(m.ClinicID, 10)
		info := clinicNameMap[clIDStr]
		meClinicList = append(meClinicList, MeClinicMembership{
			ClinicID:   clIDStr,
			ClinicName: info,
			IsMain:     clIDStr == mainClinicID,
		})
	}
	permMap := make(map[string][]string)
	for _, p := range userData.Permissions {
		clIDStr := strconv.FormatUint(p.ClinicID, 10)
		permMap[clIDStr] = append(permMap[clIDStr], string(p.Permission))
	}
	var jobTitle *string
	if userData.UserAccount.JobTitle != nil {
		jt := userData.UserAccount.JobTitle.Name
		jobTitle = &jt
	}
	var staffRole *string
	if userData.UserAccount.Staff != nil {
		sr := string(userData.UserAccount.Staff.StaffRole)
		staffRole = &sr
	}
	var avatarURL *string
	if userData.UserAccount.AvatarURL != "" {
		av := userData.UserAccount.AvatarURL
		avatarURL = &av
	}
	meResp := &MeResponse{
		ID:           strconv.FormatUint(userData.UserAccount.ID, 10),
		Email:        userData.UserAccount.Email,
		DisplayName:  userData.UserAccount.DisplayName,
		UserType:     string(userData.UserAccount.UserType),
		StaffRole:    staffRole,
		JobTitle:     jobTitle,
		AvatarURL:    avatarURL,
		MainClinicID: mainClinicID,
		Clinics:      meClinicList,
		Permissions:  permMap,
	}

	// httpOnly Cookie でトークンを保存（XSS攻撃でのトークン窃取を防止）
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenStr,
		HttpOnly: true,
		Secure:   gin.Mode() == gin.ReleaseMode,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   86400, // 24時間
	})

	c.JSON(http.StatusOK, LoginResponse{
		ExpiresAt: expiresAt.Unix(),
		UserType:  claims.UserType,
		User:      meResp,
	})
}

// Logout godoc
func (h *Handler) Logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		HttpOnly: true,
		Secure:   gin.Mode() == gin.ReleaseMode,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
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
		slog.ErrorContext(ctx, "failed to get user account", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// クリニック情報（全クリニック名を解決するため Clinic サービス経由）
	allClinics, err := h.svc.Clinic.ListClinics(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list clinics for /me", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	clinicNameMap := make(map[string]string, len(allClinics))
	for i := range allClinics {
		cl := &allClinics[i]
		clinicNameMap[strconv.FormatUint(cl.ID, 10)] = cl.Name
	}

	meClinicList := make([]MeClinicMembership, 0, len(data.Memberships))
	for _, m := range data.Memberships {
		clIDStr := strconv.FormatUint(m.ClinicID, 10)
		meClinicList = append(meClinicList, MeClinicMembership{
			ClinicID:   clIDStr,
			ClinicName: clinicNameMap[clIDStr],
			IsMain:     clIDStr == mainClinicIDStr,
		})
	}

	// 権限マップ: clinic_id → []permission
	permMap := make(map[string][]string)
	for _, p := range data.Permissions {
		clIDStr := strconv.FormatUint(p.ClinicID, 10)
		permMap[clIDStr] = append(permMap[clIDStr], string(p.Permission))
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

	c.JSON(http.StatusOK, MeResponse{
		ID:           strconv.FormatUint(data.UserAccount.ID, 10),
		Email:        data.UserAccount.Email,
		DisplayName:  data.UserAccount.DisplayName,
		UserType:     string(data.UserAccount.UserType),
		StaffRole:    staffRole,
		JobTitle:     jobTitle,
		AvatarURL:    avatarURL,
		MainClinicID: mainClinicIDStr,
		Clinics:      meClinicList,
		Permissions:  permMap,
	})
}
