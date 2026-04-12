package handler

import "time"

type createReservationAdminRequest struct {
	StartTime         time.Time      `json:"start_time"         binding:"required"`
	EndTime           time.Time      `json:"end_time"           binding:"required"`
	OwnerID           *uint64        `json:"owner_id"`
	PetID             *uint64        `json:"pet_id"`
	VisitType         string         `json:"visit_type"`
	ReservationTypeID uint64         `json:"reservation_type_id"    binding:"required"`
	DoctorID          *uint64        `json:"doctor_id"`
	IsDesignated      bool           `json:"is_designated"`
	Notes             string         `json:"notes"`
	LineCustomerID    *uint64        `json:"line_customer_id"`
	IsStaffDelegated  bool           `json:"is_staff_delegated"`
	CustomerFields    jsonRawOrEmpty `json:"customer_fields"`
}
