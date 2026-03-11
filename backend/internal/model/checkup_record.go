package model

import (
	"time"

	"github.com/google/uuid"
)

type CheckupRecord struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID *uuid.UUID `gorm:"type:uuid"                                      json:"medical_record_id,omitempty"`
	PetID           *uuid.UUID `gorm:"type:uuid"                                      json:"pet_id,omitempty"`
	OwnerName       string     `gorm:"not null;default:''"                            json:"owner_name"`
	PetName         string     `gorm:"not null;default:''"                            json:"pet_name"`
	CheckupTypeID   uuid.UUID  `gorm:"type:uuid;not null"                             json:"checkup_type_id"`
	Date            time.Time  `gorm:"type:date;not null"                             json:"date"`
	NextDate        *time.Time `gorm:"type:date"                                      json:"next_date,omitempty"`
	DoctorID        *uuid.UUID `gorm:"type:uuid"                                      json:"doctor_id,omitempty"`
	Result          string     `gorm:"default:''"                                     json:"result"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	CheckupType   CheckupType    `gorm:"foreignKey:CheckupTypeID"   json:"checkup_type,omitempty"`
	Doctor        *Staff         `gorm:"foreignKey:DoctorID"        json:"doctor,omitempty"`
}

func (CheckupRecord) TableName() string { return "checkup_records" }
