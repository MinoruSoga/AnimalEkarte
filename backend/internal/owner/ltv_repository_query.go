package owner

import (
	"fmt"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/textsearch"
)

// ownerLTVSelectSQL は FindOwnerLTV の集計 SELECT。%s は period CASE / amount / billing count /
// WHERE / HAVING / ORDER BY。プレースホルダ引数の順序は assembleLTVQueryArgs と一致必須。
const ownerLTVSelectSQL = `
SELECT
  o.id               AS owner_id,
  o.name             AS owner_name,
  o.line_user_id,
  o.lstep_opt_out,
  COALESCE(ba.total_amount, 0)                                                      AS total_amount,
  COALESCE(vs.total_visit_count, 0)                                                 AS total_visit_count,
  COALESCE(vs.annual_visit_count, 0)                                                AS annual_visit_count,
  vs.last_visit_date                                                                AS last_visit_date,
  vs.first_visit_date                                                               AS first_visit_date,
  COALESCE(ba.annual_amount, 0)                                                     AS annual_amount,
  COALESCE(ba.billing_count, 0)                                                     AS billing_count,
  COALESCE(vs.period_visit_count, 0)                                                AS period_visit_count,
  EXTRACT(DAY FROM NOW() - vs.last_visit_date)::int                                 AS days_since_last_visit,
  CASE
    WHEN vs.last_visit_date IS NULL THEN 'no_visit'
    WHEN EXTRACT(DAY FROM NOW() - vs.last_visit_date) < 90 THEN 'within_3m'
    WHEN EXTRACT(DAY FROM NOW() - vs.last_visit_date) < 180 THEN 'over_3m'
    WHEN EXTRACT(DAY FROM NOW() - vs.last_visit_date) < 365 THEN 'over_6m'
    ELSE 'over_1y'
  END AS last_visit_bucket,
  COALESCE(maxb.max_single_visit_amount, 0)                                         AS max_single_visit_amount
FROM owners o
LEFT JOIN (
  SELECT
    p.owner_id,
    p.clinic_id,
    COUNT(DISTINCT mr.date) AS total_visit_count,
    COUNT(DISTINCT CASE WHEN mr.date >= NOW() - INTERVAL '365 days' THEN mr.date END) AS annual_visit_count,
    MAX(mr.date) AS last_visit_date,
    MIN(mr.date) AS first_visit_date,
    COUNT(DISTINCT CASE WHEN TRUE %s THEN mr.date END) AS period_visit_count
  FROM medical_records mr
  INNER JOIN pets p
    ON p.id = mr.pet_id
   AND p.clinic_id = mr.clinic_id
  WHERE mr.clinic_id = ?
    AND mr.deleted_at IS NULL
  GROUP BY p.owner_id, p.clinic_id
) vs ON vs.clinic_id = o.clinic_id AND vs.owner_id = o.id
LEFT JOIN (
  SELECT
    b.clinic_id,
    COALESCE(b.owner_id, bmr.owner_id) AS owner_id,
    COALESCE(SUM(b.total_amount), 0) AS total_amount,
    %s AS annual_amount,
    %s AS billing_count
  FROM billings b
  LEFT JOIN medical_records bmr
    ON bmr.id = b.medical_record_id
    AND bmr.clinic_id = b.clinic_id
  LEFT JOIN (
    SELECT p.billing_id, b0.clinic_id, SUM(p.billing_amount) AS billing_amount
    FROM payments p
    INNER JOIN billings b0
      ON b0.id = p.billing_id
     AND b0.deleted_at IS NULL
     AND b0.clinic_id = ?
    WHERE p.deleted_at IS NULL
    GROUP BY p.billing_id, b0.clinic_id
  ) p ON p.billing_id = b.id AND p.clinic_id = b.clinic_id
  LEFT JOIN (
    SELECT billing_id, clinic_id, SUM(amount) AS amount
    FROM billing_refunds
    GROUP BY billing_id, clinic_id
  ) br ON br.billing_id = b.id AND br.clinic_id = b.clinic_id
  WHERE b.clinic_id = ?
    AND b.deleted_at IS NULL
    AND b.status = ?
    AND (
      (b.medical_record_id IS NULL AND b.owner_id IS NOT NULL)
      OR (
        bmr.id IS NOT NULL
        AND bmr.owner_id IS NOT NULL
        AND (b.owner_id IS NULL OR b.owner_id = bmr.owner_id)
      )
    )
  GROUP BY b.clinic_id, COALESCE(b.owner_id, bmr.owner_id)
) ba ON ba.clinic_id = o.clinic_id AND ba.owner_id = o.id
LEFT JOIN (
  SELECT
    b2.clinic_id,
    b2.owner_id,
    MAX(b2.total_amount) AS max_single_visit_amount
  FROM billings b2
  WHERE b2.clinic_id = ?
    AND b2.status = ?
    AND b2.deleted_at IS NULL
    AND b2.owner_id IS NOT NULL
    AND (
      b2.medical_record_id IS NULL
      OR EXISTS (
        SELECT 1
        FROM medical_records mr2
        WHERE mr2.id = b2.medical_record_id
          AND mr2.clinic_id = b2.clinic_id
          AND mr2.owner_id = b2.owner_id
          AND mr2.deleted_at IS NULL
      )
    )
  GROUP BY b2.clinic_id, b2.owner_id
) maxb ON maxb.clinic_id = o.clinic_id AND maxb.owner_id = o.id
WHERE %s
%s
ORDER BY %s
`

