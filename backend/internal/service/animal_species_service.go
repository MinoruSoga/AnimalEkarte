// Package service provides business logic implementations for AnimalSpecies entity.
package service

import (
	"context"
	"fmt"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// 列名定数
const (
	colAnimalSpeciesName      = "name"
	colAnimalSpeciesIsActive  = "is_active"
	colAnimalSpeciesSortOrder = "sort_order"
)

// ---- Input DTOs ----

// CreateAnimalSpeciesInput は動物種類作成の入力DTO
type CreateAnimalSpeciesInput struct {
	Name      string
	IsActive  bool
	SortOrder int
}

// UpdateAnimalSpeciesInput は動物種類更新の入力DTO（nil = 未指定 = 更新しない）
type UpdateAnimalSpeciesInput struct {
	Name      *string
	IsActive  *bool
	SortOrder *int
}

// AnimalSpeciesService はペット種類マスタのビジネスロジック層
type AnimalSpeciesService interface {
	List(ctx context.Context) ([]model.AnimalSpecies, error)
	GetByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error)
	Create(ctx context.Context, input *CreateAnimalSpeciesInput) (*model.AnimalSpecies, error)
	Update(ctx context.Context, id uint64, input *UpdateAnimalSpeciesInput) (*model.AnimalSpecies, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, ids []uint64) error
}

type animalSpeciesService struct {
	repo repository.AnimalSpeciesRepository
}

// NewAnimalSpeciesService はAnimalSpeciesServiceを初期化して返す
func NewAnimalSpeciesService(repo repository.AnimalSpeciesRepository) AnimalSpeciesService {
	return &animalSpeciesService{repo: repo}
}

func (s *animalSpeciesService) List(ctx context.Context) ([]model.AnimalSpecies, error) {
	return s.repo.FindAll(ctx)
}

func (s *animalSpeciesService) GetByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *animalSpeciesService) Create(ctx context.Context, input *CreateAnimalSpeciesInput) (*model.AnimalSpecies, error) {
	species := &model.AnimalSpecies{
		Name:      input.Name,
		IsActive:  input.IsActive,
		SortOrder: input.SortOrder,
	}
	if err := s.repo.Create(ctx, species); err != nil {
		return nil, fmt.Errorf("failed to create animal species: %w", err)
	}
	slog.InfoContext(ctx, "animal species created",
		slog.Uint64("species_id", species.ID),
		slog.String("name", species.Name))
	return species, nil
}

func (s *animalSpeciesService) Update(ctx context.Context, id uint64, input *UpdateAnimalSpeciesInput) (*model.AnimalSpecies, error) {
	fields := buildAnimalSpeciesUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, id, fields); err != nil {
		return nil, fmt.Errorf("failed to update animal species: %w", err)
	}
	slog.InfoContext(ctx, "animal species updated",
		slog.Uint64("species_id", id))
	return s.repo.FindByID(ctx, id)
}

func (s *animalSpeciesService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *animalSpeciesService) Reorder(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, ids)
}

// buildAnimalSpeciesUpdateFields はポインタが非 nil のフィールドのみ map に追加する
func buildAnimalSpeciesUpdateFields(input *UpdateAnimalSpeciesInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colAnimalSpeciesName] = *input.Name
	}
	if input.IsActive != nil {
		fields[colAnimalSpeciesIsActive] = *input.IsActive
	}
	if input.SortOrder != nil {
		fields[colAnimalSpeciesSortOrder] = *input.SortOrder
	}
	return fields
}
