package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// PetChronicConditionRepository は慢性疾患フラグの永続化インターフェース（BE-012）。
type PetChronicConditionRepository interface {
	FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.PetChronicCondition, error)
	Create(ctx context.Context, record *model.PetChronicCondition) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	// FindActiveConditionCodesByOwner は飼い主の全生存ペットのアクティブ疾患コードを返す。
	FindActiveConditionCodesByOwner(ctx context.Context, clinicID, ownerID uint64) ([]string, error)
}

type petChronicConditionRepository struct {
	db *gorm.DB
}

// NewPetChronicConditionRepository は PetChronicConditionRepository を初期化して返す。
func NewPetChronicConditionRepository(db *gorm.DB) PetChronicConditionRepository {
	return &petChronicConditionRepository{db: db}
}

func (r *petChronicConditionRepository) FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error) {
	var records []model.PetChronicCondition
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND pet_id = ? AND deleted_at IS NULL", clinicID, petID).
		Order("diagnosed_at DESC").
		Find(&records).Error; err != nil {
		return nil, apperrors.FromGORM(err, "pet_chronic_condition", "")
	}
	return records, nil
}

func (r *petChronicConditionRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.PetChronicCondition, error) {
	var record model.PetChronicCondition
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND id = ?", clinicID, id).
		First(&record).Error; err != nil {
		return nil, apperrors.FromGORM(err, "pet_chronic_condition", fmt.Sprintf("%d", id))
	}
	return &record, nil
}

func (r *petChronicConditionRepository) Create(ctx context.Context, record *model.PetChronicCondition) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return apperrors.FromGORM(err, "pet_chronic_condition", "")
	}
	return nil
}

func (r *petChronicConditionRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return updateScopedByID(ctx, r.db, &model.PetChronicCondition{}, "pet_chronic_condition", clinicID, id, fields)
}

func (r *petChronicConditionRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return deleteScopedByID(ctx, r.db, &model.PetChronicCondition{}, "pet_chronic_condition", clinicID, id)
}

func (r *petChronicConditionRepository) FindActiveConditionCodesByOwner(ctx context.Context, clinicID, ownerID uint64) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).Raw(`
SELECT DISTINCT pcc.condition_code
FROM pet_chronic_conditions pcc
INNER JOIN pets p ON p.id = pcc.pet_id AND p.deleted_at IS NULL
WHERE pcc.clinic_id = ?
  AND p.owner_id    = ?
  AND pcc.is_active = TRUE
  AND pcc.deleted_at IS NULL
  AND p.deceased_at IS NULL
ORDER BY pcc.condition_code
`, clinicID, ownerID).Scan(&codes).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet_chronic_condition", "")
	}
	return codes, nil
}
