package handler

// shiftBreakRequest は休憩時間リクエスト
type shiftBreakRequest struct {
	BreakStart string `json:"break_start" binding:"required"`
	BreakEnd   string `json:"break_end"   binding:"required"`
}

// createShiftRequest はシフト登録リクエスト
type createShiftRequest struct {
	StaffID   uint64              `json:"staff_id"   binding:"required"`
	Date      string              `json:"date"       binding:"required"` // YYYY-MM-DD
	ShiftType string              `json:"shift_type" binding:"required,oneof=full morning afternoon off paid_leave"`
	StartTime string              `json:"start_time"`
	EndTime   string              `json:"end_time"`
	Note      string              `json:"note"`
	Breaks    []shiftBreakRequest `json:"breaks"`
}

// updateShiftRequest はシフト更新リクエスト（PATCH）
type updateShiftRequest struct {
	ShiftType *string              `json:"shift_type" binding:"omitempty,oneof=full morning afternoon off paid_leave"`
	StartTime *string              `json:"start_time"`
	EndTime   *string              `json:"end_time"`
	Note      *string              `json:"note"`
	Breaks    *[]shiftBreakRequest `json:"breaks"`
}
