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

// LockForImageUpdate returns the authoritative before-value under an UPDATE
// lock, preserving the course -> ordered image-key lock acquisition order.
func (r *trimmingCourseRepository) LockForImageUpdate(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("course image update requires an ambient transaction")
	}
	var course model.TrimmingCourse
	if err := persistence.DBOrTx(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"}).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&course).Error; err != nil {
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

// Delete はコースの使用中判定ができないため、常に拒否する（EMR-549 C1 F3）。
// 旧判定の appointment_trimming_details.course_id は 001_init.sql L10892 で DROP 済みで、置換先
// （trimming_booking_revisions.course_id）の使用数 reader は EMR-480 の担当で未接続。論理削除は
// FK RESTRICT に守られないため、使用中のコースも消えてしまう。使用数を確認できない削除は行わない。
func (r *trimmingCourseRepository) Delete(_ context.Context, _, _ uint64) error {
	return apperrors.WrapTrimmingMasterUsageUnverified()
}

// CountUsageByTrimmingCourseID は使用数を確認できないため、常に拒否エラーを返す（EMR-549 C1 F3）。
// 旧集計の appointment_trimming_details.course_id は DROP 済みで、置換先の reader が未接続。
func (r *trimmingCourseRepository) CountUsageByTrimmingCourseID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, apperrors.WrapTrimmingMasterUsageUnverified()
}

func (r *trimmingCourseRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.TrimmingCourse{}, "trimming_course", clinicID, ids, "sort_order")
}
