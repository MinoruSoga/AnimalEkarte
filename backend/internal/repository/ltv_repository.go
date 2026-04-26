package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"gorm.io/gorm"
)

// OwnerLTVRow はLTV集計クエリの結果行。
type OwnerLTVRow struct {
	OwnerID            uint64     `gorm:"column:owner_id"`
	OwnerName          string     `gorm:"column:owner_name"`
	LineUserID         *string    `gorm:"column:line_user_id"`
	LstepOptOut        bool       `gorm:"column:lstep_opt_out"`
	TotalAmount        int64      `gorm:"column:total_amount"`
	TotalVisitCount    int64      `gorm:"column:total_visit_count"`
	AnnualVisitCount   int64      `gorm:"column:annual_visit_count"`
	LastVisitDate      *time.Time `gorm:"column:last_visit_date"`
	FirstVisitDate     *time.Time `gorm:"column:first_visit_date"`
	AnnualAmount       *int64     `gorm:"column:annual_amount"`
	BillingCount       *int64     `gorm:"column:billing_count"`
	PeriodVisitCount   *int64     `gorm:"column:period_visit_count"`
	DaysSinceLastVisit *int       `gorm:"column:days_since_last_visit"`
	LastVisitBucket    *string    `gorm:"column:last_visit_bucket"`
}

// FindOwnerLTVParams はLTV一覧検索のパラメータ。
type FindOwnerLTVParams struct {
	// 既存フィールド
	ClinicID       uint64
	Sort           string
	MinTotalAmount *int64
	MaxTotalAmount *int64
	MinVisitCount  *int64
	LineLinked     bool
	// AGG-BE-001 年間売上ランキング用
	Year        *int
	From        *string // YYYY-MM-DD format
	To          *string // YYYY-MM-DD format
	AmountBasis string  // gross_total_amount (default) / paid_amount / net_paid_amount
	IncludeZero bool
	Search      string // 飼い主名部分一致検索
	// AGG-BE-002 来院回数用
	PeriodPreset  string // last_3_months / last_6_months / last_12_months / calendar_year
	MaxVisitCount *int64
	// AGG-BE-003 最終来院分類用
	LastVisitBucket string // within_3m / over_3m / over_6m / over_1y / no_visit or ""
	IncludeNoVisit  bool
	Order           string // asc / desc
}

// LtvRepository はLTV集計のリポジトリインターフェース（BE-010）。
type LtvRepository interface {
	FindOwnerLTV(ctx context.Context, params FindOwnerLTVParams) ([]OwnerLTVRow, error)
}

type ltvRepository struct {
	db *gorm.DB
}

// NewLtvRepository は LtvRepository を初期化して返す。
func NewLtvRepository(db *gorm.DB) LtvRepository {
	return &ltvRepository{db: db}
}

