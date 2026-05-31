package handler

import (
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// shiftTemplateBreakRequest は休憩時間テンプレートのリクエストDTO
type shiftTemplateBreakRequest struct {
	BreakStart string `json:"break_start" binding:"required"`
	BreakEnd   string `json:"break_end"   binding:"required"`
}

func toShiftBreakTemplateInputs(reqs []shiftTemplateBreakRequest) []service.ShiftBreakTemplateInput {
	breaks := make([]service.ShiftBreakTemplateInput, 0, len(reqs))
	for _, req := range reqs {
		breaks = append(breaks, service.ShiftBreakTemplateInput{
			BreakStart: req.BreakStart,
			BreakEnd:   req.BreakEnd,
		})
	}
	return breaks
}

// createShiftTemplateRequest はシフトテンプレート作成リクエスト
type createShiftTemplateRequest struct {
	Name      string                      `json:"name"       binding:"required"`
	ShiftType string                      `json:"shift_type" binding:"required,oneof=full morning afternoon off paid_leave"`
	StartTime string                      `json:"start_time"`
	EndTime   string                      `json:"end_time"`
	Notes     string                      `json:"notes"`
	SortOrder int                         `json:"sort_order"`
	IsActive  *bool                       `json:"is_active"`
	Breaks    []shiftTemplateBreakRequest `json:"breaks"`
}

func (r *createShiftTemplateRequest) toServiceInput() *service.CreateShiftTemplateInput {
	return &service.CreateShiftTemplateInput{
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
	Name      *string                      `json:"name"`
	ShiftType *string                      `json:"shift_type" binding:"omitempty,oneof=full morning afternoon off paid_leave"`
	StartTime *string                      `json:"start_time"`
	EndTime   *string                      `json:"end_time"`
	Notes     *string                      `json:"notes"`
	SortOrder *int                         `json:"sort_order"`
	IsActive  *bool                        `json:"is_active"`
	Breaks    *[]shiftTemplateBreakRequest `json:"breaks"`
}

func (r updateShiftTemplateRequest) toServiceInput() *service.UpdateShiftTemplateInput {
	input := &service.UpdateShiftTemplateInput{
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
