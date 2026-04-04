package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/middleware"
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

	api := r.Group("/api/v1")

	// 認証関連（保護なし）
	api.POST("/login", h.Login)
	api.POST("/logout", h.Logout)

	protected := api.Group("")
	protected.Use(middleware.Auth(h.cfg.JWTSecret))
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
	h.registerEstimateRoutesWithAuth(protected)
	h.RegisterShiftRoutes(protected)
	h.RegisterCompanyRoutes(protected)
	h.RegisterGlobalCheckupRoutes(protected)
	h.RegisterBillingItemRoutes(protected)
}

// registerOwnerRoutesWithAuth は飼主ルートに create/edit/delete 権限チェックを追加する（BUG-020）
func (h *Handler) registerOwnerRoutesWithAuth(rg *gin.RouterGroup) {
	owners := rg.Group("/owners")
	owners.GET("", h.ListOwners)
	owners.GET("/:id", h.GetOwner)
	owners.POST("",
		h.CreateOwner)
	owners.PATCH("/:id",
		h.UpdateOwner)
	owners.DELETE("/:id",
		h.DeleteOwner)
}

// registerMedicalRecordRoutesWithAuth はカルテルートに create/edit/delete 権限チェックを追加する（BUG-020）
func (h *Handler) registerMedicalRecordRoutesWithAuth(rg *gin.RouterGroup) {
	records := rg.Group("/medical-records")
	records.GET("", h.ListMedicalRecords)
	records.GET("/:id", h.GetMedicalRecord)
	records.POST("",
		h.CreateMedicalRecord)
	records.PATCH("/:id",
		h.UpdateMedicalRecord)
	records.DELETE("/:id",
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
		h.CreateHospitalization)
	hospitalizations.PATCH("/:id",
		h.UpdateHospitalization)
	hospitalizations.DELETE("/:id",
		h.DeleteHospitalization)
	hospitalizations.POST("/:id/discharge-with-billing",
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
		h.CreateTrimming)
	trimmings.PATCH("/:id",
		h.UpdateTrimming)
	trimmings.DELETE("/:id",
		h.DeleteTrimming)
}

// registerExaminationRoutesWithAuth は検査ルートに create/edit 権限チェックを追加する（BUG-020）
func (h *Handler) registerExaminationRoutesWithAuth(rg *gin.RouterGroup) {
	examinations := rg.Group("/examinations")
	examinations.GET("", h.ListExaminations)
	examinations.GET("/:id", h.GetExamination)
	examinations.POST("",
		h.CreateExamination)
	examinations.PATCH("/:id",
		h.UpdateExamination)
	examinations.DELETE("/:id",
		h.DeleteExamination)
}

// registerVaccinationRoutesWithAuth はワクチンルートに create/edit 権限チェックを追加する（BUG-020）
func (h *Handler) registerVaccinationRoutesWithAuth(rg *gin.RouterGroup) {
	vaccinations := rg.Group("/vaccinations")
	vaccinations.GET("", h.ListVaccinations)
	vaccinations.GET("/:id", h.GetVaccination)
	vaccinations.POST("",
		h.CreateVaccination)
	vaccinations.PATCH("/:id",
		h.UpdateVaccination)
	vaccinations.DELETE("/:id",
		h.DeleteVaccination)
}

// registerAccountingRoutesWithAuth は会計ルートに create/edit 権限チェックを追加する（BUG-020）
func (h *Handler) registerAccountingRoutesWithAuth(rg *gin.RouterGroup) {
	accountings := rg.Group("/accountings")
	accountings.GET("", h.ListAccountings)
	accountings.GET("/:id", h.GetAccounting)
	accountings.GET("/:id/refunds", h.ListRefunds)
	accountings.POST("",
		h.CreateAccounting)
	accountings.PATCH("/:id",
		h.UpdateAccounting)
	accountings.DELETE("/:id",
		h.DeleteAccounting)
	accountings.POST("/:id/refunds",
		h.CreateRefund)
}

// registerInventoryRoutesWithAuth は在庫ルートに create/edit 権限チェックを追加する（BUG-020）
func (h *Handler) registerInventoryRoutesWithAuth(rg *gin.RouterGroup) {
	inventory := rg.Group("/inventory")
	inventory.GET("", h.ListInventory)
	inventory.GET("/:id", h.GetInventory)
	inventory.POST("",
		h.CreateInventory)
	inventory.PATCH("/:id",
		h.UpdateInventory)
	inventory.DELETE("/:id",
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
		h.CreateEstimate)
	estimates.PATCH("/:id",
		h.UpdateEstimate)
	estimates.DELETE("/:id",
		h.DeleteEstimate)
}

// registerPermissionGroupRoutesWithAuth は権限グループ管理ルートを認可ミドルウェア付きで登録する
