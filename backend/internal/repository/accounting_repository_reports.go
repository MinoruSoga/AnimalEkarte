package repository

import (
	"context"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// PaymentMethodTotal は支払方法別の売上合計。BUG-368
type PaymentMethodTotal struct {
	Method string `json:"method"`
	Total  int64  `json:"total"`
}

// CategoryTotal は診療区分別の売上合計。BUG-368
type CategoryTotal struct {
	Category string `json:"category"`
	Total    int64  `json:"total"`
}

// DailySummaryResult はレジ締め日次集計の結果。BUG-368
type DailySummaryResult struct {
	PaymentTotals  []PaymentMethodTotal `json:"payment_totals"`
	CategoryTotals []CategoryTotal      `json:"category_totals"`
	BillingCount   int64                `json:"billing_count"`
	GrandTotal     int64                `json:"grand_total"`
}

// GetDailySummary は指定日（JST）の会計完了分を集計する。BUG-368
// payment_splits テーブルを正として集計する（混在会計対応）。
func (r *accountingRepository) GetDailySummary(ctx context.Context, clinicID uint64, date time.Time) (*DailySummaryResult, error) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	jstDate := date.In(jst).Format("2006-01-02")

	// 合計件数・売上合計（payment_splits ベース）
	var base struct {
		BillingCount int64
		GrandTotal   int64
	}
	if err := r.db.WithContext(ctx).
		Table("billings").
		Joins("JOIN payment_splits ON payment_splits.billing_id = billings.id").
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billings.status = ?", model.BillingStatusCompleted).
		Where("DATE(billings.completed_at AT TIME ZONE 'Asia/Tokyo') = ?", jstDate).
		Select("COUNT(DISTINCT billings.id) AS billing_count, COALESCE(SUM(payment_splits.amount), 0) AS grand_total").
		Scan(&base).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing", "")
	}

	// 支払方法別合計（payment_splits.method ベース）
	paymentTotals := make([]PaymentMethodTotal, 0)
	if err := r.db.WithContext(ctx).
		Table("billings").
		Joins("JOIN payment_splits ON payment_splits.billing_id = billings.id").
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billings.status = ?", model.BillingStatusCompleted).
		Where("DATE(billings.completed_at AT TIME ZONE 'Asia/Tokyo') = ?", jstDate).
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
		Where("DATE(billings.completed_at AT TIME ZONE 'Asia/Tokyo') = ?", jstDate).
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

// ---- FEAT-368: 集計・締め機能 ----

// GetCloseAggregateInput は締めプレビュー集計の入力パラメータ
type GetCloseAggregateInput struct {
	ClinicID    uint64
	PeriodStart time.Time // JST timestamptz
	PeriodEnd   time.Time // JST timestamptz
}

// PaymentAggregateRow は支払方法別の集計結果1行
type PaymentAggregateRow struct {
	PaymentMethodID *uint64
	Amount          int64
}

// CategoryAggregateRow はカテゴリ別の集計結果1行
type CategoryAggregateRow struct {
	Category string
	Amount   int64
}

// CloseBillingDetail は個別会計一覧の1レコード
type CloseBillingDetail struct {
	BillingID         uint64    `json:"billing_id"`
	PaidAt            time.Time `json:"paid_at"`
	OwnerName         string    `json:"owner_name"`
	PetName           string    `json:"pet_name"`
	IsHospitalization bool      `json:"is_hospitalization"`
	Category          string    `json:"category"`
	PaymentMethodID   *uint64   `json:"payment_method_id,omitempty"`
	BillingAmount     int64     `json:"billing_amount"`
	RefundAmount      int64     `json:"refund_amount"`
	NetAmount         int64     `json:"net_amount"`
}

// TaxBreakdownRow は税率別の集計結果1行
type TaxBreakdownRow struct {
	TaxRate       int64 // 整数パーセント: 10 (標準) or 8 (軽減)
	TaxableAmount int64
	TaxAmount     int64
}

