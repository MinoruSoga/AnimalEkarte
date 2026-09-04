package medicalrecord

import (
	"context"
	"fmt"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *labImportRevertService) revertLabImportInTx(
	txCtx context.Context,
	input RevertLabImportInput,
	reason, requestHash string,
) (*model.LabImportRevertResponse, error) {
	job, err := s.jobs.LockByIDForUpdate(txCtx, input.ClinicID, input.JobID)
	if err != nil {
		return nil, err
	}

	existing, findErr := s.reverts.LockByIdempotencyKey(txCtx, input.ClinicID, input.IdempotencyKey)
	if findErr == nil && existing != nil {
		if existing.JobID != input.JobID || existing.RequestHash != requestHash {
			return nil, apperrors.WrapConflict("idempotency key already used with a different payload")
		}
		ids, parseErr := parseRetractedExamIDs(existing.RetractedExamIDs)
		if parseErr != nil {
			return nil, apperrors.WrapInternalServerError("corrupt revert receipt")
		}
		return &model.LabImportRevertResponse{
			JobID:            existing.JobID,
			Status:           existing.ResultStatus,
			RetractedExamIDs: ids,
			IdempotentReplay: true,
		}, nil
	}
	if findErr != nil && !apperrors.IsNotFound(findErr) {
		return nil, findErr
	}

	if job.Status == model.LabImportJobStatusReverted {
		return nil, apperrors.WrapConflict("lab import job is already reverted")
	}
	if job.Status != model.LabImportJobStatusPersisted {
		return nil, apperrors.WrapConflict(
			fmt.Sprintf("lab import job status %s cannot be reverted; only persisted jobs are eligible", job.Status),
		)
	}

	hasTracking, err := s.events.HasEventType(txCtx, input.ClinicID, input.JobID, model.LabImportEventTypeUsageTrackingStarted)
	if err != nil {
		return nil, err
	}
	if !hasTracking {
		return nil, apperrors.WrapConflict("usage_unknown: lab import job has no usage_tracking_started marker; revert is refused")
	}

	exams, err := s.lockLinkedExamsByJob(txCtx, input.ClinicID, input.JobID)
	if err != nil {
		return nil, err
	}
	receipts, err := s.usage.LockByJobForUpdate(txCtx, input.ClinicID, input.JobID)
	if err != nil {
		return nil, err
	}
	if err := s.assertRevertSafe(txCtx, input.ClinicID, input.JobID, exams, receipts); err != nil {
		return nil, err
	}

	retractedIDs := make([]uint64, 0, len(exams))
	for _, exam := range exams {
		if err := s.retractExam(txCtx, input, reason, exam); err != nil {
			return nil, err
		}
		retractedIDs = append(retractedIDs, exam.ID)
	}

	now := time.Now()
	affected, err := s.jobs.CompareAndSetStatus(
		txCtx, input.ClinicID, input.JobID,
		model.LabImportJobStatusPersisted, model.LabImportJobStatusReverted, &now,
	)
	if err != nil {
		return nil, err
	}
	if affected != 1 {
		return nil, apperrors.WrapConflict("lab import job status changed concurrently; revert refused")
	}

	from := model.LabImportJobStatusPersisted
	to := model.LabImportJobStatusReverted
	if err := s.events.Create(txCtx, &model.LabImportEvent{
		ClinicID:   input.ClinicID,
		JobID:      input.JobID,
		EventType:  model.LabImportEventTypeRevertRequested,
		FromStatus: &from,
		ToStatus:   &to,
	}); err != nil {
		return nil, err
	}

	idsJSON := MustJSON(retractedIDs)
	if err := s.reverts.Create(txCtx, &model.LabImportRevertReceipt{
		ClinicID:         input.ClinicID,
		JobID:            input.JobID,
		IdempotencyKey:   input.IdempotencyKey,
		RequestHash:      requestHash,
		Reason:           reason,
		ActorID:          input.ActorID,
		ResultStatus:     string(model.LabImportJobStatusReverted),
		RetractedExamIDs: idsJSON,
	}); err != nil {
		return nil, err
	}
	if s.audit != nil {
		s.audit.LogRevertSucceeded(txCtx, input.ClinicID, input.ActorID, input.JobID, len(retractedIDs))
	}
	return &model.LabImportRevertResponse{
		JobID:            input.JobID,
		Status:           string(model.LabImportJobStatusReverted),
		RetractedExamIDs: retractedIDs,
		IdempotentReplay: false,
	}, nil
}
