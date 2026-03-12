package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/animal-ekarte/backend/internal/middleware"
)

// devAdminUserID は dev モードで使用する固定ユーザー UUID
const devAdminUserID = "00000000-0000-0000-0000-000000000010"

// devAdminClinicID は dev モードで使用する固定クリニック UUID（002_seed_master.sql の渋谷院）
const devAdminClinicID = "00000000-0000-0000-0000-000000000001"

// allPermissions は管理者に付与する全権限リスト
var allPermissions = []string{
	"account_admin", "medical", "medical_read", "trimming",
	"billing", "reception", "hospitalization", "master_admin",
	"shift_admin", "inventory",
}

// MeClinicMembership は GET /me のクリニック所属情報
type MeClinicMembership struct {
	ClinicID   string `json:"clinic_id"`
	ClinicName string `json:"clinic_name"`
	BranchName string `json:"branch_name"`
	IsMain     bool   `json:"is_main"`
}

// MeResponse は GET /me のレスポンス（フロントエンド AuthUser と対応）
type MeResponse struct {
	ID           string                `json:"id"`
	Email        string                `json:"email"`
	DisplayName  string                `json:"display_name"`
	UserType     string                `json:"user_type"`
	JobTitle     *string               `json:"job_title"`
	AvatarURL    *string               `json:"avatar_url"`
	MainClinicID string                `json:"main_clinic_id"`
	Clinics      []MeClinicMembership  `json:"clinics"`
	Permissions  map[string][]string   `json:"permissions"`
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

// Login はメール/パスワードで認証してJWTトークンを返す
//
// 現在は DEV_ADMIN_EMAIL / DEV_ADMIN_PASSWORD 環境変数で設定した認証情報のみ受け付ける。
// TODO: UserAccount モデルに password_hash フィールドが存在しない。
// 本番実装時は以下の対応が必要:
//  1. user_accounts テーブルに password_hash TEXT NOT NULL カラムを追加
//  2. マイグレーションを作成
//  3. UserAccount モデル(internal/model/clinic.go)に PasswordHash フィールドを追加
//  4. UserAccountRepository に FindByEmail メソッドを追加
//  5. bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) で検証
func (h *Handler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Dev mode: 環境変数で設定した管理者認証情報でのみ認証を許可する。
	// DEV_ADMIN_EMAIL が空の場合は無効化される。
	// TODO: 本番実装では UserAccount をメールアドレスで検索し、
	// bcrypt でパスワードを検証すること。
	//  1. user_accounts テーブルに password_hash TEXT NOT NULL カラムを追加
	//  2. マイグレーションを作成
	//  3. UserAccount モデル(internal/model/clinic.go)に PasswordHash フィールドを追加
	//  4. UserAccountRepository に FindByEmail メソッドを追加
	//  5. bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) で検証
	if h.cfg.DevAdminEmail == "" || input.Email != h.cfg.DevAdminEmail || input.Password != h.cfg.DevAdminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	claims := &middleware.JWTClaims{
		UserID:   devAdminUserID,   // dev モード固定 UUID
		ClinicID: devAdminClinicID, // 002_seed_master.sql の渋谷院
		UserType: "clinic_admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to sign JWT",
			slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt.Unix(),
		UserType:  claims.UserType,
	})
}

// GetMe はJWTクレームからログインユーザー情報を返す。
// dev モードでは全クリニック・全権限を返す。
func (h *Handler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	userIDVal, _ := c.Get("user_id")
	clinicIDVal, _ := c.Get("clinic_id")
	userTypeVal, _ := c.Get("user_type")

	mainClinicIDStr := fmt.Sprintf("%v", clinicIDVal)

	clinics, err := h.svc.Clinic.ListClinics(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list clinics for /me",
			slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	meClinicList := make([]MeClinicMembership, 0, len(clinics))
	permissions := make(map[string][]string, len(clinics))
	for _, cl := range clinics {
		clIDStr := cl.ID.String()
		meClinicList = append(meClinicList, MeClinicMembership{
			ClinicID:   clIDStr,
			ClinicName: cl.Name,
			BranchName: cl.BranchName,
			IsMain:     clIDStr == mainClinicIDStr,
		})
		permissions[clIDStr] = allPermissions
	}

	c.JSON(http.StatusOK, MeResponse{
		ID:           fmt.Sprintf("%v", userIDVal),
		Email:        h.cfg.DevAdminEmail,
		DisplayName:  "管理者",
		UserType:     fmt.Sprintf("%v", userTypeVal),
		JobTitle:     nil,
		AvatarURL:    nil,
		MainClinicID: mainClinicIDStr,
		Clinics:      meClinicList,
		Permissions:  permissions,
	})
}
