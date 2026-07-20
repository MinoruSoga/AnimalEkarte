package service

import (
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func validateRequiredName(name string) error {
	return sharedkernel.ValidateRequiredName(name)
}

// validateOptionalName は nil 許容の名称バリデーション。
// PATCH 系で nil の場合は更新しない意味なのでスキップ、非 nil の場合のみ検証する。
func validateOptionalName(name *string) error {
	return sharedkernel.ValidateOptionalName(name)
}

// validateDiscountRate は割引率が 0〜100 の範囲内かを検証する
