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

// ValidatePetGender / Status / AcquisitionType / DangerLevel は owner と pet の境界で
// 同一契約を共有する（POC-11 / X-09: copy-paste drift 防止）。空文字は未指定として許可。

func ValidatePetGender(gender string) error {
	if gender == "" {
		return nil
	}
	switch model.PetGender(gender) {
	case model.PetGenderMale, model.PetGenderFemale, model.PetGenderUnknown:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("性別の値が不正です: %s", gender))
	}
}

func ValidatePetStatus(status string) error {
	if status == "" {
		return nil
	}
	switch model.PetStatus(status) {
	case model.PetStatusAlive, model.PetStatusDeceased:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("ステータスの値が不正です: %s", status))
	}
}

func ValidatePetAcquisitionType(acquisitionType string) error {
	if acquisitionType == "" {
		return nil
	}
	switch model.AcquisitionType(acquisitionType) {
	case model.AcquisitionTypePurchase, model.AcquisitionTypeTransfer,
		model.AcquisitionTypeRescued, model.AcquisitionTypeOther:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("入手経路の値が不正です: %s", acquisitionType))
	}
}

func ValidatePetDangerLevel(level string) error {
	if level == "" {
		return nil
	}
	switch model.DangerLevel(level) {
	case model.DangerLevelLow, model.DangerLevelMedium, model.DangerLevelHigh:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("危険度の値が不正です: %s", level))
	}
}
