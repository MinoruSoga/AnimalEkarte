package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *vitalService) createVitalInTx(
	txCtx context.Context,
	medicalRecordID uint64,
	input *CreateVitalInput,
	vital *model.VitalRecord,
) error {
	// HC-006 + BE-refactor.md X-11: 親カルテが確定済みの場合は作成拒否。LockByIDForUpdate の
	// 行ロックで finalize と直列化し、確定と同時のバイタル追加が確定済みカルテに混入する競合を防ぐ。
	parent, err := s.lockDraftParent(
		txCtx,
		input.ClinicID,
		medicalRecordID,
		"failed to find medical record",
		"確定済みカルテにバイタルを追加できません",
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
		slog.ErrorContext(txCtx, "failed to create vital record", "error", err)
		return apperrors.Wrap(err, "failed to create vital record")
	}
	// BUG-015: vital create audit は ambient tx 参加の LogEntryTx で fail-closed。
	var actorID *uint64
	if input.StaffID != nil {
		actorID = input.StaffID
	}
	return s.auditVitalTx(txCtx, input.ClinicID, actorID, "create", vital.ID, medicalRecordID, nil, extractVitalImportantFields(vital))
}
