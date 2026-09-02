package owner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/textsearch"
)

// 最終来院分類バケット名の定数（C-10）。
// SQL 内 CASE 式（本ファイル内、"Go 定数 ltvBucket* と一致必須" コメント参照）と
// Go 側の後段フィルタ・FindOwnerLTVParams.LastVisitBucket が同じ文字列を参照する。
const (
	ltvBucketNoVisit  = "no_visit"
	ltvBucketWithin3m = "within_3m"
	ltvBucketOver3m   = "over_3m"
	ltvBucketOver6m   = "over_6m"
	ltvBucketOver1y   = "over_1y"
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
		// translate() で DB 列のカタカナをひらがなに、U+3000 を ASCII 空白に正規化し、
		// 検索語も NormalizeQuerySpaces + NormalizeKana で同じ表現に揃える (BUG-001)。
		qSearch := textsearch.NormalizeQuerySpaces(params.Search)
		if qSearch == "" {
			where += " AND 1 = 0"
		} else {
			where += " AND translate(o.name, ?, ?) ILIKE ? ESCAPE '\\'"
			whereArgs = append(
				whereArgs,
				textsearch.KanaAndSpaceSourceChars,
				textsearch.KanaAndSpaceTargetChars,
				"%"+textsearch.EscapeLike(textsearch.NormalizeKana(qSearch))+"%",
			)
		}
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
	amountExpr, amountExprArgs := buildLTVAmountExpr(amountBasis, fromDate, toDate)

	// 外付けフィルタ（旧 HAVING）。来院は preagg 後の列参照なので GROUP BY 不要 → WHERE に AND 結合。
	having, havingArgs := buildLTVHaving(params, "COALESCE(ba.annual_amount, 0)", nil, fromDate, toDate)

	havingClause := ""
	if len(having) > 0 {
		havingClause = "AND " + strings.Join(having, " AND ")
	}

	// ORDER BY構築
	orderBy := r.buildOrderBy(params.Sort, params.Order)

	// 期間フィルタを CASE 式で適用（total_visit_count は全期間、period_visit_count のみ期間制限）。
	// 来院は owners ⟂ medical_records の nested loop ではなく、clinic 単位で pets 経由に
	// 事前集約してから owners に LEFT JOIN する（S10 / BUG-012: 40万件 MR で 20s timeout 回避）。
	periodVisitCountCondition := ""
	var periodVisitCountArgs []any
	billingCountExpr := "COUNT(DISTINCT b.id)"
	var billingCountArgs []any
	if fromDate != nil && toDate != nil {
		periodVisitCountCondition = "AND mr.date >= ? AND mr.date <= ?"
		periodVisitCountArgs = append(periodVisitCountArgs, fromDate, toDate)
		billingCountExpr = "COUNT(DISTINCT CASE WHEN COALESCE(bmr.date, b.scheduled_date) >= ? AND COALESCE(bmr.date, b.scheduled_date) <= ? THEN b.id END)"
		billingCountArgs = append(billingCountArgs, fromDate, toDate)
	}

	// last_visit_bucket CASE 式のリテラル ('no_visit'/'within_3m'/'over_3m'/'over_6m'/'over_1y')
	// は Go 定数 ltvBucket* と一致必須（C-10）。
	// 会計集計は医院単位で一度だけ飼主別に行い、来院行との直積による金額重複を防ぐ。
	// 来院帰属は medical_records.owner_id ではなく **現在の pets.owner_id**（譲渡後は現飼主）。
	query := fmt.Sprintf(`
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
`, periodVisitCountCondition, amountExpr, billingCountExpr, where, havingClause, orderBy)

	// Assemble args: visit period CASE, visit clinic, annual amount, billing count,
	// payments clinic, ba clinic/status, maxb clinic/status, owner scope, outer filters. (BUG-012 / S10)
	args := make([]any, 0, len(periodVisitCountArgs)+len(amountExprArgs)+len(billingCountArgs)+6+len(whereArgs)+len(havingArgs))
	args = append(args, periodVisitCountArgs...)
	args = append(args, params.ClinicID) // visit preagg clinic scope
	args = append(args, amountExprArgs...)
	args = append(args, billingCountArgs...)
	args = append(args, params.ClinicID) // payments clinic scope
	args = append(args, params.ClinicID, model.BillingStatusCompleted)
	args = append(args, params.ClinicID, model.BillingStatusCompleted) // maxb
	args = append(args, whereArgs...)
	args = append(args, havingArgs...)

	var rows []OwnerLTVRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "owner_ltv", "")
	}

	// Post-processing: フィルタリング（include_zero, include_no_visit）
	return filterLTVRows(rows, params), nil
}
