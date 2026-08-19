package model

import "time"

// LabDevice is a clinic-owned analyzer. Name and exam binding live here, not in frontend constants.
// Write owner: medicalrecord (ADR-007). source_type is the serial protocol, not the display name.
type LabDevice struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID   uint64    `gorm:"not null;uniqueIndex:uq_lab_devices_clinic_source;uniqueIndex:uq_lab_devices_clinic_name" json:"clinic_id"`
	SourceType string    `gorm:"type:varchar(32);not null;uniqueIndex:uq_lab_devices_clinic_source" json:"source_type"`
	Name       string    `gorm:"type:varchar(100);not null;uniqueIndex:uq_lab_devices_clinic_name" json:"name"`
	ExamTypeID *uint64   `gorm:"column:exam_type_id" json:"exam_type_id,omitempty"`
	IsActive   bool      `gorm:"not null;default:true" json:"is_active"`
	SortOrder  int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (LabDevice) TableName() string { return "lab_devices" }
