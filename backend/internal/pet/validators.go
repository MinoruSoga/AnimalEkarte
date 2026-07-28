package pet

import (
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const ErrMsgWeightZeroOrMore = "体重は0以上の値を入力してください"

func validatePetGender(gender string) error {
	return sharedkernel.ValidatePetGender(gender)
}

// validatePetStatus は生存ステータスの値がドメイン上有効かを検証する
func validatePetStatus(status string) error {
	return sharedkernel.ValidatePetStatus(status)
}

// validatePetAcquisitionType は入手経路の値がドメイン上有効かを検証する
func validatePetAcquisitionType(t string) error {
	return sharedkernel.ValidatePetAcquisitionType(t)
}

// validatePetDangerLevel は危険度の値がドメイン上有効かを検証する
func validatePetDangerLevel(level string) error {
	return sharedkernel.ValidatePetDangerLevel(level)
}

func validateCreatePetInput(input *CreatePetInput) error {
	if err := sharedkernel.ValidateRequiredName(input.Name); err != nil {
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
		if err := sharedkernel.ValidateRequiredName(*input.Name); err != nil {
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
