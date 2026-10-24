package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const examinationDeceasedPetMessage = "死亡したペットには検査記録を登録できません"

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
	petLife sharedkernel.PetLifeStateReader,
	clinicID uint64,
	petID *uint64,
) error {
	if petID == nil {
		return nil
	}
	if petLife == nil {
		return apperrors.WrapInternalServerError("examination pet status verifier is required")
	}
	return sharedkernel.ValidatePetNotDeceased(
		ctx,
		petLife,
		clinicID,
		*petID,
		examinationDeceasedPetMessage,
	)
}
