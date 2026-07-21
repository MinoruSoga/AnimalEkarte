package billing

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetCloseAggregate は指定期間内の会計を集計する。FEAT-368
// payment_splits を正として集計（SUM(DISTINCT) hack を除去）。
// カテゴリ別集計: billing_items を CTE で per-billing 合算し、payment_splits と JOIN して按分なしで集計。
func (r *accountingRepository) GetCloseAggregate(ctx context.Context, input GetCloseAggregateInput) (*CloseAggregateResult, error) {
	// 集計行: カテゴリ×支払方法別の純売上
	// payment_splits を正として使い、Cartesian 積バグを回避する。
	// カテゴリは billing_items から1会計1行に集約し、payment_splits と billing_id で結合する。
	// Cartesian 積を避けるため payment_splits / billing_items を別クエリで集計する
	cArgs := []any{input.ClinicID, model.BillingStatusCompleted, input.PeriodStart.In(time.Local), input.PeriodEnd.In(time.Local)}
	completedCTE := completedBillingsCTE("id")

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
		WHERE ps.billing_id IN (SELECT id FROM completed_billings)
		GROUP BY ps.payment_method_id
		`, cArgs...).Scan(&pmRows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate payment splits for close")
	}
	paymentRows := make([]PaymentAggregateRow, 0, len(pmRows))
	for _, r := range pmRows {
		paymentRows = append(paymentRows, PaymentAggregateRow(r))
	}

	// Query 2: カテゴリ別合計 (billing_items のみ)
	type catRow struct {
		Category string
		Amount   int64
	}
	var catRows []catRow
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
	categoryRows := make([]CategoryAggregateRow, 0, len(catRows))
	for _, r := range catRows {
		categoryRows = append(categoryRows, CategoryAggregateRow(r))
	}

	// Query 3: 返金合計
	totalRefund, err := r.sumRefundsForCompletedBillings(ctx, input.ClinicID, input.PeriodStart.In(time.Local), input.PeriodEnd.In(time.Local))
	if err != nil {
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
	if err := r.db.WithContext(ctx).Raw(`
		WITH completed_billings AS (
			SELECT id, completed_at, owner_id, pet_id, hospitalization_id
			FROM billings
			WHERE clinic_id = ? AND deleted_at IS NULL AND status = ?
			  -- G7-3: sargable な直接比較に統一(CTE本体と同型)。DSN TimeZone=Asia/Tokyo 固定(config.go)のため
			  -- timestamptz 直接比較と AT TIME ZONE 'Asia/Tokyo' 経由比較は同一瞬間を指し、結果は不変。
			  AND completed_at >= ?
			  AND completed_at < ?
		),
		refund_totals AS (
			SELECT billing_id, COALESCE(SUM(amount), 0) AS refund_amount
			FROM billing_refunds
			WHERE billing_id IN (SELECT id FROM completed_billings)
			GROUP BY billing_id
		),
		billing_categories AS (
			SELECT billing_id, MIN(category::text) AS category
			FROM billing_items
			WHERE billing_id IN (SELECT id FROM completed_billings) AND deleted_at IS NULL
			GROUP BY billing_id
		)
		SELECT
			cb.id AS billing_id,
			cb.completed_at AS paid_at,
			COALESCE(o.name, '') AS owner_name,
			COALESCE(p.name, '') AS pet_name,
			cb.hospitalization_id,
			bc.category,
			ps.payment_method_id,
			ps.amount AS billing_amount,
			COALESCE(rt.refund_amount, 0) AS refund_amount
		FROM completed_billings cb
		JOIN payment_splits ps ON ps.billing_id = cb.id
		JOIN billing_categories bc ON bc.billing_id = cb.id
		LEFT JOIN owners o ON o.id = cb.owner_id AND o.deleted_at IS NULL
		LEFT JOIN pets p ON p.id = cb.pet_id AND p.deleted_at IS NULL
		LEFT JOIN refund_totals rt ON rt.billing_id = cb.id
		ORDER BY cb.completed_at ASC
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

	// 税率別集計（billing_items ベース — Cartesian 積なし: payments JOIN 除去済み）
	taxBreakdown, err := r.aggregateTaxBreakdown(ctx, input.ClinicID, input.PeriodStart.In(time.Local), input.PeriodEnd.In(time.Local))
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate tax breakdown for close")
	}

	return &CloseAggregateResult{
		PaymentRows:    paymentRows,
		CategoryRows:   categoryRows,
		TotalRefund:    totalRefund,
		BillingDetails: details,
		TaxBreakdown:   taxBreakdown,
	}, nil
}
