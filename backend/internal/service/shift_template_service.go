package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ShiftBreakTemplateInput は休憩時間テンプレートの入力DTO
type ShiftBreakTemplateInput struct {
	BreakStart string
	BreakEnd   string
}

// CreateShiftTemplateInput はシフトテンプレート作成の入力DTO
type CreateShiftTemplateInput struct {
	Name      string
	ShiftType model.ShiftType
	StartTime *string
	EndTime   *string
	Notes     string
	SortOrder int
	IsActive  bool
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

// ShiftTemplateService はシフトテンプレートのビジネスロジックインターフェース
type ShiftTemplateService interface {
	List(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error)
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

func (s *shiftTemplateService) List(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list shift templates")
	}
	return items, nil
}

func (s *shiftTemplateService) Create(ctx context.Context, clinicID uint64, input *CreateShiftTemplateInput) (*model.ShiftTemplate, error) {
	startTime := normalizeTimeString(input.StartTime)
	endTime := normalizeTimeString(input.EndTime)
	if err := validateShiftTimes(input.ShiftType, startTime, endTime); err != nil {
		return nil, err
	}
	tpl := &model.ShiftTemplate{
		ClinicID:  clinicID,
		Name:      input.Name,
		ShiftType: input.ShiftType,
		StartTime: startTime,
		EndTime:   endTime,
		Notes:     input.Notes,
		SortOrder: input.SortOrder,
		IsActive:  input.IsActive,
	}
	if err := s.repo.Create(ctx, tpl); err != nil {
		return nil, apperrors.Wrap(err, "failed to create shift template")
	}
	if len(input.Breaks) > 0 {
		breaks := make([]model.ShiftTemplateBreak, 0, len(input.Breaks))
		for _, b := range input.Breaks {
			breaks = append(breaks, model.ShiftTemplateBreak{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
		}
		if err := s.repo.ReplaceBreaks(ctx, tpl.ID, breaks); err != nil {
			return nil, apperrors.Wrap(err, "failed to save shift template breaks")
		}
	}
	slog.InfoContext(ctx, "shift template created", slog.Uint64("id", tpl.ID), slog.Uint64("clinic_id", clinicID))
	result, err := s.repo.FindByID(ctx, clinicID, tpl.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get shift template after create")
	}
	return result, nil
}

func (s *shiftTemplateService) Update(ctx context.Context, clinicID, id uint64, input *UpdateShiftTemplateInput) (*model.ShiftTemplate, error) {
	fields := buildShiftTemplateUpdateFields(input)
	if len(fields) == 0 && input.Breaks == nil {
		existing, err := s.repo.FindByID(ctx, clinicID, id)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to find shift template")
		}
		return existing, nil
	}
	if len(fields) > 0 {
		if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
			return nil, apperrors.Wrap(err, "failed to update shift template")
		}
	}
	if input.Breaks != nil {
		breaks := make([]model.ShiftTemplateBreak, 0, len(*input.Breaks))
		for _, b := range *input.Breaks {
			breaks = append(breaks, model.ShiftTemplateBreak{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
		}
		if err := s.repo.ReplaceBreaks(ctx, id, breaks); err != nil {
			return nil, apperrors.Wrap(err, "failed to save shift template breaks")
		}
	}
	slog.InfoContext(ctx, "shift template updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("id", id))
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get shift template after update")
	}
	return result, nil
}

func (s *shiftTemplateService) Delete(ctx context.Context, clinicID, id uint64) error {
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete shift template")
	}
	slog.InfoContext(ctx, "shift template deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("id", id))
	return nil
}

func (s *shiftTemplateService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder shift templates")
	}
	return nil
}

func buildShiftTemplateUpdateFields(input *UpdateShiftTemplateInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.ShiftType != nil {
		fields["shift_type"] = *input.ShiftType
	}
	if input.StartTime != nil {
		fields["start_time"] = normalizeTimeString(input.StartTime)
	}
	if input.EndTime != nil {
		fields["end_time"] = normalizeTimeString(input.EndTime)
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	return fields
}
