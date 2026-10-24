package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *vitalService) createVitalInTx(
	txCtx context.Context,
	medicalRecordID uint64,
	input *CreateVitalInput,
	vital *model.VitalRecord,
) error {
	// HC-006 + BE-refactor.md X-11: 親カルテを行ロックし、同一性（clinic/ID）を検証する。
	// EMR-495: 旧 finalized 拒否は廃止した。
	parent, err := s.lockDraftParent(
		txCtx,
		input.ClinicID,
		medicalRecordID,
		"failed to find medical record",
	)
	if err != nil {
		return err
	}
	if err := validateVitalMedicalRecordRelation(parent, input.ClinicID, medicalRecordID, input.PetID); err != nil {
		return err
	}
	petID := input.PetID
	if err := validateClinicalRelations(
		txCtx,
		s.relations,
		input.ClinicID,
		parent,
		&petID,
		nil,
	); err != nil {
		return err
	}
	if err := validateClinicalStaffReference(
		txCtx,
		input.ClinicID,
		input.StaffID,
		s.staffs,
		s.staffAssignments,
	); err != nil {
		return err
	}
	if err := s.repo.Create(txCtx, vital); err != nil {
		return apperrors.Wrap(err, "failed to create vital record")
	}
	// BUG-015: vital create audit は ambient tx 参加の LogEntryTx で fail-closed。
	var actorID *uint64
	if input.StaffID != nil {
		actorID = input.StaffID
	}
	return s.auditVitalTx(txCtx, input.ClinicID, actorID, "create", vital.ID, medicalRecordID, nil, extractVitalImportantFields(vital))
}
