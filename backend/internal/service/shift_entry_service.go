package service

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// normalizeTimeString は "15:04" または "15:04:05" 形式の時刻文字列を "15:04:05" に正規化する。
// 不正な形式の場合は nil を返す。
// normalizeTimeString は sharedkernel.NormalizeTimeString への delegate（BE9-2C R② 昇格）。
func normalizeTimeString(s *string) *string {
	return sharedkernel.NormalizeTimeString(s)
}

// ShiftBreakInput は休憩時間の入力DTO
type ShiftBreakInput struct {
	BreakStart string
	BreakEnd   string
}

// CreateShiftEntryInput はシフト作成の入力DTO
type CreateShiftEntryInput struct {
	StaffID   uint64
	Date      time.Time
	ShiftType string
	StartTime *string
	EndTime   *string
	Notes     string
	Breaks    []ShiftBreakInput
}

// UpdateShiftEntryInput はシフト更新の入力DTO（nil = 未変更）
type UpdateShiftEntryInput struct {
	ShiftType *string
	StartTime *string
	EndTime   *string
	Notes     *string
	Breaks    *[]ShiftBreakInput
}

// ShiftEntryService はシフト管理のビジネスロジックインターフェース

func buildShiftEntryUpdate(input *UpdateShiftEntryInput) map[string]any {
	fields := map[string]any{}
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
	return fields
}

type ShiftEntryService interface {
	List(ctx context.Context, clinicID uint64, yearMonth string, staffID *uint64) ([]model.ShiftEntry, error)
	Create(ctx context.Context, clinicID uint64, input *CreateShiftEntryInput) (*model.ShiftEntry, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateShiftEntryInput) (*model.ShiftEntry, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// GetOnDutyStaffs は指定日に出勤しているスタッフ一覧を返す (BUG-344)
	GetOnDutyStaffs(ctx context.Context, clinicID uint64, date time.Time) ([]model.Staff, error)
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
			return nil, apperrors.Wrap(err, "failed to validate year month")
		}
	}
	filter := repository.ShiftEntryFilter{
		YearMonth: yearMonth,
		StaffID:   staffID,
	}
	items, err := s.repo.FindAll(ctx, clinicID, filter)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list shift entries", "error", err)
		return nil, apperrors.Wrap(err, "failed to list shift entries")
	}
	return items, nil
}

// validateShiftType は入力文字列が有効な ShiftType 値かどうかを検証する。
func validateShiftType(s model.ShiftType) error {
	switch s {
	case model.ShiftTypeFull, model.ShiftTypeMorning, model.ShiftTypeAfternoon,
		model.ShiftTypeOff, model.ShiftTypePaidLeave:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid shift_type: %s", s))
	}
}

// requiresTimeSlot は時刻（start_time/end_time）が必要なシフト種別かどうかを返す。
// off・paid_leave は時刻不要。
func requiresTimeSlot(shiftType model.ShiftType) bool {
	return sharedkernel.RequiresTimeSlot(shiftType)
}

// validateShiftTimes は時刻が必要なシフト種別で end_time <= start_time でないかを検証する。
// startTime/endTime は "15:04:05" 形式の文字列。
func validateShiftTimes(shiftType model.ShiftType, startTime, endTime *string) error {
	return sharedkernel.ValidateShiftTimes(shiftType, startTime, endTime)
}

