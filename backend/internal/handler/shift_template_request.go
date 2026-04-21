package handler

// shiftTemplateBreakRequest は休憩時間テンプレートのリクエストDTO
type shiftTemplateBreakRequest struct {
	BreakStart string `json:"break_start" binding:"required"`
	BreakEnd   string `json:"break_end"   binding:"required"`
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
