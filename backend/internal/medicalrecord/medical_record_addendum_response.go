package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/httpapi"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type MedicalRecordAddendumResponse struct {
	ID              uint64    `json:"id"`
	MedicalRecordID uint64    `json:"medical_record_id"`
	ClinicID        uint64    `json:"clinic_id"`
	AuthorUserID    uint64    `json:"author_user_id"`
	BeforeText      string    `json:"before_text"`
	AfterText       string    `json:"after_text"`
	Reason          string    `json:"reason"`
	CreatedAt       time.Time `json:"created_at"`
}

func toMedicalRecordAddendumResponse(a *model.MedicalRecordAddendum) MedicalRecordAddendumResponse {
	return MedicalRecordAddendumResponse{
		ID:              a.ID,
		MedicalRecordID: a.MedicalRecordID,
		ClinicID:        a.ClinicID,
		AuthorUserID:    a.AuthorUserID,
		BeforeText:      a.BeforeText,
		AfterText:       a.AfterText,
		Reason:          a.Reason,
		CreatedAt:       httpapi.LocalTime(a.CreatedAt),
	}
}

func toMedicalRecordAddendumListResponse(addenda []*model.MedicalRecordAddendum) []MedicalRecordAddendumResponse {
	res := make([]MedicalRecordAddendumResponse, len(addenda))
	for i, a := range addenda {
		res[i] = toMedicalRecordAddendumResponse(a)
	}
	return res
}
