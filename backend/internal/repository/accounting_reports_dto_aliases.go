package repository

import "github.com/animal-ekarte/backend/internal/billing"

// BE9-2C B④ transitional aliases — 会計レポート DTO は internal/billing へ移動済み。
// B⑤完了時点の実残留 consumer は internal/service の lstep 系テスト
// （lstep_tag_sync_visit_{cpm,ltv}_test.go / perf_n1_regression_test.go / mocks_accounting_test.go
// が mockAccountingRepository 経由で参照）のみ。REMOVE: lstep domain 移行時。
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
