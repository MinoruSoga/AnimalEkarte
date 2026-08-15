package model

import (
	"time"

	"gorm.io/gorm"
)

// ClinicalPlan は診察所見・診断・治療方針（診察/治療タブ, v9.0追加）
type ClinicalPlan struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID  uint64         `gorm:"not null;uniqueIndex"                           json:"medical_record_id"`
	PhysicalExam     string         `gorm:"default:''"                                     json:"physical_exam"`
	DiagnosisTypeID  *uint64        `gorm:"column:diagnosis_type_id"                       json:"diagnosis_type_id,omitempty"`
	DiagnosisNameID  *uint64        `                                                      json:"diagnosis_name_id,omitempty"`
	Diagnosis2TypeID *uint64        `gorm:"column:diagnosis_2_type_id"                     json:"diagnosis_2_type_id,omitempty"`
	Diagnosis2NameID *uint64        `gorm:"column:diagnosis_2_name_id"                     json:"diagnosis_2_name_id,omitempty"`
	DiagnosisDetails string         `gorm:"default:''"                                     json:"diagnosis_details"`
	TreatmentPolicy  string         `gorm:"default:''"                                     json:"treatment_policy"`
	Version          int            `gorm:"default:1"                                      json:"version"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index"                                          json:"-"`

	// Relations
	MedicalRecord  *MedicalRecord `gorm:"foreignKey:MedicalRecordID"  json:"medical_record,omitempty"`
	DiagnosisType  *DiagnosisType `gorm:"foreignKey:DiagnosisTypeID"  json:"diagnosis_type,omitempty"`
	DiagnosisName  *DiagnosisName `gorm:"foreignKey:DiagnosisNameID"  json:"diagnosis_name,omitempty"`
	Diagnosis2Type *DiagnosisType `gorm:"foreignKey:Diagnosis2TypeID" json:"diagnosis_2_type,omitempty"`
	Diagnosis2Name *DiagnosisName `gorm:"foreignKey:Diagnosis2NameID" json:"diagnosis_2_name,omitempty"`
}

func (ClinicalPlan) TableName() string { return "clinical_plans" }
