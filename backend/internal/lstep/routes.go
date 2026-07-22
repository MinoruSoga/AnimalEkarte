package lstep

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// PermissionMiddleware builds the gin.HandlerFunc that gates a route on (resource, action)
// （他 domain と同型・composition root が具象を注入する）。
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// PermissionRequirement identifies one resource/action pair for an OR authorization gate.
type PermissionRequirement struct {
	Resource string
	Action   string
}

// PermissionAnyMiddleware builds a Gin gate that allows any listed requirement.
type PermissionAnyMiddleware func(...PermissionRequirement) gin.HandlerFunc

// Handler composes this slice's per-entity handlers and registers their routes under a
// single, package-unique RegisterRoutes entry point（openapi_route_drift_test.go 規約）。
type Handler struct {
	lstepSettings        *LstepSettingsHandler
	lineSend             *LineSendHandler
	lineLink             *LineLinkHandler
	lineCustomer         *LineCustomerHandler
	tag                  LstepTagService
	tagCodeMapping       LstepTagCodeMappingService
	tagConfig            LstepTagConfigService
	tagSummary           LstepTagSummaryService
	checkupSync          CheckupSyncService
	lifecycle            LstepLifecycleService
	deliveryMonitor      LstepDeliveryMonitorService
	triggerPriority      LstepTriggerPriorityService
	aggregation          AggregationService
	csvImport            LstepCsvImportService
	analytics            LstepAnalyticsService
	sharedFile           SharedFileService
	ownerLineLinker      OwnerLineLinker
	requirePermission    PermissionMiddleware
	requireAnyPermission PermissionAnyMiddleware
}

type OwnerLineLinker interface {
	LinkLineUserID(ctx context.Context, clinicID, ownerID uint64, lineUserID *string, actorID *uint64) error
}

// NewHandler は lstep domain の routing composition を構築する。
func NewHandler(
	lstepSettings *LstepSettingsHandler,
	lineSend *LineSendHandler,
	lineLink *LineLinkHandler,
	lineCustomer *LineCustomerHandler,
	tag LstepTagService,
	tagCodeMapping LstepTagCodeMappingService,
	tagConfig LstepTagConfigService,
	tagSummary LstepTagSummaryService,
	checkupSync CheckupSyncService,
	lifecycle LstepLifecycleService,
	deliveryMonitor LstepDeliveryMonitorService,
	triggerPriority LstepTriggerPriorityService,
	aggregation AggregationService,
	csvImport LstepCsvImportService,
	analytics LstepAnalyticsService,
	sharedFile SharedFileService,
	ownerLineLinker OwnerLineLinker,
	requirePermission PermissionMiddleware,
	requireAnyPermission PermissionAnyMiddleware,
) *Handler {
	return &Handler{
		lstepSettings:        lstepSettings,
		lineSend:             lineSend,
		lineLink:             lineLink,
		lineCustomer:         lineCustomer,
		tag:                  tag,
		tagCodeMapping:       tagCodeMapping,
		tagConfig:            tagConfig,
		tagSummary:           tagSummary,
		checkupSync:          checkupSync,
		lifecycle:            lifecycle,
		deliveryMonitor:      deliveryMonitor,
		triggerPriority:      triggerPriority,
		aggregation:          aggregation,
		csvImport:            csvImport,
		analytics:            analytics,
		sharedFile:           sharedFile,
		ownerLineLinker:      ownerLineLinker,
		requirePermission:    requirePermission,
		requireAnyPermission: requireAnyPermission,
	}
}

