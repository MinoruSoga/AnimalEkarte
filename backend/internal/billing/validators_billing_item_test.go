package billing

// validators_billing_item_test.go — BE9-2C B③: service/validators_test.go から明細enum検証テストを移動。

import (
	"testing"

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
	}
	for _, cat := range valid {
		assert.NoError(t, validateItemCategory(string(cat)), "category %q", cat)
	}
	assert.Error(t, validateItemCategory("unknown"))
	assert.Error(t, validateItemCategory("invalid_category"))
}

func TestValidateItemSource(t *testing.T) {
	assert.NoError(t, validateItemSource(string(model.ItemSourceMedicalRecord)))
	assert.Error(t, validateItemSource("invalid_source"))
}
