package owner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
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
	where, whereArgs := buildLTVWhere(params)

	fromDate, toDate, err := r.calculateDateRange(params)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to parse date range parameters")
	}

	amountBasis := params.AmountBasis
	if amountBasis == "" {
		amountBasis = "gross_total_amount"
	}
	amountExpr, amountExprArgs := buildLTVAmountExpr(amountBasis, fromDate, toDate)
	having, havingArgs := buildLTVHaving(params, "COALESCE(ba.annual_amount, 0)", nil, fromDate, toDate)
	havingClause := ""
	if len(having) > 0 {
		havingClause = "AND " + strings.Join(having, " AND ")
	}
	orderBy := r.buildOrderBy(params.Sort, params.Order)
	periodVisitCountCondition, periodVisitCountArgs, billingCountExpr, billingCountArgs := buildLTVPeriodVisit(fromDate, toDate)

	query := fmt.Sprintf(
		ownerLTVSelectSQL,
		periodVisitCountCondition,
		amountExpr,
		billingCountExpr,
		where,
		havingClause,
		orderBy,
	)
	args := assembleLTVQueryArgs(params, periodVisitCountArgs, amountExprArgs, billingCountArgs, whereArgs, havingArgs)

	var rows []OwnerLTVRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "owner_ltv", "")
	}
	return filterLTVRows(rows, params), nil
}
