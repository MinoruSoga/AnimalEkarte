package model

import (
	"time"
)

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
	ID              uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID        uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	SourceType      string    `gorm:"column:source_type;type:varchar(32);not null" json:"source_type"`
	DeviceItemCode  string    `gorm:"column:device_item_code;type:varchar(64);not null" json:"device_item_code"`
	Unit            string    `gorm:"column:unit;type:varchar(32);not null;default:''" json:"unit"`
	ValueShape      string    `gorm:"column:value_shape;type:varchar(32);not null" json:"value_shape"`
	ExamTypeFieldID *uint64   `gorm:"column:exam_type_field_id;type:bigint" json:"exam_type_field_id,omitempty"`
	SortOrder       int       `gorm:"column:sort_order;type:integer;not null;default:0" json:"sort_order"`
	IsActive        bool      `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (LabDeviceItemMaster) TableName() string { return "lab_device_item_masters" }
