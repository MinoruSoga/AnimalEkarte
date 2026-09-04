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
	mArgs := []any{clinicID, model.BillingStatusCompleted, start, end}
	mCompletedCTE := completedBillingsCTE("id, clinic_id, completed_at")

	mPmRows, err := r.scanMonthlyPaymentRows(ctx, mCompletedCTE, mArgs)
	if err != nil {
		return nil, err
	}
	mCatRows, err := r.scanMonthlyCategoryRows(ctx, mCompletedCTE, mArgs)
	if err != nil {
		return nil, err
	}
	mTotalRefund, err := r.sumRefundsForCompletedBillings(ctx, clinicID, start, end)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate monthly refunds")
	}
	mCountRows, err := r.scanMonthlyCountRows(ctx, mCompletedCTE, mArgs)
	if err != nil {
		return nil, err
	}
	closedAM, closedPM, err := r.loadMonthlyCloseMaps(ctx, clinicID, start, end)
	if err != nil {
		return nil, err
	}

	payRows := make([]MonthlyPaymentRow, 0, len(mPmRows))
	var grandTotal int64
	for _, row := range mPmRows {
		payRows = append(payRows, MonthlyPaymentRow(row))
		grandTotal += row.Amount
	}

	catRows := make([]MonthlyCategoryRow, 0, len(mCatRows))
	for _, row := range mCatRows {
		catRows = append(catRows, MonthlyCategoryRow(row))
	}

	dailyBillingCount := make(map[string]int64, len(mCountRows))
	for _, row := range mCountRows {
		dailyBillingCount[row.Date] = row.Count
	}

	taxBreakdown, err := r.aggregateTaxBreakdown(ctx, clinicID, start, end)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate tax breakdown for monthly report")
	}

	var billingCount int64
	if err := r.db.WithContext(ctx).
		Model(&model.Billing{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("status = ?", model.BillingStatusCompleted).
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