func (r *ltvRepository) FindOwnerLTV(ctx context.Context, params FindOwnerLTVParams) ([]OwnerLTVRow, error) {
	var args []any

	where := "o.clinic_id = ? AND o.deleted_at IS NULL"
	args = append(args, params.ClinicID)

	if params.LineLinked {
		where += " AND o.line_user_id IS NOT NULL"
	}

	if params.Search != "" {
		where += " AND o.name ILIKE ?"
		args = append(args, "%"+params.Search+"%")
	}

	// 期間決定（AGG-BE-001/002/003）
	fromDate, toDate := r.calculateDateRange(params)

	// 期間フィルタ（WHERE句に追加）
	if fromDate != nil && toDate != nil {
		where += " AND mr.date >= ? AND mr.date <= ?"
		args = append(args, fromDate, toDate)
	}

	// 金額基準選択（AGG-BE-001）
	amountBasis := params.AmountBasis
	if amountBasis == "" {
		amountBasis = "gross_total_amount"
	}

	var amountExpr string
	switch amountBasis {
	case "paid_amount":
		amountExpr = "COALESCE(SUM(p.billing_amount), 0)"
	case "net_paid_amount":
		amountExpr = "COALESCE(SUM(p.billing_amount) - COALESCE(SUM(br.amount), 0), 0)"
	default: // gross_total_amount
		amountExpr = "COALESCE(SUM(b.total_amount), 0)"
	}

	// HAVING句構築
	var having []string

	// 全期間の会計額フィルタ（AGG-BE-001: min_amount/max_amount は期間内）
	if params.MinTotalAmount != nil {
		having = append(having, fmt.Sprintf("%s >= %d", amountExpr, *params.MinTotalAmount))
	}
	if params.MaxTotalAmount != nil {
		having = append(having, fmt.Sprintf("%s <= %d", amountExpr, *params.MaxTotalAmount))
	}

	// 来院回数フィルタ（AGG-BE-002）
	if params.MinVisitCount != nil {
		having = append(having, fmt.Sprintf("COUNT(DISTINCT CASE WHEN (mr.date >= COALESCE(?::date, mr.date) AND mr.date <= COALESCE(?::date, mr.date)) THEN mr.date END) >= %d", *params.MinVisitCount))
		args = append(args, fromDate, toDate)
	}
	if params.MaxVisitCount != nil {
		having = append(having, fmt.Sprintf("COUNT(DISTINCT CASE WHEN (mr.date >= COALESCE(?::date, mr.date) AND mr.date <= COALESCE(?::date, mr.date)) THEN mr.date END) <= %d", *params.MaxVisitCount))
		args = append(args, fromDate, toDate)
	}

	havingClause := ""
	if len(having) > 0 {
		havingClause = "HAVING " + strings.Join(having, " AND ")
	}

	// ORDER BY構築
	orderBy := r.buildOrderBy(params.Sort, params.Order)

	query := fmt.Sprintf(`
SELECT
  o.id               AS owner_id,
  o.name             AS owner_name,
  o.line_user_id,
  o.lstep_opt_out,
  COALESCE(SUM(b.total_amount), 0)                                                  AS total_amount,
  COUNT(DISTINCT mr.date)                                                            AS total_visit_count,
  COUNT(DISTINCT CASE WHEN mr.date >= NOW() - INTERVAL '365 days' THEN mr.date END) AS annual_visit_count,
  MAX(mr.date)                                                                        AS last_visit_date,
  MIN(mr.date)                                                                        AS first_visit_date,
  %s                                                                                   AS annual_amount,
  COUNT(DISTINCT CASE WHEN mr.clinic_id = o.clinic_id THEN b.id END)                 AS billing_count,
  COUNT(DISTINCT CASE WHEN mr.clinic_id = o.clinic_id THEN mr.date END)              AS period_visit_count,
  EXTRACT(DAY FROM NOW() - MAX(mr.date))::int                                        AS days_since_last_visit,
  CASE
    WHEN MAX(mr.date) IS NULL THEN 'no_visit'
    WHEN EXTRACT(DAY FROM NOW() - MAX(mr.date)) < 90 THEN 'within_3m'
    WHEN EXTRACT(DAY FROM NOW() - MAX(mr.date)) < 180 THEN 'over_3m'
    WHEN EXTRACT(DAY FROM NOW() - MAX(mr.date)) < 365 THEN 'over_6m'
    ELSE 'over_1y'
  END AS last_visit_bucket
FROM owners o
LEFT JOIN medical_records mr ON mr.owner_id = o.id AND mr.clinic_id = o.clinic_id AND mr.deleted_at IS NULL
LEFT JOIN billings b ON b.medical_record_id = mr.id AND b.clinic_id = o.clinic_id AND b.deleted_at IS NULL AND b.status = 'completed'
LEFT JOIN payments p ON p.billing_id = b.id AND p.deleted_at IS NULL
LEFT JOIN billing_refunds br ON br.billing_id = b.id
WHERE %s
GROUP BY o.id, o.name, o.line_user_id, o.lstep_opt_out
%s
ORDER BY %s
`, amountExpr, where, havingClause, orderBy)

	var rows []OwnerLTVRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "owner_ltv", "")
	}

	// Post-processing: フィルタリング（include_zero, include_no_visit）
	var filtered []OwnerLTVRow
	for _, row := range rows {
		// include_zero フィルタ（AGG-BE-001）
		if !params.IncludeZero && row.AnnualAmount != nil && *row.AnnualAmount == 0 {
			continue
		}
		// include_no_visit フィルタ（AGG-BE-003）
		if !params.IncludeNoVisit && row.LastVisitBucket != nil && *row.LastVisitBucket == "no_visit" {
			continue
		}
		// last_visit_bucket フィルタ（AGG-BE-003）
		if params.LastVisitBucket != "" && (row.LastVisitBucket == nil || *row.LastVisitBucket != params.LastVisitBucket) {
			continue
		}
		filtered = append(filtered, row)
	}

	return filtered, nil
}

