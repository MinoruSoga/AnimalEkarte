package staff

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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

// ShiftEntryStaffAssignmentLocker はシフト作成時の有効な医院所属確認に必要な最小依存である。
// 実装は soft-delete 済み割当を除外し、同一トランザクションの完了まで有効な所属行を
// FOR SHARE で保持する。これにより所属確認直後の並行解除との TOCTOU を防ぐ。
type ShiftEntryStaffAssignmentLocker interface {
	LockActiveByStaffAndClinic(
		ctx context.Context,
		staffID, clinicID uint64,
	) (*model.StaffClinicAssignment, error)
}

// ShiftEntryStaffLocker はシフト作成時に有効な staff identity を assignment より
// 先に FOR SHARE で保持するための consumer-side port である。
type ShiftEntryStaffLocker interface {
	LockActiveByIDForShare(ctx context.Context, staffID uint64) (*model.Staff, error)
}

type shiftEntryService struct {
	repo                ShiftEntryRepository
	staffRepo           ShiftEntryStaffLocker
	staffAssignmentRepo ShiftEntryStaffAssignmentLocker
	tx                  Transactor
}

// NewShiftEntryService はShiftEntryServiceを初期化して返す
func NewShiftEntryService(
	repo ShiftEntryRepository,
	staffRepo ShiftEntryStaffLocker,
	staffAssignmentRepo ShiftEntryStaffAssignmentLocker,
	tx Transactor,
) ShiftEntryService {
	return &shiftEntryService{
		repo:                repo,
		staffRepo:           staffRepo,
		staffAssignmentRepo: staffAssignmentRepo,
		tx:                  tx,
	}
}

func (s *shiftEntryService) List(ctx context.Context, clinicID uint64, yearMonth string, staffID *uint64) ([]model.ShiftEntry, error) {
	// G2F-10: empty year_month defaults to the current local month so list stays bounded.
	if yearMonth == "" {
		yearMonth = time.Now().In(time.Local).Format("2006-01")
	}
	if err := validateYearMonth(yearMonth); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate year month")
	}
	filter := ShiftEntryFilter{
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
	if input == nil {
		return nil, apperrors.WrapInvalidInput("shift entry input is required")
	}
	if clinicID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_id is required")
	}
	if input.StaffID == 0 {
		return nil, apperrors.WrapInvalidInput("staff_id is required")
	}
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

	breaks := make([]model.ShiftEntryBreak, 0, len(input.Breaks))
	for _, b := range input.Breaks {
		breaks = append(breaks, model.ShiftEntryBreak{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
	}
	if s.repo == nil || s.staffRepo == nil || s.staffAssignmentRepo == nil || s.tx == nil {
		return nil, apperrors.WrapInternalServerError("shift entry dependencies are not configured")
	}

	var result *model.ShiftEntry
	if err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		staff, lockStaffErr := s.staffRepo.LockActiveByIDForShare(ctx, input.StaffID)
		if lockStaffErr != nil {
			return apperrors.Wrap(lockStaffErr, "failed to lock active staff")
		}
		if staff == nil || staff.ID != input.StaffID {
			return apperrors.WrapInternalServerError("staff lock returned an invalid record")
		}

		assignment, lockErr := s.staffAssignmentRepo.LockActiveByStaffAndClinic(ctx, input.StaffID, clinicID)
		if lockErr != nil {
			return apperrors.Wrap(lockErr, "failed to lock staff clinic assignment")
		}
		if assignment == nil || assignment.StaffID != input.StaffID || assignment.ClinicID != clinicID {
			return apperrors.WrapInternalServerError("staff clinic assignment lock returned an invalid record")
		}

		if createErr := s.repo.Create(ctx, entry); createErr != nil {
			slog.ErrorContext(ctx, "failed to create shift entry", "error", createErr)
			return apperrors.Wrap(createErr, "failed to create shift entry")
		}
		if len(breaks) > 0 {
			if replaceErr := s.repo.ReplaceBreaks(ctx, entry.ID, breaks); replaceErr != nil {
				slog.ErrorContext(ctx, "failed to save shift breaks", "error", replaceErr)
				return apperrors.Wrap(replaceErr, "failed to save shift breaks")
			}
		}

		created, findErr := s.repo.FindByID(ctx, clinicID, entry.ID)
		if findErr != nil {
			slog.ErrorContext(ctx, "failed to get shift entry after create", "error", findErr)
			return apperrors.Wrap(findErr, "failed to get shift entry after create")
		}
		result = created
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to create shift entry")
	}

	slog.InfoContext(ctx, "shift entry created",
		slog.Uint64("shift_entry_id", entry.ID),
		slog.Uint64("clinic_id", clinicID))
	return result, nil
}

func (s *shiftEntryService) Update(ctx context.Context, clinicID, id uint64, input *UpdateShiftEntryInput) (*model.ShiftEntry, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if s.repo == nil || s.tx == nil {
		return nil, apperrors.WrapInternalServerError("shift entry update dependencies are not configured")
	}
	fields := buildShiftEntryUpdate(input)
	if len(fields) == 0 && input.Breaks == nil {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	var result *model.ShiftEntry
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		existing, err := s.repo.LockActiveByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock shift entry for update")
		}

		effectiveShiftType := existing.ShiftType
		if input.ShiftType != nil {
			shiftType := model.ShiftType(*input.ShiftType)
			if err := validateShiftType(shiftType); err != nil {
				return apperrors.Wrap(err, "failed to validate shift type")
			}
			effectiveShiftType = shiftType
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
			if err := s.repo.Update(txCtx, clinicID, id, fields); err != nil {
				return apperrors.Wrap(err, "failed to update shift entry")
			}
		}
		if input.Breaks != nil {
			breaks := make([]model.ShiftEntryBreak, 0, len(*input.Breaks))
			for _, b := range *input.Breaks {
				breaks = append(
					breaks,
					model.ShiftEntryBreak{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd},
				)
			}
			if err := s.repo.ReplaceBreaks(txCtx, id, breaks); err != nil {
				return apperrors.Wrap(err, "failed to save shift breaks")
			}
		}
		updated, err := s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get shift entry after update")
		}
		result = updated
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to update shift entry", "error", err)
		return nil, apperrors.Wrap(err, "failed to update shift entry")
	}

	slog.InfoContext(ctx, "shift entry updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("shift_entry_id", id))
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
