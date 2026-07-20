package service

import (
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
	"github.com/animal-ekarte/backend/internal/model"
)

func validateMedicalImageType(t string) error {
	if t == "" {
		return nil
	}
	switch model.MedicalImageType(t) {
	case model.MedicalImageTypeXray, model.MedicalImageTypeEcho, model.MedicalImageTypePhoto,
		model.MedicalImageTypeEndoscope, model.MedicalImageTypeCT, model.MedicalImageTypeMRI,
		model.MedicalImageTypeMicroscope, model.MedicalImageTypeOther:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("画像種別の値が不正です: %s", t))
	}
}

// validateAnesthesiaType は麻酔種別がドメイン上有効かを検証する
func validateAnesthesiaType(anesthesia string) error {
	return sharedkernel.ValidateAnesthesiaType(anesthesia)
}

// validateTaxType は税種別がドメイン上有効かを検証する
func validateCageType(cageType string) error {
	return sharedkernel.ValidateCageType(cageType)
}

// validateNonNegativePrice は価格フィールドが 0 以上かを検証する（nil の場合はスキップ）(BUG-380)
func validateCoverageRate(rate int) error {
	if rate < 0 || rate > 100 {
		return apperrors.WrapInvalidInput("補償率は0〜100の範囲で入力してください")
	}
	return nil
}

// validateOptionalCoverageRate は nil 許容の保険補償率バリデーション (BUG-398)
func validateOptionalCoverageRate(rate *int) error {
	if rate == nil {
		return nil
	}
	return validateCoverageRate(*rate)
}

// validatePassword はパスワード複雑性を検証する（BUG-139）。
// 8文字以上、英字1文字以上、数字1文字以上が必須。
func validateCageSize(cageSize string) error {
	return sharedkernel.ValidateCageSize(cageSize)
}
