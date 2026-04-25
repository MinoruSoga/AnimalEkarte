package model

import (
	"time"

	"gorm.io/gorm"
)

// Prescription は処方薬記録（LSTEP-BE-009）
type Prescription struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement"   json:"id"`
	ClinicID        uint64         `gorm:"not null"                   json:"clinic_id"`
	OwnerID         uint64         `gorm:"not null"                   json:"owner_id"`
	PetID           *uint64        `                                  json:"pet_id,omitempty"`
	MedicalRecordID *uint64        `                                  json:"medical_record_id,omitempty"`
	PrescribedAt    time.Time      `gorm:"type:date;not null"         json:"prescribed_at"`
	DurationDays    int            `gorm:"not null;default:0"         json:"duration_days"`
	DeletedAt       gorm.DeletedAt `                                  json:"-"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"             json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"             json:"updated_at"`

	// Relations
	Owner         *Owner         `gorm:"foreignKey:OwnerID"         json:"owner,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
}

func (Prescription) TableName() string { return "prescriptions" }
