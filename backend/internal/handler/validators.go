package handler

import (
	"fmt"
	"unicode"

	"github.com/animal-ekarte/backend/internal/apperrors"
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
