package handler

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type prescriptionResponse struct {
	ID              string  `json:"id"`
	ClinicID        string  `json:"clinic_id"`
	OwnerID         string  `json:"owner_id"`
	PetID           *string `json:"pet_id,omitempty"`
	MedicalRecordID *string `json:"medical_record_id,omitempty"`
	PrescribedAt    string  `json:"prescribed_at"`
	DurationDays    int     `json:"duration_days"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func toPrescriptionResponse(p *model.Prescription) prescriptionResponse {
	r := prescriptionResponse{
		ID:           strconv.FormatUint(p.ID, 10),
		ClinicID:     strconv.FormatUint(p.ClinicID, 10),
		OwnerID:      strconv.FormatUint(p.OwnerID, 10),
		PrescribedAt: p.PrescribedAt.Format("2006-01-02"),
		DurationDays: p.DurationDays,
		CreatedAt:    p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    p.UpdatedAt.Format(time.RFC3339),
	}
	if p.PetID != nil {
		s := strconv.FormatUint(*p.PetID, 10)
		r.PetID = &s
	}
	if p.MedicalRecordID != nil {
		s := strconv.FormatUint(*p.MedicalRecordID, 10)
		r.MedicalRecordID = &s
	}
	return r
}
