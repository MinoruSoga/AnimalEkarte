package model

import (
	"gorm.io/gorm"
	"time"
)

type NextScheduleType string

const (
	NextScheduleType3Weeks NextScheduleType = "3weeks"
	NextScheduleType4Weeks NextScheduleType = "4weeks"
	NextScheduleType1Year  NextScheduleType = "1year"
	NextScheduleTypeOther  NextScheduleType = "other"
)

type Vaccination struct {
	ID               uint64            `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID         uint64            `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	MedicalRecordID  *uint64           `gorm:"column:medical_record_id;type:bigint" json:"medical_record_id,omitempty"`
	PetID            *uint64           `gorm:"column:pet_id;type:bigint" json:"pet_id,omitempty"`
	VaccineID        uint64            `gorm:"column:vaccine_id;type:bigint;not null" json:"vaccine_id"`
	Date             time.Time         `gorm:"column:date;type:date;not null" json:"date"`
	NextDate         *time.Time        `gorm:"column:next_date;type:date" json:"next_date,omitempty"`
	NextScheduleType *NextScheduleType `gorm:"column:next_schedule_type;type:next_schedule_type" json:"next_schedule_type,omitempty"`
	DoctorID         *uint64           `gorm:"column:doctor_id;type:bigint" json:"doctor_id,omitempty"`
	Supplemental     string            `gorm:"column:supplemental;type:text;not null;default:''" json:"supplemental"`
	Lot1             string            `gorm:"column:lot1;type:text;not null;default:''" json:"lot1"`
	Lot2             string            `gorm:"column:lot2;type:text;not null;default:''" json:"lot2"`
	Lot3             string            `gorm:"column:lot3;type:text;not null;default:''" json:"lot3"`
	Lot4             string            `gorm:"column:lot4;type:text;not null;default:''" json:"lot4"`
	Remarks          string            `gorm:"column:remarks;type:text;not null;default:''" json:"remarks"`
	CreatedAt        time.Time         `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time         `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt    `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Version          int               `gorm:"column:version;type:integer;not null;default:1" json:"version"`
	MedicalRecord    *MedicalRecord    `gorm:"foreignKey:MedicalRecordID,ClinicID;references:ID,ClinicID" json:"medical_record,omitempty"`
	Pet              *Pet              `gorm:"foreignKey:ClinicID,PetID;references:ClinicID,ID"           json:"pet,omitempty"`
	Vaccine          *Vaccine          `gorm:"foreignKey:VaccineID;references:ID"       json:"vaccine,omitempty"`
	Doctor           *Staff            `gorm:"foreignKey:DoctorID;references:ID"        json:"doctor,omitempty"`
}

func (Vaccination) TableName() string { return "vaccinations" }
