package handler

import (
	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/model"
)

// BE9-2C B① transitional alias — 支払方法レスポンスは internal/billing へ移動済み。
// 残留 consumer（cash_register=B⑤）互換のための alias。REMOVE: B⑤移動時。
type paymentMethodResponse = billing.PaymentMethodResponse

func toPaymentMethodResponse(pm *model.PaymentMethodMaster) paymentMethodResponse {
	return billing.ToPaymentMethodResponse(pm)
}
