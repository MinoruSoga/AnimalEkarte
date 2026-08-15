package owner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/pet"
)

type captureOwnerRegistrationWriter struct {
	intent pet.OwnerRegistrationIntent
}

func (w *captureOwnerRegistrationWriter) CreateForOwnerRegistration(
	_ context.Context,
	intent pet.OwnerRegistrationIntent,
) ([]model.Pet, error) {
	w.intent = intent
	return []model.Pet{{Name: intent.Pets[0].Name}}, nil
}

func TestPetOwnerRegistrationAdapter_MapsOwnerIntentToPetWriteOwner(t *testing.T) {
	t.Parallel()

	writer := &captureOwnerRegistrationWriter{}
	registrar := pet.NewOwnerRegistrationAdapter(writer)
	var _ owner.PetRegistrar = registrar

	insuranceID := uint64(23)
	created, err := registrar.CreateForOwnerRegistration(context.Background(), owner.PetRegistrationIntent{
		ClinicID: 11,
		OwnerID:  17,
		Pets: []owner.PetRegistrationDraft{{
			AnimalSpeciesID: 19,
			Name:            "こむぎ",
			InsuranceID:     &insuranceID,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(11), writer.intent.ClinicID)
	require.Equal(t, uint64(17), writer.intent.OwnerID)
	require.Equal(t, uint64(19), writer.intent.Pets[0].AnimalSpeciesID)
	require.Equal(t, &insuranceID, writer.intent.Pets[0].InsuranceID)
	require.Equal(t, "こむぎ", created[0].Name)
}

func TestPetOwnerRegistrationAdapter_MissingWriterFailsClosed(t *testing.T) {
	t.Parallel()

	_, err := pet.NewOwnerRegistrationAdapter(nil).CreateForOwnerRegistration(
		context.Background(),
		owner.PetRegistrationIntent{ClinicID: 11, OwnerID: 17},
	)
	require.Error(t, err)
}
