package staff

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
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
