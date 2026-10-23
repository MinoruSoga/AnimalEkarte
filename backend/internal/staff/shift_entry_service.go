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
	// WorkDate は更新対象行の勤務日。service がロック済み行から設定する（呼び出し元は設定しない）。
	// StartTime/EndTime を timestamptz（starts_at/ends_at）へ変換するために必要。
	WorkDate time.Time
}

// ShiftEntryService はシフト管理のビジネスロジックインターフェース

// shiftEntryUpdateIsEmpty は更新対象のスカラー項目が1つも指定されていないかを返す。
func shiftEntryUpdateIsEmpty(input *UpdateShiftEntryInput) bool {
	return input.ShiftType == nil && input.StartTime == nil && input.EndTime == nil && input.Notes == nil
}

// buildShiftEntryUpdate は更新項目を GORM の fields マップへ変換する。
// HTTP/DTO の start_time/end_time（時刻文字列）は勤務日 input.WorkDate と組み合わせて
// starts_at/ends_at（timestamptz, time.Local）へ変換する。時刻指定があるのに WorkDate が
// 未設定の場合は fail-closed でエラーにする。
func buildShiftEntryUpdate(input *UpdateShiftEntryInput) (map[string]any, error) {
	fields := map[string]any{}
	if input.ShiftType != nil {
		fields["shift_type"] = *input.ShiftType
	}
	if input.StartTime != nil || input.EndTime != nil {
		if input.WorkDate.IsZero() {
			return nil, apperrors.WrapInternalServerError("shift entry work date is required to update times")
		}
	}
	if input.StartTime != nil {
		startsAt, err := sharedkernel.ShiftClockToTimestamp(input.WorkDate, normalizeTimeString(input.StartTime))
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to convert start_time")
		}
		fields["starts_at"] = startsAt
	}
	if input.EndTime != nil {
		endsAt, err := sharedkernel.ShiftClockToTimestamp(input.WorkDate, normalizeTimeString(input.EndTime))
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to convert end_time")
		}
		fields["ends_at"] = endsAt
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	return fields, nil
}

// buildShiftEntryBreaks は休憩入力（HH:MM / HH:MM:SS）を勤務日基準の starts_at/ends_at へ変換する。
func buildShiftEntryBreaks(workDate time.Time, inputs []ShiftBreakInput) ([]model.ShiftEntryBreak, error) {
	breaks := make([]model.ShiftEntryBreak, 0, len(inputs))
	for _, b := range inputs {
		startsAt, err := sharedkernel.ShiftClockToTimestamp(workDate, &b.BreakStart)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to convert break_start")
		}
		endsAt, err := sharedkernel.ShiftClockToTimestamp(workDate, &b.BreakEnd)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to convert break_end")
		}
		if startsAt == nil || endsAt == nil {
			return nil, apperrors.WrapInvalidInput("break_start and break_end are required")
		}
		breaks = append(breaks, model.ShiftEntryBreak{StartsAt: startsAt, EndsAt: endsAt})
	}
	return breaks, nil
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

	startsAt, err := sharedkernel.ShiftClockToTimestamp(input.Date, startTime)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to convert start_time")
	}
	endsAt, err := sharedkernel.ShiftClockToTimestamp(input.Date, endTime)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to convert end_time")
	}

	entry := &model.ShiftEntry{
		ClinicID:  clinicID,
		StaffID:   input.StaffID,
		Date:      input.Date,
		ShiftType: shiftType,
		StartsAt:  startsAt,
		EndsAt:    endsAt,
		Notes:     input.Notes,
	}

	breaks, err := buildShiftEntryBreaks(input.Date, input.Breaks)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to validate shift breaks")
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
			return apperrors.Wrap(createErr, "failed to create shift entry")
		}
		if len(breaks) > 0 {
			if replaceErr := s.repo.ReplaceBreaks(ctx, entry.ID, breaks); replaceErr != nil {
				return apperrors.Wrap(replaceErr, "failed to save shift breaks")
			}
		}

		created, findErr := s.repo.FindByID(ctx, clinicID, entry.ID)
		if findErr != nil {
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
	if shiftEntryUpdateIsEmpty(input) && input.Breaks == nil {
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
		effectiveStart := sharedkernel.ShiftTimestampToClock(existing.StartsAt)
		effectiveEnd := sharedkernel.ShiftTimestampToClock(existing.EndsAt)
		if input.StartTime != nil {
			effectiveStart = normalizeTimeString(input.StartTime)
		}
		if input.EndTime != nil {
			effectiveEnd = normalizeTimeString(input.EndTime)
		}
		if err := validateShiftTimes(effectiveShiftType, effectiveStart, effectiveEnd); err != nil {
			return apperrors.Wrap(err, "failed to validate shift times")
		}

		if !shiftEntryUpdateIsEmpty(input) {
			cmd := *input
			cmd.WorkDate = existing.Date
			if err := s.repo.Update(txCtx, clinicID, id, cmd); err != nil {
				return apperrors.Wrap(err, "failed to update shift entry")
			}
		}
		if input.Breaks != nil {
			breaks, err := buildShiftEntryBreaks(existing.Date, *input.Breaks)
			if err != nil {
				return apperrors.Wrap(err, "failed to validate shift breaks")
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
		return nil, apperrors.Wrap(err, "failed to get on-duty staffs")
	}
	return staffs, nil
}

var _ ShiftEntryService = (*shiftEntryService)(nil)
