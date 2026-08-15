package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const examinationDeceasedPetMessage = "死亡したペットには検査記録を登録できません"

type examinationPetByIDInClinicFinder interface {
	FindPetByIDInClinic(ctx context.Context, clinicID, petID uint64) (*model.Pet, error)
}

type examinationPetByIDAdapter struct {
	repo examinationPetByIDInClinicFinder
}

func (a examinationPetByIDAdapter) FindByID(ctx context.Context, clinicID, petID uint64) (*model.Pet, error) {
	return a.repo.FindPetByIDInClinic(ctx, clinicID, petID)
}

func effectiveExaminationPetID(explicitPetID *uint64, record *model.MedicalRecord) *uint64 {
	if explicitPetID != nil {
		return explicitPetID
	}
	if record != nil {
		return record.PetID
	}
	return nil
}

func validateExaminationPetNotDeceased(
	ctx context.Context,
	petRepo examinationPetByIDInClinicFinder,
	clinicID uint64,
	petID *uint64,
) error {
	if petID == nil {
		return nil
	}
	if petRepo == nil {
		return apperrors.WrapInternalServerError("examination pet status verifier is required")
	}
	return sharedkernel.ValidatePetNotDeceased(
		ctx,
		examinationPetByIDAdapter{repo: petRepo},
		clinicID,
		*petID,
		examinationDeceasedPetMessage,
	)
}
