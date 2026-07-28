package billing

import (
	"context"

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
	return persistence.FindByIDScoped[model.Insurance](ctx, r.db, "insurance", clinicID, id)
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
	if err := persistence.UpdateScopedByID(ctx, r.db, &model.Insurance{}, "insurance", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *insuranceRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DeleteScopedByID(ctx, r.db, &model.Insurance{}, "insurance", clinicID, id)
}

func (r *insuranceRepository) CountUsageByInsuranceID(ctx context.Context, clinicID, insuranceID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
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
