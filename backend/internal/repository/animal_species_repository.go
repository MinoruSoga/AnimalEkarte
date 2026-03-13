// Package repository provides data access implementations for AnimalSpecies entity.
package repository

import (
	"context"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// AnimalSpeciesRepository はペット種類マスタのデータアクセス層
type AnimalSpeciesRepository interface {
	FindAll(ctx context.Context) ([]model.AnimalSpecies, error)
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
