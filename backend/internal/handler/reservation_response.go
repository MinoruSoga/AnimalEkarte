package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type reservationResponse struct {
	ID            uint64                      `json:"id"`
	ClinicID      uint64                      `json:"clinic_id"`
	StartTime     time.Time                   `json:"start_time"`
	EndTime       time.Time                   `json:"end_time"`
	OwnerID       *uint64                     `json:"owner_id,omitempty"`
	PetID         *uint64                     `json:"pet_id,omitempty"`
	VisitType     string                      `json:"visit_type"`
	ServiceTypeID uint64                      `json:"service_type_id"`
	DoctorID      *uint64                     `json:"doctor_id,omitempty"`
	IsDesignated  bool                        `json:"is_designated"`
	Status        string                      `json:"status"`
	Notes         string                      `json:"notes"`
	CreatedAt     time.Time                   `json:"created_at"`
	UpdatedAt     time.Time                   `json:"updated_at"`
	Pet           *petSummaryResponse         `json:"pet,omitempty"`
	ServiceType   *serviceTypeSummaryResponse `json:"service_type,omitempty"`
	Doctor        *staffSummaryResponse       `json:"doctor,omitempty"`
}

func toReservationResponse(r *model.ReservationAppointment) reservationResponse {
	return reservationResponse{
		ID:            r.ID,
		ClinicID:      r.ClinicID,
		StartTime:     r.StartTime,
		EndTime:       r.EndTime,
		OwnerID:       r.OwnerID,
		PetID:         r.PetID,
		VisitType:     string(r.VisitType),
		ServiceTypeID: r.ServiceTypeID,
		DoctorID:      r.DoctorID,
		IsDesignated:  r.IsDesignated,
		Status:        string(r.Status),
		Notes:         r.Notes,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
		Pet:           toPetSummary(r.Pet),
		ServiceType:   toServiceTypeSummary(r.ServiceType),
		Doctor:        toStaffSummary(r.Doctor),
	}
}
