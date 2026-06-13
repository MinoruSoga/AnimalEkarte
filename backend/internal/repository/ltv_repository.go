package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
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
	// MaxSingleVisitAmount は完了済み請求の単一最大額（CPMスポット判定用、ISSUE-006）。
	// タグ同期側 AccountingRepository.MaxSingleVisitAmountByOwner と同一の集計範囲を保つため、
	// medical_record の JOIN を経由せず billings から直接取得する。
	MaxSingleVisitAmount int64 `gorm:"column:max_single_visit_amount"`
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
	FindOwnerLTV(ctx context.Context, params *FindOwnerLTVParams) ([]OwnerLTVRow, error)
}

type ltvRepository struct {
	db *gorm.DB
}

// NewLtvRepository は LtvRepository を初期化して返す。
func NewLtvRepository(db *gorm.DB) LtvRepository {
	return &ltvRepository{db: db}
}

func (r *ltvRepository) FindOwnerLTV(ctx context.Context, params *FindOwnerLTVParams) ([]OwnerLTVRow, error) {
	// Build all string components first, collecting args in separate slices
	where := "o.clinic_id = ? AND o.deleted_at IS NULL"
	var whereArgs []any
	whereArgs = append(whereArgs, params.ClinicID)

	if params.LineLinked {
		where += " AND o.line_user_id IS NOT NULL"
	}

	if params.Search != "" {
		where += " AND o.name ILIKE ?"
		whereArgs = append(whereArgs, "%"+params.Search+"%")
	}

	// 期間決定（AGG-BE-001/002/003）
	fromDate, toDate, err := r.calculateDateRange(params)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to parse date range parameters")
	}

	// 金額基準選択（AGG-BE-001）
	amountBasis := params.AmountBasis
	if amountBasis == "" {
		amountBasis = "gross_total_amount"
	}

	// 金額式の構築（期間フィルタ付き）
	var amountExpr string
	hasPeriodFilter := fromDate != nil && toDate != nil
	switch amountBasis {
	case "paid_amount":
		if hasPeriodFilter {
			amountExpr = "COALESCE(SUM(CASE WHEN mr.date >= ? AND mr.date <= ? THEN p.billing_amount ELSE 0 END), 0)"
		} else {
			amountExpr = "COALESCE(SUM(p.billing_amount), 0)"
		}
	case "net_paid_amount":
		if hasPeriodFilter {
			amountExpr = "COALESCE(SUM(CASE WHEN mr.date >= ? AND mr.date <= ? THEN p.billing_amount ELSE 0 END) - COALESCE(SUM(CASE WHEN mr.date >= ? AND mr.date <= ? THEN br.amount ELSE 0 END), 0), 0)"
		} else {
			amountExpr = "COALESCE(SUM(p.billing_amount) - COALESCE(SUM(br.amount), 0), 0)"
		}
	default: // gross_total_amount
		if hasPeriodFilter {
			amountExpr = "COALESCE(SUM(CASE WHEN mr.date >= ? AND mr.date <= ? THEN b.total_amount ELSE 0 END), 0)"
		} else {
			amountExpr = "COALESCE(SUM(b.total_amount), 0)"
		}
	}
	var amountExprArgs []any
	if hasPeriodFilter {
		amountExprArgs = append(amountExprArgs, fromDate, toDate)
		if amountBasis == "net_paid_amount" {
			amountExprArgs = append(amountExprArgs, fromDate, toDate)
		}
	}

	// HAVING句構築
	var having []string
	var havingArgs []any

	// 全期間の会計額フィルタ（AGG-BE-001: min_amount/max_amount は期間内）
	if params.MinTotalAmount != nil {
		having, havingArgs = appendAmountHaving(having, havingArgs, amountExpr, amountExprArgs, ">=", *params.MinTotalAmount)
	}
	if params.MaxTotalAmount != nil {
		having, havingArgs = appendAmountHaving(having, havingArgs, amountExpr, amountExprArgs, "<=", *params.MaxTotalAmount)
	}

	// 来院回数フィルタ（AGG-BE-002）
	if params.MinVisitCount != nil {
		having = append(having, "COUNT(DISTINCT CASE WHEN (mr.date >= COALESCE(?::date, mr.date) AND mr.date <= COALESCE(?::date, mr.date)) THEN mr.date END) >= ?")
		havingArgs = append(havingArgs, fromDate, toDate, *params.MinVisitCount)
	}
	if params.MaxVisitCount != nil {
		having = append(having, "COUNT(DISTINCT CASE WHEN (mr.date >= COALESCE(?::date, mr.date) AND mr.date <= COALESCE(?::date, mr.date)) THEN mr.date END) <= ?")
		havingArgs = append(havingArgs, fromDate, toDate, *params.MaxVisitCount)
	}

	havingClause := ""
	if len(having) > 0 {
		havingClause = "HAVING " + strings.Join(having, " AND ")
	}

	// ORDER BY構築
	orderBy := r.buildOrderBy(params.Sort, params.Order)

	// 期間フィルタをCASE式で適用（total_visit_countは全期間、period_visit_countのみ期間制限）
	periodFilter := ""
	var periodFilterArgs []any
	if fromDate != nil && toDate != nil {
		periodFilter = "(mr.date >= ? AND mr.date <= ?)"
		// amountExpr用: net_paid_amount の場合は4個、その他は2個のargs
		periodFilterArgs = append(periodFilterArgs, amountExprArgs...)
	}

	// period_visit_count用フィルタ条件（CASE WHEN の条件部分）
	// 注意: periodVisitCountCondition は query内で2回使われるため、fromDate/toDate を2回分必要
	periodVisitCountCondition := ""
	if periodFilter != "" {
		periodVisitCountCondition = "AND " + periodFilter
		// periodVisitCountCondition が query内で2回使われるため、args を2回分追加
		periodFilterArgs = append(periodFilterArgs, fromDate, toDate, fromDate, toDate)
	}

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
  COUNT(DISTINCT CASE WHEN mr.clinic_id = o.clinic_id %s THEN b.id END)              AS billing_count,
  COUNT(DISTINCT CASE WHEN mr.clinic_id = o.clinic_id %s THEN mr.date END)           AS period_visit_count,
  EXTRACT(DAY FROM NOW() - MAX(mr.date))::int                                        AS days_since_last_visit,
  CASE
    WHEN MAX(mr.date) IS NULL THEN 'no_visit'
    WHEN EXTRACT(DAY FROM NOW() - MAX(mr.date)) < 90 THEN 'within_3m'
    WHEN EXTRACT(DAY FROM NOW() - MAX(mr.date)) < 180 THEN 'over_3m'
    WHEN EXTRACT(DAY FROM NOW() - MAX(mr.date)) < 365 THEN 'over_6m'
    ELSE 'over_1y'
  END AS last_visit_bucket,
  COALESCE((
    SELECT MAX(b2.total_amount)
    FROM billings b2
    WHERE b2.clinic_id = o.clinic_id
      AND b2.owner_id = o.id
      AND b2.status = 'completed'
      AND b2.deleted_at IS NULL
  ), 0)                                                                              AS max_single_visit_amount
