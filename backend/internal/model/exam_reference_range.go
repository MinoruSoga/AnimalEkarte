package model

import "time"

// ExamReferenceRange is a clinic-owned reference range for an examination field and species.
type ExamReferenceRange struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"     json:"id"`
	ClinicID        uint64    `gorm:"not null"                     json:"-"`
	ExamTypeFieldID uint64    `gorm:"not null"                     json:"exam_type_field_id"`
	AnimalSpeciesID uint64    `gorm:"not null"                     json:"animal_species_id"`
	RefMin          *float64  `gorm:"type:decimal(10,4)"           json:"ref_min,omitempty"`
	RefMax          *float64  `gorm:"type:decimal(10,4)"           json:"ref_max,omitempty"`
	QualitativeMin  *string   `                                  json:"qualitative_min,omitempty"`
	QualitativeMax  *string   `                                  json:"qualitative_max,omitempty"`
	CreatedAt       time.Time `gorm:"autoCreateTime"               json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"               json:"updated_at"`
}

func (ExamReferenceRange) TableName() string { return "exam_reference_ranges" }
