package owner

import (
	"fmt"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func validateDiscountRate(rate float64) error {
	return sharedkernel.ValidateDiscountRate(rate)
}

// validateMedicalImageType は診療画像種別がドメイン上有効かを検証する
func validateMembershipType(t model.MembershipType) error {
	if t == "" {
		return nil
	}
	switch t {
	case model.MembershipTypeNonMember, model.MembershipTypeMember,
		model.MembershipTypeDeceased, model.MembershipTypeTransferred:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("会員種別の値が不正です: %s", t))
	}
}

func validateCreateOwnerInput(input *CreateOwnerInput) error {
	if err := sharedkernel.ValidateRequiredName(input.OwnerName); err != nil {
		return apperrors.Wrap(err, "failed to validate required name")
	}
	input.OwnerName = strings.TrimSpace(input.OwnerName)

	if err := validateEmailFormat(input.Email); err != nil {
		return apperrors.Wrap(err, "failed to validate email format")
	}
	if err := validatePhoneFormat(input.Phone); err != nil {
		return apperrors.Wrap(err, "failed to validate phone format")
	}
	if err := validatePostalCodeFormat(input.PostalCode); err != nil {
		return apperrors.Wrap(err, "failed to validate postal code format")
	}
	if err := validateDiscountRate(input.DiscountRate); err != nil {
		return apperrors.Wrap(err, "failed to validate discount rate")
	}
	if err := validateMembershipType(input.MembershipType); err != nil {
		return apperrors.Wrap(err, "failed to validate membership type")
	}
	for i := range input.Pets {
		if err := validatePetForOwnerInput(&input.Pets[i]); err != nil {
			return apperrors.Wrap(err, fmt.Sprintf("pets[%d]", i))
		}
	}
	return nil
}

func validateUpdateOwnerInput(input *UpdateOwnerInput) error {
	if input.OwnerName != nil {
		if err := sharedkernel.ValidateRequiredName(*input.OwnerName); err != nil {
			return apperrors.Wrap(err, "failed to validate required name")
		}
		trimmed := strings.TrimSpace(*input.OwnerName)
		input.OwnerName = &trimmed
	}
	if input.Email != nil {
		if err := validateEmailFormat(*input.Email); err != nil {
			return apperrors.Wrap(err, "failed to validate email format")
		}
	}
	if input.Phone != nil {
		if err := validatePhoneFormat(*input.Phone); err != nil {
			return apperrors.Wrap(err, "failed to validate phone format")
		}
	}
	if input.PostalCode != nil {
		if err := validatePostalCodeFormat(*input.PostalCode); err != nil {
			return apperrors.Wrap(err, "failed to validate postal code format")
		}
	}
	if input.DiscountRate != nil {
		if err := validateDiscountRate(*input.DiscountRate); err != nil {
			return apperrors.Wrap(err, "failed to validate discount rate")
		}
	}
	if input.MembershipType != nil {
		if err := validateMembershipType(*input.MembershipType); err != nil {
			return apperrors.Wrap(err, "failed to validate membership type")
		}
	}
	return nil
}

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

func validatePetGender(gender string) error {
	return sharedkernel.ValidatePetGender(gender)
}

func validatePetStatus(status string) error {
	return sharedkernel.ValidatePetStatus(status)
}

func validatePetAcquisitionType(acquisitionType string) error {
	return sharedkernel.ValidatePetAcquisitionType(acquisitionType)
}

func validatePetDangerLevel(level string) error {
	return sharedkernel.ValidatePetDangerLevel(level)
}
