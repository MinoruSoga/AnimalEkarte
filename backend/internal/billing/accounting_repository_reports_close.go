package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
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
	// 集計行: カテゴリ×支払方法別の純売上
	// payment_splits を正として使い、Cartesian 積バグを回避する。
	// カテゴリは billing_items から1会計1行に集約し、payment_splits と billing_id で結合する。
	// Cartesian 積を避けるため payment_splits / billing_items を別クエリで集計する
	cArgs := []any{input.ClinicID, model.BillingStatusCompleted, input.PeriodStart.In(time.Local), input.PeriodEnd.In(time.Local)}
	// SEC-SWEEP-02-BILL-B1a: use pet-correlated CTE for all close aggregates (not shared
	// completedBillingsCTE, which has no pets parent join).
	completedCTE := completedBillingsPetClinicCTE("billings.id, billings.clinic_id")

	// Query 1: 支払方法別合計 (payment_splits のみ)
	type pmRow struct {
		PaymentMethodID *uint64
		Amount          int64
	}
	var pmRows []pmRow
	if err := r.db.WithContext(ctx).Raw(
		completedCTE+`
		SELECT ps.payment_method_id, COALESCE(SUM(ps.amount), 0) AS amount
		FROM payment_splits ps
		JOIN completed_billings cb
		  ON cb.id = ps.billing_id
		 AND cb.clinic_id = ps.clinic_id
		GROUP BY ps.payment_method_id
		`, cArgs...).Scan(&pmRows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate payment splits for close")
	}
	paymentRows := make([]PaymentAggregateRow, 0, len(pmRows))
	for _, r := range pmRows {
		paymentRows = append(paymentRows, PaymentAggregateRow(r))
	}

	// Query 2: カテゴリ別合計 (billing_items のみ)
	var catRows []closeCategoryRow
	if err := r.db.WithContext(ctx).Raw(
		completedCTE+`
		SELECT bi.category::text AS category,
		       COALESCE(SUM(ROUND(bi.unit_price * bi.quantity::numeric)), 0) AS amount
		FROM billing_items bi
		WHERE bi.billing_id IN (SELECT id FROM completed_billings) AND bi.deleted_at IS NULL
		GROUP BY bi.category
		`, cArgs...).Scan(&catRows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate categories for close")
	}
	categoryRows, err := toCategoryAggregateRows(catRows)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to validate category aggregate for close")
	}

	// Query 3: 返金合計（pet-correlated completed set）
	var totalRefund int64
	if err := r.db.WithContext(ctx).Raw(
		completedCTE+`
		SELECT COALESCE(SUM(br.amount), 0)
		FROM billing_refunds br
		JOIN completed_billings cb
		  ON cb.id = br.billing_id
		 AND cb.clinic_id = br.clinic_id
		`, cArgs...).Scan(&totalRefund).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate refunds for close")
	}

	// 個別会計一覧（payment_splits ベース: 混在支払いは複数行）
	type detailRow struct {
		BillingID         uint64
		PaidAt            time.Time
		OwnerName         string
		PetName           string
		HospitalizationID *uint64
		Category          string
		PaymentMethodID   *uint64
		BillingAmount     int64
		RefundAmount      int64
	}
	var detailRows []detailRow
	// Parent pets clinic correlation (SEC-SWEEP-02-BILL-B1a-FIX): CTE keeps display
	// LEFT JOIN for name blanking without row exclusion; outer LEFT JOIN uses table
	// alias `billings` so lint sees pets.id=billings.pet_id + pets.clinic_id=
	// billings.clinic_id. No pets.deleted_at / deceased_at on the parent correlation.
	if err := r.db.WithContext(ctx).Raw(
		completedBillingsPetClinicCTE("billings.id, billings.clinic_id, billings.completed_at, billings.owner_id, billings.pet_id, billings.hospitalization_id")+`,
		refund_totals AS (
			SELECT br.billing_id, COALESCE(SUM(br.amount), 0) AS refund_amount
			FROM billing_refunds br
			JOIN completed_billings cb
			  ON cb.id = br.billing_id
			 AND cb.clinic_id = br.clinic_id
			GROUP BY br.billing_id
		),
		billing_categories AS (
			SELECT billing_id, MIN(category::text) AS category
			FROM billing_items
			WHERE billing_id IN (SELECT id FROM completed_billings) AND deleted_at IS NULL
			GROUP BY billing_id
		)
		SELECT
			billings.id AS billing_id,
			billings.completed_at AS paid_at,
			COALESCE(o.name, '') AS owner_name,
			COALESCE(p.name, '') AS pet_name,
			billings.hospitalization_id,
			bc.category,
			ps.payment_method_id,
			ps.amount AS billing_amount,
			COALESCE(rt.refund_amount, 0) AS refund_amount
		FROM completed_billings AS billings
		JOIN payment_splits ps
		  ON ps.billing_id = billings.id
		 AND ps.clinic_id = billings.clinic_id
		JOIN billing_categories bc ON bc.billing_id = billings.id
		LEFT JOIN owners o
		  ON o.id = billings.owner_id
		 AND o.clinic_id = billings.clinic_id
		 AND o.deleted_at IS NULL
		LEFT JOIN pets p
		  ON p.id = billings.pet_id
		 AND p.clinic_id = billings.clinic_id
		LEFT JOIN refund_totals rt ON rt.billing_id = billings.id
		ORDER BY billings.completed_at ASC
	`, input.ClinicID, model.BillingStatusCompleted, input.PeriodStart.In(time.Local), input.PeriodEnd.In(time.Local)).
		Scan(&detailRows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to get billing details for close")
	}

	details := make([]CloseBillingDetailRow, 0, len(detailRows))
	for _, d := range detailRows {
		details = append(details, CloseBillingDetailRow{
			BillingID:         d.BillingID,
			PaidAt:            d.PaidAt,
			OwnerName:         d.OwnerName,
			PetName:           d.PetName,
			IsHospitalization: d.HospitalizationID != nil,
			Category:          d.Category,
			PaymentMethodID:   d.PaymentMethodID,
			BillingAmount:     d.BillingAmount,
			RefundAmount:      d.RefundAmount,
			NetAmount:         d.BillingAmount - d.RefundAmount,
		})
	}

	// 税率別集計（pet-correlated completed set — Cartesian 積なし）
	var taxBreakdown []TaxBreakdownRow
	if err := r.db.WithContext(ctx).Raw(
		completedCTE+`
		SELECT
			ROUND(bi.tax_rate * 100)::bigint AS tax_rate,
			COALESCE(SUM(ROUND(bi.unit_price * bi.quantity::numeric)), 0) AS taxable_amount,
			COALESCE(SUM(ROUND(bi.unit_price * bi.quantity::numeric * bi.tax_rate)), 0) AS tax_amount
		FROM billing_items bi
		JOIN completed_billings cb ON cb.id = bi.billing_id
		WHERE bi.deleted_at IS NULL
		GROUP BY bi.tax_rate
		`, cArgs...).Scan(&taxBreakdown).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate tax breakdown for close")
	}

	// DEC-40: 未分類・要確認件数 = category=other 明細を1件以上持つ会計の distinct 数。
	// 既存 detail の MIN(category) / payment_splits 展開とは独立に集計し、混在会計の欠落と
	// 明細行・split 行の過大計上を避ける（detail/report 全体の再設計はしない）。
	var unclassifiedOtherCount int64
	if err := r.db.WithContext(ctx).Raw(
		completedCTE+`
		SELECT COUNT(DISTINCT bi.billing_id)
		FROM billing_items bi
		WHERE bi.billing_id IN (SELECT id FROM completed_billings)
		  AND bi.deleted_at IS NULL
		  AND bi.category = ?
		`, append(append([]any{}, cArgs...), model.ItemCategoryOther)...).
		Scan(&unclassifiedOtherCount).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to count unclassified other accountings for close")
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
