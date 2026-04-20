package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TrimmingCourseRepository はトリミングコースのデータアクセスインターフェース
type TrimmingCourseRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
	Create(ctx context.Context, course *model.TrimmingCourse) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourse, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountRecordsByCourseID(ctx context.Context, clinicID, courseID uint64) (int64, error)
}

type trimmingCourseRepository struct{ db *gorm.DB }

// NewTrimmingCourseRepository は TrimmingCourseRepository を生成する
func NewTrimmingCourseRepository(db *gorm.DB) TrimmingCourseRepository {
	return &trimmingCourseRepository{db: db}
}

func (r *trimmingCourseRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	courses := make([]model.TrimmingCourse, 0)
	if err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&courses).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course", "")
	}
	return courses, nil
}

func (r *trimmingCourseRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	var course model.TrimmingCourse
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&course).Error
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
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
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
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.TrimmingCourse{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "trimming_course", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("trimming_course", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountRecordsByCourseID は指定コースを使用しているトリミング詳細数を返す（BUG-111）
func (r *trimmingCourseRepository) CountRecordsByCourseID(ctx context.Context, clinicID, courseID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.AppointmentTrimmingDetail{}).
		Scopes(clinicScope(clinicID)).
		Where("course_id = ?", courseID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "appointment_trimming_detail", "")
	}
	return count, nil
}

func (r *trimmingCourseRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.TrimmingCourse{}, "trimming_course", clinicID, ids)
}
