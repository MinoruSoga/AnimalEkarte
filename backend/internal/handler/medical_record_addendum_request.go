package handler

type CreateMedicalRecordAddendumRequest struct {
	AfterText string `json:"after_text" binding:"required"`
	Reason    string `json:"reason"     binding:"required"`
}
