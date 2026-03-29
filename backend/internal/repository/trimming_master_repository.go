// Package repository provides data access implementations for TrimmingCourse and TrimmingOption entities.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- TrimmingCourse ----

type TrimmingCourseRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	FindByID(ctx context.Context, id uint64) (*model.TrimmingCourse, error)
	Create(ctx context.Context, course *model.TrimmingCourse) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourse, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type trimmingCourseRepository struct{ db *gorm.DB }

func NewTrimmingCourseRepository(db *gorm.DB) TrimmingCourseRepository {
	return &trimmingCourseRepository{db: db}
}

func (r *trimmingCourseRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	courses := make([]model.TrimmingCourse, 0)
	if err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).Order("sort_order ASC, name ASC").Find(&courses).Error; err != nil {
		return nil, apperrors.Wrap(err, "find trimming courses")
	}
	return courses, nil
}

func (r *trimmingCourseRepository) FindByID(ctx context.Context, id uint64) (*model.TrimmingCourse, error) {
	var course model.TrimmingCourse
	err := r.db.WithContext(ctx).First(&course, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course", fmt.Sprintf("%d", id))
	}
	return &course, nil
}

func (r *trimmingCourseRepository) Create(ctx context.Context, course *model.TrimmingCourse) error {
	if err := r.db.WithContext(ctx).Create(course).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("trimming_course", course.Name)
		}
		return apperrors.Wrap(err, "create trimming course")
	}
	return nil
}

func (r *trimmingCourseRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourse, error) {
	result := r.db.WithContext(ctx).
		Model(&model.TrimmingCourse{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.Wrap(result.Error, "update trimming course")
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("trimming_course", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, id)
}

func (r *trimmingCourseRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.TrimmingCourse{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete trimming course")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("trimming_course", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *trimmingCourseRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.TrimmingCourse{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, "reorder trimming course")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("trimming_course id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("reorder trimming course: %w", err)
	}
	return nil
}

// ---- TrimmingOption ----

type TrimmingOptionRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	FindByID(ctx context.Context, id uint64) (*model.TrimmingOption, error)
	Create(ctx context.Context, option *model.TrimmingOption) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingOption, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type trimmingOptionRepository struct{ db *gorm.DB }

func NewTrimmingOptionRepository(db *gorm.DB) TrimmingOptionRepository {
	return &trimmingOptionRepository{db: db}
}

func (r *trimmingOptionRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	options := make([]model.TrimmingOption, 0)
	if err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).Order("sort_order ASC, name ASC").Find(&options).Error; err != nil {
		return nil, apperrors.Wrap(err, "find trimming options")
	}
	return options, nil
}

func (r *trimmingOptionRepository) FindByID(ctx context.Context, id uint64) (*model.TrimmingOption, error) {
	var option model.TrimmingOption
	err := r.db.WithContext(ctx).First(&option, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("%d", id))
	}
	return &option, nil
}

func (r *trimmingOptionRepository) Create(ctx context.Context, option *model.TrimmingOption) error {
	if err := r.db.WithContext(ctx).Create(option).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("trimming_option", option.Name)
		}
		return apperrors.Wrap(err, "create trimming option")
	}
	return nil
}

func (r *trimmingOptionRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingOption, error) {
	result := r.db.WithContext(ctx).
		Model(&model.TrimmingOption{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.Wrap(result.Error, "update trimming option")
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("trimming_option", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, id)
}

func (r *trimmingOptionRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.TrimmingOption{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete trimming option")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("trimming_option", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *trimmingOptionRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.TrimmingOption{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, "reorder trimming option")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("trimming_option id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("reorder trimming options: %w", err)
	}
	return nil
}
