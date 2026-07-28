package pet

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ChronicConditionRepository は慢性疾患フラグの永続化インターフェース（BE-012）。
type ChronicConditionRepository interface {
	FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error)
	FindByID(ctx context.Context, clinicID, petID, id uint64) (*model.PetChronicCondition, error)
	Create(ctx context.Context, record *model.PetChronicCondition) error
	Update(ctx context.Context, clinicID, petID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, petID, id uint64) error
	// FindActiveConditionCodesByOwner は飼い主の全生存ペットのアクティブ疾患コードを返す。
	FindActiveConditionCodesByOwner(ctx context.Context, clinicID, ownerID uint64) ([]string, error)
}

type chronicConditionRepository struct {
	db *gorm.DB
}

// NewChronicConditionRepository は ChronicConditionRepository を初期化して返す。
func NewChronicConditionRepository(db *gorm.DB) ChronicConditionRepository {
	return &chronicConditionRepository{db: db}
}

func (r *chronicConditionRepository) FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error) {
	var records []model.PetChronicCondition
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND pet_id = ? AND deleted_at IS NULL", clinicID, petID).
		Where("EXISTS (SELECT 1 FROM pets p WHERE p.id = pet_chronic_conditions.pet_id AND p.clinic_id = pet_chronic_conditions.clinic_id)").
		Order("diagnosed_at DESC").
		Find(&records).Error; err != nil {
		return nil, apperrors.FromGORM(err, "pet_chronic_condition", "")
	}
	return records, nil
}

func (r *chronicConditionRepository) FindByID(
	ctx context.Context,
	clinicID, petID, id uint64,
) (*model.PetChronicCondition, error) {
	var record model.PetChronicCondition
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND pet_id = ? AND id = ?", clinicID, petID, id).
		Where("EXISTS (SELECT 1 FROM pets p WHERE p.id = pet_chronic_conditions.pet_id AND p.clinic_id = pet_chronic_conditions.clinic_id)").
		First(&record).Error; err != nil {
		return nil, apperrors.FromGORM(err, "pet_chronic_condition", fmt.Sprintf("%d", id))
	}
	return &record, nil
}

func (r *chronicConditionRepository) Create(ctx context.Context, record *model.PetChronicCondition) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return apperrors.FromGORM(err, "pet_chronic_condition", "")
	}
	return nil
}

func (r *chronicConditionRepository) Update(
	ctx context.Context,
	clinicID, petID, id uint64,
	fields map[string]any,
) error {
	result := r.db.WithContext(ctx).
		Model(&model.PetChronicCondition{}).
		Where("clinic_id = ? AND pet_id = ? AND id = ? AND deleted_at IS NULL", clinicID, petID, id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "pet_chronic_condition", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("pet_chronic_condition", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *chronicConditionRepository) Delete(ctx context.Context, clinicID, petID, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("clinic_id = ? AND pet_id = ? AND id = ? AND deleted_at IS NULL", clinicID, petID, id).
		Delete(&model.PetChronicCondition{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "pet_chronic_condition", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("pet_chronic_condition", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *chronicConditionRepository) FindActiveConditionCodesByOwner(ctx context.Context, clinicID, ownerID uint64) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).Raw(`
SELECT DISTINCT pcc.condition_code
FROM pet_chronic_conditions pcc
INNER JOIN pets p ON p.id = pcc.pet_id
  AND p.clinic_id = pcc.clinic_id
  AND p.deleted_at IS NULL
WHERE pcc.clinic_id = ?
  AND p.clinic_id   = ?
  AND p.owner_id    = ?
  AND pcc.is_active = TRUE
  AND pcc.deleted_at IS NULL
  AND p.deceased_at IS NULL
ORDER BY pcc.condition_code
`, clinicID, clinicID, ownerID).Scan(&codes).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet_chronic_condition", "")
	}
	return codes, nil
}
