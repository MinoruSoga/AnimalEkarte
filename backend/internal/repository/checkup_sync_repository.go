package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// CheckupSyncPreviewRow はプレビュークエリの結果行。
type CheckupSyncPreviewRow struct {
	OwnerID       uint64     `gorm:"column:owner_id"`
	OwnerName     string     `gorm:"column:owner_name"`
	LineUserID    *string    `gorm:"column:line_user_id"`
	LstepOptOut   bool       `gorm:"column:lstep_opt_out"`
	PetNamesCSV   string     `gorm:"column:pet_names"`
	LastVisitDate *time.Time `gorm:"column:last_visit_date"`
}

// FindCheckupSyncPreviewParams はプレビュー検索パラメータ。
type FindCheckupSyncPreviewParams struct {
	ClinicID        uint64
	Species         string
	LastVisitBefore *time.Time
	LastVisitAfter  *time.Time
}

// CheckupSyncRepository は健診同期プレビューのリポジトリインターフェース（BE-004）。
type CheckupSyncRepository interface {
	FindCheckupSyncPreview(ctx context.Context, params FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error)
}

type checkupSyncRepository struct {
	db *gorm.DB
}

// NewCheckupSyncRepository は CheckupSyncRepository を初期化して返す。
func NewCheckupSyncRepository(db *gorm.DB) CheckupSyncRepository {
	return &checkupSyncRepository{db: db}
}

func (r *checkupSyncRepository) FindCheckupSyncPreview(ctx context.Context, params FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error) {
	var args []any

	where := "o.clinic_id = ? AND o.deleted_at IS NULL"
	args = append(args, params.ClinicID)

	if params.Species != "" {
		where += " AND EXISTS (SELECT 1 FROM pets p2 JOIN animal_species sp ON sp.id = p2.animal_species_id WHERE p2.owner_id = o.id AND p2.clinic_id = o.clinic_id AND p2.deleted_at IS NULL AND LOWER(sp.name) = LOWER(?))"
		args = append(args, params.Species)
	}

	var having []string
	if params.LastVisitBefore != nil {
		having = append(having, "MAX(mr.date) <= ?")
		args = append(args, params.LastVisitBefore.Format("2006-01-02"))
	}
	if params.LastVisitAfter != nil {
		having = append(having, "MAX(mr.date) >= ?")
		args = append(args, params.LastVisitAfter.Format("2006-01-02"))
	}

	havingClause := ""
	if len(having) > 0 {
		havingClause = "HAVING " + strings.Join(having, " AND ")
	}

	query := fmt.Sprintf(`
SELECT
  o.id          AS owner_id,
  o.name        AS owner_name,
  o.line_user_id,
  o.lstep_opt_out,
  COALESCE(STRING_AGG(DISTINCT p.name, ',' ORDER BY p.name), '') AS pet_names,
  MAX(mr.date)  AS last_visit_date
FROM owners o
LEFT JOIN pets p ON p.owner_id = o.id AND p.clinic_id = o.clinic_id AND p.deleted_at IS NULL
LEFT JOIN medical_records mr ON mr.owner_id = o.id AND mr.clinic_id = o.clinic_id AND mr.deleted_at IS NULL
WHERE %s
GROUP BY o.id, o.name, o.line_user_id, o.lstep_opt_out
%s
ORDER BY MAX(mr.date) DESC NULLS LAST
`, where, havingClause)

	var rows []CheckupSyncPreviewRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "checkup_sync_preview", "")
	}
	return rows, nil
}
