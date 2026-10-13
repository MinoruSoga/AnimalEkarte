package billing

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (r *accountingRepository) scanClosePaymentRows(ctx context.Context, completedCTE string, cArgs []any) ([]PaymentAggregateRow, error) {
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
	for _, row := range pmRows {
		paymentRows = append(paymentRows, PaymentAggregateRow(row))
	}
	return paymentRows, nil
}

func (r *accountingRepository) scanCloseCategoryRows(ctx context.Context, completedCTE string, cArgs []any) ([]CategoryAggregateRow, error) {
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
	return categoryRows, nil
}

func (r *accountingRepository) scanCloseRefundTotal(ctx context.Context, completedCTE string, cArgs []any) (int64, error) {
	var totalRefund int64
	if err := r.db.WithContext(ctx).Raw(
		completedCTE+`
		SELECT COALESCE(SUM(br.amount), 0)
		FROM billing_refunds br
		JOIN completed_billings cb
		  ON cb.id = br.billing_id
		 AND cb.clinic_id = br.clinic_id
		`, cArgs...).Scan(&totalRefund).Error; err != nil {
		return 0, apperrors.Wrap(err, "failed to aggregate refunds for close")
	}
	return totalRefund, nil
}

func (r *accountingRepository) scanCloseBillingDetails(ctx context.Context, input GetCloseAggregateInput) ([]CloseBillingDetailRow, error) {
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
			COALESCE(bc.category, '') AS category,
			ps.payment_method_id,
			ps.amount AS billing_amount,
			COALESCE(rt.refund_amount, 0) AS refund_amount
		FROM completed_billings AS billings
		JOIN payment_splits ps
		  ON ps.billing_id = billings.id
		 AND ps.clinic_id = billings.clinic_id
		-- EMR-354: billing_items を持たない完了済み支払済み会計も明細から落とさない（LEFT JOIN）。
		-- category は従来通り MIN(category) の互換スカラ（なければ空文字）。部門別は category_amounts を参照。
		LEFT JOIN billing_categories bc ON bc.billing_id = billings.id
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
	billingIDs := make([]uint64, 0, len(detailRows))
	seen := make(map[uint64]struct{}, len(detailRows))
	for _, d := range detailRows {
		if _, ok := seen[d.BillingID]; !ok {
			seen[d.BillingID] = struct{}{}
			billingIDs = append(billingIDs, d.BillingID)
		}
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
	if err := r.attachCloseDetailCategoryAmounts(ctx, input.ClinicID, billingIDs, details); err != nil {
		return nil, err
	}
	return details, nil
}

// closeDetailCategoryMethodKey は明細行の部門別 GROSS 按分に使う固定の非空メソッドキー。
// MatrixRowTotals で畳み込むため値自体は表示されないが、AllocateBillingPayments は
// 空キーを捨てるため安定した非空キーが必要（EMR-354）。
const closeDetailCategoryMethodKey = "split"

// attachCloseDetailCategoryAmounts は各行の split 支払額（BillingAmount）を、
// 親会計の billing_items 明細金額重みで按分した部門別 GROSS 内訳を載せる（EMR-354）。
// 既存の AllocateBillingPayments + MatrixRowTotals を再利用し、allocation 数学は変更しない。
// 重みは認証済み clinic かつ distinct billing_id で一括ロードする（N+1 なし）。
// 重みのない会計は既存フォールバックで other に全額が乗る。返金は含まない
// （RefundAmount/NetAmount は従来の全期間スカラのまま）。
func (r *accountingRepository) attachCloseDetailCategoryAmounts(
	ctx context.Context,
	clinicID uint64,
	billingIDs []uint64,
	details []CloseBillingDetailRow,
) error {
	if len(details) == 0 {
		return nil
	}
	weights, err := r.loadCategoryWeightsForBillings(ctx, clinicID, billingIDs)
	if err != nil {
		return err
	}
	weightsByBilling := make(map[uint64]map[string]int64, len(billingIDs))
	for _, w := range weights {
		m, ok := weightsByBilling[w.BillingID]
		if !ok {
			m = make(map[string]int64)
			weightsByBilling[w.BillingID] = m
		}
		m[w.Category] += w.Amount
	}
	for i := range details {
		matrix := AllocateBillingPayments(AllocationBillingInput{
			CategoryWeights: weightsByBilling[details[i].BillingID],
			Payments: []AllocationPayment{
				{MethodKey: closeDetailCategoryMethodKey, Amount: details[i].BillingAmount},
			},
		})
		details[i].CategoryAmounts = MatrixRowTotals(matrix)
	}
	return nil
}

func (r *accountingRepository) scanCloseTaxBreakdown(ctx context.Context, completedCTE string, cArgs []any) ([]TaxBreakdownRow, error) {
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
	return taxBreakdown, nil
}

func (r *accountingRepository) scanCloseUnclassifiedOtherCount(ctx context.Context, completedCTE string, cArgs []any) (int64, error) {
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
		return 0, apperrors.Wrap(err, "failed to count unclassified other accountings for close")
	}
	return unclassifiedOtherCount, nil
}
