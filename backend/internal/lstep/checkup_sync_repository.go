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
	// C-1: SELECT 句の total_amount / max_single_visit_amount 副問い合わせの2箇所は
	// WHERE 句より前に query テキスト中へ現れるため、他の args より先にバインドする。
	where, whereArgs := buildCheckupSyncWhere(params)
	having, havingArgs := buildCheckupSyncHaving(params)
	args := make([]any, 0, 2+len(whereArgs)+len(havingArgs))
	args = append(args, model.BillingStatusCompleted, model.BillingStatusCompleted)
	args = append(args, whereArgs...)
	args = append(args, havingArgs...)

	havingClause := ""
	if len(having) > 0 {
		havingClause = "HAVING " + strings.Join(having, " AND ")
	}

	// ISSUE-005:
	//   - PetNames / LivingPetCount は生存ペット（deceased_at IS NULL）のみ集計する。
	//   - 死亡ペットのみの飼い主は LivingPetCount=0 でレコードが返り、サービス層で「生存ペットなし」として除外理由を付与する。
	// ISSUE-009:
	//   - 年齢/総診療費/年間来院/最終健診/CPM 判定材料を SELECT に含める。
	//   - 累計診療費・単回最大診療費・最終健診日 はスカラー副問い合わせで集約（pets LEFT JOIN による cartesian inflation 回避）。
	query := fmt.Sprintf(`
SELECT
  o.id          AS owner_id,
  o.name        AS owner_name,
  o.line_user_id,
  o.lstep_opt_out,
  COALESCE(STRING_AGG(DISTINCT p.name, ',' ORDER BY p.name) FILTER (WHERE p.deceased_at IS NULL), '') AS pet_names,
  COUNT(DISTINCT p.id) FILTER (WHERE p.deceased_at IS NULL) AS living_pet_count,
  MAX(mr.date)  AS last_visit_date,
  MIN(mr.date)  AS first_visit_date,
  COUNT(DISTINCT mr.date) AS total_visit_count,
  COUNT(DISTINCT CASE WHEN mr.date >= NOW() - INTERVAL '365 days' THEN mr.date END) AS annual_visit_count,
  MIN(EXTRACT(YEAR FROM AGE(NOW(), p.birth_date))::int) FILTER (WHERE p.deceased_at IS NULL AND p.birth_date IS NOT NULL) AS min_pet_age_years,
  MAX(EXTRACT(YEAR FROM AGE(NOW(), p.birth_date))::int) FILTER (WHERE p.deceased_at IS NULL AND p.birth_date IS NOT NULL) AS max_pet_age_years,
  EXISTS (
    SELECT 1 FROM pet_chronic_conditions pcc
    INNER JOIN pets pp ON pp.id = pcc.pet_id AND pp.deleted_at IS NULL AND pp.deceased_at IS NULL
    WHERE pp.owner_id = o.id AND pp.clinic_id = o.clinic_id AND pcc.clinic_id = o.clinic_id
      AND pcc.is_active = TRUE AND pcc.deleted_at IS NULL
  ) AS has_chronic_condition,
  COALESCE((
    SELECT SUM(b.total_amount) FROM billings b
    WHERE b.clinic_id = o.clinic_id AND b.owner_id = o.id
      AND b.status = ? AND b.deleted_at IS NULL
  ), 0) AS total_amount,
  COALESCE((
    SELECT MAX(b2.total_amount) FROM billings b2
    WHERE b2.clinic_id = o.clinic_id AND b2.owner_id = o.id
      AND b2.status = ? AND b2.deleted_at IS NULL
  ), 0) AS max_single_visit_amount,
  (
    SELECT MAX(c.date) FROM checkups c
    INNER JOIN medical_records mrc ON mrc.id = c.medical_record_id AND mrc.deleted_at IS NULL
    WHERE c.clinic_id = o.clinic_id AND c.deleted_at IS NULL
      AND mrc.clinic_id = o.clinic_id
      AND EXISTS (
        SELECT 1 FROM pets current_checkup_owner_pet
        WHERE current_checkup_owner_pet.id = mrc.pet_id
          AND current_checkup_owner_pet.clinic_id = mrc.clinic_id
          AND current_checkup_owner_pet.owner_id = o.id
      )
  ) AS last_checkup_date
FROM owners o
LEFT JOIN pets p ON p.owner_id = o.id AND p.clinic_id = o.clinic_id AND p.deleted_at IS NULL
LEFT JOIN medical_records mr
  ON mr.clinic_id = o.clinic_id
 AND mr.deleted_at IS NULL
 AND EXISTS (
   SELECT 1 FROM pets current_visit_owner_pet
   WHERE current_visit_owner_pet.id = mr.pet_id
     AND current_visit_owner_pet.clinic_id = mr.clinic_id
     AND current_visit_owner_pet.owner_id = o.id
 )
WHERE %s
GROUP BY o.id, o.name, o.line_user_id, o.lstep_opt_out
%s
ORDER BY MAX(mr.date) DESC NULLS LAST
LIMIT ?
`, where, havingClause)
	args = append(args, CheckupSyncPreviewRowLimit)

	var rows []CheckupSyncPreviewRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "checkup_sync_preview", "")
	}
	return rows, nil
}

