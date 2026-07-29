package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// medicalRecordImageResponse は診療画像の API レスポンス
type medicalRecordImageResponse struct {
	ID              uint64                `json:"id"`
	MedicalRecordID uint64                `json:"medical_record_id"`
	ImageURL        string                `json:"image_url"`
	ThumbnailURL    string                `json:"thumbnail_url"`
	FileName        string                `json:"file_name"`
	FileSize        int64                 `json:"file_size"`
	MimeType        string                `json:"mime_type"`
	ImageType       string                `json:"image_type"`
	Description     string                `json:"description"`
	TakenAt         *time.Time            `json:"taken_at,omitempty"`
	ExamID          *uint64               `json:"exam_id,omitempty"`
	StaffID         *uint64               `json:"staff_id,omitempty"`
	SortOrder       int                   `json:"sort_order"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	Staff           *StaffSummaryResponse `json:"staff,omitempty"`
}

// toMedicalRecordImageResponse は model.MedicalRecordImage を medicalRecordImageResponse に変換する
func toMedicalRecordImageResponse(img *model.MedicalRecordImage) medicalRecordImageResponse {
	r := medicalRecordImageResponse{
		ID:              img.ID,
		MedicalRecordID: img.MedicalRecordID,
		ImageURL:        img.ImageURL,
		ThumbnailURL:    img.ThumbnailURL,
		FileName:        img.FileName,
		FileSize:        img.FileSize,
		MimeType:        img.MimeType,
		ImageType:       string(img.ImageType),
		Description:     img.Description,
		TakenAt:         httpapi.LocalTimePtr(img.TakenAt),
		ExamID:          img.ExamID,
		StaffID:         img.StaffID,
		SortOrder:       img.SortOrder,
		CreatedAt:       httpapi.LocalTime(img.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(img.UpdatedAt),
	}
	if img.Staff != nil {
		r.Staff = toStaffSummary(img.Staff)
	}
	return r
}
