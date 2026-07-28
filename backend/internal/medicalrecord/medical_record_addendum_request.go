package medicalrecord

type CreateMedicalRecordAddendumRequest struct {
	AfterText string `json:"after_text" binding:"required"`
	Reason    string `json:"reason"     binding:"required"`
}

func (r CreateMedicalRecordAddendumRequest) toServiceInput(medicalRecordID, authorUserID uint64) CreateMedicalRecordAddendumInput {
	return CreateMedicalRecordAddendumInput{
		MedicalRecordID: medicalRecordID,
		AuthorUserID:    authorUserID,
		AfterText:       r.AfterText,
		Reason:          r.Reason,
	}
}
