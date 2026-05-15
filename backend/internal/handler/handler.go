package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/infra"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/service"
)

// Handler はHTTPハンドラーのルートコンテナ
type Handler struct {
	cfg      *config.Config
	svc      *service.Services
	repos    *repository.Repositories
	uploader infra.FileUploader
}

// New はHandlerを初期化して返す
func New(cfg *config.Config, svc *service.Services, repos *repository.Repositories, uploader infra.FileUploader) *Handler {
	return &Handler{cfg: cfg, svc: svc, repos: repos, uploader: uploader}
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

// RegisterRoutes はすべてのルートを登録する。
// ctx はバックグラウンドゴルーチン（rate limiter cleanup 等）のライフタイム管理に使用する。
func (h *Handler) RegisterRoutes(ctx context.Context, r *gin.Engine) {
	// Health check エンドポイント（ルートレベル）
	r.GET("/health", h.Health)

	// Static file serving for uploaded images（BUG-141: ディレクトリリスティング無効化）
	r.StaticFS("/uploads", gin.Dir("/app/uploads", false))

	api := r.Group("/api/v1")

	// 認証関連（保護なし）— ログインにはレートリミット適用（BUG-130: ブルートフォース対策）
	loginRateStore := middleware.NewRateLimitStore(ctx)
	// パスワードリセット系も独立したストアでレートリミット（3回/分・バースト3）
	passwordRateStore := middleware.NewRateLimitStore(ctx)
	api.POST("/login", middleware.RateLimit(loginRateStore, 5.0/60, 5), h.Login) // 5回/分
	api.POST("/logout", h.Logout)
	api.POST("/auth/refresh", h.RefreshToken) // BUG-136: refresh token エンドポイント
	api.POST("/auth/forgot-password", middleware.RateLimit(passwordRateStore, 3.0/60, 3), h.ForgotPassword)
	api.POST("/auth/reset-password", middleware.RateLimit(passwordRateStore, 3.0/60, 3), h.ResetPassword)

	protected := api.Group("")
	var auditSvc service.AuditService
	if h.svc != nil {
		auditSvc = h.svc.Audit
	}
	protected.Use(middleware.Auth(h.cfg.JWTSecret, h.cfg.GinMode == "release", auditSvc))
	// NOTE: SanitizeNullBytes は main.go でグローバル登録済み（BUG-067）

	protected.GET("/me", h.GetMe)
	protected.PUT("/users/me/password", h.ChangeMyPassword) // BUG-148: 自分のパスワード変更

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
	h.RegisterMasterRoutes(protected)
	h.RegisterClinicRoutes(protected)
	h.registerEstimateRoutesWithAuth(protected)
	h.RegisterShiftRoutes(protected)
	h.RegisterShiftTemplateRoutes(protected)
	h.RegisterClinicHolidayRoutes(protected)
	h.RegisterCompanyRoutes(protected)
	h.RegisterGlobalCheckupRoutes(protected)
	h.RegisterBillingItemRoutes(protected)
	h.RegisterLineReservationRoutes(protected)
	// FEAT-368: 集計・締め機能
	h.RegisterClosingSettingsRoutes(protected)
	h.RegisterCashRegisterRoutes(protected)
	h.RegisterAccountingReportRoutes(protected)
	h.RegisterPaymentMethodMasterRoutes(protected)

	// LSTEP / LINE連携
	h.RegisterLstepSettingsRoutes(protected)
	h.RegisterSharedFileRoutes(protected)
	h.RegisterAggregationRoutes(protected)
	// LSTEP-BE-020: タグ集計・タグ別飼い主検索
	h.RegisterLstepTagSummaryRoutes(protected)
	// LSTEP-BE-004: 健診対象者抽出・一括タグ連携
	h.RegisterCheckupSyncRoutes(protected)
	// Q23: トリガー優先順位設定
	h.RegisterLstepTriggerPriorityRoutes(protected)
	// FEAT-379: タグコードマッピング設定
	h.RegisterLstepTagCodeMappingRoutes(protected)
	// 自動管理タグプレフィックス・条件タグ・送信目的タグ設定
	h.RegisterLstepTagConfigRoutes(protected)
	// FEAT-384: 自動配信トリガー監視
	h.RegisterLstepDeliveryMonitorRoutes(protected)
	// FEAT-385: Lステップ CSV インポート・分析
	h.RegisterLstepCsvImportRoutes(protected)
	h.RegisterLstepAnalyticsRoutes(protected)

	// LIFF公開API（JWT認証なし・LINE IDトークン認証）
	h.RegisterLiffRoutes(r)

	// BE-021: LINE Webhook（JWT認証なし・HMAC-SHA256署名検証）
	r.POST("/api/line/webhook", h.ReceiveLineWebhook)
}

// registerOwnerRoutesWithAuth は飼主ルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerOwnerRoutesWithAuth(rg *gin.RouterGroup) {
	owners := rg.Group("/owners")
	owners.GET("", h.RequirePermission(string(model.ResourceOwners), "view"), h.ListOwners)
	owners.GET("/:id", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetOwner)
	owners.POST("", h.RequirePermission(string(model.ResourceOwners), "create"), h.CreateOwner)
	owners.PATCH("/:id", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwner)
	owners.DELETE("/:id", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwner)
	// BE-005: LINE User ID 連携・解除
	owners.PATCH("/:id/line-user-id", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerLineUserID)
	// ISSUE-001: FE統一エンドポイントのエイリアス
	owners.PATCH("/:id/line", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerLineUserID)
	owners.DELETE("/:id/line", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwnerLine)
	// BE-017: Lステップオプトアウト・オプトイン（旧エンドポイント互換保持）
	owners.POST("/:id/lstep-opt-out", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PostOwnerLstepOptOut)
	owners.DELETE("/:id/lstep-opt-out", h.RequirePermission(string(model.ResourceOwners), "edit"), h.DeleteOwnerLstepOptOut)
	// ISSUE-001: 統合opt-outエンドポイント
	owners.PATCH("/:id/lstep/opt-out", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerLstepOptOut)
	// FEAT-381: 配信除外・転院・LINE ID確認
	owners.PATCH("/:id/delivery-exclusion", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerDeliveryExclusion)
	// FEAT-381-2: 配信注意フラグ
	owners.PATCH("/:id/delivery-caution", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerDeliveryCaution)
	owners.PATCH("/:id/transfer-status", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerTransferStatus)
	owners.PATCH("/:id/line-id-confirm", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerLineIDConfirm)
	// BE-019: Lステップタグ CRUD
	owners.GET("/:id/lstep/tags", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetOwnerLstepTags)
	owners.POST("/:id/lstep/tags", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PostOwnerLstepTag)
	owners.DELETE("/:id/lstep/tags/:tag_name", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwnerLstepTag)
	// BE-013: LINE個別送信
	h.RegisterLineSendRoutes(owners)
	// BE-021: LINE User ID 自動取得・飼い主紐付けトークン発行
	owners.POST("/:id/line/link-token", h.RequirePermission(string(model.ResourceOwners), "edit"), h.GenerateLineLinkToken)

	// ISSUE-001/ISSUE-002: FE が /clinics/:clinic_id/owners/:id/... プレフィックスで呼ぶルートのエイリアス
	// extractClinicID は JWT コンテキストから clinic_id を取得するため :clinic_id URL パラムは無視される
	co := rg.Group("/clinics/:clinic_id/owners")
	co.PATCH("/:id/line-user-id", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerLineUserID)
	co.PATCH("/:id/delivery-exclusion", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerDeliveryExclusion)
	co.PATCH("/:id/delivery-caution", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerDeliveryCaution)
	co.PATCH("/:id/transfer-status", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerTransferStatus)
	co.PATCH("/:id/line-id-confirm", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerLineIDConfirm)
	co.POST("/:id/lstep-opt-out", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PostOwnerLstepOptOut)
	co.DELETE("/:id/lstep-opt-out", h.RequirePermission(string(model.ResourceOwners), "edit"), h.DeleteOwnerLstepOptOut)
	co.GET("/:id/lstep/tags", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetOwnerLstepTags)
	co.POST("/:id/lstep/tags", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PostOwnerLstepTag)
	co.DELETE("/:id/lstep/tags/:tag_name", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwnerLstepTag)
	co.POST("/:id/line/send", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PostLineSend)
	co.GET("/:id/line/send-logs", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetLineSendLogs)
	// FEAT-385: 飼主の最新 Lステップ友だち属性
	co.GET("/:id/lstep/friend-attributes", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepOwnerFriendAttributes)
}

