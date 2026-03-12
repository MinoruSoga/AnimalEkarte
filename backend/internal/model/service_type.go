package model

import (
	"time"

	"github.com/google/uuid"
)

type ServiceType struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID    uuid.UUID    `gorm:"type:uuid;not null"                             json:"clinic_id"`
	Code        string       `gorm:"default:''"                                     json:"code"`
	Name        string       `gorm:"not null"                                       json:"name"`
	IsActive    bool         `gorm:"default:true"                                   json:"is_active"`
	Description string       `gorm:"default:''"                                     json:"description"`
	Color       string       `gorm:"default:'#3B82F6'"                              json:"color"`
	SortOrder   int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (ServiceType) TableName() string { return "service_types" }
