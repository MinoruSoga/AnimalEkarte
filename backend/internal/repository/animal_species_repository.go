// Package repository provides data access implementations for AnimalSpecies entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// AnimalSpeciesRepository は動物種マスタのデータアクセス層。
// 動物種はシステム全体で共有されるグローバルマスタであり、clinic_id を持たない。
type AnimalSpeciesRepository interface {
	FindAll(ctx context.Context) ([]model.AnimalSpecies, error)
	FindByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error)
	Create(ctx context.Context, species *model.AnimalSpecies) error
	Update(ctx context.Context, id uint64, fields map[string]any) (*model.AnimalSpecies, error)
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
		Order("sort_order ASC, name ASC").
		Find(&items).Error; err != nil {
		return nil, apperrors.FromGORM(err, "animal_species", "")
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
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "animal_species", "")
	}
	return nil
}

func (r *animalSpeciesRepository) Update(ctx context.Context, id uint64, fields map[string]any) (*model.AnimalSpecies, error) {
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

func (r *animalSpeciesRepository) Delete(ctx context.Context, id uint64) error {
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

// Reorder はトランザクション内で sort_order を ids の順序で更新する
func (r *animalSpeciesRepository) Reorder(ctx context.Context, ids []uint64) error {
	return reorderGlobal(ctx, r.db, &model.AnimalSpecies{}, "animal_species", ids)
}
