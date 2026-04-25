package model

import (
	"time"

	"gorm.io/gorm"
)

// PetChronicCondition はペットの慢性疾患フラグ（LSTEP-BE-012）。
type PetChronicCondition struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement"  json:"id"`
	ClinicID      uint64         `gorm:"not null"                  json:"clinic_id"`
	PetID         uint64         `gorm:"not null"                  json:"pet_id"`
	ConditionCode string         `gorm:"size:50;not null"          json:"condition_code"`
	ConditionName string         `gorm:"size:100;not null"         json:"condition_name"`
	DiagnosedAt   time.Time      `gorm:"type:date;not null"        json:"diagnosed_at"`
	Notes         *string        `                                 json:"notes,omitempty"`
	IsActive      bool           `gorm:"not null;default:true"     json:"is_active"`
	DeletedAt     gorm.DeletedAt `                                 json:"-"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"            json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"            json:"updated_at"`

	// Relations
	Pet *Pet `gorm:"foreignKey:PetID" json:"pet,omitempty"`
}

func (PetChronicCondition) TableName() string { return "pet_chronic_conditions" }
