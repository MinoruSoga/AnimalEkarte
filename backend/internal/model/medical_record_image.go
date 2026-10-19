package model

import (
	"time"
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

// MedicalRecordImage は診療画像（v7.0追加）
type MedicalRecordImage struct {
	ID              uint64           `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	MedicalRecordID uint64           `gorm:"column:medical_record_id;type:bigint;not null" json:"medical_record_id"`
	ImageURL        string           `gorm:"column:image_url;type:text;not null;default:''" json:"image_url"`
	ThumbnailURL    string           `gorm:"column:thumbnail_url;type:text;not null;default:''" json:"thumbnail_url"`
	FileName        string           `gorm:"column:file_name;type:text;not null;default:''" json:"file_name"`
	FileSize        *int64           `gorm:"column:file_size;type:bigint;default:0" json:"file_size"`
	MimeType        string           `gorm:"column:mime_type;type:text;not null;default:''" json:"mime_type"`
	ImageType       MedicalImageType `gorm:"column:image_type;type:medical_image_type;not null;default:'other'" json:"image_type"`
	Description     string           `gorm:"column:description;type:text;not null;default:''" json:"description"`
	TakenAt         *time.Time       `gorm:"column:taken_at;type:timestamptz" json:"taken_at,omitempty"`
	ExamID          *uint64          `gorm:"column:exam_id;type:bigint" json:"exam_id,omitempty"`
	StaffID         *uint64          `gorm:"column:staff_id;type:bigint" json:"staff_id,omitempty"`
	SortOrder       *int             `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt       time.Time        `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time        `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	MedicalRecord   *MedicalRecord   `gorm:"foreignKey:MedicalRecordID;references:ID" json:"medical_record,omitempty"`
	Examination     *Examination     `gorm:"foreignKey:ExamID;references:ID"          json:"exam,omitempty"`
	Staff           *Staff           `gorm:"foreignKey:StaffID;references:ID"         json:"staff,omitempty"`
}

func (MedicalRecordImage) TableName() string { return "medical_record_images" }
