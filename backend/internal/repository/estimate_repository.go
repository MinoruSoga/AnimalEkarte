package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
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
	Delete(ctx context.Context, clinicID, id uint64) error
	// CountItemsByEstimateID は BE-refactor.md R2-5 (D12) で clinic_id 述語を追加した
	// （唯一の呼び出し元 estimateService.Delete が clinicID を既に保持・ownership 検証済み）。
	// estimate_items 自体は clinic_id を持たないため estimates への JOIN で述語を付与する。
	CountItemsByEstimateID(ctx context.Context, clinicID, estimateID uint64) (int64, error)
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

	q := r.db.WithContext(ctx).Model(&model.Estimate{}).Scopes(clinicScope(clinicID))
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
		Preload("Items", "deleted_at IS NULL").
		Scopes(paginate(page, limit)).Order("created_at DESC").Find(&estimates).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "estimate", "")
	}
	return estimates, total, nil
}

func (r *estimateRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error) {
	var estimate model.Estimate
	err := r.db.WithContext(ctx).
		Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Items", "deleted_at IS NULL").
		Preload("CreatedStaff", "deleted_at IS NULL").
		Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&estimate).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "estimate", fmt.Sprintf("%d", id))
	}
	return &estimate, nil
}

func (r *estimateRepository) Create(ctx context.Context, estimate *model.Estimate) error {
	err := r.db.WithContext(ctx).Create(estimate).Error
	if err != nil {
		return apperrors.FromGORM(err, "estimate", "")
	}
	return nil
}

func (r *estimateRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return updateScopedByID(ctx, r.db, &model.Estimate{}, "estimate", clinicID, id, fields)
}

func (r *estimateRepository) UpdateIfNotLocked(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Estimate, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Estimate{}).
		Scopes(clinicScope(clinicID)).
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

func (r *estimateRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return deleteScopedByID(ctx, r.db, &model.Estimate{}, "estimate", clinicID, id)
}

// CountItemsByEstimateID は見積書に紐付く明細行の件数を返す（BUG-201）
// BE-refactor.md R2-5 (D12): clinic_id 述語を追加。estimate_items は clinic_id カラムを持たないため
// estimates への JOIN で述語を付与する。
func (r *estimateRepository) CountItemsByEstimateID(ctx context.Context, clinicID, estimateID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.EstimateItem{}).
		Joins("JOIN estimates ON estimates.id = estimate_items.estimate_id AND estimates.clinic_id = ? AND estimates.deleted_at IS NULL", clinicID).
		Where("estimate_items.estimate_id = ? AND estimate_items.deleted_at IS NULL", estimateID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "estimate_item", "")
	}
	return count, nil
}