// buildCheckupSyncWhere は WHERE 句の条件文字列とバインド引数を構築する
// （BE-refactor.md E-13: FindCheckupSyncPreview の位置引数結合を隔離する純粋抽出）。
func buildCheckupSyncWhere(params *FindCheckupSyncPreviewParams) (where string, args []any) {
	where = "o.clinic_id = ? AND o.deleted_at IS NULL"
	args = append(args, params.ClinicID)

	if params.Species != "" {
		// ISSUE-005: 種別判定は生存ペットに限定する（死亡ペットの種別だけで対象化しない）。
		where += " AND EXISTS (SELECT 1 FROM pets p2 JOIN animal_species sp ON sp.id = p2.animal_species_id WHERE p2.owner_id = o.id AND p2.clinic_id = o.clinic_id AND p2.deleted_at IS NULL AND p2.deceased_at IS NULL AND LOWER(sp.name) = LOWER(?))"
		args = append(args, params.Species)
	}

	// ISSUE-009: 慢性疾患フィルタは飼い主単位の EXISTS で WHERE 評価する。
	if params.HasChronicCondition != nil {
		if *params.HasChronicCondition {
			where += " AND EXISTS (SELECT 1 FROM pet_chronic_conditions pcc INNER JOIN pets pp ON pp.id = pcc.pet_id AND pp.deleted_at IS NULL AND pp.deceased_at IS NULL WHERE pp.owner_id = o.id AND pp.clinic_id = o.clinic_id AND pcc.clinic_id = o.clinic_id AND pcc.is_active = TRUE AND pcc.deleted_at IS NULL)"
		} else {
			where += " AND NOT EXISTS (SELECT 1 FROM pet_chronic_conditions pcc INNER JOIN pets pp ON pp.id = pcc.pet_id AND pp.deleted_at IS NULL AND pp.deceased_at IS NULL WHERE pp.owner_id = o.id AND pp.clinic_id = o.clinic_id AND pcc.clinic_id = o.clinic_id AND pcc.is_active = TRUE AND pcc.deleted_at IS NULL)"
		}
	}
	return where, args
}

// buildCheckupSyncHaving は HAVING 句の条件断片とバインド引数を構築する（BE-refactor.md E-13）。
func buildCheckupSyncHaving(params *FindCheckupSyncPreviewParams) (having []string, args []any) {
	if params.LastVisitBefore != nil {
		having = append(having, "MAX(mr.date) <= ?")
		args = append(args, params.LastVisitBefore.Format(time.DateOnly))
	}
	if params.LastVisitAfter != nil {
		having = append(having, "MAX(mr.date) >= ?")
		args = append(args, params.LastVisitAfter.Format(time.DateOnly))
	}

	// ISSUE-009: 年齢フィルタ — 生存ペットの年齢レンジで包含判定（誰か1匹でもレンジに該当すれば対象）。
	// COALESCE で「生存ペット無し / 誕生日未登録」owner も評価できるようにする（マッチしないようなセンチネル値を使う）。
	if params.MinAgeYears != nil {
		having = append(having, "COALESCE(MAX(EXTRACT(YEAR FROM AGE(NOW(), p.birth_date))::int) FILTER (WHERE p.deceased_at IS NULL AND p.birth_date IS NOT NULL), -1) >= ?")
		args = append(args, *params.MinAgeYears)
	}
	if params.MaxAgeYears != nil {
		having = append(having, "COALESCE(MIN(EXTRACT(YEAR FROM AGE(NOW(), p.birth_date))::int) FILTER (WHERE p.deceased_at IS NULL AND p.birth_date IS NOT NULL), 9999) <= ?")
		args = append(args, *params.MaxAgeYears)
	}
	// ISSUE-009: 年間来院回数フィルタ（過去365日の distinct visit 数）。
	if params.MinAnnualVisitCount != nil {
		having = append(having, "COUNT(DISTINCT CASE WHEN mr.date >= NOW() - INTERVAL '365 days' THEN mr.date END) >= ?")
		args = append(args, *params.MinAnnualVisitCount)
	}
	// ISSUE-009: 累計診療費フィルタ — billings 集計はスカラー副問い合わせで取得（pets cartesian の影響を回避）。
	if params.MinTotalAmount != nil {
		having = append(having, "COALESCE((SELECT SUM(bf.total_amount) FROM billings bf WHERE bf.clinic_id = o.clinic_id AND bf.owner_id = o.id AND bf.status = ? AND bf.deleted_at IS NULL), 0) >= ?")
		args = append(args, model.BillingStatusCompleted, *params.MinTotalAmount)
	}
	// ISSUE-009: 最終健診実施日フィルタ。checkups は medical_records 経由で owner と紐づける。
	if params.LastCheckupBefore != nil {
		having = append(having, "(SELECT MAX(cf.date) FROM checkups cf INNER JOIN medical_records mrf ON mrf.id = cf.medical_record_id AND mrf.deleted_at IS NULL WHERE cf.clinic_id = o.clinic_id AND cf.deleted_at IS NULL AND mrf.clinic_id = o.clinic_id AND EXISTS (SELECT 1 FROM pets current_checkup_owner_pet WHERE current_checkup_owner_pet.id = mrf.pet_id AND current_checkup_owner_pet.clinic_id = mrf.clinic_id AND current_checkup_owner_pet.owner_id = o.id)) <= ?")
		args = append(args, params.LastCheckupBefore.Format(time.DateOnly))
	}
	if params.LastCheckupAfter != nil {
		having = append(having, "(SELECT MAX(cf.date) FROM checkups cf INNER JOIN medical_records mrf ON mrf.id = cf.medical_record_id AND mrf.deleted_at IS NULL WHERE cf.clinic_id = o.clinic_id AND cf.deleted_at IS NULL AND mrf.clinic_id = o.clinic_id AND EXISTS (SELECT 1 FROM pets current_checkup_owner_pet WHERE current_checkup_owner_pet.id = mrf.pet_id AND current_checkup_owner_pet.clinic_id = mrf.clinic_id AND current_checkup_owner_pet.owner_id = o.id)) >= ?")
		args = append(args, params.LastCheckupAfter.Format(time.DateOnly))
	}
	return having, args
}
