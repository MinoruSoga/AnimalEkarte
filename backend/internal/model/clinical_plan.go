package model

import (
	"time"
)

// ClinicalPlan は診察所見・診断・治療方針（診察/治療タブ, v9.0追加）
// @name ClinicalPlan
type ClinicalPlan struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID     uint64    `gorm:"not null;uniqueIndex"                           json:"medical_record_id"`
	PhysicalExam        string    `gorm:"default:''"                                     json:"physical_exam"`
	DiagnosisCategoryID *uint64   `                                                      json:"diagnosis_category_id,omitempty"`
	DiagnosisNameID     *uint64   `                                                      json:"diagnosis_name_id,omitempty"`
	DiagnosisDetails    string    `gorm:"default:''"                                     json:"diagnosis_details"`
	TreatmentPolicy     string    `gorm:"default:''"                                     json:"treatment_policy"`
	CreatedAt           time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord     *MedicalRecord     `gorm:"foreignKey:MedicalRecordID"     json:"medical_record,omitempty"`
	DiagnosisCategory *DiagnosisCategory `gorm:"foreignKey:DiagnosisCategoryID" json:"diagnosis_category,omitempty"`
	DiagnosisName     *DiagnosisName     `gorm:"foreignKey:DiagnosisNameID"     json:"diagnosis_name,omitempty"`
}

func (ClinicalPlan) TableName() string { return "clinical_plans" }
