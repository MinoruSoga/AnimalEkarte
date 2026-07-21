package service

import (
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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

// validatePassword はパスワード複雑性を検証する（BUG-139）。
// 8文字以上、英字1文字以上、数字1文字以上が必須。
func validateCageSize(cageSize string) error {
	return sharedkernel.ValidateCageSize(cageSize)
}
