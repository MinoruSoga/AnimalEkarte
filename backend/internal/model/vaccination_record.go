package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NextScheduleType string

const (
	NextScheduleType3Weeks NextScheduleType = "3weeks"
	NextScheduleType4Weeks NextScheduleType = "4weeks"
	NextScheduleType1Year  NextScheduleType = "1year"
	NextScheduleTypeOther  NextScheduleType = "other"
)

type Vaccination struct {
	ID               uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID  uuid.UUID         `gorm:"type:uuid;not null"                             json:"medical_record_id"`
	PetID            *uuid.UUID        `gorm:"type:uuid"                                      json:"pet_id,omitempty"`
	VaccineID        uuid.UUID         `gorm:"type:uuid;not null"                             json:"vaccine_id"`
	Date             time.Time         `gorm:"type:date;not null"                             json:"date"`
	DoctorID         *uuid.UUID        `gorm:"type:uuid"                                      json:"doctor_id,omitempty"`
	NextDate         *time.Time        `gorm:"type:date"                                      json:"next_date,omitempty"`
	NextScheduleType *NextScheduleType `gorm:"type:next_schedule_type"                       json:"next_schedule_type,omitempty"`
	Supplemental     string            `gorm:"default:''"                                     json:"supplemental"`
	Lot1             string            `gorm:"default:''"                                     json:"lot1"`
	Lot2             string            `gorm:"default:''"                                     json:"lot2"`
	Lot3             string            `gorm:"default:''"                                     json:"lot3"`
	Lot4             string            `gorm:"default:''"                                     json:"lot4"`
	Remarks          string            `gorm:"default:''"                                     json:"remarks"`
	DeletedAt        gorm.DeletedAt    `                                                      json:"deleted_at"`
	CreatedAt        time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt        time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	Vaccine       *Vaccine       `gorm:"foreignKey:VaccineID"       json:"vaccine,omitempty"`
	Doctor        *Staff         `gorm:"foreignKey:DoctorID"        json:"doctor,omitempty"`
}

func (Vaccination) TableName() string { return "vaccinations" }
