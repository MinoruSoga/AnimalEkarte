package pet

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Repository is the data access interface for animal species (global master).
type AnimalSpeciesRepository interface {
	FindAll(ctx context.Context) ([]model.AnimalSpecies, error)
	FindByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error)
	Create(ctx context.Context, species *model.AnimalSpecies) error
	Update(ctx context.Context, id uint64, fields map[string]any) (*model.AnimalSpecies, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, ids []uint64) error
}

type animalSpeciesRepository struct{ db *gorm.DB }

// NewAnimalSpeciesRepository constructs the global animal-species master repository.
func NewAnimalSpeciesRepository(db *gorm.DB) AnimalSpeciesRepository {
	return &animalSpeciesRepository{db: db}
}

func (r *animalSpeciesRepository) FindAll(ctx context.Context) ([]model.AnimalSpecies, error) {
	items := make([]model.AnimalSpecies, 0)
	if err := persistence.DBOrTx(ctx, r.db).
		Order("sort_order ASC, name ASC").
		Find(&items).Error; err != nil {
		return nil, apperrors.FromGORM(err, "animal_species", "")
	}
	return items, nil
}

func (r *animalSpeciesRepository) FindByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error) {
	var species model.AnimalSpecies
	err := persistence.DBOrTx(ctx, r.db).
		First(&species, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "animal_species", fmt.Sprintf("%d", id))
	}
	return &species, nil
}

func (r *animalSpeciesRepository) Create(ctx context.Context, species *model.AnimalSpecies) error {
	db := persistence.DBOrTx(ctx, r.db)
	// Capture intent before Create: gorm default:true omits zero bools from
	// INSERT and may write the DB default back into the struct.
	wantActive := species.IsActive
	if err := db.Create(species).Error; err != nil {
		return apperrors.FromGORM(err, "animal_species", "")
	}
	if !wantActive {
		if err := db.Model(species).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "animal_species", fmt.Sprintf("%d", species.ID))
		}
		species.IsActive = false
	}
	return nil
}

func (r *animalSpeciesRepository) Update(ctx context.Context, id uint64, fields map[string]any) (*model.AnimalSpecies, error) {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.AnimalSpecies{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "animal_species", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("animal_species", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, id)
}

func (r *animalSpeciesRepository) Delete(ctx context.Context, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Where("id = ?", id).
		Delete(&model.AnimalSpecies{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "animal_species", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("animal_species", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *animalSpeciesRepository) Reorder(ctx context.Context, ids []uint64) error {
	return persistence.ReorderGlobal(ctx, r.db, &model.AnimalSpecies{}, "animal_species", ids, "sort_order")
}