// RegisterRoutes は lstep domain の全 route を登録する（旧登録箇所からの RBAC 逐語転記）。
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	ls := rg.Group("/lstep-settings")
	ls.GET("", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.lstepSettings.GetLstepSettings)
	ls.PATCH("", h.requirePermission(string(model.ResourceHospitalSettings), "edit"), h.lstepSettings.UpdateLstepSettings)
	ls.DELETE("", h.requirePermission(string(model.ResourceHospitalSettings), "delete"), h.lstepSettings.DeleteLstepSettings)
	ls.POST("/test-connection", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.lstepSettings.TestLstepConnection)

	// ISSUE-003: FE が /clinics/:clinic_id/lstep-settings で呼ぶエイリアス
	alias := rg.Group("/clinics/:clinic_id/lstep-settings")
	alias.GET("", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.lstepSettings.GetLstepSettings)
	alias.PATCH("", h.requirePermission(string(model.ResourceHospitalSettings), "edit"), h.lstepSettings.UpdateLstepSettings)
	alias.DELETE("", h.requirePermission(string(model.ResourceHospitalSettings), "delete"), h.lstepSettings.DeleteLstepSettings)
	alias.POST("/test-connection", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.lstepSettings.TestLstepConnection)

	// LINE 個別送信（旧 line_send_handler.go RegisterLineSendRoutes 逐語・/owners group 共存）
	owners := rg.Group("/owners")
	owners.POST("/:id/line/send", h.requirePermission(string(model.ResourceOwners), "edit"), h.lineSend.SendLineMessage)
	owners.GET("/:id/line/send-logs", h.requirePermission(string(model.ResourceOwners), "view"), h.lineSend.GetLineSendLogs)
	// ISSUE-002: FE統一エンドポイントのエイリアス
	owners.POST("/:id/lstep/send", h.requirePermission(string(model.ResourceOwners), "edit"), h.lineSend.SendLineMessage)
	owners.GET("/:id/lstep/send-history", h.requirePermission(string(model.ResourceOwners), "view"), h.lineSend.GetLineSendLogs)

	// BE-021: LINE User ID 自動取得・飼い主紐付けトークン発行（co側に無い理由は未文書化・現状維持）
	owners.POST("/:id/line/link-token", h.requirePermission(string(model.ResourceOwners), "edit"), h.lineLink.GenerateLineLinkToken)

	// ISSUE-001/ISSUE-002: /clinics/:clinic_id/owners エイリアス（:clinic_id は無視され JWT の clinic_id を使う）
	// co側は送信2ルートのみ（/lstep/send・/lstep/send-history エイリアスが無い理由は未文書化・現状維持）
	co := rg.Group("/clinics/:clinic_id/owners")
	co.POST("/:id/line/send", h.requirePermission(string(model.ResourceOwners), "edit"), h.lineSend.SendLineMessage)
	co.GET("/:id/line/send-logs", h.requirePermission(string(model.ResourceOwners), "view"), h.lineSend.GetLineSendLogs)

	// BE-019: manual owner tag CRUD (canonical + clinic alias).
	owners.GET("/:id/lstep/tags", h.requirePermission(string(model.ResourceOwners), "view"), h.GetOwnerLstepTags)
	owners.POST("/:id/lstep/tags", h.requirePermission(string(model.ResourceOwners), "edit"), h.AddOwnerLstepTag)
	owners.DELETE("/:id/lstep/tags/:tag_name", h.requirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwnerLstepTag)
	co.GET("/:id/lstep/tags", h.requirePermission(string(model.ResourceOwners), "view"), h.GetOwnerLstepTags)
	co.POST("/:id/lstep/tags", h.requirePermission(string(model.ResourceOwners), "edit"), h.AddOwnerLstepTag)
	co.DELETE("/:id/lstep/tags/:tag_name", h.requirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwnerLstepTag)

	// BE-020: tag analytics (canonical + clinic alias).
	lstepRoutes := rg.Group("/lstep")
	lstepRoutes.GET("/tag-summary", h.requirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepTagSummary)
	lstepRoutes.GET("/owners", h.requirePermission(string(model.ResourceLstepAnalytics), "view"), h.SearchLstepOwnersByTag)
	clinicLstep := rg.Group("/clinics/:clinic_id/lstep")
	clinicLstep.GET("/tag-summary", h.requirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepTagSummary)
	clinicLstep.GET("/owners", h.requirePermission(string(model.ResourceLstepAnalytics), "view"), h.SearchLstepOwnersByTag)

	// FEAT-379: tag-code mapping settings (canonical + clinic alias).
	rg.GET("/lstep-tag-code-mappings", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.ListTagCodeMappings)
	rg.PUT("/lstep-tag-code-mappings/:tag_name", h.requirePermission(string(model.ResourceHospitalSettings), "edit"), h.ReplaceTagCodeMappingsForTag)
	codeMappingAlias := rg.Group("/clinics/:clinic_id/lstep-tag-code-mappings")
	codeMappingAlias.GET("", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.ListTagCodeMappings)
	codeMappingAlias.PUT("/:tag_name", h.requirePermission(string(model.ResourceHospitalSettings), "edit"), h.ReplaceTagCodeMappingsForTag)

	// Dynamic tag configuration (nine hospital-settings routes).
	prefixes := rg.Group("/lstep-tag-config/auto-managed-prefixes")
	prefixes.GET("", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.ListAutoManagedPrefixes)
	prefixes.POST("", h.requirePermission(string(model.ResourceHospitalSettings), "create"), h.CreateAutoManagedPrefix)
	prefixes.DELETE("/:id", h.requirePermission(string(model.ResourceHospitalSettings), "delete"), h.DeleteAutoManagedPrefix)
	conditions := rg.Group("/lstep-tag-config/condition-tag-mappings")
	conditions.GET("", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.ListConditionTagMappings)
	conditions.POST("", h.requirePermission(string(model.ResourceHospitalSettings), "create"), h.CreateConditionTagMapping)
	conditions.DELETE("/:id", h.requirePermission(string(model.ResourceHospitalSettings), "delete"), h.DeleteConditionTagMapping)
	purposes := rg.Group("/lstep-tag-config/send-purpose-tag-prefixes")
	purposes.GET("", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.ListSendPurposeTagPrefixes)
	purposes.POST("", h.requirePermission(string(model.ResourceHospitalSettings), "create"), h.CreateSendPurposeTagPrefix)
	purposes.DELETE("/:id", h.requirePermission(string(model.ResourceHospitalSettings), "delete"), h.DeleteSendPurposeTagPrefix)

	// LSTEP-BE-004: 健診対象者抽出・一括タグ連携。
	clinics := rg.Group("/clinics/:clinic_id")
	clinics.GET("/lstep/checkup-sync/preview", h.requirePermission(string(model.ResourceOwners), "view"), h.GetCheckupSyncPreview)
	clinics.POST("/lstep/checkup-sync", h.requirePermission(string(model.ResourceOwners), "edit"), h.CreateCheckupSync)

	// BE-017 / ISSUE-001: lifecycle routes (canonical + clinic alias).
	owners.DELETE("/:id/line", h.requirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwnerLine)
	owners.POST("/:id/lstep-opt-out", h.requirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerLstepOptOut)
	owners.PATCH("/:id/lstep/opt-out", h.requirePermission(string(model.ResourceOwners), "edit"), h.PatchOwnerLstepOptOut)
	co.POST("/:id/lstep-opt-out", h.requirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerLstepOptOut)
	pets := rg.Group("/pets")
	pets.PATCH("/:id/death", h.requirePermission(string(model.ResourceOwners), "edit"), h.UpdatePetDeath)
	pets.DELETE("/:id/death", h.requirePermission(string(model.ResourceOwners), "edit"), h.DeletePetDeath)
	clinicPets := rg.Group("/clinics/:clinic_id/pets")
	clinicPets.PATCH("/:id/death", h.requirePermission(string(model.ResourceOwners), "edit"), h.UpdatePetDeath)
	clinicPets.DELETE("/:id/death", h.requirePermission(string(model.ResourceOwners), "edit"), h.DeletePetDeath)

	// FEAT-384 / Q23: delivery monitor and trigger priority settings.
	h.RegisterLstepDeliveryMonitorRoutes(rg)
	h.RegisterLstepTriggerPriorityRoutes(rg)

	// BE-010 / FEAT-385: aggregation, CSV import, and analytics.
	h.RegisterAggregationRoutes(rg)
	h.RegisterLstepCsvImportRoutes(rg)
	h.RegisterLstepAnalyticsRoutes(rg)
	co.GET("/:id/lstep/friend-attributes", h.requirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepOwnerFriendAttributes)

	// 顧客管理（旧 reservation_line_routes.go 逐語）
	customers := clinics.Group("/line-customers")
	customers.GET("", h.requirePermission(string(model.ResourceOwners), "view"), h.lineCustomer.ListLineCustomers)
	customers.PATCH("/:customerId/link-owner", h.requirePermission(string(model.ResourceOwners), "edit"), h.lineCustomer.LinkOwnerToLineCustomer)

	// SharedFile routes are canonical-only and derive clinic/staff identity from JWT context.
	h.RegisterSharedFileRoutes(rg)
}

// RegisterWebhookRoutes は LINE Webhook（JWT 認証なし・HMAC-SHA256 署名検証）を engine 直下へ登録する
// （旧 handler.go の r.POST("/api/line/webhook", h.ReceiveLineWebhook) 逐語）。
func (h *Handler) RegisterWebhookRoutes(r *gin.Engine) {
	r.POST("/api/line/webhook", h.lineLink.ReceiveLineWebhook)
}
