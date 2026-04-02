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

	// Static file serving for uploaded images
	r.Static("/uploads", "/app/uploads")

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
	protected.Use(middleware.SanitizeNullBytes()) // BUG-067: NULL バイト・制御文字を除去

	protected.GET("/me", h.GetMe)

	// BUG-020: 各リソースの write 操作に権限チェックを適用
	h.registerOwnerRoutesWithAuth(protected)
	h.RegisterPetRoutes(protected)
	h.RegisterReservationRoutes(protected)
	h.registerMedicalRecordRoutesWithAuth(protected)
	h.registerHospitalizationRoutesWithAuth(protected)
	h.registerAccountingRoutesWithAuth(protected)
	h.registerTrimmingRoutesWithAuth(protected)
	h.registerExaminationRoutesWithAuth(protected)
	h.registerVaccinationRoutesWithAuth(protected)
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

// registerOwnerRoutesWithAuth は飼主ルートに create/edit/delete 権限チェックを追加する（BUG-020）
func (h *Handler) registerOwnerRoutesWithAuth(rg *gin.RouterGroup) {
	owners := rg.Group("/owners")
	owners.GET("", h.ListOwners)
	owners.GET("/:id", h.GetOwner)
	owners.POST("",
		middleware.RequirePermission(model.ResourceOwners, "create", h.repos.PermissionGroup),
		h.CreateOwner)
	owners.PATCH("/:id",
		middleware.RequirePermission(model.ResourceOwners, "edit", h.repos.PermissionGroup),
		h.UpdateOwner)
	owners.DELETE("/:id",
		middleware.RequirePermission(model.ResourceOwners, "delete", h.repos.PermissionGroup),
		h.DeleteOwner)
}

