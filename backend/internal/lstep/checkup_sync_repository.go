// Checkup sync preview aggregation reads across owners, pets, medical records,
// billings, and checkups without owning a persisted table of its own.
package lstep

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CheckupSyncPreviewRow はプレビュークエリの結果行。
// PetNamesCSV と LivingPetCount は生存ペット（pets.deceased_at IS NULL）のみを集計する。
//
// ISSUE-009 で追加された集計値は CPM ステージ判定（service層 CalculateCPMStage）に必要となる。
// repository は SQL で済む絞り込みを担当し、CPM ステージ絞り込みは service 層で post-filter する。
type CheckupSyncPreviewRow struct {
	OwnerID        uint64     `gorm:"column:owner_id"`
	OwnerName      string     `gorm:"column:owner_name"`
	LineUserID     *string    `gorm:"column:line_user_id"`
	LstepOptOut    bool       `gorm:"column:lstep_opt_out"`
	PetNamesCSV    string     `gorm:"column:pet_names"`
	LivingPetCount int64      `gorm:"column:living_pet_count"`
	LastVisitDate  *time.Time `gorm:"column:last_visit_date"`

	// ISSUE-009: 追加フィルタ／表示用の集計値
	MaxPetAgeYears       *int       `gorm:"column:max_pet_age_years"`       // 生存ペット最大年齢（years）
	MinPetAgeYears       *int       `gorm:"column:min_pet_age_years"`       // 生存ペット最小年齢（years）
	HasChronicCondition  bool       `gorm:"column:has_chronic_condition"`   // 生存ペットにアクティブ慢性疾患があるか
	TotalAmount          int64      `gorm:"column:total_amount"`            // 累計診療費（completed billings）
	AnnualVisitCount     int64      `gorm:"column:annual_visit_count"`      // 過去365日の distinct visit 数
	LastCheckupDate      *time.Time `gorm:"column:last_checkup_date"`       // 最終健診実施日（checkups.date MAX）
	TotalVisitCount      int64      `gorm:"column:total_visit_count"`       // CPM 用：通算 distinct visit 数
	FirstVisitDate       *time.Time `gorm:"column:first_visit_date"`        // CPM 用：初診日
	MaxSingleVisitAmount int64      `gorm:"column:max_single_visit_amount"` // CPM 用：単回最大支払額（cpm_spot 判定）
}

// FindCheckupSyncPreviewParams はプレビュー検索パラメータ。
//
// ISSUE-009: 年齢／慢性疾患／累計診療費／年間来院回数／最終健診実施日の絞り込みを SQL 層で行う。
// CPM ステージは集計値ベースで判定するため service 層で post-filter する。
type FindCheckupSyncPreviewParams struct {
	ClinicID        uint64
	Species         string
	LastVisitBefore *time.Time
	LastVisitAfter  *time.Time

	// ISSUE-009: 追加フィルタ
	MinAgeYears         *int       // 生存ペットの最小年齢（years 以上）
	MaxAgeYears         *int       // 生存ペットの最大年齢（years 以下）
	HasChronicCondition *bool      // true: 慢性疾患あり / false: 慢性疾患なし / nil: 絞らない
	MinTotalAmount      *int64     // 累計診療費（円）以上
	MinAnnualVisitCount *int64     // 年間来院回数（過去365日）以上
	LastCheckupBefore   *time.Time // 最終健診実施日 <= この日
	LastCheckupAfter    *time.Time // 最終健診実施日 >= この日
}

// CheckupSyncPreviewRowLimit bounds SQL result size (BUG-032). Aligns with FE
// CHECKUP_SYNC_OWNER_LIMIT (100) with headroom for post-SQL CPM filtering.
const CheckupSyncPreviewRowLimit = 500

// CheckupSyncRepository は健診同期プレビューのリポジトリインターフェース（BE-004）。
type CheckupSyncRepository interface {
	FindCheckupSyncPreview(ctx context.Context, params *FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error)
}

type checkupSyncRepository struct {
	db *gorm.DB
}

// NewCheckupSyncRepository は CheckupSyncRepository を初期化して返す。
func NewCheckupSyncRepository(db *gorm.DB) CheckupSyncRepository {
	return &checkupSyncRepository{db: db}
}

