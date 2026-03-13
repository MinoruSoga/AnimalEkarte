// Package repository provides data access implementations for ExamType entity.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ExaminationType ----

type ExamTypeRepository interface {
	FindAll(ctx context.Context) ([]model.ExamType, error)
	FindByID(ctx context.Context, id uint64) (*model.ExamType, error)
	Create(ctx context.Context, exType *model.ExamType) error
	Update(ctx context.Context, exType *model.ExamType) error
	Delete(ctx context.Context, id uint64) error
}

type examTypeRepository struct{ db *gorm.DB }

func NewExamTypeRepository(db *gorm.DB) ExamTypeRepository {
	return &examTypeRepository{db: db}
}

func (r *examTypeRepository) FindAll(ctx context.Context) ([]model.ExamType, error) {
	var exTypes []model.ExamType
	if err := r.db.WithContext(ctx).Preload("Items").Order("sort_order ASC, name ASC").Find(&exTypes).Error; err != nil {
		return nil, apperrors.Wrap(err, "find examination types")
	}
	return exTypes, nil
}

func (r *examTypeRepository) FindByID(ctx context.Context, id uint64) (*model.ExamType, error) {
	var exType model.ExamType
	if err := r.db.WithContext(ctx).Preload("Items").First(&exType, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("examination_type", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find examination type by id")
	}
	return &exType, nil
}

func (r *examTypeRepository) Create(ctx context.Context, exType *model.ExamType) error {
	if err := r.db.WithContext(ctx).Create(exType).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("examination_type", exType.Name)
		}
		return apperrors.Wrap(err, "create examination type")
	}
	return nil
}

func (r *examTypeRepository) Update(ctx context.Context, exType *model.ExamType) error {
	result := r.db.WithContext(ctx).
		Model(&model.ExamType{}).
		Where("id = ? AND clinic_id = ?", exType.ID, exType.ClinicID).
		Updates(exType)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update examination type")
	}
	if result.RowsAffected == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "update examination type")
	}
	return nil
}

func (r *examTypeRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.ExamType{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete examination type")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("examination_type", fmt.Sprintf("%d", id))
	}
	return nil
}
