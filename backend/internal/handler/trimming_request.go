package handler

import "time"

// createTrimmingRequest はトリミング予約作成のバインド struct（BE-119: appointments ベース）
type createTrimmingRequest struct {
	ReservationTypeID uint64     `json:"reservation_type_id" binding:"required"`
	StartTime         *time.Time `json:"start_time"          binding:"required"`
	EndTime           *time.Time `json:"end_time"            binding:"required"`
	PetID             *uint64    `json:"pet_id"               binding:"required"`
	StaffID           *uint64    `json:"staff_id"`
	Status            string     `json:"status"              binding:"omitempty,oneof=confirmed pending cancelled checked_in in_consultation accounting completed"`
	// トリミング詳細
	CourseID       *uint64  `json:"course_id"`
	StyleRequest   string   `json:"style_request"`
	BW             *float64 `json:"bw"`
	BWUnit         string   `json:"bw_unit"`
	BT             *float64 `json:"bt"`
	UsedShampoo    string   `json:"used_shampoo"`
	UsedRibbon     string   `json:"used_ribbon"`
	Remarks        string   `json:"remarks"`
	StyleImage     string   `json:"style_image"`
	CompletedImage string   `json:"completed_image"`
	OptionIDs      []uint64 `json:"option_ids"`
}

// updateTrimmingRequest はトリミング予約更新のバインド struct（BE-119）
type updateTrimmingRequest struct {
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	PetID     *uint64    `json:"pet_id"`
	StaffID   *uint64    `json:"staff_id"`
	Status    *string    `json:"status"      binding:"omitempty,oneof=confirmed pending cancelled checked_in in_consultation accounting completed"`
	// トリミング詳細
	CourseID       *uint64  `json:"course_id"`
	StyleRequest   *string  `json:"style_request"`
	BW             *float64 `json:"bw"`
	BWUnit         *string  `json:"bw_unit"`
	BT             *float64 `json:"bt"`
	UsedShampoo    *string  `json:"used_shampoo"`
	UsedRibbon     *string  `json:"used_ribbon"`
	Remarks        *string  `json:"remarks"`
	StyleImage     *string  `json:"style_image"`
	CompletedImage *string  `json:"completed_image"`
	// nil = 変更なし、non-nil（空スライス含む）= 全置換
	OptionIDs *[]uint64 `json:"option_ids"`
}