func buildLTVWhere(params *FindOwnerLTVParams) (where string, whereArgs []any) {
	where = "o.clinic_id = ? AND o.deleted_at IS NULL"
	whereArgs = append(whereArgs, params.ClinicID)
	if params.LineLinked {
		where += " AND o.line_user_id IS NOT NULL"
	}
	if params.Search == "" {
		return where, whereArgs
	}
	qSearch := textsearch.NormalizeQuerySpaces(params.Search)
	if qSearch == "" {
		return where + " AND 1 = 0", whereArgs
	}
	where += " AND translate(o.name, ?, ?) ILIKE ? ESCAPE '\\'"
	whereArgs = append(
		whereArgs,
		textsearch.KanaAndSpaceSourceChars,
		textsearch.KanaAndSpaceTargetChars,
		"%"+textsearch.EscapeLike(textsearch.NormalizeKana(qSearch))+"%",
	)
	return where, whereArgs
}

func buildLTVPeriodVisit(fromDate, toDate *time.Time) (periodVisitCountCondition string, periodVisitCountArgs []any, billingCountExpr string, billingCountArgs []any) {
	billingCountExpr = "COUNT(DISTINCT b.id)"
	if fromDate == nil || toDate == nil {
		return "", nil, billingCountExpr, nil
	}
	periodVisitCountCondition = "AND mr.date >= ? AND mr.date <= ?"
	periodVisitCountArgs = append(periodVisitCountArgs, fromDate, toDate)
	billingCountExpr = "COUNT(DISTINCT CASE WHEN COALESCE(bmr.date, b.scheduled_date) >= ? AND COALESCE(bmr.date, b.scheduled_date) <= ? THEN b.id END)"
	billingCountArgs = append(billingCountArgs, fromDate, toDate)
	return periodVisitCountCondition, periodVisitCountArgs, billingCountExpr, billingCountArgs
}

func assembleLTVQueryArgs(
	params *FindOwnerLTVParams,
	periodVisitCountArgs, amountExprArgs, billingCountArgs, whereArgs, havingArgs []any,
) []any {
	args := make([]any, 0, len(periodVisitCountArgs)+len(amountExprArgs)+len(billingCountArgs)+6+len(whereArgs)+len(havingArgs))
	args = append(args, periodVisitCountArgs...)
	args = append(args, params.ClinicID)
	args = append(args, amountExprArgs...)
	args = append(args, billingCountArgs...)
	args = append(args,
		params.ClinicID,
		params.ClinicID, model.BillingStatusCompleted,
		params.ClinicID, model.BillingStatusCompleted,
	)
	args = append(args, whereArgs...)
	args = append(args, havingArgs...)
	return args
}

