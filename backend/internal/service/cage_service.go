// Package service provides business logic implementations for Cage entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- CageService ----

type CageService interface {
	List(ctx context.Context, cageType *string) ([]model.Cage, error)
	GetByID(ctx context.Context, id uint64) (*model.Cage, error)
	Create(ctx context.Context, cage *model.Cage) error
	Update(ctx context.Context, clinicID, id uint64, input UpdateCageInput) (*model.Cage, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type cageService struct {
	repo                repository.CageRepository
	hospitalizationRepo repository.HospitalizationRepository
}

func NewCageService(repo repository.CageRepository, hospitalizationRepo repository.HospitalizationRepository) CageService {
	return &cageService{repo: repo, hospitalizationRepo: hospitalizationRepo}
}

func (s *cageService) List(ctx context.Context, cageType *string) ([]model.Cage, error) {
	return s.repo.FindAll(ctx, cageType)
}
func (s *cageService) GetByID(ctx context.Context, id uint64) (*model.Cage, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *cageService) Create(ctx context.Context, cage *model.Cage) error {
	return s.repo.Create(ctx, cage)
}
func (s *cageService) Update(ctx context.Context, clinicID, id uint64, input UpdateCageInput) (*model.Cage, error) {
	fields := buildCageUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	cage, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update cage")
	}
	slog.InfoContext(ctx, "cage updated", slog.Uint64("cage_id", id))
	return cage, nil
}
func (s *cageService) Delete(ctx context.Context, id uint64) error {
	exists, err := s.hospitalizationRepo.ExistsByCageID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check hospitalization dependency")
	}
	if exists {
		return apperrors.WrapAlreadyExists("cage", "このケージは入院データで使用中のため削除できません")
	}
	return s.repo.Delete(ctx, id)
}

func (s *cageService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
}

// UpdateCageInput はケージ更新のサービス入力 DTO
type UpdateCageInput struct {
	Name        *string
	CageType    *model.CageType
	CageSize    *model.CageSize
	Price       *int64
	IsActive    *bool
	Description *string
	SortOrder   *int
}

func buildCageUpdateFields(input UpdateCageInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.CageType != nil {
		fields["cage_type"] = *input.CageType
	}
	if input.CageSize != nil {
		fields["cage_size"] = *input.CageSize
	}
	if input.Price != nil {
		fields["price"] = input.Price
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}
