package staff

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	colShiftTemplateName      = "name"
	colShiftTemplateShiftType = "shift_type"
	colShiftTemplateStartTime = "start_time"
	colShiftTemplateEndTime   = "end_time"
	colShiftTemplateNotes     = "notes"
	colShiftTemplateSortOrder = "sort_order"
	colShiftTemplateIsActive  = "is_active"
)

// ShiftBreakTemplateInput は休憩時間テンプレートの入力DTO
type ShiftBreakTemplateInput struct {
	BreakStart string
	BreakEnd   string
}

// CreateShiftTemplateInput はシフトテンプレート作成の入力DTO
type CreateShiftTemplateInput struct {
	Name      string
	ShiftType string
	StartTime string
	EndTime   string
	Notes     string
	SortOrder int
	IsActive  *bool // nil = デフォルト true（Service内で処理）
	Breaks    []ShiftBreakTemplateInput
}

// UpdateShiftTemplateInput はシフトテンプレート更新の入力DTO（nil = 未変更）
type UpdateShiftTemplateInput struct {
	Name      *string
	ShiftType *model.ShiftType
	StartTime *string
	EndTime   *string
	Notes     *string
	SortOrder *int
	IsActive  *bool
	Breaks    *[]ShiftBreakTemplateInput
}

func buildShiftTemplateUpdate(input *UpdateShiftTemplateInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields[colShiftTemplateName] = *input.Name
	}
	if input.ShiftType != nil {
		fields[colShiftTemplateShiftType] = *input.ShiftType
	}
	if input.StartTime != nil {
		fields[colShiftTemplateStartTime] = normalizeTimeString(input.StartTime)
	}
	if input.EndTime != nil {
		fields[colShiftTemplateEndTime] = normalizeTimeString(input.EndTime)
	}
	if input.Notes != nil {
		fields[colShiftTemplateNotes] = *input.Notes
	}
	if input.SortOrder != nil {
		fields[colShiftTemplateSortOrder] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields[colShiftTemplateIsActive] = *input.IsActive
	}
	return fields
}

// ShiftTemplateService はシフトテンプレートのビジネスロジックインターフェース
type ShiftTemplateService interface {
	List(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error)
	Create(ctx context.Context, clinicID uint64, input *CreateShiftTemplateInput) (*model.ShiftTemplate, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateShiftTemplateInput) (*model.ShiftTemplate, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type shiftTemplateService struct {
	repo ShiftTemplateRepository
}

// NewShiftTemplateService は ShiftTemplateService を初期化して返す
func NewShiftTemplateService(repo ShiftTemplateRepository) ShiftTemplateService {
	return &shiftTemplateService{repo: repo}
}

func (s *shiftTemplateService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get shift template", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get shift template")
	}
	return result, nil
}

func (s *shiftTemplateService) List(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list shift templates", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list shift templates")
	}
	return items, nil
}

func (s *shiftTemplateService) Create(ctx context.Context, clinicID uint64, input *CreateShiftTemplateInput) (*model.ShiftTemplate, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("shift template input is required")
	}
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}
	var startTimePtr, endTimePtr *string
	if input.StartTime != "" {
		startTimePtr = &input.StartTime
	}
	if input.EndTime != "" {
		endTimePtr = &input.EndTime
	}
	shiftType := model.ShiftType(input.ShiftType)
	if err := validateShiftType(shiftType); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate shift type")
	}
	startTime := normalizeTimeString(startTimePtr)
	endTime := normalizeTimeString(endTimePtr)
	if err := validateShiftTimes(shiftType, startTime, endTime); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate shift times")
	}
	template := &model.ShiftTemplate{
		ClinicID:  clinicID,
		Name:      input.Name,
		ShiftType: shiftType,
		StartTime: startTime,
		EndTime:   endTime,
		Notes:     input.Notes,
		SortOrder: input.SortOrder,
		IsActive:  isActive,
	}
	breaks := make([]model.ShiftTemplateBreak, 0, len(input.Breaks))
	for _, item := range input.Breaks {
		breaks = append(
			breaks,
			model.ShiftTemplateBreak{BreakStart: item.BreakStart, BreakEnd: item.BreakEnd},
		)
	}

	var result *model.ShiftTemplate
	if err := s.repo.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, template); err != nil {
			// BUG-026: map measured name unique_violation to stable domain code + name echo.
			if conflict := apperrors.AsNameUniqueConflict(
				err,
				input.Name,
				apperrors.ConstraintShiftTemplateName,
				apperrors.CodeShiftTemplateNameConflict,
			); conflict != nil {
				return conflict
			}
			return apperrors.Wrap(err, "failed to create shift template")
		}
		if len(breaks) > 0 {
			if err := s.repo.UpdateBreaks(txCtx, template.ID, breaks); err != nil {
				return apperrors.Wrap(err, "failed to save shift template breaks")
			}
		}
		created, err := s.repo.FindByID(txCtx, clinicID, template.ID)
		if err != nil {
			return apperrors.Wrap(err, "failed to get shift template after create")
		}
		result = created
		return nil
	}); err != nil {
		if apperrors.IsNameConflict(err, apperrors.CodeShiftTemplateNameConflict) {
			return nil, err
		}
		slog.ErrorContext(ctx, "failed to create shift template", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create shift template")
	}
	slog.InfoContext(ctx, "shift template created", slog.Uint64("clinic_id", clinicID), slog.Uint64("shift_template_id", template.ID))
	return result, nil
}

