package model

import (
	"time"
)

type DiagnosisType struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	Description string    `gorm:"default:''"                                     json:"description"`
	SortOrder   int       `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Names []DiagnosisName `gorm:"foreignKey:DiagnosisTypeID" json:"names,omitempty"`
}

func (DiagnosisType) TableName() string { return "diagnosis_types" }

type DiagnosisName struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID        uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name            string    `gorm:"not null"                                       json:"name"`
	IsActive        bool      `gorm:"default:true"                                   json:"is_active"`
	Description     string    `gorm:"default:''"                                     json:"description"`
	DiagnosisTypeID uint64    `gorm:"column:diagnosis_type_id;not null"              json:"diagnosis_type_id"`
	SortOrder       int       `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt       time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Category *DiagnosisType `gorm:"foreignKey:DiagnosisTypeID" json:"category,omitempty"`
}

func (DiagnosisName) TableName() string { return "diagnosis_names" }
