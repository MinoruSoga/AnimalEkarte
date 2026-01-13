package repository

import (
	"context"
	"errors"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *Repository) GetAllPets(ctx context.Context) ([]model.Pet, error) {
	var pets []model.Pet
	result := r.db.WithContext(ctx).Order("created_at DESC").Find(&pets)
	if result.Error != nil {
		return nil, apperrors.Wrap(result.Error, "failed to get pets")
	}
	return pets, nil
}

func (r *Repository) GetPetByID(ctx context.Context, id uuid.UUID) (*model.Pet, error) {
	var pet model.Pet
	result := r.db.WithContext(ctx).First(&pet, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("pet", id.String())
		}
		return nil, apperrors.Wrap(result.Error, "failed to get pet")
	}
	return &pet, nil
}

func (r *Repository) CreatePet(ctx context.Context, pet *model.Pet) error {
	if err := r.db.WithContext(ctx).Create(pet).Error; err != nil {
		return apperrors.Wrap(err, "failed to create pet")
	}
	return nil
}

func (r *Repository) UpdatePet(ctx context.Context, pet *model.Pet) error {
	if err := r.db.WithContext(ctx).Save(pet).Error; err != nil {
		return apperrors.Wrap(err, "failed to update pet")
	}
	return nil
}

func (r *Repository) DeletePet(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Pet{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "failed to delete pet")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("pet", id.String())
	}
	return nil
}
