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

// CreateCageInput はケージ作成の入力DTO
type CreateCageInput struct {
	Name        string
	CageType    string
	CageSize    string
	Price       *int64
	IsActive    bool
	Description string
	SortOrder   int
}

type CageService interface {
	List(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	Create(ctx context.Context, clinicID uint64, input *CreateCageInput) (*model.Cage, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type cageService struct {
	repo                repository.CageRepository
	hospitalizationRepo repository.HospitalizationRepository
}

func NewCageService(repo repository.CageRepository, hospitalizationRepo repository.HospitalizationRepository) CageService {
	return &cageService{repo: repo, hospitalizationRepo: hospitalizationRepo}
}

func (s *cageService) List(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error) {
	result, err := s.repo.FindAll(ctx, clinicID, cageType)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list cage")
	}
	return result, nil
}
func (s *cageService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get cage")
	}
	return result, nil
}
func (s *cageService) Create(ctx context.Context, clinicID uint64, input *CreateCageInput) (*model.Cage, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	if err := validateCageType(input.CageType); err != nil {
		return nil, err
	}
	if err := validateCageSize(input.CageSize); err != nil {
		return nil, err
	}
	cage := &model.Cage{
		ClinicID:    clinicID,
		Name:        input.Name,
		CageType:    model.CageType(input.CageType),
		CageSize:    model.CageSize(input.CageSize),
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, cage); err != nil {
		return nil, apperrors.Wrap(err, "failed to create cage")
	}
	slog.InfoContext(ctx, "cage created",
		slog.Uint64("clinic_id", cage.ClinicID),
		slog.Uint64("cage_id", cage.ID))
	return cage, nil
}
func (s *cageService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error) {
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	if input.CageType != nil {
		if err := validateCageType(*input.CageType); err != nil {
			return nil, err
		}
	}
	if input.CageSize != nil {
		if err := validateCageSize(*input.CageSize); err != nil {
			return nil, err
		}
	}
	fields := buildCageUpdateFields(*input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	cage, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update cage")
	}
	slog.InfoContext(ctx, "cage updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("cage_id", id))
	return cage, nil
}
func (s *cageService) Delete(ctx context.Context, clinicID, id uint64) error {
	exists, err := s.hospitalizationRepo.ExistsByCageID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check hospitalization dependency")
	}
	if exists {
		return apperrors.WrapConflict("このケージは入院データで使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete cage")
	}
	slog.InfoContext(ctx, "cage deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("cage_id", id))
	return nil
}

func (s *cageService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder cage")
	}
	slog.InfoContext(ctx, "cage reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}

// UpdateCageInput はケージ更新のサービス入力 DTO
type UpdateCageInput struct {
	Name        *string
	CageType    *string
	CageSize    *string
	Price       *int64
	IsActive    *bool
	Description *string
	SortOrder   *int
}

const (
	colCageName        = "name"
	colCageCageType    = "cage_type"
	colCageCageSize    = "cage_size"
	colCagePrice       = "price"
	colCageIsActive    = "is_active"
	colCageDescription = "description"
	colCageSortOrder   = "sort_order"
)

func buildCageUpdateFields(input UpdateCageInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colCageName] = *input.Name
	}
	if input.CageType != nil {
		fields[colCageCageType] = model.CageType(*input.CageType)
	}
	if input.CageSize != nil {
		fields[colCageCageSize] = model.CageSize(*input.CageSize)
	}
	if input.Price != nil {
		fields[colCagePrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colCageIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colCageDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colCageSortOrder] = *input.SortOrder
	}
	return fields
}
