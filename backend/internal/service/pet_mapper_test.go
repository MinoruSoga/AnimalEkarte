package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestBuildPetModel(t *testing.T) {
	t.Run("maps required fields and leaves optional enum/nullable fields unset when zero-valued", func(t *testing.T) {
		input := &CreatePetInput{
			OwnerID:         5,
			AnimalSpeciesID: 2,
			Name:            "ポチ",
			PetNameKana:     "ポチ",
		}
		pet := buildPetModel(1, "P0001", input)

		assert.Equal(t, uint64(1), pet.ClinicID)
		assert.Equal(t, uint64(5), pet.OwnerID)
		assert.Equal(t, uint64(2), pet.AnimalSpeciesID)
		assert.Equal(t, "P0001", pet.PetNumber)
		assert.Equal(t, "ポチ", pet.Name)
		assert.Equal(t, model.PetGender(""), pet.Gender)
		assert.Equal(t, model.PetStatus(""), pet.Status)
		assert.Nil(t, pet.AcquisitionType)
		assert.Equal(t, model.DangerLevel(""), pet.DangerLevel)
		assert.Nil(t, pet.BloodType)
		assert.Nil(t, pet.MicrochipNumber)
	})

	t.Run("maps optional enum and nullable fields when provided", func(t *testing.T) {
		birthDate := time.Date(2020, 4, 20, 0, 0, 0, 0, time.UTC)
		weight := 4.5
		input := &CreatePetInput{
			OwnerID:         5,
			AnimalSpeciesID: 2,
			Name:            "ハチ",
			BirthDate:       &birthDate,
			Weight:          &weight,
			Gender:          "female",
			Status:          "alive",
			AcquisitionType: "rescued",
			DangerLevel:     "medium",
			BloodType:       "DEA1.1陽性",
			MicrochipNumber: "392140000123456",
		}
		pet := buildPetModel(1, "P0002", input)

		assert.Equal(t, model.PetGenderFemale, pet.Gender)
		assert.Equal(t, model.PetStatusAlive, pet.Status)
		require.NotNil(t, pet.AcquisitionType)
		assert.Equal(t, model.AcquisitionTypeRescued, *pet.AcquisitionType)
		assert.Equal(t, model.DangerLevelMedium, pet.DangerLevel)
		require.NotNil(t, pet.BloodType)
		assert.Equal(t, "DEA1.1陽性", *pet.BloodType)
		require.NotNil(t, pet.MicrochipNumber)
		assert.Equal(t, "392140000123456", *pet.MicrochipNumber)
		assert.Equal(t, &birthDate, pet.BirthDate)
		assert.Equal(t, &weight, pet.Weight)
	})
}
