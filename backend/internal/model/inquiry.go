package model

import (
	"time"
)

// AppetiteLevel は食欲レベル
type AppetiteLevel string

const (
	AppetiteLevelNormal    AppetiteLevel = "normal"
	AppetiteLevelIncreased AppetiteLevel = "increased"
	AppetiteLevelDecreased AppetiteLevel = "decreased"
	AppetiteLevelNone      AppetiteLevel = "none"
)

// WaterIntakeLevel は水分摂取レベル
type WaterIntakeLevel string

const (
	WaterIntakeLevelNormal    WaterIntakeLevel = "normal"
	WaterIntakeLevelIncreased WaterIntakeLevel = "increased"
	WaterIntakeLevelDecreased WaterIntakeLevel = "decreased"
	WaterIntakeLevelNone      WaterIntakeLevel = "none"
)

// Inquiry は問診情報（カルテ問診タブ, v7.0追加）
type Inquiry struct {
	ID                   uint64              `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	MedicalRecordID      uint64              `gorm:"column:medical_record_id;type:bigint;not null" json:"medical_record_id"`
	ChiefComplaintTypeID *uint64             `gorm:"column:chief_complaint_type_id;type:bigint" json:"chief_complaint_type_id,omitempty"`
	ChiefComplaint       string              `gorm:"column:chief_complaint;type:text;not null;default:''" json:"chief_complaint"`
	History              string              `gorm:"column:history;type:text;not null;default:''" json:"history"`
	CurrentMedications   string              `gorm:"column:current_medications;type:text;not null;default:''" json:"current_medications"`
	AllergyInfo          string              `gorm:"column:allergy_info;type:text;not null;default:''" json:"allergy_info"`
	LastMeal             string              `gorm:"column:last_meal;type:text;not null;default:''" json:"last_meal"`
	LastDefecation       string              `gorm:"column:last_defecation;type:text;not null;default:''" json:"last_defecation"`
	LastUrination        string              `gorm:"column:last_urination;type:text;not null;default:''" json:"last_urination"`
	Appetite             *AppetiteLevel      `gorm:"column:appetite;type:appetite_level" json:"appetite,omitempty"`
	WaterIntake          *WaterIntakeLevel   `gorm:"column:water_intake;type:water_intake_level" json:"water_intake,omitempty"`
	OwnerObservations    string              `gorm:"column:owner_observations;type:text;not null;default:''" json:"owner_observations"`
	Notes                string              `gorm:"column:notes;type:text;not null;default:''" json:"notes"`
	StaffID              *uint64             `gorm:"column:staff_id;type:bigint" json:"staff_id,omitempty"`
	CreatedAt            time.Time           `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time           `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	MedicalRecord        *MedicalRecord      `gorm:"foreignKey:MedicalRecordID;references:ID"          json:"medical_record,omitempty"`
	ChiefComplaintType   *ChiefComplaintType `gorm:"foreignKey:ChiefComplaintTypeID;references:ID" json:"chief_complaint_type,omitempty"`
	Staff                *Staff              `gorm:"foreignKey:StaffID;references:ID"                  json:"staff,omitempty"`
}

func (Inquiry) TableName() string { return "inquiries" }
