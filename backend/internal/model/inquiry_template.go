package model

import (
	"time"

	"github.com/google/uuid"
)

// InquiryTemplate は問診定型文マスタ（v11.0追加）
type InquiryTemplate struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID  uuid.UUID    `gorm:"type:uuid;not null"                             json:"clinic_id"`
	Category  string       `gorm:"default:''"                                     json:"category"`
	Title     string       `gorm:"not null"                                       json:"title"`
	Content   string       `gorm:"default:''"                                     json:"content"`
	Status    MasterStatus `gorm:"type:master_status;default:'active'"            json:"status"`
	SortOrder int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (InquiryTemplate) TableName() string { return "inquiry_templates" }
