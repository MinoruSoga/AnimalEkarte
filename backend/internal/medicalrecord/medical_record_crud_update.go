package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func (s *medicalRecordService) updateMedicalRecordInTx(
	txCtx context.Context,
	clinicID, id uint64,
	existing *model.MedicalRecord,
	input UpdateMedicalRecordInput,
	fields map[string]any,
	finalOwnerID, finalPetID, finalDoctorID *uint64,
	needsLinkValidation, needsContextValidation, isBecomingFinalized bool,
) (*model.MedicalRecord, error) {
	if err := s.validateMedicalRecordUpdateLinks(txCtx, clinicID, finalOwnerID, finalPetID, needsLinkValidation, isBecomingFinalized); err != nil {
		return nil, err
	}
	if input.DoctorID != nil || isBecomingFinalized {
		if err := s.validateMedicalRecordDoctor(txCtx, clinicID, finalDoctorID); err != nil {
			return nil, err
		}
	}
	if isBecomingFinalized && existing.AppointmentID != nil {
		if s.reservationRepo == nil {
			return nil, apperrors.WrapInternalServerError("reservation lifecycle dependency is required to finalize an appointment-linked medical record")
		}
		if err := s.reservationRepo.PrepareForMedicalRecordFinalization(txCtx, clinicID, *existing.AppointmentID); err != nil {
			return nil, apperrors.Wrap(err, "failed to prepare appointment for medical record finalization")
		}
	}
	if existing.AppointmentID != nil && (needsContextValidation || isBecomingFinalized) {
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
	if isBecomingFinalized && s.auditTx == nil {
		return nil, apperrors.WrapInternalServerError("medical record finalize audit dependency is required")
	}
	updated, err := s.repo.Update(txCtx, clinicID, id, fields, input.Version)
	if err != nil {
		slog.ErrorContext(txCtx, "failed to update medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to update medical record")
	}
	if isBecomingFinalized {
		resourceID := id
		if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
			ClinicID:   &clinicID,
			ActorID:    input.ActorID,
			ActorType:  sharedkernel.AuditActorTypeFor(input.ActorID),
			Action:     "finalize",
			Resource:   "medical_record",
			ResourceID: &resourceID,
			NewValue:   extractMedicalRecordImportantFields(updated),
		}); err != nil {
			return nil, apperrors.Wrap(err, "failed to audit medical record finalize")
		}
	}
	return updated, nil
}
