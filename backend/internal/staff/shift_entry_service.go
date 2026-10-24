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

// ShiftBreakInput は休憩の入力DTO（明示日時。EMR-549: 時刻文字列の wire は受け付けない）
type ShiftBreakInput struct {
	StartsAt time.Time
	EndsAt   time.Time
}

// CreateShiftEntryInput はシフト作成の入力DTO
type CreateShiftEntryInput struct {
	StaffID   uint64
	Date      time.Time
	ShiftType string
	StartsAt  *time.Time
	EndsAt    *time.Time
	Notes     string
	Breaks    []ShiftBreakInput
}

// OptionalShiftTime は PATCH の starts_at / ends_at を表す。
// Set=false は未指定（保存値を維持）、Set=true かつ Value=nil は明示 null（終日非勤務化）。
type OptionalShiftTime struct {
	Set   bool
	Value *time.Time
}

// UpdateShiftEntryInput はシフト更新の入力DTO（nil / Set=false = 未変更）
type UpdateShiftEntryInput struct {
	ShiftType *string
	StartsAt  OptionalShiftTime
	EndsAt    OptionalShiftTime
	Notes     *string
	Breaks    *[]ShiftBreakInput
	// ExpectedVersion は勤務 aggregate の期待版。nil は版照合なし（HTTP 境界では必須にするため、内部呼出し専用）。
	ExpectedVersion *int64
}

// ShiftEntryService はシフト管理のビジネスロジックインターフェース

// shiftEntryUpdateIsEmpty は更新対象のスカラー項目が1つも指定されていないかを返す。
func shiftEntryUpdateIsEmpty(input *UpdateShiftEntryInput) bool {
	return input.ShiftType == nil && !input.StartsAt.Set && !input.EndsAt.Set && input.Notes == nil
}

// buildShiftEntryUpdate は更新項目を GORM の fields マップへ変換する。
// starts_at / ends_at は明示日時のまま保存し、明示 null は NULL として書く。
// 時刻文字列や勤務日からの組立ては行わない（EMR-549）。
func buildShiftEntryUpdate(input *UpdateShiftEntryInput) map[string]any {
	fields := map[string]any{}
	if input.ShiftType != nil {
		fields["shift_type"] = *input.ShiftType
	}
	if input.StartsAt.Set {
		fields["starts_at"] = input.StartsAt.Value
	}
	if input.EndsAt.Set {
		fields["ends_at"] = input.EndsAt.Value
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	return fields
}

// buildShiftEntryBreaks は休憩入力（明示日時）を保存行へ写す。
func buildShiftEntryBreaks(inputs []ShiftBreakInput) []model.ShiftEntryBreak {
	breaks := make([]model.ShiftEntryBreak, 0, len(inputs))
	for _, b := range inputs {
		startsAt, endsAt := b.StartsAt, b.EndsAt
		breaks = append(breaks, model.ShiftEntryBreak{StartsAt: &startsAt, EndsAt: &endsAt})
	}
	return breaks
}

// shiftBreakWindows は休憩入力を検証用の区間へ写す。
func shiftBreakWindows(inputs []ShiftBreakInput) []sharedkernel.ShiftWindow {
	windows := make([]sharedkernel.ShiftWindow, 0, len(inputs))
	for _, b := range inputs {
		windows = append(windows, sharedkernel.ShiftWindow{StartsAt: b.StartsAt, EndsAt: b.EndsAt})
	}
	return windows
}

// existingShiftBreakWindows は保存済み休憩を検証用の区間へ写す。
// 呼出し側で interval_state=known（全休憩の日時が非 NULL）を確認済みであること。
func existingShiftBreakWindows(breaks []model.ShiftEntryBreak) []sharedkernel.ShiftWindow {
	windows := make([]sharedkernel.ShiftWindow, 0, len(breaks))
	for _, b := range breaks {
		if b.StartsAt == nil || b.EndsAt == nil {
			continue
		}
		windows = append(windows, sharedkernel.ShiftWindow{StartsAt: *b.StartsAt, EndsAt: *b.EndsAt})
	}
	return windows
}

// api.yaml ShiftContractError.code の固定値（409）。
const (
	codeShiftVersionConflict  = "version_conflict"
	codeShiftSourceUnresolved = "source_unresolved"
	codeShiftExists           = "shift_exists"
)

// newShiftContractConflict は固定 code を持つ 409 を返す。code は error 文字列ではなく分岐に使われる。
func newShiftContractConflict(code, message string) error {
	return &apperrors.AppError{Code: code, Message: message, Err: apperrors.ErrConflict}
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
	// 勤務種別・明示日時・勤務日（starts_at の JST 暦日）・休憩を契約どおり検証する（EMR-549）。
	// 日跨ぎ勤務（ends_at が翌日以降）は許可し、時刻から翌日を推測しない。
	if err := sharedkernel.ValidateShiftEntryShape(
		shiftType, input.Date, input.StartsAt, input.EndsAt, shiftBreakWindows(input.Breaks),
	); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate shift times")
	}

	entry := &model.ShiftEntry{
		ClinicID:  clinicID,
		StaffID:   input.StaffID,
		Date:      input.Date,
		ShiftType: shiftType,
		StartsAt:  input.StartsAt,
		EndsAt:    input.EndsAt,
		Notes:     input.Notes,
	}

	breaks := buildShiftEntryBreaks(input.Breaks)
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

		// 版照合（write 直前・行ロック下）。不一致は 409 version_conflict で、何も保存しない。
		if input.ExpectedVersion != nil && existing.Version != *input.ExpectedVersion {
			return newShiftContractConflict(codeShiftVersionConflict, "shift entry version does not match")
		}
		// 保存済みの勤務日時または休憩が unknown の行は、訂正確認まで保存を停止する。
		// NULL を休日・勤務可能と扱わず、値を黙って補完・削除もしない（api.yaml ShiftEntry）。
		if sharedkernel.ShiftEntryIntervalState(existing, existing.Breaks) == sharedkernel.ShiftIntervalStateUnknown {
			return newShiftContractConflict(codeShiftSourceUnresolved, "shift entry interval is unresolved")
		}

		effectiveShiftType := existing.ShiftType
		if input.ShiftType != nil {
			shiftType := model.ShiftType(*input.ShiftType)
			if err := validateShiftType(shiftType); err != nil {
				return apperrors.Wrap(err, "failed to validate shift type")
			}
			effectiveShiftType = shiftType
		}
		effectiveStart, effectiveEnd := existing.StartsAt, existing.EndsAt
		if input.StartsAt.Set {
			effectiveStart = input.StartsAt.Value
		}
		if input.EndsAt.Set {
			effectiveEnd = input.EndsAt.Value
		}
		effectiveBreaks := existingShiftBreakWindows(existing.Breaks)
		if input.Breaks != nil {
			effectiveBreaks = shiftBreakWindows(*input.Breaks)
		}
		if err := sharedkernel.ValidateShiftEntryShape(
			effectiveShiftType, existing.Date, effectiveStart, effectiveEnd, effectiveBreaks,
		); err != nil {
			return apperrors.Wrap(err, "failed to validate shift times")
		}

		if !shiftEntryUpdateIsEmpty(input) {
			if err := s.repo.Update(txCtx, clinicID, id, *input); err != nil {
				return apperrors.Wrap(err, "failed to update shift entry")
			}
		}
		if input.Breaks != nil {
			if err := s.repo.ReplaceBreaks(txCtx, id, buildShiftEntryBreaks(*input.Breaks)); err != nil {
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
