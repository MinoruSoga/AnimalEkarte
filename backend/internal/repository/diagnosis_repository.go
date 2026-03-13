// Package repository provides data access implementations for DiagnosisCategory and DiagnosisName entities.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- DiagnosisCategory ----

type DiagnosisCategoryRepository interface {
	FindAll(ctx context.Context) ([]model.DiagnosisCategory, error)
	FindByID(ctx context.Context, id uint64) (*model.DiagnosisCategory, error)
	Create(ctx context.Context, category *model.DiagnosisCategory) error
	Update(ctx context.Context, category *model.DiagnosisCategory) error
	Delete(ctx context.Context, id uint64) error
}

type diagnosisCategoryRepository struct{ db *gorm.DB }

func NewDiagnosisCategoryRepository(db *gorm.DB) DiagnosisCategoryRepository {
	return &diagnosisCategoryRepository{db: db}
}

func (r *diagnosisCategoryRepository) FindAll(ctx context.Context) ([]model.DiagnosisCategory, error) {
	var categories []model.DiagnosisCategory
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, apperrors.Wrap(err, "find diagnosis categories")
	}
	return categories, nil
}

func (r *diagnosisCategoryRepository) FindByID(ctx context.Context, id uint64) (*model.DiagnosisCategory, error) {
	var category model.DiagnosisCategory
	if err := r.db.WithContext(ctx).Preload("Names").First(&category, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("diagnosis_category", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find diagnosis category by id")
	}
	return &category, nil
}

func (r *diagnosisCategoryRepository) Create(ctx context.Context, category *model.DiagnosisCategory) error {
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("diagnosis_category", category.Name)
		}
		return apperrors.Wrap(err, "create diagnosis category")
	}
	return nil
}

func (r *diagnosisCategoryRepository) Update(ctx context.Context, category *model.DiagnosisCategory) error {
	if err := r.db.WithContext(ctx).Save(category).Error; err != nil {
		return apperrors.Wrap(err, "update diagnosis category")
	}
	return nil
}

func (r *diagnosisCategoryRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.DiagnosisCategory{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete diagnosis category")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("diagnosis_category", fmt.Sprintf("%d", id))
	}
	return nil
}

// ---- DiagnosisName ----

type DiagnosisNameRepository interface {
	FindAll(ctx context.Context) ([]model.DiagnosisName, error)
	FindByCategoryID(ctx context.Context, categoryID uint64) ([]model.DiagnosisName, error)
	FindByID(ctx context.Context, id uint64) (*model.DiagnosisName, error)
	Create(ctx context.Context, name *model.DiagnosisName) error
	Update(ctx context.Context, name *model.DiagnosisName) error
	Delete(ctx context.Context, id uint64) error
}

type diagnosisNameRepository struct{ db *gorm.DB }

func NewDiagnosisNameRepository(db *gorm.DB) DiagnosisNameRepository {
	return &diagnosisNameRepository{db: db}
}

func (r *diagnosisNameRepository) FindAll(ctx context.Context) ([]model.DiagnosisName, error) {
	var names []model.DiagnosisName
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&names).Error; err != nil {
		return nil, apperrors.Wrap(err, "find diagnosis names")
	}
	return names, nil
}

func (r *diagnosisNameRepository) FindByCategoryID(ctx context.Context, categoryID uint64) ([]model.DiagnosisName, error) {
	var names []model.DiagnosisName
	if err := r.db.WithContext(ctx).
		Where("diagnosis_category_id = ?", categoryID).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, apperrors.Wrap(err, "find diagnosis names by category id")
	}
	return names, nil
}

func (r *diagnosisNameRepository) FindByID(ctx context.Context, id uint64) (*model.DiagnosisName, error) {
	var name model.DiagnosisName
	if err := r.db.WithContext(ctx).First(&name, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("diagnosis_name", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find diagnosis name by id")
	}
	return &name, nil
}

func (r *diagnosisNameRepository) Create(ctx context.Context, name *model.DiagnosisName) error {
	if err := r.db.WithContext(ctx).Create(name).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("diagnosis_name", name.Name)
		}
		return apperrors.Wrap(err, "create diagnosis name")
	}
	return nil
}

func (r *diagnosisNameRepository) Update(ctx context.Context, name *model.DiagnosisName) error {
	if err := r.db.WithContext(ctx).Save(name).Error; err != nil {
		return apperrors.Wrap(err, "update diagnosis name")
	}
	return nil
}

func (r *diagnosisNameRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.DiagnosisName{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete diagnosis name")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("diagnosis_name", fmt.Sprintf("%d", id))
	}
	return nil
}
