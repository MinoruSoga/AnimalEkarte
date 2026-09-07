package medicalrecord

// Moved from internal/repository/cage (BE9-2D ⑥ Batch A roll-up・BE8-4 subpackage)。
// generic Repository/New は entity-specific 名へ改名（①⑤先例）— 外部は facade alias 経由で不変。

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Repository is the data access interface for cages.
type CageRepository interface {
	FindAll(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	// LockByIDForUpdate takes FOR UPDATE under an ambient transaction (SEC-CS-F13).
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	CountUsageByCageID(ctx context.Context, clinicID, id uint64) (int64, error)
	Create(ctx context.Context, cage *model.Cage) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateCageInput) (*model.Cage, error)
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
	q := persistence.DBOrTx(ctx, r.db).Model(&model.Cage{}).Scopes(persistence.ClinicScope(clinicID))
	if cageType != nil {
		q = q.Where("cage_type = ?", *cageType)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&cages).Error; err != nil {
		return nil, apperrors.FromGORM(err, "cage", "")
	}
	return cages, nil
}

// FindByID loads a clinic-scoped cage. When called under an ambient transaction it
// takes FOR SHARE so concurrent soft-delete (FOR UPDATE) waits until the caller
// commits — hospitalization cage FK validation serialization (SEC-CS-F13).
func (r *cageRepositoryImpl) FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	var cage model.Cage
	db := persistence.DBOrTx(ctx, r.db)
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	if err := db.Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&cage).Error; err != nil {
		return nil, apperrors.FromGORM(err, "cage", fmt.Sprintf("%d", id))
	}
	return &cage, nil
}

// LockByIDForUpdate exclusive-locks a cage row for the ambient transaction so usage
// check + soft-delete serialize with hospitalization assignment (SEC-CS-F13).
func (r *cageRepositoryImpl) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("cage lock requires ambient transaction")
	}
	var cage model.Cage
	db := persistence.DBOrTx(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"})
	if err := db.Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&cage).Error; err != nil {
		return nil, apperrors.FromGORM(err, "cage", fmt.Sprintf("%d", id))
	}
	return &cage, nil
}

func (r *cageRepositoryImpl) Create(ctx context.Context, cage *model.Cage) error {
	db := persistence.DBOrTx(ctx, r.db)
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

func (r *cageRepositoryImpl) Update(ctx context.Context, clinicID, id uint64, cmd UpdateCageInput) (*model.Cage, error) {
	if err := r.update(ctx, clinicID, id, buildCageUpdate(&cmd)); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *cageRepositoryImpl) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.Cage{}, "cage", clinicID, id, fields)
}

func (r *cageRepositoryImpl) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM hospitalizations
			WHERE hospitalizations.cage_id = cages.id
			  AND hospitalizations.clinic_id = ?
			  AND hospitalizations.deleted_at IS NULL
		)`, clinicID).
		Delete(&model.Cage{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "cage", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeCageDeleteMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *cageRepositoryImpl) normalizeCageDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	return apperrors.WrapConflict("このケージは入院データで使用中のため削除できません")
}

func (r *cageRepositoryImpl) CountUsageByCageID(ctx context.Context, clinicID, id uint64) (int64, error) {
	// DBOrTx so delete's ambient transaction sees uncommitted hospitalization assignments.
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
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
