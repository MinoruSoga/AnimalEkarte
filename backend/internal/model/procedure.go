package model

import (
	"time"

	"gorm.io/gorm"
)

type AnesthesiaType string

const (
	AnesthesiaTypeNone     AnesthesiaType = "none"
	AnesthesiaTypeLocal    AnesthesiaType = "local"
	AnesthesiaTypeSedation AnesthesiaType = "sedation"
	AnesthesiaTypeGeneral  AnesthesiaType = "general"
)

type Procedure struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64         `gorm:"not null"                                       json:"clinic_id"`
	Name        string         `gorm:"not null"                                       json:"name"`
	Price       *int64         `gorm:"type:bigint"                                    json:"price,omitempty"`
	IsActive    bool           `gorm:"default:true"                                   json:"is_active"`
	Description string         `gorm:"default:''"                                     json:"description"`
	Duration    *int           `gorm:"type:integer"                                   json:"duration,omitempty"`
	Anesthesia  AnesthesiaType `gorm:"type:anesthesia_type;default:'none'"            json:"anesthesia"`
	ParentID    *uint64        `gorm:"column:parent_id"                               json:"parent_id,omitempty"`
	TaxType     TaxType        `gorm:"type:tax_type;not null;default:excluded"        json:"tax_type"`
	TaxRate     float64        `gorm:"type:numeric;not null;default:0.10"             json:"tax_rate"`
	SortOrder   int            `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt   gorm.DeletedAt `                                                      json:"-"`
}

func (Procedure) TableName() string { return "procedures" }
