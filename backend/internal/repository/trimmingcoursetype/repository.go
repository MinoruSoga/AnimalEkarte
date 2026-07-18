// Package trimmingcoursetype owns trimming_course_types master data access.
package trimmingcoursetype

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for trimming course types.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error)
	Create(ctx context.Context, m *model.TrimmingCourseType) (*model.TrimmingCourseType, error)
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourseType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	CountUsageByCourseTypeID(ctx context.Context, clinicID, id uint64) (int64, error)
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error) {
	var ms []model.TrimmingCourseType
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&ms).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course_type", "")
	}
	return ms, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error) {
	var m model.TrimmingCourseType
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		First(&m, id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course_type", fmt.Sprintf("%d", id))
	}
	return &m, nil
}

func (r *repository) Create(ctx context.Context, m *model.TrimmingCourseType) (*model.TrimmingCourseType, error) {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course_type", "")
	}
	return m, nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourseType, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.TrimmingCourseType{}, "trimming_course_type", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.TrimmingCourseType{}, "trimming_course_type", clinicID, id)
}

func (r *repository) CountUsageByCourseTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.TrimmingCourse{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("course_type_id = ? AND deleted_at IS NULL", id).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "trimming_course_type", fmt.Sprintf("%d", id))
	}
	return count, nil
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.TrimmingCourseType{}, "trimming_course_type", clinicID, ids, "sort_order")
}
