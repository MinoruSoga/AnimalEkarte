package staff

import "github.com/animal-ekarte/backend/internal/model"

// shiftTemplateBreakResponse は休憩時間テンプレートのレスポンスDTO
type shiftTemplateBreakResponse struct {
	ID         uint64 `json:"id"`
	BreakStart string `json:"break_start"`
	BreakEnd   string `json:"break_end"`
}

// shiftTemplateResponse はシフトテンプレートのレスポンスDTO
type shiftTemplateResponse struct {
	ID        uint64                       `json:"id"`
	ClinicID  uint64                       `json:"clinic_id"`
	Name      string                       `json:"name"`
	ShiftType string                       `json:"shift_type"`
	StartTime string                       `json:"start_time"`
	EndTime   string                       `json:"end_time"`
	Notes     string                       `json:"notes"`
	SortOrder int                          `json:"sort_order"`
	IsActive  bool                         `json:"is_active"`
	Breaks    []shiftTemplateBreakResponse `json:"breaks"`
}

func toShiftTemplateResponse(tpl *model.ShiftTemplate) shiftTemplateResponse {
	breaks := make([]shiftTemplateBreakResponse, 0, len(tpl.Breaks))
	for _, b := range tpl.Breaks {
		breaks = append(breaks, shiftTemplateBreakResponse{
			ID:         b.ID,
			BreakStart: b.BreakStart,
			BreakEnd:   b.BreakEnd,
		})
	}
	startTime := ""
	if tpl.StartTime != nil {
		startTime = *tpl.StartTime
	}
	endTime := ""
	if tpl.EndTime != nil {
		endTime = *tpl.EndTime
	}
	return shiftTemplateResponse{
		ID:        tpl.ID,
		ClinicID:  tpl.ClinicID,
		Name:      tpl.Name,
		ShiftType: string(tpl.ShiftType),
		StartTime: startTime,
		EndTime:   endTime,
		Notes:     tpl.Notes,
		SortOrder: tpl.SortOrder,
		IsActive:  tpl.IsActive,
		Breaks:    breaks,
	}
}
