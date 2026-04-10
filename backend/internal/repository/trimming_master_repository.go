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
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
	Create(ctx context.Context, course *model.TrimmingCourse) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourse, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountRecordsByCourseID(ctx context.Context, courseID uint64) (int64, error)
}

type trimmingCourseRepository struct{ db *gorm.DB }

func NewTrimmingCourseRepository(db *gorm.DB) TrimmingCourseRepository {
	return &trimmingCourseRepository{db: db}
}

func (r *trimmingCourseRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	courses := make([]model.TrimmingCourse, 0)
	if err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).Order("sort_order ASC, name ASC").Find(&courses).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course", "")
	}
	return courses, nil
}

func (r *trimmingCourseRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	var course model.TrimmingCourse
	err := r.db.WithContext(ctx).First(&course, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course", fmt.Sprintf("%d", id))
	}
	return &course, nil
}

func (r *trimmingCourseRepository) Create(ctx context.Context, course *model.TrimmingCourse) error {
	if err := r.db.WithContext(ctx).Create(course).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "trimming_course", "")
	}
	return nil
}

func (r *trimmingCourseRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourse, error) {
	result := r.db.WithContext(ctx).
		Model(&model.TrimmingCourse{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "trimming_course", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("trimming_course", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *trimmingCourseRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.TrimmingCourse{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "trimming_course", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("trimming_course", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountRecordsByCourseID は指定コースを使用しているトリミング記録数を返す（BUG-111）
func (r *trimmingCourseRepository) CountRecordsByCourseID(ctx context.Context, courseID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.TrimmingRecord{}).
		Where("course_id = ?", courseID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "trimming_record", "")
	}
	return count, nil
}

func (r *trimmingCourseRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(r.db, ctx, &model.TrimmingCourse{}, "trimming_course", clinicID, ids)
}

// ---- TrimmingOption ----

type TrimmingOptionRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
	Create(ctx context.Context, option *model.TrimmingOption) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingOption, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountRecordsByOptionID(ctx context.Context, optionID uint64) (int64, error)
}

type trimmingOptionRepository struct{ db *gorm.DB }

func NewTrimmingOptionRepository(db *gorm.DB) TrimmingOptionRepository {
	return &trimmingOptionRepository{db: db}
}

func (r *trimmingOptionRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	options := make([]model.TrimmingOption, 0)
	if err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).Order("sort_order ASC, name ASC").Find(&options).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_option", "")
	}
	return options, nil
}

func (r *trimmingOptionRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	var option model.TrimmingOption
	err := r.db.WithContext(ctx).First(&option, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("%d", id))
	}
	return &option, nil
}

func (r *trimmingOptionRepository) Create(ctx context.Context, option *model.TrimmingOption) error {
	if err := r.db.WithContext(ctx).Create(option).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "trimming_option", "")
	}
	return nil
}

func (r *trimmingOptionRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingOption, error) {
	result := r.db.WithContext(ctx).
		Model(&model.TrimmingOption{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "trimming_option", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("trimming_option", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *trimmingOptionRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.TrimmingOption{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "trimming_option", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("trimming_option", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *trimmingOptionRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(r.db, ctx, &model.TrimmingOption{}, "trimming_option", clinicID, ids)
}

// CountRecordsByOptionID は指定オプションを使用しているトリミング記録数を返す（BUG-201）
func (r *trimmingOptionRepository) CountRecordsByOptionID(ctx context.Context, optionID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.TrimmingRecordOption{}).
		Where("option_id = ?", optionID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "trimming_record_option", "")
	}
	return count, nil
}
