package medicalrecord

import "github.com/animal-ekarte/backend/internal/sharedkernel"

// 共有カーネル昇格batch: ①/④b 以来の internal/service との複製（validateNonNegativePrice/
// validateDiscountRate/errMsg*）は sharedkernel へ一本化。本ファイルは呼び出し面互換 delegate のみ。

const errMsgPriceZeroOrMore = sharedkernel.ErrMsgPriceZeroOrMore
const errMsgQuantityPositive = sharedkernel.ErrMsgQuantityPositive

func validateNonNegativePrice(price *int64) error {
	return sharedkernel.ValidateNonNegativePrice(price)
}

func validateDiscountRate(rate float64) error {
	return sharedkernel.ValidateDiscountRate(rate)
}
