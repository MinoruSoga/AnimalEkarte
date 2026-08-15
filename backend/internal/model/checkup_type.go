package model

import (
	"time"

	"gorm.io/gorm"
)

type CheckupType struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID        uint64         `gorm:"not null"                                       json:"clinic_id"`
	Name            string         `gorm:"not null"                                       json:"name"`
	Price           *int64         `gorm:"type:bigint"                                    json:"price,omitempty"`
	IsActive        bool           `gorm:"default:true"                                   json:"is_active"`
	Description     string         `gorm:"default:''"                                     json:"description"`
	Interval        string         `gorm:"default:''"                                     json:"interval"`
	TargetAge       string         `gorm:"default:''"                                     json:"target_age"`
	ParentID        *uint64        `gorm:"column:parent_id"                               json:"parent_id,omitempty"`
	// ImportNamespace / ImportKey are nullable stable keys for versioned package import (TASK-374).
	ImportNamespace *string        `gorm:"column:import_namespace"                        json:"import_namespace,omitempty"`
	ImportKey       *string        `gorm:"column:import_key"                              json:"import_key,omitempty"`
	SortOrder       int            `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt       gorm.DeletedAt `                                                      json:"-"`
}

func (CheckupType) TableName() string { return "checkup_types" }
