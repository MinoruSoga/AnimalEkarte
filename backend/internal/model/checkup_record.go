package model

import (
	"gorm.io/gorm"
	"time"
)

type Checkup struct {
	ID              uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	MedicalRecordID uint64         `gorm:"column:medical_record_id;type:bigint;not null" json:"medical_record_id"`
	ClinicID        uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	PetID           *uint64        `gorm:"column:pet_id;type:bigint" json:"pet_id,omitempty"`
	CheckupTypeID   uint64         `gorm:"column:checkup_type_id;type:bigint;not null" json:"checkup_type_id"`
	Date            time.Time      `gorm:"column:date;type:date;not null" json:"date"`
	NextDate        *time.Time     `gorm:"column:next_date;type:date" json:"next_date,omitempty"`
	DoctorID        *uint64        `gorm:"column:doctor_id;type:bigint" json:"doctor_id,omitempty"`
	Result          string         `gorm:"column:result;type:text;not null;default:''" json:"result"`
	CreatedAt       time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Version         int64          `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	MedicalRecord   *MedicalRecord `gorm:"foreignKey:MedicalRecordID;references:ID" json:"medical_record,omitempty"`
	Pet             *Pet           `gorm:"foreignKey:PetID;references:ID"           json:"pet,omitempty"`
	CheckupType     *CheckupType   `gorm:"foreignKey:CheckupTypeID;references:ID"   json:"checkup_type,omitempty"`
	Doctor          *Staff         `gorm:"foreignKey:DoctorID;references:ID"        json:"doctor,omitempty"`
}

func (Checkup) TableName() string { return "checkups" }
