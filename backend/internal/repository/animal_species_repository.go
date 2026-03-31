// Package repository provides data access implementations for AnimalSpecies entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// AnimalSpeciesRepository はペット種類マスタのデータアクセス層
type AnimalSpeciesRepository interface {
	FindAll(ctx context.Context) ([]model.AnimalSpecies, error)
	FindByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error)
	Create(ctx context.Context, species *model.AnimalSpecies) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, ids []uint64) error
}

type animalSpeciesRepository struct{ db *gorm.DB }

// NewAnimalSpeciesRepository はAnimalSpeciesRepositoryを初期化して返す
func NewAnimalSpeciesRepository(db *gorm.DB) AnimalSpeciesRepository {
	return &animalSpeciesRepository{db: db}
}

func (r *animalSpeciesRepository) FindAll(ctx context.Context) ([]model.AnimalSpecies, error) {
	items := make([]model.AnimalSpecies, 0)
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Find(&items).Error; err != nil {
		return nil, apperrors.Wrap(err, "find animal species")
	}
	return items, nil
}

func (r *animalSpeciesRepository) FindByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error) {
	var species model.AnimalSpecies
	err := r.db.WithContext(ctx).
		First(&species, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "animal_species", fmt.Sprintf("%d", id))
	}
	return &species, nil
}

func (r *animalSpeciesRepository) Create(ctx context.Context, species *model.AnimalSpecies) error {
	if err := r.db.WithContext(ctx).Create(species).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("animal_species", species.Name)
		}
		return apperrors.Wrap(err, "create animal species")
	}
	return nil
}

func (r *animalSpeciesRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.AnimalSpecies{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update animal species")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("animal_species", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *animalSpeciesRepository) Delete(ctx context.Context, id uint64) error {
	// pets テーブルで使用中かチェック
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Pet{}).
		Where("animal_species_id = ?", id).
		Count(&count).Error; err != nil {
		return apperrors.Wrap(err, "check animal species usage")
	}
	if count > 0 {
		return apperrors.WrapAlreadyExists("animal_species", fmt.Sprintf("id %d is referenced by %d pets", id, count))
	}

	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.AnimalSpecies{})
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete animal species")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("animal_species", fmt.Sprintf("%d", id))
	}
	return nil
}

// Reorder はトランザクション内で sort_order を ids の順序で更新する
func (r *animalSpeciesRepository) Reorder(ctx context.Context, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.AnimalSpecies{}).
				Where("id = ?", id).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, "reorder animal species")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("animal_species id %d not found", id))
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "reorder animal species")
	}
	return nil
}
