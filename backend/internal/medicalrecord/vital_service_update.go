package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *vitalService) updateVitalInTx(
	txCtx context.Context,
	clinicID, medicalRecordID, vitalID uint64,
	input *UpdateVitalInput,
	fields map[string]any,
	existing *model.VitalRecord,
) (*model.VitalRecord, error) {
	// BE-refactor.md X-11: LockByIDForUpdate の行ロックで finalize と直列化し、確定と同時の
	// バイタル編集が確定済みカルテに混入する競合を防ぐ。
	parent, err := s.lockDraftParent(
		txCtx,
		clinicID,
		medicalRecordID,
		"failed to find medical record",
		"確定済みカルテのバイタルは編集できません",
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
	if err := s.repo.Update(txCtx, clinicID, vitalID, fields); err != nil {
		slog.ErrorContext(txCtx, "failed to update vital record", "error", err)
		return nil, apperrors.Wrap(err, "failed to update vital record")
	}
	result, err := s.repo.FindByID(txCtx, clinicID, vitalID)
	if err != nil {
		slog.ErrorContext(txCtx, "failed to get vital record after update", "error", err)
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
