// Package repository provides data access implementations for DiagnosisType and DiagnosisName entities.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- DiagnosisType ----

type DiagnosisTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error)
	Create(ctx context.Context, category *model.DiagnosisType) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountNamesByCategoryID(ctx context.Context, categoryID uint64) (int64, error)
}

type diagnosisTypeRepository struct{ db *gorm.DB }

func NewDiagnosisTypeRepository(db *gorm.DB) DiagnosisTypeRepository {
	return &diagnosisTypeRepository{db: db}
}

func (r *diagnosisTypeRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, error) {
	categories := make([]model.DiagnosisType, 0)
	if err := r.db.WithContext(ctx).Model(&model.DiagnosisType{}).Scopes(clinicScope(clinicID)).
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, apperrors.FromGORM(err, "diagnosis_type", "")
	}
	return categories, nil
}

func (r *diagnosisTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error) {
	var category model.DiagnosisType
	err := r.db.WithContext(ctx).
		Preload("Names").
		Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "diagnosis_type", fmt.Sprintf("%d", id))
	}
	return &category, nil
}

func (r *diagnosisTypeRepository) Create(ctx context.Context, category *model.DiagnosisType) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "diagnosis_type", "")
	}
	return nil
}

func (r *diagnosisTypeRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisType, error) {
	result := r.db.WithContext(ctx).
		Model(&model.DiagnosisType{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "diagnosis_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("diagnosis_type", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *diagnosisTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Delete(&model.DiagnosisType{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "diagnosis_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("diagnosis_type", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountNamesByCategoryID は指定カテゴリに属する diagnosis_names の件数を返す（BUG-113 補足）
func (r *diagnosisTypeRepository) CountNamesByCategoryID(ctx context.Context, categoryID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.DiagnosisName{}).
		Where("diagnosis_type_id = ?", categoryID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "diagnosis_name", "")
	}
	return count, nil
}

// Reorder はトランザクション内でカテゴリの sort_order を ids の順序で更新する (#019)
func (r *diagnosisTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.DiagnosisType{}, "diagnosis_type", clinicID, ids)
}

// ---- DiagnosisName ----

type DiagnosisNameRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, error)
	FindByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error)
	Create(ctx context.Context, name *model.DiagnosisName) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisName, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountClinicalPlansByDiagnosisNameID(ctx context.Context, diagnosisNameID uint64) (int64, error)
}

type diagnosisNameRepository struct{ db *gorm.DB }

func NewDiagnosisNameRepository(db *gorm.DB) DiagnosisNameRepository {
	return &diagnosisNameRepository{db: db}
}

func (r *diagnosisNameRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, error) {
	names := make([]model.DiagnosisName, 0)
	if err := r.db.WithContext(ctx).Model(&model.DiagnosisName{}).Scopes(clinicScope(clinicID)).
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, apperrors.FromGORM(err, "diagnosis_name", "")
	}
	return names, nil
}

func (r *diagnosisNameRepository) FindByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, error) {
	names := make([]model.DiagnosisName, 0)
	if err := r.db.WithContext(ctx).Model(&model.DiagnosisName{}).
		Scopes(clinicScope(clinicID)).
		Where("diagnosis_type_id = ?", categoryID).
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, apperrors.FromGORM(err, "diagnosis_name", "")
	}
	return names, nil
}

func (r *diagnosisNameRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error) {
	var name model.DiagnosisName
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&name).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "diagnosis_name", fmt.Sprintf("%d", id))
	}
	return &name, nil
}

func (r *diagnosisNameRepository) Create(ctx context.Context, name *model.DiagnosisName) error {
	err := r.db.WithContext(ctx).Create(name).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "diagnosis_name", "")
	}
	return nil
}

func (r *diagnosisNameRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisName, error) {
	result := r.db.WithContext(ctx).
		Model(&model.DiagnosisName{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "diagnosis_name", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("diagnosis_name", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *diagnosisNameRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
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
	return reorderByClinicID(ctx, r.db, &model.DiagnosisName{}, "diagnosis_name", clinicID, ids)
}
