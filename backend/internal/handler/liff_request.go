package handler

import "encoding/json"

// liffCreateReservationRequest は予約確定リクエスト。
type liffCreateReservationRequest struct {
	CourseID       uint64          `json:"course_id"        binding:"required"`
	StaffID        uint64          `json:"staff_id"`                            // 0 = 指名なし
	Date           string          `json:"date"             binding:"required"` // "YYYY-MM-DD"
	StartTime      string          `json:"start_time"       binding:"required"` // "HHMM"
	EndTime        string          `json:"end_time"         binding:"required"` // "HHMM"
	CustomerFields json.RawMessage `json:"customer_fields"`
	RequestText    string          `json:"request_text"`
}