// registerMedicalRecordRoutesWithAuth はカルテルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerMedicalRecordRoutesWithAuth(rg *gin.RouterGroup) {
	records := rg.Group("/medical-records")
	records.GET("", h.RequirePermission(string(model.ResourceMedicalRecords), "view"), h.ListMedicalRecords)
	records.GET("/:id", h.RequirePermission(string(model.ResourceMedicalRecords), "view"), h.GetMedicalRecord)
	records.POST("", h.RequirePermission(string(model.ResourceMedicalRecords), "create"), h.CreateMedicalRecord)
	records.PATCH("/:id", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.UpdateMedicalRecord)
	records.DELETE("/:id", h.RequirePermission(string(model.ResourceMedicalRecords), "delete"), h.DeleteMedicalRecord)
	records.PATCH("/:id/recommendation-reason", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.PatchMedicalRecordRecommendationReason)

	h.RegisterVitalRoutes(records)
	h.RegisterTreatmentRoutes(records)
	h.RegisterBillingConfirmationRoutes(records)
	h.RegisterMedicalRecordImageRoutes(records)
	h.RegisterTreatmentPlanMedicalRecordRoutes(records)
	h.RegisterClinicalPlanRoutes(records)
	h.RegisterCheckupRoutes(records)
	h.RegisterPrescriptionRoutes(records)
	h.RegisterInquiryRoutes(records)
	h.RegisterMedicalRecordAddendumRoutes(records)
}

