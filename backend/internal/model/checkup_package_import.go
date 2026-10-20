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
	ID                  uint64                     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID            uint64                     `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Namespace           string                     `gorm:"column:namespace;type:text;not null" json:"namespace"`
	Version             string                     `gorm:"column:version;type:text;not null" json:"version"`
	ContentDigest       string                     `gorm:"column:content_digest;type:text;not null" json:"content_digest"`
	Status              CheckupPackageImportStatus `gorm:"column:status;type:text;not null" json:"status"`
	ActorID             uint64                     `gorm:"column:actor_id;type:bigint;not null" json:"actor_id"`
	TypesCreated        int                        `gorm:"column:types_created;type:integer;not null;default:0" json:"types_created"`
	FieldsCreated       int                        `gorm:"column:fields_created;type:integer;not null;default:0" json:"fields_created"`
	ResourceMapping     json.RawMessage            `gorm:"column:resource_mapping;type:jsonb;not null;default:'{}'::jsonb" json:"resource_mapping"`
	ClinicalApprovalRef string                     `gorm:"column:clinical_approval_ref;type:text;not null;default:''" json:"clinical_approval_ref"`
	CreatedAt           time.Time                  `gorm:"column:created_at;type:timestamptz;not null;default:CURRENT_TIMESTAMP;autoCreateTime" json:"created_at"`
}

func (CheckupPackageImportReceipt) TableName() string {
	return "checkup_package_import_receipts"
}
