package model

import (
	"time"

	"github.com/google/uuid"
)

type DiagnosisCategory struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code        string       `gorm:"default:''"                                     json:"code"`
	Name        string       `gorm:"not null"                                       json:"name"`
	Status      MasterStatus `gorm:"type:master_status;default:'active'"            json:"status"`
	Description string       `gorm:"default:''"                                     json:"description"`
	SortOrder   int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Names []DiagnosisName `gorm:"foreignKey:DiagnosisCategoryID" json:"names,omitempty"`
}

func (DiagnosisCategory) TableName() string { return "diagnosis_categories" }

type DiagnosisName struct {
	ID                  uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code                string       `gorm:"default:''"                                     json:"code"`
	Name                string       `gorm:"not null"                                       json:"name"`
	Status              MasterStatus `gorm:"type:master_status;default:'active'"            json:"status"`
	Description         string       `gorm:"default:''"                                     json:"description"`
	DiagnosisCategoryID uuid.UUID    `gorm:"type:uuid;not null"                             json:"diagnosis_category_id"`
	SortOrder           int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt           time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt           time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Category DiagnosisCategory `gorm:"foreignKey:DiagnosisCategoryID" json:"category,omitempty"`
}

func (DiagnosisName) TableName() string { return "diagnosis_names" }
