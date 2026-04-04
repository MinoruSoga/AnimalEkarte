package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type PetRepository interface {
	FindAll(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
	CountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	CountByAnimalSpeciesID(ctx context.Context, speciesID uint64) (int64, error)
	Create(ctx context.Context, pet *model.Pet) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type petRepository struct {
	db *gorm.DB
}

func NewPetRepository(db *gorm.DB) PetRepository {
	return &petRepository{db: db}
}

func (r *petRepository) FindAll(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error) {
	pets := make([]model.Pet, 0)
	var total int64

	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Pet{}).Where("pets.clinic_id = ?", clinicID)
		if ownerID != nil {
			q = q.Where("pets.owner_id = ?", *ownerID)
		}
		if search != "" {
			escaped := escapeLike(search)
			pattern := "%" + escaped + "%"
			q = q.Joins("LEFT JOIN owners ON owners.id = pets.owner_id AND owners.deleted_at IS NULL").
				Where(
					`(pets.name ILIKE ? ESCAPE '\' OR pets.pet_name_kana ILIKE ? ESCAPE '\' OR owners.owner_name ILIKE ? ESCAPE '\')`,
					pattern, pattern, pattern,
				)
		}
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count pets")
	}
	if err := buildBase().Preload("Owner").Preload("AnimalSpecies").Preload("Insurance").
		Offset((page - 1) * limit).Limit(limit).Order("pets.created_at DESC").Find(&pets).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find pets")
	}
	return pets, total, nil
}

func (r *petRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	var pet model.Pet
	err := r.db.WithContext(ctx).
		Preload("Owner").
		Preload("AnimalSpecies").
		Preload("Insurance").
		First(&pet, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", id))
	}
	return &pet, nil
}

func (r *petRepository) CountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Pet{}).
		Where("clinic_id = ? AND owner_id = ?", clinicID, ownerID).
		Count(&count).Error; err != nil {
		return 0, apperrors.Wrap(err, "count pets by owner")
	}
	return count, nil
}

func (r *petRepository) CountByAnimalSpeciesID(ctx context.Context, speciesID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Pet{}).
		Where("animal_species_id = ? AND deleted_at IS NULL", speciesID).
		Count(&count).Error; err != nil {
		return 0, apperrors.Wrap(err, "count pets by animal species")
	}
	return count, nil
}

func (r *petRepository) Create(ctx context.Context, pet *model.Pet) error {
	if err := r.db.WithContext(ctx).Create(pet).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("pet", "pet number already registered")
		}
		return apperrors.Wrap(err, "create pet")
	}
	loaded, err := r.FindByID(ctx, pet.ClinicID, pet.ID)
	if err != nil {
		return apperrors.Wrap(err, "reload pet after create")
	}
	*pet = *loaded
	return nil
}

func (r *petRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.Pet{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update pet")
	}
	if result.RowsAffected == 0 {
		var count int64
		r.db.WithContext(ctx).Model(&model.Pet{}).
			Where("id = ? AND clinic_id = ?", id, clinicID).
			Count(&count)
		if count == 0 {
			return apperrors.WrapNotFound("pet", fmt.Sprintf("%d", id))
		}
		// レコードは存在するが値が変わらなかった → success
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
