// Package cage owns cages master data access.
package medicalrecord

// Moved from internal/repository/cage (BE9-2D ⑥ Batch A roll-up・BE8-4 subpackage)。
// generic Repository/New は entity-specific 名へ改名（①⑤先例）— 外部は facade alias 経由で不変。

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Repository is the data access interface for cages.
type CageRepository interface {
	FindAll(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	CountUsageByCageID(ctx context.Context, clinicID, id uint64) (int64, error)
	Create(ctx context.Context, cage *model.Cage) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Cage, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type cageRepositoryImpl struct{ db *gorm.DB }

// New constructs a Repository.
func NewCageRepository(db *gorm.DB) CageRepository {
	return &cageRepositoryImpl{db: db}
}

func (r *cageRepositoryImpl) FindAll(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error) {
	cages := make([]model.Cage, 0)
	q := r.db.WithContext(ctx).Model(&model.Cage{}).Scopes(persistence.ClinicScope(clinicID))
	if cageType != nil {
		q = q.Where("cage_type = ?", *cageType)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&cages).Error; err != nil {
		return nil, apperrors.FromGORM(err, "cage", "")
	}
	return cages, nil
}

func (r *cageRepositoryImpl) FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	return persistence.FindByIDScoped[model.Cage](ctx, r.db, "cage", clinicID, id)
}

func (r *cageRepositoryImpl) Create(ctx context.Context, cage *model.Cage) error {
	db := r.db.WithContext(ctx)
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	wantActive := cage.IsActive
	if err := db.Create(cage).Error; err != nil {
		return apperrors.FromGORM(err, "cage", "")
	}
	if !wantActive {
		if err := db.Model(cage).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "cage", fmt.Sprintf("%d", cage.ID))
		}
		cage.IsActive = false
	}
	return nil
}

func (r *cageRepositoryImpl) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Cage, error) {
	if err := persistence.UpdateScopedByID(ctx, r.db, &model.Cage{}, "cage", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *cageRepositoryImpl) Delete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DeleteScopedByID(ctx, r.db, &model.Cage{}, "cage", clinicID, id)
}

func (r *cageRepositoryImpl) CountUsageByCageID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Hospitalization{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("cage_id = ? AND deleted_at IS NULL", id).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "hospitalization", "")
	}
	return count, nil
}

func (r *cageRepositoryImpl) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.Cage{}, "cage", clinicID, ids, "sort_order")
}
