package pet

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestValidatePetGender(t *testing.T) {
	assert.NoError(t, validatePetGender(""))
	assert.NoError(t, validatePetGender(string(model.PetGenderMale)))
	assert.Error(t, validatePetGender("invalid_gender"))
}

func TestValidatePetStatus(t *testing.T) {
	assert.NoError(t, validatePetStatus(""))
	assert.NoError(t, validatePetStatus(string(model.PetStatusAlive)))
	assert.Error(t, validatePetStatus("invalid_status"))
}

func TestValidatePetAcquisitionType(t *testing.T) {
	assert.NoError(t, validatePetAcquisitionType(""))
	assert.NoError(t, validatePetAcquisitionType(string(model.AcquisitionTypePurchase)))
	assert.Error(t, validatePetAcquisitionType("invalid_acquisition"))
}

func TestValidatePetDangerLevel(t *testing.T) {
	assert.NoError(t, validatePetDangerLevel(""))
	assert.NoError(t, validatePetDangerLevel(string(model.DangerLevelLow)))
	assert.Error(t, validatePetDangerLevel("invalid_danger"))
}

func TestValidateCreatePetInput(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		weight := 5.5
		input := &CreatePetInput{
			Name:            "pet",
			Weight:          &weight,
			Gender:          string(model.PetGenderMale),
			Status:          string(model.PetStatusAlive),
			AcquisitionType: string(model.AcquisitionTypePurchase),
			DangerLevel:     string(model.DangerLevelLow),
		}
		assert.NoError(t, validateCreatePetInput(input))
	})

	t.Run("invalid name", func(t *testing.T) {
		assert.Error(t, validateCreatePetInput(&CreatePetInput{Name: ""}))
	})

	t.Run("invalid weight", func(t *testing.T) {
		weight := -1.0
		assert.Error(t, validateCreatePetInput(&CreatePetInput{Name: "pet", Weight: &weight}))
	})

	t.Run("invalid gender", func(t *testing.T) {
		assert.Error(t, validateCreatePetInput(&CreatePetInput{Name: "pet", Gender: "invalid"}))
	})

	t.Run("invalid status", func(t *testing.T) {
		assert.Error(t, validateCreatePetInput(&CreatePetInput{Name: "pet", Status: "invalid"}))
	})

	t.Run("invalid acquisition", func(t *testing.T) {
		assert.Error(t, validateCreatePetInput(&CreatePetInput{Name: "pet", AcquisitionType: "invalid"}))
	})

	t.Run("invalid danger", func(t *testing.T) {
		assert.Error(t, validateCreatePetInput(&CreatePetInput{Name: "pet", DangerLevel: "invalid"}))
	})
}

func TestValidateUpdatePetInput(t *testing.T) {
	name := "pet"
	weight := 5.5
	gender := string(model.PetGenderMale)
	acquisitionType := string(model.AcquisitionTypePurchase)
	dangerLevel := string(model.DangerLevelLow)

	t.Run("valid", func(t *testing.T) {
		input := &UpdatePetInput{
			Name:            &name,
			Weight:          &weight,
			Gender:          &gender,
			AcquisitionType: &acquisitionType,
			DangerLevel:     &dangerLevel,
		}
		assert.NoError(t, validateUpdatePetInput(input))
	})

	t.Run("invalid name", func(t *testing.T) {
		value := ""
		assert.Error(t, validateUpdatePetInput(&UpdatePetInput{Name: &value}))
	})

	t.Run("invalid weight", func(t *testing.T) {
		value := -1.0
		assert.Error(t, validateUpdatePetInput(&UpdatePetInput{Weight: &value}))
	})

	t.Run("invalid gender", func(t *testing.T) {
		value := "bad"
		assert.Error(t, validateUpdatePetInput(&UpdatePetInput{Gender: &value}))
	})

	t.Run("invalid acquisition", func(t *testing.T) {
		value := "bad"
		assert.Error(t, validateUpdatePetInput(&UpdatePetInput{AcquisitionType: &value}))
	})

	t.Run("invalid danger", func(t *testing.T) {
		value := "bad"
		assert.Error(t, validateUpdatePetInput(&UpdatePetInput{DangerLevel: &value}))
	})
}
