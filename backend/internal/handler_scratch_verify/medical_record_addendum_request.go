package handler

import "github.com/animal-ekarte/backend/internal/service"

type CreateMedicalRecordAddendumRequest struct {
	AfterText string `json:"after_text" binding:"required"`
	Reason    string `json:"reason"     binding:"required"`
}

func (r CreateMedicalRecordAddendumRequest) toServiceInput(medicalRecordID, authorUserID uint64) service.CreateMedicalRecordAddendumInput {
	return service.CreateMedicalRecordAddendumInput{
		MedicalRecordID: medicalRecordID,
		AuthorUserID:    authorUserID,
		AfterText:       r.AfterText,
		Reason:          r.Reason,
	}
}
