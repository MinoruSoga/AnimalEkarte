// Trimming course type persistence belongs to package trimming.
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

// TrimmingCourseTypeRepository is the trimming course type persistence contract.
type TrimmingCourseTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error)
	Create(ctx context.Context, m *model.TrimmingCourseType) (*model.TrimmingCourseType, error)
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	CountUsageByCourseTypeID(ctx context.Context, clinicID, id uint64) (int64, error)
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type trimmingCourseTypeRepository struct{ db *gorm.DB }

// NewTrimmingCourseTypeRepository constructs a trimming course type repository.
func NewTrimmingCourseTypeRepository(db *gorm.DB) TrimmingCourseTypeRepository {
	return &trimmingCourseTypeRepository{db: db}
}

func (r *trimmingCourseTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error) {
	var ms []model.TrimmingCourseType
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&ms).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course_type", "")
	}
	return ms, nil
}

func (r *trimmingCourseTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error) {
	var m model.TrimmingCourseType
	db := persistence.DBOrTx(ctx, r.db)
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Scopes(persistence.ClinicScope(clinicID)).
		First(&m, id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course_type", fmt.Sprintf("%d", id))
	}
	return &m, nil
}

func (r *trimmingCourseTypeRepository) Create(ctx context.Context, m *model.TrimmingCourseType) (*model.TrimmingCourseType, error) {
	if err := persistence.DBOrTx(ctx, r.db).Create(m).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_course_type", "")
	}
	return m, nil
}

func (r *trimmingCourseTypeRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error) {
	if err := r.update(ctx, clinicID, id, buildTrimmingCourseTypeUpdate(&cmd)); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *trimmingCourseTypeRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.TrimmingCourseType{}, "trimming_course_type", clinicID, id, fields)
}

func (r *trimmingCourseTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			First(&model.TrimmingCourseType{}).Error; err != nil {
			return apperrors.FromGORM(err, "trimming_course_type", fmt.Sprintf("%d", id))
		}
		result := tx.
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			Where(`NOT EXISTS (
				SELECT 1 FROM trimming_courses
				WHERE trimming_courses.course_type_id = trimming_course_types.id
				  AND trimming_courses.clinic_id = ?
				  AND trimming_courses.deleted_at IS NULL
			)`, clinicID).
			Delete(&model.TrimmingCourseType{})
		if result.Error != nil {
			return apperrors.FromGORM(result.Error, "trimming_course_type", fmt.Sprintf("%d", id))
		}
		if result.RowsAffected == 0 {
			return r.normalizeTrimmingCourseTypeDeleteMiss(persistence.WithTxValue(ctx, tx), clinicID, id)
		}
		return nil
	})
}

func (r *trimmingCourseTypeRepository) normalizeTrimmingCourseTypeDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	return apperrors.WrapConflict("この種別は使用中のため削除できません")
}

func (r *trimmingCourseTypeRepository) CountUsageByCourseTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.TrimmingCourse{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("course_type_id = ? AND deleted_at IS NULL", id).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "trimming_course_type", fmt.Sprintf("%d", id))
	}
	return count, nil
}

func (r *trimmingCourseTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.TrimmingCourseType{}, "trimming_course_type", clinicID, ids, "sort_order")
}
