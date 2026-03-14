package model

import (
	"time"
)

// JobTitle は職種マスタ（v16.0 job_title ENUM廃止→マスタテーブル化）
type JobTitle struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	Description string    `gorm:"default:''"                                     json:"description"`
	SortOrder   int       `gorm:"default:0"                                      json:"sort_order"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (JobTitle) TableName() string { return "job_titles" }
