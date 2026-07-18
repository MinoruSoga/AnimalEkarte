// Package cage owns cages master data access.
package cage

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for cages.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	CountUsageByCageID(ctx context.Context, clinicID, id uint64) (int64, error)
	Create(ctx context.Context, cage *model.Cage) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Cage, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error) {
	cages := make([]model.Cage, 0)
	q := r.db.WithContext(ctx).Model(&model.Cage{}).Scopes(repohelpers.ClinicScope(clinicID))
	if cageType != nil {
		q = q.Where("cage_type = ?", *cageType)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&cages).Error; err != nil {
		return nil, apperrors.FromGORM(err, "cage", "")
	}
	return cages, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	return repohelpers.FindByIDScoped[model.Cage](ctx, r.db, "cage", clinicID, id)
}

func (r *repository) Create(ctx context.Context, cage *model.Cage) error {
	if err := r.db.WithContext(ctx).Create(cage).Error; err != nil {
		return apperrors.FromGORM(err, "cage", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Cage, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.Cage{}, "cage", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.Cage{}, "cage", clinicID, id)
}

func (r *repository) CountUsageByCageID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Hospitalization{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("cage_id = ? AND deleted_at IS NULL", id).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "hospitalization", "")
	}
	return count, nil
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.Cage{}, "cage", clinicID, ids, "sort_order")
}
