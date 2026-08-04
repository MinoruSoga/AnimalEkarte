package model

import (
	"encoding/json"
	"time"
)

// CheckupPackageImportStatus is the durable provenance status for one clinic/namespace/version.
type CheckupPackageImportStatus string

const (
	CheckupPackageImportStatusApplied  CheckupPackageImportStatus = "applied"
	CheckupPackageImportStatusNoop     CheckupPackageImportStatus = "noop"
	CheckupPackageImportStatusConflict CheckupPackageImportStatus = "conflict"
	CheckupPackageImportStatusFailed   CheckupPackageImportStatus = "failed"
)

// CheckupPackageImportReceipt is the clinic-scoped import provenance row (internal sink).
// Operator receipt DTOs and application logs must not expose actor/clinic/digest/mapping.
type CheckupPackageImportReceipt struct {
	ID                   uint64                      `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID             uint64                      `gorm:"not null"                 json:"clinic_id"`
	Namespace            string                      `gorm:"not null"                 json:"namespace"`
	Version              string                      `gorm:"not null"                 json:"version"`
	ContentDigest        string                      `gorm:"not null"                 json:"content_digest"`
	Status               CheckupPackageImportStatus  `gorm:"type:text;not null"       json:"status"`
	ActorID              uint64                      `gorm:"not null"                 json:"actor_id"`
	TypesCreated         int                         `gorm:"not null;default:0"       json:"types_created"`
	FieldsCreated        int                         `gorm:"not null;default:0"       json:"fields_created"`
	ResourceMapping      json.RawMessage             `gorm:"type:jsonb;not null"      json:"resource_mapping"`
	ClinicalApprovalRef  string                      `gorm:"not null;default:''"      json:"clinical_approval_ref"`
	CreatedAt            time.Time                   `gorm:"not null;autoCreateTime"  json:"created_at"`
}

func (CheckupPackageImportReceipt) TableName() string {
	return "checkup_package_import_receipts"
}
