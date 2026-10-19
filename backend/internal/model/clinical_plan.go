package model

import (
	"gorm.io/gorm"
	"time"
)

// ClinicalPlan は診察所見・診断・治療方針（診察/治療タブ, v9.0追加）
type ClinicalPlan struct {
	ID               uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	MedicalRecordID  uint64         `gorm:"column:medical_record_id;type:bigint;not null" json:"medical_record_id"`
	PhysicalExam     string         `gorm:"column:physical_exam;type:text;not null;default:''" json:"physical_exam"`
	DiagnosisDetails string         `gorm:"column:diagnosis_details;type:text;not null;default:''" json:"diagnosis_details"`
	TreatmentPolicy  string         `gorm:"column:treatment_policy;type:text;not null;default:''" json:"treatment_policy"`
	Version          int            `gorm:"column:version;type:integer;not null;default:1" json:"version"`
	CreatedAt        time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	ClinicID         uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	MedicalRecord    *MedicalRecord `gorm:"foreignKey:MedicalRecordID,ClinicID;references:ID,ClinicID"  json:"medical_record,omitempty"`
	DiagnosisType    *DiagnosisType `gorm:"-"  json:"diagnosis_type,omitempty"`
	DiagnosisName    *DiagnosisName `gorm:"-"  json:"diagnosis_name,omitempty"`
	Diagnosis2Type   *DiagnosisType `gorm:"-" json:"diagnosis_2_type,omitempty"`
	Diagnosis2Name   *DiagnosisName `gorm:"-" json:"diagnosis_2_name,omitempty"`
}

func (ClinicalPlan) TableName() string { return "clinical_plans" }
