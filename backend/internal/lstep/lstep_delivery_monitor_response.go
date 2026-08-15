package lstep

import (
	"strconv"

	"github.com/animal-ekarte/backend/internal/httpapi"
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

func toDeliveryTriggerSummaryResponse(s DeliveryTriggerSummary) deliveryTriggerSummaryResponse {
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
func toDeliveryTriggerLogItemResponse(item *DeliveryTriggerLogItem) deliveryTriggerLogItemResponse {
	var firedAt *string
	if item.FiredAt != nil {
		s := httpapi.LocalTimeRFC3339(*item.FiredAt)
		firedAt = &s
	}
	return deliveryTriggerLogItemResponse{
		ID:             strconv.FormatUint(item.ID, 10),
		OwnerID:        strconv.FormatUint(item.OwnerID, 10),
		OwnerName:      item.OwnerName,
		TriggerType:    item.TriggerType,
		ScheduledAt:    httpapi.LocalTimeRFC3339(item.ScheduledAt),
		Status:         item.Status,
		FiredAt:        firedAt,
		ExcludedReason: item.ExcludedReason,
	}
}

func toDeliveryTriggerLogsPageResponse(page DeliveryTriggerLogsPage) deliveryTriggerLogsPageResponse {
	items := make([]deliveryTriggerLogItemResponse, len(page.Items))
	for i := range page.Items {
		items[i] = toDeliveryTriggerLogItemResponse(&page.Items[i])
	}
	return deliveryTriggerLogsPageResponse{Items: items, Total: page.Total, Page: page.Page, PerPage: page.PerPage}
}
