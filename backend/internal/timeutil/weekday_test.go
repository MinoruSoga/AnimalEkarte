package timeutil

import (
	"testing"
	"time"
)

func TestWeekdayJP(t *testing.T) {
	t.Parallel()

	sunday := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC) // 2026-07-12 is Sunday
	if got := WeekdayJP(sunday); got != "日" {
		t.Errorf("WeekdayJP(sunday) = %q, want %q", got, "日")
	}

	saturday := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC) // 2026-07-18 is Saturday
	if got := WeekdayJP(saturday); got != "土" {
		t.Errorf("WeekdayJP(saturday) = %q, want %q", got, "土")
	}
}
