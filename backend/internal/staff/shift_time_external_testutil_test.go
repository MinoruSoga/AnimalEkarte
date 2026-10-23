package staff_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// shiftTS は勤務日と時刻文字列から starts_at/ends_at 用のタイムスタンプを作る（EMR-493）。
func shiftTS(t testing.TB, date time.Time, clock string) *time.Time {
	t.Helper()
	ts, err := sharedkernel.ShiftClockToTimestamp(date, &clock)
	require.NoError(t, err)
	return ts
}

// shiftBreakAt は勤務日基準の休憩行を作る。
func shiftBreakAt(t testing.TB, date time.Time, start, end string) model.ShiftEntryBreak {
	t.Helper()
	return model.ShiftEntryBreak{StartsAt: shiftTS(t, date, start), EndsAt: shiftTS(t, date, end)}
}

// shiftClock は starts_at/ends_at を wire 形式の "HH:MM:SS" へ戻す（nil は空文字）。
func shiftClock(ts *time.Time) string {
	if clock := sharedkernel.ShiftTimestampToClock(ts); clock != nil {
		return *clock
	}
	return ""
}
