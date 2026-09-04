package model

import (
	"time"

	"github.com/google/uuid"
)

// LabImportJobItem is one decoded device line on a job. Write owner: medicalrecord.
type LabImportJobItem struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID        uint64    `gorm:"not null" json:"clinic_id"`
	JobID           uuid.UUID `gorm:"type:uuid;not null" json:"job_id"`
	DeviceItemCode  string    `gorm:"type:varchar(64);not null" json:"device_item_code"`
	ValueRaw        string    `gorm:"type:varchar(64);not null;default:''" json:"value_raw"`
	Unit            string    `gorm:"type:varchar(32);not null;default:''" json:"unit"`
	Flag            string    `gorm:"type:varchar(32);not null;default:''" json:"flag"`
	ExamTypeFieldID *uint64   `json:"exam_type_field_id,omitempty"`
	NeedsReview     bool      `gorm:"not null;default:false" json:"needs_review"`
	SortOrder       int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (LabImportJobItem) TableName() string { return "lab_import_job_items" }

// LabDeviceWait is the single active wait per clinic. Write owner: medicalrecord.
type LabDeviceWait struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID  uint64     `gorm:"not null" json:"clinic_id"`
	PetID     uint64     `gorm:"not null" json:"pet_id"`
	StaffID   uint64     `gorm:"not null" json:"staff_id"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	ClearedAt *time.Time `json:"cleared_at,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (LabDeviceWait) TableName() string { return "lab_device_waits" }

// LabDeviceStationSettings holds wait TTL and logical slots. Write owner: medicalrecord.
type LabDeviceStationSettings struct {
	ClinicID       uint64    `gorm:"primaryKey" json:"clinic_id"`
	WaitTTLSeconds int       `gorm:"not null;default:1800" json:"wait_ttl_seconds"`
	SlotsJSON      string    `gorm:"type:jsonb;not null;default:'[]'" json:"slots_json"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (LabDeviceStationSettings) TableName() string { return "lab_device_station_settings" }
