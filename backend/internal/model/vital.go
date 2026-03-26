package model

import "time"

type BodyWeightUnit string

const (
	BodyWeightUnitKg BodyWeightUnit = "Kg"
	BodyWeightUnitG  BodyWeightUnit = "g"
)

type VitalRecord struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement"`
	PetID           uint64 `gorm:"not null"`
	MedicalRecordID *uint64
	DailyRecordID   *uint64
	RecordedAt      time.Time `gorm:"not null;default:now()"`
	StaffID         *uint64
	Temperature     *float64
	HeartRate       *int
	RespirationRate *int
	Weight          *float64
	WeightUnit      BodyWeightUnit `gorm:"type:body_weight_unit;default:'Kg'"`
	Notes           string         `gorm:"not null;default:''"`
	CreatedAt       time.Time      `gorm:"not null;default:now()"`
	UpdatedAt       time.Time      `gorm:"not null;default:now()"`

	Pet           *Pet           `gorm:"foreignKey:PetID"`
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID"`
	DailyRecord   *DailyRecord   `gorm:"foreignKey:DailyRecordID"`
	Staff         *Staff         `gorm:"foreignKey:StaffID"`
}

func (VitalRecord) TableName() string {
	return "vital_records"
}
