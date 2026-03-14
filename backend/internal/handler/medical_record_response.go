package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type medicalRecordResponse struct {
	ID                       uint64                `json:"id"`
	ClinicID                 uint64                `json:"clinic_id"`
	RecordNo                 string                `json:"record_no"`
	Date                     time.Time             `json:"date"`
	OwnerID                  *uint64               `json:"owner_id,omitempty"`
	PetID                    *uint64               `json:"pet_id,omitempty"`
	DoctorID                 *uint64               `json:"doctor_id,omitempty"`
	ReservationAppointmentID *uint64               `json:"reservation_appointment_id,omitempty"`
	Status                   string                `json:"status"`
	CreatedAt                time.Time             `json:"created_at"`
	UpdatedAt                time.Time             `json:"updated_at"`
	Owner                    *ownerSummaryResponse `json:"owner,omitempty"`
	Pet                      *petSummaryResponse   `json:"pet,omitempty"`
	Doctor                   *staffSummaryResponse `json:"doctor,omitempty"`
}

func toMedicalRecordResponse(r *model.MedicalRecord) medicalRecordResponse {
	return medicalRecordResponse{
		ID:                       r.ID,
		ClinicID:                 r.ClinicID,
		RecordNo:                 r.RecordNo,
		Date:                     r.Date,
		OwnerID:                  r.OwnerID,
		PetID:                    r.PetID,
		DoctorID:                 r.DoctorID,
		ReservationAppointmentID: r.ReservationAppointmentID,
		Status:                   string(r.Status),
		CreatedAt:                r.CreatedAt,
		UpdatedAt:                r.UpdatedAt,
		Owner:                    toOwnerSummary(r.Owner),
		Pet:                      toPetSummary(r.Pet),
		Doctor:                   toStaffSummary(r.Doctor),
	}
}
