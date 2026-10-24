package reservation

import (
	"time"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type shiftEntryBreakResponse struct {
	ID         uint64 `json:"id"`
	BreakStart string `json:"break_start"`
	BreakEnd   string `json:"break_end"`
}

type scheduleEntryResponse struct {
	ID        uint64                    `json:"id"`
	ClinicID  uint64                    `json:"clinic_id"`
	StaffID   uint64                    `json:"staff_id"`
	Date      string                    `json:"date"`
	ShiftType string                    `json:"shift_type"`
	WorkStart *string                   `json:"work_start,omitempty"` // リクエストと統一（work_start/work_end）
	WorkEnd   *string                   `json:"work_end,omitempty"`
	Note      string                    `json:"note"`
	Breaks    []shiftEntryBreakResponse `json:"breaks"`
	CreatedAt string                    `json:"created_at"`
	UpdatedAt string                    `json:"updated_at"`
}

func toScheduleEntryResponse(e *ScheduleEntry) scheduleEntryResponse {
	breaks := make([]shiftEntryBreakResponse, 0, len(e.Breaks))
	for _, b := range e.Breaks {
		breaks = append(breaks, toShiftEntryBreakResponse(b))
	}
	return scheduleEntryResponse{
		ID:        e.Entry.ID,
		ClinicID:  e.Entry.ClinicID,
		StaffID:   e.Entry.StaffID,
		Date:      e.Entry.Date.In(time.Local).Format(time.DateOnly),
		ShiftType: string(e.Entry.ShiftType),
		WorkStart: formatScheduleTimestamp(e.Entry.StartsAt),
		WorkEnd:   formatScheduleTimestamp(e.Entry.EndsAt),
		Note:      e.Entry.Notes,
		Breaks:    breaks,
		CreatedAt: httpapi.LocalTimeRFC3339(e.Entry.CreatedAt),
		UpdatedAt: httpapi.LocalTimeRFC3339(e.Entry.UpdatedAt),
	}
}

// formatScheduleTimestamp は starts_at / ends_at を JST の RFC 3339 date-time で返す。nil は nil。
func formatScheduleTimestamp(ts *time.Time) *string {
	if ts == nil {
		return nil
	}
	formatted := ts.In(config.JST).Format(time.RFC3339Nano)
	return &formatted
}

// scheduleTimestampOrEmpty は休憩の starts_at / ends_at を JST の RFC 3339 date-time で返す。nil は空文字（現行どおり）。
func scheduleTimestampOrEmpty(ts *time.Time) string {
	if formatted := formatScheduleTimestamp(ts); formatted != nil {
		return *formatted
	}
	return ""
}

func toShiftEntryBreakResponse(b model.ShiftEntryBreak) shiftEntryBreakResponse {
	return shiftEntryBreakResponse{
		ID:         b.ID,
		BreakStart: scheduleTimestampOrEmpty(b.StartsAt),
		BreakEnd:   scheduleTimestampOrEmpty(b.EndsAt),
	}
}
