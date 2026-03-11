package model

import (
	"time"

	"github.com/google/uuid"
)

type AnesthesiaType string

const (
	AnesthesiaTypeNone    AnesthesiaType = "none"
	AnesthesiaTypeLocal   AnesthesiaType = "local"
	AnesthesiaTypeGeneral AnesthesiaType = "general"
)

type Procedure struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code        string         `gorm:"default:''"                                     json:"code"`
	Name        string         `gorm:"not null"                                       json:"name"`
	Price       *float64       `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	Status      MasterStatus   `gorm:"type:master_status;default:'active'"            json:"status"`
	Description string         `gorm:"default:''"                                     json:"description"`
	Duration    string         `gorm:"default:''"                                     json:"duration"`
	Anesthesia  AnesthesiaType `gorm:"type:anesthesia_type;default:'none'"            json:"anesthesia"`
	SortOrder   int            `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Procedure) TableName() string { return "procedures" }
