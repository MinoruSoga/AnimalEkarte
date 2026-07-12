package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetLstepDeliveryTriggerSummary godoc
// GET /api/clinics/:clinic_id/lstep/delivery-monitor/summary — ステータス別集計を返す（FEAT-384）。
func (h *Handler) GetLstepDeliveryTriggerSummary(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	query, err := newLstepDeliveryMonitorSummaryQuery(clinicID, c.Request.URL.Query(), time.Now())
	if err != nil {
		RespondError(c, err)
		return
	}
	result, err := h.svc.LstepDeliveryMonitor.GetSummary(c.Request.Context(), query.toSummaryServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDeliveryTriggerSummaryResponse(result))
}

// GetLstepDeliveryTriggerLogs godoc
// GET /api/clinics/:clinic_id/lstep/delivery-monitor/logs — ページングログ一覧を返す（FEAT-384）。
func (h *Handler) GetLstepDeliveryTriggerLogs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	query, err := newLstepDeliveryMonitorLogsQuery(clinicID, c.Request.URL.Query(), time.Now())
	if err != nil {
		RespondError(c, err)
		return
	}
	result, err := h.svc.LstepDeliveryMonitor.GetLogs(c.Request.Context(), query.toLogsServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDeliveryTriggerLogsPageResponse(result))
}

// RegisterLstepDeliveryMonitorRoutes は FEAT-384 のルートを登録する。
func (h *Handler) RegisterLstepDeliveryMonitorRoutes(rg *gin.RouterGroup) {
	lstep := rg.Group("/lstep")
	lstep.GET("/delivery-monitor/summary", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepDeliveryTriggerSummary)
	lstep.GET("/delivery-monitor/logs", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepDeliveryTriggerLogs)

	// FE が /clinics/:clinic_id/lstep/... で呼ぶエイリアス
	clinicLstep := rg.Group("/clinics/:clinic_id/lstep")
	clinicLstep.GET("/delivery-monitor/summary", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepDeliveryTriggerSummary)
	clinicLstep.GET("/delivery-monitor/logs", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepDeliveryTriggerLogs)
}
