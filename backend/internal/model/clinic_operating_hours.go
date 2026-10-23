package model

import (
	"encoding/json"
	"time"
)

// ClinicOperatingHoursSchemaV1 is the stored schema marker for the common
// clinic operating-hours aggregate (contract §2.1).
const ClinicOperatingHoursSchemaV1 = "clinic-operating-hours/v1"

// ClinicOperatingHours maps clinic_operating_hours: one row per clinic holding
// the common weekly hours, closed dates, and the national-holiday policy.
// An absent row means unconfigured and is never complemented to all-open or
// all-closed. Strict JSON element validation (weekday 0..6 unique ascending,
// HH:MM intervals, non-overlapping intervals, unique ascending dates) belongs
// to the server strict decoder, not to this storage type. HTTP DTOs are owned
// by a different session; this model is not the wire shape.
type ClinicOperatingHours struct {
	ClinicID              uint64          `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"clinic_id"`
	SchemaVersion         string          `gorm:"column:schema_version;type:text;not null" json:"schema_version"`
	Version               int64           `gorm:"column:version;type:bigint;not null;default:1" json:"version"`
	WeeklyHours           json.RawMessage `gorm:"column:weekly_hours;type:jsonb;not null" json:"weekly_hours"`
	ClosedDates           json.RawMessage `gorm:"column:closed_dates;type:jsonb;not null" json:"closed_dates"`
	NationalHolidayClosed bool            `gorm:"column:national_holiday_closed;type:boolean;not null" json:"national_holiday_closed"`
	CreatedAt             time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:clock_timestamp();autoUpdateTime" json:"updated_at"`
}

// TableName returns the clinic operating hours table.
func (ClinicOperatingHours) TableName() string { return "clinic_operating_hours" }
