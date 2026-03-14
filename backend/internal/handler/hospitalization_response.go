package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type hospitalizationResponse struct {
	ID                  uint64                `json:"id"`
	ClinicID            uint64                `json:"clinic_id"`
	OwnerID             uint64                `json:"owner_id"`
	PetID               uint64                `json:"pet_id"`
	HospitalizationType string                `json:"hospitalization_type"`
	StartDate           time.Time             `json:"start_date"`
	EndDate             time.Time             `json:"end_date"`
	Status              string                `json:"status"`
	CageID              *uint64               `json:"cage_id,omitempty"`
	DoctorID            *uint64               `json:"doctor_id,omitempty"`
	Memo                string                `json:"memo"`
	OwnerRequest        string                `json:"owner_request"`
	StaffNotes          string                `json:"staff_notes"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
	Owner               *ownerSummaryResponse `json:"owner,omitempty"`
	Pet                 *petSummaryResponse   `json:"pet,omitempty"`
	Cage                *cageSummaryResponse  `json:"cage,omitempty"`
	Doctor              *staffSummaryResponse `json:"doctor,omitempty"`
}

func toHospitalizationResponse(h *model.Hospitalization) hospitalizationResponse {
	return hospitalizationResponse{
		ID:                  h.ID,
		ClinicID:            h.ClinicID,
		OwnerID:             h.OwnerID,
		PetID:               h.PetID,
		HospitalizationType: string(h.HospitalizationType),
		StartDate:           h.StartDate,
		EndDate:             h.EndDate,
		Status:              string(h.Status),
		CageID:              h.CageID,
		DoctorID:            h.DoctorID,
		Memo:                h.Memo,
		OwnerRequest:        h.OwnerRequest,
		StaffNotes:          h.StaffNotes,
		CreatedAt:           h.CreatedAt,
		UpdatedAt:           h.UpdatedAt,
		Owner:               toOwnerSummary(h.Owner),
		Pet:                 toPetSummary(h.Pet),
		Cage:                toCageSummary(h.Cage),
		Doctor:              toStaffSummary(h.Doctor),
	}
}
