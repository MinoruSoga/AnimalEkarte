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
	CountUsageByCheckupTypeID(ctx context.Context, checkupTypeID uint64) (int64, error)
}

type checkupTypeRepository struct{ db *gorm.DB }

func NewCheckupTypeRepository(db *gorm.DB) CheckupTypeRepository {
	return &checkupTypeRepository{db: db}
}

func (r *checkupTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.CheckupType, error) {
	checkupTypes := make([]model.CheckupType, 0)
	err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).Order("sort_order ASC, name ASC").Find(&checkupTypes).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup_type", "")
	}
	return checkupTypes, nil
}

func (r *checkupTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error) {
	var checkupType model.CheckupType
	err := r.db.WithContext(ctx).First(&checkupType, "id = ? AND clinic_id = ?", id, clinicID).Error
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
		Where("id = ? AND clinic_id = ?", id, clinicID).
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
	result := r.db.WithContext(ctx).Delete(&model.CheckupType{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "checkup_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("checkup_type", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *checkupTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.CheckupType{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "checkup_type", fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("checkup_type id %d not found in this clinic", id))
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// CountUsageByCheckupTypeID は定期健診種別を参照している checkup_records の件数を返す（BUG-107）
func (r *checkupTypeRepository) CountUsageByCheckupTypeID(ctx context.Context, checkupTypeID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Checkup{}).
		Where("checkup_type_id = ?", checkupTypeID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "checkup_record", "")
	}
	return count, nil
}
