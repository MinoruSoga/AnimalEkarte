package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type inquiryResponse struct {
	ID                   uint64    `json:"id"`
	MedicalRecordID      uint64    `json:"medical_record_id"`
	ChiefComplaintTypeID *uint64   `json:"chief_complaint_type_id,omitempty"`
	ChiefComplaint       string    `json:"chief_complaint"`
	History              string    `json:"history"`
	CurrentMedications   string    `json:"current_medications"`
	AllergyInfo          string    `json:"allergy_info"`
	LastMeal             string    `json:"last_meal"`
	LastDefecation       string    `json:"last_defecation"`
	LastUrination        string    `json:"last_urination"`
	Appetite             *string   `json:"appetite,omitempty"`
	WaterIntake          *string   `json:"water_intake,omitempty"`
	OwnerObservations    string    `json:"owner_observations"`
	Notes                string    `json:"notes"`
	StaffID              *uint64   `json:"staff_id,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func toInquiryResponse(inquiry *model.Inquiry) inquiryResponse {
	var appetite *string
	if inquiry.Appetite != nil {
		s := string(*inquiry.Appetite)
		appetite = &s
	}
	var waterIntake *string
	if inquiry.WaterIntake != nil {
		s := string(*inquiry.WaterIntake)
		waterIntake = &s
	}
	return inquiryResponse{
		ID:                   inquiry.ID,
		MedicalRecordID:      inquiry.MedicalRecordID,
		ChiefComplaintTypeID: inquiry.ChiefComplaintTypeID,
		ChiefComplaint:       inquiry.ChiefComplaint,
		History:              inquiry.History,
		CurrentMedications:   inquiry.CurrentMedications,
		AllergyInfo:          inquiry.AllergyInfo,
		LastMeal:             inquiry.LastMeal,
		LastDefecation:       inquiry.LastDefecation,
		LastUrination:        inquiry.LastUrination,
		Appetite:             appetite,
		WaterIntake:          waterIntake,
		OwnerObservations:    inquiry.OwnerObservations,
		Notes:                inquiry.Notes,
		StaffID:              inquiry.StaffID,
		CreatedAt:            httpapi.LocalTime(inquiry.CreatedAt),
		UpdatedAt:            httpapi.LocalTime(inquiry.UpdatedAt),
	}
}
