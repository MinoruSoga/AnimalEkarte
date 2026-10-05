package staff

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
)

// CredentialMutationAudit is trusted request metadata for an administrator
// replacing a staff account password. It intentionally contains no credential,
// token, email, or password-hash material.
type CredentialMutationAudit struct {
	ClinicID      uint64
	ActorStaffID  uint64
	TargetStaffID uint64
	IPAddress     string
	UserAgent     string
}

// CredentialAuditEntry is staff's consumer-owned view of a credential audit.
type CredentialAuditEntry struct {
	ClinicID   *uint64
	ActorID    *uint64
	ActorType  string
	Action     string
	Resource   string
	ResourceID *uint64
	NewValue   any
	IPAddress  string
	UserAgent  string
}

// CredentialAuditTxLogger persists a credential audit entry in the caller's
// ambient transaction.
type CredentialAuditTxLogger interface {
	LogEntryTx(ctx context.Context, entry CredentialAuditEntry) error
}

func validateStaffCredentialAudit(
	clinicID, targetStaffID uint64,
	audit *CredentialMutationAudit,
) error {
	if audit == nil ||
		audit.ClinicID == 0 ||
		audit.ClinicID != clinicID ||
		audit.ActorStaffID == 0 ||
		audit.TargetStaffID == 0 ||
		audit.TargetStaffID != targetStaffID {
		return apperrors.WrapInternalServerError(
			"staff credential audit metadata is invalid",
		)
	}
	return nil
}

func staffCredentialAuditEntry(
	audit CredentialMutationAudit,
	accountID uint64,
) CredentialAuditEntry {
	clinicID := audit.ClinicID
	actorID := audit.ActorStaffID
	resourceID := accountID
	return CredentialAuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     model.AuditActionAuthPasswordAdminReplace,
		Resource:   model.AuditResourceAccount,
		ResourceID: &resourceID,
		NewValue: map[string]any{
			"staff_id": audit.TargetStaffID,
		},
		IPAddress: audit.IPAddress,
		UserAgent: audit.UserAgent,
	}
}

func staffAccountAttachAuditEntry(
	audit CredentialMutationAudit,
	accountID uint64,
) CredentialAuditEntry {
	clinicID := audit.ClinicID
	actorID := audit.ActorStaffID
	resourceID := accountID
	return CredentialAuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     model.AuditActionStaffAccountAttach,
		Resource:   model.AuditResourceAccount,
		ResourceID: &resourceID,
		NewValue: map[string]any{
			"staff_id": audit.TargetStaffID,
		},
		IPAddress: audit.IPAddress,
		UserAgent: audit.UserAgent,
	}
}

// auditActionStaffAccountDeactivate is the durable audit action recorded when a
// staff mutation disables the linked login account in the same transaction.
// Declared here because internal/model is outside this change's owned paths.
const auditActionStaffAccountDeactivate = "staff.account.deactivate"

// writeAccountDeactivationAuditTx persists the linked-account deactivation in
// audit_logs on the caller's transaction (PR#551 review 4178809086), so the
// credential kill and its audit row commit or roll back together — an audit
// failure rolls back the staff/account mutation (fail-closed). Requests that
// carry an authenticated actor attribute the row to that staff; paths without
// actor plumbing (repository Delete) record a system actor.
func writeAccountDeactivationAuditTx(
	tx *gorm.DB,
	clinicID, staffID, accountID, actorStaffID uint64,
) error {
	entry := &model.AuditLog{
		ClinicID:   &clinicID,
		ActorType:  model.AuditActorTypeSystem,
		Action:     auditActionStaffAccountDeactivate,
		Resource:   model.AuditResourceAccount,
		ResourceID: &accountID,
		NewValue: mustJSON(map[string]any{
			"staff_id":  staffID,
			"is_active": false,
		}),
	}
	if actorStaffID != 0 {
		entry.ActorID = &actorStaffID
		entry.ActorType = model.AuditActorTypeStaff
	}
	if err := audit.ValidateLog(entry); err != nil {
		return err
	}
	if err := tx.Create(entry).Error; err != nil {
		return apperrors.FromGORM(err, "audit_log", "")
	}
	return nil
}
