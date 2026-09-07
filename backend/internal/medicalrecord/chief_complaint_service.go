package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	colChiefComplaintTypeName        = "name"
	colChiefComplaintTypeDescription = "description"
	colChiefComplaintTypeSortOrder   = "sort_order"
	colChiefComplaintTypeIsActive    = "is_active"
)

func buildChiefComplaintTypeUpdate(input *UpdateChiefComplaintTypeInput) map[string]any {
	fields := make(map[string]any)
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
	repo ChiefComplaintTypeRepository
}

func NewChiefComplaintTypeService(repo ChiefComplaintTypeRepository) ChiefComplaintTypeService {
	return &chiefComplaintTypeService{repo: repo}
}

func (s *chiefComplaintTypeService) List(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list chief complaint types")
	}
	return items, nil
}

func (s *chiefComplaintTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get chief complaint type")
	}
	return result, nil
}

func (s *chiefComplaintTypeService) Create(ctx context.Context, clinicID uint64, input *CreateChiefComplaintTypeInput) (*model.ChiefComplaintType, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	category := &model.ChiefComplaintType{
		ClinicID:    clinicID,
		Name:        input.Name,
		Description: input.Description,
		IsActive:    input.IsActive,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, category); err != nil {
		return nil, apperrors.Wrap(err, "failed to create chief complaint type")
	}
	slog.InfoContext(ctx, "chief complaint type created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("chief_complaint_type_id", category.ID))
	return category, nil
}

func (s *chiefComplaintTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateChiefComplaintTypeInput) (*model.ChiefComplaintType, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get chief complaint type")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	fields := buildChiefComplaintTypeUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	result, err := s.repo.Update(ctx, clinicID, id, *input)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update chief complaint type")
	}
	slog.InfoContext(ctx, "chief complaint type updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("chief_complaint_type_id", id))
	return result, nil
}

func (s *chiefComplaintTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get chief complaint type")
	}
	count, err := s.repo.CountUsageByChiefComplaintTypeID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check chief complaint type usage")
	}
	if count > 0 {
		return apperrors.WrapConflict("この主訴カテゴリは問診記録で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete chief complaint type")
	}
	slog.InfoContext(ctx, "chief complaint type deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("chief_complaint_type_id", id))
	return nil
}

func (s *chiefComplaintTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(errMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder chief complaint types")
	}
	slog.InfoContext(ctx, "chief complaint type reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}
