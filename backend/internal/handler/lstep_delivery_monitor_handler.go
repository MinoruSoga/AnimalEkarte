package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// deliveryTriggerSummaryResponse は GET /lstep/delivery-monitor/summary のJSONレスポンス。
type deliveryTriggerSummaryResponse struct {
	Scheduled               int64            `json:"scheduled"`
	Fired                   int64            `json:"fired"`
	Excluded                int64            `json:"excluded"`
	Failed                  int64            `json:"failed"`
	SuppressedByPriority    int64            `json:"suppressed_by_priority"`
	ExcludedReasonBreakdown map[string]int64 `json:"excluded_reason_breakdown"`
}

// deliveryTriggerLogItemResponse は配信トリガーログ1件のJSONレスポンス。
type deliveryTriggerLogItemResponse struct {
	ID             string  `json:"id"`
	OwnerID        string  `json:"owner_id"`
	OwnerName      string  `json:"owner_name"`
	TriggerType    string  `json:"trigger_type"`
	ScheduledAt    string  `json:"scheduled_at"`
	Status         string  `json:"status"`
	FiredAt        *string `json:"fired_at,omitempty"`
	ExcludedReason *string `json:"excluded_reason,omitempty"`
}

// deliveryTriggerLogsPageResponse は GET /lstep/delivery-monitor/logs のJSONレスポンス。
type deliveryTriggerLogsPageResponse struct {
	Items   []deliveryTriggerLogItemResponse `json:"items"`
	Total   int64                            `json:"total"`
	Page    int                              `json:"page"`
	PerPage int                              `json:"per_page"`
}

func toDeliveryTriggerSummaryResponse(s service.DeliveryTriggerSummary) deliveryTriggerSummaryResponse {
	rb := s.ExcludedReasonBreakdown
	if rb == nil {
		rb = map[string]int64{}
	}
	return deliveryTriggerSummaryResponse{
		Scheduled:               s.Scheduled,
		Fired:                   s.Fired,
		Excluded:                s.Excluded,
		Failed:                  s.Failed,
		SuppressedByPriority:    s.SuppressedByPriority,
		ExcludedReasonBreakdown: rb,
	}
}

// caller must pass non-nil item; panics on nil.
func toDeliveryTriggerLogItemResponse(item *service.DeliveryTriggerLogItem) deliveryTriggerLogItemResponse {
	var firedAt *string
	if item.FiredAt != nil {
		s := item.FiredAt.Format(time.RFC3339)
		firedAt = &s
	}
	return deliveryTriggerLogItemResponse{
		ID:             strconv.FormatUint(item.ID, 10),
		OwnerID:        strconv.FormatUint(item.OwnerID, 10),
		OwnerName:      item.OwnerName,
		TriggerType:    item.TriggerType,
		ScheduledAt:    item.ScheduledAt.Format(time.RFC3339),
		Status:         item.Status,
		FiredAt:        firedAt,
		ExcludedReason: item.ExcludedReason,
	}
}

func toDeliveryTriggerLogsPageResponse(page service.DeliveryTriggerLogsPage) deliveryTriggerLogsPageResponse {
	items := make([]deliveryTriggerLogItemResponse, len(page.Items))
	for i := range page.Items {
		items[i] = toDeliveryTriggerLogItemResponse(&page.Items[i])
	}
	return deliveryTriggerLogsPageResponse{Items: items, Total: page.Total, Page: page.Page, PerPage: page.PerPage}
}

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
