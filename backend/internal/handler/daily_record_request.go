package handler

// addVitalRecordRequest はバイタル記録追加のバインド struct
type addVitalRecordRequest struct {
	Time            string   `json:"time"             binding:"required"`
	Temperature     *float64 `json:"temperature"`
	HeartRate       *int     `json:"heart_rate"`
	RespirationRate *int     `json:"respiration_rate"`
	Weight          *float64 `json:"weight"`
	Notes           string   `json:"notes"`
	StaffID         *uint64  `json:"staff_id"`
}

// addCareLogRequest はケアログ記録追加のバインド struct
type addCareLogRequest struct {
	Time    string  `json:"time"    binding:"required"`
	Type    string  `json:"type"    binding:"required,oneof=food excretion medicine treatment other"`
	Status  string  `json:"status"  binding:"omitempty,oneof=completed partial skipped"`
	Value   string  `json:"value"`
	StaffID *uint64 `json:"staff_id"`
	Notes   string  `json:"notes"`
}

// addStaffNoteRequest はスタッフメモ記録追加のバインド struct
type addStaffNoteRequest struct {
	Time    string  `json:"time"    binding:"required"`
	Content string  `json:"content" binding:"required"`
	StaffID *uint64 `json:"staff_id"`
}
