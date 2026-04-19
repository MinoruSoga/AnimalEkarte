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

// CreateInquiryTemplateInput は問診定型文作成の入力DTO
type CreateInquiryTemplateInput struct {
	Category  string
	Title     string
	Content   string
	IsActive  bool
	SortOrder int
}

type InquiryTemplateService interface {
	List(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error)
	Create(ctx context.Context, clinicID uint64, input *CreateInquiryTemplateInput) (*model.InquiryTemplate, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateInquiryTemplateInput) (*model.InquiryTemplate, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
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

func (s *inquiryTemplateService) Create(ctx context.Context, clinicID uint64, input *CreateInquiryTemplateInput) (*model.InquiryTemplate, error) {
	if err := validateRequiredName(input.Title); err != nil {
		return nil, err
	}
	template := &model.InquiryTemplate{
		ClinicID:  clinicID,
		Category:  input.Category,
		Title:     input.Title,
		Content:   input.Content,
		IsActive:  input.IsActive,
		SortOrder: input.SortOrder,
	}
	if err := s.repo.Create(ctx, template); err != nil {
		return nil, apperrors.Wrap(err, "failed to create inquiry template")
	}
	slog.InfoContext(ctx, "inquiry template created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("inquiry_template_id", template.ID))
	return template, nil
}

func (s *inquiryTemplateService) Update(ctx context.Context, clinicID, id uint64, input *UpdateInquiryTemplateInput) (*model.InquiryTemplate, error) {
	if err := validateOptionalName(input.Title); err != nil {
		return nil, err
	}
	fields := buildInquiryTemplateUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	result, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update inquiry template")
	}
	slog.InfoContext(ctx, "inquiry template updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("inquiry_template_id", id))
	return result, nil
}

func (s *inquiryTemplateService) Delete(ctx context.Context, clinicID, id uint64) error {
	// 設計上の意図: inquiry_templates を参照する外部テーブル（inquiry_template_id FK）は
	// 現スキーマに存在しないため、依存チェックなしで直接削除する。
	// 将来 inquiry_answers 等が追加された場合は CountUsageByTemplateID を実装すること。
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete inquiry template")
	}
	slog.InfoContext(ctx, "inquiry template deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("inquiry_template_id", id))
	return nil
}

func (s *inquiryTemplateService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder inquiry templates")
	}
	slog.InfoContext(ctx, "inquiry templates reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}

const (
	colInquiryTemplateCategory  = "category"
	colInquiryTemplateTitle     = "title"
	colInquiryTemplateContent   = "content"
	colInquiryTemplateIsActive  = "is_active"
	colInquiryTemplateSortOrder = "sort_order"
)

func buildInquiryTemplateUpdateFields(input *UpdateInquiryTemplateInput) map[string]any {
	fields := map[string]any{}
	if input.Category != nil {
		fields[colInquiryTemplateCategory] = *input.Category
	}
	if input.Title != nil {
		fields[colInquiryTemplateTitle] = *input.Title
	}
	if input.Content != nil {
		fields[colInquiryTemplateContent] = *input.Content
	}
	if input.IsActive != nil {
		fields[colInquiryTemplateIsActive] = *input.IsActive
	}
	if input.SortOrder != nil {
		fields[colInquiryTemplateSortOrder] = *input.SortOrder
	}
	return fields
}
