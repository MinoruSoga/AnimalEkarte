package handler

import "time"

// createVitalRequest はバイタル作成のバインド struct
type createVitalRequest struct {
	RecordedAt      time.Time `json:"recorded_at"       binding:"required"`
	StaffID         *uint64   `json:"staff_id"`
	Temperature     *float64  `json:"temperature"`
	HeartRate       *int      `json:"heart_rate"`
	RespirationRate *int      `json:"respiration_rate"`
	Weight          *float64  `json:"weight"`
	Notes           string    `json:"notes"`
}

// updateVitalRequest はバイタル更新のバインド struct（全フィールドがオプション）
type updateVitalRequest struct {
	RecordedAt      *time.Time `json:"recorded_at"`
	StaffID         *uint64    `json:"staff_id"`
	Temperature     *float64   `json:"temperature"`
	HeartRate       *int       `json:"heart_rate"`
	RespirationRate *int       `json:"respiration_rate"`
	Weight          *float64   `json:"weight"`
	Notes           *string    `json:"notes"`
}