FROM owners o
LEFT JOIN medical_records mr ON mr.owner_id = o.id AND mr.clinic_id = o.clinic_id AND mr.deleted_at IS NULL
LEFT JOIN billings b ON b.medical_record_id = mr.id AND b.clinic_id = o.clinic_id AND b.deleted_at IS NULL AND b.status = 'completed'
LEFT JOIN payments p ON p.billing_id = b.id AND p.deleted_at IS NULL
LEFT JOIN billing_refunds br ON br.billing_id = b.id
WHERE %s
GROUP BY o.id, o.name, o.line_user_id, o.lstep_opt_out
%s
ORDER BY %s
`, amountExpr, periodVisitCountCondition, periodVisitCountCondition, where, havingClause, orderBy)

	// Assemble args in the correct order: periodFilter args, then where args, then having args
	args := make([]any, 0, len(periodFilterArgs)+len(whereArgs)+len(havingArgs))
	args = append(args, periodFilterArgs...)
	args = append(args, whereArgs...)
	args = append(args, havingArgs...)

	var rows []OwnerLTVRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "owner_ltv", "")
	}

	// Post-processing: フィルタリング（include_zero, include_no_visit）
	var filtered []OwnerLTVRow
	for i := range rows {
		row := &rows[i]
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
		filtered = append(filtered, *row)
	}

	return filtered, nil
}

// appendAmountHaving はローカル生成した集計式だけを SQL 断片として埋め込み、
// 動的な閾値は必ずプレースホルダでバインドする。
func appendAmountHaving(having []string, args []any, amountExpr string, amountExprArgs []any, op string, amount int64) ([]string, []any) {
	having = append(having, fmt.Sprintf("%s %s ?", amountExpr, op))
	args = append(args, amountExprArgs...)
	args = append(args, amount)
	return having, args
}

// calculateDateRange は year/from/to/period_preset から集計期間を決定する。
func (r *ltvRepository) calculateDateRange(params *FindOwnerLTVParams) (fromDate, toDate *time.Time, err error) {
	now := time.Now().In(time.Local)
	currentYear := now.Year()

	// from/to が明示的に指定されている場合はそれを優先（優先度1）
	if params.From != nil && params.To != nil {
		// YYYY-MM-DD 形式をパース（エラーを明示的に処理）
		from, err := time.ParseInLocation("2006-01-02", *params.From, time.Local)
		if err != nil {
			return nil, nil, apperrors.Wrap(err, fmt.Sprintf("invalid From date format: %s (expected YYYY-MM-DD)", *params.From))
		}
		to, err := time.ParseInLocation("2006-01-02", *params.To, time.Local)
		if err != nil {
			return nil, nil, apperrors.Wrap(err, fmt.Sprintf("invalid To date format: %s (expected YYYY-MM-DD)", *params.To))
		}
		return &from, &to, nil
	}

	// year が指定されている場合（優先度2）（AGG-BE-001）
	if params.Year != nil {
		from := time.Date(*params.Year, 1, 1, 0, 0, 0, 0, time.Local)
		to := time.Date(*params.Year, 12, 31, 23, 59, 59, 0, time.Local)
		return &from, &to, nil
	}

	// period_preset に基づいて期間を決定（優先度3）（AGG-BE-002）
	switch params.PeriodPreset {
	case "last_3_months":
		from := now.AddDate(0, -3, 0)
		return &from, &now, nil
	case "last_6_months":
		from := now.AddDate(0, -6, 0)
		return &from, &now, nil
	case "last_12_months":
		from := now.AddDate(-1, 0, 0)
		return &from, &now, nil
	case "calendar_year":
		from := time.Date(currentYear, 1, 1, 0, 0, 0, 0, now.Location())
		return &from, &now, nil
	}

	// デフォルト: from/to は nil（全期間）
	return nil, nil, nil
}

// buildOrderBy はソートフィールドと順序から ORDER BY 句を構築する。
func (r *ltvRepository) buildOrderBy(sort, order string) string {
	// Whitelist order parameter to prevent SQL injection
	if order != "asc" && order != "desc" {
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
