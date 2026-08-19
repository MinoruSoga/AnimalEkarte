package billing

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// EstimateRepository は見積書のデータアクセスインターフェース
type EstimateRepository interface {
	FindAll(ctx context.Context, clinicID uint64, ownerID, medicalRecordID *uint64, status *string, page, limit int) ([]model.Estimate, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error)
	Create(ctx context.Context, estimate *model.Estimate) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	// UpdateIfNotLocked は status NOT IN (approved, rejected) のときだけ更新する。
	// FindByID→isEstimateLocked→Update の TOCTOU を防ぐ（UpdateIfNotDischarged と同型）。
	// RowsAffected==0（ロック済み・未存在・他院）は Conflict に正規化する。
	UpdateIfNotLocked(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Estimate, error)
	// DeleteIfNotLocked は status NOT IN (approved, rejected) かつ active 明細が 0 のときだけ削除する。
	// FindByID→isEstimateLocked→CountItems→Delete の TOCTOU を防ぐ（UpdateIfNotLocked と同型）。
	// RowsAffected==0 は再読取で locked / 明細あり / NotFound を可能な範囲で区別する。
	DeleteIfNotLocked(ctx context.Context, clinicID, id uint64) error
	// CountItemsByEstimateID は BE-refactor.md R2-5 (D12) で clinic_id 述語を追加した
	// （唯一の呼び出し元 estimateService.Delete が clinicID を既に保持・ownership 検証済み）。
	// estimate_items 自体は clinic_id を持たないため estimates への JOIN で述語を付与する。
	CountItemsByEstimateID(ctx context.Context, clinicID, estimateID uint64) (int64, error)
	// ReplaceItems は estimate_items を一括置換する（soft-delete 既存 → 挿入）。
	ReplaceItems(ctx context.Context, clinicID, estimateID uint64, items []model.EstimateItem) error
	// AllocateNextEstimateNo は clinic スコープで見積番号 EST-{N} を原子的に採番する。
	// 呼び出し元の ambient tx 内で pg_advisory_xact_lock を取り、EST-% の最大数値接尾辞+1 を返す。
	// EST-% が無い場合は max(id)+1 をフォールバックとする（空テーブルは EST-1）。
	// soft-deleted 行も番号占有のため Unscoped で走査する（unique index が deleted_at を見ない）。
	AllocateNextEstimateNo(ctx context.Context, clinicID uint64) (string, error)
}

func estimateNoClinicLockKey(clinicID uint64) string {
	return fmt.Sprintf("estimate_no:clinic:%d", clinicID)
}

type estimateRepository struct {
	db *gorm.DB
}

// NewEstimateRepository はEstimateRepositoryを初期化して返す
func NewEstimateRepository(db *gorm.DB) EstimateRepository {
	return &estimateRepository{db: db}
}

func (r *estimateRepository) FindAll(ctx context.Context, clinicID uint64, ownerID, medicalRecordID *uint64, status *string, page, limit int) ([]model.Estimate, int64, error) {
	estimates := make([]model.Estimate, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Estimate{}).Scopes(persistence.ClinicScope(clinicID))
	if ownerID != nil {
		q = q.Where("owner_id = ?", *ownerID)
	}
	if medicalRecordID != nil {
		q = q.Where("medical_record_id = ?", *medicalRecordID)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "estimate", "")
	}
	// AUD-005: Owner Preload clinic-scoped (callers: List/Create refetch FindByID).
	if err := q.Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted_at IS NULL").Order("sort_order ASC, id ASC")
		}).
		Scopes(persistence.Paginate(page, limit)).Order("created_at DESC").Find(&estimates).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "estimate", "")
	}
	return estimates, total, nil
}

// FindByID は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（UpdateIfNotLocked の tx 内再取得・
// SD-2 系ガード監査の estimateService.Create/Update/Delete が LockByIDForUpdate と同一 tx に
// 収めるため）。
func (r *estimateRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error) {
	var estimate model.Estimate
	err := persistence.DBOrTx(ctx, r.db).
		Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted_at IS NULL").Order("sort_order ASC, id ASC")
		}).
		Preload("CreatedStaff", "deleted_at IS NULL").
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&estimate).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "estimate", fmt.Sprintf("%d", id))
	}
	return &estimate, nil
}

// Create は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（SD-2 系ガード監査: estimateService.Create が
// 確定済みカルテガードのため LockByIDForUpdate と同一 tx に束ねる）。
func (r *estimateRepository) Create(ctx context.Context, estimate *model.Estimate) error {
	err := persistence.DBOrTx(ctx, r.db).Create(estimate).Error
	if err != nil {
		return apperrors.FromGORM(err, "estimate", "")
	}
	return nil
}

func (r *estimateRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, r.db, &model.Estimate{}, "estimate", clinicID, id, fields)
}

// UpdateIfNotLocked は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（SD-2 系ガード監査:
// estimateService.Update が確定済みカルテガードのため LockByIDForUpdate と同一 tx に束ねる）。
func (r *estimateRepository) UpdateIfNotLocked(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Estimate, error) {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Estimate{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND status NOT IN ?", id, []model.EstimateStatus{
			model.EstimateStatusApproved,
			model.EstimateStatusRejected,
		}).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "estimate", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapConflict("承認済みまたは却下済みの見積書は編集できません")
	}
	return r.FindByID(ctx, clinicID, id)
}

