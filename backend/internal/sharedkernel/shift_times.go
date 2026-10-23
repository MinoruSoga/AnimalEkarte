// shift_times.go — シフト時刻の正規化・検証（BE9-2C R②: service/shift_entry_service.go から昇格。
// staff（shift_entry/shift_template）と reservation（reservation_schedule）の恒久ドメイン跨ぎ）。
// ParseHHMM は closing/clinic と billing/cash_register の締め時刻パース正本（POC-16 / X-09）。
package sharedkernel

import (
	"fmt"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
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

// ShiftClockToTimestamp は勤務日と時刻文字列（HH:MM / HH:MM:SS）から time.Local の
// タイムスタンプを組み立てる（shift_entries.starts_at/ends_at, shift_entry_breaks.starts_at/ends_at 用）。
// 年月日は date.In(time.Local) から取る（レスポンスの Date 整形と同じゾーン）。
// clock が nil または空文字なら (nil, nil)。解析できない clock は invalid input。
func ShiftClockToTimestamp(date time.Time, clock *string) (*time.Time, error) {
	if clock == nil || *clock == "" {
		return nil, nil
	}
	normalized := NormalizeTimeString(clock)
	if normalized == nil {
		return nil, apperrors.WrapInvalidInput("時刻は HH:MM 形式で指定してください")
	}
	parsed, err := time.ParseInLocation("15:04:05", *normalized, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("時刻の解析に失敗しました")
	}
	year, month, day := date.In(time.Local).Date()
	ts := time.Date(year, month, day, parsed.Hour(), parsed.Minute(), parsed.Second(), 0, time.Local)
	return &ts, nil
}

// ShiftTimestampToClock は starts_at/ends_at を time.Local の "HH:MM:SS" へ戻す。nil は nil。
func ShiftTimestampToClock(ts *time.Time) *string {
	if ts == nil {
		return nil
	}
	clock := ts.In(time.Local).Format("15:04:05")
	return &clock
}
