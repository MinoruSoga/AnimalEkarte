package model

import (
	"gorm.io/gorm"
	"time"
)

type CheckupType struct {
	ID              uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID        uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name            string         `gorm:"column:name;type:text;not null" json:"name"`
	Price           *int64         `gorm:"column:price;type:bigint" json:"price,omitempty"`
	IsActive        bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	Description     string         `gorm:"column:description;type:text;not null;default:''" json:"description"`
	Interval        string         `gorm:"column:interval;type:text;not null;default:''" json:"interval"`
	TargetAge       string         `gorm:"column:target_age;type:text;not null;default:''" json:"target_age"`
	ParentID        *uint64        `gorm:"column:parent_id;type:bigint" json:"parent_id,omitempty"`
	SortOrder       *int           `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt       time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	ImportNamespace *string        `gorm:"column:import_namespace;type:text" json:"import_namespace,omitempty"`
	ImportKey       *string        `gorm:"column:import_key;type:text" json:"import_key,omitempty"`
	IsNonInsurance  *bool          `gorm:"column:is_non_insurance;type:boolean" json:"-"`
	Version         int64          `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
}

func (CheckupType) TableName() string { return "checkup_types" }
