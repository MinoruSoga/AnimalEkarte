package billing

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// InsuranceRepository is the data access interface for insurance masters.
type InsuranceRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Insurance, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error)
	Create(ctx context.Context, insurance *model.Insurance) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Insurance, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByInsuranceID(ctx context.Context, clinicID, insuranceID uint64) (int64, error)
}

type insuranceRepository struct{ db *gorm.DB }

// New constructs a InsuranceRepository.
func NewInsuranceRepository(db *gorm.DB) InsuranceRepository {
	return &insuranceRepository{db: db}
}

func (r *insuranceRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Insurance, error) {
	insurances := make([]model.Insurance, 0)
	err := r.db.WithContext(ctx).Scopes(persistence.ClinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&insurances).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "insurance", "")
	}
	return insurances, nil
}

func (r *insuranceRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error) {
	return persistence.FindByIDScoped[model.Insurance](ctx, persistence.DBOrTx(ctx, r.db), "insurance", clinicID, id)
}

func (r *insuranceRepository) Create(ctx context.Context, insurance *model.Insurance) error {
	// Capture intent before Create: gorm default:true omits zero bools from
	// INSERT and may write the DB default back into the struct (BUG-455-S3).
	wantActive := insurance.IsActive
	if err := r.db.WithContext(ctx).Create(insurance).Error; err != nil {
		return apperrors.FromGORM(err, "insurance", "")
	}
	if !wantActive {
		if err := r.db.WithContext(ctx).Model(insurance).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "insurance", "")
		}
		insurance.IsActive = false
	}
	return nil
}

func (r *insuranceRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Insurance, error) {
	var loaded *model.Insurance
	err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := persistence.UpdateScopedByID(txCtx, tx, &model.Insurance{}, "insurance", clinicID, id, fields); err != nil {
			return err
		}
		reloaded, err := r.FindByID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "reload insurance after update")
		}
		loaded = reloaded
		return nil
	})
	if err != nil {
		return nil, err
	}
	return loaded, nil
}

func (r *insuranceRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	// clinic_id + id + pets.insurance_id 不在を同一 DELETE で要求し、
	// CountUsage→Delete 間の参照追加 TOCTOU を原子的に塞ぐ。
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM pets
			WHERE pets.insurance_id = insurances.id
			  AND pets.clinic_id = ?
			  AND pets.deleted_at IS NULL
		)`, clinicID).
		Delete(&model.Insurance{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "insurance", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeDeleteIfUnusedMiss(ctx, clinicID, id)
	}
	return nil
}

// normalizeDeleteIfUnusedMiss は原子 DELETE が 0 行だった理由を再読取で区別する。
func (r *insuranceRepository) normalizeDeleteIfUnusedMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	count, err := r.CountUsageByInsuranceID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return apperrors.WrapConflict("この保険はペット情報で使用中のため削除できません")
	}
	return apperrors.WrapConflict("この保険はペット情報で使用中のため削除できません")
}

func (r *insuranceRepository) CountUsageByInsuranceID(ctx context.Context, clinicID, insuranceID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Pet{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("insurance_id = ? AND deleted_at IS NULL", insuranceID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "pet", "")
	}
	return count, nil
}

func (r *insuranceRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.Insurance{}, "insurance", clinicID, ids, "sort_order")
}
