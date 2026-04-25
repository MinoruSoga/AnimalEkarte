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
	OwnerID          uint64     `gorm:"column:owner_id"`
	OwnerName        string     `gorm:"column:owner_name"`
	LineUserID       *string    `gorm:"column:line_user_id"`
	LstepOptOut      bool       `gorm:"column:lstep_opt_out"`
	TotalAmount      int64      `gorm:"column:total_amount"`
	TotalVisitCount  int64      `gorm:"column:total_visit_count"`
	AnnualVisitCount int64      `gorm:"column:annual_visit_count"`
	LastVisitDate    *time.Time `gorm:"column:last_visit_date"`
	FirstVisitDate   *time.Time `gorm:"column:first_visit_date"`
}

// FindOwnerLTVParams はLTV一覧検索のパラメータ。
type FindOwnerLTVParams struct {
	ClinicID       uint64
	Sort           string // total_amount_desc / visit_count_desc / last_visit_desc
	MinTotalAmount *int64
	MaxTotalAmount *int64
	MinVisitCount  *int64
	LineLinked     bool
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

	var having []string
	if params.MinTotalAmount != nil {
		having = append(having, fmt.Sprintf("COALESCE(SUM(b.total_amount), 0) >= %d", *params.MinTotalAmount))
	}
	if params.MaxTotalAmount != nil {
		having = append(having, fmt.Sprintf("COALESCE(SUM(b.total_amount), 0) <= %d", *params.MaxTotalAmount))
	}
	if params.MinVisitCount != nil {
		having = append(having, fmt.Sprintf("COUNT(DISTINCT mr.date) >= %d", *params.MinVisitCount))
	}

	havingClause := ""
	if len(having) > 0 {
		havingClause = "HAVING " + strings.Join(having, " AND ")
	}

	orderBy := "total_amount DESC"
	switch params.Sort {
	case "visit_count_desc":
		orderBy = "total_visit_count DESC"
	case "last_visit_desc":
		orderBy = "last_visit_date DESC NULLS LAST"
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
  MAX(mr.date) AS last_visit_date,
  MIN(mr.date) AS first_visit_date
FROM owners o
LEFT JOIN medical_records mr ON mr.owner_id = o.id AND mr.clinic_id = o.clinic_id AND mr.deleted_at IS NULL
LEFT JOIN billings b ON b.medical_record_id = mr.id AND b.clinic_id = o.clinic_id AND b.deleted_at IS NULL
WHERE %s
GROUP BY o.id, o.name, o.line_user_id, o.lstep_opt_out
%s
ORDER BY %s
`, where, havingClause, orderBy)

	var rows []OwnerLTVRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "owner_ltv", "")
	}
	return rows, nil
}
