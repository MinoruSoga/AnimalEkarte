package owner

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type insuranceFinderFunc func(context.Context, uint64, uint64) (*model.Insurance, error)

func (f insuranceFinderFunc) FindByID(
	ctx context.Context,
	clinicID, insuranceID uint64,
) (*model.Insurance, error) {
	return f(ctx, clinicID, insuranceID)
}

func TestOwnerService_CreateWithPets_PreservesUnknownInsuranceLookupError(t *testing.T) {
	insuranceID := uint64(23)
	unexpected := errors.New("insurance database unavailable")
	service := NewService(
		&mockOwnerRepository{},
		insuranceFinderFunc(func(context.Context, uint64, uint64) (*model.Insurance, error) {
			return nil, unexpected
		}),
		nil,
		nil,
	)

	_, err := service.CreateWithPets(context.Background(), 11, &CreateOwnerInput{
		OwnerName: "飼主",
		Pets: []CreatePetForOwnerInput{{
			Name:            "こむぎ",
			AnimalSpeciesID: 19,
			InsuranceID:     &insuranceID,
		}},
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, unexpected)
	assert.False(t, apperrors.IsInvalidInput(err))
}

func TestOwnerService_CreateWithPets_MapsMissingInsuranceToInvalidInput(t *testing.T) {
	insuranceID := uint64(23)
	service := NewService(
		&mockOwnerRepository{},
		insuranceFinderFunc(func(context.Context, uint64, uint64) (*model.Insurance, error) {
			return nil, apperrors.WrapNotFound("insurance", "23")
		}),
		nil,
		nil,
	)

	_, err := service.CreateWithPets(context.Background(), 11, &CreateOwnerInput{
		OwnerName: "飼主",
		Pets: []CreatePetForOwnerInput{{
			Name:            "こむぎ",
			AnimalSpeciesID: 19,
			InsuranceID:     &insuranceID,
		}},
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}
