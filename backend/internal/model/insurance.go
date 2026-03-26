package model

import (
	"time"
)

type Insurance struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID     uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name         string    `gorm:"not null"                                       json:"name"`
	IsActive     bool      `gorm:"default:true"                                   json:"is_active"`
	Description  string    `gorm:"default:''"                                     json:"description"`
	CoverageRate int       `gorm:"type:integer;not null;default:0"                json:"coverage_rate"`
	ContactPhone string    `gorm:"default:''"                                     json:"contact_phone"`
	SortOrder    int       `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt    time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Insurance) TableName() string { return "insurances" }
