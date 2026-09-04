// Package occupation owns occupations master data access.
package staff

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Repository is the data access interface for occupations.
type OccupationRepository interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
	FindAll(ctx context.Context, clinicID uint64) ([]model.Occupation, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error)
	LockActiveByIDForShare(ctx context.Context, clinicID, id uint64) (*model.Occupation, error)
	LockActiveByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Occupation, error)
	Create(ctx context.Context, occupation *model.Occupation) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByOccupationID(ctx context.Context, clinicID, occupationID uint64) (int64, error)
}

type occupationRepository struct{ db *gorm.DB }

// New constructs a Repository.
func NewOccupationRepository(db *gorm.DB) OccupationRepository {
	return &occupationRepository{db: db}
}

func (r *occupationRepository) WithTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx))
	}); err != nil {
		return apperrors.Wrap(err, "occupation transaction failed")
	}
	return nil
}

func (r *occupationRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Occupation, error) {
	occupations := make([]model.Occupation, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&occupations).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "occupation", "")
	}
	return occupations, nil
}

func (r *occupationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error) {
	return persistence.FindByIDScoped[model.Occupation](
		ctx,
		persistence.DBOrTx(ctx, r.db),
		"occupation",
		clinicID,
		id,
	)
}

func (r *occupationRepository) LockActiveByIDForShare(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Occupation, error) {
	return r.lockActiveByID(ctx, clinicID, id, "SHARE")
}

func (r *occupationRepository) LockActiveByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Occupation, error) {
	return r.lockActiveByID(ctx, clinicID, id, "UPDATE")
}

func (r *occupationRepository) lockActiveByID(
	ctx context.Context,
	clinicID, id uint64,
	strength string,
) (*model.Occupation, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError(
			"occupation lock requires an active transaction",
		)
	}
	var occupation model.Occupation
	if err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: strength}).
		Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", clinicID, id).
		First(&occupation).Error; err != nil {
		return nil, apperrors.FromGORM(err, "occupation", fmt.Sprintf("%d", id))
	}
	return &occupation, nil
}

func (r *occupationRepository) Create(ctx context.Context, occupation *model.Occupation) error {
	db := persistence.DBOrTx(ctx, r.db)
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	wantActive := occupation.IsActive
	if err := db.Create(occupation).Error; err != nil {
		return apperrors.FromGORM(err, "occupation", "")
	}
	if !wantActive {
		if err := db.Model(occupation).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "occupation", fmt.Sprintf("%d", occupation.ID))
		}
		occupation.IsActive = false
	}
	return nil
}

func (r *occupationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error) {
	if err := persistence.UpdateScopedByID(
		ctx,
		persistence.DBOrTx(ctx, r.db),
		&model.Occupation{},
		"occupation",
		clinicID,
		id,
		fields,
	); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *occupationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM staffs
			JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id
			  AND staff_clinic_assignments.clinic_id = ?
			  AND staff_clinic_assignments.deleted_at IS NULL
			WHERE staffs.occupation_id = occupations.id
			  AND staffs.deleted_at IS NULL
		)`, clinicID).
		Delete(&model.Occupation{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "occupation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeOccupationDeleteMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *occupationRepository) normalizeOccupationDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	return apperrors.WrapConflict("この役職はスタッフ情報で使用中のため削除できません")
}

// CountUsageByOccupationID returns staff references for the occupation (BUG-112).
// staffs lack clinic_id; tenant isolation is via staff_clinic_assignments JOIN.
func (r *occupationRepository) CountUsageByOccupationID(ctx context.Context, clinicID, occupationID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Staff{}).
		Joins("JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL", clinicID).
		Where("staffs.occupation_id = ? AND staffs.deleted_at IS NULL", occupationID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "staff", "")
	}
	return count, nil
}

func (r *occupationRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.Occupation{}, "occupation", clinicID, ids, "sort_order")
}
