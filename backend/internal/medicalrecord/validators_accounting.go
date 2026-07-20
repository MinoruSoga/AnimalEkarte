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

// errMsgQuantityPositive mirrors internal/service.ErrMsgQuantityPositive (validators.go) —
// treatmentService (BE9-2D ④b) 用の複製。原本は internal/service の残存 service 群が共有し続ける
// ため移動しない（validateNonNegativePrice と同方針。errMsgAtLeastOneField は validators.go に既存）。
const errMsgQuantityPositive = "数量は0より大きい値を入力してください"

// validateDiscountRate is a documented, byte-for-byte duplicate of
// internal/service.validateDiscountRate (validators_owner.go). It stays duplicated because
// internal/service's owner service still shares it; this copy serves treatmentService only
// (BE9-2D ④b). Pure, stateless — collapse onto the shared kernel package when it exists.
func validateDiscountRate(rate float64) error {
	if rate < 0 || rate > 100 {
		return apperrors.WrapInvalidInput("割引率は0〜100の範囲で入力してください")
	}
	return nil
}
