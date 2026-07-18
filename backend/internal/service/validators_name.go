package service

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func validateRequiredName(name string) error {
	if strings.TrimSpace(name) == "" {
		return apperrors.WrapInvalidInput("名前を入力してください")
	}
	if utf8.RuneCountInString(name) > MasterNameMaxLength {
		return apperrors.WrapInvalidInput(fmt.Sprintf("名前は%d文字以内で入力してください", MasterNameMaxLength))
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

// validateOptionalName は nil 許容の名称バリデーション。
// PATCH 系で nil の場合は更新しない意味なのでスキップ、非 nil の場合のみ検証する。
func validateOptionalName(name *string) error {
	if name == nil {
		return nil
	}
	return validateRequiredName(*name)
}

// validateDiscountRate は割引率が 0〜100 の範囲内かを検証する
