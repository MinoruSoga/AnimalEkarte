package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type VaccinationRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	Create(ctx context.Context, vaccination *model.Vaccination) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type vaccinationRepository struct {
	db *gorm.DB
}

func NewVaccinationRepository(db *gorm.DB) VaccinationRepository {
	return &vaccinationRepository{db: db}
}

func (r *vaccinationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error) {
	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Vaccination{}).
			Where("vaccinations.clinic_id = ?", clinicID)
		if petID != nil {
			q = q.Where("vaccinations.pet_id = ?", *petID)
		}
		if ownerID != nil {
			q = q.Joins("JOIN pets ON pets.id = vaccinations.pet_id AND pets.deleted_at IS NULL").Where("pets.owner_id = ?", *ownerID)
		}
		if startDate != nil {
			q = q.Where("vaccinations.date >= ?", *startDate)
		}
		if endDate != nil {
			q = q.Where("vaccinations.date <= ?", *endDate)
		}
		return q
	}

	var total int64
	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "vaccination", "")
	}

	vaccinations := make([]model.Vaccination, 0)
	if err := buildBase().
		Preload("Vaccine", "deleted_at IS NULL").
		Preload("Pet", "deleted_at IS NULL").
		Preload("Pet.Owner", "deleted_at IS NULL").
		Preload("Doctor", "deleted_at IS NULL").
		Offset((page - 1) * limit).Limit(limit).Order("vaccinations.date DESC, vaccinations.created_at DESC").
		Find(&vaccinations).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "vaccination", "")
	}
	return vaccinations, total, nil
}

func (r *vaccinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	var vaccination model.Vaccination
	err := r.db.WithContext(ctx).
		Where("vaccinations.id = ? AND vaccinations.clinic_id = ?", id, clinicID).
		Preload("Vaccine", "deleted_at IS NULL").Preload("Pet", "deleted_at IS NULL").Preload("Pet.Owner", "deleted_at IS NULL").Preload("Doctor", "deleted_at IS NULL").
		First(&vaccination).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vaccination", fmt.Sprintf("%d", id))
	}
	return &vaccination, nil
}

func (r *vaccinationRepository) Create(ctx context.Context, vaccination *model.Vaccination) error {
	if err := r.db.WithContext(ctx).Create(vaccination).Error; err != nil {
		return apperrors.FromGORM(err, "vaccination", "")
	}
	return nil
}

func (r *vaccinationRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Vaccination{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "vaccination", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("vaccination", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *vaccinationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Delete(&model.Vaccination{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "vaccination", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vaccination", fmt.Sprintf("%d", id))
	}
	return nil
}
