package pet

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestCreatePetDraftFromModel_DropsCallerOwnedIdentityAndAssociations(t *testing.T) {
	source := model.Pet{
		ID:              99,
		ClinicID:        88,
		OwnerID:         77,
		AnimalSpeciesID: 66,
		PetNumber:       "caller-controlled",
		Name:            "draft source",
		Owner:           &model.Owner{ID: 55},
		AnimalSpecies:   &model.AnimalSpecies{ID: 44},
		Insurance:       &model.Insurance{ID: 33},
	}

	draft := CreatePetDraftFromModel(source)
	persisted := draft.model(1, 2, "2-1")

	assert.Zero(t, persisted.ID)
	assert.Equal(t, uint64(1), persisted.ClinicID)
	assert.Equal(t, uint64(2), persisted.OwnerID)
	assert.Equal(t, "2-1", persisted.PetNumber)
	assert.Equal(t, uint64(66), persisted.AnimalSpeciesID)
	assert.Equal(t, "draft source", persisted.Name)
	assert.Nil(t, persisted.Owner)
	assert.Nil(t, persisted.AnimalSpecies)
	assert.Nil(t, persisted.Insurance)
}

func TestWriter_CreateFailsClosedWithoutDatabase(t *testing.T) {
	created, err := NewWriter(nil).Create(context.Background(), CreateIntent{})

	assert.Nil(t, created)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "INTERNAL", appErr.Code)
}

func TestOwnerRegistrationWriter_FailsClosedWithoutAmbientTransaction(t *testing.T) {
	created, err := NewOwnerRegistrationWriter().CreateForOwnerRegistration(
		context.Background(),
		OwnerRegistrationIntent{},
	)

	assert.Nil(t, created)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "INTERNAL", appErr.Code)
}
