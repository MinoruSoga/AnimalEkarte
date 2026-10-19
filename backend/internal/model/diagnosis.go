package model

import (
	"gorm.io/gorm"
	"time"
)

type DiagnosisType struct {
	ID          uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID    uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name        string          `gorm:"column:name;type:text;not null" json:"name"`
	IsActive    bool            `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	Description string          `gorm:"column:description;type:text;not null;default:''" json:"description"`
	SortOrder   *int            `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt   time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Names       []DiagnosisName `gorm:"foreignKey:DiagnosisTypeID;references:ID" json:"names,omitempty"`
}

func (DiagnosisType) TableName() string { return "diagnosis_types" }

type DiagnosisName struct {
	ID              uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID        uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Name            string         `gorm:"column:name;type:text;not null" json:"name"`
	IsActive        bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	Description     string         `gorm:"column:description;type:text;not null;default:''" json:"description"`
	DiagnosisTypeID uint64         `gorm:"column:diagnosis_type_id;type:bigint;not null" json:"diagnosis_type_id"`
	SortOrder       *int           `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt       time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Category        *DiagnosisType `gorm:"foreignKey:DiagnosisTypeID;references:ID" json:"category,omitempty"`
}

func (DiagnosisName) TableName() string { return "diagnosis_names" }
