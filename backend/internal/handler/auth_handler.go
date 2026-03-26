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

// MeClinicMembership は GET /me のクリニック所属情報
type MeClinicMembership struct {
	ClinicID   string `json:"clinic_id"`
	ClinicName string `json:"clinic_name"`
	IsMain     bool   `json:"is_main"`
}

// MeClinicInfo は GET /me のメイン医院詳細情報
type MeClinicInfo struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	PostalCode         string  `json:"postal_code"`
	Address            string  `json:"address"`
	PhoneNumber        string  `json:"phone_number"`
	FaxNumber          string  `json:"fax_number"`
	RegistrationNumber string  `json:"registration_number"`
	DirectorName       string  `json:"director_name"`
	Email              string  `json:"email"`
	Website            string  `json:"website"`
	LogoURL            *string `json:"logo_url"`
}

// ResourcePermission は1リソースのCRUD権限
type ResourcePermission struct {
	View   bool `json:"view"`
	Create bool `json:"create"`
	Edit   bool `json:"edit"`
	Delete bool `json:"delete"`
}

// ClinicEffectivePermissions は clinicID → resource → CRUD のネストマップ
// 例: {"1": {"accounting": {View: true, Create: true, Edit: false, Delete: false}}}
type ClinicEffectivePermissions = map[string]map[string]ResourcePermission

// allResources はフロントエンド側のページ識別子一覧（system_admin/clinic_admin 全権限バイパス用）
var allResources = []string{
	"dashboard", "owners", "reservations", "medical-records", "hospitalization",
	"trimming", "examinations", "accounting", "vaccinations", "checkups",
	"inventory", "estimates", "shifts", "master", "hospital-settings",
}

// MeResponse は GET /me のレスポンス（フロントエンド AuthUser と対応）
type MeResponse struct {
	ID           string                     `json:"id"`
	Email        string                     `json:"email"`
	DisplayName  string                     `json:"display_name"`
	UserType     string                     `json:"user_type"`
	StaffRole    *string                    `json:"staff_role"`
	JobTitle     *string                    `json:"job_title"`
	AvatarURL    *string                    `json:"avatar_url"`
	MainClinicID string                     `json:"main_clinic_id"`
	Clinic       *MeClinicInfo              `json:"clinic"`
	Clinics      []MeClinicMembership       `json:"clinics"`
	Permissions  ClinicEffectivePermissions `json:"permissions"`
}

// LoginInput はログインリクエストのボディ
type LoginInput struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginResponse はログイン成功時のレスポンス
type LoginResponse struct {
	Token     string      `json:"token"` // JWT トークン（Authorization Bearer で送信）
	ExpiresAt int64       `json:"expires_at"`
	UserType  string      `json:"user_type"`
	User      *MeResponse `json:"user"`
}

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

	// 実効権限マップを構築する
	// system_admin / clinic_admin は全リソース全CRUD true（DB問い合わせ不要）
	permMap := make(ClinicEffectivePermissions)
	if data.UserAccount.UserType == model.UserTypeSystemAdmin || data.UserAccount.UserType == model.UserTypeClinicAdmin {
		for _, m := range data.Memberships {
			clIDStr := strconv.FormatUint(m.ClinicID, 10)
			permMap[clIDStr] = buildAllPermissionsForClinic()
		}
	} else {
		// staff はグループのUNIONで実効権限を計算（DBから取得済み）
		for _, row := range data.EffectivePermRows {
			clIDStr := strconv.FormatUint(row.ClinicID, 10)
			if permMap[clIDStr] == nil {
				permMap[clIDStr] = make(map[string]ResourcePermission)
			}
			permMap[clIDStr][row.Resource] = ResourcePermission{
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
		RespondError(c, apperrors.Wrap(err, "failed to sign jwt"))
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

	// JWT トークンはレスポンスボディに含める（Authorization Bearer ヘッダで管理）
	// フロントが sessionStorage に保存 → 各リクエストで Authorization ヘッダに含める
	// Cookie は使わない（cross-domain third-party Cookie ブロック対策）

	c.JSON(http.StatusOK, LoginResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt.Unix(),
		UserType:  claims.UserType,
		User:      buildMeResponse(userData, mainClinicID, clinicNameMap, allClinics),
	})
}

// Logout godoc
// フロント側で sessionStorage から token を削除するため、バック側では何もしない
func (h *Handler) Logout(c *gin.Context) {
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

// buildAllPermissionsForClinic は全リソースに対して全CRUD true のマップを返す。
// system_admin / clinic_admin はグループ設定に関係なく全権限を持つ。
func buildAllPermissionsForClinic() map[string]ResourcePermission {
	m := make(map[string]ResourcePermission, len(allResources))
	for _, res := range allResources {
		m[res] = ResourcePermission{View: true, Create: true, Edit: true, Delete: true}
	}
	return m
}
