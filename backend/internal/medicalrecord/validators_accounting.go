package medicalrecord

import "github.com/animal-ekarte/backend/internal/apperrors"

// errMsgPriceZeroOrMore mirrors internal/service.ErrMsgPriceZeroOrMore (validators.go, BUG-380).
const errMsgPriceZeroOrMore = "金額は0以上を入力してください"

// validateNonNegativePrice is a documented, byte-for-byte duplicate of
// internal/service.validateNonNegativePrice (validators_accounting.go). It stays duplicated
// (not moved) because internal/service's procedure/billing/merchandise/treatment services still
// share it; this copy serves vaccineService only. Pure, stateless — no interface warranted (see
// validators.go rationale). Follow-up: collapse both copies onto a shared package once justified.
func validateNonNegativePrice(price *int64) error {
	if price == nil {
		return nil
	}
	if *price < 0 {
		return apperrors.WrapInvalidInput(errMsgPriceZeroOrMore)
	}
	return nil
}
