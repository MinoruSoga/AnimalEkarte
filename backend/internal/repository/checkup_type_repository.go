// Package repository provides data access implementations for CheckupType entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- CheckupType ----

type CheckupTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.CheckupType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error)
	Create(ctx context.Context, checkupType *model.CheckupType) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.CheckupType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type checkupTypeRepository struct{ db *gorm.DB }

func NewCheckupTypeRepository(db *gorm.DB) CheckupTypeRepository {
	return &checkupTypeRepository{db: db}
}

func (r *checkupTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.CheckupType, error) {
	checkupTypes := make([]model.CheckupType, 0)
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&checkupTypes).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup_type", "")
	}
	return checkupTypes, nil
}

func (r *checkupTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error) {
	var checkupType model.CheckupType
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&checkupType).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup_type", fmt.Sprintf("%d", id))
	}
	return &checkupType, nil
}

func (r *checkupTypeRepository) Create(ctx context.Context, checkupType *model.CheckupType) error {
	err := r.db.WithContext(ctx).Create(checkupType).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "checkup_type", "")
	}
	return nil
}

func (r *checkupTypeRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.CheckupType, error) {
	result := r.db.WithContext(ctx).
		Model(&model.CheckupType{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "checkup_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("checkup_type", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *checkupTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.CheckupType{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "checkup_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("checkup_type", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *checkupTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.CheckupType{}, "checkup_type", clinicID, ids)
}

// CountUsageByCheckupTypeID は定期健診種別を参照している checkups の件数を返す（BUG-107）
// checkups テーブルは直接 clinic_id を持つためテナント分離を直接適用する
func (r *checkupTypeRepository) CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Checkup{}).
		Where("checkup_type_id = ? AND clinic_id = ?", checkupTypeID, clinicID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "checkup_record", "")
	}
	return count, nil
}

// CountChildrenByParentID は指定した親 ID を持つ子健診種別の件数を返す。
func (r *checkupTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.CheckupType{}).
		Scopes(clinicScope(clinicID)).
		Where("parent_id = ?", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "checkup_type", "")
	}
	return count, nil
}
