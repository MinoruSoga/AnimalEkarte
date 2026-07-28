package auth

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const credentialAuditTimeout = 5 * time.Second

// CredentialMutationAudit is trusted request metadata for one credential
// mutation. AccountID and the persisted audit payload are derived by the use
// case so credentials, tokens, and account identity strings never enter it.
type CredentialMutationAudit struct {
	ClinicID      uint64
	ActorID       *uint64
	ActorType     string
	Action        string
	TargetStaffID uint64
	IPAddress     string
	UserAgent     string
}

// CredentialAuditTxLogger persists a credential audit entry in the caller's
// ambient transaction.
type CredentialAuditTxLogger interface {
	LogEntryTx(ctx context.Context, entry AuthAuditEntry) error
}

// CredentialAuditSubject is the account-to-staff/clinic identity needed for a
// password reset audit after the reset token identifies its account.
type CredentialAuditSubject struct {
	ClinicID uint64
	StaffID  uint64
}

// CredentialAuditSubjectResolver resolves the credential subject with
// transaction-scoped plain reads using the caller's ambient transaction. It
// must not acquire staff or assignment row locks because reset already holds
// the account lock and staff-admin password replacement uses the opposite
// staff/assignment-to-account lock order.
type CredentialAuditSubjectResolver interface {
	ResolveCredentialAuditSubject(
		ctx context.Context,
		accountID uint64,
	) (CredentialAuditSubject, error)
}

func validateCredentialMutationAudit(
	audit CredentialMutationAudit,
	expectedAction string,
) error {
	if audit.Action != expectedAction {
		return apperrors.WrapInternalServerError(
			"credential audit action is invalid",
		)
	}
	switch expectedAction {
	case model.AuditActionAuthPasswordChange:
		if audit.ClinicID == 0 ||
			audit.ActorID == nil ||
			*audit.ActorID == 0 ||
			audit.ActorType != model.AuditActorTypeStaff ||
			audit.TargetStaffID == 0 ||
			*audit.ActorID != audit.TargetStaffID {
			return apperrors.WrapInternalServerError(
				"password change audit metadata is invalid",
			)
		}
	case model.AuditActionAuthPasswordReset:
		if audit.ActorID != nil ||
			audit.ActorType != model.AuditActorTypeSystem ||
			audit.ClinicID != 0 ||
			audit.TargetStaffID != 0 {
			return apperrors.WrapInternalServerError(
				"password reset audit metadata is invalid",
			)
		}
	default:
		return apperrors.WrapInternalServerError(
			"credential audit action is unsupported",
		)
	}
	return nil
}

func credentialAuditEntry(
	audit CredentialMutationAudit,
	accountID uint64,
) AuthAuditEntry {
	clinicID := audit.ClinicID
	resourceID := accountID
	return AuthAuditEntry{
		ClinicID:   &clinicID,
		ActorID:    audit.ActorID,
		ActorType:  audit.ActorType,
		Action:     audit.Action,
		Resource:   model.AuditResourceAccount,
		ResourceID: &resourceID,
		NewValue: map[string]any{
			"staff_id": audit.TargetStaffID,
		},
		IPAddress: audit.IPAddress,
		UserAgent: audit.UserAgent,
	}
}

func auditClinicIDFromStaffAssignments(
	staffID uint64,
	assignments []model.StaffClinicAssignment,
) (uint64, bool) {
	var fallback uint64
	for i := range assignments {
		assignment := &assignments[i]
		if assignment.StaffID != staffID ||
			assignment.ClinicID == 0 ||
			assignment.DeletedAt.Valid {
			continue
		}
		if fallback == 0 {
			fallback = assignment.ClinicID
		}
		if assignment.IsMain {
			return assignment.ClinicID, true
		}
	}
	return fallback, fallback != 0
}

func (h *HTTPHandler) resolveSelfPasswordAuditClinic(
	ctx context.Context,
	staff *model.Staff,
	requestedClinicID uint64,
) (uint64, error) {
	if staff == nil || staff.ID == 0 || staff.AccountID == nil {
		return 0, apperrors.WrapInternalServerError(
			"password audit identity is invalid",
		)
	}
	if h.deps.StaffAssignments == nil {
		return 0, apperrors.WrapInternalServerError(
			"password audit clinic assignments are not configured",
		)
	}
	assignments, err := h.deps.StaffAssignments.FindAllByStaffID(ctx, staff.ID)
	if err != nil {
		return 0, apperrors.Wrap(err, "failed to resolve password audit clinic")
	}
	for i := range assignments {
		assignment := &assignments[i]
		if assignment.StaffID == staff.ID &&
			assignment.ClinicID == requestedClinicID &&
			!assignment.DeletedAt.Valid {
			return requestedClinicID, nil
		}
	}
	if err := h.validateSystemAdminPasswordAuditClinic(
		ctx,
		*staff.AccountID,
		requestedClinicID,
	); err == nil {
		return requestedClinicID, nil
	} else {
		return 0, err
	}
}

func (h *HTTPHandler) validateSystemAdminPasswordAuditClinic(
	ctx context.Context,
	accountID, requestedClinicID uint64,
) error {
	if requestedClinicID == 0 ||
		h.deps.Accounts == nil ||
		h.deps.Clinics == nil {
		return apperrors.WrapInternalServerError(
			"system administrator password audit dependencies are not configured",
		)
	}
	account, err := h.deps.Accounts.GetByID(ctx, accountID)
	if err != nil {
		return apperrors.Wrap(
			err,
			"failed to resolve system administrator password audit account",
		)
	}
	if account == nil ||
		account.ID != accountID ||
		!account.IsActive ||
		account.DeletedAt.Valid ||
		!account.IsSystemAdmin {
		return apperrors.WrapForbidden("no auditable clinic access is available")
	}
	clinics, err := h.deps.Clinics.ListClinics(ctx)
	if err != nil {
		return apperrors.Wrap(
			err,
			"failed to resolve system administrator password audit clinic",
		)
	}
	for i := range clinics {
		if clinics[i].ID == requestedClinicID && clinics[i].IsActive {
			return nil
		}
	}
	return apperrors.WrapForbidden("no auditable clinic access is available")
}
