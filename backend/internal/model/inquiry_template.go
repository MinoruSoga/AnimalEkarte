package model

import (
	"time"
)

// InquiryTemplate は問診定型文マスタ（v11.0追加）
type InquiryTemplate struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID  uint64    `gorm:"not null"                                       json:"clinic_id"`
	Category  string    `gorm:"default:''"                                     json:"category"`
	Title     string    `gorm:"not null"                                       json:"title"`
	Content   string    `gorm:"default:''"                                     json:"content"`
	IsActive  bool      `gorm:"default:true"                                   json:"is_active"`
	SortOrder int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (InquiryTemplate) TableName() string { return "inquiry_templates" }
