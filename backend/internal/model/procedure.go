package model

import (
	"time"
)

// @name AnesthesiaType
type AnesthesiaType string

const (
	AnesthesiaTypeNone    AnesthesiaType = "none"
	AnesthesiaTypeLocal   AnesthesiaType = "local"
	AnesthesiaTypeGeneral AnesthesiaType = "general"
)

// @name Procedure
type Procedure struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64         `gorm:"not null"                                       json:"clinic_id"`
	Name        string         `gorm:"not null"                                       json:"name"`
	Price       *float64       `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool           `gorm:"default:true"                                   json:"is_active"`
	Description string         `gorm:"default:''"                                     json:"description"`
	Duration    *int           `gorm:"type:integer"                                   json:"duration,omitempty"`
	Anesthesia  AnesthesiaType `gorm:"type:anesthesia_type;default:'none'"            json:"anesthesia"`
	SortOrder   int            `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Procedure) TableName() string { return "procedures" }
