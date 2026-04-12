// Package repository provides data access implementations for ChiefComplaintType entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ChiefComplaintType ----

type ChiefComplaintTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error)
	Create(ctx context.Context, category *model.ChiefComplaintType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type chiefComplaintTypeRepository struct{ db *gorm.DB }

func NewChiefComplaintTypeRepository(db *gorm.DB) ChiefComplaintTypeRepository {
	return &chiefComplaintTypeRepository{db: db}
}

func (r *chiefComplaintTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error) {
	categories := make([]model.ChiefComplaintType, 0)
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "chief_complaint_type", "")
	}
	return categories, nil
}

func (r *chiefComplaintTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error) {
	var category model.ChiefComplaintType
	err := r.db.WithContext(ctx).First(&category, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "chief_complaint_type", fmt.Sprintf("%d", id))
	}
	return &category, nil
}

func (r *chiefComplaintTypeRepository) Create(ctx context.Context, category *model.ChiefComplaintType) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "chief_complaint_type", "")
	}
	return nil
}

func (r *chiefComplaintTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ChiefComplaintType{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "chief_complaint_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("chief_complaint_type", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *chiefComplaintTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.ChiefComplaintType{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "chief_complaint_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("chief_complaint_type", fmt.Sprintf("%d", id))
	}
	return nil
}
