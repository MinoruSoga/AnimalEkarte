package billing

// validators_insurance.go — BE9-2C B①: service/validators_master.go から保険補償率検証を移動
// （billing domain 固有・service 側は accounting 残留 consumer 向け delegate）。

import (
	"github.com/animal-ekarte/backend/internal/apperrors"
)

// validateNonNegativePrice は価格フィールドが 0 以上かを検証する（nil の場合はスキップ）(BUG-380)
func ValidateCoverageRate(rate int) error {
	if rate < 0 || rate > 100 {
		return apperrors.WrapInvalidInput("補償率は0〜100の範囲で入力してください")
	}
	return nil
}

// validateOptionalCoverageRate は nil 許容の保険補償率バリデーション (BUG-398)
func ValidateOptionalCoverageRate(rate *int) error {
	if rate == nil {
		return nil
	}
	return ValidateCoverageRate(*rate)
}
