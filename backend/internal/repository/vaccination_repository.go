package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type VaccinationRepository interface {
	FindAll(ctx context.Context, petID *uuid.UUID, ownerID *uuid.UUID, page, limit int) ([]model.Vaccination, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Vaccination, error)
	Create(ctx context.Context, vaccination *model.Vaccination) error
	Update(ctx context.Context, vaccination *model.Vaccination) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type vaccinationRepository struct {
	db *gorm.DB
}

func NewVaccinationRepository(db *gorm.DB) VaccinationRepository {
	return &vaccinationRepository{db: db}
}

func (r *vaccinationRepository) FindAll(ctx context.Context, petID *uuid.UUID, ownerID *uuid.UUID, page, limit int) ([]model.Vaccination, int64, error) {
	var vaccinations []model.Vaccination
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Vaccination{})
	if petID != nil {
		q = q.Where("vaccinations.pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Joins("JOIN pets ON pets.id = vaccinations.pet_id").Where("pets.owner_id = ?", *ownerID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count vaccinations")
	}
	if err := q.Preload("Vaccine").Preload("Pet").Preload("Doctor").
		Offset((page - 1) * limit).Limit(limit).Order("vaccinations.date DESC, vaccinations.created_at DESC").
		Find(&vaccinations).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find vaccinations")
	}
	return vaccinations, total, nil
}

func (r *vaccinationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Vaccination, error) {
	var vaccination model.Vaccination
	if err := r.db.WithContext(ctx).
		Preload("Vaccine").
		Preload("Pet").
		Preload("Doctor").
		First(&vaccination, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("vaccination", id.String())
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

func (r *vaccinationRepository) Update(ctx context.Context, vaccination *model.Vaccination) error {
	result := r.db.WithContext(ctx).Where("id = ?", vaccination.ID).Save(vaccination)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update vaccination")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vaccination", vaccination.ID.String())
	}
	return nil
}

func (r *vaccinationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Vaccination{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete vaccination")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vaccination", id.String())
	}
	return nil
}
