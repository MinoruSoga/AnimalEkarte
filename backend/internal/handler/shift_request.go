package handler

// createShiftRequest はシフト登録リクエスト
type createShiftRequest struct {
	StaffID   uint64 `json:"staff_id"   binding:"required"`
	Date      string `json:"date"       binding:"required"` // YYYY-MM-DD
	ShiftType string `json:"shift_type" binding:"required,oneof=full morning afternoon off paid_leave"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Note      string `json:"note"`
}

// updateShiftRequest はシフト更新リクエスト（PATCH）
type updateShiftRequest struct {
	ShiftType *string `json:"shift_type" binding:"omitempty,oneof=full morning afternoon off paid_leave"`
	StartTime *string `json:"start_time"`
	EndTime   *string `json:"end_time"`
	Note      *string `json:"note"`
}
