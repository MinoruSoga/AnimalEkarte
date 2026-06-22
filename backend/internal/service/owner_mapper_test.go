package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildOwnerPetModelsIncludesNullableLegacyReportFields(t *testing.T) {
	pets := buildOwnerPetModels([]CreatePetForOwnerInput{
		{
			Name:            "ポチ",
			AnimalSpeciesID: 1,
			BloodType:       "DEA1.1陽性",
			MicrochipNumber: "392140000123456",
		},
		{
			Name:            "未設定ペット",
			AnimalSpeciesID: 1,
		},
	})

	require.Len(t, pets, 2)
	require.NotNil(t, pets[0].BloodType)
	assert.Equal(t, "DEA1.1陽性", *pets[0].BloodType)
	require.NotNil(t, pets[0].MicrochipNumber)
	assert.Equal(t, "392140000123456", *pets[0].MicrochipNumber)
	assert.Nil(t, pets[1].BloodType)
	assert.Nil(t, pets[1].MicrochipNumber)
}

func TestBuildOwnerUpdateCanClearDMPreference(t *testing.T) {
	input := &UpdateOwnerInput{}
	input.DMPreference = new(*bool)

	fields := buildOwnerUpdate(input)

	assert.Contains(t, fields, colDMPreference)
	assert.Nil(t, fields[colDMPreference])
}
