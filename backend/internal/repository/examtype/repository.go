// Package examtype owns examination_types master data access.
package examtype

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for examination types.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error)
	Create(ctx context.Context, exType *model.ExaminationType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByExamTypeID(ctx context.Context, clinicID, examTypeID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error) {
	exTypes := make([]model.ExaminationType, 0)
	err := r.db.WithContext(ctx).Scopes(repohelpers.ClinicScope(clinicID)).Preload("Items").Order("sort_order ASC, name ASC").Find(&exTypes).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "examination_type", "")
	}
	return exTypes, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
	var exType model.ExaminationType
	err := r.db.WithContext(ctx).Preload("Items").Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).First(&exType).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "examination_type", fmt.Sprintf("%d", id))
	}
	return &exType, nil
}

func (r *repository) Create(ctx context.Context, exType *model.ExaminationType) error {
	err := r.db.WithContext(ctx).Create(exType).Error
	if err != nil {
		return apperrors.FromGORM(err, "examination_type", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.ExaminationType{}, "examination_type", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.ExaminationType{}, "examination_type", clinicID, id)
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.ExaminationType{}, "examination_type", clinicID, ids, "sort_order")
}

// CountUsageByExamTypeID returns exam references (BUG-107).
func (r *repository) CountUsageByExamTypeID(ctx context.Context, clinicID, examTypeID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Examination{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("exam_type_id = ? AND deleted_at IS NULL", examTypeID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "exam_type", "")
	}
	return count, nil
}

// CountChildrenByParentID returns child examination-type count before parent delete.
func (r *repository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.ExaminationType{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "examination_type", "")
	}
	return count, nil
}
