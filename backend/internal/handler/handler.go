package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/service"
)

// Handler はHTTPハンドラーのルートコンテナ
type Handler struct {
	cfg       *config.Config
	svc       *service.Services
	repos     *repository.Repositories
	auditRepo repository.AuditRepository
}

// New はHandlerを初期化して返す
func New(cfg *config.Config, svc *service.Services, repos *repository.Repositories) *Handler {
	return &Handler{cfg: cfg, svc: svc, repos: repos, auditRepo: repos.Audit}
}

// PaginatedResponse はページネーション付きレスポンスの共通構造
type PaginatedResponse[T any] struct {
	Data  T     `json:"data"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// newPaginatedResponse はPaginatedResponseを型推論で生成するヘルパー
func newPaginatedResponse[T any](data T, total int64, page, limit int) PaginatedResponse[T] {
	return PaginatedResponse[T]{Data: data, Total: total, Page: page, Limit: limit}
}

// Health はサーバーの稼働状態を返す
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RegisterRoutes はすべてのルートを登録する
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// Health check エンドポイント（ルートレベル）
	r.GET("/health", h.Health)

	// レートリミッター初期化
	rateLimitStore := middleware.NewRateLimitStore()

	api := r.Group("/api/v1")

	// ログイン・ログアウトのレート制限（10 req/min, burst 5）
	loginGroup := api.Group("")
	loginGroup.Use(middleware.RateLimit(rateLimitStore, 0.167, 5))
	loginGroup.POST("/login", h.Login)
	loginGroup.POST("/logout", h.Logout)
	loginGroup.POST("/auth/refresh", h.RefreshToken)
	loginGroup.POST("/auth/forgot-password", h.ForgotPassword)
	loginGroup.POST("/auth/reset-password", h.ResetPassword)

	protected := api.Group("")
	protected.Use(middleware.Auth(h.cfg.JWTSecret, h.repos.UserAccount))

	protected.GET("/me", h.GetMe)

	h.RegisterOwnerRoutes(protected)
	h.RegisterPetRoutes(protected)
	h.RegisterReservationRoutes(protected)
	h.registerMedicalRecordRoutesWithAuth(protected)
	h.RegisterHospitalizationRoutes(protected)
	h.registerAccountingRoutesWithAuth(protected)
	h.RegisterTrimmingRoutes(protected)
	h.RegisterExaminationRoutes(protected)
	h.RegisterVaccinationRoutes(protected)
	h.registerInventoryRoutesWithAuth(protected)
	h.registerMasterRoutesWithAuth(protected)
	h.RegisterClinicRoutes(protected)
	h.RegisterUserRoutes(protected)
	h.registerEstimateRoutesWithAuth(protected)
	h.RegisterShiftRoutes(protected)
	h.RegisterCompanyRoutes(protected)
	h.RegisterGlobalCheckupRoutes(protected)
	h.RegisterBillingItemRoutes(protected)
	h.registerPermissionGroupRoutesWithAuth(protected)
}

// registerMedicalRecordRoutesWithAuth はカルテルートにDELETE権限チェックを追加する
func (h *Handler) registerMedicalRecordRoutesWithAuth(rg *gin.RouterGroup) {
	records := rg.Group("/medical-records")
	records.GET("", h.ListMedicalRecords)
	records.POST("", h.CreateMedicalRecord)
	records.GET("/:id", h.GetMedicalRecord)
	records.PATCH("/:id", h.UpdateMedicalRecord)
	records.DELETE("/:id",
		middleware.RequirePermission(model.ResourceMedicalRecords, "delete", h.repos.PermissionGroup),
		h.DeleteMedicalRecord)

	h.RegisterVitalRoutes(records)
	h.RegisterTreatmentRoutes(records)
	h.RegisterBillingReviewRoutes(records)
	h.RegisterRecordImageRoutes(records)
	h.RegisterTreatmentPlanMedicalRecordRoutes(records)
	h.RegisterClinicalPlanRoutes(records)
	h.RegisterCheckupRoutes(records)
	h.RegisterInquiryRoutes(records)
}

// registerAccountingRoutesWithAuth は会計ルートにDELETE権限チェックを追加する
func (h *Handler) registerAccountingRoutesWithAuth(rg *gin.RouterGroup) {
	h.RegisterAccountingRoutes(rg)
}

// registerInventoryRoutesWithAuth は在庫ルートにDELETE権限チェックを追加する
func (h *Handler) registerInventoryRoutesWithAuth(rg *gin.RouterGroup) {
	h.RegisterInventoryRoutes(rg)
}

// registerMasterRoutesWithAuth はマスタルートに権限チェックを追加する
func (h *Handler) registerMasterRoutesWithAuth(rg *gin.RouterGroup) {
	h.RegisterMasterRoutes(rg)
}

// registerEstimateRoutesWithAuth は見積書ルートにDELETE権限チェックを追加する
func (h *Handler) registerEstimateRoutesWithAuth(rg *gin.RouterGroup) {
	h.RegisterEstimateRoutes(rg)
}

// registerPermissionGroupRoutesWithAuth は権限グループ管理ルートを認可ミドルウェア付きで登録する
func (h *Handler) registerPermissionGroupRoutesWithAuth(rg *gin.RouterGroup) {
	pg := rg.Group("/permission-groups")
	// 閲覧は全員許可
	pg.GET("", h.ListPermissionGroups)
	pg.GET("/:id", h.GetPermissionGroup)

	// 作成・更新・削除・ルール変更は clinic_admin 以上のみ
	adminPG := pg.Group("")
	adminPG.Use(middleware.RequireClinicAdmin())
	adminPG.POST("", h.CreatePermissionGroup)
	adminPG.PATCH("/:id", h.UpdatePermissionGroup)
	adminPG.DELETE("/:id", h.DeletePermissionGroup)
	adminPG.PUT("/:id/rules", h.SetPermissionGroupRules)
}
