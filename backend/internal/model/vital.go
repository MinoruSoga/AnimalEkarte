package model

import (
	"gorm.io/gorm"
	"time"
)

type BodyWeightUnit string

const (
	BodyWeightUnitKg BodyWeightUnit = "Kg"
	BodyWeightUnitG  BodyWeightUnit = "g"
)

type VitalRecord struct {
	ID              uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID        uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	PetID           uint64          `gorm:"column:pet_id;type:bigint;not null" json:"pet_id"`
	MedicalRecordID *uint64         `gorm:"column:medical_record_id;type:bigint" json:"medical_record_id"`
	DailyRecordID   *uint64         `gorm:"column:daily_record_id;type:bigint" json:"daily_record_id"`
	RecordedAt      time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:now()" json:"recorded_at"`
	StaffID         *uint64         `gorm:"column:staff_id;type:bigint" json:"staff_id"`
	Temperature     *ExactNumeric   `gorm:"column:temperature;type:numeric" json:"temperature"`
	HeartRate       *int            `gorm:"column:heart_rate;type:integer" json:"heart_rate"`
	RespirationRate *int            `gorm:"column:respiration_rate;type:integer" json:"respiration_rate"`
	Weight          *ExactNumeric   `gorm:"column:weight;type:numeric" json:"weight"`
	WeightUnit      *BodyWeightUnit `gorm:"column:weight_unit;type:body_weight_unit;default:'Kg'" json:"weight_unit"`
	Notes           string          `gorm:"column:notes;type:text;not null;default:''" json:"notes"`
	CreatedAt       time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:now()" json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Version         int             `gorm:"column:version;type:integer;not null;default:1" json:"version"`
	Pet             *Pet            `gorm:"foreignKey:PetID;references:ID"           json:"pet,omitempty"`
	MedicalRecord   *MedicalRecord  `gorm:"foreignKey:MedicalRecordID;references:ID" json:"-"`
	DailyRecord     *DailyRecord    `gorm:"foreignKey:DailyRecordID;references:ID"   json:"-"`
	Staff           *Staff          `gorm:"foreignKey:StaffID;references:ID"         json:"staff,omitempty"`
}

func (VitalRecord) TableName() string {
	return "vital_records"
}
