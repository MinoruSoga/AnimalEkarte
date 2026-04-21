package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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

func buildShiftTemplateUpdateFields(input *UpdateShiftTemplateInput) map[string]any {
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
	repo repository.ShiftTemplateRepository
}

// NewShiftTemplateService は ShiftTemplateService を初期化して返す
func NewShiftTemplateService(repo repository.ShiftTemplateRepository) ShiftTemplateService {
	return &shiftTemplateService{repo: repo}
}

func (s *shiftTemplateService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get shift template", "error", err)
		return nil, apperrors.Wrap(err, "failed to get shift template")
	}
	return result, nil
}

func (s *shiftTemplateService) List(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list shift templates", "error", err)
		return nil, apperrors.Wrap(err, "failed to list shift templates")
	}
	return items, nil
}

func (s *shiftTemplateService) Create(ctx context.Context, clinicID uint64, input *CreateShiftTemplateInput) (*model.ShiftTemplate, error) {
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
	startTime := normalizeTimeString(startTimePtr)
	endTime := normalizeTimeString(endTimePtr)
	if err := validateShiftTimes(shiftType, startTime, endTime); err != nil {
		return nil, err
	}
	tpl := &model.ShiftTemplate{
		ClinicID:  clinicID,
		Name:      input.Name,
		ShiftType: shiftType,
		StartTime: startTime,
		EndTime:   endTime,
		Notes:     input.Notes,
		SortOrder: input.SortOrder,
		IsActive:  isActive,
	}
	if err := s.repo.Create(ctx, tpl); err != nil {
		slog.ErrorContext(ctx, "failed to create shift template", "error", err)
		return nil, apperrors.Wrap(err, "failed to create shift template")
	}
	if len(input.Breaks) > 0 {
		breaks := make([]model.ShiftTemplateBreak, 0, len(input.Breaks))
		for _, b := range input.Breaks {
			breaks = append(breaks, model.ShiftTemplateBreak{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
		}
		if err := s.repo.ReplaceBreaks(ctx, tpl.ID, breaks); err != nil {
			slog.ErrorContext(ctx, "failed to save shift template breaks", "error", err)
			return nil, apperrors.Wrap(err, "failed to save shift template breaks")
		}
	}
	slog.InfoContext(ctx, "shift template created", slog.Uint64("clinic_id", clinicID), slog.Uint64("shift_template_id", tpl.ID))
	result, err := s.repo.FindByID(ctx, clinicID, tpl.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get shift template after create", "error", err)
		return nil, apperrors.Wrap(err, "failed to get shift template after create")
	}
	return result, nil
}

func (s *shiftTemplateService) Update(ctx context.Context, clinicID, id uint64, input *UpdateShiftTemplateInput) (*model.ShiftTemplate, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get shift template", "error", err)
		return nil, apperrors.Wrap(err, "failed to get shift template")
	}
	fields := buildShiftTemplateUpdateFields(input)
	if len(fields) == 0 && input.Breaks == nil {
		existing, err := s.repo.FindByID(ctx, clinicID, id)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find shift template", "error", err)
			return nil, apperrors.Wrap(err, "failed to find shift template")
		}
		return existing, nil
	}
	var result *model.ShiftTemplate
	if len(fields) > 0 {
		var err error
		result, err = s.repo.UpdateFields(ctx, clinicID, id, fields)
		if err != nil {
			slog.ErrorContext(ctx, "failed to update shift template", "error", err)
			return nil, apperrors.Wrap(err, "failed to update shift template")
		}
	}
	if input.Breaks != nil {
		breaks := make([]model.ShiftTemplateBreak, 0, len(*input.Breaks))
		for _, b := range *input.Breaks {
			breaks = append(breaks, model.ShiftTemplateBreak{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
		}
		if err := s.repo.ReplaceBreaks(ctx, id, breaks); err != nil {
			slog.ErrorContext(ctx, "failed to save shift template breaks", "error", err)
			return nil, apperrors.Wrap(err, "failed to save shift template breaks")
		}
		var err error
		result, err = s.repo.FindByID(ctx, clinicID, id)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get shift template after update", "error", err)
			return nil, apperrors.Wrap(err, "failed to get shift template after update")
		}
	}
	slog.InfoContext(ctx, "shift template updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("shift_template_id", id))
	return result, nil
}

func (s *shiftTemplateService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get shift template", "error", err)
		return apperrors.Wrap(err, "failed to get shift template")
	}
	count, err := s.repo.CountUsageByShiftTemplateID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count shift template usage", "error", err)
		return apperrors.Wrap(err, "failed to check shift template usage")
	}
	if count > 0 {
		return apperrors.WrapConflict("このシフトテンプレートは使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete shift template", "error", err)
		return apperrors.Wrap(err, "failed to delete shift template")
	}
	slog.InfoContext(ctx, "shift template deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("shift_template_id", id))
	return nil
}

func (s *shiftTemplateService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder shift templates", "error", err)
		return apperrors.Wrap(err, "failed to reorder shift templates")
	}
	slog.InfoContext(ctx, "shift templates reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}
