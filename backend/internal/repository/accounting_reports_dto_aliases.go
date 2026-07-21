package repository

import "github.com/animal-ekarte/backend/internal/billing"

// BE9-2C B④ transitional aliases — 会計レポート DTO は internal/billing へ移動済み。
// 残留 consumer（cash_register=B⑤等）互換のための alias。REMOVE: B⑤移動時。
type (
	DailySummaryResult     = billing.DailySummaryResult
	PaymentMethodTotal     = billing.PaymentMethodTotal
	CategoryTotal          = billing.CategoryTotal
	GetCloseAggregateInput = billing.GetCloseAggregateInput
	CloseAggregateResult   = billing.CloseAggregateResult
	MonthlyReportResult    = billing.MonthlyReportResult
	PaymentAggregateRow    = billing.PaymentAggregateRow
	CategoryAggregateRow   = billing.CategoryAggregateRow
	CloseBillingDetail     = billing.CloseBillingDetail
	TaxBreakdownRow        = billing.TaxBreakdownRow
	MonthlyPaymentRow      = billing.MonthlyPaymentRow
	MonthlyCategoryRow     = billing.MonthlyCategoryRow
	UnpaidOwnerAggregate   = billing.UnpaidOwnerAggregate
	UnpaidSummary          = billing.UnpaidSummary
	OwnerUnpaidBalance     = billing.OwnerUnpaidBalance
	MonthlyUnpaidOwnerPet  = billing.MonthlyUnpaidOwnerPet
	MonthlyUnpaidSummary   = billing.MonthlyUnpaidSummary
	OwnerAnnualRevenue     = billing.OwnerAnnualRevenue
)
