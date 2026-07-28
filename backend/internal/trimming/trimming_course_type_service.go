package trimming

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CreateTrimmingCourseTypeInput はトリミングコース種別作成の入力 (#73)
type CreateTrimmingCourseTypeInput struct {
	Name      string
	SortOrder int
}

// UpdateTrimmingCourseTypeInput はトリミングコース種別更新の入力
type UpdateTrimmingCourseTypeInput struct {
	Name      *string
	SortOrder *int
	IsActive  *bool
}

const (
	colTrimmingCourseTypeName      = "name"
	colTrimmingCourseTypeSortOrder = "sort_order"
	colTrimmingCourseTypeIsActive  = "is_active"
)

func buildTrimmingCourseTypeUpdate(input *UpdateTrimmingCourseTypeInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colTrimmingCourseTypeName] = *input.Name
	}
	if input.SortOrder != nil {
		fields[colTrimmingCourseTypeSortOrder] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields[colTrimmingCourseTypeIsActive] = *input.IsActive
	}
	return fields
}

// TrimmingCourseTypeService はトリミングコース種別マスタのビジネスロジックインターフェース
type TrimmingCourseTypeService interface {
	List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type trimmingCourseTypeService struct {
	repo       TrimmingCourseTypeRepository
	transactor Transactor
}

// NewTrimmingCourseTypeService は TrimmingCourseTypeService を初期化して返す
func NewTrimmingCourseTypeService(repo TrimmingCourseTypeRepository, transactor Transactor) TrimmingCourseTypeService {
	return &trimmingCourseTypeService{repo: repo, transactor: transactor}
}

func (s *trimmingCourseTypeService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list trimming course types", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list trimming course types")
	}
	return items, nil
}

func (s *trimmingCourseTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get trimming course type", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get trimming course type")
	}
	return result, nil
}

func (s *trimmingCourseTypeService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	m := &model.TrimmingCourseType{
		ClinicID:  clinicID,
		Name:      input.Name,
		SortOrder: input.SortOrder,
	}
	result, err := s.repo.Create(ctx, m)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create trimming course type", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create trimming course type")
	}
	slog.InfoContext(ctx, "trimming course type created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("id", result.ID))
	return result, nil
}

func (s *trimmingCourseTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get trimming course type", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get trimming course type")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	fields := buildTrimmingCourseTypeUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	result, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update trimming course type", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update trimming course type")
	}
	slog.InfoContext(ctx, "trimming course type updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("id", id))
	return result, nil
}

func (s *trimmingCourseTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("trimming course type transaction dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete trimming course type")
		}
		count, err := s.repo.CountUsageByCourseTypeID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check trimming course type usage")
		}
		if count > 0 {
			return apperrors.WrapConflict("この種別は使用中のため削除できません")
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to delete trimming course type", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete trimming course type")
	}
	slog.InfoContext(ctx, "trimming course type deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("id", id))
	return nil
}

func (s *trimmingCourseTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder trimming course types", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder trimming course types")
	}
	slog.InfoContext(ctx, "trimming course types reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}
