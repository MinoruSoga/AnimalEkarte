package service

import (
	"fmt"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func validatePetGender(gender string) error {
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

// validatePetStatus は生存ステータスの値がドメイン上有効かを検証する
func validatePetStatus(status string) error {
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

// validatePetAcquisitionType は入手経路の値がドメイン上有効かを検証する
func validatePetAcquisitionType(t string) error {
	if t == "" {
		return nil
	}
	switch model.AcquisitionType(t) {
	case model.AcquisitionTypePurchase, model.AcquisitionTypeTransfer,
		model.AcquisitionTypeRescued, model.AcquisitionTypeOther:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("入手経路の値が不正です: %s", t))
	}
}

// validatePetDangerLevel は危険度の値がドメイン上有効かを検証する
func validatePetDangerLevel(level string) error {
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

// validateRequiredName は必須名前フィールドのバリデーションを行う。
// スペースのみ・NULL バイト・制御文字・255文字超が含まれていないかを検証する。
func validatePetForOwnerInput(input *CreatePetForOwnerInput) error {
	if err := validatePetGender(input.Gender); err != nil {
		return err
	}
	if err := validatePetStatus(input.Status); err != nil {
		return err
	}
	if err := validatePetAcquisitionType(input.AcquisitionType); err != nil {
		return err
	}
	return validatePetDangerLevel(input.DangerLevel)
}

func validateCreatePetInput(input *CreatePetInput) error {
	if err := validateRequiredName(input.Name); err != nil {
		return apperrors.Wrap(err, "failed to validate required name")
	}
	input.Name = strings.TrimSpace(input.Name)

	if input.Weight != nil && *input.Weight < 0 {
		return apperrors.WrapInvalidInput(ErrMsgWeightZeroOrMore)
	}
	if err := validatePetGender(input.Gender); err != nil {
		return apperrors.Wrap(err, "failed to validate pet gender")
	}
	if err := validatePetStatus(input.Status); err != nil {
		return apperrors.Wrap(err, "failed to validate pet status")
	}
	if err := validatePetAcquisitionType(input.AcquisitionType); err != nil {
		return apperrors.Wrap(err, "failed to validate pet acquisition type")
	}
	if err := validatePetDangerLevel(input.DangerLevel); err != nil {
		return apperrors.Wrap(err, "failed to validate pet danger level")
	}
	return nil
}

func validateUpdatePetInput(input *UpdatePetInput) error {
	if input.Name != nil {
		if err := validateRequiredName(*input.Name); err != nil {
			return apperrors.Wrap(err, "failed to validate required name")
		}
		trimmed := strings.TrimSpace(*input.Name)
		input.Name = &trimmed
	}
	if input.Weight != nil && *input.Weight < 0 {
		return apperrors.WrapInvalidInput(ErrMsgWeightZeroOrMore)
	}
	if input.Gender != nil {
		if err := validatePetGender(*input.Gender); err != nil {
			return apperrors.Wrap(err, "failed to validate pet gender")
		}
	}
	if input.AcquisitionType != nil {
		if err := validatePetAcquisitionType(*input.AcquisitionType); err != nil {
			return apperrors.Wrap(err, "failed to validate pet acquisition type")
		}
	}
	if input.DangerLevel != nil {
		if err := validatePetDangerLevel(*input.DangerLevel); err != nil {
			return apperrors.Wrap(err, "failed to validate pet danger level")
		}
	}
	return nil
}

// validateEmailFormat はメールアドレス形式が有効かを検証する（空の場合はスキップ）
