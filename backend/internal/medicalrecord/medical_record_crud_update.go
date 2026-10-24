package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func (s *medicalRecordService) updateMedicalRecordInTx(
	txCtx context.Context,
	clinicID, id uint64,
	existing *model.MedicalRecord,
	input UpdateMedicalRecordInput,
	finalOwnerID, finalPetID, finalDoctorID *uint64,
	needsLinkValidation, needsContextValidation bool,
) (*model.MedicalRecord, error) {
	if needsLinkValidation {
		if err := s.validateMedicalRecordOwnerPetLinks(txCtx, clinicID, finalOwnerID, finalPetID); err != nil {
			return nil, err
		}
	}
	if input.DoctorID != nil {
		if err := s.validateMedicalRecordDoctor(txCtx, clinicID, finalDoctorID); err != nil {
			return nil, err
		}
	}
	if existing.AppointmentID != nil && needsContextValidation {
		if s.reservationRepo == nil {
			return nil, apperrors.WrapInternalServerError("reservation lifecycle dependency is required to validate an appointment-linked medical record")
		}
		if err := s.reservationRepo.BackfillForMedicalRecord(
			txCtx,
			clinicID,
			*existing.AppointmentID,
			finalOwnerID,
			finalPetID,
			finalDoctorID,
		); err != nil {
			return nil, apperrors.Wrap(err, "failed to lock and validate appointment for medical record update")
		}
	}
	updated, err := s.repo.Update(txCtx, clinicID, id, input, input.Version)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update medical record")
	}
	// EMR-495（監査 fail-closed）: 更新の監査は更新と同じ tx で書く。監査の書込に失敗したら更新ごと rollback し、
	// エラーを返す（成功した編集の編集者ログを best-effort にしない。lifecycle §1 / 495 契約 §6）。
	if err := s.auditMedicalRecordUpdateTx(txCtx, clinicID, id, input.ActorID, existing, updated); err != nil {
		return nil, err
	}
	return updated, nil
}

// auditMedicalRecordUpdateTx は通常更新の監査を ambient transaction へ参加させる（LogEntryTx → CreateTx）。
// 重要項目に差分がなければ書かない。差分があるのに監査依存が無い、または書込に失敗した場合は error を返し、
// 呼び出し元の tx（= 更新本体）を rollback させる。
func (s *medicalRecordService) auditMedicalRecordUpdateTx(
	ctx context.Context,
	clinicID, id uint64,
	actorID *uint64,
	existing, updated *model.MedicalRecord,
) error {
	oldDiff, newDiff := diffMedicalRecordImportantFields(existing, updated)
	if oldDiff == nil {
		return nil
	}
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("medical record audit dependency is required")
	}
	resourceID := id
	if err := s.auditTx.LogEntryTx(ctx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  sharedkernel.AuditActorTypeFor(actorID),
		Action:     "update",
		Resource:   "medical_record",
		ResourceID: &resourceID,
		OldValue:   oldDiff,
		NewValue:   newDiff,
	}); err != nil {
		return apperrors.Wrap(err, "failed to write medical record update audit")
	}
	return nil
}
