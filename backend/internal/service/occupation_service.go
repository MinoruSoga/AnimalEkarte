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
	Create(ctx context.Context, occupation *model.Occupation) error
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
	return s.repo.FindAll(ctx, clinicID)
}

func (s *occupationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *occupationService) Create(ctx context.Context, occupation *model.Occupation) error {
	if err := s.repo.Create(ctx, occupation); err != nil {
		return apperrors.Wrap(err, "failed to create occupation")
	}
	slog.InfoContext(ctx, "occupation created",
		slog.Uint64("occupation_id", occupation.ID),
		slog.Uint64("clinic_id", occupation.ClinicID))
	return nil
}

func (s *occupationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateOccupationInput) (*model.Occupation, error) {
	fields := buildOccupationUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update occupation")
	}
	slog.InfoContext(ctx, "occupation updated",
		slog.Uint64("occupation_id", id),
		slog.Uint64("clinic_id", clinicID))
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *occupationService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountStaffsByOccupationID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check occupation dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この役職はスタッフ情報で使用中のため削除できません")
	}
	return s.repo.Delete(ctx, clinicID, id)
}

func (s *occupationService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return s.repo.Reorder(ctx, clinicID, ids)
}

func buildOccupationUpdateFields(input *UpdateOccupationInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	return fields
}
