package pet

import (
	"fmt"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const ErrMsgWeightZeroOrMore = "体重は0以上の値を入力してください"

// ValidateGender validates the optional public pet-gender value.
func ValidateGender(gender string) error {
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
func ValidateStatus(status string) error {
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
func ValidateAcquisitionType(t string) error {
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
func ValidateDangerLevel(level string) error {
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

func validateCreatePetInput(input *CreatePetInput) error {
	if err := sharedkernel.ValidateRequiredName(input.Name); err != nil {
		return apperrors.Wrap(err, "failed to validate required name")
	}
	input.Name = strings.TrimSpace(input.Name)

	if input.Weight != nil && *input.Weight < 0 {
		return apperrors.WrapInvalidInput(ErrMsgWeightZeroOrMore)
	}
	if err := ValidateGender(input.Gender); err != nil {
		return apperrors.Wrap(err, "failed to validate pet gender")
	}
	if err := ValidateStatus(input.Status); err != nil {
		return apperrors.Wrap(err, "failed to validate pet status")
	}
	if err := ValidateAcquisitionType(input.AcquisitionType); err != nil {
		return apperrors.Wrap(err, "failed to validate pet acquisition type")
	}
	if err := ValidateDangerLevel(input.DangerLevel); err != nil {
		return apperrors.Wrap(err, "failed to validate pet danger level")
	}
	return nil
}

// ValidateCreateInput validates and normalizes a pet create command.
// It is exported for temporary legacy compatibility facades.
func ValidateCreateInput(input *CreatePetInput) error {
	return validateCreatePetInput(input)
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
		if err := ValidateGender(*input.Gender); err != nil {
			return apperrors.Wrap(err, "failed to validate pet gender")
		}
	}
	if input.AcquisitionType != nil {
		if err := ValidateAcquisitionType(*input.AcquisitionType); err != nil {
			return apperrors.Wrap(err, "failed to validate pet acquisition type")
		}
	}
	if input.DangerLevel != nil {
		if err := ValidateDangerLevel(*input.DangerLevel); err != nil {
			return apperrors.Wrap(err, "failed to validate pet danger level")
		}
	}
	return nil
}

// ValidateUpdateInput validates and normalizes a pet update command.
// It is exported for temporary legacy compatibility facades.
func ValidateUpdateInput(input *UpdatePetInput) error {
	return validateUpdatePetInput(input)
}
