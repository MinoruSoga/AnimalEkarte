package model

import (
	"time"

	"github.com/google/uuid"
)

type Consultation struct {
	ID            uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID      uuid.UUID    `gorm:"type:uuid;not null"                             json:"clinic_id"`
	Code          string       `gorm:"default:''"                                     json:"code"`
	Name          string       `gorm:"not null"                                       json:"name"`
	Price         *float64     `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive      bool         `gorm:"default:true"                                   json:"is_active"`
	Description   string       `gorm:"default:''"                                     json:"description"`
	TimeCondition string       `gorm:"default:''"                                     json:"time_condition"`
	Duration      *int         `gorm:"type:integer"                                   json:"duration,omitempty"`
	SortOrder     int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt     time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt     time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Consultation) TableName() string { return "consultations" }
