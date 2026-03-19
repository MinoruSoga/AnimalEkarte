package handler

import "time"

// createTrimmingRequest はトリミング作成のバインド struct
type createTrimmingRequest struct {
	Date           *time.Time `json:"date"`
	PetID          *uint64    `json:"pet_id"`
	StaffID        *uint64    `json:"staff_id"`
	CourseID       *uint64    `json:"course_id"`
	Status         string     `json:"status"`
	StyleRequest   string     `json:"style_request"`
	BW             *float64   `json:"bw"`
	BWUnit         string     `json:"bw_unit"`
	BT             *float64   `json:"bt"`
	UsedShampoo    string     `json:"used_shampoo"`
	UsedRibbon     string     `json:"used_ribbon"`
	Remarks        string     `json:"remarks"`
	StyleImage     string     `json:"style_image"`
	CompletedImage string     `json:"completed_image"`
	OptionIDs      []uint64   `json:"option_ids"`
}

// updateTrimmingRequest はトリミング更新のバインド struct
type updateTrimmingRequest struct {
	Date           *time.Time `json:"date"`
	PetID          *uint64    `json:"pet_id"`
	StaffID        *uint64    `json:"staff_id"`
	CourseID       *uint64    `json:"course_id"`
	Status         *string    `json:"status"`
	StyleRequest   *string    `json:"style_request"`
	BW             **float64  `json:"bw"`
	BWUnit         *string    `json:"bw_unit"`
	BT             **float64  `json:"bt"`
	UsedShampoo    *string    `json:"used_shampoo"`
	UsedRibbon     *string    `json:"used_ribbon"`
	Remarks        *string    `json:"remarks"`
	StyleImage     *string    `json:"style_image"`
	CompletedImage *string    `json:"completed_image"`
	// nil = 変更なし、non-nil（空スライス含む）= 全置換
	OptionIDs *[]uint64 `json:"option_ids"`
}
