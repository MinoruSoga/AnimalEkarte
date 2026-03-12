package model

import (
	"time"

	"github.com/google/uuid"
)

// JobTitle は職種マスタ（v16.0 job_title ENUM廃止→マスタテーブル化）
type JobTitle struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID  uuid.UUID    `gorm:"type:uuid;not null"                             json:"clinic_id"`
	Name      string       `gorm:"not null"                                       json:"name"`
	Code      string       `gorm:"default:''"                                     json:"code"`
	SortOrder int          `gorm:"default:0"                                      json:"sort_order"`
	Status    MasterStatus `gorm:"type:master_status;default:'active'"            json:"status"`
	CreatedAt time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (JobTitle) TableName() string { return "job_titles" }
