package staff_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// shiftTS は勤務日（JST 暦日）と時刻文字列から starts_at/ends_at 用の明示日時を作る（EMR-493 / EMR-549）。
// 日跨ぎは date を翌日にして表し、時刻の大小から翌日は推測しない。
func shiftTS(t testing.TB, date time.Time, clock string) *time.Time {
	t.Helper()
	var parsed time.Time
	var err error
	for _, layout := range []string{"15:04:05", "15:04"} {
		if parsed, err = time.Parse(layout, clock); err == nil {
			break
		}
	}
	require.NoError(t, err)
	year, month, day := date.In(config.JST).Date()
	ts := time.Date(year, month, day, parsed.Hour(), parsed.Minute(), parsed.Second(), 0, config.JST)
	return &ts
}

// shiftBreakAt は勤務日基準の休憩行を作る。
func shiftBreakAt(t testing.TB, date time.Time, start, end string) model.ShiftEntryBreak {
	t.Helper()
	return model.ShiftEntryBreak{StartsAt: shiftTS(t, date, start), EndsAt: shiftTS(t, date, end)}
}

// shiftClock は starts_at/ends_at を JST の "HH:MM:SS" へ戻す（nil は空文字）。
func shiftClock(ts *time.Time) string {
	if ts == nil {
		return ""
	}
	return ts.In(config.JST).Format("15:04:05")
}
