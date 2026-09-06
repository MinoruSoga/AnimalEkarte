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

// VaccineRepository is the data access interface for vaccines (master).
// Moved from internal/repository/vaccine (BE8-4) — BE9-2D roll-up. Renamed from that
// subpackage's generic "Repository" to this entity-specific name only because medicalrecord
// holds multiple repository interfaces in one package; every external caller only ever saw
// this name via the internal/repository facade, so no call site changes.
type VaccineRepository interface {
	FindAll(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error)
	Create(ctx context.Context, vaccine *model.Vaccine) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateVaccineInput) (*model.Vaccine, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByVaccineID(ctx context.Context, clinicID, vaccineID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type vaccineRepository struct{ db *gorm.DB }

// NewVaccineRepository constructs a VaccineRepository.
func NewVaccineRepository(db *gorm.DB) VaccineRepository {
	return &vaccineRepository{db: db}
}

func (r *vaccineRepository) FindAll(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error) {
	vaccines := make([]model.Vaccine, 0)
	q := r.db.WithContext(ctx).Model(&model.Vaccine{}).Scopes(persistence.ClinicScope(clinicID))
	if species != nil {
		q = q.Where("species = ?", *species)
	}
	if err := q.Order("sort_order ASC, name ASC").Limit(persistence.MaxMasterListRows).Find(&vaccines).Error; err != nil {
		return nil, apperrors.FromGORM(err, "vaccine", "")
	}
	return vaccines, nil
}

func (r *vaccineRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error) {
	var vaccine model.Vaccine
	db := persistence.DBOrTx(ctx, r.db)
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	if err := db.Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&vaccine).Error; err != nil {
		return nil, apperrors.FromGORM(err, "vaccine", fmt.Sprintf("%d", id))
	}
	return &vaccine, nil
}

func (r *vaccineRepository) Create(ctx context.Context, vaccine *model.Vaccine) error {
	db := r.db.WithContext(ctx)
	wantActive := vaccine.IsActive
	if err := db.Create(vaccine).Error; err != nil {
		return apperrors.FromGORM(err, "vaccine", "")
	}
	if !wantActive {
		if err := db.Model(vaccine).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "vaccine", fmt.Sprintf("%d", vaccine.ID))
		}
		vaccine.IsActive = false
	}
	return nil
}

func (r *vaccineRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateVaccineInput) (*model.Vaccine, error) {
	if err := r.update(ctx, clinicID, id, buildVaccineUpdate(&cmd)); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *vaccineRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, r.db, &model.Vaccine{}, "vaccine", clinicID, id, fields)
}

func (r *vaccineRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM vaccines children
			WHERE children.parent_id = vaccines.id
			  AND children.clinic_id = ?
			  AND children.deleted_at IS NULL
		)`, clinicID).
		Where(`NOT EXISTS (
			SELECT 1 FROM vaccinations
			WHERE vaccinations.vaccine_id = vaccines.id
			  AND vaccinations.clinic_id = ?
			  AND vaccinations.deleted_at IS NULL
		)`, clinicID).
		Delete(&model.Vaccine{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "vaccine", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeVaccineDeleteMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *vaccineRepository) normalizeVaccineDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	childCount, err := r.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return apperrors.WrapConflict("このワクチンは子ワクチンが存在するため削除できません")
	}
	return apperrors.WrapConflict("このワクチンはワクチン接種記録で使用中のため削除できません")
}

func (r *vaccineRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.Vaccine{}, "vaccine", clinicID, ids, "sort_order")
}

// CountUsageByVaccineID returns vaccination references (BUG-107).
func (r *vaccineRepository) CountUsageByVaccineID(ctx context.Context, clinicID, vaccineID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Vaccination{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("vaccine_id = ? AND deleted_at IS NULL", vaccineID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "vaccination_record", "")
	}
	return count, nil
}

// CountChildrenByParentID returns child vaccine count (BUG-390).
func (r *vaccineRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Vaccine{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "vaccine", fmt.Sprintf("%d", parentID))
	}
	return count, nil
}
