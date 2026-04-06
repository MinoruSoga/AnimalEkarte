package handler

import (
	"fmt"
	"unicode"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// validateTaxType は課税種別文字列がドメイン上有効かを検証する
func validateTaxType(v string) error {
	switch model.TaxType(v) {
	case model.TaxTypeIncluded, model.TaxTypeExcluded, model.TaxTypeExempt:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid tax_type: %s", v))
	}
}

// validateItemCategory は明細カテゴリ文字列がドメイン上有効かを検証する
func validateItemCategory(v string) error {
	switch model.ItemCategory(v) {
	case model.ItemCategoryExamination, model.ItemCategoryTest, model.ItemCategoryProcedure,
		model.ItemCategorySurgery, model.ItemCategoryMedicine, model.ItemCategoryFood,
		model.ItemCategoryGoods, model.ItemCategoryOther:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid category: %s", v))
	}
}

// validateItemSource は明細ソース文字列がドメイン上有効かを検証する
func validateItemSource(v string) error {
	switch model.ItemSource(v) {
	case model.ItemSourceMedicalRecord, model.ItemSourceManual, model.ItemSourceHospitalization:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid source: %s", v))
	}
}

// validateTreatmentItemType は治療項目種別文字列がドメイン上有効かを検証する
func validateTreatmentItemType(v string) error {
	switch model.TreatmentItemType(v) {
	case model.TreatmentItemTypeConsultation, model.TreatmentItemTypeProcedure,
		model.TreatmentItemTypeMedicine, model.TreatmentItemTypeOther:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid item_type: %s", v))
	}
}

// validatePassword はパスワード複雑性を検証する（BUG-139）。
// 8文字以上、英字1文字以上、数字1文字以上が必須。
func validatePassword(pw string) error {
	if len(pw) < 8 {
		return apperrors.WrapInvalidInput("パスワードは8文字以上で入力してください")
	}
	var hasLetter, hasDigit bool
	for _, r := range pw {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return apperrors.WrapInvalidInput("パスワードは英字と数字の両方を含めてください")
	}
	return nil
}
