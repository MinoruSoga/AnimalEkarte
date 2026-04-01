package service

import (
	"fmt"
	"regexp"
	"strings"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// RFC 5322簡易的なメール形式パターン
var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

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

// validatePetAcquisitionType は入手経路の値がドメイン上有効かを検証する
func validatePetAcquisitionType(t string) error {
	if t == "" {
		return nil
	}
	switch model.AcquisitionType(t) {
	case model.AcquisitionTypePurchase, model.AcquisitionTypeTransfer,
		model.AcquisitionTypeProtected, model.AcquisitionTypeOther:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid acquisition_type: %s", t))
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
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid danger_level: %s", level))
	}
}

// validateRequiredName は必須名前フィールドのバリデーションを行う。
// スペースのみ・NULL バイト・制御文字が含まれていないかを検証する。
func validateRequiredName(name string) error {
	if strings.TrimSpace(name) == "" {
		return apperrors.WrapInvalidInput("名前を入力してください")
	}
	for _, r := range name {
		if r == '\u0000' {
			return apperrors.WrapInvalidInput("名前に無効な文字が含まれています")
		}
		if r < 0x20 && r != '\t' && r != '\n' {
			return apperrors.WrapInvalidInput("名前に無効な文字が含まれています")
		}
	}
	return nil
}

// validateOwnerName はオーナー名のバリデーションを行う（validateRequiredName のエイリアス）
func validateOwnerName(name string) error {
	return validateRequiredName(name)
}

// validateDiscountRate は割引率が 0〜100 の範囲内かを検証する
func validateDiscountRate(rate float64) error {
	if rate < 0 || rate > 100 {
		return apperrors.WrapInvalidInput("discount_rate must be between 0 and 100")
	}
	return nil
}

// validateMedicalImageType は診療画像種別がドメイン上有効かを検証する
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
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid image_type: %s", t))
	}
}

// validateStaffRole はスタッフ役職がドメイン上有効かを検証する
func validateStaffRole(role model.StaffRole) error {
	switch role {
	case model.StaffRoleVeterinarian, model.StaffRoleNurse, model.StaffRoleTrimmer,
		model.StaffRoleReception, model.StaffRoleManager:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid staff_role: %s", role))
	}
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

// validateEmailFormat はメールアドレス形式が有効かを検証する（空の場合はスキップ）
func validateEmailFormat(email string) error {
	if email == "" {
		return nil
	}
	if !emailPattern.MatchString(email) {
		return apperrors.WrapInvalidInput("メールアドレスの形式が正しくありません")
	}
	return nil
}

// validatePhoneFormat は電話番号形式が有効かを検証する（空の場合はスキップ）
// 許可形式: 03-1234-5678, 090-1234-5678, 09012345678, 0312345678
func validatePhoneFormat(phone string) error {
	if phone == "" {
		return nil
	}
	// ハイフンなし、またはハイフン付き電話番号を許可
	phonePattern := regexp.MustCompile(`^0\d{1,4}-?\d{1,4}-?\d{4}$`)
	if !phonePattern.MatchString(phone) {
		return apperrors.WrapInvalidInput("電話番号の形式が正しくありません（例：090-1234-5678 または 09012345678）")
	}
	return nil
}

// validatePostalCodeFormat は郵便番号形式が有効かを検証する（空の場合はスキップ）
// 許可形式: 123-4567 または 1234567
func validatePostalCodeFormat(postalCode string) error {
	if postalCode == "" {
		return nil
	}
	postalCodePattern := regexp.MustCompile(`^\d{3}-?\d{4}$`)
	if !postalCodePattern.MatchString(postalCode) {
		return apperrors.WrapInvalidInput("郵便番号の形式が正しくありません（例：123-4567 または 1234567）")
	}
	return nil
}
