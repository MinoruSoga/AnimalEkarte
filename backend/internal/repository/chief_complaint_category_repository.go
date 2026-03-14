// Package repository provides data access implementations for ChiefComplaintCategory entity.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ChiefComplaintCategory ----

type ChiefComplaintCategoryRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintCategory, error)
	FindByID(ctx context.Context, id uint64) (*model.ChiefComplaintCategory, error)
	Create(ctx context.Context, category *model.ChiefComplaintCategory) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
}

type chiefComplaintCategoryRepository struct{ db *gorm.DB }

func NewChiefComplaintCategoryRepository(db *gorm.DB) ChiefComplaintCategoryRepository {
	return &chiefComplaintCategoryRepository{db: db}
}

func (r *chiefComplaintCategoryRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintCategory, error) {
	categories := make([]model.ChiefComplaintCategory, 0)
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, apperrors.Wrap(err, "find chief complaint categories")
	}
	return categories, nil
}

func (r *chiefComplaintCategoryRepository) FindByID(ctx context.Context, id uint64) (*model.ChiefComplaintCategory, error) {
	var category model.ChiefComplaintCategory
	if err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("chief_complaint_category", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find chief complaint category by id")
	}
	return &category, nil
}

func (r *chiefComplaintCategoryRepository) Create(ctx context.Context, category *model.ChiefComplaintCategory) error {
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("chief_complaint_category", category.Name)
		}
		return apperrors.Wrap(err, "create chief complaint category")
	}
	return nil
}

func (r *chiefComplaintCategoryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ChiefComplaintCategory{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update chief complaint category")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("chief_complaint_category", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *chiefComplaintCategoryRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.ChiefComplaintCategory{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete chief complaint category")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("chief_complaint_category", fmt.Sprintf("%d", id))
	}
	return nil
}
