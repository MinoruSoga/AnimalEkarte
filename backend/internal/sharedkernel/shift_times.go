// shift_times.go — シフト時刻の正規化・検証（BE9-2C R②: service/shift_entry_service.go から昇格。
// staff（shift_entry/shift_template）と reservation（reservation_schedule）の恒久ドメイン跨ぎ）。
// ParseHHMM は closing/clinic と billing/cash_register の締め時刻パース正本（POC-16 / X-09）。
package sharedkernel

import (
	"fmt"
	"sort"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// ParseHHMM parses "HH:MM" or PostgreSQL time "HH:MM:SS" into hour and minute (POC-16).
// Seconds are discarded; invalid formats return invalid input.
func ParseHHMM(s string) (h, m int, err error) {
	// PostgreSQL time 型は "HH:MM:SS" で返るので秒部分を除去する
	if len(s) == 8 && s[2] == ':' && s[5] == ':' {
		s = s[:5]
	}
	if len(s) != 5 || s[2] != ':' {
		return 0, 0, apperrors.WrapInvalidInput("時刻は HH:MM 形式で指定してください")
	}
	var hh, mm int
	_, parseErr := fmt.Sscanf(s, "%d:%d", &hh, &mm)
	if parseErr != nil {
		return 0, 0, apperrors.WrapInvalidInput("時刻の解析に失敗しました")
	}
	return hh, mm, nil
}

// NormalizeTimeString は HH:MM / HH:MM:SS を HH:MM:SS へ正規化する（不正・空は nil）。
func NormalizeTimeString(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	for _, layout := range []string{"15:04:05", "15:04"} {
		if t, err := time.ParseInLocation(layout, *s, time.Local); err == nil {
			normalized := t.Format("15:04:05")
			return &normalized
		}
	}
	return nil
}

// RequiresTimeSlot は勤務時間帯入力を要する ShiftType か判定する。
func RequiresTimeSlot(shiftType model.ShiftType) bool {
	switch shiftType {
	case model.ShiftTypeOff, model.ShiftTypePaidLeave:
		return false
	default:
		return true
	}
}

// ValidateShiftTimes はシフト種別に応じた開始/終了時刻の整合を検証する（BUG-028 / BUG-036 / EMR-241）。
// start_time / end_time は両方指定するか両方省略するかのどちらかのみ許容する。
// 片方のみの指定は off・paid_leave を含む全 ShiftType で拒否する。
// off・paid_leave 以外では start_time と end_time の両方が必須。
func ValidateShiftTimes(shiftType model.ShiftType, startTime, endTime *string) error {
	if (startTime == nil) != (endTime == nil) {
		return apperrors.Wrap(apperrors.ErrInvalidInput, "start_time and end_time must both be provided or both be omitted")
	}
	if !RequiresTimeSlot(shiftType) {
		return nil
	}
	if startTime == nil || endTime == nil {
		return apperrors.Wrap(apperrors.ErrInvalidInput, "start_time and end_time are required for this shift type")
	}
	const layout = "15:04:05"
	st, err := time.ParseInLocation(layout, *startTime, time.Local)
	if err != nil {
		return apperrors.Wrap(apperrors.ErrInvalidInput, "invalid start_time format: expected HH:MM:SS")
	}
	et, err := time.ParseInLocation(layout, *endTime, time.Local)
	if err != nil {
		return apperrors.Wrap(apperrors.ErrInvalidInput, "invalid end_time format: expected HH:MM:SS")
	}
	if !et.After(st) {
		return apperrors.Wrap(apperrors.ErrInvalidInput, "end_time must be after start_time")
	}
	return nil
}

// ShiftIntervalState の値。api.yaml の ShiftEntry.interval_state / ShiftWorkBreakView.interval_state と一致させる。
const (
	ShiftIntervalStateKnown   = "known"
	ShiftIntervalStateUnknown = "unknown"
)

// ShiftWindow は勤務内の休憩検証に使う明示日時の区間 [StartsAt, EndsAt)。
type ShiftWindow struct {
	StartsAt time.Time
	EndsAt   time.Time
}

// ValidateShiftInterval は勤務種別に応じた starts_at / ends_at の整合を検証する（EMR-493 / EMR-549）。
// 勤務（full / morning / afternoon）は両日時が必須で starts_at < ends_at。ends_at が翌日以降の日跨ぎ勤務も許し、
// 時刻だけから翌日を推測しない。off / paid_leave は終日非勤務のため両日時とも NULL。
func ValidateShiftInterval(shiftType model.ShiftType, startsAt, endsAt *time.Time) error {
	if !RequiresTimeSlot(shiftType) {
		if startsAt != nil || endsAt != nil {
			return apperrors.WrapInvalidInput("starts_at and ends_at must be null for off and paid_leave")
		}
		return nil
	}
	if startsAt == nil || endsAt == nil {
		return apperrors.WrapInvalidInput("starts_at and ends_at are required for this shift type")
	}
	if !startsAt.Before(*endsAt) {
		return apperrors.WrapInvalidInput("starts_at must be before ends_at")
	}
	return nil
}

// ValidateShiftDate は勤務日 date が starts_at の JST 暦日と一致することを検証する。
// starts_at が nil（終日非勤務）のときは検査しない。
func ValidateShiftDate(date time.Time, startsAt *time.Time) error {
	if startsAt == nil {
		return nil
	}
	if shiftDateKey(date) != shiftDateKey(*startsAt) {
		return apperrors.WrapInvalidInput("date must equal the JST date of starts_at")
	}
	return nil
}

// shiftDateKey は時刻の JST 暦日を YYYY-MM-DD で返す。DATE 列は UTC 0 時で読まれうるため、必ず JST へ直して比較する。
func shiftDateKey(t time.Time) string {
	return t.In(config.JST).Format(time.DateOnly)
}

// ValidateShiftBreaks は休憩が勤務区間 [starts_at, ends_at) の内側にあり、各休憩の starts_at < ends_at で、
// 互いに正の時間幅で重複しないことを検証する。区間は半開のため、休憩の終了と次の休憩の開始がちょうど接する場合は
// 重複とみなさない。
func ValidateShiftBreaks(startsAt, endsAt *time.Time, breaks []ShiftWindow) error {
	if len(breaks) == 0 {
		return nil
	}
	if startsAt == nil || endsAt == nil {
		return apperrors.WrapInvalidInput("breaks require a shift interval")
	}
	sorted := make([]ShiftWindow, len(breaks))
	copy(sorted, breaks)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].StartsAt.Before(sorted[j].StartsAt) })
	for i, b := range sorted {
		if !b.StartsAt.Before(b.EndsAt) {
			return apperrors.WrapInvalidInput("break starts_at must be before ends_at")
		}
		if b.StartsAt.Before(*startsAt) || b.EndsAt.After(*endsAt) {
			return apperrors.WrapInvalidInput("break must be within the shift interval")
		}
		if i > 0 && sorted[i-1].EndsAt.After(b.StartsAt) {
			return apperrors.WrapInvalidInput("breaks must not overlap")
		}
	}
	return nil
}

