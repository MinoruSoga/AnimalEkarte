package staff

import (
	"unicode"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const (
	ErrMsgAtLeastOneField = sharedkernel.ErrMsgAtLeastOneField
	ErrMsgIDsNotEmpty     = sharedkernel.ErrMsgIDsNotEmpty
	ErrMsgInputNotNil     = sharedkernel.ErrMsgInputNotNil
)

func validateRequiredName(name string) error {
	return sharedkernel.ValidateRequiredName(name)
}

func validateOptionalName(name *string) error {
	return sharedkernel.ValidateOptionalName(name)
}

func validatePassword(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return apperrors.WrapInvalidInput("パスワードは8文字以上で入力してください")
	}
	if len([]byte(password)) > 72 {
		return apperrors.WrapInvalidInput("パスワードは72バイト以下で入力してください")
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		hasLetter = hasLetter || unicode.IsLetter(r)
		hasDigit = hasDigit || unicode.IsDigit(r)
	}
	if !hasLetter || !hasDigit {
		return apperrors.WrapInvalidInput("パスワードは英字と数字の両方を含めてください")
	}
	return nil
}
