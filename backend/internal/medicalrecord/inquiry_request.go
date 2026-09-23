package medicalrecord

type updateInquiryRequest struct {
	ChiefComplaint *string `json:"chief_complaint"`
	// ChiefComplaintTypeID は PATCH DTO で「未送信」(保持) と「JSON null」(解除) を区別する。
	// clinical_plan_request.go の diagnosis_2_*_id と同じ nullableUint64RequestField 契約。
	ChiefComplaintTypeID nullableUint64RequestField `json:"chief_complaint_type_id"`
	Notes                *string                    `json:"notes"`
}

func (r updateInquiryRequest) toServiceInput(clinicID, medicalRecordID uint64) UpsertInquiryInput {
	return UpsertInquiryInput{
		ClinicID:             clinicID,
		MedicalRecordID:      medicalRecordID,
		ChiefComplaintTypeID: r.ChiefComplaintTypeID.toServiceInput(),
		ChiefComplaint:       r.ChiefComplaint,
		Notes:                r.Notes,
	}
}
