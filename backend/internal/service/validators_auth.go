package service

import (
	"unicode"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func validatePassword(pw string) error {
	if len(pw) < 8 {
		return apperrors.WrapInvalidInput("パスワードは8文字以上で入力してください")
	}
	var hasLetter, hasDigit bool
	for _, r := range pw {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return apperrors.WrapInvalidInput("パスワードは英字と数字の両方を含めてください")
	}
	return nil
}

// validateCageSize はケージサイズがドメイン上有効かを検証する
