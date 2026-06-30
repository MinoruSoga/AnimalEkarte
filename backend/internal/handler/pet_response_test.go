package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToPetListResponseIncludesOwnerReportDetailFields(t *testing.T) {
	birthDate := time.Date(2015, 4, 14, 0, 0, 0, 0, time.UTC)
	neuteredDate := time.Date(2016, 5, 20, 0, 0, 0, 0, time.UTC)
	lastVisit := time.Date(2024, 8, 25, 0, 0, 0, 0, time.UTC)
	weight := 7.35
	acquisitionType := model.AcquisitionTypePurchase
	insuranceID := uint64(9)
	bloodType := "DEA1.1陽性"
	microchipNumber := "392140000123456"

	resp := toPetListResponse(&model.Pet{
		ID:              7,
		OwnerID:         42,
		AnimalSpeciesID: 3,
		PetNumber:       "42-1",
		Name:            "ポチ",
		NameKana:        "ぽち",
		Gender:          model.PetGenderMale,
		Status:          model.PetStatusAlive,
		BirthDate:       &birthDate,
		Breed:           "柴犬",
		Color:           "赤",
		BloodType:       &bloodType,
		MicrochipNumber: &microchipNumber,
		Weight:          &weight,
		NeuteredDate:    &neuteredDate,
		AcquisitionType: &acquisitionType,
		DangerLevel:     model.DangerLevelMedium,
		Food:            "療法食",
		Environment:     "室内",
		LastVisit:       &lastVisit,
		InsuranceID:     &insuranceID,
		Remarks:         "咬傷注意",
		AnimalSpecies:   &model.AnimalSpecies{ID: 3, Name: "犬", SortOrder: 1},
		Insurance:       &model.Insurance{ID: insuranceID, Name: "アニコム", CoverageRate: 70, ContactPhone: "03-0000-0000"},
	})

	assert.Equal(t, "柴犬", resp.Breed)
	assert.Equal(t, "赤", resp.Color)
	assert.Equal(t, "DEA1.1陽性", *resp.BloodType)
	assert.Equal(t, "392140000123456", *resp.MicrochipNumber)
	// neutered_date は canonical 規約 (localTimePtr) でローカル化されるため、
	// tz 表現ではなくカレンダー日付で検証する (格納値 2016-05-20 を保持)。
	require.NotNil(t, resp.NeuteredDate)
	assert.Equal(t, "2016-05-20", resp.NeuteredDate.Format("2006-01-02"))
	require.NotNil(t, resp.AcquisitionType)
	assert.Equal(t, "purchased", *resp.AcquisitionType)
	assert.Equal(t, "medium", resp.DangerLevel)
	assert.Equal(t, "療法食", resp.Food)
	assert.Equal(t, "室内", resp.Environment)
	assert.Equal(t, &insuranceID, resp.InsuranceID)
	assert.Equal(t, "咬傷注意", resp.Remarks)
	require.NotNil(t, resp.Insurance)
	assert.Equal(t, "アニコム", resp.Insurance.Name)
	assert.Equal(t, 70, resp.Insurance.CoverageRate)
}
