package billing

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetDailySummary は指定日（JST）の会計完了分を集計する。BUG-368
// payment_splits テーブルを正として集計する（混在会計対応）。
func (r *accountingRepository) GetDailySummary(ctx context.Context, clinicID uint64, date time.Time) (*DailySummaryResult, error) {
	jstDateStart := time.Date(date.In(time.Local).Year(), date.In(time.Local).Month(), date.In(time.Local).Day(), 0, 0, 0, 0, time.Local)
	jstDateEnd := jstDateStart.AddDate(0, 0, 1)

	// 合計件数・売上合計（payment_splits ベース）
	var base struct {
		BillingCount int64
		GrandTotal   int64
	}
	if err := r.db.WithContext(ctx).
		Table("billings").
		Joins(
			"JOIN payment_splits ON payment_splits.billing_id = billings.id"+
				" AND payment_splits.clinic_id = billings.clinic_id",
		).
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billings.status = ?", model.BillingStatusCompleted).
		Where("billings.completed_at >= ? AND billings.completed_at < ?", jstDateStart, jstDateEnd).
		Select("COUNT(DISTINCT billings.id) AS billing_count, COALESCE(SUM(payment_splits.amount), 0) AS grand_total").
		Scan(&base).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing", "")
	}

	// 支払方法別合計（payment_splits.method ベース）
	paymentTotals := make([]PaymentMethodTotal, 0)
	if err := r.db.WithContext(ctx).
		Table("billings").
		Joins(
			"JOIN payment_splits ON payment_splits.billing_id = billings.id"+
				" AND payment_splits.clinic_id = billings.clinic_id",
		).
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billings.status = ?", model.BillingStatusCompleted).
		Where("billings.completed_at >= ? AND billings.completed_at < ?", jstDateStart, jstDateEnd).
		Select("payment_splits.method::text AS method, COALESCE(SUM(payment_splits.amount), 0) AS total").
		Group("payment_splits.method").
		Scan(&paymentTotals).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing", "")
	}

	// 診療区分別合計（billing_items ベース — 変更なし）
	categoryTotals := make([]CategoryTotal, 0)
	if err := r.db.WithContext(ctx).
		Table("billings").
		Joins("JOIN billing_items ON billing_items.billing_id = billings.id AND billing_items.deleted_at IS NULL").
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billings.status = ?", model.BillingStatusCompleted).
		Where("billings.completed_at >= ? AND billings.completed_at < ?", jstDateStart, jstDateEnd).
		Select("billing_items.category::text AS category, COALESCE(SUM(ROUND(billing_items.unit_price * billing_items.quantity::numeric)), 0) AS total").
		Group("billing_items.category").
		Scan(&categoryTotals).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing_item", "")
	}

	return &DailySummaryResult{
		PaymentTotals:  paymentTotals,
		CategoryTotals: categoryTotals,
		BillingCount:   base.BillingCount,
		GrandTotal:     base.GrandTotal,
	}, nil
}
