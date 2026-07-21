package service

import "github.com/animal-ekarte/backend/internal/billing"

// BE9-2C B① transitional aliases — campaign discount 系の型は internal/billing へ移動済み。
// 残留 consumer（billing_item=B③）互換のための alias。REMOVE: B③移動時。
type (
	DiscountSuggestion = billing.DiscountSuggestion
	billingAuditEntry  = billing.AuditEntry
)
