package model

import (
	"gorm.io/gorm"
	"time"
)

// PetChronicCondition はペットの慢性疾患フラグ（LSTEP-BE-012）。
type PetChronicCondition struct {
	ID            uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID      uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	PetID         uint64         `gorm:"column:pet_id;type:bigint;not null" json:"pet_id"`
	ConditionCode string         `gorm:"column:condition_code;type:varchar(50);not null" json:"condition_code"`
	ConditionName string         `gorm:"column:condition_name;type:varchar(100);not null" json:"condition_name"`
	DiagnosedAt   time.Time      `gorm:"column:diagnosed_at;type:date;not null" json:"diagnosed_at"`
	Notes         *string        `gorm:"column:notes;type:text" json:"notes,omitempty"`
	IsActive      bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	CreatedAt     time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Pet           *Pet           `gorm:"foreignKey:PetID;references:ID" json:"pet,omitempty"`
}

func (PetChronicCondition) TableName() string { return "pet_chronic_conditions" }
