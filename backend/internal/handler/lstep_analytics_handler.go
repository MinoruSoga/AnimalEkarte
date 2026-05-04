package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
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
		SnapshotTakenAt: m.SnapshotTakenAt.Format(time.RFC3339),
		CreatedAt:       m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       m.UpdatedAt.Format(time.RFC3339),
	}
	if m.RegisteredAt != nil {
		s := m.RegisteredAt.Format(time.RFC3339)
		r.RegisteredAt = &s
	}
	if m.LastMessageAt != nil {
		s := m.LastMessageAt.Format(time.RFC3339)
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
	yearMonth := c.Query("year_month")
	if yearMonth == "" {
		RespondError(c, apperrors.WrapInvalidInput("year_month クエリパラメータは必須です"))
		return
	}
	stats, err := h.svc.LstepAnalytics.GetMonthlyDeliveryStats(c.Request.Context(), clinicID, yearMonth)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepMonthlyDeliveryStatsResponse(stats))
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
}
