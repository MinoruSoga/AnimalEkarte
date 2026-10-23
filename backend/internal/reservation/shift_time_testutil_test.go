package reservation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// rsvShiftDate は構造体だけで完結するテストで starts_at/ends_at の勤務日として使う固定日。
var rsvShiftDate = time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

// rsvTS は勤務日と時刻文字列から starts_at/ends_at 用のタイムスタンプを作る（EMR-493）。
func rsvTS(t testing.TB, date time.Time, clock string) *time.Time {
	t.Helper()
	ts, err := sharedkernel.ShiftClockToTimestamp(date, &clock)
	require.NoError(t, err)
	return ts
}

// rsvBreak は勤務日基準の休憩行を作る。
func rsvBreak(t testing.TB, date time.Time, start, end string) model.ShiftEntryBreak {
	t.Helper()
	return model.ShiftEntryBreak{StartsAt: rsvTS(t, date, start), EndsAt: rsvTS(t, date, end)}
}

// rsvClock は starts_at/ends_at を wire 形式の "HH:MM:SS" へ戻す（nil は空文字）。
func rsvClock(ts *time.Time) string {
	if clock := sharedkernel.ShiftTimestampToClock(ts); clock != nil {
		return *clock
	}
	return ""
}
