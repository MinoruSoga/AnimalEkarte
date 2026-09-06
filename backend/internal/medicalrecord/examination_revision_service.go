package medicalrecord

import (
	"context"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const maxExaminationUnconfirmReasonRunes = 500

func examinationOptionalIDChanged(current, requested *uint64) bool {
	if requested == nil {
		return false
	}
	return current == nil || *current != *requested
}

func examinationItemsToUpsertInputs(items []model.ExamResult) []UpsertExamItemInput {
	inputs := make([]UpsertExamItemInput, 0, len(items))
	for i := range items {
		item := items[i]
		inputs = append(inputs, UpsertExamItemInput{
			ExamTypeFieldID: cloneOptionalUint64(item.ExamTypeItemID),
			Name:            item.Name,
			InspectionValue: item.InspectionValue,
			NormalValue:     item.NormalValue,
			Result:          item.Result,
			Unit:            item.Unit,
			ReferenceValue:  item.ReferenceValue,
			SortOrder:       item.SortOrder,
		})
	}
	return inputs
}

// ExaminationOfficialProjection keeps the immutable revision snapshot separate from
// the mutable parent row's current revision pointer.
type ExaminationOfficialProjection struct {
	model.Examination
	OfficialVersion uint64
}

// ExaminationOfficialReader is the separate Slice-A read capability for immutable
// official snapshots. The legacy general examination service interface remains unchanged.
type ExaminationOfficialReader interface {
	GetOfficialByID(ctx context.Context, clinicID, examinationID uint64) (*ExaminationOfficialProjection, error)
}

func (s *examinationService) confirmFirstRevisionTx(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	current, before *model.Examination,
	action, operation string,
) (*model.Examination, error) {
	if s.revisions == nil {
		return nil, apperrors.WrapInternalServerError("examination revision repository capability is required")
	}
	if actorID == nil || *actorID == 0 {
		return nil, apperrors.WrapInvalidInput("authenticated staff actor is required for examination confirmation")
	}
	if current == nil {
		return nil, apperrors.WrapInternalServerError("locked examination is required for confirmation")
	}
	if current.CurrentRevisionVersion != nil {
		return nil, apperrors.WrapConflict("revisioned examination requires the reconfirm workflow")
	}

	version, err := s.revisions.AppendOfficialRevision(
		ctx,
		clinicID,
		current.ID,
		*actorID,
		examinationInitialConfirmReason,
	)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to append official examination revision")
	}
	restored, err := s.repo.FindByID(ctx, clinicID, current.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to load confirmed examination snapshot")
	}
	after := *restored
	after.Status = model.ExaminationStatusConfirmed
	after.CurrentRevisionVersion = cloneUint64(version)
	if err := s.logParentMutationTx(ctx, clinicID, actorID, action, operation, before, &after); err != nil {
		return nil, err
	}
	confirmed, err := s.revisions.ConfirmWithRevisionCAS(
		ctx,
		clinicID,
		current.ID,
		current.Status,
		version,
	)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to advance examination revision pointer")
	}
	return confirmed, nil
}

func (s *examinationService) Unconfirm(
	ctx context.Context,
	clinicID, examinationID uint64,
	input UnconfirmExaminationInput,
) (*model.Examination, error) {
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("examination write transaction dependency is required")
	}
	if s.revisionWorkflow == nil {
		return nil, apperrors.WrapInternalServerError("examination revision workflow repository capability is required")
	}
	if err := s.validateParentMutationAudit(input.ActorID); err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return nil, apperrors.WrapInvalidInput("unconfirm reason is required")
	}
	if utf8.RuneCountInString(reason) > maxExaminationUnconfirmReasonRunes {
		return nil, apperrors.WrapInvalidInput("unconfirm reason must be 500 characters or fewer")
	}

	var unconfirmed *model.Examination
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		locked, err := s.repo.LockByIDForUpdate(txCtx, clinicID, examinationID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock examination")
		}
		if locked.Status != model.ExaminationStatusConfirmed {
			return apperrors.WrapConflict("only a confirmed examination can be unconfirmed")
		}
		if locked.CurrentRevisionVersion == nil {
			return apperrors.WrapConflict("legacy confirmed examination has no official revision")
		}
		if err := s.lockRevisionMedicalRecordDraftTx(txCtx, clinicID, locked); err != nil {
			return err
		}
		currentVersion := *locked.CurrentRevisionVersion
		nextVersion, err := s.revisionWorkflow.AppendWorkingRevisionFromOfficial(
			txCtx,
			clinicID,
			examinationID,
			currentVersion,
			*input.ActorID,
			reason,
		)
		if err != nil {
			return apperrors.Wrap(err, "failed to append working examination revision")
		}
		restored, err := s.repo.FindByID(txCtx, clinicID, examinationID)
		if err != nil {
			return apperrors.Wrap(err, "failed to load restored examination")
		}
		after := *restored
		after.Status = model.ExaminationStatusCompleted
		after.CurrentRevisionVersion = cloneUint64(nextVersion)
		if err := s.logParentMutationWithReasonTx(
			txCtx,
			clinicID,
			input.ActorID,
			model.AuditActionExaminationUnconfirm,
			"unconfirm",
			reason,
			locked,
			&after,
		); err != nil {
			return err
		}
		unconfirmed, err = s.revisionWorkflow.AdvanceRevisionCAS(
			txCtx,
			clinicID,
			examinationID,
			model.ExaminationStatusConfirmed,
			currentVersion,
			model.ExaminationStatusCompleted,
			nextVersion,
		)
		if err != nil {
			return err
		}
		// TASK-032: unconfirm is a manual mutation of an import-linked exam when job_id is set.
		return s.usage().RecordManualMutation(txCtx, clinicID, unconfirmed, input.ActorID)
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "examination unconfirmed",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("examination_id", examinationID),
	)
	return unconfirmed, nil
}

