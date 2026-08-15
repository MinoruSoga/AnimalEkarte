package owner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
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

func TestBuildOwnerPetModelsMapsOptionalEnumFields(t *testing.T) {
	pets := buildOwnerPetModels([]CreatePetForOwnerInput{
		{
			Name:            "ハチ",
			AnimalSpeciesID: 1,
			Gender:          "male",
			Status:          "alive",
			AcquisitionType: "purchased",
			DangerLevel:     "low",
		},
		{
			Name:            "未設定",
			AnimalSpeciesID: 1,
		},
	})

	require.Len(t, pets, 2)
	assert.Equal(t, model.PetGenderMale, pets[0].Gender)
	assert.Equal(t, model.PetStatusAlive, pets[0].Status)
	require.NotNil(t, pets[0].AcquisitionType)
	assert.Equal(t, model.AcquisitionTypePurchase, *pets[0].AcquisitionType)
	assert.Equal(t, model.DangerLevelLow, pets[0].DangerLevel)

	// zero-value input leaves the enum fields at their Go zero value (unset)
	assert.Equal(t, model.PetGender(""), pets[1].Gender)
	assert.Equal(t, model.PetStatus(""), pets[1].Status)
	assert.Nil(t, pets[1].AcquisitionType)
	assert.Equal(t, model.DangerLevel(""), pets[1].DangerLevel)
}

func TestBuildOwnerPetModelsEmptyInput(t *testing.T) {
	pets := buildOwnerPetModels(nil)
	assert.Empty(t, pets)
}

func TestBuildOwnerModel(t *testing.T) {
	t.Run("defaults membership type to non_member when empty", func(t *testing.T) {
		owner := buildOwnerModel(1, &CreateOwnerInput{OwnerName: "山田太郎"})
		assert.Equal(t, model.MembershipTypeNonMember, owner.MembershipType)
		assert.Equal(t, uint64(1), owner.ClinicID)
		assert.Equal(t, "山田太郎", owner.Name)
	})

	t.Run("keeps explicit membership type", func(t *testing.T) {
		owner := buildOwnerModel(1, &CreateOwnerInput{OwnerName: "鈴木花子", MembershipType: model.MembershipTypeMember})
		assert.Equal(t, model.MembershipTypeMember, owner.MembershipType)
	})
}

func TestBuildOwnerUpdateCanClearDMPreference(t *testing.T) {
	input := &UpdateOwnerInput{}
	input.DMPreference = new(*bool)

	fields := buildOwnerUpdate(input)

	assert.Contains(t, fields, colDMPreference)
	assert.Nil(t, fields[colDMPreference])
}
