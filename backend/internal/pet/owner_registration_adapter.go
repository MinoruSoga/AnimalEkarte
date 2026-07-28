package pet

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/owner"
)

// OwnerRegistrationAdapter implements owner's consumer-side registration port
// without moving pet persistence or numbering outside the pet write owner.
type OwnerRegistrationAdapter struct {
	writer OwnerRegistrationWriter
}

// NewOwnerRegistrationAdapter adapts the existing pet write owner to owner's
// cross-domain command.
func NewOwnerRegistrationAdapter(writer OwnerRegistrationWriter) *OwnerRegistrationAdapter {
	return &OwnerRegistrationAdapter{writer: writer}
}

func (a *OwnerRegistrationAdapter) CreateForOwnerRegistration(
	ctx context.Context,
	intent owner.PetRegistrationIntent,
) ([]model.Pet, error) {
	if a == nil || a.writer == nil {
		return nil, apperrors.WrapInternalServerError("pet owner-registration writer is not configured")
	}

	pets := make([]CreatePetDraft, 0, len(intent.Pets))
	for i := range intent.Pets {
		draft := intent.Pets[i]
		pets = append(pets, CreatePetDraft{
			AnimalSpeciesID: draft.AnimalSpeciesID,
			Name:            draft.Name,
			NameKana:        draft.NameKana,
			Breed:           draft.Breed,
			Color:           draft.Color,
			BloodType:       draft.BloodType,
			MicrochipNumber: draft.MicrochipNumber,
			Gender:          draft.Gender,
			Status:          draft.Status,
			BirthDate:       draft.BirthDate,
			Weight:          draft.Weight,
			NeuteredDate:    draft.NeuteredDate,
			AcquisitionType: draft.AcquisitionType,
			DangerLevel:     draft.DangerLevel,
			DangerReason:    draft.DangerReason,
			Food:            draft.Food,
			Environment:     draft.Environment,
			Phone:           draft.Phone,
			InsuranceID:     draft.InsuranceID,
			Remarks:         draft.Remarks,
		})
	}

	return a.writer.CreateForOwnerRegistration(ctx, OwnerRegistrationIntent{
		ClinicID: intent.ClinicID,
		OwnerID:  intent.OwnerID,
		Pets:     pets,
	})
}
