package pet

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func uniqueAnimalSpeciesNameErr() error {
	return apperrors.FromGORM(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: apperrors.ConstraintAnimalSpeciesName,
	}, "animal_species", "")
}

func TestAnimalSpeciesService_Create_NameConflictMapsDomainCode(t *testing.T) {
	repo := &mockAnimalSpeciesRepository{
		createFn: func(_ context.Context, _ *model.AnimalSpecies) error {
			return uniqueAnimalSpeciesNameErr()
		},
	}
	svc := NewAnimalSpeciesService(repo, &mockPetRepository{})

	species, err := svc.Create(context.Background(), &CreateAnimalSpeciesInput{
		Name:     "V04動物種類",
		IsActive: true,
	}, AnimalSpeciesMutationMeta{})

	require.Error(t, err)
	assert.Nil(t, species)
	assert.True(t, apperrors.IsNameConflict(err, apperrors.CodeAnimalSpeciesNameConflict))
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "V04動物種類", appErr.Params["name"])
	assert.NotContains(t, appErr.Message, "animal_species '' already exists")
	assert.NotContains(t, err.Error(), "idx_animal_species_name")
}

func TestAnimalSpeciesService_Update_NameConflictMapsDomainCode(t *testing.T) {
	name := "重複種"
	repo := &mockAnimalSpeciesRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.AnimalSpecies, error) {
			return &model.AnimalSpecies{ID: id, Name: "旧"}, nil
		},
		updateFieldsFn: func(_ context.Context, _ uint64, _ UpdateAnimalSpeciesInput) (*model.AnimalSpecies, error) {
			return nil, uniqueAnimalSpeciesNameErr()
		},
	}
	svc := NewAnimalSpeciesService(repo, &mockPetRepository{})

	result, err := svc.Update(
		context.Background(),
		1,
		&UpdateAnimalSpeciesInput{Name: &name},
		AnimalSpeciesMutationMeta{},
	)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsNameConflict(err, apperrors.CodeAnimalSpeciesNameConflict))
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "重複種", appErr.Params["name"])
}

func TestAnimalSpeciesService_Create_GenericDBErrorNotElevated(t *testing.T) {
	repo := &mockAnimalSpeciesRepository{
		createFn: func(_ context.Context, _ *model.AnimalSpecies) error {
			// AlreadyExists without recoverable constraint — fail-closed, no domain code.
			return apperrors.WrapAlreadyExists("animal_species", "")
		},
	}
	svc := NewAnimalSpeciesService(repo, &mockPetRepository{})

	_, err := svc.Create(context.Background(), &CreateAnimalSpeciesInput{
		Name: "X", IsActive: true,
	}, AnimalSpeciesMutationMeta{})

	require.Error(t, err)
	assert.False(t, apperrors.IsNameConflict(err, apperrors.CodeAnimalSpeciesNameConflict))
	assert.True(t, apperrors.IsAlreadyExists(err))
}