// registerHospitalizationRoutesWithAuth は入院ルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerHospitalizationRoutesWithAuth(rg *gin.RouterGroup) {
	hospitalizations := rg.Group("/hospitalizations")
	hospitalizations.GET("", h.RequirePermission(string(model.ResourceHospitalization), "view"), h.ListHospitalizations)
	hospitalizations.GET("/:id", h.RequirePermission(string(model.ResourceHospitalization), "view"), h.GetHospitalization)
	hospitalizations.POST("", h.RequirePermission(string(model.ResourceHospitalization), "create"), h.CreateHospitalization)
	hospitalizations.PATCH("/:id", h.RequirePermission(string(model.ResourceHospitalization), "edit"), h.UpdateHospitalization)
	hospitalizations.DELETE("/:id", h.RequirePermission(string(model.ResourceHospitalization), "delete"), h.DeleteHospitalization)
	hospitalizations.POST("/:id/discharge-with-billing", h.RequirePermission(string(model.ResourceHospitalization), "edit"), h.DischargeWithBilling)

	h.RegisterDailyRecordRoutes(hospitalizations)
	h.RegisterCarePlanItemRoutes(hospitalizations)
	h.RegisterTreatmentPlanHospitalizationRoutes(hospitalizations)
}

// registerTrimmingRoutesWithAuth はトリミングルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerTrimmingRoutesWithAuth(rg *gin.RouterGroup) {
	trimmings := rg.Group("/trimmings")
	trimmings.GET("", h.RequirePermission(string(model.ResourceTrimming), "view"), h.ListTrimmings)
	trimmings.GET("/:id", h.RequirePermission(string(model.ResourceTrimming), "view"), h.GetTrimming)
	trimmings.POST("", h.RequirePermission(string(model.ResourceTrimming), "create"), h.CreateTrimming)
	trimmings.PATCH("/:id", h.RequirePermission(string(model.ResourceTrimming), "edit"), h.UpdateTrimming)
	trimmings.DELETE("/:id", h.RequirePermission(string(model.ResourceTrimming), "delete"), h.DeleteTrimming)
}

// registerExaminationRoutesWithAuth は検査ルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerExaminationRoutesWithAuth(rg *gin.RouterGroup) {
	examinations := rg.Group("/examinations")
	examinations.GET("", h.RequirePermission(string(model.ResourceExaminations), "view"), h.ListExaminations)
	examinations.GET("/:id", h.RequirePermission(string(model.ResourceExaminations), "view"), h.GetExamination)
	examinations.POST("", h.RequirePermission(string(model.ResourceExaminations), "create"), h.CreateExamination)
	examinations.PATCH("/:id", h.RequirePermission(string(model.ResourceExaminations), "edit"), h.UpdateExamination)
	examinations.DELETE("/:id", h.RequirePermission(string(model.ResourceExaminations), "delete"), h.DeleteExamination)

	// 検査項目（exam_results）— PUT 一括置換セマンティクス
	examinations.GET("/:id/items", h.RequirePermission(string(model.ResourceExaminations), "view"), h.ListExaminationItems)
	examinations.PUT("/:id/items", h.RequirePermission(string(model.ResourceExaminations), "edit"), h.ReplaceExaminationItems)
}

