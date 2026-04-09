// Package service provides business logic implementations for ChiefComplaintCategory entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- ChiefComplaintCategoryService ----

// UpdateChiefComplaintCategoryInput holds the fields that can be updated via PATCH.
// All fields are pointers: nil means "not provided / skip".
type UpdateChiefComplaintCategoryInput struct {
	Name        *string
	Description *string
	SortOrder   *int
	IsActive    *bool
}

type ChiefComplaintCategoryService interface {
	List(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintCategory, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintCategory, error)
	Create(ctx context.Context, category *model.ChiefComplaintCategory) error
	Update(ctx context.Context, clinicID, id uint64, input *UpdateChiefComplaintCategoryInput) (*model.ChiefComplaintCategory, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type chiefComplaintCategoryService struct {
	repo        repository.ChiefComplaintCategoryRepository
	inquiryRepo repository.InquiryRepository
}

func NewChiefComplaintCategoryService(repo repository.ChiefComplaintCategoryRepository, inquiryRepo repository.InquiryRepository) ChiefComplaintCategoryService {
	return &chiefComplaintCategoryService{repo: repo, inquiryRepo: inquiryRepo}
}

func (s *chiefComplaintCategoryService) List(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintCategory, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list chief complaint categories")
	}
	return items, nil
}

func (s *chiefComplaintCategoryService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintCategory, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get chief complaint category")
	}
	return result, nil
}

func (s *chiefComplaintCategoryService) Create(ctx context.Context, category *model.ChiefComplaintCategory) error {
	if err := s.repo.Create(ctx, category); err != nil {
		return apperrors.Wrap(err, "failed to create chief complaint category")
	}
	slog.InfoContext(ctx, "chief complaint category created",
		slog.Uint64("category_id", category.ID),
		slog.Uint64("clinic_id", category.ClinicID))
	return nil
}

func (s *chiefComplaintCategoryService) Update(ctx context.Context, clinicID, id uint64, input *UpdateChiefComplaintCategoryInput) (*model.ChiefComplaintCategory, error) {
	fields := buildChiefComplaintCategoryUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update chief complaint category")
	}
	slog.InfoContext(ctx, "chief complaint category updated",
		slog.Uint64("category_id", id),
		slog.Uint64("clinic_id", clinicID))
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get chief complaint category after update")
	}
	return result, nil
}

func (s *chiefComplaintCategoryService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.inquiryRepo.CountByChiefComplaintCategoryID(ctx, id)
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

func buildChiefComplaintCategoryUpdateFields(input *UpdateChiefComplaintCategoryInput) map[string]any {
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
