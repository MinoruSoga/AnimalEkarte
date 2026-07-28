// shift_times.go — シフト時刻の正規化・検証（BE9-2C R②: service/shift_entry_service.go から昇格。
// staff（shift_entry/shift_template）と reservation（reservation_schedule）の恒久ドメイン跨ぎ）。
package sharedkernel

import (
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

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

// ValidateShiftTimes はシフト種別に応じた開始/終了時刻の整合を検証する（BUG-028）。
func ValidateShiftTimes(shiftType model.ShiftType, startTime, endTime *string) error {
	if !RequiresTimeSlot(shiftType) {
		return nil
	}
	if startTime == nil || endTime == nil {
		return nil
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
