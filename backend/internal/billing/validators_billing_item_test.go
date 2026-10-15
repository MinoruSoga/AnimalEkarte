package billing

// validators_billing_item_test.go — BE9-2C B③: service/validators_test.go から明細enum検証テストを移動。

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestValidateItemCategory(t *testing.T) {
	// Called by go test for BE7-1; production callers use validateItemCategory via billing_item CreateItem.
	valid := []model.ItemCategory{
		model.ItemCategoryExamination,
		model.ItemCategoryTest,
		model.ItemCategoryProcedure,
		model.ItemCategorySurgery,
		model.ItemCategoryMedicine,
		model.ItemCategoryFood,
		model.ItemCategoryGoods,
		model.ItemCategoryOther,
		model.ItemCategoryVaccine,
		model.ItemCategoryTrimming,
		model.ItemCategoryHotel,
		model.ItemCategoryTraining,
		model.ItemCategory("house_call"),
	}
	for _, cat := range valid {
		assert.NoError(t, validateItemCategory(string(cat)), "category %q", cat)
	}
	assert.Error(t, validateItemCategory("unknown"))
	assert.Error(t, validateItemCategory("invalid_category"))
}

func TestItemCategoryBinding_HouseCall(t *testing.T) {
	t.Run("completeAccountingItemRequest validates category", func(t *testing.T) {
		req := completeAccountingItemRequest{Category: "house_call", Name: "往診", Quantity: 1}
		assert.NoError(t, binding.Validator.ValidateStruct(&req))
		req.Category = "housecall"
		err := binding.Validator.ValidateStruct(&req)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Category")
	})
	t.Run("createBillingItemRequest validates category", func(t *testing.T) {
		req := createBillingItemRequest{BillingID: 1, Category: "house_call", Name: "往診"}
		assert.NoError(t, binding.Validator.ValidateStruct(&req))
		req.Category = "housecall"
		err := binding.Validator.ValidateStruct(&req)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Category")
	})
	t.Run("campaign create validates target_categories", func(t *testing.T) {
		create := createCampaignRequest{
			Name:             "キャンペーン",
			StartDate:        "2026-10-01",
			EndDate:          "2026-10-31",
			DiscountType:     "rate",
			TargetCategories: []string{"house_call"},
		}
		assert.NoError(t, binding.Validator.ValidateStruct(&create))
		create.TargetCategories = []string{"housecall"}
		err := binding.Validator.ValidateStruct(&create)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "TargetCategories")
	})
	t.Run("campaign update validates target_categories", func(t *testing.T) {
		update := updateCampaignRequest{TargetCategories: &[]string{"house_call"}}
		assert.NoError(t, binding.Validator.ValidateStruct(&update))
		update.TargetCategories = &[]string{"housecall"}
		err := binding.Validator.ValidateStruct(&update)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "TargetCategories")
	})
}

func TestValidateItemSource(t *testing.T) {
	assert.NoError(t, validateItemSource(string(model.ItemSourceMedicalRecord)))
	assert.Error(t, validateItemSource("invalid_source"))
}
