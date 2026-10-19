package model

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// LabImportJobItem is one decoded device line on a job. Write owner: medicalrecord.
type LabImportJobItem struct {
	ID              uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID        uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	JobID           uuid.UUID `gorm:"column:job_id;type:uuid;not null" json:"job_id"`
	DeviceItemCode  string    `gorm:"column:device_item_code;type:varchar(64);not null" json:"device_item_code"`
	ValueRaw        string    `gorm:"column:value_raw;type:varchar(64);not null;default:''" json:"value_raw"`
	Unit            string    `gorm:"column:unit;type:varchar(32);not null;default:''" json:"unit"`
	Flag            string    `gorm:"column:flag;type:varchar(32);not null;default:''" json:"flag"`
	ExamTypeFieldID *uint64   `gorm:"column:exam_type_field_id;type:bigint" json:"exam_type_field_id,omitempty"`
	NeedsReview     bool      `gorm:"column:needs_review;type:boolean;not null;default:false" json:"needs_review"`
	SortOrder       int       `gorm:"column:sort_order;type:integer;not null;default:0" json:"sort_order"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
}

func (LabImportJobItem) TableName() string { return "lab_import_job_items" }

// LabDeviceWait is the single active wait per clinic. Write owner: medicalrecord.
type LabDeviceWait struct {
	ID        uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID  uint64     `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	PetID     uint64     `gorm:"column:pet_id;type:bigint;not null" json:"pet_id"`
	StaffID   uint64     `gorm:"column:staff_id;type:bigint;not null" json:"staff_id"`
	ExpiresAt time.Time  `gorm:"column:expires_at;type:timestamptz;not null" json:"expires_at"`
	ClearedAt *time.Time `gorm:"column:cleared_at;type:timestamptz" json:"cleared_at,omitempty"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (LabDeviceWait) TableName() string { return "lab_device_waits" }

// LabDeviceStationSettings holds wait TTL and logical slots. Write owner: medicalrecord.
type LabDeviceStationSettings struct {
	ClinicID       uint64          `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"clinic_id"`
	WaitTTLSeconds int             `gorm:"column:wait_ttl_seconds;type:integer;not null;default:1800" json:"wait_ttl_seconds"`
	SlotsJSON      json.RawMessage `gorm:"column:slots_json;type:jsonb;not null;default:'[]'::jsonb" json:"slots_json"`
	CreatedAt      time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (LabDeviceStationSettings) TableName() string { return "lab_device_station_settings" }
