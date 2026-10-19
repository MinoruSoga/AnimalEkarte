package model

import (
	"time"
)

// ExamReferenceRange is a clinic-owned reference range for an examination field and species.
type ExamReferenceRange struct {
	ID              uint64        `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID        uint64        `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	ExamTypeFieldID uint64        `gorm:"column:exam_type_field_id;type:bigint;not null" json:"exam_type_field_id"`
	AnimalSpeciesID uint64        `gorm:"column:animal_species_id;type:bigint;not null" json:"animal_species_id"`
	RefMin          *ExactNumeric `gorm:"column:ref_min;type:decimal(10,4)" json:"ref_min,omitempty"`
	RefMax          *ExactNumeric `gorm:"column:ref_max;type:decimal(10,4)" json:"ref_max,omitempty"`
	QualitativeMin  *string       `gorm:"column:qualitative_min;type:text" json:"qualitative_min,omitempty"`
	QualitativeMax  *string       `gorm:"column:qualitative_max;type:text" json:"qualitative_max,omitempty"`
	CreatedAt       time.Time     `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time     `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (ExamReferenceRange) TableName() string { return "exam_reference_ranges" }
