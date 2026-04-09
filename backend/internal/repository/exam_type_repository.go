// Package repository provides data access implementations for ExaminationType entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ExaminationType ----

type ExamTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error)
	Create(ctx context.Context, exType *model.ExaminationType) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByExamTypeID(ctx context.Context, examTypeID uint64) (int64, error)
}

type examTypeRepository struct{ db *gorm.DB }

func NewExamTypeRepository(db *gorm.DB) ExamTypeRepository {
	return &examTypeRepository{db: db}
}

func (r *examTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error) {
	exTypes := make([]model.ExaminationType, 0)
	err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).Preload("Items").Order("sort_order ASC, name ASC").Find(&exTypes).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "examination_type", "")
	}
	return exTypes, nil
}

func (r *examTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
	var exType model.ExaminationType
	err := r.db.WithContext(ctx).Preload("Items").First(&exType, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "examination_type", fmt.Sprintf("%d", id))
	}
	return &exType, nil
}

func (r *examTypeRepository) Create(ctx context.Context, exType *model.ExaminationType) error {
	err := r.db.WithContext(ctx).Create(exType).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "examination_type", "")
	}
	return nil
}

func (r *examTypeRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error) {
	result := r.db.WithContext(ctx).
		Model(&model.ExaminationType{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "examination_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("examination_type", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *examTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.ExaminationType{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "examination_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("examination_type", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *examTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.ExaminationType{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "examination_type", fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("examination_type id %d not found in this clinic", id))
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// CountUsageByExamTypeID は検査種別を参照している examination_records の件数を返す（BUG-107）
func (r *examTypeRepository) CountUsageByExamTypeID(ctx context.Context, examTypeID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Examination{}).
		Where("exam_type_id = ?", examTypeID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "examination_record", "")
	}
	return count, nil
}
