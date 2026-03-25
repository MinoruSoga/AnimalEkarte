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
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisCategory, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisCategory, error)
	Create(ctx context.Context, category *model.DiagnosisCategory) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type diagnosisCategoryRepository struct{ db *gorm.DB }

func NewDiagnosisCategoryRepository(db *gorm.DB) DiagnosisCategoryRepository {
	return &diagnosisCategoryRepository{db: db}
}

func (r *diagnosisCategoryRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisCategory, int64, error) {
	categories := make([]model.DiagnosisCategory, 0)
	var total int64

	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&model.DiagnosisCategory{}).Where("clinic_id = ?", clinicID)
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count diagnosis categories")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find diagnosis categories")
	}
	return categories, total, nil
}

func (r *diagnosisCategoryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisCategory, error) {
	var category model.DiagnosisCategory
	if err := r.db.WithContext(ctx).
		Preload("Names").
		First(&category, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
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

func (r *diagnosisCategoryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.DiagnosisCategory{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update diagnosis category")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("diagnosis_category", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *diagnosisCategoryRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Delete(&model.DiagnosisCategory{})
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete diagnosis category")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("diagnosis_category", fmt.Sprintf("%d", id))
	}
	return nil
}

// Reorder はトランザクション内でカテゴリの sort_order を ids の順序で更新する (#019)
func (r *diagnosisCategoryRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.DiagnosisCategory{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, "reorder diagnosis category")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("diagnosis_category id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("reorder diagnosis category: %w", err)
	}
	return nil
}

// ---- DiagnosisName ----

type DiagnosisNameRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	FindByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error)
	Create(ctx context.Context, name *model.DiagnosisName) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type diagnosisNameRepository struct{ db *gorm.DB }

func NewDiagnosisNameRepository(db *gorm.DB) DiagnosisNameRepository {
	return &diagnosisNameRepository{db: db}
}

func (r *diagnosisNameRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	names := make([]model.DiagnosisName, 0)
	var total int64

	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&model.DiagnosisName{}).Where("clinic_id = ?", clinicID)
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count diagnosis names")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find diagnosis names")
	}
	return names, total, nil
}

func (r *diagnosisNameRepository) FindByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	names := make([]model.DiagnosisName, 0)
	var total int64

	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&model.DiagnosisName{}).
			Where("clinic_id = ? AND diagnosis_category_id = ?", clinicID, categoryID)
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count diagnosis names by category id")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find diagnosis names by category id")
	}
	return names, total, nil
}

func (r *diagnosisNameRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error) {
	var name model.DiagnosisName
	if err := r.db.WithContext(ctx).
		First(&name, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
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

func (r *diagnosisNameRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.DiagnosisName{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update diagnosis name")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("diagnosis_name", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *diagnosisNameRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Delete(&model.DiagnosisName{})
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete diagnosis name")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("diagnosis_name", fmt.Sprintf("%d", id))
	}
	return nil
}

// Reorder はトランザクション内で診断名の sort_order を ids の順序で更新する (#019)
func (r *diagnosisNameRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.DiagnosisName{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, "reorder diagnosis name")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("diagnosis_name id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("reorder diagnosis name: %w", err)
	}
	return nil
}
