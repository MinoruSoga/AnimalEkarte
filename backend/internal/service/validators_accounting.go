package service

import (
	"fmt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func validateTaxType(taxType string) error {
	if taxType == "" {
		return nil
	}
	switch model.TaxType(taxType) {
	case model.TaxTypeIncluded, model.TaxTypeExcluded, model.TaxTypeExempt:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("税種別の値が不正です: %s", taxType))
	}
}

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

// validateCageType はケージ種別がドメイン上有効かを検証する
func validateNonNegativePrice(price *int64) error {
	if price == nil {
		return nil
	}
	if *price < 0 {
		return apperrors.WrapInvalidInput(ErrMsgPriceZeroOrMore)
	}
	return nil
}

// validateCoverageRate は保険補償率が 0〜100 の範囲内かを検証する (BUG-398)
