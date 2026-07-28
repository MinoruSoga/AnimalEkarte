package billing

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// completedBillingsCTE は完了会計を期間・クリニックで絞り込む共有 CTE を返す
// （BE-refactor.md E-11: monthly/close で完全同一だった CTE を共有化。列リストのみ
// 呼び出し元ごとに異なる）。プレースホルダは clinicID, status, start, end の順。
func completedBillingsCTE(cols string) string {
	return `WITH completed_billings AS (
		SELECT ` + cols + ` FROM billings
		WHERE clinic_id = ? AND deleted_at IS NULL AND status = ?
		  AND completed_at >= ?
		  AND completed_at < ?
	)`
}

// sumRefundsForCompletedBillings は指定期間内の完了会計に紐づく返金合計を集計する
// （BE-refactor.md E-11: monthly/close で完全同一だった返金集計クエリを共有化）。
func (r *accountingRepository) sumRefundsForCompletedBillings(ctx context.Context, clinicID uint64, start, end time.Time) (int64, error) {
	args := []any{clinicID, model.BillingStatusCompleted, start, end}
	var total int64
	if err := r.db.WithContext(ctx).Raw(
		completedBillingsCTE("id, clinic_id")+`
		SELECT COALESCE(SUM(br.amount), 0)
		FROM billing_refunds br
		JOIN completed_billings cb
		  ON cb.id = br.billing_id
		 AND cb.clinic_id = br.clinic_id
		`, args...).Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// aggregateTaxBreakdown は指定期間内の完了会計を tax_rate 別に集計する（BE-refactor.md E-11:
// monthly/close で期間引数以外同一だった税内訳クエリを共有化。G7-3: sargable な直接比較を使用）。
func (r *accountingRepository) aggregateTaxBreakdown(ctx context.Context, clinicID uint64, start, end time.Time) ([]TaxBreakdownRow, error) {
	var taxRows []TaxBreakdownRow
	if err := r.db.WithContext(ctx).
		Table("billing_items").
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.deleted_at IS NULL").
		Where("billings.clinic_id = ? AND billing_items.deleted_at IS NULL", clinicID).
		Where("billings.status = ?", model.BillingStatusCompleted).
		// G7-3: sargable な直接比較に統一(idx_billings_clinic_completed_at partial index を使えるようにする、
		// CTE本体と同型)。DSN TimeZone=Asia/Tokyo 固定(config.go)のため timestamptz 直接比較と
		// AT TIME ZONE 'Asia/Tokyo' 経由比較は同一瞬間を指し、結果は不変。
		Where("billings.completed_at >= ?", start).
		Where("billings.completed_at < ?", end).
		Select(
			"ROUND(billing_items.tax_rate * 100)::bigint AS tax_rate," +
				" COALESCE(SUM(ROUND(billing_items.unit_price * billing_items.quantity::numeric)), 0) AS taxable_amount," +
				" COALESCE(SUM(ROUND(billing_items.unit_price * billing_items.quantity::numeric * billing_items.tax_rate)), 0) AS tax_amount",
		).
		Group("billing_items.tax_rate").
		Scan(&taxRows).Error; err != nil {
		return nil, err
	}
	return taxRows, nil
}
