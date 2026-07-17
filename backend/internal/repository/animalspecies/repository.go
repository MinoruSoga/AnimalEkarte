// Package animalspecies owns global animal_species master data access (no clinic_id).
package animalspecies

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for animal species (global master).
type Repository interface {
	FindAll(ctx context.Context) ([]model.AnimalSpecies, error)
	FindByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error)
	Create(ctx context.Context, species *model.AnimalSpecies) error
	Update(ctx context.Context, id uint64, fields map[string]any) (*model.AnimalSpecies, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, ids []uint64) error
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context) ([]model.AnimalSpecies, error) {
	items := make([]model.AnimalSpecies, 0)
	if err := r.db.WithContext(ctx).
		Order("sort_order ASC, name ASC").
		Find(&items).Error; err != nil {
		return nil, apperrors.FromGORM(err, "animal_species", "")
	}
	return items, nil
}

func (r *repository) FindByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error) {
	var species model.AnimalSpecies
	err := r.db.WithContext(ctx).
		First(&species, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "animal_species", fmt.Sprintf("%d", id))
	}
	return &species, nil
}

func (r *repository) Create(ctx context.Context, species *model.AnimalSpecies) error {
	if err := r.db.WithContext(ctx).Create(species).Error; err != nil {
		return apperrors.FromGORM(err, "animal_species", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, id uint64, fields map[string]any) (*model.AnimalSpecies, error) {
	result := r.db.WithContext(ctx).
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

func (r *repository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
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

func (r *repository) Reorder(ctx context.Context, ids []uint64) error {
	return repohelpers.ReorderGlobal(ctx, r.db, &model.AnimalSpecies{}, "animal_species", ids, "sort_order")
}
