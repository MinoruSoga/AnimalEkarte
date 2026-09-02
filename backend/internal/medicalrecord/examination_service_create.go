package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *examinationService) createExaminationInTx(
	txCtx context.Context,
	clinicID uint64,
	input *CreateExaminationInput,
	exam *model.Examination,
	targetStatus model.ExaminationStatus,
) (*model.Examination, error) {
	record, err := lockOptionalDraftMedicalRecord(txCtx, s.medRec, clinicID, input.MedicalRecordID, "確定済みカルテに検査を追加できません")
	if err != nil {
		return nil, err
	}
	if err := validateClinicalRelations(txCtx, s.relations, clinicID, record, input.PetID, input.DoctorID); err != nil {
		return nil, err
	}
	petID := effectiveExaminationPetID(input.PetID, record)
	if err := validateExaminationPetNotDeceased(txCtx, s.petStatuses, clinicID, petID); err != nil {
		return nil, err
	}

	// クロステナント write 防止: 別 clinic の exam_type を紐付けると、その exam_type が持つ
	// 異常値判定の基準値/単位（exam_type_fields）が検査記録に混入する（#124 同型）。所有権を検証する。
	if input.ExamTypeID != 0 {
		if _, err := s.examTypeRepo.FindByID(txCtx, clinicID, input.ExamTypeID); err != nil {
			return nil, apperrors.Wrap(err, "failed to verify exam type ownership")
		}
	}

	if err := s.repo.Create(txCtx, exam); err != nil {
		slog.ErrorContext(txCtx, "failed to create examination", "error", err)
		return nil, apperrors.Wrap(err, "failed to create examination")
	}
	if input.Items != nil {
		if _, err := s.replaceItemsTx(txCtx, clinicID, exam, input.ActorID, *input.Items); err != nil {
			return nil, err
		}
	}
	if targetStatus == model.ExaminationStatusConfirmed {
		locked, err := s.repo.LockByIDForUpdate(txCtx, clinicID, exam.ID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to lock created examination for confirmation")
		}
		return s.confirmFirstRevisionTx(
			txCtx,
			clinicID,
			input.ActorID,
			locked,
			nil,
			model.AuditActionExaminationCreate,
			"create",
		)
	}
	// Create does not reload database-normalized columns (notably the PostgreSQL date value).
	// Audit and response data must describe the durable parent row, not the caller's timestamp.
	persisted, err := s.repo.FindByID(txCtx, clinicID, exam.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to reload created examination")
	}
	if err := s.logParentMutationTx(
		txCtx,
		clinicID,
		input.ActorID,
		model.AuditActionExaminationCreate,
		"create",
		nil,
		persisted,
	); err != nil {
		return nil, err
	}
	return persisted, nil
}
