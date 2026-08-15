package trimming

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	colTrimmingOptionName         = "name"
	colTrimmingOptionPrice        = "price"
	colTrimmingOptionIsActive     = "is_active"
	colTrimmingOptionDescription  = "description"
	colTrimmingOptionDuration     = "duration"
	colTrimmingOptionIsCombinable = "is_combinable"
	colTrimmingOptionSortOrder    = "sort_order"
)

func buildTrimmingOptionUpdate(input *UpdateTrimmingOptionInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colTrimmingOptionName] = *input.Name
	}
	if input.Price != nil {
		fields[colTrimmingOptionPrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colTrimmingOptionIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colTrimmingOptionDescription] = *input.Description
	}
	if input.Duration != nil {
		fields[colTrimmingOptionDuration] = *input.Duration
	}
	if input.IsCombinable != nil {
		fields[colTrimmingOptionIsCombinable] = *input.IsCombinable
	}
	if input.SortOrder != nil {
		fields[colTrimmingOptionSortOrder] = *input.SortOrder
	}
	return fields
}

// CreateTrimmingOptionInput はトリミングオプション作成の入力DTO
type CreateTrimmingOptionInput struct {
	Name         string
	Price        *int64
	IsActive     bool
	Description  string
	Duration     *int
	IsCombinable bool
	SortOrder    int
}

// UpdateTrimmingOptionInput はトリミングオプション更新のサービス入力 DTO
type UpdateTrimmingOptionInput struct {
	Name         *string
	Price        *int64
	IsActive     *bool
	Description  *string
	Duration     *int
	IsCombinable *bool
	SortOrder    *int
}

// TrimmingOptionService はトリミングオプションのビジネスロジックインターフェース
type TrimmingOptionService interface {
	List(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
	Create(ctx context.Context, clinicID uint64, input *CreateTrimmingOptionInput) (*model.TrimmingOption, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingOptionInput) (*model.TrimmingOption, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type trimmingOptionService struct {
	repo       TrimmingOptionRepository
	transactor Transactor
}

// NewTrimmingOptionService は TrimmingOptionService を生成する
func NewTrimmingOptionService(repo TrimmingOptionRepository, transactor Transactor) TrimmingOptionService {
	return &trimmingOptionService{repo: repo, transactor: transactor}
}

func (s *trimmingOptionService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	result, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list trimming options", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list trimming options")
	}
	return result, nil
}

func (s *trimmingOptionService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get trimming option", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get trimming option")
	}
	return result, nil
}

func (s *trimmingOptionService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingOptionInput) (*model.TrimmingOption, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	option := &model.TrimmingOption{
		ClinicID:     clinicID,
		Name:         input.Name,
		Price:        input.Price,
		IsActive:     input.IsActive,
		Description:  input.Description,
		Duration:     input.Duration,
		IsCombinable: input.IsCombinable,
		SortOrder:    input.SortOrder,
	}
	if err := s.repo.Create(ctx, option); err != nil {
		slog.ErrorContext(ctx, "failed to create trimming option", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create trimming option")
	}
	slog.InfoContext(ctx, "trimming option created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("trimming_option_id", option.ID))
	return option, nil
}

func (s *trimmingOptionService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingOptionInput) (*model.TrimmingOption, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get trimming option", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get trimming option")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	fields := buildTrimmingOptionUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	option, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update trimming option", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update trimming option")
	}
	slog.InfoContext(ctx, "trimming option updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("trimming_option_id", id))
	return option, nil
}

func (s *trimmingOptionService) Delete(ctx context.Context, clinicID, id uint64) error {
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("trimming option transaction dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete trimming option")
		}
		count, err := s.repo.CountUsageByTrimmingOptionID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check trimming option dependencies")
		}
		if count > 0 {
			return apperrors.WrapConflict("このトリミングオプションはトリミング記録で使用中のため削除できません")
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to delete trimming option", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete trimming option")
	}
	slog.InfoContext(ctx, "trimming option deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("trimming_option_id", id))
	return nil
}

func (s *trimmingOptionService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder trimming options", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder trimming options")
	}
	slog.InfoContext(ctx, "trimming options reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}