// registerVaccinationRoutesWithAuth はワクチンルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerVaccinationRoutesWithAuth(rg *gin.RouterGroup) {
	vaccinations := rg.Group("/vaccinations")
	vaccinations.GET("", h.RequirePermission(string(model.ResourceVaccinations), "view"), h.ListVaccinations)
	vaccinations.GET("/:id", h.RequirePermission(string(model.ResourceVaccinations), "view"), h.GetVaccination)
	vaccinations.POST("", h.RequirePermission(string(model.ResourceVaccinations), "create"), h.CreateVaccination)
	vaccinations.PATCH("/:id", h.RequirePermission(string(model.ResourceVaccinations), "edit"), h.UpdateVaccination)
	vaccinations.DELETE("/:id", h.RequirePermission(string(model.ResourceVaccinations), "delete"), h.DeleteVaccination)
}

// registerAccountingRoutesWithAuth は会計ルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerAccountingRoutesWithAuth(rg *gin.RouterGroup) {
	accountings := rg.Group("/accountings")
	accountings.GET("", h.RequirePermission(string(model.ResourceAccounting), "view"), h.ListAccountings)
	// BUG-370: 月末未納者一覧
	accountings.GET("/unpaid", h.RequirePermission(string(model.ResourceAccounting), "view"), h.ListUnpaidBillings)
	// BUG-368: レジ締め日次集計
	accountings.GET("/daily-summary", h.RequirePermission(string(model.ResourceAccounting), "view"), h.GetDailySummary)
	accountings.GET("/:id", h.RequirePermission(string(model.ResourceAccounting), "view"), h.GetAccounting)
	accountings.GET("/:id/refunds", h.RequirePermission(string(model.ResourceAccounting), "view"), h.ListRefunds)
	accountings.POST("", h.RequirePermission(string(model.ResourceAccounting), "create"), h.CreateAccounting)
	accountings.PATCH("/:id", h.RequirePermission(string(model.ResourceAccounting), "edit"), h.UpdateAccounting)
	// BUG-371: DELETE は廃止し論理削除 (POST /:id/cancel) に統合。status=cancelled に遷移させる。
	accountings.POST("/:id/cancel", h.RequirePermission(string(model.ResourceAccounting), "delete"), h.CancelAccounting)
	accountings.POST("/:id/refunds", h.RequirePermission(string(model.ResourceAccounting), "create"), h.CreateRefund)
}

// registerInventoryRoutesWithAuth は在庫ルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerInventoryRoutesWithAuth(rg *gin.RouterGroup) {
	inventory := rg.Group("/inventory")
	inventory.GET("", h.RequirePermission(string(model.ResourceInventory), "view"), h.ListInventory)
	inventory.GET("/:id", h.RequirePermission(string(model.ResourceInventory), "view"), h.GetInventory)
	inventory.POST("", h.RequirePermission(string(model.ResourceInventory), "create"), h.CreateInventory)
	inventory.PATCH("/:id", h.RequirePermission(string(model.ResourceInventory), "edit"), h.UpdateInventory)
	inventory.DELETE("/:id", h.RequirePermission(string(model.ResourceInventory), "delete"), h.DeleteInventory)
}

// registerEstimateRoutesWithAuth は見積書ルートに RBAC 権限チェックを適用する（BUG-125: CRUD個別ガード）
func (h *Handler) registerEstimateRoutesWithAuth(rg *gin.RouterGroup) {
	estimates := rg.Group("/estimates")
	estimates.GET("", h.RequirePermission(string(model.ResourceEstimates), "view"), h.ListEstimates)
	estimates.GET("/:id", h.RequirePermission(string(model.ResourceEstimates), "view"), h.GetEstimate)
	estimates.POST("", h.RequirePermission(string(model.ResourceEstimates), "create"), h.CreateEstimate)
	estimates.PATCH("/:id", h.RequirePermission(string(model.ResourceEstimates), "edit"), h.UpdateEstimate)
	estimates.DELETE("/:id", h.RequirePermission(string(model.ResourceEstimates), "delete"), h.DeleteEstimate)
}
