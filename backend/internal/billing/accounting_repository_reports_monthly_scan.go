package billing

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type monthlyPaymentScanRow struct {
	Date            string
	PaymentMethodID *uint64
	Amount          int64
}

type monthlyCategoryScanRow struct {
	Date     string
	Category string
	Amount   int64
}

type monthlyCountScanRow struct {
	Date  string
	Count int64
}

func (r *accountingRepository) scanMonthlyPaymentRows(
	ctx context.Context,
	completedCTE string,
	args []any,
) ([]monthlyPaymentScanRow, error) {
	var rows []monthlyPaymentScanRow
	if err := r.db.WithContext(ctx).Raw(
		completedCTE+`
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
		`, args...).Scan(&rows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to get monthly payment report")
	}
	return rows, nil
}

func (r *accountingRepository) scanMonthlyCategoryRows(
	ctx context.Context,
	completedCTE string,
	args []any,
) ([]monthlyCategoryScanRow, error) {
	var rows []monthlyCategoryScanRow
	if err := r.db.WithContext(ctx).Raw(
		completedCTE+`
		SELECT
			TO_CHAR(cb.completed_at AT TIME ZONE 'Asia/Tokyo', 'YYYY-MM-DD') AS date,
			bi.category::text AS category,
			COALESCE(SUM(ROUND(bi.unit_price * bi.quantity::numeric)), 0) AS amount
		FROM completed_billings cb
		JOIN billing_items bi ON bi.billing_id = cb.id AND bi.deleted_at IS NULL
		GROUP BY date, bi.category
		ORDER BY date ASC
		`, args...).Scan(&rows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to get monthly category report")
	}
	return rows, nil
}

func (r *accountingRepository) scanMonthlyCountRows(
	ctx context.Context,
	completedCTE string,
	args []any,
) ([]monthlyCountScanRow, error) {
	var rows []monthlyCountScanRow
	if err := r.db.WithContext(ctx).Raw(
		completedCTE+`
		SELECT
			TO_CHAR(cb.completed_at AT TIME ZONE 'Asia/Tokyo', 'YYYY-MM-DD') AS date,
			COUNT(cb.id) AS count
		FROM completed_billings cb
		GROUP BY date
		`, args...).Scan(&rows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to count daily billings for monthly report")
	}
	return rows, nil
}

func (r *accountingRepository) loadMonthlyCloseMaps(
	ctx context.Context,
	clinicID uint64,
	start, end time.Time,
) (map[string]bool, map[string]bool, error) {
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
		return nil, nil, apperrors.Wrap(err, "failed to get cash register closes for monthly report")
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
	return closedAM, closedPM, nil
}
