package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *examinationService) updateExaminationInTx(
	txCtx context.Context,
	clinicID, id uint64,
	input UpdateExaminationInput,
	fields map[string]any,
	confirming bool,
) (*model.Examination, error) {
	locked, err := s.repo.LockByIDForUpdate(txCtx, clinicID, id)
	if err != nil {
		slog.ErrorContext(txCtx, "failed to lock examination", "error", err)
		return nil, apperrors.Wrap(err, "failed to lock examination")
	}
	if examinationFullyLocked(locked) {
		return nil, errExaminationFullyLocked()
	}
	if examinationResultsLocked(locked) && input.Items != nil {
		return nil, errExaminationResultsLocked(locked)
	}
	before := *locked
	revisioned := locked.CurrentRevisionVersion != nil
	petChanged := examinationOptionalIDChanged(locked.PetID, input.PetID)
	medicalRecordChanged := examinationOptionalIDChanged(locked.MedicalRecordID, input.MedicalRecordID)
	if revisioned && (petChanged || medicalRecordChanged) {
		return nil, apperrors.WrapConflict("revision history exists; examination patient relation cannot be changed")
	}
	if revisioned && s.revisionWorkflow == nil {
		return nil, apperrors.WrapInternalServerError("examination revision workflow repository capability is required")
	}
	if confirming && revisioned {
		if len(fields) != 0 || input.Items != nil {
			return nil, apperrors.WrapConflict("save working examination changes before reconfirming")
		}
		return s.reconfirmRevisionTx(txCtx, clinicID, input.ActorID, locked)
	}

	medicalRecordID, petID, doctorID := effectiveExaminationRelations(locked, input)
	record, err := s.lockExaminationUpdateMedicalRecords(
		txCtx,
		clinicID,
		locked.MedicalRecordID,
		medicalRecordID,
	)
	if err != nil {
		return nil, err
	}
	if err := validateClinicalRelations(txCtx, s.relations, clinicID, record, petID, doctorID); err != nil {
		return nil, err
	}
	if petChanged || medicalRecordChanged {
		targetPetID := effectiveExaminationPetID(petID, record)
		if err := validateExaminationPetNotDeceased(txCtx, s.petStatuses, clinicID, targetPetID); err != nil {
			return nil, err
		}
	}

	if err := validateOwnedMasterFK(txCtx, "exam type", clinicID, input.ExamTypeID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.examTypeRepo.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return nil, err
	}

	itemsToReplace := input.Items
	if petChanged && itemsToReplace == nil {
		existingItems, err := s.repo.FindAllItemsByExamID(txCtx, clinicID, id)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to load examination items for patient reassessment")
		}
		reassessedInputs := examinationItemsToUpsertInputs(existingItems)
		itemsToReplace = &reassessedInputs
	}

	exam := locked
	if len(fields) > 0 {
		updated, err := s.repo.Update(txCtx, clinicID, id, fields)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to update examination", "error", err)
			return nil, apperrors.Wrap(err, "failed to update examination")
		}
		exam = updated
	}
	if itemsToReplace != nil {
		if _, err := s.replaceItemsTx(txCtx, clinicID, exam, input.ActorID, *itemsToReplace); err != nil {
			return nil, err
		}
	}
	if confirming {
		exam, err = s.confirmFirstRevisionTx(
			txCtx,
			clinicID,
			input.ActorID,
			exam,
			&before,
			model.AuditActionExaminationConfirm,
			"confirm",
		)
		if err != nil {
			return nil, err
		}
		return exam, s.usage().RecordManualMutation(txCtx, clinicID, exam, input.ActorID)
	}
	if revisioned {
		exam, err = s.appendWorkingRevisionTx(txCtx, clinicID, input.ActorID, &before, exam, examinationWorkingUpdateReason)
		if err != nil {
			return nil, err
		}
		return exam, s.usage().RecordManualMutation(txCtx, clinicID, exam, input.ActorID)
	}
	if err := s.logParentMutationTx(
		txCtx,
		clinicID,
		input.ActorID,
		model.AuditActionExaminationUpdate,
		"update",
		&before,
		exam,
	); err != nil {
		return nil, err
	}
	return exam, s.usage().RecordManualMutation(txCtx, clinicID, exam, input.ActorID)
}