// buildLTVAmountExpr は AmountBasis に応じた金額集計式と、その式が要するプレースホルダ引数を
// 構築する（BE-refactor.md E-12: FindOwnerLTV の位置引数結合を事故源から隔離する純粋抽出）。
func buildLTVAmountExpr(basis string, from, to *time.Time) (amountExpr string, amountExprArgs []any) {
	hasPeriodFilter := from != nil && to != nil
	switch basis {
	case "paid_amount":
		if hasPeriodFilter {
			amountExpr = "COALESCE(SUM(CASE WHEN COALESCE(bmr.date, b.scheduled_date) >= ? AND COALESCE(bmr.date, b.scheduled_date) <= ? THEN p.billing_amount ELSE 0 END), 0)"
		} else {
			amountExpr = "COALESCE(SUM(p.billing_amount), 0)"
		}
	case "net_paid_amount":
		if hasPeriodFilter {
			amountExpr = "COALESCE(SUM(CASE WHEN COALESCE(bmr.date, b.scheduled_date) >= ? AND COALESCE(bmr.date, b.scheduled_date) <= ? THEN p.billing_amount ELSE 0 END) - COALESCE(SUM(CASE WHEN COALESCE(bmr.date, b.scheduled_date) >= ? AND COALESCE(bmr.date, b.scheduled_date) <= ? THEN br.amount ELSE 0 END), 0), 0)"
		} else {
			amountExpr = "COALESCE(SUM(p.billing_amount) - COALESCE(SUM(br.amount), 0), 0)"
		}
	default: // gross_total_amount
		if hasPeriodFilter {
			amountExpr = "COALESCE(SUM(CASE WHEN COALESCE(bmr.date, b.scheduled_date) >= ? AND COALESCE(bmr.date, b.scheduled_date) <= ? THEN b.total_amount ELSE 0 END), 0)"
		} else {
			amountExpr = "COALESCE(SUM(b.total_amount), 0)"
		}
	}
	if hasPeriodFilter {
		amountExprArgs = append(amountExprArgs, from, to)
		if basis == "net_paid_amount" {
			amountExprArgs = append(amountExprArgs, from, to)
		}
	}
	return amountExpr, amountExprArgs
}

// buildLTVHaving は HAVING 句の条件断片とバインド引数を構築する（BE-refactor.md E-12）。
func buildLTVHaving(params *FindOwnerLTVParams, amountExpr string, amountExprArgs []any, from, to *time.Time) (having []string, havingArgs []any) {
	_ = from
	_ = to

	// 会計額フィルタ（AGG-BE-001: min_amount/max_amount は期間内 annual_amount）
	if params.MinTotalAmount != nil {
		having, havingArgs = appendAmountHaving(having, havingArgs, amountExpr, amountExprArgs, ">=", *params.MinTotalAmount)
	}
	if params.MaxTotalAmount != nil {
		having, havingArgs = appendAmountHaving(having, havingArgs, amountExpr, amountExprArgs, "<=", *params.MaxTotalAmount)
	}

	// 来院回数フィルタ（AGG-BE-002）— vs.period_visit_count は期間指定時のみ期間内、未指定時は全期間。
	if params.MinVisitCount != nil {
		having = append(having, "COALESCE(vs.period_visit_count, 0) >= ?")
		havingArgs = append(havingArgs, *params.MinVisitCount)
	}
	if params.MaxVisitCount != nil {
		having = append(having, "COALESCE(vs.period_visit_count, 0) <= ?")
		havingArgs = append(havingArgs, *params.MaxVisitCount)
	}
	return having, havingArgs
}

