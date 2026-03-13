package model

import (
	"time"

	"gorm.io/gorm"
)

type Checkup struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID uint64         `gorm:"not null"                                       json:"medical_record_id"`
	PetID           *uint64        `                                                      json:"pet_id,omitempty"`
	CheckupTypeID   uint64         `gorm:"not null"                                       json:"checkup_type_id"`
	Date            time.Time      `gorm:"type:date;not null"                             json:"date"`
	DoctorID        *uint64        `                                                      json:"doctor_id,omitempty"`
	Result          string         `gorm:"default:''"                                     json:"result"`
	NextDate        *time.Time     `gorm:"type:date"                                      json:"next_date,omitempty"`
	DeletedAt       gorm.DeletedAt `                                                      json:"-" swaggerignore:"true"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	CheckupType   *CheckupType   `gorm:"foreignKey:CheckupTypeID"   json:"checkup_type,omitempty"`
	Doctor        *Staff         `gorm:"foreignKey:DoctorID"        json:"doctor,omitempty"`
}

func (Checkup) TableName() string { return "checkups" }
