package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// examinationAuditReasonAuthenticatedRequest is the bounded DEC-53 reason taxonomy for routine
// parent CRUD: the mutation occurred because an authenticated staff request passed the normal
// authorization path. It is not a caller-supplied business justification. Operations that require
// an explicit human reason (such as TASK-027 unconfirm) must define and validate that input separately.
const examinationAuditReasonAuthenticatedRequest = "authenticated_user_request"

func examinationAuditOptionalID(id *uint64) any {
	if id == nil {
		return nil
	}
	return *id
}

func examinationAuditJobID(exam *model.Examination) any {
	if exam.JobID == nil {
		return nil
	}
	return exam.JobID.String()
}

// examinationAuditValue builds the durable parent-row snapshot required by DEC-53.
// It intentionally excludes preloaded owner/pet/staff objects and records only the clinic-scoped
// foreign keys and examination values needed to reconstruct the mutation.
func examinationAuditValue(exam *model.Examination) map[string]any {
	if exam == nil {
		return nil
	}
	return map[string]any{
		"id":                       exam.ID,
		"medical_record_id":        examinationAuditOptionalID(exam.MedicalRecordID),
		"pet_id":                   examinationAuditOptionalID(exam.PetID),
		"exam_type_id":             exam.ExamTypeID,
		"doctor_id":                examinationAuditOptionalID(exam.DoctorID),
		"job_id":                   examinationAuditJobID(exam),
		"date":                     exam.Date,
		"result_summary":           exam.ResultSummary,
		"machine":                  exam.Machine,
		"status":                   exam.Status,
		"current_revision_version": examinationAuditOptionalID(exam.CurrentRevisionVersion),
	}
}

func (s *examinationService) validateParentMutationAudit(actorID *uint64) error {
	if actorID == nil || *actorID == 0 {
		return apperrors.WrapInvalidInput("authenticated staff actor is required for examination mutation")
	}
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("examination audit dependency is required")
	}
	return nil
}

func (s *examinationService) logParentMutationTx(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	action, operation string,
	before, after *model.Examination,
) error {
	return s.logParentMutationWithReasonTx(
		ctx,
		clinicID,
		actorID,
		action,
		operation,
		examinationAuditReasonAuthenticatedRequest,
		before,
		after,
	)
}

func (s *examinationService) logParentMutationWithReasonTx(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	action, operation, reason string,
	before, after *model.Examination,
) error {
	resourceID := uint64(0)
	if after != nil {
		resourceID = after.ID
	} else if before != nil {
		resourceID = before.ID
	}
	entry := &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     action,
		Resource:   model.AuditResourceExamination,
		ResourceID: &resourceID,
		OldValue:   examinationAuditValue(before),
		NewValue:   examinationAuditValue(after),
		Metadata: map[string]any{
			"operation_type": operation,
			"reason":         reason,
		},
	}
	if err := s.auditTx.LogEntryTx(ctx, entry); err != nil {
		slog.ErrorContext(ctx, "failed to audit parent examination mutation",
			"error", err,
			"clinic_id", clinicID,
			"examination_id", resourceID,
			"operation_type", operation,
		)
		return apperrors.Wrap(err, "failed to write examination mutation audit")
	}
	return nil
}
