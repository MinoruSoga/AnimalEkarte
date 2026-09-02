package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type closeCategoryRow struct {
	Category string
	Amount   int64
}

func validateCloseAggregateCategory(category string) error {
	for _, allowed := range model.AllItemCategories() {
		if category == string(allowed) {
			return nil
		}
	}

	return fmt.Errorf("unknown item category in close aggregate: %q", category)
}

func toCategoryAggregateRows(rows []closeCategoryRow) ([]CategoryAggregateRow, error) {
	for _, row := range rows {
		if err := validateCloseAggregateCategory(row.Category); err != nil {
			return nil, err
		}
	}

	result := make([]CategoryAggregateRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, CategoryAggregateRow(row))
	}
	return result, nil
}

// completedBillingsPetClinicCTE selects completed billings in the close period.
// Parent pets use clinic-correlated LEFT JOIN for lint parent correlation only;
// rows are never filtered out by pet_id parent clinic (SEC-SWEEP-02-BILL-B1a-FIX).
// Foreign pet/owner name blanking is done in the details query outer joins.
// No pets.deleted_at / deceased_at on correlation (historical close totals must stay).
// cols must be billings-qualified (JOIN pets would make bare id/clinic_id ambiguous).
func completedBillingsPetClinicCTE(cols string) string {
	return `WITH completed_billings AS (
		SELECT ` + cols + `
		FROM billings
		LEFT JOIN pets ON pets.id = billings.pet_id AND pets.clinic_id = billings.clinic_id
		WHERE billings.clinic_id = ? AND billings.deleted_at IS NULL AND billings.status = ?
		  AND billings.completed_at >= ?
		  AND billings.completed_at < ?
	)`
}

// GetCloseAggregate は指定期間内の会計を集計する。FEAT-368
// payment_splits を正として集計（SUM(DISTINCT) hack を除去）。
// カテゴリ別集計: billing_items を CTE で per-billing 合算し、payment_splits と JOIN して按分なしで集計。
func (r *accountingRepository) GetCloseAggregate(ctx context.Context, input GetCloseAggregateInput) (*CloseAggregateResult, error) {
	cArgs := []any{input.ClinicID, model.BillingStatusCompleted, input.PeriodStart.In(time.Local), input.PeriodEnd.In(time.Local)}
	completedCTE := completedBillingsPetClinicCTE("billings.id, billings.clinic_id")

	paymentRows, err := r.scanClosePaymentRows(ctx, completedCTE, cArgs)
	if err != nil {
		return nil, err
	}
	categoryRows, err := r.scanCloseCategoryRows(ctx, completedCTE, cArgs)
	if err != nil {
		return nil, err
	}
	totalRefund, err := r.scanCloseRefundTotal(ctx, completedCTE, cArgs)
	if err != nil {
		return nil, err
	}
	details, err := r.scanCloseBillingDetails(ctx, input)
	if err != nil {
		return nil, err
	}
	taxBreakdown, err := r.scanCloseTaxBreakdown(ctx, completedCTE, cArgs)
	if err != nil {
		return nil, err
	}
	unclassifiedOtherCount, err := r.scanCloseUnclassifiedOtherCount(ctx, completedCTE, cArgs)
	if err != nil {
		return nil, err
	}

	return &CloseAggregateResult{
		PaymentRows:            paymentRows,
		CategoryRows:           categoryRows,
		TotalRefund:            totalRefund,
		BillingDetails:         details,
		TaxBreakdown:           taxBreakdown,
		UnclassifiedOtherCount: unclassifiedOtherCount,
	}, nil
}
