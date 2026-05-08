package model

import "time"

type MedicalRecordAddendum struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MedicalRecordID uint64    `gorm:"not null"                 json:"medical_record_id"`
	ClinicID        uint64    `gorm:"not null"                 json:"clinic_id"`
	AuthorUserID    uint64    `gorm:"not null"                 json:"author_user_id"`
	BeforeText      string    `gorm:"not null;default:''"      json:"before_text"`
	AfterText       string    `gorm:"not null"                 json:"after_text"`
	Reason          string    `gorm:"not null"                 json:"reason"`
	CreatedAt       time.Time `gorm:"autoCreateTime"           json:"created_at"`
}

func (MedicalRecordAddendum) TableName() string { return "medical_record_addenda" }
