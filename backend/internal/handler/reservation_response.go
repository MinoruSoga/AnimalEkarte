package handler

import (
	"encoding/json"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type reservationResponse struct {
	ID                uint64          `json:"id"`
	ClinicID          uint64          `json:"clinic_id"`
	StartTime         time.Time       `json:"start_time"`
	EndTime           time.Time       `json:"end_time"`
	OwnerID           *uint64         `json:"owner_id,omitempty"`
	PetID             *uint64         `json:"pet_id,omitempty"`
	VisitType         string          `json:"visit_type"`
	ReservationTypeID uint64          `json:"reservation_type_id"`
	DoctorID          *uint64         `json:"doctor_id,omitempty"`
	IsDesignated      bool            `json:"is_designated"`
	Status            string          `json:"status"`
	Notes             string          `json:"notes"`
	Source            string          `json:"source"`
	CreatedBy         *uint64         `json:"created_by,omitempty"`
	LineCustomerID    *uint64         `json:"line_customer_id,omitempty"`
	IsStaffDelegated  bool            `json:"is_staff_delegated"`
	CustomerFields    json.RawMessage `json:"customer_fields"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func toReservationResponse(r *model.Reservation) reservationResponse {
	return reservationResponse{
		ID:                r.ID,
		ClinicID:          r.ClinicID,
		StartTime:         r.StartTime,
		EndTime:           r.EndTime,
		OwnerID:           r.OwnerID,
		PetID:             r.PetID,
		VisitType:         string(r.VisitType),
		ReservationTypeID: r.ReservationTypeID,
		DoctorID:          r.DoctorID,
		IsDesignated:      r.IsDesignated,
		Status:            string(r.Status),
		Notes:             r.Notes,
		Source:            string(r.Source),
		CreatedBy:         r.CreatedBy,
		LineCustomerID:    r.LineCustomerID,
		IsStaffDelegated:  r.IsStaffDelegated,
		CustomerFields:    r.CustomerFields,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}
