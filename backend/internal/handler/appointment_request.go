package handler

import "time"

// createReservationRequest は予約作成のバインド struct
type createReservationRequest struct {
	StartTime         time.Time `json:"start_time"      binding:"required"`
	EndTime           time.Time `json:"end_time"        binding:"required"`
	OwnerID           *uint64   `json:"owner_id"`
	PetID             *uint64   `json:"pet_id"`
	VisitType         string    `json:"visit_type"          binding:"omitempty,oneof=first revisit"`
	ReservationTypeID uint64    `json:"reservation_type_id" binding:"required"`
	DoctorID          *uint64   `json:"doctor_id"`
	IsDesignated      bool      `json:"is_designated"`
	Status            string    `json:"status"              binding:"omitempty,oneof=confirmed pending cancelled checked_in in_consultation accounting completed"`
	Notes             string    `json:"notes"`
	Source            string    `json:"source"              binding:"omitempty,oneof=manual line"`
}

// patchReservationReservationRouteRequest は予約経路更新リクエスト（FEAT-381-2）。
type patchReservationReservationRouteRequest struct {
	Route string `json:"route" binding:"max=20"`
}

// updateReservationRequest は予約更新のバインド struct
type updateReservationRequest struct {
	StartTime         *time.Time `json:"start_time"`
	EndTime           *time.Time `json:"end_time"`
	OwnerID           *uint64    `json:"owner_id"`
	PetID             *uint64    `json:"pet_id"`
	VisitType         *string    `json:"visit_type"           binding:"omitempty,oneof=first revisit"`
	ReservationTypeID *uint64    `json:"reservation_type_id"`
	DoctorID          *uint64    `json:"doctor_id"`
	IsDesignated      *bool      `json:"is_designated"`
	Status            *string    `json:"status"               binding:"omitempty,oneof=confirmed pending cancelled checked_in in_consultation accounting completed"`
	Notes             *string    `json:"notes"`
}
