package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Checkup struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID uuid.UUID      `gorm:"type:uuid;not null"                             json:"medical_record_id"`
	PetID           *uuid.UUID     `gorm:"type:uuid"                                      json:"pet_id,omitempty"`
	CheckupTypeID   uuid.UUID      `gorm:"type:uuid;not null"                             json:"checkup_type_id"`
	Date            time.Time      `gorm:"type:date;not null"                             json:"date"`
	DoctorID        *uuid.UUID     `gorm:"type:uuid"                                      json:"doctor_id,omitempty"`
	Result          string         `gorm:"default:''"                                     json:"result"`
	NextDate        *time.Time     `gorm:"type:date"                                      json:"next_date,omitempty"`
	DeletedAt       gorm.DeletedAt `                                                      json:"deleted_at"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	CheckupType   *CheckupType   `gorm:"foreignKey:CheckupTypeID"   json:"checkup_type,omitempty"`
	Doctor        *Staff         `gorm:"foreignKey:DoctorID"        json:"doctor,omitempty"`
}

func (Checkup) TableName() string { return "checkups" }
