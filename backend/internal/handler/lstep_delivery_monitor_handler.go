package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// deliveryTriggerSummaryResponse は GET /lstep/delivery-monitor/summary のJSONレスポンス。
type deliveryTriggerSummaryResponse struct {
	Scheduled               int64            `json:"scheduled"`
	Fired                   int64            `json:"fired"`
	Excluded                int64            `json:"excluded"`
	Failed                  int64            `json:"failed"`
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
		ExcludedReasonBreakdown: rb,
	}
}

func toDeliveryTriggerLogItemResponse(item service.DeliveryTriggerLogItem) deliveryTriggerLogItemResponse {
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
	for i, it := range page.Items {
		items[i] = toDeliveryTriggerLogItemResponse(it)
	}
	return deliveryTriggerLogsPageResponse{Items: items, Total: page.Total, Page: page.Page, PerPage: page.PerPage}
}

// parseDateRange は from/to クエリパラメータを JST 基準でパースする。
// 未指定時は当日の範囲（当日0時 〜 翌日0時）をデフォルトとする。
// to は日付指定を inclusive end として扱い、内部的に翌日0時（exclusive）に変換する。
func parseDateRange(c *gin.Context) (from, to time.Time, ok bool) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Now().In(jst)
	defaultDate := now.Format("2006-01-02")

	fromStr := c.DefaultQuery("from", defaultDate)
	toStr := c.DefaultQuery("to", defaultDate)

	var err error
	from, err = time.ParseInLocation("2006-01-02", fromStr, jst)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("from は YYYY-MM-DD 形式で指定してください"))
		return time.Time{}, time.Time{}, false
	}
	to, err = time.ParseInLocation("2006-01-02", toStr, jst)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("to は YYYY-MM-DD 形式で指定してください"))
		return time.Time{}, time.Time{}, false
	}
	to = to.AddDate(0, 0, 1) // inclusive end → exclusive upper bound
	return from, to, true
}

// GetLstepDeliveryTriggerSummary godoc
// GET /api/clinics/:clinic_id/lstep/delivery-monitor/summary — ステータス別集計を返す（FEAT-384）。
func (h *Handler) GetLstepDeliveryTriggerSummary(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	from, to, ok := parseDateRange(c)
	if !ok {
		return
	}
	triggerType := c.Query("trigger_type")

	result, err := h.svc.LstepDeliveryMonitor.GetSummary(c.Request.Context(), service.GetDeliveryMonitorSummaryInput{
		ClinicID:    clinicID,
		From:        from,
		To:          to,
		TriggerType: triggerType,
	})
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
	from, to, ok := parseDateRange(c)
	if !ok {
		return
	}

	triggerType := c.Query("trigger_type")
	status := c.Query("status")

	page, pageErr := strconv.Atoi(c.DefaultQuery("page", "1"))
	if pageErr != nil || page < 1 {
		RespondError(c, apperrors.WrapInvalidInput("page は1以上の整数で指定してください"))
		return
	}
	perPage, ppErr := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if ppErr != nil || perPage < 1 || perPage > 100 {
		RespondError(c, apperrors.WrapInvalidInput("per_page は1〜100の範囲で指定してください"))
		return
	}

	result, err := h.svc.LstepDeliveryMonitor.GetLogs(c.Request.Context(), service.GetDeliveryMonitorLogsInput{
		ClinicID:    clinicID,
		From:        from,
		To:          to,
		TriggerType: triggerType,
		Status:      status,
		Page:        page,
		PerPage:     perPage,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDeliveryTriggerLogsPageResponse(result))
}

// RegisterLstepDeliveryMonitorRoutes は FEAT-384 のルートを登録する。
func (h *Handler) RegisterLstepDeliveryMonitorRoutes(rg *gin.RouterGroup) {
	lstep := rg.Group("/lstep")
	lstep.GET("/delivery-monitor/summary", h.GetLstepDeliveryTriggerSummary)
	lstep.GET("/delivery-monitor/logs", h.GetLstepDeliveryTriggerLogs)

	// FE が /clinics/:clinic_id/lstep/... で呼ぶエイリアス
	clinicLstep := rg.Group("/clinics/:clinic_id/lstep")
	clinicLstep.GET("/delivery-monitor/summary", h.GetLstepDeliveryTriggerSummary)
	clinicLstep.GET("/delivery-monitor/logs", h.GetLstepDeliveryTriggerLogs)
}