// DeleteIfNotLocked は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（SD-2 系ガード監査:
// estimateService.Delete が確定済みカルテガードのため LockByIDForUpdate と同一 tx に束ねる）。
func (r *estimateRepository) DeleteIfNotLocked(ctx context.Context, clinicID, id uint64) error {
	// status ロック条件と active estimate_items=0 を同一 DELETE で要求し、
	// CountItems→Delete 間の明細追加 TOCTOU を原子的に塞ぐ。
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND status NOT IN ?", id, []model.EstimateStatus{
			model.EstimateStatusApproved,
			model.EstimateStatusRejected,
		}).
		Where(`NOT EXISTS (
			SELECT 1 FROM estimate_items ei
			WHERE ei.estimate_id = estimates.id
			  AND ei.deleted_at IS NULL
		)`).
		Delete(&model.Estimate{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "estimate", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeDeleteIfNotLockedMiss(ctx, clinicID, id)
	}
	return nil
}

// normalizeDeleteIfNotLockedMiss は原子 DELETE が 0 行だった理由を再読取で区別する。
func (r *estimateRepository) normalizeDeleteIfNotLockedMiss(ctx context.Context, clinicID, id uint64) error {
	existing, err := r.FindByID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if existing.Status == model.EstimateStatusApproved || existing.Status == model.EstimateStatusRejected {
		return apperrors.WrapConflict("承認済みまたは却下済みの見積書は削除できません")
	}
	count, err := r.CountItemsByEstimateID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return apperrors.WrapConflict("この見積書には明細が登録されているため削除できません")
	}
	// DELETE と再読取の間に status がロックへ遷移した等の残余レース
	return apperrors.WrapConflict("承認済みまたは却下済みの見積書は削除できません")
}

// ReplaceItems は親見積を FOR UPDATE で固定してから明細を置換する。
// estimate_items は clinic_id を持たないため、親行の clinic スコープ＋行ロックで越境と TOCTOU を防ぐ。
// FOR UPDATE 依存のため ambient tx 不在は fail-closed。
func (r *estimateRepository) ReplaceItems(ctx context.Context, clinicID, estimateID uint64, items []model.EstimateItem) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("estimate item replace requires an active transaction")
	}
	db := persistence.DBOrTx(ctx, r.db)
	var parent model.Estimate
	if err := db.Model(&model.Estimate{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Scopes(persistence.ClinicScope(clinicID)).
		Select("id").
		Where("id = ?", estimateID).
		First(&parent).Error; err != nil {
		return apperrors.FromGORM(err, "estimate", fmt.Sprintf("%d", estimateID))
	}
	if err := db.Where("estimate_id = ?", estimateID).Delete(&model.EstimateItem{}).Error; err != nil {
		return apperrors.FromGORM(err, "estimate_item", "")
	}
	if len(items) == 0 {
		return nil
	}
	for i := range items {
		items[i].EstimateID = estimateID
		items[i].ID = 0
	}
	if err := db.Create(&items).Error; err != nil {
		return apperrors.FromGORM(err, "estimate_item", "")
	}
	return nil
}

func (r *estimateRepository) CountItemsByEstimateID(ctx context.Context, clinicID, estimateID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.EstimateItem{}).
		Joins("JOIN estimates ON estimates.id = estimate_items.estimate_id AND estimates.clinic_id = ? AND estimates.deleted_at IS NULL", clinicID).
		Where("estimate_items.estimate_id = ? AND estimate_items.deleted_at IS NULL", estimateID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "estimate_item", "")
	}
	return count, nil
}

// AllocateNextEstimateNo は clinic スコープで次の EST-{N} を採番する（ambient tx + advisory xact lock）。
func (r *estimateRepository) AllocateNextEstimateNo(ctx context.Context, clinicID uint64) (string, error) {
	db := persistence.DBOrTx(ctx, r.db)
	if err := db.Exec(
		"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))",
		estimateNoClinicLockKey(clinicID),
	).Error; err != nil {
		return "", apperrors.Wrap(err, "failed to acquire estimate number lock")
	}

	// soft-deleted も unique index 上は番号を占有するため Unscoped。
	var nos []string
	if err := db.Unscoped().
		Model(&model.Estimate{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("estimate_no LIKE ?", "EST-%").
		Pluck("estimate_no", &nos).Error; err != nil {
		return "", apperrors.FromGORM(err, "estimate", "")
	}

	var maxSuffix int64
	foundEST := false
	for _, no := range nos {
		suffix := strings.TrimPrefix(no, "EST-")
		if suffix == no {
			continue
		}
		n, err := strconv.ParseInt(suffix, 10, 64)
		if err != nil || n < 0 {
			continue
		}
		foundEST = true
		if n > maxSuffix {
			maxSuffix = n
		}
	}

	next := maxSuffix + 1
	if !foundEST {
		// EST-% が無い場合: max(id)+1 をフォールバック（空なら 1）。
		var maxID int64
		if err := db.Unscoped().
			Model(&model.Estimate{}).
			Scopes(persistence.ClinicScope(clinicID)).
			Select("COALESCE(MAX(id), 0)").
			Scan(&maxID).Error; err != nil {
			return "", apperrors.FromGORM(err, "estimate", "")
		}
		next = maxID + 1
		if next < 1 {
			next = 1
		}
	}

	return fmt.Sprintf("EST-%d", next), nil
}
