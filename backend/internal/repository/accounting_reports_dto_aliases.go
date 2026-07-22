package repository

import "github.com/animal-ekarte/backend/internal/billing"

// BE9-2C B④ transitional aliases — 会計レポート DTO は internal/billing へ移動済み。
// L⑥時点の実残留 consumer は internal/service/mocks_accounting_test.go。LSTEP consumer
// ではないため本sliceでは維持し、owner/pet等のservice test整理（BE9-2E）または最終
// compatibility cleanup（BE9-2F）でconsumerと一緒に削除する。
type (
	DailySummaryResult     = billing.DailySummaryResult
	PaymentMethodTotal     = billing.PaymentMethodTotal
	CategoryTotal          = billing.CategoryTotal
	GetCloseAggregateInput = billing.GetCloseAggregateInput
	CloseAggregateResult   = billing.CloseAggregateResult
	MonthlyReportResult    = billing.MonthlyReportResult
	PaymentAggregateRow    = billing.PaymentAggregateRow
	CategoryAggregateRow   = billing.CategoryAggregateRow
	CloseBillingDetailRow  = billing.CloseBillingDetailRow
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