func (s *shiftTemplateService) Update(ctx context.Context, clinicID, id uint64, input *UpdateShiftTemplateInput) (*model.ShiftTemplate, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("shift template input is required")
	}
	fields := buildShiftTemplateUpdate(input)
	if len(fields) == 0 && input.Breaks == nil {
		existing, err := s.repo.FindByID(ctx, clinicID, id)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find shift template", "error", err, "id", id, "clinic_id", clinicID)
			return nil, apperrors.Wrap(err, "failed to find shift template")
		}
		return existing, nil
	}
	breaks := make([]model.ShiftTemplateBreak, 0)
	if input.Breaks != nil {
		breaks = make([]model.ShiftTemplateBreak, 0, len(*input.Breaks))
		for _, item := range *input.Breaks {
			breaks = append(
				breaks,
				model.ShiftTemplateBreak{
					BreakStart: item.BreakStart,
					BreakEnd:   item.BreakEnd,
				},
			)
		}
	}

	var result *model.ShiftTemplate
	if err := s.repo.WithTx(ctx, func(txCtx context.Context) error {
		existing, err := s.repo.LockActiveByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock shift template for update")
		}

		effectiveShiftType := existing.ShiftType
		if input.ShiftType != nil {
			if err := validateShiftType(*input.ShiftType); err != nil {
				return apperrors.Wrap(err, "failed to validate shift type")
			}
			effectiveShiftType = *input.ShiftType
		}
		effectiveStart := existing.StartTime
		effectiveEnd := existing.EndTime
		if input.StartTime != nil {
			effectiveStart = normalizeTimeString(input.StartTime)
		}
		if input.EndTime != nil {
			effectiveEnd = normalizeTimeString(input.EndTime)
		}
		if err := validateShiftTimes(effectiveShiftType, effectiveStart, effectiveEnd); err != nil {
			return apperrors.Wrap(err, "failed to validate shift times")
		}

		if len(fields) > 0 {
			if err := s.applyShiftTemplateFieldUpdate(txCtx, clinicID, id, fields, input); err != nil {
				return err
			}
		}
		if input.Breaks != nil {
			if err := s.repo.UpdateBreaks(txCtx, id, breaks); err != nil {
				return apperrors.Wrap(err, "failed to save shift template breaks")
			}
		}
		updated, err := s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get shift template after update")
		}
		result = updated
		return nil
	}); err != nil {
		if apperrors.IsNameConflict(err, apperrors.CodeShiftTemplateNameConflict) {
			return nil, err
		}
		slog.ErrorContext(ctx, "failed to update shift template", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update shift template")
	}
	slog.InfoContext(ctx, "shift template updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("shift_template_id", id))
	return result, nil
}

func (s *shiftTemplateService) Delete(ctx context.Context, clinicID, id uint64) error {
	if err := s.repo.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.LockActiveByIDForUpdate(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to lock shift template for delete")
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete shift template")
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to delete shift template", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete shift template")
	}
	slog.InfoContext(ctx, "shift template deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("shift_template_id", id))
	return nil
}

func (s *shiftTemplateService) applyShiftTemplateFieldUpdate(
	ctx context.Context,
	clinicID, id uint64,
	fields map[string]any,
	input *UpdateShiftTemplateInput,
) error {
	if _, err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		nameForConflict := ""
		if input.Name != nil {
			nameForConflict = *input.Name
		}
		if conflict := apperrors.AsNameUniqueConflict(
			err,
			nameForConflict,
			apperrors.ConstraintShiftTemplateName,
			apperrors.CodeShiftTemplateNameConflict,
		); conflict != nil {
			return conflict
		}
		return apperrors.Wrap(err, "failed to update shift template")
	}
	return nil
}

func (s *shiftTemplateService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder shift templates", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder shift templates")
	}
	slog.InfoContext(ctx, "shift templates reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}
