package model

import (
	"encoding/json"
	"time"
)

// LabDeviceMappingHead is the current mapping revision counter for one
// (clinic, source_type, species_scope). species_scope "" is generic.
// Every settings writer increments Revision in the same transaction; readers
// stay inactive until all writers participate.
type LabDeviceMappingHead struct {
	ClinicID     uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"clinic_id"`
	SourceType   string    `gorm:"column:source_type;type:varchar(32);primaryKey;autoIncrement:false;not null" json:"source_type"`
	SpeciesScope string    `gorm:"column:species_scope;type:varchar(32);primaryKey;autoIncrement:false;not null" json:"species_scope"`
	Revision     int64     `gorm:"column:revision;type:bigint;not null" json:"revision"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:clock_timestamp();autoUpdateTime" json:"updated_at"`
}

// TableName returns the mapping heads table.
func (LabDeviceMappingHead) TableName() string { return "lab_device_mapping_heads" }

// LabDeviceMappingRevision is one immutable applied mapping version.
// SnapshotJSON is schema "lab-mapping-revision/v1" and SnapshotDigest is
// SHA-256(JCS(snapshot_json)); both stay in the database and out of the API.
// OperationID and RecordedByStaffID are NULL for legacy_baseline rows.
type LabDeviceMappingRevision struct {
	ID                       uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID                 uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	SourceType               string          `gorm:"column:source_type;type:varchar(32);not null" json:"source_type"`
	SpeciesScope             string          `gorm:"column:species_scope;type:varchar(32);not null" json:"species_scope"`
	Revision                 int64           `gorm:"column:revision;type:bigint;not null" json:"revision"`
	SnapshotJSON             json.RawMessage `gorm:"column:snapshot_json;type:jsonb;not null" json:"-"`
	SnapshotDigest           []byte          `gorm:"column:snapshot_digest;type:bytea;not null" json:"-"`
	FieldDefinitionSnapshots json.RawMessage `gorm:"column:field_definition_snapshots;type:jsonb;not null" json:"-"`
	ConfirmedDraftID         *uint64         `gorm:"column:confirmed_draft_id;type:bigint" json:"confirmed_draft_id,omitempty"`
	OperationID              *uint64         `gorm:"column:operation_id;type:bigint" json:"operation_id,omitempty"`
	RecordedByStaffID        *uint64         `gorm:"column:recorded_by_staff_id;type:bigint" json:"recorded_by_staff_id,omitempty"`
	RecordedAt               time.Time       `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime" json:"recorded_at"`
}

// TableName returns the mapping revisions table.
func (LabDeviceMappingRevision) TableName() string { return "lab_device_mapping_revisions" }
