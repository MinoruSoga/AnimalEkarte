package model

import (
	"encoding/json"
	"time"
)

// TrimmingRegisteredStartsSchemaV1 is the stored schema marker for the
// per-staff trimming registered start times (contract §4.1).
const TrimmingRegisteredStartsSchemaV1 = "trimming-registered-starts/v1"

// LineStaffStartPolicy maps line_staff_start_policies: the registered
// trimming start times for one (clinic, reservation_type, assignment).
// An absent row means unconfigured; an explicit empty weekday means no
// registered starts that day; a date override replaces the weekly set.
// DeletedAt is a tombstone read by the management GET — it is deliberately
// not gorm.DeletedAt so no soft-delete scope is applied to queries.
// Duty/capability/duration/capacity are not duplicated here. Strict JSON
// validation belongs to the server strict decoder.
type LineStaffStartPolicy struct {
	ClinicID          uint64          `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"clinic_id"`
	ReservationTypeID uint64          `gorm:"column:reservation_type_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"reservation_type_id"`
	AssignmentID      uint64          `gorm:"column:assignment_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"assignment_id"`
	StaffID           uint64          `gorm:"column:staff_id;type:bigint;not null" json:"staff_id"`
	SchemaVersion     string          `gorm:"column:schema_version;type:text;not null" json:"schema_version"`
	Version           int64           `gorm:"column:version;type:bigint;not null;default:1" json:"version"`
	WeeklyTimes       json.RawMessage `gorm:"column:weekly_times;type:jsonb;not null" json:"weekly_times"`
	DateOverrides     json.RawMessage `gorm:"column:date_overrides;type:jsonb;not null" json:"date_overrides"`
	DeletedAt         *time.Time      `gorm:"column:deleted_at;type:timestamptz" json:"deleted_at"`
	CreatedAt         time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:clock_timestamp();autoUpdateTime" json:"updated_at"`
}

// TableName returns the line staff start policies table.
func (LineStaffStartPolicy) TableName() string { return "line_staff_start_policies" }
