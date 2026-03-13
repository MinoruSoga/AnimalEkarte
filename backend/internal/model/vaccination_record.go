package model

import (
	"time"

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
	ID               uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID  uint64            `gorm:"not null"                                       json:"medical_record_id"`
	PetID            *uint64           `                                                      json:"pet_id,omitempty"`
	VaccineID        uint64            `gorm:"not null"                                       json:"vaccine_id"`
	Date             time.Time         `gorm:"type:date;not null"                             json:"date"`
	DoctorID         *uint64           `                                                      json:"doctor_id,omitempty"`
	NextDate         *time.Time        `gorm:"type:date"                                      json:"next_date,omitempty"`
	NextScheduleType *NextScheduleType `gorm:"type:next_schedule_type"                       json:"next_schedule_type,omitempty"`
	Supplemental     string            `gorm:"default:''"                                     json:"supplemental"`
	Lot1             string            `gorm:"default:''"                                     json:"lot1"`
	Lot2             string            `gorm:"default:''"                                     json:"lot2"`
	Lot3             string            `gorm:"default:''"                                     json:"lot3"`
	Lot4             string            `gorm:"default:''"                                     json:"lot4"`
	Remarks          string            `gorm:"default:''"                                     json:"remarks"`
	DeletedAt        gorm.DeletedAt    `                                                      json:"deleted_at" swaggerignore:"true"`
	CreatedAt        time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt        time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	Vaccine       *Vaccine       `gorm:"foreignKey:VaccineID"       json:"vaccine,omitempty"`
	Doctor        *Staff         `gorm:"foreignKey:DoctorID"        json:"doctor,omitempty"`
}

func (Vaccination) TableName() string { return "vaccinations" }
