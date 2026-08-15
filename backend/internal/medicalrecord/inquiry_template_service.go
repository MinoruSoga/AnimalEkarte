package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	colInquiryTemplateCategory  = "category"
	colInquiryTemplateTitle     = "title"
	colInquiryTemplateContent   = "content"
	colInquiryTemplateIsActive  = "is_active"
	colInquiryTemplateSortOrder = "sort_order"
)

func buildInquiryTemplateUpdate(input *UpdateInquiryTemplateInput) map[string]any {
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

// ---- InquiryTemplateService ----

// CreateInquiryTemplateInput は問診定型文作成の入力DTO
type CreateInquiryTemplateInput struct {
	Category  string
	Title     string
	Content   string
	IsActive  bool
	SortOrder int
}

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
	Create(ctx context.Context, clinicID uint64, input *CreateInquiryTemplateInput) (*model.InquiryTemplate, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateInquiryTemplateInput) (*model.InquiryTemplate, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type inquiryTemplateService struct {
	repo InquiryTemplateRepository
}

func NewInquiryTemplateService(repo InquiryTemplateRepository) InquiryTemplateService {
	return &inquiryTemplateService{repo: repo}
}

func (s *inquiryTemplateService) List(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list inquiry templates", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list inquiry templates")
	}
	return items, nil
}

func (s *inquiryTemplateService) GetByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get inquiry template", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get inquiry template")
	}
	return result, nil
}

func (s *inquiryTemplateService) Create(ctx context.Context, clinicID uint64, input *CreateInquiryTemplateInput) (*model.InquiryTemplate, error) {
	if err := validateRequiredName(input.Title); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
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
		slog.ErrorContext(ctx, "failed to create inquiry template", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create inquiry template")
	}
	slog.InfoContext(ctx, "inquiry template created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("inquiry_template_id", template.ID))
	return template, nil
}

func (s *inquiryTemplateService) Update(ctx context.Context, clinicID, id uint64, input *UpdateInquiryTemplateInput) (*model.InquiryTemplate, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get inquiry template", "error", err)
		return nil, apperrors.Wrap(err, "failed to get inquiry template")
	}
	if err := validateOptionalName(input.Title); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	fields := buildInquiryTemplateUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	result, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update inquiry template", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update inquiry template")
	}
	slog.InfoContext(ctx, "inquiry template updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("inquiry_template_id", id))
	return result, nil
}

func (s *inquiryTemplateService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find inquiry template")
	}
	count, err := s.repo.CountUsageByInquiryTemplateID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check inquiry template usage", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check inquiry template usage")
	}
	if count > 0 {
		return apperrors.WrapConflict("この問診定型文は使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete inquiry template", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete inquiry template")
	}
	slog.InfoContext(ctx, "inquiry template deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("inquiry_template_id", id))
	return nil
}

func (s *inquiryTemplateService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(errMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder inquiry templates", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder inquiry templates")
	}
	slog.InfoContext(ctx, "inquiry templates reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}
