package model

import (
	"time"
)

// ChiefComplaintCategory は主訴区分マスタ（v11.0追加）
type ChiefComplaintCategory struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	Description string    `gorm:"default:''"                                     json:"description"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	SortOrder   int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (ChiefComplaintCategory) TableName() string { return "chief_complaint_categories" }
