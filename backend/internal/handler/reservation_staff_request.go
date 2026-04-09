package handler

type createReservationStaffRequest struct {
	Name               string   `json:"name"               binding:"required"`
	StaffType          string   `json:"staff_type"`
	ReservationVisible bool     `json:"reservation_visible"`
	ReservationComment string   `json:"reservation_comment"`
	SortOrder          int      `json:"sort_order"`
	ExcludedCourseIDs  []uint64 `json:"excluded_course_ids"`
}

type updateReservationStaffRequest struct {
	Name               *string   `json:"name"`
	StaffType          *string   `json:"staff_type"`
	ReservationVisible *bool     `json:"reservation_visible"`
	ReservationComment *string   `json:"reservation_comment"`
	SortOrder          *int      `json:"sort_order"`
	ExcludedCourseIDs  *[]uint64 `json:"excluded_course_ids"`
}

type patchReservationStaffStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type patchReservationStaffSortOrderRequest struct {
	Direction string `json:"direction" binding:"required,oneof=up down"`
}
