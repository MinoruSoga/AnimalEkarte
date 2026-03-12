package handler

import (
	"log/slog"
	"net/http"
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
	BranchName string `json:"branch_name"`
	IsMain     bool   `json:"is_main"`
}

// MeResponse は GET /me のレスポンス（フロントエンド AuthUser と対応）
type MeResponse struct {
	ID           string               `json:"id"`
	Email        string               `json:"email"`
	DisplayName  string               `json:"display_name"`
	UserType     string               `json:"user_type"`
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
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	UserType  string `json:"user_type"`
}

// Login godoc
// @Summary ログイン
// @Description メール/パスワードで認証してJWTトークンを返す
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginInput true "ログイン情報"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
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
	if mErr == nil {
		for _, m := range memberships {
			if m.IsMain {
				mainClinicID = m.ClinicID.String()
				break
			}
		}
		// isMain がなければ先頭を使う
		if mainClinicID == "" && len(memberships) > 0 {
			mainClinicID = memberships[0].ClinicID.String()
		}
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	claims := &middleware.JWTClaims{
		UserID:   account.ID.String(),
		ClinicID: mainClinicID,
		UserType: string(account.UserType),
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

	c.JSON(http.StatusOK, LoginResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt.Unix(),
		UserType:  claims.UserType,
	})
}

// GetMe godoc
// @Summary ログインユーザー情報取得
// @Description JWTクレームからログインユーザーの詳細情報を返す
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MeResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /me [get]
// GetMe はJWTクレームからログインユーザー情報を返す。
func (h *Handler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	userIDVal, _ := c.Get("user_id")
	mainClinicIDVal, _ := c.Get("clinic_id")
	mainClinicIDStr, _ := mainClinicIDVal.(string)

	data, err := h.svc.UserAccount.GetWithMemberships(ctx, userIDVal.(string))
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
	clinicNameMap := make(map[string]struct{ Name, BranchName string }, len(allClinics))
	for _, cl := range allClinics {
		clinicNameMap[cl.ID.String()] = struct{ Name, BranchName string }{cl.Name, cl.BranchName}
	}

	meClinicList := make([]MeClinicMembership, 0, len(data.Memberships))
	for _, m := range data.Memberships {
		clIDStr := m.ClinicID.String()
		info := clinicNameMap[clIDStr]
		meClinicList = append(meClinicList, MeClinicMembership{
			ClinicID:   clIDStr,
			ClinicName: info.Name,
			BranchName: info.BranchName,
			IsMain:     clIDStr == mainClinicIDStr,
		})
	}

	// 権限マップ: clinic_id → []permission
	permMap := make(map[string][]string)
	for _, p := range data.Permissions {
		clIDStr := p.ClinicID.String()
		permMap[clIDStr] = append(permMap[clIDStr], string(p.Permission))
	}

	var jobTitle *string
	if data.JobTitle != nil {
		jt := data.JobTitle.Name
		jobTitle = &jt
	}
	var avatarURL *string
	if data.UserAccount.AvatarURL != "" {
		av := data.UserAccount.AvatarURL
		avatarURL = &av
	}

	c.JSON(http.StatusOK, MeResponse{
		ID:           data.UserAccount.ID.String(),
		Email:        data.UserAccount.Email,
		DisplayName:  data.UserAccount.DisplayName,
		UserType:     string(data.UserAccount.UserType),
		JobTitle:     jobTitle,
		AvatarURL:    avatarURL,
		MainClinicID: mainClinicIDStr,
		Clinics:      meClinicList,
		Permissions:  permMap,
	})
}
