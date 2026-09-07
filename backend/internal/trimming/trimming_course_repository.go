// Package trimming owns trimming HTTP, application, and persistence behavior.
package trimming

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// TrimmingCourseRepository is the trimming course persistence contract.
type TrimmingCourseRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
	Create(ctx context.Context, course *model.TrimmingCourse) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingCourseInput) (*model.TrimmingCourse, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByTrimmingCourseID(ctx context.Context, clinicID, courseID uint64) (int64, error)
}

type trimmingCourseRepository struct{ db *gorm.DB }

// NewTrimmingCourseRepository constructs a trimming course repository.
func NewTrimmingCourseRepository(db *gorm.DB) TrimmingCourseRepository {
	return &trimmingCourseRepository{db: db}
}

func (r *trimmingCourseRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	courses := make([]model.TrimmingCourse, 0)
	if err := persistence.DBOrTx(ctx, r.db).Scopes(persistence.ClinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&courses).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course", "")
	}
	return courses, nil
}

func (r *trimmingCourseRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	var course model.TrimmingCourse
	db := persistence.DBOrTx(ctx, r.db)
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	if err := db.Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&course).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course", fmt.Sprintf("%d", id))
	}
	return &course, nil
}

func (r *trimmingCourseRepository) Create(ctx context.Context, course *model.TrimmingCourse) error {
	db := persistence.DBOrTx(ctx, r.db)
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	wantActive := course.IsActive
	if err := db.Create(course).Error; err != nil {
		return apperrors.FromGORM(err, "trimming_course", "")
	}
	if !wantActive {
		if err := db.Model(course).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "trimming_course", fmt.Sprintf("%d", course.ID))
		}
		course.IsActive = false
	}
	return nil
}

func (r *trimmingCourseRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
	if err := r.update(ctx, clinicID, id, buildTrimmingCourseUpdate(&cmd)); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *trimmingCourseRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.TrimmingCourse{}, "trimming_course", clinicID, id, fields)
}

func (r *trimmingCourseRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			First(&model.TrimmingCourse{}).Error; err != nil {
			return apperrors.FromGORM(err, "trimming_course", fmt.Sprintf("%d", id))
		}
		result := tx.
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			Where(`NOT EXISTS (
				SELECT 1 FROM appointment_trimming_details
				JOIN appointments ON appointments.id = appointment_trimming_details.appointment_id
				  AND appointments.clinic_id = ?
				  AND appointments.deleted_at IS NULL
				WHERE appointment_trimming_details.clinic_id = ?
				  AND appointment_trimming_details.course_id = trimming_courses.id
			)`, clinicID, clinicID).
			Delete(&model.TrimmingCourse{})
		if result.Error != nil {
			return apperrors.FromGORM(result.Error, "trimming_course", fmt.Sprintf("%d", id))
		}
		if result.RowsAffected == 0 {
			return r.normalizeTrimmingCourseDeleteMiss(persistence.WithTxValue(ctx, tx), clinicID, id)
		}
		return nil
	})
}

func (r *trimmingCourseRepository) normalizeTrimmingCourseDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	return apperrors.WrapConflict("このトリミングコースはトリミング記録で使用中のため削除できません")
}

// CountUsageByTrimmingCourseID は指定コースを使用しているトリミング詳細数を返す（BUG-111）
// appointment_trimming_details は deleted_at を持たないため appointments を JOIN して論理削除を考慮する
func (r *trimmingCourseRepository) CountUsageByTrimmingCourseID(ctx context.Context, clinicID, courseID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.AppointmentTrimmingDetail{}).
		Joins("JOIN appointments ON appointments.id = appointment_trimming_details.appointment_id AND appointments.clinic_id = ? AND appointments.deleted_at IS NULL", clinicID).
		Where("appointment_trimming_details.clinic_id = ?", clinicID).
		Where("appointment_trimming_details.course_id = ?", courseID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "appointment_trimming_detail", "")
	}
	return count, nil
}

func (r *trimmingCourseRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.TrimmingCourse{}, "trimming_course", clinicID, ids, "sort_order")
}
