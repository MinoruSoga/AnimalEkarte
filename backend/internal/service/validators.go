package service

import (
	"fmt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// validatePetGender は性別の値がドメイン上有効かを検証する
func validatePetGender(gender string) error {
	if gender == "" {
		return nil
	}
	switch model.PetGender(gender) {
	case model.PetGenderMale, model.PetGenderFemale, model.PetGenderUnknown:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid gender: %s", gender))
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
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid status: %s", status))
	}
}

// validateDiscountRate は割引率が 0〜100 の範囲内かを検証する
func validateDiscountRate(rate float64) error {
	if rate < 0 || rate > 100 {
		return apperrors.WrapInvalidInput("discount_rate must be between 0 and 100")
	}
	return nil
}

// validateMembershipType は会員種別がドメイン上有効かを検証する
func validateMembershipType(t model.MembershipType) error {
	if t == "" {
		return nil
	}
	switch t {
	case model.MembershipTypeNonMember, model.MembershipTypeMember,
		model.MembershipTypeDeceased, model.MembershipTypeTransferred:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid membership_type: %s", t))
	}
}
