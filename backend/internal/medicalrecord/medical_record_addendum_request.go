package medicalrecord

type CreateMedicalRecordAddendumRequest struct {
	AfterText string `json:"after_text" binding:"required,max=1000"`
	Reason    string `json:"reason"     binding:"required,max=500"`
}

func (r CreateMedicalRecordAddendumRequest) toServiceInput(medicalRecordID, authorUserID uint64) CreateMedicalRecordAddendumInput {
	return CreateMedicalRecordAddendumInput{
		MedicalRecordID: medicalRecordID,
		AuthorUserID:    authorUserID,
		AfterText:       r.AfterText,
		Reason:          r.Reason,
	}
}
