package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type PetRepository interface {
	FindAll(ctx context.Context, ownerID *uuid.UUID, page, limit int, search string) ([]model.Pet, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Pet, error)
	Create(ctx context.Context, pet *model.Pet) error
	Update(ctx context.Context, pet *model.Pet) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type petRepository struct {
	db *gorm.DB
}

func NewPetRepository(db *gorm.DB) PetRepository {
	return &petRepository{db: db}
}

func (r *petRepository) FindAll(ctx context.Context, ownerID *uuid.UUID, page, limit int, search string) ([]model.Pet, int64, error) {
	var pets []model.Pet
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Pet{})
	if ownerID != nil {
		q = q.Where("owner_id = ?", ownerID)
	}
	if search != "" {
		q = q.Where("name ILIKE ? OR pet_name_kana ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count pets")
	}
	if err := q.Offset((page-1)*limit).Limit(limit).Order("created_at DESC").Find(&pets).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find pets")
	}
	return pets, total, nil
}

func (r *petRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Pet, error) {
	var pet model.Pet
	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Preload("Insurance").
		First(&pet, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("pet", id.String())
		}
		return nil, apperrors.Wrap(err, "find pet by id")
	}
	return &pet, nil
}

func (r *petRepository) Create(ctx context.Context, pet *model.Pet) error {
	if err := r.db.WithContext(ctx).Create(pet).Error; err != nil {
		return apperrors.Wrap(err, "create pet")
	}
	return nil
}

func (r *petRepository) Update(ctx context.Context, pet *model.Pet) error {
	if err := r.db.WithContext(ctx).Save(pet).Error; err != nil {
		return apperrors.Wrap(err, "update pet")
	}
	return nil
}

func (r *petRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.Pet{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete pet")
	}
	return nil
}