// calculateDateRange は year/from/to/period_preset から集計期間を決定する。
func (r *ltvRepository) calculateDateRange(params FindOwnerLTVParams) (*time.Time, *time.Time) {
	now := time.Now()
	currentYear := now.Year()

	// from/to が明示的に指定されている場合はそれを優先
	if params.From != nil && params.To != nil {
		// YYYY-MM-DD 形式をパース
		from, _ := time.Parse("2006-01-02", *params.From)
		to, _ := time.Parse("2006-01-02", *params.To)
		return &from, &to
	}

	// period_preset に基づいて期間を決定（AGG-BE-002）
	switch params.PeriodPreset {
	case "last_3_months":
		from := now.AddDate(0, -3, 0)
		return &from, &now
	case "last_6_months":
		from := now.AddDate(0, -6, 0)
		return &from, &now
	case "last_12_months":
		from := now.AddDate(-1, 0, 0)
		return &from, &now
	case "calendar_year":
		from := time.Date(currentYear, 1, 1, 0, 0, 0, 0, now.Location())
		return &from, &now
	}

	// year が指定されている場合（AGG-BE-001）
	if params.Year != nil {
		from := time.Date(*params.Year, 1, 1, 0, 0, 0, 0, now.Location())
		to := time.Date(*params.Year, 12, 31, 23, 59, 59, 0, now.Location())
		return &from, &to
	}

	// デフォルト: from/to は nil（全期間）
	return nil, nil
}

// buildOrderBy はソートフィールドと順序から ORDER BY 句を構築する。
func (r *ltvRepository) buildOrderBy(sort string, order string) string {
	if order == "" {
		order = "desc"
	}

	switch sort {
	case "annual_amount":
		return fmt.Sprintf("annual_amount %s NULLS LAST", strings.ToUpper(order))
	case "total_amount":
		return fmt.Sprintf("total_amount %s NULLS LAST", strings.ToUpper(order))
	case "visit_count":
		return fmt.Sprintf("period_visit_count %s NULLS LAST", strings.ToUpper(order))
	case "total_visit_count":
		return fmt.Sprintf("total_visit_count %s NULLS LAST", strings.ToUpper(order))
	case "annual_visit_count":
		return fmt.Sprintf("annual_visit_count %s NULLS LAST", strings.ToUpper(order))
	case "last_visit_date":
		return fmt.Sprintf("last_visit_date %s NULLS LAST", strings.ToUpper(order))
	case "days_since_last_visit":
		return fmt.Sprintf("days_since_last_visit %s NULLS LAST", strings.ToUpper(order))
	case "owner_name":
		return fmt.Sprintf("owner_name %s", strings.ToUpper(order))
	default:
		return fmt.Sprintf("total_amount %s NULLS LAST", strings.ToUpper(order))
	}
}
