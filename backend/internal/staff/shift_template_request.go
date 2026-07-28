package staff

import (
	"github.com/animal-ekarte/backend/internal/model"
)

// shiftTemplateBreakRequest は休憩時間テンプレートのリクエストDTO
type shiftTemplateBreakRequest struct {
	BreakStart string `json:"break_start" binding:"required,max=8"`
	BreakEnd   string `json:"break_end"   binding:"required,max=8"`
}

func toShiftBreakTemplateInputs(reqs []shiftTemplateBreakRequest) []ShiftBreakTemplateInput {
	breaks := make([]ShiftBreakTemplateInput, 0, len(reqs))
	for _, req := range reqs {
		breaks = append(breaks, ShiftBreakTemplateInput{
			BreakStart: req.BreakStart,
			BreakEnd:   req.BreakEnd,
		})
	}
	return breaks
}

// createShiftTemplateRequest はシフトテンプレート作成リクエスト
type createShiftTemplateRequest struct {
	Name      string                      `json:"name"       binding:"required,max=255"`
	ShiftType string                      `json:"shift_type" binding:"required,max=10,oneof=full morning afternoon off paid_leave"`
	StartTime string                      `json:"start_time" binding:"max=8"`
	EndTime   string                      `json:"end_time"   binding:"max=8"`
	Notes     string                      `json:"notes"      binding:"max=2000"`
	SortOrder int                         `json:"sort_order"`
	IsActive  *bool                       `json:"is_active"`
	Breaks    []shiftTemplateBreakRequest `json:"breaks" binding:"max=50,dive"`
}

func (r *createShiftTemplateRequest) toServiceInput() *CreateShiftTemplateInput {
	return &CreateShiftTemplateInput{
		Name:      r.Name,
		ShiftType: r.ShiftType,
		StartTime: r.StartTime,
		EndTime:   r.EndTime,
		Notes:     r.Notes,
		SortOrder: r.SortOrder,
		IsActive:  r.IsActive,
		Breaks:    toShiftBreakTemplateInputs(r.Breaks),
	}
}

// updateShiftTemplateRequest はシフトテンプレート更新リクエスト（PATCH）
type updateShiftTemplateRequest struct {
	Name      *string                      `json:"name"       binding:"omitempty,max=255"`
	ShiftType *string                      `json:"shift_type" binding:"omitempty,max=10,oneof=full morning afternoon off paid_leave"`
	StartTime *string                      `json:"start_time" binding:"omitempty,max=8"`
	EndTime   *string                      `json:"end_time"   binding:"omitempty,max=8"`
	Notes     *string                      `json:"notes"      binding:"omitempty,max=2000"`
	SortOrder *int                         `json:"sort_order"`
	IsActive  *bool                        `json:"is_active"`
	Breaks    *[]shiftTemplateBreakRequest `json:"breaks" binding:"omitempty,max=50,dive"`
}

func (r updateShiftTemplateRequest) toServiceInput() *UpdateShiftTemplateInput {
	input := &UpdateShiftTemplateInput{
		Name:      r.Name,
		Notes:     r.Notes,
		SortOrder: r.SortOrder,
		IsActive:  r.IsActive,
		StartTime: r.StartTime,
		EndTime:   r.EndTime,
	}
	if r.ShiftType != nil {
		shiftType := model.ShiftType(*r.ShiftType)
		input.ShiftType = &shiftType
	}
	if r.Breaks != nil {
		breaks := toShiftBreakTemplateInputs(*r.Breaks)
		input.Breaks = &breaks
	}
	return input
}