// registerMedicalRecordRoutesWithAuth はカルテルートに create/edit/delete 権限チェックを追加する（BUG-020）
func (h *Handler) registerMedicalRecordRoutesWithAuth(rg *gin.RouterGroup) {
	records := rg.Group("/medical-records")
	records.GET("", h.ListMedicalRecords)
	records.GET("/:id", h.GetMedicalRecord)
	records.POST("",
		middleware.RequirePermission(model.ResourceMedicalRecords, "create", h.repos.PermissionGroup),
		h.CreateMedicalRecord)
	records.PATCH("/:id",
		middleware.RequirePermission(model.ResourceMedicalRecords, "edit", h.repos.PermissionGroup),
		h.UpdateMedicalRecord)
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

// registerHospitalizationRoutesWithAuth は入院ルートに create/edit 権限チェックを追加する（BUG-020）
func (h *Handler) registerHospitalizationRoutesWithAuth(rg *gin.RouterGroup) {
	hospitalizations := rg.Group("/hospitalizations")
	hospitalizations.GET("", h.ListHospitalizations)
	hospitalizations.GET("/:id", h.GetHospitalization)
	hospitalizations.POST("",
		middleware.RequirePermission(model.ResourceHospitalization, "create", h.repos.PermissionGroup),
		h.CreateHospitalization)
	hospitalizations.PATCH("/:id",
		middleware.RequirePermission(model.ResourceHospitalization, "edit", h.repos.PermissionGroup),
		h.UpdateHospitalization)
	hospitalizations.DELETE("/:id",
		middleware.RequirePermission(model.ResourceHospitalization, "delete", h.repos.PermissionGroup),
		h.DeleteHospitalization)
	hospitalizations.POST("/:id/discharge-with-billing",
		middleware.RequirePermission(model.ResourceHospitalization, "edit", h.repos.PermissionGroup),
		h.DischargeWithBilling)
	h.RegisterDailyRecordRoutes(hospitalizations)
	h.RegisterCarePlanItemRoutes(hospitalizations)
	h.RegisterTreatmentPlanHospitalizationRoutes(hospitalizations)
}

// registerTrimmingRoutesWithAuth はトリミングルートに create/edit 権限チェックを追加する（BUG-020）
func (h *Handler) registerTrimmingRoutesWithAuth(rg *gin.RouterGroup) {
	trimmings := rg.Group("/trimmings")
	trimmings.GET("", h.ListTrimmings)
	trimmings.GET("/:id", h.GetTrimming)
	trimmings.POST("",
		middleware.RequirePermission(model.ResourceTrimming, "create", h.repos.PermissionGroup),
		h.CreateTrimming)
	trimmings.PATCH("/:id",
		middleware.RequirePermission(model.ResourceTrimming, "edit", h.repos.PermissionGroup),
		h.UpdateTrimming)
	trimmings.DELETE("/:id",
		middleware.RequirePermission(model.ResourceTrimming, "delete", h.repos.PermissionGroup),
		h.DeleteTrimming)
}

// registerExaminationRoutesWithAuth は検査ルートに create/edit 権限チェックを追加する（BUG-020）
func (h *Handler) registerExaminationRoutesWithAuth(rg *gin.RouterGroup) {
	examinations := rg.Group("/examinations")
	examinations.GET("", h.ListExaminations)
	examinations.GET("/:id", h.GetExamination)
	examinations.POST("",
		middleware.RequirePermission(model.ResourceExaminations, "create", h.repos.PermissionGroup),
		h.CreateExamination)
	examinations.PATCH("/:id",
		middleware.RequirePermission(model.ResourceExaminations, "edit", h.repos.PermissionGroup),
		h.UpdateExamination)
	examinations.DELETE("/:id",
		middleware.RequirePermission(model.ResourceExaminations, "delete", h.repos.PermissionGroup),
		h.DeleteExamination)
}

// registerVaccinationRoutesWithAuth はワクチンルートに create/edit 権限チェックを追加する（BUG-020）
func (h *Handler) registerVaccinationRoutesWithAuth(rg *gin.RouterGroup) {
	vaccinations := rg.Group("/vaccinations")
	vaccinations.GET("", h.ListVaccinations)
	vaccinations.GET("/:id", h.GetVaccination)
	vaccinations.POST("",
		middleware.RequirePermission(model.ResourceVaccinations, "create", h.repos.PermissionGroup),
		h.CreateVaccination)
	vaccinations.PATCH("/:id",
		middleware.RequirePermission(model.ResourceVaccinations, "edit", h.repos.PermissionGroup),
		h.UpdateVaccination)
	vaccinations.DELETE("/:id",
		middleware.RequirePermission(model.ResourceVaccinations, "delete", h.repos.PermissionGroup),
		h.DeleteVaccination)
}

// registerAccountingRoutesWithAuth は会計ルートに create/edit 権限チェックを追加する（BUG-020）
func (h *Handler) registerAccountingRoutesWithAuth(rg *gin.RouterGroup) {
	accountings := rg.Group("/accountings")
	accountings.GET("", h.ListAccountings)
	accountings.GET("/:id", h.GetAccounting)
	accountings.GET("/:id/refunds", h.ListRefunds)
	accountings.POST("",
		middleware.RequirePermission(model.ResourceAccounting, "create", h.repos.PermissionGroup),
		h.CreateAccounting)
	accountings.PATCH("/:id",
		middleware.RequirePermission(model.ResourceAccounting, "edit", h.repos.PermissionGroup),
		h.UpdateAccounting)
	accountings.DELETE("/:id",
		middleware.RequirePermission(model.ResourceAccounting, "delete", h.repos.PermissionGroup),
		h.DeleteAccounting)
	accountings.POST("/:id/refunds",
		middleware.RequirePermission(model.ResourceAccounting, "edit", h.repos.PermissionGroup),
		h.CreateRefund)
}

// registerInventoryRoutesWithAuth は在庫ルートに create/edit 権限チェックを追加する（BUG-020）
func (h *Handler) registerInventoryRoutesWithAuth(rg *gin.RouterGroup) {
	inventory := rg.Group("/inventory")
	inventory.GET("", h.ListInventory)
	inventory.GET("/:id", h.GetInventory)
	inventory.POST("",
		middleware.RequirePermission(model.ResourceInventory, "create", h.repos.PermissionGroup),
		h.CreateInventory)
	inventory.PATCH("/:id",
		middleware.RequirePermission(model.ResourceInventory, "edit", h.repos.PermissionGroup),
		h.UpdateInventory)
	inventory.DELETE("/:id",
		middleware.RequirePermission(model.ResourceInventory, "delete", h.repos.PermissionGroup),
		h.DeleteInventory)
}

// registerMasterRoutesWithAuth はマスタルートに権限チェックを追加する（BUG-020）
func (h *Handler) registerMasterRoutesWithAuth(rg *gin.RouterGroup) {
	h.RegisterMasterRoutes(rg)
}

// registerEstimateRoutesWithAuth は見積書ルートに create/edit 権限チェックを追加する（BUG-020）
func (h *Handler) registerEstimateRoutesWithAuth(rg *gin.RouterGroup) {
	estimates := rg.Group("/estimates")
	estimates.GET("", h.ListEstimates)
	estimates.GET("/:id", h.GetEstimate)
	estimates.POST("",
		middleware.RequirePermission(model.ResourceEstimates, "create", h.repos.PermissionGroup),
		h.CreateEstimate)
	estimates.PATCH("/:id",
		middleware.RequirePermission(model.ResourceEstimates, "edit", h.repos.PermissionGroup),
		h.UpdateEstimate)
	estimates.DELETE("/:id",
		middleware.RequirePermission(model.ResourceEstimates, "delete", h.repos.PermissionGroup),
		h.DeleteEstimate)
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
