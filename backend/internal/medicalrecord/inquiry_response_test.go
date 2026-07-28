package medicalrecord

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToInquiryResponse(t *testing.T) {
	t.Run("maps all fields including non-nil Appetite and WaterIntake", func(t *testing.T) {
		typeID := uint64(3)
		staffID := uint64(7)
		appetite := model.AppetiteLevelDecreased
		waterIntake := model.WaterIntakeLevelIncreased
		createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		updatedAt := time.Date(2026, 1, 3, 4, 5, 6, 0, time.UTC)

		inquiry := &model.Inquiry{
			ID:                   1,
			MedicalRecordID:      10,
			ChiefComplaintTypeID: &typeID,
			ChiefComplaint:       "嘔吐",
			History:              "3日前から",
			CurrentMedications:   "なし",
			AllergyInfo:          "なし",
			LastMeal:             "昨夜",
			LastDefecation:       "今朝",
			LastUrination:        "今朝",
			Appetite:             &appetite,
			WaterIntake:          &waterIntake,
			OwnerObservations:    "元気がない",
			Notes:                "経過観察",
			StaffID:              &staffID,
			CreatedAt:            createdAt,
			UpdatedAt:            updatedAt,
		}

		resp := toInquiryResponse(inquiry)

		assert.Equal(t, uint64(1), resp.ID)
		assert.Equal(t, uint64(10), resp.MedicalRecordID)
		require.NotNil(t, resp.ChiefComplaintTypeID)
		assert.Equal(t, typeID, *resp.ChiefComplaintTypeID)
		assert.Equal(t, "嘔吐", resp.ChiefComplaint)
		assert.Equal(t, "3日前から", resp.History)
		assert.Equal(t, "なし", resp.CurrentMedications)
		assert.Equal(t, "なし", resp.AllergyInfo)
		assert.Equal(t, "昨夜", resp.LastMeal)
		assert.Equal(t, "今朝", resp.LastDefecation)
		assert.Equal(t, "今朝", resp.LastUrination)
		require.NotNil(t, resp.Appetite)
		assert.Equal(t, "decreased", *resp.Appetite)
		require.NotNil(t, resp.WaterIntake)
		assert.Equal(t, "increased", *resp.WaterIntake)
		assert.Equal(t, "元気がない", resp.OwnerObservations)
		assert.Equal(t, "経過観察", resp.Notes)
		require.NotNil(t, resp.StaffID)
		assert.Equal(t, staffID, *resp.StaffID)
	})

	t.Run("nil Appetite, WaterIntake, ChiefComplaintTypeID, StaffID stay nil", func(t *testing.T) {
		inquiry := &model.Inquiry{
			ID:              2,
			MedicalRecordID: 20,
		}

		resp := toInquiryResponse(inquiry)

		assert.Nil(t, resp.ChiefComplaintTypeID)
		assert.Nil(t, resp.Appetite)
		assert.Nil(t, resp.WaterIntake)
		assert.Nil(t, resp.StaffID)
	})
}
