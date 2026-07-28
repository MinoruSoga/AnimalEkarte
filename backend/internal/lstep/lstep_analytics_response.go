package lstep

import (
	"encoding/json"
	"strconv"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
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

func toLstepMonthlyDeliveryStatsResponse(s *MonthlyDeliveryStats) lstepMonthlyDeliveryStatsResponse {
	rows := make([]lstepDeliveryStatsRowResponse, len(s.Rows))
	for i, r := range s.Rows {
		rows[i] = lstepDeliveryStatsRowResponse(r)
	}
	return lstepMonthlyDeliveryStatsResponse{YearMonth: s.YearMonth, Rows: rows}
}

func toLstepVisitConversionResponse(s *VisitConversionSummary) lstepVisitConversionResponse {
	rows := make([]lstepVisitConversionRowResponse, len(s.Rows))
	for i, r := range s.Rows {
		rows[i] = lstepVisitConversionRowResponse(r)
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
		SnapshotTakenAt: httpapi.LocalTimeRFC3339(m.SnapshotTakenAt),
		CreatedAt:       httpapi.LocalTimeRFC3339(m.CreatedAt),
		UpdatedAt:       httpapi.LocalTimeRFC3339(m.UpdatedAt),
	}
	if m.RegisteredAt != nil {
		s := httpapi.LocalTimeRFC3339(*m.RegisteredAt)
		r.RegisteredAt = &s
	}
	if m.LastMessageAt != nil {
		s := httpapi.LocalTimeRFC3339(*m.LastMessageAt)
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
