package medicalrecord

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// ClinicalRelationVerifier is medicalrecord's consumer-side view of the
// reservation-owned patient and doctor assignment checks.
type ClinicalRelationVerifier interface {
	sharedkernel.OwnerPetLinkVerifier
	AssertMedicalRecordDoctorInClinic(ctx context.Context, clinicID, doctorID uint64) error
}

func lockClinicalMedicalRecord(
	ctx context.Context,
	records medicalRecordLocker,
	clinicID, medicalRecordID uint64,
) (*model.MedicalRecord, error) {
	if records == nil {
		return nil, apperrors.WrapInternalServerError("medical record relation locker is required")
	}
	record, err := records.LockByIDForUpdate(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to verify medical record ownership")
	}
	if record == nil ||
		record.ID != medicalRecordID ||
		record.ClinicID != clinicID ||
		record.DeletedAt.Valid {
		return nil, apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", medicalRecordID))
	}
	return record, nil
}

func validateClinicalRelations(
	ctx context.Context,
	relations ClinicalRelationVerifier,
	clinicID uint64,
	record *model.MedicalRecord,
	petID, doctorID *uint64,
) error {
	if record == nil && petID == nil && doctorID == nil {
		return nil
	}
	if relations == nil {
		return apperrors.WrapInternalServerError("clinical relation verifier is required")
	}

	if record != nil {
		if record.OwnerID != nil {
			if err := relations.AssertOwnerInClinic(ctx, clinicID, *record.OwnerID); err != nil {
				return apperrors.Wrap(err, "failed to verify medical record owner ownership")
			}
		}
		if record.PetID != nil {
			// Clinic-scoped pet existence (and any ambient FOR SHARE) stays required.
			// Owner equality with the request pet is unreachable after pet ID match.
			if _, err := relations.FindPetOwnerInClinic(ctx, clinicID, *record.PetID); err != nil {
				return apperrors.Wrap(err, "failed to verify medical record pet ownership")
			}
		}
	}

	if petID != nil {
		if _, err := relations.FindPetOwnerInClinic(ctx, clinicID, *petID); err != nil {
			return apperrors.Wrap(err, "failed to verify patient ownership")
		}
	}

	if record != nil && petID != nil {
		if record.PetID == nil || *record.PetID != *petID {
			return apperrors.WrapNotFound("medical_record", "relation")
		}
	}

	if doctorID != nil {
		if err := relations.AssertMedicalRecordDoctorInClinic(ctx, clinicID, *doctorID); err != nil {
			return apperrors.Wrap(err, "failed to verify doctor ownership")
		}
	}
	return nil
}
