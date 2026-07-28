package staff

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// shiftBreakResponse は休憩時間レスポンス
type shiftBreakResponse struct {
	ID         string `json:"id"`
	BreakStart string `json:"break_start"`
	BreakEnd   string `json:"break_end"`
}

// shiftResponse はシフトエントリのレスポンス
type shiftResponse struct {
	ID        string                `json:"id"`
	ClinicID  string                `json:"clinic_id"`
	StaffID   string                `json:"staff_id"`
	StaffName string                `json:"staff_name,omitempty"`
	Date      string                `json:"date"` // YYYY-MM-DD
	ShiftType model.ShiftType       `json:"shift_type"`
	StartTime string                `json:"start_time"`
	EndTime   string                `json:"end_time"`
	Notes     string                `json:"notes"`
	Breaks    []shiftBreakResponse  `json:"breaks"`
	CreatedAt string                `json:"created_at"`
	UpdatedAt string                `json:"updated_at"`
	Staff     *staffSummaryResponse `json:"staff,omitempty"`
}

func toShiftResponse(s *model.ShiftEntry) shiftResponse {
	var startTime, endTime string
	if s.StartTime != nil {
		startTime = *s.StartTime
	}
	if s.EndTime != nil {
		endTime = *s.EndTime
	}
	breaks := make([]shiftBreakResponse, 0, len(s.Breaks))
	for _, b := range s.Breaks {
		breaks = append(breaks, shiftBreakResponse{
			ID:         strconv.FormatUint(b.ID, 10),
			BreakStart: b.BreakStart,
			BreakEnd:   b.BreakEnd,
		})
	}
	r := shiftResponse{
		ID:        strconv.FormatUint(s.ID, 10),
		ClinicID:  strconv.FormatUint(s.ClinicID, 10),
		StaffID:   strconv.FormatUint(s.StaffID, 10),
		Date:      s.Date.In(time.Local).Format(time.DateOnly),
		ShiftType: s.ShiftType,
		StartTime: startTime,
		EndTime:   endTime,
		Notes:     s.Notes,
		Breaks:    breaks,
		CreatedAt: localTimeRFC3339(s.CreatedAt),
		UpdatedAt: localTimeRFC3339(s.UpdatedAt),
	}
	if s.Staff != nil && s.Staff.ID != 0 {
		r.StaffName = s.Staff.Name
		r.Staff = &staffSummaryResponse{ID: s.Staff.ID, Name: s.Staff.Name}
	}
	return r
}
