package billing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestCreateCampaignRequest_ToServiceInput(t *testing.T) {
	t.Run("converts valid request", func(t *testing.T) {
		req := createCampaignRequest{
			Name:             "夏の割引キャンペーン",
			StartDate:        "2026-07-01",
			EndDate:          "2026-08-31",
			DiscountType:     "rate",
			DiscountValue:    10,
			IsActive:         true,
			SortOrder:        2,
			TargetCategories: []string{"examination", "medicine"},
			TargetItemIDs:    []uint64{1, 2, 3},
		}

		input, err := req.toServiceInput()

		require.NoError(t, err)
		assert.Equal(t, "夏の割引キャンペーン", input.Name)
		assert.Equal(t, "2026-07-01", input.StartDate.Format("2006-01-02"))
		assert.Equal(t, "2026-08-31", input.EndDate.Format("2006-01-02"))
		assert.Equal(t, model.CampaignDiscountType("rate"), input.DiscountType)
		assert.Equal(t, 10.0, input.DiscountValue)
		assert.True(t, input.IsActive)
		assert.Equal(t, 2, input.SortOrder)
		assert.Equal(t, []model.ItemCategory{"examination", "medicine"}, input.TargetCategories)
		assert.Equal(t, []uint64{1, 2, 3}, input.TargetItemIDs)
	})

	t.Run("returns error for invalid start_date format", func(t *testing.T) {
		req := createCampaignRequest{
			Name:      "テスト",
			StartDate: "2026/07/01",
			EndDate:   "2026-08-31",
		}

		input, err := req.toServiceInput()

		assert.Nil(t, input)
		require.Error(t, err)
	})

	t.Run("returns error for invalid end_date format", func(t *testing.T) {
		req := createCampaignRequest{
			Name:      "テスト",
			StartDate: "2026-07-01",
			EndDate:   "not-a-date",
		}

		input, err := req.toServiceInput()

		assert.Nil(t, input)
		require.Error(t, err)
	})

	t.Run("handles empty target categories", func(t *testing.T) {
		req := createCampaignRequest{
			Name:      "テスト",
			StartDate: "2026-07-01",
			EndDate:   "2026-08-31",
		}

		input, err := req.toServiceInput()

		require.NoError(t, err)
		assert.Empty(t, input.TargetCategories)
	})
}

func TestUpdateCampaignRequest_ToServiceInput(t *testing.T) {
	t.Run("returns nil fields when nothing specified", func(t *testing.T) {
		input, err := (&updateCampaignRequest{}).toServiceInput()

		require.NoError(t, err)
		assert.Nil(t, input.Name)
		assert.Nil(t, input.StartDate)
		assert.Nil(t, input.EndDate)
		assert.Nil(t, input.DiscountType)
		assert.Nil(t, input.DiscountValue)
		assert.Nil(t, input.IsActive)
		assert.Nil(t, input.SortOrder)
		assert.Nil(t, input.TargetCategories)
		assert.Nil(t, input.TargetItemIDs)
	})

	t.Run("converts all provided fields", func(t *testing.T) {
		name := "秋の割引キャンペーン"
		start := "2026-09-01"
		end := "2026-09-30"
		discountType := "amount"
		discountValue := 500.0
		isActive := false
		sortOrder := 3
		targetCategories := []string{"vaccine"}
		targetItemIDs := []uint64{5, 6}

		req := updateCampaignRequest{
			Name:             &name,
			StartDate:        &start,
			EndDate:          &end,
			DiscountType:     &discountType,
			DiscountValue:    &discountValue,
			IsActive:         &isActive,
			SortOrder:        &sortOrder,
			TargetCategories: &targetCategories,
			TargetItemIDs:    &targetItemIDs,
		}

		input, err := req.toServiceInput()

		require.NoError(t, err)
		require.NotNil(t, input.Name)
		assert.Equal(t, name, *input.Name)
		require.NotNil(t, input.StartDate)
		assert.Equal(t, "2026-09-01", input.StartDate.Format("2006-01-02"))
		require.NotNil(t, input.EndDate)
		assert.Equal(t, "2026-09-30", input.EndDate.Format("2006-01-02"))
		require.NotNil(t, input.DiscountType)
		assert.Equal(t, model.CampaignDiscountType("amount"), *input.DiscountType)
		assert.Equal(t, &discountValue, input.DiscountValue)
		assert.Equal(t, &isActive, input.IsActive)
		assert.Equal(t, &sortOrder, input.SortOrder)
		require.NotNil(t, input.TargetCategories)
		assert.Equal(t, []model.ItemCategory{"vaccine"}, *input.TargetCategories)
		assert.Equal(t, &targetItemIDs, input.TargetItemIDs)
	})

	t.Run("returns error for invalid start_date format", func(t *testing.T) {
		start := "invalid"
		req := updateCampaignRequest{StartDate: &start}

		input, err := req.toServiceInput()

		assert.Nil(t, input)
		require.Error(t, err)
	})

	t.Run("returns error for invalid end_date format", func(t *testing.T) {
		end := "invalid"
		req := updateCampaignRequest{EndDate: &end}

		input, err := req.toServiceInput()

		assert.Nil(t, input)
		require.Error(t, err)
	})
}
