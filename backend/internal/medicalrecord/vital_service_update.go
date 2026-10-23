package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *vitalService) updateVitalInTx(
	txCtx context.Context,
	clinicID, medicalRecordID, vitalID uint64,
	input *UpdateVitalInput,
	existing *model.VitalRecord,
) (*model.VitalRecord, error) {
	// BE-refactor.md X-11: 親カルテを行ロックし、同一性（clinic/ID）を検証する。
	// EMR-495: 旧 finalized 拒否は廃止した。
	parent, err := s.lockDraftParent(
		txCtx,
		clinicID,
		medicalRecordID,
		"failed to find medical record",
	)
	if err != nil {
		return nil, err
	}
	if err := validateVitalMedicalRecordRelation(parent, clinicID, medicalRecordID, existing.PetID); err != nil {
		return nil, err
	}
	petID := existing.PetID
	if err := validateClinicalRelations(
		txCtx,
		s.relations,
		clinicID,
		parent,
		&petID,
		nil,
	); err != nil {
		return nil, err
	}
	if existing.ClinicID != clinicID {
		return nil, apperrors.WrapNotFound("vital", "not found in medical record")
	}
	if err := validateClinicalStaffReference(
		txCtx,
		clinicID,
		input.StaffID,
		s.staffs,
		s.staffAssignments,
	); err != nil {
		return nil, err
	}
	if err := s.repo.Update(txCtx, clinicID, vitalID, *input); err != nil {
		return nil, apperrors.Wrap(err, "failed to update vital record")
	}
	result, err := s.repo.FindByID(txCtx, clinicID, vitalID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get vital record after update")
	}
	if err := validateUpdatedVitalRelation(
		result,
		clinicID,
		medicalRecordID,
		vitalID,
		existing.PetID,
	); err != nil {
		return nil, err
	}
	if err := validateClinicalStaffReference(
		txCtx,
		clinicID,
		result.StaffID,
		s.staffs,
		s.staffAssignments,
	); err != nil {
		return nil, err
	}
	if result.Staff != nil &&
		(result.StaffID == nil ||
			result.Staff.ID != *result.StaffID ||
			!result.Staff.IsActive) {
		return nil, apperrors.WrapNotFound("staff", "nested relation")
	}
	// BUG-015: vital update audit は ambient tx 参加の LogEntryTx で fail-closed。
	oldDiff, newDiff := diffVitalImportantFields(existing, result)
	if oldDiff != nil {
		if err := s.auditVitalTx(txCtx, clinicID, input.ActorID, "update", vitalID, medicalRecordID, oldDiff, newDiff); err != nil {
			return nil, err
		}
	}
	return result, nil
}
