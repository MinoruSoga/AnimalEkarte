// Package service provides business logic implementations for Occupation entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- OccupationService ----

// CreateOccupationInput は職種作成の入力DTO
type CreateOccupationInput struct {
	Name        string
	Description string
	SortOrder   int
	IsActive    bool
}

// UpdateOccupationInput holds the fields that can be updated via PATCH.
// All fields are pointers: nil means "not provided / skip".
type UpdateOccupationInput struct {
	Name        *string
	Description *string
	SortOrder   *int
	IsActive    *bool
}

type OccupationService interface {
	List(ctx context.Context, clinicID uint64) ([]model.Occupation, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error)
	Create(ctx context.Context, clinicID uint64, input *CreateOccupationInput) (*model.Occupation, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateOccupationInput) (*model.Occupation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type occupationService struct {
	repo repository.OccupationRepository
}

func NewOccupationService(repo repository.OccupationRepository) OccupationService {
	return &occupationService{repo: repo}
}

func (s *occupationService) List(ctx context.Context, clinicID uint64) ([]model.Occupation, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list occupations")
	}
	return items, nil
}

func (s *occupationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get occupation")
	}
	return result, nil
}

func (s *occupationService) Create(ctx context.Context, clinicID uint64, input *CreateOccupationInput) (*model.Occupation, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	occupation := &model.Occupation{
		ClinicID:    clinicID,
		Name:        input.Name,
		Description: input.Description,
		SortOrder:   input.SortOrder,
		IsActive:    input.IsActive,
	}
	if err := s.repo.Create(ctx, occupation); err != nil {
		return nil, apperrors.Wrap(err, "failed to create occupation")
	}
	slog.InfoContext(ctx, "occupation created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("occupation_id", occupation.ID))
	return occupation, nil
}

func (s *occupationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateOccupationInput) (*model.Occupation, error) {
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	fields := buildOccupationUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	result, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update occupation")
	}
	slog.InfoContext(ctx, "occupation updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("occupation_id", id))
	return result, nil
}

func (s *occupationService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get occupation")
	}
	count, err := s.repo.CountStaffsByOccupationID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check occupation dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この役職はスタッフ情報で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete occupation")
	}
	slog.InfoContext(ctx, "occupation deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("occupation_id", id))
	return nil
}

func (s *occupationService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder occupations")
	}
	slog.InfoContext(ctx, "occupations reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}

const (
	colOccupationName        = "name"
	colOccupationDescription = "description"
	colOccupationSortOrder   = "sort_order"
	colOccupationIsActive    = "is_active"
)

func buildOccupationUpdateFields(input *UpdateOccupationInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colOccupationName] = *input.Name
	}
	if input.Description != nil {
		fields[colOccupationDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colOccupationSortOrder] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields[colOccupationIsActive] = *input.IsActive
	}
	return fields
}
