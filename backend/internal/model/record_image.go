package model

import (
	"time"

	"github.com/google/uuid"
)

// MedicalImageType は診療画像種別
type MedicalImageType string

const (
	MedicalImageTypeXray       MedicalImageType = "xray"
	MedicalImageTypeEcho       MedicalImageType = "echo"
	MedicalImageTypePhoto      MedicalImageType = "photo"
	MedicalImageTypeEndoscope  MedicalImageType = "endoscope"
	MedicalImageTypeCT         MedicalImageType = "ct"
	MedicalImageTypeMRI        MedicalImageType = "mri"
	MedicalImageTypeMicroscope MedicalImageType = "microscope"
	MedicalImageTypeOther      MedicalImageType = "other"
)

// RecordImage は診療画像（v7.0追加）
type RecordImage struct {
	ID              uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID uuid.UUID        `gorm:"type:uuid;not null"                             json:"medical_record_id"`
	ImageURL        string           `gorm:"not null;default:''"                            json:"image_url"`
	ThumbnailURL    string           `gorm:"default:''"                                     json:"thumbnail_url"`
	FileName        string           `gorm:"default:''"                                     json:"file_name"`
	FileSize        int64            `gorm:"column:file_size"                               json:"file_size"`
	MimeType        string           `gorm:"default:''"                                     json:"mime_type"`
	ImageType       MedicalImageType `gorm:"type:medical_image_type;default:'other'"        json:"image_type"`
	Description     string           `gorm:"default:''"                                     json:"description"`
	TakenAt         *time.Time       `gorm:"column:taken_at"                                json:"taken_at,omitempty"`
	ExamID          *uuid.UUID       `gorm:"type:uuid"                                      json:"exam_id,omitempty"`
	StaffID         *uuid.UUID       `gorm:"type:uuid"                                      json:"staff_id,omitempty"`
	SortOrder       int              `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt       time.Time        `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time        `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Exam          *Exam          `gorm:"foreignKey:ExamID"          json:"exam,omitempty"`
	Staff         *Staff         `gorm:"foreignKey:StaffID"         json:"staff,omitempty"`
}

func (RecordImage) TableName() string { return "record_images" }
