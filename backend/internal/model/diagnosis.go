package model

import (
	"time"
)

type DiagnosisCategory struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	Description string    `gorm:"default:''"                                     json:"description"`
	SortOrder   int       `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Names []DiagnosisName `gorm:"foreignKey:DiagnosisCategoryID" json:"names,omitempty"`
}

func (DiagnosisCategory) TableName() string { return "diagnosis_categories" }

type DiagnosisName struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID            uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name                string    `gorm:"not null"                                       json:"name"`
	IsActive            bool      `gorm:"default:true"                                   json:"is_active"`
	Description         string    `gorm:"default:''"                                     json:"description"`
	DiagnosisCategoryID uint64    `gorm:"not null"                                       json:"diagnosis_category_id"`
	SortOrder           int       `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt           time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Category *DiagnosisCategory `gorm:"foreignKey:DiagnosisCategoryID" json:"category,omitempty"`
}

func (DiagnosisName) TableName() string { return "diagnosis_names" }
