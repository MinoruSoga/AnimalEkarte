package reservation

type createReservationStaffRequest struct {
	Name      string `json:"name"               binding:"required"`
	StaffType string `json:"staff_type" binding:"omitempty,oneof=doctor nurse groomer other"`
	// ReservationVisible is *bool so omitted stays default true; explicit false is preserved.
	ReservationVisible *bool  `json:"reservation_visible"`
	ReservationComment string `json:"reservation_comment"`
	SortOrder          int    `json:"sort_order"`
	// TASK-021 UNIT-021-A: excluded_type_ids removed from request DTO / OpenAPI.
	// Capability seed on Create uses inverse facade with empty exclusion (full universe).
}

func (r *createReservationStaffRequest) toServiceInput() *CreateReservationStaffInput {
	return &CreateReservationStaffInput{
		Name:               r.Name,
		StaffType:          r.StaffType,
		ReservationVisible: resolveBoolDefaultTrue(r.ReservationVisible),
		ReservationComment: r.ReservationComment,
		SortOrder:          r.SortOrder,
	}
}

type updateReservationStaffRequest struct {
	Name               *string `json:"name"`
	StaffType          *string `json:"staff_type" binding:"omitempty,oneof=doctor nurse groomer other"`
	ReservationVisible *bool   `json:"reservation_visible"`
	ReservationComment *string `json:"reservation_comment"`
	SortOrder          *int    `json:"sort_order"`
	// TASK-021 UNIT-021-A: excluded_type_ids removed. Capability changes use capable-reservation-types.
}

func (r updateReservationStaffRequest) toServiceInput() *UpdateReservationStaffInput {
	return &UpdateReservationStaffInput{
		Name:               r.Name,
		StaffType:          r.StaffType,
		ReservationVisible: r.ReservationVisible,
		ReservationComment: r.ReservationComment,
		SortOrder:          r.SortOrder,
	}
}

type patchReservationStaffStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type patchReservationStaffSortOrderRequest struct {
	Direction string `json:"direction" binding:"required,oneof=up down"`
}
