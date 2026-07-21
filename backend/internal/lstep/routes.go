package lstep

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// PermissionMiddleware builds the gin.HandlerFunc that gates a route on (resource, action)
// （他 domain と同型・composition root が具象を注入する）。
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// Handler composes this slice's per-entity handlers and registers their routes under a
// single, package-unique RegisterRoutes entry point（openapi_route_drift_test.go 規約）。
type Handler struct {
	lstepSettings     *LstepSettingsHandler
	lineSend          *LineSendHandler
	lineLink          *LineLinkHandler
	lineCustomer      *LineCustomerHandler
	requirePermission PermissionMiddleware
}

// NewHandler は lstep domain の routing composition を構築する。
func NewHandler(
	lstepSettings *LstepSettingsHandler,
	lineSend *LineSendHandler,
	lineLink *LineLinkHandler,
	lineCustomer *LineCustomerHandler,
	requirePermission PermissionMiddleware,
) *Handler {
	return &Handler{
		lstepSettings:     lstepSettings,
		lineSend:          lineSend,
		lineLink:          lineLink,
		lineCustomer:      lineCustomer,
		requirePermission: requirePermission,
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

	// 顧客管理（旧 reservation_line_routes.go 逐語）
	clinics := rg.Group("/clinics/:clinic_id")
	customers := clinics.Group("/line-customers")
	customers.GET("", h.requirePermission(string(model.ResourceOwners), "view"), h.lineCustomer.ListLineCustomers)
	customers.PATCH("/:customerId/link-owner", h.requirePermission(string(model.ResourceOwners), "edit"), h.lineCustomer.LinkOwnerToLineCustomer)
}

// RegisterWebhookRoutes は LINE Webhook（JWT 認証なし・HMAC-SHA256 署名検証）を engine 直下へ登録する
// （旧 handler.go の r.POST("/api/line/webhook", h.ReceiveLineWebhook) 逐語）。
func (h *Handler) RegisterWebhookRoutes(r *gin.Engine) {
	r.POST("/api/line/webhook", h.lineLink.ReceiveLineWebhook)
}
