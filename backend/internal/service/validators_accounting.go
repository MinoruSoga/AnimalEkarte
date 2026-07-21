package service

import (
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func validateTaxType(taxType string) error {
	return sharedkernel.ValidateTaxType(taxType)
}

// validateCageType はケージ種別がドメイン上有効かを検証する
func validateNonNegativePrice(price *int64) error {
	return sharedkernel.ValidateNonNegativePrice(price)
}

// validateCoverageRate は保険補償率が 0〜100 の範囲内かを検証する (BUG-398)
