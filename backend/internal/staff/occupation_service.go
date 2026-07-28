// Package service provides business logic implementations for Occupation entity.
package staff

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	colOccupationName        = "name"
	colOccupationDescription = "description"
	colOccupationIsActive    = "is_active"
	colOccupationSortOrder   = "sort_order"
)

func buildOccupationUpdate(input *UpdateOccupationInput) map[string]any {
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
	repo OccupationRepository
}

func NewOccupationService(repo OccupationRepository) OccupationService {
	return &occupationService{repo: repo}
}

func (s *occupationService) List(ctx context.Context, clinicID uint64) ([]model.Occupation, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list occupations", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list occupations")
	}
	return items, nil
}

func (s *occupationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get occupation", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get occupation")
	}
	return result, nil
}

func (s *occupationService) Create(ctx context.Context, clinicID uint64, input *CreateOccupationInput) (*model.Occupation, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	occupation := &model.Occupation{
		ClinicID:    clinicID,
		Name:        input.Name,
		Description: input.Description,
		SortOrder:   input.SortOrder,
		IsActive:    input.IsActive,
	}
	if err := s.repo.Create(ctx, occupation); err != nil {
		slog.ErrorContext(ctx, "failed to create occupation", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create occupation")
	}
	slog.InfoContext(ctx, "occupation created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("occupation_id", occupation.ID))
	return occupation, nil
}

func (s *occupationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateOccupationInput) (*model.Occupation, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	fields := buildOccupationUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	var result *model.Occupation
	if err := s.repo.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.LockActiveByIDForUpdate(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to lock occupation for update")
		}
		updated, err := s.repo.Update(txCtx, clinicID, id, fields)
		if err != nil {
			return apperrors.Wrap(err, "failed to update occupation")
		}
		result = updated
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to update occupation", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update occupation")
	}
	slog.InfoContext(ctx, "occupation updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("occupation_id", id))
	return result, nil
}

func (s *occupationService) Delete(ctx context.Context, clinicID, id uint64) error {
	if err := s.repo.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.LockActiveByIDForUpdate(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to lock occupation for delete")
		}
		count, err := s.repo.CountUsageByOccupationID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check occupation dependencies")
		}
		if count > 0 {
			return apperrors.WrapConflict("この役職はスタッフ情報で使用中のため削除できません")
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete occupation")
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to delete occupation", "error", err, "id", id, "clinic_id", clinicID)
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
		slog.ErrorContext(ctx, "failed to reorder occupations", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder occupations")
	}
	slog.InfoContext(ctx, "occupations reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}
