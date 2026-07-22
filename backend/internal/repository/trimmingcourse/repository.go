// Package trimmingcourse owns trimming course master data access.
package trimmingcourse

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for trimming courses.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
	Create(ctx context.Context, course *model.TrimmingCourse) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourse, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByTrimmingCourseID(ctx context.Context, clinicID, courseID uint64) (int64, error)
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	courses := make([]model.TrimmingCourse, 0)
	if err := r.db.WithContext(ctx).Scopes(repohelpers.ClinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&courses).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course", "")
	}
	return courses, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	var course model.TrimmingCourse
	db := repohelpers.DBOrTx(ctx, r.db)
	if repohelpers.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	if err := db.Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).First(&course).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course", fmt.Sprintf("%d", id))
	}
	return &course, nil
}

func (r *repository) Create(ctx context.Context, course *model.TrimmingCourse) error {
	if err := r.db.WithContext(ctx).Create(course).Error; err != nil {
		return apperrors.FromGORM(err, "trimming_course", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourse, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.TrimmingCourse{}, "trimming_course", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.TrimmingCourse{}, "trimming_course", clinicID, id)
}

// CountUsageByTrimmingCourseID は指定コースを使用しているトリミング詳細数を返す（BUG-111）
// appointment_trimming_details は deleted_at を持たないため appointments を JOIN して論理削除を考慮する
func (r *repository) CountUsageByTrimmingCourseID(ctx context.Context, clinicID, courseID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.AppointmentTrimmingDetail{}).
		Joins("JOIN appointments ON appointments.id = appointment_trimming_details.appointment_id AND appointments.clinic_id = ? AND appointments.deleted_at IS NULL", clinicID).
		Where("appointment_trimming_details.course_id = ?", courseID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "appointment_trimming_detail", "")
	}
	return count, nil
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.TrimmingCourse{}, "trimming_course", clinicID, ids, "sort_order")
}
