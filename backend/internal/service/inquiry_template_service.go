// Package service provides business logic implementations for InquiryTemplate entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- InquiryTemplateService ----

// UpdateInquiryTemplateInput holds the fields that can be updated via PATCH.
// All fields are pointers: nil means "not provided / skip".
type UpdateInquiryTemplateInput struct {
	Category  *string
	Title     *string
	Content   *string
	IsActive  *bool
	SortOrder *int
}

type InquiryTemplateService interface {
	List(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error)
	Create(ctx context.Context, template *model.InquiryTemplate) error
	Update(ctx context.Context, clinicID, id uint64, input *UpdateInquiryTemplateInput) (*model.InquiryTemplate, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type inquiryTemplateService struct {
	repo repository.InquiryTemplateRepository
}

func NewInquiryTemplateService(repo repository.InquiryTemplateRepository) InquiryTemplateService {
	return &inquiryTemplateService{repo: repo}
}

func (s *inquiryTemplateService) List(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list inquiry templates")
	}
	return items, nil
}

func (s *inquiryTemplateService) GetByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get inquiry template")
	}
	return result, nil
}

func (s *inquiryTemplateService) Create(ctx context.Context, template *model.InquiryTemplate) error {
	if err := validateRequiredName(template.Title); err != nil {
		return err
	}
	if err := s.repo.Create(ctx, template); err != nil {
		return apperrors.Wrap(err, "failed to create inquiry template")
	}
	slog.InfoContext(ctx, "inquiry template created",
		slog.Uint64("template_id", template.ID),
		slog.Uint64("clinic_id", template.ClinicID))
	return nil
}

func (s *inquiryTemplateService) Update(ctx context.Context, clinicID, id uint64, input *UpdateInquiryTemplateInput) (*model.InquiryTemplate, error) {
	if err := validateOptionalName(input.Title); err != nil {
		return nil, err
	}
	fields := buildInquiryTemplateUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update inquiry template")
	}
	slog.InfoContext(ctx, "inquiry template updated",
		slog.Uint64("template_id", id),
		slog.Uint64("clinic_id", clinicID))
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get inquiry template after update")
	}
	return result, nil
}

func (s *inquiryTemplateService) Delete(ctx context.Context, clinicID, id uint64) error {
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete inquiry template")
	}
	slog.InfoContext(ctx, "inquiry template deleted",
		slog.Uint64("template_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

func buildInquiryTemplateUpdateFields(input *UpdateInquiryTemplateInput) map[string]any {
	fields := map[string]any{}
	if input.Category != nil {
		fields["category"] = *input.Category
	}
	if input.Title != nil {
		fields["title"] = *input.Title
	}
	if input.Content != nil {
		fields["content"] = *input.Content
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}
