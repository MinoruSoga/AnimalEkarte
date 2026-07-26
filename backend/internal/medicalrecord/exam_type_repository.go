package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ExamTypeRepository is the data access interface for examination types. Moved from
// internal/repository/examtype — BE9-2C roll-up (see diagnosis_type_repository.go's header
// for the rename rationale).
type ExamTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error)
	Create(ctx context.Context, exType *model.ExaminationType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByExamTypeID(ctx context.Context, clinicID, examTypeID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type examTypeRepository struct{ db *gorm.DB }

// NewExamTypeRepository constructs an ExamTypeRepository.
func NewExamTypeRepository(db *gorm.DB) ExamTypeRepository {
	return &examTypeRepository{db: db}
}

func (r *examTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error) {
	exTypes := make([]model.ExaminationType, 0)
	err := r.db.WithContext(ctx).Scopes(persistence.ClinicScope(clinicID)).Preload("Items", "clinic_id = ?", clinicID).Order("sort_order ASC, name ASC").Find(&exTypes).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "examination_type", "")
	}
	return exTypes, nil
}

func (r *examTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
	var exType model.ExaminationType
	db := persistence.DBOrTx(ctx, r.db)
	lockForValidation := persistence.TxFromContext(ctx) != nil
	if lockForValidation {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.Preload("Items", func(itemsDB *gorm.DB) *gorm.DB {
		itemsDB = itemsDB.Where("clinic_id = ?", clinicID)
		if lockForValidation {
			itemsDB = itemsDB.Clauses(clause.Locking{Strength: "SHARE"})
		}
		return itemsDB
	}).Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&exType).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "examination_type", fmt.Sprintf("%d", id))
	}
	return &exType, nil
}

func (r *examTypeRepository) Create(ctx context.Context, exType *model.ExaminationType) error {
	err := r.db.WithContext(ctx).Create(exType).Error
	if err != nil {
		return apperrors.FromGORM(err, "examination_type", "")
	}
	return nil
}

func (r *examTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error) {
	if err := persistence.UpdateScopedByID(ctx, r.db, &model.ExaminationType{}, "examination_type", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *examTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DeleteScopedByID(ctx, r.db, &model.ExaminationType{}, "examination_type", clinicID, id)
}

func (r *examTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.ExaminationType{}, "examination_type", clinicID, ids, "sort_order")
}

// CountUsageByExamTypeID returns exam references (BUG-107).
func (r *examTypeRepository) CountUsageByExamTypeID(ctx context.Context, clinicID, examTypeID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Examination{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("exam_type_id = ? AND deleted_at IS NULL", examTypeID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "exam_type", "")
	}
	return count, nil
}

// CountChildrenByParentID returns child examination-type count before parent delete.
func (r *examTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.ExaminationType{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "examination_type", "")
	}
	return count, nil
}
