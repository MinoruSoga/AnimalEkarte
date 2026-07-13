package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TrimmingCourseTypeRepository はトリミングコース種別マスタのデータアクセスインターフェース (#73)
type TrimmingCourseTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error)
	Create(ctx context.Context, m *model.TrimmingCourseType) (*model.TrimmingCourseType, error)
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourseType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	CountUsageByCourseTypeID(ctx context.Context, clinicID, id uint64) (int64, error)
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type trimmingCourseTypeRepository struct{ db *gorm.DB }

// NewTrimmingCourseTypeRepository は TrimmingCourseTypeRepository を初期化して返す
func NewTrimmingCourseTypeRepository(db *gorm.DB) TrimmingCourseTypeRepository {
	return &trimmingCourseTypeRepository{db: db}
}

func (r *trimmingCourseTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error) {
	var ms []model.TrimmingCourseType
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&ms).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course_type", "")
	}
	return ms, nil
}

func (r *trimmingCourseTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error) {
	var m model.TrimmingCourseType
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		First(&m, id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course_type", fmt.Sprintf("%d", id))
	}
	return &m, nil
}

func (r *trimmingCourseTypeRepository) Create(ctx context.Context, m *model.TrimmingCourseType) (*model.TrimmingCourseType, error) {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course_type", "")
	}
	return m, nil
}

func (r *trimmingCourseTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourseType, error) {
	if err := updateScopedByID(ctx, r.db, &model.TrimmingCourseType{}, "trimming_course_type", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *trimmingCourseTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return deleteScopedByID(ctx, r.db, &model.TrimmingCourseType{}, "trimming_course_type", clinicID, id)
}

// CountUsageByCourseTypeID は指定種別を参照している trimming_courses の件数を返す。
// trimming_courses は直接 clinic_id を持つため JOIN は不要。
func (r *trimmingCourseTypeRepository) CountUsageByCourseTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.TrimmingCourse{}).
		Scopes(clinicScope(clinicID)).
		Where("course_type_id = ? AND deleted_at IS NULL", id).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "trimming_course_type", fmt.Sprintf("%d", id))
	}
	return count, nil
}

func (r *trimmingCourseTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.TrimmingCourseType{}, "trimming_course_type", clinicID, ids, "sort_order")
}
