package owner

import (
	"regexp"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// RFC 5322簡易的なメール形式パターン
var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// 電話番号パターン: 03-1234-5678, 090-1234-5678, 09012345678 等
var phonePattern = regexp.MustCompile(`^0\d{1,4}-?\d{1,4}-?\d{4}$`)

// 郵便番号パターン: 123-4567 または 1234567
var postalCodePattern = regexp.MustCompile(`^\d{3}-?\d{4}$`)

// validateEmailFormat はメールアドレス形式が有効かを検証する（空の場合はスキップ）
// POC-17: clinic/company use the same patterns via jp_email/jp_phone/jp_postal binding tags.
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
	if !postalCodePattern.MatchString(postalCode) {
		return apperrors.WrapInvalidInput("郵便番号の形式が正しくありません（例：123-4567 または 1234567）")
	}
	return nil
}

// ValidateEmailFormat exports contact email validation for cross-domain binding parity tests.
func ValidateEmailFormat(email string) error { return validateEmailFormat(email) }

// ValidatePhoneFormat exports contact phone validation for cross-domain binding parity tests.
func ValidatePhoneFormat(phone string) error { return validatePhoneFormat(phone) }

// ValidatePostalCodeFormat exports postal validation for cross-domain binding parity tests.
func ValidatePostalCodeFormat(postalCode string) error {
	return validatePostalCodeFormat(postalCode)
}
