package model

import (
	"time"
)

type MedicalRecordAddendum struct {
	ID              uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	MedicalRecordID uint64    `gorm:"column:medical_record_id;type:bigint;not null" json:"medical_record_id"`
	ClinicID        uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	AuthorUserID    uint64    `gorm:"column:author_user_id;type:bigint;not null" json:"author_user_id"`
	BeforeText      string    `gorm:"column:before_text;type:text;not null;default:''" json:"before_text"`
	AfterText       string    `gorm:"column:after_text;type:text;not null" json:"after_text"`
	Reason          string    `gorm:"column:reason;type:text;not null" json:"reason"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamptz;not null;default:NOW();autoCreateTime" json:"created_at"`
	AuthorStaff     *Staff    `gorm:"foreignKey:AuthorUserID;references:ID" json:"-" tygo:"-"`
}

func (MedicalRecordAddendum) TableName() string { return "medical_record_addenda" }
