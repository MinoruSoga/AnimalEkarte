package handler

import "encoding/json"

// liffCreateReservationRequest は予約確定リクエスト。
type liffCreateReservationRequest struct {
	TypeID               uint64          `json:"course_id"        binding:"required"`
	StaffID              uint64          `json:"staff_id"`                            // 0 = 指名なし
	Date                 string          `json:"date"             binding:"required"` // "YYYY-MM-DD"
	StartTime            string          `json:"start_time"       binding:"required"` // "HHMM"
	EndTime              string          `json:"end_time"         binding:"required"` // "HHMM"
	CustomerFields       json.RawMessage `json:"customer_fields"`
	RequestText          string          `json:"request_text"`
	TrimmingCourseID     *uint64         `json:"trimming_course_id"`     // BE-120
	TrimmingOptionIDs    []uint64        `json:"trimming_option_ids"`    // BE-120
	TrimmingStyleRequest string          `json:"trimming_style_request"` // BE-120
}
