package model

import (
	"time"
)
type CheckupType struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	Price       *float64  `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	Description string    `gorm:"default:''"                                     json:"description"`
	Interval    string    `gorm:"default:''"                                     json:"interval"`
	TargetAge   string    `gorm:"default:''"                                     json:"target_age"`
	SortOrder   int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (CheckupType) TableName() string { return "checkup_types" }