func (s *examinationService) appendWorkingRevisionTx(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	before, current *model.Examination,
	changeReason string,
) (*model.Examination, error) {
	if s.revisionWorkflow == nil {
		return nil, apperrors.WrapInternalServerError("examination revision workflow repository capability is required")
	}
	if actorID == nil || *actorID == 0 {
		return nil, apperrors.WrapInvalidInput("authenticated staff actor is required for examination revision")
	}
	if before == nil || before.CurrentRevisionVersion == nil || current == nil {
		return nil, apperrors.WrapConflict("working examination revision pointer is missing")
	}
	currentVersion := *before.CurrentRevisionVersion
	nextVersion, err := s.revisionWorkflow.AppendWorkingRevisionFromCurrent(
		ctx,
		clinicID,
		current.ID,
		currentVersion,
		*actorID,
		changeReason,
	)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to append working examination revision")
	}
	after := *current
	after.CurrentRevisionVersion = cloneUint64(nextVersion)
	if err := s.logParentMutationTx(
		ctx,
		clinicID,
		actorID,
		model.AuditActionExaminationUpdate,
		"working_update",
		before,
		&after,
	); err != nil {
		return nil, err
	}
	advanced, err := s.revisionWorkflow.AdvanceRevisionCAS(
		ctx,
		clinicID,
		current.ID,
		current.Status,
		currentVersion,
		current.Status,
		nextVersion,
	)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to advance working examination revision")
	}
	return advanced, nil
}

func (s *examinationService) reconfirmRevisionTx(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	current *model.Examination,
) (*model.Examination, error) {
	if s.revisionWorkflow == nil {
		return nil, apperrors.WrapInternalServerError("examination revision workflow repository capability is required")
	}
	if actorID == nil || *actorID == 0 {
		return nil, apperrors.WrapInvalidInput("authenticated staff actor is required for examination reconfirmation")
	}
	if current == nil || current.CurrentRevisionVersion == nil {
		return nil, apperrors.WrapConflict("working examination revision pointer is missing")
	}
	if err := s.lockRevisionMedicalRecordDraftTx(ctx, clinicID, current); err != nil {
		return nil, err
	}
	currentVersion := *current.CurrentRevisionVersion
	nextVersion, err := s.revisionWorkflow.AppendOfficialRevisionFromWorking(
		ctx,
		clinicID,
		current.ID,
		currentVersion,
		*actorID,
		examinationReconfirmReason,
	)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to append official examination revision")
	}
	restored, err := s.repo.FindByID(ctx, clinicID, current.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to load reconfirmed examination snapshot")
	}
	after := *restored
	after.Status = model.ExaminationStatusConfirmed
	after.CurrentRevisionVersion = cloneUint64(nextVersion)
	if err := s.logParentMutationTx(
		ctx,
		clinicID,
		actorID,
		model.AuditActionExaminationConfirm,
		"reconfirm",
		current,
		&after,
	); err != nil {
		return nil, err
	}
	confirmed, err := s.revisionWorkflow.AdvanceRevisionCAS(
		ctx,
		clinicID,
		current.ID,
		current.Status,
		currentVersion,
		model.ExaminationStatusConfirmed,
		nextVersion,
	)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to advance official examination revision")
	}
	return confirmed, nil
}

func (s *examinationService) lockRevisionMedicalRecordDraftTx(
	ctx context.Context,
	clinicID uint64,
	exam *model.Examination,
) error {
	if exam == nil || exam.MedicalRecordID == nil {
		return nil
	}
	return lockDraftMedicalRecord(
		ctx,
		s.medRec,
		clinicID,
		*exam.MedicalRecordID,
		"failed to find medical record",
		"確定済みカルテの検査確定状態は変更できません",
	)
}

// GetOfficialByID reads only the append-only official revision projection and fails
// closed when a legacy confirmed parent has no official revision.
func (s *examinationService) GetOfficialByID(
	ctx context.Context,
	clinicID, examinationID uint64,
) (*ExaminationOfficialProjection, error) {
	if s.revisions == nil {
		return nil, apperrors.WrapInternalServerError("examination revision repository capability is required")
	}
	exam, err := s.revisions.FindOfficialByID(ctx, clinicID, examinationID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to read official examination revision")
	}
	return exam, nil
}
