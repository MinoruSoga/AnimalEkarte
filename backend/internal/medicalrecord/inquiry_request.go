package medicalrecord

type updateInquiryRequest struct {
	ChiefComplaint       *string `json:"chief_complaint"`
	ChiefComplaintTypeID *uint64 `json:"chief_complaint_type_id"`
	Notes                *string `json:"notes"`
}

func (r updateInquiryRequest) toServiceInput(clinicID, medicalRecordID uint64) UpsertInquiryInput {
	return UpsertInquiryInput{
		ClinicID:             clinicID,
		MedicalRecordID:      medicalRecordID,
		ChiefComplaintTypeID: r.ChiefComplaintTypeID,
		ChiefComplaint:       r.ChiefComplaint,
		Notes:                r.Notes,
	}
}
