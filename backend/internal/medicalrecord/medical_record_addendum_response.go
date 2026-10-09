package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/model"
)

type MedicalRecordAddendumResponse struct {
	ID              uint64 `json:"id"`
	MedicalRecordID uint64 `json:"medical_record_id"`
	ClinicID        uint64 `json:"clinic_id"`
	AuthorUserID    uint64 `json:"author_user_id"`
	// AuthorName は親カルテの clinic スコープで解決できた追記者氏名。
	// 未解決（削除・異動・他医院）または空氏名のときは nil → omitempty でキー省略し、
	// FE は author_user_id へフォールバック表示する（EMR-326）。
	AuthorName *string   `json:"author_name,omitempty"`
	BeforeText string    `json:"before_text"`
	AfterText  string    `json:"after_text"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}

func toMedicalRecordAddendumResponse(a *model.MedicalRecordAddendum) MedicalRecordAddendumResponse {
	res := MedicalRecordAddendumResponse{
		ID:              a.ID,
		MedicalRecordID: a.MedicalRecordID,
		ClinicID:        a.ClinicID,
		AuthorUserID:    a.AuthorUserID,
		BeforeText:      a.BeforeText,
		AfterText:       a.AfterText,
		Reason:          a.Reason,
		CreatedAt:       httpapi.LocalTime(a.CreatedAt),
	}
	// 空文字の氏名は送らない: FE は author_name ?? "スタッフID" でフォールバック
	// するため、"" を返すと ID 表示が空文字に潰される。
	if a.AuthorStaff != nil && a.AuthorStaff.Name != "" {
		name := a.AuthorStaff.Name
		res.AuthorName = &name
	}
	return res
}

func toMedicalRecordAddendumListResponse(addenda []*model.MedicalRecordAddendum) []MedicalRecordAddendumResponse {
	res := make([]MedicalRecordAddendumResponse, len(addenda))
	for i, a := range addenda {
		res[i] = toMedicalRecordAddendumResponse(a)
	}
	return res
}
