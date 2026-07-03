package handler

import "github.com/animal-ekarte/backend/internal/service"

type updateInquiryRequest struct {
	ChiefComplaint       *string `json:"chief_complaint"`
	ChiefComplaintTypeID *uint64 `json:"chief_complaint_type_id"`
	Notes                *string `json:"notes"`
}

func (r updateInquiryRequest) toServiceInput(clinicID, medicalRecordID uint64) service.UpsertInquiryInput {
	return service.UpsertInquiryInput{
		ClinicID:             clinicID,
		MedicalRecordID:      medicalRecordID,
		ChiefComplaintTypeID: r.ChiefComplaintTypeID,
		ChiefComplaint:       r.ChiefComplaint,
		Notes:                r.Notes,
	}
}