// shouldExcludeZeroAnnualAmount は include_zero=false 時に annual_amount=0 を落とすかを決める（BUG-012）。
// include_zero は AGG-BE-001 売上ランキング向け。来院回数・最終来院軸では 0 円除外を適用しない
// （UI の「0円を含む」は売上タブのみ。来院あり・会計 0 の飼主が常に消えるのを防ぐ）。
func shouldExcludeZeroAnnualAmount(params *FindOwnerLTVParams) bool {
	if params.IncludeZero {
		return false
	}
	// 金額レンジ明示時は常に売上フィルタ
	if params.MinTotalAmount != nil || params.MaxTotalAmount != nil {
		return true
	}
	// 売上タブ既定: year / amount_basis / 金額ソート
	if params.Year != nil {
		return true
	}
	if params.AmountBasis != "" {
		return true
	}
	switch params.Sort {
	case "annual_amount", "total_amount":
		return true
	}
	// 来院・最終来院中心クエリは 0 円除外しない
	if params.PeriodPreset != "" ||
		params.LastVisitBucket != "" ||
		params.MinVisitCount != nil ||
		params.MaxVisitCount != nil ||
		params.Sort == "visit_count" ||
		params.Sort == "period_visit_count" ||
		params.Sort == "total_visit_count" ||
		params.Sort == "annual_visit_count" ||
		params.Sort == "last_visit_date" ||
		params.Sort == "days_since_last_visit" {
		return false
	}
	// 後方互換: 素の LTV 一覧（既定 sort=total_amount 相当）は 0 円除外
	return true
}

// filterLTVRows は include_zero / include_no_visit / last_visit_bucket の Go 側後段フィルタを
// 適用する（BE-refactor.md E-12）。
func filterLTVRows(rows []OwnerLTVRow, params *FindOwnerLTVParams) []OwnerLTVRow {
	var filtered []OwnerLTVRow
	excludeZero := shouldExcludeZeroAnnualAmount(params)
	for i := range rows {
		row := &rows[i]
		// include_zero フィルタ（AGG-BE-001 / BUG-012: 売上軸のみ）
		if excludeZero && row.AnnualAmount != nil && *row.AnnualAmount == 0 {
			continue
		}
		// include_no_visit フィルタ（AGG-BE-003）
		if !params.IncludeNoVisit && row.LastVisitBucket != nil && *row.LastVisitBucket == ltvBucketNoVisit {
			continue
		}
		// last_visit_bucket フィルタ（AGG-BE-003 / BUG-008）
		// 区分（over_3m 等）選択時に「来院なしを含む」が ON なら、その区分 OR no_visit を残す。
		if params.LastVisitBucket != "" && params.LastVisitBucket != ltvBucketNoVisit {
			if row.LastVisitBucket == nil {
				continue
			}
			matchesBucket := *row.LastVisitBucket == params.LastVisitBucket
			matchesNoVisit := params.IncludeNoVisit && *row.LastVisitBucket == ltvBucketNoVisit
			if !matchesBucket && !matchesNoVisit {
				continue
			}
		} else if params.LastVisitBucket != "" {
			if row.LastVisitBucket == nil || *row.LastVisitBucket != params.LastVisitBucket {
				continue
			}
		}
		filtered = append(filtered, *row)
	}
	return filtered
}

// appendAmountHaving はローカル生成した集計式だけを SQL 断片として埋め込み、
// 動的な閾値は必ずプレースホルダでバインドする。
func appendAmountHaving(having []string, args []any, amountExpr string, amountExprArgs []any, op string, amount int64) (outHaving []string, outArgs []any) {
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
		from, err := time.ParseInLocation(time.DateOnly, *params.From, time.Local)
		if err != nil {
			return nil, nil, apperrors.Wrap(err, fmt.Sprintf("invalid From date format: %s (expected YYYY-MM-DD)", *params.From))
		}
		to, err := time.ParseInLocation(time.DateOnly, *params.To, time.Local)
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
