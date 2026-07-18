// Package checkuptype owns checkup_types data access (BE8-4 batch5 — leaf domain).
package checkuptype

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for checkup types.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.CheckupType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error)
	Create(ctx context.Context, checkupType *model.CheckupType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.CheckupType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64) ([]model.CheckupType, error) {
	checkupTypes := make([]model.CheckupType, 0)
	err := r.db.WithContext(ctx).Scopes(repohelpers.ClinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&checkupTypes).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup_type", "")
	}
	return checkupTypes, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error) {
	return repohelpers.FindByIDScoped[model.CheckupType](ctx, r.db, "checkup_type", clinicID, id)
}

func (r *repository) Create(ctx context.Context, checkupType *model.CheckupType) error {
	err := r.db.WithContext(ctx).Create(checkupType).Error
	if err != nil {
		return apperrors.FromGORM(err, "checkup_type", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.CheckupType, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.CheckupType{}, "checkup_type", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.CheckupType{}, "checkup_type", clinicID, id)
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.CheckupType{}, "checkup_type", clinicID, ids, "sort_order")
}

// CountUsageByCheckupTypeID は定期健診種別を参照している checkups の件数を返す（BUG-107）
// checkups テーブルは直接 clinic_id を持つためテナント分離を直接適用する
func (r *repository) CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Checkup{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("checkup_type_id = ? AND deleted_at IS NULL", checkupTypeID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "checkup_type", "")
	}
	return count, nil
}

// CountChildrenByParentID は指定した親 ID を持つ子健診種別の件数を返す。
func (r *repository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.CheckupType{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "checkup_type", "")
	}
	return count, nil
}
