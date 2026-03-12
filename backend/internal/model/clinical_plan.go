package model

import (
	"time"

	"github.com/google/uuid"
)

// ClinicalPlan は診察所見・診断・治療方針（診察/治療タブ, v9.0追加）
type ClinicalPlan struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex"                 json:"medical_record_id"`
	PhysicalExam        string     `gorm:"default:''"                                     json:"physical_exam"`
	DiagnosisCategoryID *uuid.UUID `gorm:"type:uuid"                                      json:"diagnosis_category_id,omitempty"`
	DiagnosisNameID     *uuid.UUID `gorm:"type:uuid"                                      json:"diagnosis_name_id,omitempty"`
	DiagnosisDetails    string     `gorm:"default:''"                                     json:"diagnosis_details"`
	TreatmentPolicy     string     `gorm:"default:''"                                     json:"treatment_policy"`
	CreatedAt           time.Time  `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord     *MedicalRecord     `gorm:"foreignKey:MedicalRecordID"     json:"medical_record,omitempty"`
	DiagnosisCategory *DiagnosisCategory `gorm:"foreignKey:DiagnosisCategoryID" json:"diagnosis_category,omitempty"`
	DiagnosisName     *DiagnosisName     `gorm:"foreignKey:DiagnosisNameID"     json:"diagnosis_name,omitempty"`
}

func (ClinicalPlan) TableName() string { return "clinical_plans" }
