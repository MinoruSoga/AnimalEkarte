package model

import (
	"time"

	"gorm.io/gorm"
)

type BodyWeightUnit string

const (
	BodyWeightUnitKg BodyWeightUnit = "Kg"
	BodyWeightUnitG  BodyWeightUnit = "g"
)

type VitalRecord struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement"          json:"id"`
	ClinicID        uint64         `gorm:"not null"                          json:"clinic_id"`
	PetID           uint64         `gorm:"not null"                          json:"pet_id"`
	MedicalRecordID *uint64        `                                         json:"medical_record_id"`
	DailyRecordID   *uint64        `                                         json:"daily_record_id"`
	RecordedAt      time.Time      `gorm:"not null;default:now()"            json:"recorded_at"`
	StaffID         *uint64        `                                         json:"staff_id"`
	Temperature     *float64       `gorm:"type:numeric"                      json:"temperature"`
	HeartRate       *int           `gorm:"type:integer"                      json:"heart_rate"`
	RespirationRate *int           `gorm:"type:integer"                      json:"respiration_rate"`
	Weight          *float64       `gorm:"type:numeric"                      json:"weight"`
	WeightUnit      BodyWeightUnit `gorm:"type:body_weight_unit;default:'Kg'" json:"weight_unit"`
	Notes           string         `gorm:"not null;default:''"               json:"notes"`
	CreatedAt       time.Time      `gorm:"not null;default:now()"            json:"created_at"`
	UpdatedAt       time.Time      `gorm:"not null;default:now()"            json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index"                             json:"-"`

	Pet           *Pet           `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"-"`
	DailyRecord   *DailyRecord   `gorm:"foreignKey:DailyRecordID"   json:"-"`
	Staff         *Staff         `gorm:"foreignKey:StaffID"         json:"staff,omitempty"`
}

func (VitalRecord) TableName() string {
	return "vital_records"
}
