package billing

// validators_billing_item.go — BE9-2C B③: service/validators_accounting.go から明細系
// enum 検証を移動（唯一の consumer=billing_item_service が本 package へ移動したため）。

import (
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// validateItemCategory は明細カテゴリ文字列がドメイン上有効かを検証する
func validateItemCategory(v string) error {
	switch model.ItemCategory(v) {
	case model.ItemCategoryExamination, model.ItemCategoryTest, model.ItemCategoryProcedure,
		model.ItemCategorySurgery, model.ItemCategoryMedicine, model.ItemCategoryFood,
		model.ItemCategoryGoods, model.ItemCategoryOther,
		model.ItemCategoryVaccine, model.ItemCategoryTrimming, model.ItemCategoryHotel,
		model.ItemCategoryTraining:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("明細カテゴリの値が不正です: %s", v))
	}
}

// validateItemSource は明細ソース文字列がドメイン上有効かを検証する
func validateItemSource(v string) error {
	switch model.ItemSource(v) {
	case model.ItemSourceMedicalRecord, model.ItemSourceManual, model.ItemSourceHospitalization, model.ItemSourceTrimming:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("明細ソースの値が不正です: %s", v))
	}
}
