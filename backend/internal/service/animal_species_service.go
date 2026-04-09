// Package service provides business logic implementations for AnimalSpecies entity.
package service

import (
	"context"
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
	repo    repository.AnimalSpeciesRepository
	petRepo repository.PetRepository
}

// NewAnimalSpeciesService はAnimalSpeciesServiceを初期化して返す
func NewAnimalSpeciesService(repo repository.AnimalSpeciesRepository, petRepo repository.PetRepository) AnimalSpeciesService {
	return &animalSpeciesService{repo: repo, petRepo: petRepo}
}

func (s *animalSpeciesService) List(ctx context.Context) ([]model.AnimalSpecies, error) {
	items, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list animal species")
	}
	return items, nil
}

func (s *animalSpeciesService) GetByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error) {
	result, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get animal species")
	}
	return result, nil
}

func (s *animalSpeciesService) Create(ctx context.Context, input *CreateAnimalSpeciesInput) (*model.AnimalSpecies, error) {
	species := &model.AnimalSpecies{
		Name:      input.Name,
		IsActive:  input.IsActive,
		SortOrder: input.SortOrder,
	}
	if err := s.repo.Create(ctx, species); err != nil {
		return nil, apperrors.Wrap(err, "failed to create animal species")
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
		return nil, apperrors.Wrap(err, "failed to update animal species")
	}
	slog.InfoContext(ctx, "animal species updated",
		slog.Uint64("species_id", id))
	result, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get animal species after update")
	}
	return result, nil
}

func (s *animalSpeciesService) Delete(ctx context.Context, id uint64) error {
	count, err := s.petRepo.CountByAnimalSpeciesID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check animal species dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この動物種はペット情報で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperrors.Wrap(err, "failed to delete animal species")
	}
	slog.InfoContext(ctx, "animal species deleted", slog.Uint64("species_id", id))
	return nil
}

func (s *animalSpeciesService) Reorder(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder animal species")
	}
	return nil
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
