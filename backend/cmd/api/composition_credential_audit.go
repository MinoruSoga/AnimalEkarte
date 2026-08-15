package main

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/model"
)

type credentialAuditAccountReader interface {
	GetByID(ctx context.Context, accountID uint64) (*model.Account, error)
}

type credentialAuditStaffReader interface {
	FindByAccountID(
		ctx context.Context,
		accountID uint64,
	) (*model.Staff, error)
}

type authCredentialAuditTxAdapter struct {
	logger audit.TxLogger
}

func (a authCredentialAuditTxAdapter) LogEntryTx(
	ctx context.Context,
	entry auth.AuthAuditEntry,
) error {
	if a.logger == nil {
		return fmt.Errorf("auth credential audit logger is required")
	}
	return a.logger.LogEntryTx(ctx, &audit.Entry{
		ClinicID:   entry.ClinicID,
		ActorID:    entry.ActorID,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		OldValue:   entry.OldValue,
		NewValue:   entry.NewValue,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
	})
}

type authCredentialAuditSubjectResolver struct {
	accounts    credentialAuditAccountReader
	staff       credentialAuditStaffReader
	assignments auth.StaffClinicAssignmentReader
	clinics     auth.ClinicLister
}

func (r authCredentialAuditSubjectResolver) ResolveCredentialAuditSubject(
	ctx context.Context,
	accountID uint64,
) (auth.CredentialAuditSubject, error) {
	if accountID == 0 ||
		r.accounts == nil ||
		r.staff == nil ||
		r.assignments == nil ||
		r.clinics == nil {
		return auth.CredentialAuditSubject{}, apperrors.WrapInternalServerError(
			"credential audit subject dependencies are not configured",
		)
	}

	account, err := r.accounts.GetByID(ctx, accountID)
	if err != nil {
		return auth.CredentialAuditSubject{}, apperrors.Wrap(
			err,
			"failed to resolve credential audit account",
		)
	}
	if account == nil ||
		account.ID != accountID ||
		!account.IsActive ||
		account.DeletedAt.Valid {
		return auth.CredentialAuditSubject{}, apperrors.WrapForbidden(
			"credential audit account is no longer active",
		)
	}

	staffRecord, err := r.staff.FindByAccountID(ctx, accountID)
	if err != nil {
		return auth.CredentialAuditSubject{}, apperrors.Wrap(
			err,
			"failed to resolve credential audit staff",
		)
	}
	if staffRecord == nil ||
		staffRecord.ID == 0 ||
		staffRecord.AccountID == nil ||
		*staffRecord.AccountID != accountID ||
		!staffRecord.IsActive ||
		staffRecord.DeletedAt.Valid {
		return auth.CredentialAuditSubject{}, apperrors.WrapForbidden(
			"credential audit staff is no longer active",
		)
	}

	assignments, err := r.assignments.FindAllByStaffID(ctx, staffRecord.ID)
	if err != nil {
		return auth.CredentialAuditSubject{}, apperrors.Wrap(
			err,
			"failed to resolve credential audit clinic assignments",
		)
	}
	clinics, err := r.clinics.ListClinics(ctx)
	if err != nil {
		return auth.CredentialAuditSubject{}, apperrors.Wrap(
			err,
			"failed to resolve credential audit clinics",
		)
	}

	activeClinicIDs := make([]uint64, 0, len(clinics))
	activeClinics := make(map[uint64]struct{}, len(clinics))
	for i := range clinics {
		clinic := &clinics[i]
		if clinic.ID == 0 || !clinic.IsActive {
			continue
		}
		if _, duplicate := activeClinics[clinic.ID]; duplicate {
			continue
		}
		activeClinics[clinic.ID] = struct{}{}
		activeClinicIDs = append(activeClinicIDs, clinic.ID)
	}

	var assignedClinicID uint64
	for i := range assignments {
		assignment := &assignments[i]
		if assignment.DeletedAt.Valid {
			continue
		}
		if assignment.StaffID != staffRecord.ID ||
			assignment.ClinicID == 0 {
			return auth.CredentialAuditSubject{}, apperrors.WrapInternalServerError(
				"credential audit clinic assignment is invalid",
			)
		}
		if _, active := activeClinics[assignment.ClinicID]; !active {
			continue
		}
		if assignedClinicID == 0 {
			assignedClinicID = assignment.ClinicID
		}
		if assignment.IsMain {
			assignedClinicID = assignment.ClinicID
			break
		}
	}
	if assignedClinicID == 0 &&
		account.IsSystemAdmin &&
		len(activeClinicIDs) > 0 {
		assignedClinicID = activeClinicIDs[0]
	}
	if assignedClinicID == 0 {
		return auth.CredentialAuditSubject{}, apperrors.WrapForbidden(
			"no active credential audit clinic is available",
		)
	}

	return auth.CredentialAuditSubject{
		ClinicID: assignedClinicID,
		StaffID:  staffRecord.ID,
	}, nil
}
