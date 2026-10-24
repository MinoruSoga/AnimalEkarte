package reservation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// rsvShiftDate は構造体だけで完結するテストで starts_at/ends_at の勤務日として使う固定日。
var rsvShiftDate = time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

// rsvTS は勤務日（JST 暦日）と時刻文字列から starts_at/ends_at 用の明示日時を作る（EMR-493 / EMR-549）。
// 日跨ぎは date を翌日にして表し、時刻の大小から翌日は推測しない。
func rsvTS(t testing.TB, date time.Time, clock string) *time.Time {
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

// rsvBreak は勤務日基準の休憩行を作る。
func rsvBreak(t testing.TB, date time.Time, start, end string) model.ShiftEntryBreak {
	t.Helper()
	return model.ShiftEntryBreak{StartsAt: rsvTS(t, date, start), EndsAt: rsvTS(t, date, end)}
}

// rsvClock は starts_at/ends_at を JST の "HH:MM:SS" へ戻す（nil は空文字）。
func rsvClock(ts *time.Time) string {
	if ts == nil {
		return ""
	}
	return ts.In(config.JST).Format("15:04:05")
}
