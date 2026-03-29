package service

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateShiftEntryInput はシフト作成の入力DTO
type CreateShiftEntryInput struct {
	StaffID   uint64
	Date      time.Time
	ShiftType model.ShiftType
	StartTime *string
	EndTime   *string
	Note      string
}

// UpdateShiftEntryInput はシフト更新の入力DTO（nil = 未変更）
type UpdateShiftEntryInput struct {
	ShiftType *model.ShiftType
	StartTime *string
	EndTime   *string
	Note      *string
}

// ShiftEntryService はシフト管理のビジネスロジックインターフェース
type ShiftEntryService interface {
	List(ctx context.Context, clinicID uint64, yearMonth string, staffID *uint64) ([]model.ShiftEntry, error)
	Create(ctx context.Context, clinicID uint64, input *CreateShiftEntryInput) (*model.ShiftEntry, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateShiftEntryInput) (*model.ShiftEntry, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type shiftEntryService struct {
	repo repository.ShiftEntryRepository
}

// NewShiftEntryService はShiftEntryServiceを初期化して返す
func NewShiftEntryService(repo repository.ShiftEntryRepository) ShiftEntryService {
	return &shiftEntryService{repo: repo}
}

func (s *shiftEntryService) List(ctx context.Context, clinicID uint64, yearMonth string, staffID *uint64) ([]model.ShiftEntry, error) {
	if yearMonth != "" {
		if err := validateYearMonth(yearMonth); err != nil {
			return nil, err
		}
	}
	filter := repository.ShiftEntryFilter{
		YearMonth: yearMonth,
		StaffID:   staffID,
	}
	return s.repo.List(ctx, clinicID, filter)
}

func parseTimeString(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	for _, layout := range []string{"15:04:05", "15:04"} {
		if t, err := time.Parse(layout, *s); err == nil {
			return &t
		}
	}
	return nil
}

func (s *shiftEntryService) Create(ctx context.Context, clinicID uint64, input *CreateShiftEntryInput) (*model.ShiftEntry, error) {
	entry := &model.ShiftEntry{
		ClinicID:  clinicID,
		StaffID:   input.StaffID,
		Date:      input.Date,
		ShiftType: input.ShiftType,
		StartTime: parseTimeString(input.StartTime),
		EndTime:   parseTimeString(input.EndTime),
		Note:      input.Note,
	}
	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, apperrors.Wrap(err, "failed to create shift entry")
	}
	slog.InfoContext(ctx, "shift entry created",
		slog.Uint64("shift_entry_id", entry.ID),
		slog.Uint64("clinic_id", clinicID))
	return s.repo.FindByID(ctx, clinicID, entry.ID)
}

func (s *shiftEntryService) Update(ctx context.Context, clinicID, id uint64, input *UpdateShiftEntryInput) (*model.ShiftEntry, error) {
	fields := buildShiftEntryUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update shift entry")
	}
	slog.InfoContext(ctx, "shift entry updated", slog.Uint64("shift_entry_id", id))
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *shiftEntryService) Delete(ctx context.Context, clinicID, id uint64) error {
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete shift entry")
	}
	slog.InfoContext(ctx, "shift entry deleted", slog.Uint64("shift_entry_id", id))
	return nil
}

func buildShiftEntryUpdateFields(input *UpdateShiftEntryInput) map[string]any {
	fields := map[string]any{}
	if input.ShiftType != nil {
		fields["shift_type"] = *input.ShiftType
	}
	if input.StartTime != nil {
		fields["start_time"] = parseTimeString(input.StartTime)
	}
	if input.EndTime != nil {
		fields["end_time"] = parseTimeString(input.EndTime)
	}
	if input.Note != nil {
		fields["note"] = *input.Note
	}
	return fields
}

// validateYearMonth は "YYYY-MM" 形式を検証する
func validateYearMonth(yearMonth string) error {
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}$`, yearMonth)
	if !matched {
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid year_month format: %s (expected YYYY-MM)", yearMonth))
	}
	_, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid year_month value: %s", yearMonth))
	}
	return nil
}

var _ ShiftEntryService = (*shiftEntryService)(nil)