func (r *checkupSyncRepository) FindCheckupSyncPreview(ctx context.Context, params *FindCheckupSyncPreviewParams) ([]CheckupSyncPreviewRow, error) {
	if params == nil {
		return nil, apperrors.WrapInvalidInput("params is nil")
	}

	// BUG-030: correlated subqueries + pets×medical_records cartesian over full clinic
	// (10k owners / 400k+ MR) timed out before LIMIT. Pre-aggregate per domain CTE, then
	// join/filter/limit owners. Semantics (current pet owner attribution, clinic isolation)
	// match the previous query; only the execution plan changes.
	where, whereArgs := buildCheckupSyncWhere(params)
	filter, filterArgs := buildCheckupSyncPostJoinFilters(params)

	args := make([]any, 0, 8+len(whereArgs)+len(filterArgs))
	// CTE clinic binds (visit/pet/billing/checkup/chronic) — fixed order.
	args = append(args,
		params.ClinicID,
		params.ClinicID,
		params.ClinicID, model.BillingStatusCompleted,
		params.ClinicID,
		params.ClinicID,
	)
	args = append(args, whereArgs...)
	args = append(args, filterArgs...)
	args = append(args, CheckupSyncPreviewRowLimit)

	filterClause := ""
	if len(filter) > 0 {
		filterClause = "AND " + strings.Join(filter, " AND ")
	}

	query := fmt.Sprintf(checkupSyncPreviewSQL, where, filterClause)

	var rows []CheckupSyncPreviewRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "checkup_sync_preview", "")
	}
	return rows, nil
}

// buildCheckupSyncWhere は owners に対する WHERE 句とバインド引数を構築する。
func buildCheckupSyncWhere(params *FindCheckupSyncPreviewParams) (where string, args []any) {
	where = "o.clinic_id = ? AND o.deleted_at IS NULL"
	args = append(args, params.ClinicID)

	if params.Species != "" {
		// ISSUE-005: 種別判定は生存ペットに限定する（死亡ペットの種別だけで対象化しない）。
		where += " AND EXISTS (SELECT 1 FROM pets p2 JOIN animal_species sp ON sp.id = p2.animal_species_id WHERE p2.owner_id = o.id AND p2.clinic_id = o.clinic_id AND p2.deleted_at IS NULL AND p2.deceased_at IS NULL AND LOWER(sp.name) = LOWER(?))"
		args = append(args, params.Species)
	}

	// ISSUE-009: 慢性疾患フィルタ（CTE chronic_owners を利用）。
	if params.HasChronicCondition != nil {
		if *params.HasChronicCondition {
			where += " AND co.owner_id IS NOT NULL"
		} else {
			where += " AND co.owner_id IS NULL"
		}
	}
	return where, args
}

// buildCheckupSyncPostJoinFilters は CTE 結合後に評価する集計フィルタを構築する
// （旧 HAVING 句相当。BUG-030 で cartesian GROUP BY を廃したため WHERE 側へ移動）。
func buildCheckupSyncPostJoinFilters(params *FindCheckupSyncPreviewParams) (filters []string, args []any) {
	if params.LastVisitBefore != nil {
		filters = append(filters, "va.last_visit_date <= ?")
		args = append(args, params.LastVisitBefore.Format(time.DateOnly))
	}
	if params.LastVisitAfter != nil {
		filters = append(filters, "va.last_visit_date >= ?")
		args = append(args, params.LastVisitAfter.Format(time.DateOnly))
	}

	// ISSUE-009: 年齢フィルタ — 生存ペットの年齢レンジで包含判定。
	if params.MinAgeYears != nil {
		filters = append(filters, "COALESCE(pa.max_pet_age_years, -1) >= ?")
		args = append(args, *params.MinAgeYears)
	}
	if params.MaxAgeYears != nil {
		filters = append(filters, "COALESCE(pa.min_pet_age_years, 9999) <= ?")
		args = append(args, *params.MaxAgeYears)
	}
	if params.MinAnnualVisitCount != nil {
		filters = append(filters, "COALESCE(va.annual_visit_count, 0) >= ?")
		args = append(args, *params.MinAnnualVisitCount)
	}
	if params.MinTotalAmount != nil {
		filters = append(filters, "COALESCE(ba.total_amount, 0) >= ?")
		args = append(args, *params.MinTotalAmount)
	}
	if params.LastCheckupBefore != nil {
		filters = append(filters, "ca.last_checkup_date <= ?")
		args = append(args, params.LastCheckupBefore.Format(time.DateOnly))
	}
	if params.LastCheckupAfter != nil {
		filters = append(filters, "ca.last_checkup_date >= ?")
		args = append(args, params.LastCheckupAfter.Format(time.DateOnly))
	}
	return filters, args
}
