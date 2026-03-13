package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type PetRepository interface {
	FindAll(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
	Create(ctx context.Context, pet *model.Pet) error
	Update(ctx context.Context, pet *model.Pet) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type petRepository struct {
	db *gorm.DB
}

func NewPetRepository(db *gorm.DB) PetRepository {
	return &petRepository{db: db}
}

func (r *petRepository) FindAll(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error) {
	var pets []model.Pet
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Pet{}).Where("clinic_id = ?", clinicID)
	if ownerID != nil {
		q = q.Where("owner_id = ?", ownerID)
	}
	if search != "" {
		q = q.Where("name ILIKE ? OR pet_name_kana ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count pets")
	}
	if err := q.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&pets).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find pets")
	}
	return pets, total, nil
}

func (r *petRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	var pet model.Pet
	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Preload("Insurance").
		First(&pet, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("pet", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find pet by id")
	}
	return &pet, nil
}

func (r *petRepository) Create(ctx context.Context, pet *model.Pet) error {
	if err := r.db.WithContext(ctx).Create(pet).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("pet", pet.Name)
		}
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

func (r *petRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Pet{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete pet")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("pet", fmt.Sprintf("%d", id))
	}
	return nil
}
