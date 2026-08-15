package billing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToCampaignResponse(t *testing.T) {
	t.Run("converts campaign with categories and items", func(t *testing.T) {
		m := &model.Campaign{
			ID:            1,
			ClinicID:      2,
			Name:          "夏の割引キャンペーン",
			StartDate:     time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			EndDate:       time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
			DiscountType:  model.CampaignDiscountTypeRate,
			DiscountValue: 10,
			IsActive:      true,
			SortOrder:     3,
			TargetCategories: []model.CampaignTargetCategory{
				{Category: model.ItemCategory("examination")},
				{Category: model.ItemCategory("medicine")},
			},
			TargetItems: []model.CampaignTargetItem{
				{MerchandiseItemID: 10},
				{MerchandiseItemID: 20},
			},
		}

		resp := toCampaignResponse(m)

		assert.Equal(t, uint64(1), resp.ID)
		assert.Equal(t, uint64(2), resp.ClinicID)
		assert.Equal(t, "夏の割引キャンペーン", resp.Name)
		assert.Equal(t, "2026-07-01", resp.StartDate)
		assert.Equal(t, "2026-08-31", resp.EndDate)
		assert.Equal(t, "rate", resp.DiscountType)
		assert.Equal(t, 10.0, resp.DiscountValue)
		assert.True(t, resp.IsActive)
		assert.Equal(t, 3, resp.SortOrder)
		assert.Equal(t, []string{"examination", "medicine"}, resp.TargetCategories)
		assert.Equal(t, []uint64{10, 20}, resp.TargetItemIDs)
	})

	t.Run("converts campaign with no categories or items", func(t *testing.T) {
		m := &model.Campaign{
			ID:           2,
			ClinicID:     1,
			Name:         "対象なしキャンペーン",
			StartDate:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
			DiscountType: model.CampaignDiscountTypeAmount,
		}

		resp := toCampaignResponse(m)

		assert.Empty(t, resp.TargetCategories)
		assert.Empty(t, resp.TargetItemIDs)
		assert.Equal(t, "amount", resp.DiscountType)
	})
}
