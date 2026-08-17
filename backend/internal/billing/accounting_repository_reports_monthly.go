package billing

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// GetMonthlyReport は指定年月の月次売上レポートを集計する。FEAT-368
// payment_splits を正として集計（SUM(DISTINCT) hack を除去）。
func (r *accountingRepository) GetMonthlyReport(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResult, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	return r.GetMonthlyReportByPeriod(ctx, clinicID, start, end)
}

// GetMonthlyReportByPeriod は指定期間（start 以上 end 未満）の売上レポートを集計する。
func (r *accountingRepository) GetMonthlyReportByPeriod(ctx context.Context, clinicID uint64, start, end time.Time) (*MonthlyReportResult, error) {
	// Cartesian 積を避けるため payment_splits / billing_items を別クエリで集計する
	mArgs := []any{clinicID, model.BillingStatusCompleted, start, end}
	mCompletedCTE := completedBillingsCTE("id, clinic_id, completed_at")

	// Query 1: 日×支払方法別合計 (payment_splits のみ)
	type mPmRow struct {
		Date            string
		PaymentMethodID *uint64
		Amount          int64
	}
	var mPmRows []mPmRow
	if err := r.db.WithContext(ctx).Raw(
		mCompletedCTE+`
		SELECT
			TO_CHAR(cb.completed_at AT TIME ZONE 'Asia/Tokyo', 'YYYY-MM-DD') AS date,
			ps.payment_method_id,
			COALESCE(SUM(ps.amount), 0) AS amount
		FROM completed_billings cb
		JOIN payment_splits ps
		  ON ps.billing_id = cb.id
		 AND ps.clinic_id = cb.clinic_id
		GROUP BY date, ps.payment_method_id
		ORDER BY date ASC
		`, mArgs...).Scan(&mPmRows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to get monthly payment report")
	}

	// Query 2: 日×カテゴリ別合計 (billing_items のみ)
	type mCatRow struct {
		Date     string
		Category string
		Amount   int64
	}
	var mCatRows []mCatRow
	if err := r.db.WithContext(ctx).Raw(
		mCompletedCTE+`
		SELECT
			TO_CHAR(cb.completed_at AT TIME ZONE 'Asia/Tokyo', 'YYYY-MM-DD') AS date,
			bi.category::text AS category,
			COALESCE(SUM(ROUND(bi.unit_price * bi.quantity::numeric)), 0) AS amount
		FROM completed_billings cb
		JOIN billing_items bi ON bi.billing_id = cb.id AND bi.deleted_at IS NULL
		GROUP BY date, bi.category
		ORDER BY date ASC
		`, mArgs...).Scan(&mCatRows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to get monthly category report")
	}

	// Query 3: 返金合計
	mTotalRefund, err := r.sumRefundsForCompletedBillings(ctx, clinicID, start, end)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate monthly refunds")
	}

	// Query 4: 日別会計件数
	type mCountRow struct {
		Date  string
		Count int64
	}
	var mCountRows []mCountRow
	if err := r.db.WithContext(ctx).Raw(
		mCompletedCTE+`
		SELECT
			TO_CHAR(cb.completed_at AT TIME ZONE 'Asia/Tokyo', 'YYYY-MM-DD') AS date,
			COUNT(cb.id) AS count
		FROM completed_billings cb
		GROUP BY date
		`, mArgs...).Scan(&mCountRows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to count daily billings for monthly report")
	}

	// 締めレコード取得（日付→AM/PM 締め状態マップ）
	type closeRow struct {
		CloseDate string
		Period    string
	}
	var closeRows []closeRow
	if err := r.db.WithContext(ctx).
		Table("cash_register_closes").
		Where("clinic_id = ?", clinicID).
		Where("close_date >= ? AND close_date < ?", start.Format(time.DateOnly), end.Format(time.DateOnly)).
		Select("close_date::text AS close_date, period").
		Scan(&closeRows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to get cash register closes for monthly report")
	}
	closedAM := make(map[string]bool)
	closedPM := make(map[string]bool)
	for _, cr := range closeRows {
		if cr.Period == "am" {
			closedAM[cr.CloseDate] = true
		} else {
			closedPM[cr.CloseDate] = true
		}
	}

	// 集計結果を構造体に変換
	payRows := make([]MonthlyPaymentRow, 0, len(mPmRows))
	var grandTotal int64
	for _, r := range mPmRows {
		payRows = append(payRows, MonthlyPaymentRow(r))
		grandTotal += r.Amount
	}

	catRows := make([]MonthlyCategoryRow, 0, len(mCatRows))
	for _, r := range mCatRows {
		catRows = append(catRows, MonthlyCategoryRow(r))
	}

	dailyBillingCount := make(map[string]int64, len(mCountRows))
	for _, r := range mCountRows {
		dailyBillingCount[r.Date] = r.Count
	}

	// 税率別集計
	taxBreakdown, err := r.aggregateTaxBreakdown(ctx, clinicID, start, end)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate tax breakdown for monthly report")
	}

	// 会計件数（billings 単位）
	var billingCount int64
	if err := r.db.WithContext(ctx).
		Model(&model.Billing{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("status = ?", model.BillingStatusCompleted).
		// G7-3: sargable な直接比較に統一(idx_billings_clinic_completed_at partial index を使えるようにする)。
		Where("completed_at >= ?", start).
		Where("completed_at < ?", end).
		Count(&billingCount).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to count monthly billings")
	}

	return &MonthlyReportResult{
		PaymentRows:       payRows,
		CategoryRows:      catRows,
		DailyBillingCount: dailyBillingCount,
		ClosedAM:          closedAM,
		ClosedPM:          closedPM,
		GrandTotal:        grandTotal - mTotalRefund,
		TotalRefund:       mTotalRefund,
		BillingCount:      billingCount,
		TaxBreakdown:      taxBreakdown,
	}, nil
}
