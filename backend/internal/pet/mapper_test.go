package pet

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
		assert.Equal(t, "ぽち", pet.NameKana)
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

// TestBuildPetUpdate は buildPetUpdate のポインタ→map 変換ロジックを直接検証する。
func TestBuildPetUpdate(t *testing.T) {
	t.Run("returns empty map when all fields are nil", func(t *testing.T) {
		fields := buildPetUpdate(&UpdatePetInput{})
		assert.Empty(t, fields)
	})

	t.Run("includes all provided fields with correct column names", func(t *testing.T) {
		ownerID := uint64(5)
		speciesID := uint64(2)
		petNumber := "5-1"
		name := "ポチ"
		nameKana := "ポチ"
		gender := "male"
		birthDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		breed := "柴犬"
		color := "茶"
		bloodType := "A"
		microchip := "123456789012345"
		weight := 5.5
		neuteredDate := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
		acquisitionType := "purchase"
		dangerLevel := "low"
		food := "ドライフード"
		environment := "indoor"
		phone := "090-1234-5678"
		lastVisit := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		insuranceID := uint64(3)
		insuranceIDPtr := &insuranceID
		remarks := "備考"

		input := &UpdatePetInput{
			OwnerID:         &ownerID,
			AnimalSpeciesID: &speciesID,
			PetNumber:       &petNumber,
			Name:            &name,
			PetNameKana:     &nameKana,
			Gender:          &gender,
			BirthDate:       &birthDate,
			Breed:           &breed,
			Color:           &color,
			BloodType:       &bloodType,
			MicrochipNumber: &microchip,
			Weight:          &weight,
			NeuteredDate:    &neuteredDate,
			AcquisitionType: &acquisitionType,
			DangerLevel:     &dangerLevel,
			Food:            &food,
			Environment:     &environment,
			Phone:           &phone,
			LastVisit:       &lastVisit,
			InsuranceID:     &insuranceIDPtr,
			Remarks:         &remarks,
		}

		fields := buildPetUpdate(input)

		assert.Equal(t, ownerID, fields[colPetOwnerID])
		assert.Equal(t, speciesID, fields[colPetAnimalSpeciesID])
		assert.Equal(t, petNumber, fields["pet_number"])
		assert.Equal(t, name, fields[colPetName])
		assert.Equal(t, "ぽち", fields[colPetNameKana])
		assert.Equal(t, gender, fields[colPetGender])
		assert.Equal(t, birthDate, fields[colPetBirthDate])
		assert.Equal(t, breed, fields[colPetBreed])
		assert.Equal(t, color, fields["color"])
		assert.Equal(t, bloodType, fields[colPetBloodType])
		assert.Equal(t, microchip, fields[colPetMicrochipNumber])
		assert.Equal(t, weight, fields[colPetWeight])
		assert.Equal(t, neuteredDate, fields["neutered_date"])
		assert.Equal(t, acquisitionType, fields["acquisition_type"])
		assert.Equal(t, dangerLevel, fields["danger_level"])
		assert.Equal(t, food, fields["food"])
		assert.Equal(t, environment, fields[colPetEnvironment])
		assert.Equal(t, phone, fields["phone"])
		assert.Equal(t, lastVisit, fields["last_visit"])
		assert.Equal(t, insuranceIDPtr, fields[colPetInsuranceID])
		assert.Equal(t, remarks, fields[colPetRemarks])
		// BUG-415: status は generic update から意図的に除外されている（唯一の書込元は
		// Create と HandlePetDeath/HandlePetRevival に一本化済み）。UpdatePetInput に
		// Status フィールドが存在しないため、この map に status キーが混入する経路はない。
		_, hasStatus := fields["status"]
		assert.False(t, hasStatus, "buildPetUpdate は status を書き込んではならない(BUG-415)")
		assert.Len(t, fields, 21)
	})

	t.Run("clears insurance_id when InsuranceID points to a nil pointer", func(t *testing.T) {
		var nilInsurance *uint64
		input := &UpdatePetInput{InsuranceID: &nilInsurance}

		fields := buildPetUpdate(input)

		assert.Contains(t, fields, colPetInsuranceID)
		assert.Nil(t, fields[colPetInsuranceID])
	})
}