// ValidateShiftEntryShape は保存後の勤務全体（種別・日時・勤務日・休憩）を shift 契約どおり検証する。
// off / paid_leave は休憩を持てない（終日非勤務）。
func ValidateShiftEntryShape(shiftType model.ShiftType, date time.Time, startsAt, endsAt *time.Time, breaks []ShiftWindow) error {
	if err := ValidateShiftInterval(shiftType, startsAt, endsAt); err != nil {
		return err
	}
	if !RequiresTimeSlot(shiftType) {
		if len(breaks) > 0 {
			return apperrors.WrapInvalidInput("breaks must be empty for off and paid_leave")
		}
		return nil
	}
	if err := ValidateShiftDate(date, startsAt); err != nil {
		return err
	}
	return ValidateShiftBreaks(startsAt, endsAt, breaks)
}

// ShiftBreakIntervalState は休憩の interval_state を返す。いずれかの日時が NULL なら unknown。
func ShiftBreakIntervalState(b model.ShiftEntryBreak) string {
	if b.StartsAt == nil || b.EndsAt == nil {
		return ShiftIntervalStateUnknown
	}
	return ShiftIntervalStateKnown
}

// ShiftEntryIntervalState は勤務行の interval_state を返す（api.yaml ShiftEntry）。
// 勤務で starts_at / ends_at が NULL、いずれかの休憩が unknown、または off / paid_leave に矛盾する
// 日時・休憩が残る行は unknown。unknown は休日でも勤務可能でもなく、値を補完・削除しない。
func ShiftEntryIntervalState(entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) string {
	if entry == nil {
		return ShiftIntervalStateUnknown
	}
	for _, b := range breaks {
		if ShiftBreakIntervalState(b) == ShiftIntervalStateUnknown {
			return ShiftIntervalStateUnknown
		}
	}
	if RequiresTimeSlot(entry.ShiftType) {
		if entry.StartsAt == nil || entry.EndsAt == nil {
			return ShiftIntervalStateUnknown
		}
		return ShiftIntervalStateKnown
	}
	if entry.StartsAt != nil || entry.EndsAt != nil || len(breaks) > 0 {
		return ShiftIntervalStateUnknown
	}
	return ShiftIntervalStateKnown
}
