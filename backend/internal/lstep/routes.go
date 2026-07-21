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
	requirePermission PermissionMiddleware
}

// NewHandler は lstep domain の routing composition を構築する。
func NewHandler(lstepSettings *LstepSettingsHandler, requirePermission PermissionMiddleware) *Handler {
	return &Handler{lstepSettings: lstepSettings, requirePermission: requirePermission}
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
}
