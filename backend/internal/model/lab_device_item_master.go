package model

import "time"

// LabDeviceValueShape は電文値の形。マスタ列。legacy_name_candidate は持たない。
const (
	LabDeviceValueShapeNumeric    = "numeric"
	LabDeviceValueShapeInequality = "inequality"
	LabDeviceValueShapeQualAndNum = "qual_and_num"
	LabDeviceValueShapeDash       = "dash"
	LabDeviceValueShapeText       = "text"
)

// LabDeviceItemMaster maps a device item code to an optional exam_type_field.
// Write owner: medicalrecord (ADR-007). source_type is varchar, not the jobs enum.
type LabDeviceItemMaster struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID        uint64    `gorm:"not null;uniqueIndex:uq_lab_device_item_masters_clinic_source_code" json:"clinic_id"`
	SourceType      string    `gorm:"type:varchar(32);not null;uniqueIndex:uq_lab_device_item_masters_clinic_source_code" json:"source_type"`
	DeviceItemCode  string    `gorm:"type:varchar(64);not null;uniqueIndex:uq_lab_device_item_masters_clinic_source_code" json:"device_item_code"`
	DisplayName     string    `gorm:"type:varchar(100);not null" json:"display_name"`
	Unit            string    `gorm:"type:varchar(32);not null;default:''" json:"unit"`
	ValueShape      string    `gorm:"type:varchar(32);not null" json:"value_shape"`
	ExamTypeFieldID *uint64   `gorm:"column:exam_type_field_id" json:"exam_type_field_id,omitempty"`
	SortOrder       int       `gorm:"not null;default:0" json:"sort_order"`
	IsActive        bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (LabDeviceItemMaster) TableName() string { return "lab_device_item_masters" }
