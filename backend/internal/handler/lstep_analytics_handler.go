package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

type lstepDeliveryStatsRowResponse struct {
	TriggerType string `json:"trigger_type"`
	Status      string `json:"status"`
	Count       int64  `json:"count"`
}

type lstepMonthlyDeliveryStatsResponse struct {
	YearMonth string                          `json:"year_month"`
	Rows      []lstepDeliveryStatsRowResponse `json:"rows"`
}

type lstepVisitConversionRowResponse struct {
	TriggerType    string  `json:"trigger_type"`
	DeliveredCount int64   `json:"delivered_count"`
	VisitedCount   int64   `json:"visited_count"`
	VisitRate      float64 `json:"visit_rate"`
}

type lstepVisitConversionResponse struct {
	YearMonth      string                            `json:"year_month"`
	Days           int                               `json:"days"`
	DeliveredCount int64                             `json:"delivered_count"`
	VisitedCount   int64                             `json:"visited_count"`
	VisitRate      float64                           `json:"visit_rate"`
	Rows           []lstepVisitConversionRowResponse `json:"rows"`
}

func toLstepMonthlyDeliveryStatsResponse(s *service.MonthlyDeliveryStats) lstepMonthlyDeliveryStatsResponse {
	rows := make([]lstepDeliveryStatsRowResponse, len(s.Rows))
	for i, r := range s.Rows {
		rows[i] = lstepDeliveryStatsRowResponse{
			TriggerType: r.TriggerType,
			Status:      r.Status,
			Count:       r.Count,
		}
	}
	return lstepMonthlyDeliveryStatsResponse{YearMonth: s.YearMonth, Rows: rows}
}

func toLstepVisitConversionResponse(s *service.VisitConversionSummary) lstepVisitConversionResponse {
	rows := make([]lstepVisitConversionRowResponse, len(s.Rows))
	for i, r := range s.Rows {
		rows[i] = lstepVisitConversionRowResponse{
			TriggerType:    r.TriggerType,
			DeliveredCount: r.DeliveredCount,
			VisitedCount:   r.VisitedCount,
			VisitRate:      r.VisitRate,
		}
	}
	return lstepVisitConversionResponse{
		YearMonth:      s.YearMonth,
		Days:           s.Days,
		DeliveredCount: s.DeliveredCount,
		VisitedCount:   s.VisitedCount,
		VisitRate:      s.VisitRate,
		Rows:           rows,
	}
}

type lstepFriendAttributeResponse struct {
	ID              string          `json:"id"`
	LineUserID      string          `json:"line_user_id"`
	DisplayName     *string         `json:"display_name,omitempty"`
	RegisteredAt    *string         `json:"registered_at,omitempty"`
	Tags            json.RawMessage `json:"tags,omitempty"`
	Scenarios       json.RawMessage `json:"scenarios,omitempty"`
	TrafficSource   *string         `json:"traffic_source,omitempty"`
	BlockStatus     *string         `json:"block_status,omitempty"`
	LastMessageAt   *string         `json:"last_message_at,omitempty"`
	SnapshotTakenAt string          `json:"snapshot_taken_at"`
	CsvImportID     *string         `json:"csv_import_id,omitempty"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

func toLstepFriendAttributeResponse(m *model.LstepFriendAttributeSnapshot) lstepFriendAttributeResponse {
	r := lstepFriendAttributeResponse{
		ID:              strconv.FormatUint(m.ID, 10),
		LineUserID:      m.LineUserID,
		DisplayName:     m.DisplayName,
		TrafficSource:   m.TrafficSource,
		BlockStatus:     m.BlockStatus,
		SnapshotTakenAt: localTimeRFC3339(m.SnapshotTakenAt),
		CreatedAt:       localTimeRFC3339(m.CreatedAt),
		UpdatedAt:       localTimeRFC3339(m.UpdatedAt),
	}
	if m.RegisteredAt != nil {
		s := localTimeRFC3339(*m.RegisteredAt)
		r.RegisteredAt = &s
	}
	if m.LastMessageAt != nil {
		s := localTimeRFC3339(*m.LastMessageAt)
		r.LastMessageAt = &s
	}
	if len(m.Tags) > 0 {
		r.Tags = json.RawMessage(m.Tags)
	}
	if len(m.Scenarios) > 0 {
		r.Scenarios = json.RawMessage(m.Scenarios)
	}
	if m.CsvImportID != nil {
		s := m.CsvImportID.String()
		r.CsvImportID = &s
	}
	return r
}

// GetLstepMonthlyDeliveryStats godoc
// GET /api/v1/clinics/:clinic_id/lstep/analytics/delivery-stats?year_month=YYYY-MM — 月次配信集計（FEAT-385）。
func (h *Handler) GetLstepMonthlyDeliveryStats(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	query, err := newLstepMonthlyDeliveryStatsQuery(c.Request.URL.Query())
	if err != nil {
		RespondError(c, err)
		return
	}
	stats, err := h.svc.LstepAnalytics.GetMonthlyDeliveryStats(c.Request.Context(), clinicID, query.YearMonth)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepMonthlyDeliveryStatsResponse(stats))
}

// GetLstepVisitConversionSummary godoc
// GET /api/v1/clinics/:clinic_id/lstep/analytics/visit-conversion?year_month=YYYY-MM&days=30 — 月次配信後来院率集計。
func (h *Handler) GetLstepVisitConversionSummary(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	query, err := newLstepVisitConversionQuery(c.Request.URL.Query())
	if err != nil {
		RespondError(c, err)
		return
	}
	days, err := query.toDays()
	if err != nil {
		RespondError(c, err)
		return
	}
	stats, err := h.svc.LstepAnalytics.GetVisitConversionSummary(c.Request.Context(), clinicID, query.YearMonth, days)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepVisitConversionResponse(stats))
}

// GetLstepOwnerFriendAttributes godoc
// GET /api/v1/clinics/:clinic_id/owners/:id/lstep/friend-attributes — 飼主の最新 Lステップ友だち属性（FEAT-385）。
func (h *Handler) GetLstepOwnerFriendAttributes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	snapshot, err := h.svc.LstepAnalytics.GetLatestFriendAttributes(c.Request.Context(), clinicID, ownerID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepFriendAttributeResponse(snapshot))
}

// RegisterLstepAnalyticsRoutes は FEAT-385 分析のルートを登録する。
func (h *Handler) RegisterLstepAnalyticsRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/clinics/:clinic_id/lstep/analytics")
	g.GET("/delivery-stats", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepMonthlyDeliveryStats)
	g.GET("/visit-conversion", h.RequirePermission(string(model.ResourceLstepAnalytics), "view"), h.GetLstepVisitConversionSummary)
}
