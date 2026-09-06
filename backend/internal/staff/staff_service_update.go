package staff

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *staffService) applyStaffUpdateInTx(
	txCtx context.Context,
	clinicID, id uint64,
	input *UpdateStaffInput,
	hasProfileUpdate, hasPasswordUpdate bool,
	passwordHash string,
) (*model.Staff, error) {
	lockedStaff, err := s.lockStaffForMutation(txCtx, clinicID, id, input.IsSystemAdmin)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to lock staff for update")
	}
	if lockedStaff == nil || lockedStaff.ID != id {
		return nil, apperrors.WrapInternalServerError("staff lock returned an invalid record")
	}
	assignments, err := s.assignmentRepo.LockActiveByStaff(txCtx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to lock staff clinic assignments for update")
	}
	if err := authorizeGlobalStaffUpdate(
		id,
		clinicID,
		assignments,
		input.AuthorizedClinicIDs,
		input.IsSystemAdmin,
	); err != nil {
		return nil, err
	}
	writeClinicID, err := mutationClinicID(id, clinicID, assignments, input.IsSystemAdmin)
	if err != nil {
		return nil, err
	}
	if err := s.lockOccupationOwnership(txCtx, writeClinicID, input.OccupationID); err != nil {
		return nil, err
	}
	if hasPasswordUpdate && lockedStaff.AccountID == nil {
		return nil, apperrors.WrapInvalidInput("staff does not have an account")
	}
	if input.IsActive != nil && !*input.IsActive {
		if err := s.guardStaffDeactivation(txCtx, id, lockedStaff, input.ActorStaffID); err != nil {
			return nil, err
		}
	}
	if hasProfileUpdate {
		if err := s.repo.Update(txCtx, writeClinicID, id, *input); err != nil {
			return nil, apperrors.Wrap(err, "failed to update staff")
		}
	}
	if hasPasswordUpdate {
		if err := s.accountRepo.UpdatePasswordHash(
			txCtx,
			*lockedStaff.AccountID,
			passwordHash,
			time.Now(),
		); err != nil {
			return nil, apperrors.Wrap(err, "failed to update staff account password")
		}
		if err := s.accountRepo.DeletePasswordResetTokens(
			txCtx,
			*lockedStaff.AccountID,
		); err != nil {
			return nil, apperrors.Wrap(
				err,
				"failed to revoke staff password reset tokens",
			)
		}
		if auditErr := s.credentialAudit.LogEntryTx(
			txCtx,
			staffCredentialAuditEntry(
				*input.CredentialAudit,
				*lockedStaff.AccountID,
			),
		); auditErr != nil {
			return nil, apperrors.Wrap(
				auditErr,
				"failed to write staff credential audit",
			)
		}
	}
	if input.IsSystemAdmin {
		updated, err := s.repo.FindByID(txCtx, id)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get updated staff")
		}
		return updated, nil
	}
	updated, err := s.repo.FindByIDInClinic(txCtx, writeClinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get updated staff")
	}
	return updated, nil
}