func (s *shiftEntryService) Create(ctx context.Context, clinicID uint64, input *CreateShiftEntryInput) (*model.ShiftEntry, error) {
	shiftType := model.ShiftType(input.ShiftType)
	if err := validateShiftType(shiftType); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate shift type")
	}
	startTime := normalizeTimeString(input.StartTime)
	endTime := normalizeTimeString(input.EndTime)

	// BUG-028: 時刻バリデーション（off/paid_leave は除外）
	if err := validateShiftTimes(shiftType, startTime, endTime); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate shift times")
	}

	entry := &model.ShiftEntry{
		ClinicID:  clinicID,
		StaffID:   input.StaffID,
		Date:      input.Date,
		ShiftType: shiftType,
		StartTime: startTime,
		EndTime:   endTime,
		Notes:     input.Notes,
	}
	if err := s.repo.Create(ctx, entry); err != nil {
		slog.ErrorContext(ctx, "failed to create shift entry", "error", err)
		return nil, apperrors.Wrap(err, "failed to create shift entry")
	}
	// 休憩時間を保存
	if len(input.Breaks) > 0 {
		breaks := make([]model.ShiftEntryBreak, 0, len(input.Breaks))
		for _, b := range input.Breaks {
			breaks = append(breaks, model.ShiftEntryBreak{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
		}
		if err := s.repo.ReplaceBreaks(ctx, entry.ID, breaks); err != nil {
			slog.ErrorContext(ctx, "failed to save shift breaks", "error", err)
			return nil, apperrors.Wrap(err, "failed to save shift breaks")
		}
	}
	slog.InfoContext(ctx, "shift entry created",
		slog.Uint64("shift_entry_id", entry.ID),
		slog.Uint64("clinic_id", clinicID))
	result, err := s.repo.FindByID(ctx, clinicID, entry.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get shift entry after create", "error", err)
		return nil, apperrors.Wrap(err, "failed to get shift entry after create")
	}
	return result, nil
}

func (s *shiftEntryService) Update(ctx context.Context, clinicID, id uint64, input *UpdateShiftEntryInput) (*model.ShiftEntry, error) {
	// BUG-028: 時刻バリデーション。更新後の start_time/end_time を確定させるために既存レコードを取得する。
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find shift entry for update", "error", err)
		return nil, apperrors.Wrap(err, "failed to find shift entry for update")
	}
	fields := buildShiftEntryUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	// 更新後の shiftType を決定（input に値があれば上書き、なければ既存値）
	effectiveShiftType := existing.ShiftType
	if input.ShiftType != nil {
		st := model.ShiftType(*input.ShiftType)
		if err := validateShiftType(st); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate shift type")
		}
		effectiveShiftType = st
	}
	// 更新後の start_time/end_time を決定
	effectiveStart := existing.StartTime
	effectiveEnd := existing.EndTime
	if input.StartTime != nil {
		effectiveStart = normalizeTimeString(input.StartTime)
	}
	if input.EndTime != nil {
		effectiveEnd = normalizeTimeString(input.EndTime)
	}
	if err := validateShiftTimes(effectiveShiftType, effectiveStart, effectiveEnd); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate shift times")
	}

	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update shift entry", "error", err)
		return nil, apperrors.Wrap(err, "failed to update shift entry")
	}
	// 休憩時間の更新（送信された場合のみ）
	if input.Breaks != nil {
		breaks := make([]model.ShiftEntryBreak, 0, len(*input.Breaks))
		for _, b := range *input.Breaks {
			breaks = append(breaks, model.ShiftEntryBreak{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
		}
		if err := s.repo.ReplaceBreaks(ctx, id, breaks); err != nil {
			slog.ErrorContext(ctx, "failed to save shift breaks", "error", err)
			return nil, apperrors.Wrap(err, "failed to save shift breaks")
		}
	}
	slog.InfoContext(ctx, "shift entry updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("shift_entry_id", id))
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get shift entry after update", "error", err)
		return nil, apperrors.Wrap(err, "failed to get shift entry after update")
	}
	return result, nil
}

func (s *shiftEntryService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find shift entry")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete shift entry", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete shift entry")
	}
	slog.InfoContext(ctx, "shift entry deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("shift_entry_id", id))
	return nil
}

// validateYearMonth は "YYYY-MM" 形式を検証する
func validateYearMonth(yearMonth string) error {
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}$`, yearMonth)
	if !matched {
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid year_month format: %s (expected YYYY-MM)", yearMonth))
	}
	_, err := time.ParseInLocation("2006-01", yearMonth, time.Local)
	if err != nil {
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid year_month value: %s", yearMonth))
	}
	return nil
}

// GetOnDutyStaffs は指定日に出勤しているスタッフ一覧を返す (BUG-344)
func (s *shiftEntryService) GetOnDutyStaffs(ctx context.Context, clinicID uint64, date time.Time) ([]model.Staff, error) {
	staffs, err := s.repo.FindOnDutyStaffs(ctx, clinicID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get on-duty staffs", "error", err)
		return nil, apperrors.Wrap(err, "failed to get on-duty staffs")
	}
	return staffs, nil
}

var _ ShiftEntryService = (*shiftEntryService)(nil)
