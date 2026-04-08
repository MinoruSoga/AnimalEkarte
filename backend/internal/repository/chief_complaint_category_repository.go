// Package repository provides data access implementations for ChiefComplaintCategory entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ChiefComplaintCategory ----

type ChiefComplaintCategoryRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintCategory, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintCategory, error)
	Create(ctx context.Context, category *model.ChiefComplaintCategory) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type chiefComplaintCategoryRepository struct{ db *gorm.DB }

func NewChiefComplaintCategoryRepository(db *gorm.DB) ChiefComplaintCategoryRepository {
	return &chiefComplaintCategoryRepository{db: db}
}

func (r *chiefComplaintCategoryRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintCategory, error) {
	categories := make([]model.ChiefComplaintCategory, 0)
	err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "chief_complaint_category", "")
	}
	return categories, nil
}

func (r *chiefComplaintCategoryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintCategory, error) {
	var category model.ChiefComplaintCategory
	err := r.db.WithContext(ctx).First(&category, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "chief_complaint_category", fmt.Sprintf("%d", id))
	}
	return &category, nil
}

func (r *chiefComplaintCategoryRepository) Create(ctx context.Context, category *model.ChiefComplaintCategory) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "chief_complaint_category", "")
	}
	return nil
}

func (r *chiefComplaintCategoryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ChiefComplaintCategory{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "chief_complaint_category", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("chief_complaint_category", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *chiefComplaintCategoryRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.ChiefComplaintCategory{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "chief_complaint_category", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("chief_complaint_category", fmt.Sprintf("%d", id))
	}
	return nil
}
