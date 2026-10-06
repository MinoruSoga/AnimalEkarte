package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Lab mapping draft statuses. Failed drafts cannot be confirmed.
const (
	LabMappingDraftStatusUnconfirmed = "unconfirmed"
	LabMappingDraftStatusConfirmed   = "confirmed"
	LabMappingDraftStatusFailed      = "failed"
)

// Lab mapping item sources. Dictionary rows have no LLM payload.
const (
	LabMappingItemSourceDictionary = "dictionary"
	LabMappingItemSourceLLM        = "llm"
	LabMappingItemSourceBlank      = "blank"
	LabMappingItemSourceComment    = "comment"
)

// LabDeviceMappingDraft is one device result kept for a human to confirm.
// RawPayload is the device frame. It is not a chart write and is not returned by the API.
type LabDeviceMappingDraft struct {
	ID                 uint64         `gorm:"primaryKey;autoIncrement;uniqueIndex:uq_lab_device_mapping_drafts_id_clinic,priority:2" json:"id"`
	ClinicID           uint64         `gorm:"not null;uniqueIndex:uq_lab_device_mapping_drafts_id_clinic,priority:1" json:"clinic_id"`
	SourceType         string         `gorm:"type:varchar(32);not null;default:''" json:"source_type"`
	Species            string         `gorm:"type:varchar(32);not null;default:''" json:"species"`
	Status             string         `gorm:"type:varchar(32);not null" json:"status"`
	RawPayload         []byte         `gorm:"type:bytea;not null" json:"-"`
	RawFingerprint     string         `gorm:"type:varchar(64);not null;default:''" json:"raw_fingerprint"`
	ErrorMessage       string         `gorm:"not null;default:''" json:"error_message"`
	ConfirmedAt        *time.Time     `json:"confirmed_at,omitempty"`
	ConfirmedByStaffID *uint64        `json:"confirmed_by_staff_id,omitempty"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the drafts table.
func (LabDeviceMappingDraft) TableName() string { return "lab_device_mapping_drafts" }

// LabDeviceMappingDraftItem is one code on a draft.
// ValueRaw and the LLM bodies stay in the database and out of the API.
type LabDeviceMappingDraftItem struct {
	ID                uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID          uint64          `gorm:"not null" json:"clinic_id"`
	DraftID           uint64          `gorm:"not null" json:"draft_id"`
	DeviceItemCode    string          `gorm:"type:varchar(64);not null;default:''" json:"device_item_code"`
	ValueRaw          string          `gorm:"not null;default:''" json:"-"`
	DeviceUnit        string          `gorm:"type:varchar(64);not null;default:''" json:"device_unit"`
	ChartUnit         string          `gorm:"type:varchar(32);not null;default:''" json:"chart_unit"`
	UnitMismatch      bool            `gorm:"not null;default:false" json:"unit_mismatch"`
	ExamTypeFieldID   *uint64         `json:"exam_type_field_id,omitempty"`
	ProposedFieldName string          `gorm:"type:varchar(255);not null;default:''" json:"proposed_field_name"`
	ItemSource        string          `gorm:"type:varchar(32);not null" json:"item_source"`
	LLMRequest        json.RawMessage `gorm:"type:jsonb" json:"-"`
	LLMResponse       json.RawMessage `gorm:"type:jsonb" json:"-"`
	LLMError          string          `gorm:"type:varchar(64);not null;default:''" json:"llm_error"`
	SortOrder         int             `gorm:"not null;default:0" json:"sort_order"`
	Confirmed         bool            `gorm:"not null;default:false" json:"confirmed"`
	CreatedAt         time.Time       `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt         gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName returns the draft items table.
func (LabDeviceMappingDraftItem) TableName() string { return "lab_device_mapping_draft_items" }

// LabDeviceMappingSpec is a confirmed device+species item map.
// Samples with an empty species do not read this table.
type LabDeviceMappingSpec struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID         uint64         `gorm:"not null;uniqueIndex:uq_lab_device_mapping_specs" json:"clinic_id"`
	SourceType       string         `gorm:"type:varchar(32);not null;uniqueIndex:uq_lab_device_mapping_specs" json:"source_type"`
	Species          string         `gorm:"type:varchar(32);not null;default:'';uniqueIndex:uq_lab_device_mapping_specs" json:"species"`
	DeviceItemCode   string         `gorm:"type:varchar(64);not null;uniqueIndex:uq_lab_device_mapping_specs" json:"device_item_code"`
	Unit             string         `gorm:"type:varchar(32);not null;default:''" json:"unit"`
	ValueShape       string         `gorm:"type:varchar(32);not null" json:"value_shape"`
	ExamTypeFieldID  uint64         `gorm:"not null" json:"exam_type_field_id"`
	ConfirmedDraftID uint64         `gorm:"not null" json:"confirmed_draft_id"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the species spec table.
func (LabDeviceMappingSpec) TableName() string { return "lab_device_mapping_specs" }

// LabDeviceMappingGeneration is the per-clinic switch for new drafts.
// A missing row means draft generation is allowed. Import does not read it.
type LabDeviceMappingGeneration struct {
	ClinicID  uint64    `gorm:"primaryKey" json:"clinic_id"`
	Stopped   bool      `gorm:"not null;default:false" json:"stopped"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName returns the generation switch table.
func (LabDeviceMappingGeneration) TableName() string { return "lab_device_mapping_generation" }
