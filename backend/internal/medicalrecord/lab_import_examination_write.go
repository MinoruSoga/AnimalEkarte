package medicalrecord

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *labImportExaminationService) verifyLabExamPersistOwnership(ctx context.Context, input LabExamPersistInput) error {
	if input.PetID != nil {
		if _, err := s.petRepo.FindByID(ctx, input.ClinicID, *input.PetID); err != nil {
			slog.ErrorContext(ctx, "failed to verify pet ownership",
				"error", err,
				"pet_id", *input.PetID,
				"clinic_id", input.ClinicID,
			)
			return apperrors.Wrap(err, "failed to verify pet ownership")
		}
	}
	if input.MedicalRecordID == nil {
		return nil
	}
	record, err := s.medicalRecordRepo.FindByID(ctx, input.ClinicID, *input.MedicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to verify medical record ownership",
			"error", err,
			"medical_record_id", *input.MedicalRecordID,
			"clinic_id", input.ClinicID,
		)
		return apperrors.Wrap(err, "failed to verify medical record ownership")
	}
	if input.PetID != nil && (record == nil || record.PetID == nil || *record.PetID != *input.PetID) {
		return apperrors.WrapNotFound("medical_record", "relation")
	}
	return nil
}

func (s *labImportExaminationService) writeLabExamGraph(
	txCtx context.Context,
	input LabExamPersistInput,
	exam *model.Examination,
) (duplicate bool, err error) {
	examType, err := s.examTypeRepo.FindByID(txCtx, input.ClinicID, input.ExamTypeID)
	if err != nil {
		slog.ErrorContext(txCtx, "failed to verify exam type ownership",
			"error", err,
			"exam_type_id", input.ExamTypeID,
			"clinic_id", input.ClinicID,
		)
		return false, apperrors.Wrap(err, "failed to verify exam type ownership")
	}
	if examType == nil {
		return false, apperrors.WrapNotFound("exam_type", fmt.Sprintf("%d", input.ExamTypeID))
	}
	if err := requireOwnedExamTypeFields(examType, input.Items); err != nil {
		return false, err
	}
	if err := s.examRepo.Create(txCtx, exam); err != nil {
		if apperrors.IsAlreadyExists(err) {
			return true, nil
		}
		slog.ErrorContext(txCtx, "lab import exam create failed",
			"error", err,
			"clinic_id", input.ClinicID,
			"job_id", input.JobID.String(),
		)
		return false, apperrors.Wrap(err, "failed to create exam from lab import")
	}
	if len(input.Items) == 0 {
		return false, nil
	}
	items := buildExamResults(exam.ID, input.Items)
	if _, _, err := s.examRepo.ReplaceItemsByExamID(txCtx, input.ClinicID, exam.ID, items); err != nil {
		slog.ErrorContext(txCtx, "lab import exam items save failed",
			"error", err,
			"clinic_id", input.ClinicID,
			"exam_id", exam.ID,
			"job_id", input.JobID.String(),
		)
		return false, apperrors.Wrap(err, fmt.Sprintf("failed to save exam items for exam %d", exam.ID))
	}
	return false, nil
}
