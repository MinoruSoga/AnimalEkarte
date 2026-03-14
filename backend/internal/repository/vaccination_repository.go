package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type VaccinationRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, page, limit int) ([]model.Vaccination, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	Create(ctx context.Context, vaccination *model.Vaccination) error
	Update(ctx context.Context, clinicID uint64, vaccination *model.Vaccination) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type vaccinationRepository struct {
	db *gorm.DB
}

func NewVaccinationRepository(db *gorm.DB) VaccinationRepository {
	return &vaccinationRepository{db: db}
}

func (r *vaccinationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, page, limit int) ([]model.Vaccination, int64, error) {
	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Vaccination{}).
			Where("vaccinations.clinic_id = ?", clinicID)
		if petID != nil {
			q = q.Where("vaccinations.pet_id = ?", *petID)
		}
		if ownerID != nil {
			q = q.Joins("JOIN pets ON pets.id = vaccinations.pet_id").Where("pets.owner_id = ?", *ownerID)
		}
		return q
	}

	var total int64
	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count vaccinations")
	}

	vaccinations := make([]model.Vaccination, 0)
	if err := buildBase().
		Preload("Vaccine").
		Preload("Pet").
		Preload("Pet.Owner").
		Preload("Doctor").
		Offset((page-1)*limit).Limit(limit).Order("vaccinations.date DESC, vaccinations.created_at DESC").
		Find(&vaccinations).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find vaccinations")
	}
	return vaccinations, total, nil
}

func (r *vaccinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	var vaccination model.Vaccination
	if err := r.db.WithContext(ctx).
		Where("vaccinations.id = ? AND vaccinations.clinic_id = ?", id, clinicID).
		Preload("Vaccine").Preload("Pet").Preload("Pet.Owner").Preload("Doctor").
		First(&vaccination).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("vaccination", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find vaccination by id")
	}
	return &vaccination, nil
}

func (r *vaccinationRepository) Create(ctx context.Context, vaccination *model.Vaccination) error {
	if err := r.db.WithContext(ctx).Create(vaccination).Error; err != nil {
		return apperrors.Wrap(err, "create vaccination")
	}
	return nil
}

func (r *vaccinationRepository) Update(ctx context.Context, clinicID uint64, vaccination *model.Vaccination) error {
	result := r.db.WithContext(ctx).
		Model(&model.Vaccination{}).
		Where("id = ? AND clinic_id = ?", vaccination.ID, clinicID).
		Updates(vaccination)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update vaccination")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vaccination", fmt.Sprintf("%d", vaccination.ID))
	}
	return nil
}

func (r *vaccinationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Delete(&model.Vaccination{})
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete vaccination")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vaccination", fmt.Sprintf("%d", id))
	}
	return nil
}
