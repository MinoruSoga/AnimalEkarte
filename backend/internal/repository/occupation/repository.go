// Package occupation owns occupations master data access.
package occupation

import (
	"context"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for occupations.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Occupation, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error)
	Create(ctx context.Context, occupation *model.Occupation) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByOccupationID(ctx context.Context, clinicID, occupationID uint64) (int64, error)
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64) ([]model.Occupation, error) {
	occupations := make([]model.Occupation, 0)
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&occupations).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "occupation", "")
	}
	return occupations, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error) {
	return repohelpers.FindByIDScoped[model.Occupation](ctx, r.db, "occupation", clinicID, id)
}

func (r *repository) Create(ctx context.Context, occupation *model.Occupation) error {
	err := r.db.WithContext(ctx).Create(occupation).Error
	if err != nil {
		return apperrors.FromGORM(err, "occupation", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.Occupation{}, "occupation", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.Occupation{}, "occupation", clinicID, id)
}

// CountUsageByOccupationID returns staff references for the occupation (BUG-112).
// staffs lack clinic_id; tenant isolation is via staff_clinic_assignments JOIN.
func (r *repository) CountUsageByOccupationID(ctx context.Context, clinicID, occupationID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Staff{}).
		Joins("JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL", clinicID).
		Where("staffs.occupation_id = ? AND staffs.deleted_at IS NULL", occupationID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "staff", "")
	}
	return count, nil
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.Occupation{}, "occupation", clinicID, ids, "sort_order")
}
