package model

import (
	"time"

	"github.com/google/uuid"
)

type Insurance struct {
	ID           uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID     uuid.UUID    `gorm:"type:uuid;not null"                             json:"clinic_id"`
	Code         string       `gorm:"default:''"                                     json:"code"`
	Name         string       `gorm:"not null"                                       json:"name"`
	Status       MasterStatus `gorm:"type:master_status;default:'active'"            json:"status"`
	Description  string       `gorm:"default:''"                                     json:"description"`
	CoverageRate *int         `gorm:"column:coverage_rate"                           json:"coverage_rate,omitempty"`
	ContactPhone string       `gorm:"default:''"                                     json:"contact_phone"`
	SortOrder    int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt    time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt    time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Insurance) TableName() string { return "insurances" }