// CloseAggregateResult は締めプレビュー集計の結果
type CloseAggregateResult struct {
	PaymentRows    []PaymentAggregateRow
	CategoryRows   []CategoryAggregateRow
	TotalRefund    int64
	BillingDetails []CloseBillingDetail
	TaxBreakdown   []TaxBreakdownRow `json:"-"`
}

// MonthlyPaymentRow は月次レポートの日×支払方法別集計行
type MonthlyPaymentRow struct {
	Date            string
	PaymentMethodID *uint64
	Amount          int64
}

// MonthlyCategoryRow は月次レポートの日×カテゴリ別集計行
type MonthlyCategoryRow struct {
	Date     string
	Category string
	Amount   int64
}

// MonthlyReportResult は月次売上レポートの結果
type MonthlyReportResult struct {
	PaymentRows       []MonthlyPaymentRow
	CategoryRows      []MonthlyCategoryRow
	DailyBillingCount map[string]int64
	ClosedAM          map[string]bool
	ClosedPM          map[string]bool
	GrandTotal        int64
	TotalRefund       int64
	BillingCount      int64
	TaxBreakdown      []TaxBreakdownRow `json:"-"`
}

// GetCloseAggregate は指定期間内の会計を集計する。FEAT-368
// payment_splits を正として集計（SUM(DISTINCT) hack を除去）。
// カテゴリ別集計: billing_items を CTE で per-billing 合算し、payment_splits と JOIN して按分なしで集計。
func (r *accountingRepository) GetCloseAggregate(ctx context.Context, input GetCloseAggregateInput) (*CloseAggregateResult, error) {
	// 集計行: カテゴリ×支払方法別の純売上
	// payment_splits を正として使い、Cartesian 積バグを回避する。
	// カテゴリは billing_items から1会計1行に集約し、payment_splits と billing_id で結合する。
	// Cartesian 積を避けるため payment_splits / billing_items を別クエリで集計する
	cArgs := []any{input.ClinicID, model.BillingStatusCompleted, input.PeriodStart, input.PeriodEnd}
	completedCTE := `WITH completed_billings AS (
		SELECT id FROM billings
		WHERE clinic_id = ? AND deleted_at IS NULL AND status = ?
		  AND completed_at AT TIME ZONE 'Asia/Tokyo' >= ?
		  AND completed_at AT TIME ZONE 'Asia/Tokyo' < ?
	)`

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
	var totalRefund int64
	if err := r.db.WithContext(ctx).Raw(
		completedCTE+`
		SELECT COALESCE(SUM(br.amount), 0)
		FROM billing_refunds br
		WHERE br.billing_id IN (SELECT id FROM completed_billings)
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
	if err := r.db.WithContext(ctx).Raw(`
		WITH completed_billings AS (
			SELECT id, completed_at, owner_id, pet_id, hospitalization_id
			FROM billings
			WHERE clinic_id = ? AND deleted_at IS NULL AND status = ?
			  AND completed_at AT TIME ZONE 'Asia/Tokyo' >= ?
			  AND completed_at AT TIME ZONE 'Asia/Tokyo' < ?
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
	`, input.ClinicID, model.BillingStatusCompleted, input.PeriodStart, input.PeriodEnd).
		Scan(&detailRows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to get billing details for close")
	}

	details := make([]CloseBillingDetail, 0, len(detailRows))
	for _, d := range detailRows {
		details = append(details, CloseBillingDetail{
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
	type taxRow struct {
		TaxRate       int64
		TaxableAmount int64
		TaxAmount     int64
	}
	var taxRows []taxRow
	if err := r.db.WithContext(ctx).
		Table("billing_items").
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.deleted_at IS NULL").
		Where("billings.clinic_id = ? AND billing_items.deleted_at IS NULL", input.ClinicID).
		Where("billings.status = ?", model.BillingStatusCompleted).
		Where("billings.completed_at AT TIME ZONE 'Asia/Tokyo' >= ?", input.PeriodStart).
		Where("billings.completed_at AT TIME ZONE 'Asia/Tokyo' < ?", input.PeriodEnd).
		Select(
			"ROUND(billing_items.tax_rate * 100)::bigint AS tax_rate," +
				" COALESCE(SUM(ROUND(billing_items.unit_price * billing_items.quantity::numeric)), 0) AS taxable_amount," +
				" COALESCE(SUM(ROUND(billing_items.unit_price * billing_items.quantity::numeric * billing_items.tax_rate)), 0) AS tax_amount",
		).
		Group("billing_items.tax_rate").
		Scan(&taxRows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate tax breakdown for close")
	}

	taxBreakdown := make([]TaxBreakdownRow, 0, len(taxRows))
	for _, tr := range taxRows {
		taxBreakdown = append(taxBreakdown, TaxBreakdownRow(tr))
	}

	return &CloseAggregateResult{
		PaymentRows:    paymentRows,
		CategoryRows:   categoryRows,
		TotalRefund:    totalRefund,
		BillingDetails: details,
		TaxBreakdown:   taxBreakdown,
	}, nil
}

// GetMonthlyReport は指定年月の月次売上レポートを集計する。FEAT-368
// payment_splits を正として集計（SUM(DISTINCT) hack を除去）。
func (r *accountingRepository) GetMonthlyReport(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResult, error) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, jst)
	end := start.AddDate(0, 1, 0)

	// Cartesian 積を避けるため payment_splits / billing_items を別クエリで集計する
	mArgs := []any{clinicID, model.BillingStatusCompleted, start, end}
	mCompletedCTE := `WITH completed_billings AS (
		SELECT id, completed_at FROM billings
		WHERE clinic_id = ? AND deleted_at IS NULL AND status = ?
		  AND completed_at AT TIME ZONE 'Asia/Tokyo' >= ?
		  AND completed_at AT TIME ZONE 'Asia/Tokyo' < ?
	)`

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
		JOIN payment_splits ps ON ps.billing_id = cb.id
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
	var mTotalRefund int64
	if err := r.db.WithContext(ctx).Raw(
		mCompletedCTE+`
		SELECT COALESCE(SUM(br.amount), 0)
		FROM billing_refunds br
		WHERE br.billing_id IN (SELECT id FROM completed_billings)
		`, mArgs...).Scan(&mTotalRefund).Error; err != nil {
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
		Where("clinic_id = ? AND deleted_at IS NULL", clinicID).
		Where("close_date >= ? AND close_date < ?", start.Format("2006-01-02"), end.Format("2006-01-02")).
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
	type taxRow struct {
		TaxRate       int64
		TaxableAmount int64
		TaxAmount     int64
	}
	var taxRows []taxRow
	if err := r.db.WithContext(ctx).
		Table("billing_items").
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.deleted_at IS NULL").
		Where("billings.clinic_id = ? AND billing_items.deleted_at IS NULL", clinicID).
		Where("billings.status = ?", model.BillingStatusCompleted).
		Where("billings.completed_at AT TIME ZONE 'Asia/Tokyo' >= ?", start).
		Where("billings.completed_at AT TIME ZONE 'Asia/Tokyo' < ?", end).
		Select(
			"ROUND(billing_items.tax_rate * 100)::bigint AS tax_rate," +
				" COALESCE(SUM(ROUND(billing_items.unit_price * billing_items.quantity::numeric)), 0) AS taxable_amount," +
				" COALESCE(SUM(ROUND(billing_items.unit_price * billing_items.quantity::numeric * billing_items.tax_rate)), 0) AS tax_amount",
		).
		Group("billing_items.tax_rate").
		Scan(&taxRows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to aggregate tax breakdown for monthly report")
	}

	taxBreakdown := make([]TaxBreakdownRow, 0, len(taxRows))
	for _, tr := range taxRows {
		taxBreakdown = append(taxBreakdown, TaxBreakdownRow(tr))
	}

	// 会計件数（billings 単位）
	var billingCount int64
	if err := r.db.WithContext(ctx).
		Model(&model.Billing{}).
		Scopes(clinicScope(clinicID)).
		Where("status = ?", model.BillingStatusCompleted).
		Where("completed_at AT TIME ZONE 'Asia/Tokyo' >= ?", start).
		Where("completed_at AT TIME ZONE 'Asia/Tokyo' < ?", end).
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
