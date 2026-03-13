// Package repository provides data access implementations for TrimmingCourse and TrimmingOption entities.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- TrimmingCourse ----

type TrimmingCourseRepository interface {
	FindAll(ctx context.Context) ([]model.TrimmingCourse, error)
	FindByID(ctx context.Context, id uint64) (*model.TrimmingCourse, error)
	Create(ctx context.Context, course *model.TrimmingCourse) error
	Update(ctx context.Context, course *model.TrimmingCourse) error
	Delete(ctx context.Context, id uint64) error
}

type trimmingCourseRepository struct{ db *gorm.DB }

func NewTrimmingCourseRepository(db *gorm.DB) TrimmingCourseRepository {
	return &trimmingCourseRepository{db: db}
}

func (r *trimmingCourseRepository) FindAll(ctx context.Context) ([]model.TrimmingCourse, error) {
	var courses []model.TrimmingCourse
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&courses).Error; err != nil {
		return nil, apperrors.Wrap(err, "find trimming courses")
	}
	return courses, nil
}

func (r *trimmingCourseRepository) FindByID(ctx context.Context, id uint64) (*model.TrimmingCourse, error) {
	var course model.TrimmingCourse
	if err := r.db.WithContext(ctx).First(&course, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("trimming_course", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find trimming course by id")
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

func (r *trimmingCourseRepository) Update(ctx context.Context, course *model.TrimmingCourse) error {
	result := r.db.WithContext(ctx).
		Model(&model.TrimmingCourse{}).
		Where("id = ? AND clinic_id = ?", course.ID, course.ClinicID).
		Updates(course)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update trimming course")
	}
	if result.RowsAffected == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "update trimming course")
	}
	return nil
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

// ---- TrimmingOption ----

type TrimmingOptionRepository interface {
	FindAll(ctx context.Context) ([]model.TrimmingOption, error)
	FindByID(ctx context.Context, id uint64) (*model.TrimmingOption, error)
	Create(ctx context.Context, option *model.TrimmingOption) error
	Update(ctx context.Context, option *model.TrimmingOption) error
	Delete(ctx context.Context, id uint64) error
}

type trimmingOptionRepository struct{ db *gorm.DB }

func NewTrimmingOptionRepository(db *gorm.DB) TrimmingOptionRepository {
	return &trimmingOptionRepository{db: db}
}

func (r *trimmingOptionRepository) FindAll(ctx context.Context) ([]model.TrimmingOption, error) {
	var options []model.TrimmingOption
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&options).Error; err != nil {
		return nil, apperrors.Wrap(err, "find trimming options")
	}
	return options, nil
}

func (r *trimmingOptionRepository) FindByID(ctx context.Context, id uint64) (*model.TrimmingOption, error) {
	var option model.TrimmingOption
	if err := r.db.WithContext(ctx).First(&option, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("trimming_option", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find trimming option by id")
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

func (r *trimmingOptionRepository) Update(ctx context.Context, option *model.TrimmingOption) error {
	result := r.db.WithContext(ctx).
		Model(&model.TrimmingOption{}).
		Where("id = ? AND clinic_id = ?", option.ID, option.ClinicID).
		Updates(option)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update trimming option")
	}
	if result.RowsAffected == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "update trimming option")
	}
	return nil
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
