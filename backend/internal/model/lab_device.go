package model

import (
	"time"
)

// LabDevice is a clinic-owned analyzer. Name and exam binding live here, not in frontend constants.
// Write owner: medicalrecord (ADR-007). source_type is the serial protocol, not the display name.
type LabDevice struct {
	ID         uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID   uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	SourceType string    `gorm:"column:source_type;type:varchar(32);not null" json:"source_type"`
	Name       string    `gorm:"column:name;type:varchar(100);not null" json:"name"`
	ExamTypeID *uint64   `gorm:"column:exam_type_id;type:bigint" json:"exam_type_id,omitempty"`
	IsActive   bool      `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	SortOrder  int       `gorm:"column:sort_order;type:integer;not null;default:0" json:"sort_order"`
	CreatedAt  time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (LabDevice) TableName() string { return "lab_devices" }
