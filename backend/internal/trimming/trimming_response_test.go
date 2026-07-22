package trimming

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToTrimmingResponse_MapsDetailAndRelations(t *testing.T) {
	start := time.Date(2026, 7, 22, 1, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	createdAt := start.Add(-time.Hour)
	updatedAt := end.Add(time.Hour)
	petID := uint64(20)
	staffID := uint64(30)
	courseID := uint64(40)
	weight := 5.4
	temperature := 38.2
	price := int64(8_800)

	response := toTrimmingResponse(&model.Reservation{
		ID:                10,
		ClinicID:          1,
		ReservationTypeID: 2,
		StartTime:         start,
		EndTime:           end,
		PetID:             &petID,
		DoctorID:          &staffID,
		Status:            model.ReservationStatusCompleted,
		Source:            model.ReservationSourceManual,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		Pet: &model.Pet{
			ID:        petID,
			Name:      "ポチ",
			PetNumber: "P-20",
			Weight:    &weight,
			Status:    model.PetStatusAlive,
			Breed:     "トイプードル",
			Owner:     &model.Owner{ID: 50, Name: "飼主"},
			AnimalSpecies: &model.AnimalSpecies{
				ID:   60,
				Name: "犬",
			},
		},
		Doctor: &model.Staff{ID: staffID, Name: "担当者"},
		TrimmingDetail: &model.AppointmentTrimmingDetail{
			CourseID:        &courseID,
			StyleRequest:    "短め",
			BodyWeight:      &weight,
			BWUnit:          model.BodyWeightUnitKg,
			BodyTemperature: &temperature,
			UsedShampoo:     "薬用",
			UsedRibbon:      "青",
			Remarks:         "問題なし",
			StyleImage:      "before.jpg",
			CompletedImage:  "after.jpg",
			Course:          &model.TrimmingCourse{ID: courseID, Name: "全身", Price: &price},
			Options:         []model.TrimmingOption{{ID: 70, Name: "爪切り"}},
		},
	})

	assert.Equal(t, uint64(10), response.ID)
	assert.Equal(t, start.In(time.Local), response.StartTime)
	assert.Equal(t, end.In(time.Local), response.EndTime)
	assert.True(t, response.HasDetail)
	assert.Equal(t, "短め", response.StyleRequest)
	assert.Equal(t, model.BodyWeightUnitKg, model.BodyWeightUnit(response.BWUnit))
	require.NotNil(t, response.Pet)
	assert.Equal(t, "トイプードル", response.Pet.Breed)
	require.NotNil(t, response.Pet.Owner)
	assert.Equal(t, "飼主", response.Pet.Owner.OwnerName)
	require.NotNil(t, response.Pet.AnimalSpecies)
	assert.Equal(t, "犬", response.Pet.AnimalSpecies.Name)
	require.NotNil(t, response.Staff)
	assert.Equal(t, "担当者", response.Staff.Name)
	require.NotNil(t, response.Course)
	assert.Equal(t, price, response.Course.Price)
	require.Len(t, response.Options, 1)
	assert.Equal(t, "爪切り", response.Options[0].Name)
}

func TestToTrimmingResponse_DefaultsWithoutDetailOrRelations(t *testing.T) {
	response := toTrimmingResponse(&model.Reservation{})

	assert.False(t, response.HasDetail)
	assert.Equal(t, "Kg", response.BWUnit)
	assert.Empty(t, response.Options)
	assert.Nil(t, response.Pet)
	assert.Nil(t, response.Staff)
	assert.Nil(t, toOwnerSummary(nil))
	assert.Nil(t, toPetSummary(nil))
	assert.Nil(t, toStaffSummary(nil))
}
