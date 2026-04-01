// Package repository provides data access implementations for DiagnosisCategory and DiagnosisName entities.
package repository

import (
	"context"
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
	CountNamesByCategoryID(ctx context.Context, categoryID uint64) (int64, error)
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
		return nil, 0, apperrors.FromGORM(err, "diagnosis_category", "")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "diagnosis_category", "")
	}
	return categories, total, nil
}

func (r *diagnosisCategoryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisCategory, error) {
	var category model.DiagnosisCategory
	err := r.db.WithContext(ctx).
		Preload("Names").
		First(&category, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "diagnosis_category", fmt.Sprintf("%d", id))
	}
	return &category, nil
}

func (r *diagnosisCategoryRepository) Create(ctx context.Context, category *model.DiagnosisCategory) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("diagnosis_category", category.Name)
		}
		return apperrors.FromGORM(err, "diagnosis_category", "")
	}
	return nil
}

func (r *diagnosisCategoryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.DiagnosisCategory{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "diagnosis_category", fmt.Sprintf("%d", id))
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
		return apperrors.FromGORM(result.Error, "diagnosis_category", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("diagnosis_category", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountNamesByCategoryID は指定カテゴリに属する diagnosis_names の件数を返す（BUG-113 補足）
func (r *diagnosisCategoryRepository) CountNamesByCategoryID(ctx context.Context, categoryID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.DiagnosisName{}).
		Where("diagnosis_category_id = ?", categoryID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "diagnosis_name", "")
	}
	return count, nil
}

// Reorder はトランザクション内でカテゴリの sort_order を ids の順序で更新する (#019)
func (r *diagnosisCategoryRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.DiagnosisCategory{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "diagnosis_category", fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("diagnosis_category id %d not found in this clinic", id))
			}
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err, "reorder diagnosis category")
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
	CountClinicalPlansByDiagnosisNameID(ctx context.Context, diagnosisNameID uint64) (int64, error)
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
		return nil, 0, apperrors.FromGORM(err, "diagnosis_name", "")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "diagnosis_name", "")
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
		return nil, 0, apperrors.FromGORM(err, "diagnosis_name", "")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "diagnosis_name", "")
	}
	return names, total, nil
}

func (r *diagnosisNameRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error) {
	var name model.DiagnosisName
	err := r.db.WithContext(ctx).
		First(&name, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "diagnosis_name", fmt.Sprintf("%d", id))
	}
	return &name, nil
}

func (r *diagnosisNameRepository) Create(ctx context.Context, name *model.DiagnosisName) error {
	err := r.db.WithContext(ctx).Create(name).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("diagnosis_name", name.Name)
		}
		return apperrors.FromGORM(err, "diagnosis_name", "")
	}
	return nil
}

func (r *diagnosisNameRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.DiagnosisName{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "diagnosis_name", fmt.Sprintf("%d", id))
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
		return apperrors.FromGORM(result.Error, "diagnosis_name", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("diagnosis_name", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountClinicalPlansByDiagnosisNameID は診断名を参照している clinical_plans の件数を返す（BUG-113）
// diagnosis_name_id および diagnosis_2_name_id 両方をカウントする
func (r *diagnosisNameRepository) CountClinicalPlansByDiagnosisNameID(ctx context.Context, diagnosisNameID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.ClinicalPlan{}).
		Where("diagnosis_name_id = ? OR diagnosis_2_name_id = ?", diagnosisNameID, diagnosisNameID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "clinical_plan", "")
	}
	return count, nil
}

// Reorder はトランザクション内で診断名の sort_order を ids の順序で更新する (#019)
func (r *diagnosisNameRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.DiagnosisName{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "diagnosis_name", fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("diagnosis_name id %d not found in this clinic", id))
			}
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err, "reorder diagnosis name")
	}
	return nil
}
