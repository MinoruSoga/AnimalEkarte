package medicalrecord

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToConsultationResponse(t *testing.T) {
	t.Run("converts consultation with all fields set", func(t *testing.T) {
		price := int64(3300)
		duration := 20
		parentID := uint64(8)
		createdAt := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
		updatedAt := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)

		m := &model.Consultation{
			ID:            1,
			ClinicID:      2,
			Name:          "予防接種相談",
			Price:         &price,
			IsActive:      true,
			Description:   "説明文",
			TimeCondition: "morning",
			Duration:      &duration,
			ParentID:      &parentID,
			TaxType:       model.TaxTypeIncluded,
			TaxRate:       0.1,
			SortOrder:     3,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		}

		resp := toConsultationResponse(m)

		assert.Equal(t, uint64(1), resp.ID)
		assert.Equal(t, uint64(2), resp.ClinicID)
		assert.Equal(t, "予防接種相談", resp.Name)
		require.NotNil(t, resp.Price)
		assert.Equal(t, price, *resp.Price)
		assert.True(t, resp.IsActive)
		assert.Equal(t, "説明文", resp.Description)
		assert.Equal(t, "morning", resp.TimeCondition)
		require.NotNil(t, resp.Duration)
		assert.Equal(t, duration, *resp.Duration)
		require.NotNil(t, resp.ParentID)
		assert.Equal(t, parentID, *resp.ParentID)
		assert.Equal(t, model.TaxTypeIncluded, resp.TaxType)
		assert.Equal(t, 0.1, resp.TaxRate)
		assert.Equal(t, 3, resp.SortOrder)
		assert.Equal(t, createdAt.In(time.Local), resp.CreatedAt)
		assert.Equal(t, updatedAt.In(time.Local), resp.UpdatedAt)
	})

	t.Run("converts consultation with nil optional fields", func(t *testing.T) {
		m := &model.Consultation{
			ID:       2,
			ClinicID: 1,
			Name:     "栄養相談",
			IsActive: false,
		}

		resp := toConsultationResponse(m)

		assert.Nil(t, resp.Price)
		assert.Nil(t, resp.Duration)
		assert.Nil(t, resp.ParentID)
		assert.False(t, resp.IsActive)
	})
}
