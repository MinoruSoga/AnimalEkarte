package model

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
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
	ID                 uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID           uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	SourceType         string         `gorm:"column:source_type;type:varchar(32);not null;default:''" json:"source_type"`
	Species            string         `gorm:"column:species;type:varchar(32);not null;default:''" json:"species"`
	Status             string         `gorm:"column:status;type:varchar(32);not null" json:"status"`
	RawPayload         []byte         `gorm:"column:raw_payload;type:bytea;not null" json:"-"`
	RawFingerprint     string         `gorm:"column:raw_fingerprint;type:varchar(64);not null;default:''" json:"raw_fingerprint"`
	ErrorMessage       string         `gorm:"column:error_message;type:text;not null;default:''" json:"error_message"`
	ConfirmedAt        *time.Time     `gorm:"column:confirmed_at;type:timestamptz" json:"confirmed_at,omitempty"`
	ConfirmedByStaffID *uint64        `gorm:"column:confirmed_by_staff_id;type:bigint" json:"confirmed_by_staff_id,omitempty"`
	CreatedAt          time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

// TableName returns the drafts table.
func (LabDeviceMappingDraft) TableName() string { return "lab_device_mapping_drafts" }

// LabDeviceMappingDraftItem is one code on a draft.
// ValueRaw and the LLM bodies stay in the database and out of the API.
type LabDeviceMappingDraftItem struct {
	ID                uint64           `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID          uint64           `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	DraftID           uint64           `gorm:"column:draft_id;type:bigint;not null" json:"draft_id"`
	DeviceItemCode    string           `gorm:"column:device_item_code;type:varchar(64);not null;default:''" json:"device_item_code"`
	ValueRaw          string           `gorm:"column:value_raw;type:text;not null;default:''" json:"-"`
	DeviceUnit        string           `gorm:"column:device_unit;type:varchar(64);not null;default:''" json:"device_unit"`
	ChartUnit         string           `gorm:"column:chart_unit;type:varchar(32);not null;default:''" json:"chart_unit"`
	UnitMismatch      bool             `gorm:"column:unit_mismatch;type:boolean;not null;default:false" json:"unit_mismatch"`
	ExamTypeFieldID   *uint64          `gorm:"column:exam_type_field_id;type:bigint" json:"exam_type_field_id,omitempty"`
	ProposedFieldName string           `gorm:"column:proposed_field_name;type:varchar(255);not null;default:''" json:"proposed_field_name"`
	ItemSource        string           `gorm:"column:item_source;type:varchar(32);not null" json:"item_source"`
	LLMRequest        *json.RawMessage `gorm:"column:llm_request;type:jsonb" json:"-"`
	LLMResponse       *json.RawMessage `gorm:"column:llm_response;type:jsonb" json:"-"`
	LLMError          string           `gorm:"column:llm_error;type:varchar(64);not null;default:''" json:"llm_error"`
	SortOrder         int              `gorm:"column:sort_order;type:integer;not null;default:0" json:"sort_order"`
	Confirmed         bool             `gorm:"column:confirmed;type:boolean;not null;default:false" json:"confirmed"`
	CreatedAt         time.Time        `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	DeletedAt         gorm.DeletedAt   `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

// TableName returns the draft items table.
func (LabDeviceMappingDraftItem) TableName() string { return "lab_device_mapping_draft_items" }

// LabDeviceMappingSpec is a confirmed device+species item map.
// Samples with an empty species do not read this table.
type LabDeviceMappingSpec struct {
	ID               uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID         uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	SourceType       string         `gorm:"column:source_type;type:varchar(32);not null" json:"source_type"`
	Species          string         `gorm:"column:species;type:varchar(32);not null" json:"species"`
	DeviceItemCode   string         `gorm:"column:device_item_code;type:varchar(64);not null" json:"device_item_code"`
	Unit             string         `gorm:"column:unit;type:varchar(32);not null;default:''" json:"unit"`
	ValueShape       string         `gorm:"column:value_shape;type:varchar(32);not null" json:"value_shape"`
	ExamTypeFieldID  uint64         `gorm:"column:exam_type_field_id;type:bigint;not null" json:"exam_type_field_id"`
	ConfirmedDraftID uint64         `gorm:"column:confirmed_draft_id;type:bigint;not null" json:"confirmed_draft_id"`
	CreatedAt        time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
}

// TableName returns the species spec table.
func (LabDeviceMappingSpec) TableName() string { return "lab_device_mapping_specs" }

// LabDeviceMappingGeneration is the per-clinic switch for new drafts.
// A missing row means draft generation is allowed. Import does not read it.
type LabDeviceMappingGeneration struct {
	ClinicID  uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"clinic_id"`
	Stopped   bool      `gorm:"column:stopped;type:boolean;not null;default:false" json:"stopped"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

// TableName returns the generation switch table.
func (LabDeviceMappingGeneration) TableName() string { return "lab_device_mapping_generation" }
