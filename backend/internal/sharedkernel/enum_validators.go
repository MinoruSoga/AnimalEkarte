package sharedkernel

import (
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// マスタ enum 検証（BE9-2D ⑥で昇格）。ValidateTaxType は medicalrecord（procedure/consultation/
// medicine）と billing 系（billing_item 等・未移行）の恒久跨ぎ、cage/anesthesia は medicalrecord
// 帰属だが同ファイル由来のため同時に単一実装化する。空文字は「未指定」として許可（旧契約維持）。

func ValidateAnesthesiaType(anesthesia string) error {
	if anesthesia == "" {
		return nil
	}
	switch model.AnesthesiaType(anesthesia) {
	case model.AnesthesiaTypeNone, model.AnesthesiaTypeLocal,
		model.AnesthesiaTypeSedation, model.AnesthesiaTypeGeneral:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("麻酔種別の値が不正です: %s", anesthesia))
	}
}

func ValidateCageType(cageType string) error {
	if cageType == "" {
		return nil
	}
	switch model.CageType(cageType) {
	case model.CageTypeICU, model.CageTypeDog, model.CageTypeCat, model.CageTypeGeneral:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("ケージ種別の値が不正です: %s", cageType))
	}
}

func ValidateCageSize(cageSize string) error {
	if cageSize == "" {
		return nil
	}
	switch model.CageSize(cageSize) {
	case model.CageSizeSmall, model.CageSizeMedium, model.CageSizeLarge:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid cage_size: %s", cageSize))
	}
}

func ValidateTaxType(taxType string) error {
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
