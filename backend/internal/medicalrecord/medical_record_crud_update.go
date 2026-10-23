package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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
	return updated, nil
}
