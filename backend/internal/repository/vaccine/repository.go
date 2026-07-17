// Package vaccine owns vaccines master data access.
package vaccine

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for vaccines.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error)
	Create(ctx context.Context, vaccine *model.Vaccine) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccine, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByVaccineID(ctx context.Context, clinicID, vaccineID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error) {
	vaccines := make([]model.Vaccine, 0)
	q := r.db.WithContext(ctx).Model(&model.Vaccine{}).Scopes(repohelpers.ClinicScope(clinicID))
	if species != nil {
		q = q.Where("species = ?", *species)
	}
	if err := q.Order("sort_order ASC, name ASC").Limit(repohelpers.MaxMasterListRows).Find(&vaccines).Error; err != nil {
		return nil, apperrors.FromGORM(err, "vaccine", "")
	}
	return vaccines, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error) {
	return repohelpers.FindByIDScoped[model.Vaccine](ctx, r.db, "vaccine", clinicID, id)
}

func (r *repository) Create(ctx context.Context, vaccine *model.Vaccine) error {
	if err := r.db.WithContext(ctx).Create(vaccine).Error; err != nil {
		return apperrors.FromGORM(err, "vaccine", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccine, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.Vaccine{}, "vaccine", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.Vaccine{}, "vaccine", clinicID, id)
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.Vaccine{}, "vaccine", clinicID, ids, "sort_order")
}

// CountUsageByVaccineID returns vaccination references (BUG-107).
func (r *repository) CountUsageByVaccineID(ctx context.Context, clinicID, vaccineID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Vaccination{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("vaccine_id = ? AND deleted_at IS NULL", vaccineID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "vaccination_record", "")
	}
	return count, nil
}

// CountChildrenByParentID returns child vaccine count (BUG-390).
func (r *repository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Vaccine{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "vaccine", fmt.Sprintf("%d", parentID))
	}
	return count, nil
}
