// Package reservationtypegroup owns reservation_type_groups master data access
// (BE8-4 Wave D — "reservation_type_group (optional co-master)", first of the
// reservation-type-children cut order in repository/CLAUDE.md).
package reservationtypegroup

import (
	"context"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for reservation type groups.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error)
	CountUsageByReservationTypeGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error)
	Create(ctx context.Context, g *model.ReservationTypeGroup) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationTypeGroup, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error) {
	var list []model.ReservationTypeGroup
	if err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&list).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_group", "")
	}
	return list, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error) {
	return repohelpers.FindByIDScoped[model.ReservationTypeGroup](ctx, r.db, "reservation_type_group", clinicID, id)
}

func (r *repository) CountUsageByReservationTypeGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ReservationType{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("group_id = ? AND deleted_at IS NULL", groupID).Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "reservation_type", "")
	}
	return count, nil
}

func (r *repository) Create(ctx context.Context, g *model.ReservationTypeGroup) error {
	if err := r.db.WithContext(ctx).Create(g).Error; err != nil {
		return apperrors.FromGORM(err, "reservation_type_group", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationTypeGroup, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.ReservationTypeGroup{}, "reservation_type_group", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.ReservationTypeGroup{}, "reservation_type_group", clinicID, id)
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.ReservationTypeGroup{}, "reservation_type_group", clinicID, ids, "sort_order")
}
