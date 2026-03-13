package model

import (
	"time"
)

// AppetiteLevel は食欲レベル
// @name AppetiteLevel
type AppetiteLevel string

const (
	AppetiteLevelNormal    AppetiteLevel = "normal"
	AppetiteLevelIncreased AppetiteLevel = "increased"
	AppetiteLevelDecreased AppetiteLevel = "decreased"
	AppetiteLevelNone      AppetiteLevel = "none"
)

// WaterIntakeLevel は水分摂取レベル
// @name WaterIntakeLevel
type WaterIntakeLevel string

const (
	WaterIntakeLevelNormal    WaterIntakeLevel = "normal"
	WaterIntakeLevelIncreased WaterIntakeLevel = "increased"
	WaterIntakeLevelDecreased WaterIntakeLevel = "decreased"
	WaterIntakeLevelNone      WaterIntakeLevel = "none"
)

// Inquiry は問診情報（カルテ問診タブ, v7.0追加）
// @name Inquiry
type Inquiry struct {
	ID                       uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID          uint64            `gorm:"not null"                                       json:"medical_record_id"`
	ChiefComplaintCategoryID *uint64           `                                                      json:"chief_complaint_category_id,omitempty"`
	ChiefComplaint           string            `gorm:"default:''"                                     json:"chief_complaint"`
	History                  string            `gorm:"default:''"                                     json:"history"`
	CurrentMedications       string            `gorm:"default:''"                                     json:"current_medications"`
	AllergyInfo              string            `gorm:"default:''"                                     json:"allergy_info"`
	LastMeal                 string            `gorm:"default:''"                                     json:"last_meal"`
	LastDefecation           string            `gorm:"default:''"                                     json:"last_defecation"`
	LastUrination            string            `gorm:"default:''"                                     json:"last_urination"`
	Appetite                 *AppetiteLevel    `gorm:"type:appetite_level"                            json:"appetite,omitempty"`
	WaterIntake              *WaterIntakeLevel `gorm:"type:water_intake_level"                       json:"water_intake,omitempty"`
	OwnerObservations        string            `gorm:"default:''"                                     json:"owner_observations"`
	Notes                    string            `gorm:"default:''"                                     json:"notes"`
	StaffID                  *uint64           `                                                      json:"staff_id,omitempty"`
	CreatedAt                time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt                time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord          *MedicalRecord          `gorm:"foreignKey:MedicalRecordID"          json:"medical_record,omitempty"`
	ChiefComplaintCategory *ChiefComplaintCategory `gorm:"foreignKey:ChiefComplaintCategoryID" json:"chief_complaint_category,omitempty"`
	Staff                  *Staff                  `gorm:"foreignKey:StaffID"                  json:"staff,omitempty"`
}

func (Inquiry) TableName() string { return "inquiries" }
