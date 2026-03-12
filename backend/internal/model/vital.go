package model

import (
	"time"

	"github.com/google/uuid"
)

// Vital は外来バイタル記録
type Vital struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID uuid.UUID  `gorm:"type:uuid;not null"                             json:"medical_record_id"`
	RecordedAt      time.Time  `gorm:"not null;default:now()"                         json:"recorded_at"`
	StaffID         *uuid.UUID `gorm:"type:uuid"                                      json:"staff_id,omitempty"`
	Temperature     *float64   `gorm:"type:numeric(4,1)"                              json:"temperature,omitempty"`
	HeartRate       *int       `gorm:"column:heart_rate"                              json:"heart_rate,omitempty"`
	RespirationRate *int       `gorm:"column:respiration_rate"                        json:"respiration_rate,omitempty"`
	Weight          *float64   `gorm:"type:numeric(6,2)"                              json:"weight,omitempty"`
	Notes           string     `gorm:"default:''"                                     json:"notes"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"                                 json:"created_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Staff         *Staff         `gorm:"foreignKey:StaffID"         json:"staff,omitempty"`
}

func (Vital) TableName() string { return "vitals" }
