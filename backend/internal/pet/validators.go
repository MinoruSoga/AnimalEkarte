package pet

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const ErrMsgWeightZeroOrMore = "体重は0以上の値を入力してください"

// EMR-174: 任意記録フィールドの上限（request binding と同値。service 層でも再検証する）。
const (
	nameOriginMaxRunes   = 500
	meetingStoryMaxRunes = 2000
)

// normalizeOptionalNarrative は任意記録 *string をトリムし、空白のみを nil（NULL）に
// 正規化する。上限超過は invalid input。
func normalizeOptionalNarrative(label string, value *string, maxRunes int) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if utf8.RuneCountInString(trimmed) > maxRunes {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("%sは%d文字以内で入力してください", label, maxRunes))
	}
	if trimmed == "" {
		return nil, nil
	}
	return &trimmed, nil
}

// normalizeOptionalNarrativeTriState は PATCH tri-state（nil=変更なし / &nil=クリア /
// &&value=更新）版の正規化。値が送られた場合のみトリム＋上限検証を行い、空白のみは
// NULL クリア（&nil）と同義に丸める。
func normalizeOptionalNarrativeTriState(label string, value **string, maxRunes int) (**string, error) {
	if value == nil {
		return nil, nil
	}
	normalized, err := normalizeOptionalNarrative(label, *value, maxRunes)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

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
	var err error
	if input.NameOrigin, err = normalizeOptionalNarrative("名前の由来", input.NameOrigin, nameOriginMaxRunes); err != nil {
		return apperrors.Wrap(err, "failed to validate name origin")
	}
	if input.MeetingStory, err = normalizeOptionalNarrative("出逢いのストーリー", input.MeetingStory, meetingStoryMaxRunes); err != nil {
		return apperrors.Wrap(err, "failed to validate meeting story")
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
	var err error
	if input.NameOrigin, err = normalizeOptionalNarrativeTriState("名前の由来", input.NameOrigin, nameOriginMaxRunes); err != nil {
		return apperrors.Wrap(err, "failed to validate name origin")
	}
	if input.MeetingStory, err = normalizeOptionalNarrativeTriState("出逢いのストーリー", input.MeetingStory, meetingStoryMaxRunes); err != nil {
		return apperrors.Wrap(err, "failed to validate meeting story")
	}
	return nil
}
