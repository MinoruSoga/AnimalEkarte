// Package service provides business logic implementations for ChiefComplaintType entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- ChiefComplaintTypeService ----

// CreateChiefComplaintTypeInput は主訴種別作成の入力DTO
type CreateChiefComplaintTypeInput struct {
	Name        string
	Description string
	IsActive    bool
	SortOrder   int
}

// UpdateChiefComplaintTypeInput holds the fields that can be updated via PATCH.
// All fields are pointers: nil means "not provided / skip".
type UpdateChiefComplaintTypeInput struct {
	Name        *string
	Description *string
	SortOrder   *int
	IsActive    *bool
}

type ChiefComplaintTypeService interface {
	List(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateChiefComplaintTypeInput) (*model.ChiefComplaintType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateChiefComplaintTypeInput) (*model.ChiefComplaintType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type chiefComplaintTypeService struct {
	repo        repository.ChiefComplaintTypeRepository
	inquiryRepo repository.InquiryRepository
}

func NewChiefComplaintTypeService(repo repository.ChiefComplaintTypeRepository, inquiryRepo repository.InquiryRepository) ChiefComplaintTypeService {
	return &chiefComplaintTypeService{repo: repo, inquiryRepo: inquiryRepo}
}

func (s *chiefComplaintTypeService) List(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list chief complaint categories")
	}
	return items, nil
}

func (s *chiefComplaintTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get chief complaint category")
	}
	return result, nil
}

func (s *chiefComplaintTypeService) Create(ctx context.Context, clinicID uint64, input *CreateChiefComplaintTypeInput) (*model.ChiefComplaintType, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	category := &model.ChiefComplaintType{
		ClinicID:    clinicID,
		Name:        input.Name,
		Description: input.Description,
		IsActive:    input.IsActive,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, category); err != nil {
		return nil, apperrors.Wrap(err, "failed to create chief complaint category")
	}
	slog.InfoContext(ctx, "chief complaint category created",
		slog.Uint64("category_id", category.ID),
		slog.Uint64("clinic_id", clinicID))
	return category, nil
}

func (s *chiefComplaintTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateChiefComplaintTypeInput) (*model.ChiefComplaintType, error) {
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	fields := buildChiefComplaintTypeUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	result, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update chief complaint category")
	}
	slog.InfoContext(ctx, "chief complaint category updated",
		slog.Uint64("category_id", id),
		slog.Uint64("clinic_id", clinicID))
	return result, nil
}

func (s *chiefComplaintTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder chief complaint categories")
	}
	slog.InfoContext(ctx, "chief_complaint_types reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}

func (s *chiefComplaintTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.inquiryRepo.CountByChiefComplaintTypeID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check inquiry dependency")
	}
	if count > 0 {
		return apperrors.WrapConflict("この主訴カテゴリは問診記録で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete chief complaint category")
	}
	slog.InfoContext(ctx, "chief complaint category deleted",
		slog.Uint64("category_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

// --- DB column constants ---

const (
	colChiefComplaintTypeName        = "name"
	colChiefComplaintTypeDescription = "description"
	colChiefComplaintTypeSortOrder   = "sort_order"
	colChiefComplaintTypeIsActive    = "is_active"
)

func buildChiefComplaintTypeUpdateFields(input *UpdateChiefComplaintTypeInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields[colChiefComplaintTypeName] = *input.Name
	}
	if input.Description != nil {
		fields[colChiefComplaintTypeDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colChiefComplaintTypeSortOrder] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields[colChiefComplaintTypeIsActive] = *input.IsActive
	}
	return fields
}
