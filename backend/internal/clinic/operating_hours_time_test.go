package clinic

import (
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/config"
)

// fixture は emr-clinic-line-time-selected-contract-20261022 §1/§8 の時刻境界。実行は NOT_RUN（本人指示待ち）。

func TestParseClockStartAndEnd(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		raw     string
		isEnd   bool
		want    ClockMinute
		wantErr bool
	}{
		{name: "start 00:00", raw: "00:00", want: 0},
		{name: "start 23:59", raw: "23:59", want: 1439},
		{name: "start 24:00 rejected", raw: "24:00", wantErr: true},
		{name: "end 00:01", raw: "00:01", isEnd: true, want: 1},
		{name: "end 24:00", raw: "24:00", isEnd: true, want: 1440},
		{name: "end 00:00 rejected", raw: "00:00", isEnd: true, wantErr: true},
		{name: "end 24:01 rejected", raw: "24:01", isEnd: true, wantErr: true},
		{name: "legacy HHMM rejected", raw: "0900", wantErr: true},
		{name: "single digit hour rejected", raw: "9:00", wantErr: true},
		{name: "minute 60 rejected", raw: "09:60", wantErr: true},
		{name: "full-width digit rejected", raw: "０9:00", wantErr: true},
		{name: "seconds rejected", raw: "09:00:00", wantErr: true},
		{name: "empty rejected", raw: "", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			parse := ParseClockStart
			if tc.isEnd {
				parse = ParseClockEnd
			}
			got, err := parse(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got %v", tc.raw, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestClockMinuteString(t *testing.T) {
	t.Parallel()
	if got := ClockMinute(1440).String(); got != "24:00" {
		t.Fatalf("got %q, want 24:00", got)
	}
	if got := ClockMinute(545).String(); got != "09:05" {
		t.Fatalf("got %q, want 09:05", got)
	}
}

func TestNewTimeIntervalRequiresStartBeforeEnd(t *testing.T) {
	t.Parallel()
	if _, err := NewTimeInterval("18:00", "18:00"); err == nil {
		t.Fatal("expected equal start/end to be rejected")
	}
	if _, err := NewTimeInterval("18:00", "09:00"); err == nil {
		t.Fatal("expected reversed interval to be rejected (no implicit day crossing)")
	}
	iv, err := NewTimeInterval("22:00", "24:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if iv.Start != 1320 || iv.End != 1440 {
		t.Fatalf("got %+v", iv)
	}
}

func TestTimeIntervalInstantsOnTreats2400AsNextJSTDay(t *testing.T) {
	t.Parallel()
	date, err := ParseCalendarDate("2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	start, end := TimeInterval{Start: 1320, End: 1440}.InstantsOn(date)
	wantStart := time.Date(2026, 12, 31, 22, 0, 0, 0, config.JST)
	wantEnd := time.Date(2027, 1, 1, 0, 0, 0, 0, config.JST)
	if !start.Equal(wantStart) || !end.Equal(wantEnd) {
		t.Fatalf("got %v-%v, want %v-%v", start, end, wantStart, wantEnd)
	}
}

func TestTimeIntervalContainsIsHalfOpen(t *testing.T) {
	t.Parallel()
	iv := TimeInterval{Start: 540, End: 1080}
	if !iv.Contains(540) || iv.Contains(1080) || !iv.Contains(1079) {
		t.Fatalf("half-open boundary broken: %+v", iv)
	}
}

func TestIntervalAlgebra(t *testing.T) {
	t.Parallel()
	common := []TimeInterval{{Start: 540, End: 720}, {Start: 780, End: 1080}} // 09:00-12:00, 13:00-18:00
	line := []TimeInterval{{Start: 600, End: 1140}}                           // 10:00-19:00

	assertIntervals(t, "intersect", IntersectIntervals(common, line),
		[]TimeInterval{{Start: 600, End: 720}, {Start: 780, End: 1080}})
	assertIntervals(t, "subtract", SubtractIntervals(line, common),
		[]TimeInterval{{Start: 720, End: 780}, {Start: 1080, End: 1140}})
	assertIntervals(t, "normalize merges adjacent",
		NormalizeIntervals([]TimeInterval{{Start: 780, End: 900}, {Start: 540, End: 780}}),
		[]TimeInterval{{Start: 540, End: 900}})

	if IntervalsCover(common, line) {
		t.Fatal("line exceeds common; cover must be false")
	}
	if !IntervalsCover(common, []TimeInterval{{Start: 600, End: 700}}) {
		t.Fatal("inner interval within common must be covered")
	}
	if !IntervalsCover(common, nil) {
		t.Fatal("empty inner is trivially covered")
	}
}

func TestNormalizeIntervalsDoesNotMutateInput(t *testing.T) {
	t.Parallel()
	in := []TimeInterval{{Start: 780, End: 900}, {Start: 540, End: 780}}
	_ = NormalizeIntervals(in)
	if in[0].Start != 780 {
		t.Fatalf("input mutated: %+v", in)
	}
}

func TestParseCalendarDateStrictGregorian(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"2026-02-30", "2026-2-01", "2026/02/01", "20260201", " 2026-02-01", "2026-13-01"} {
		if _, err := ParseCalendarDate(raw); err == nil {
			t.Errorf("expected %q to be rejected", raw)
		}
	}
	d, err := ParseCalendarDate("2028-02-29")
	if err != nil {
		t.Fatalf("leap day rejected: %v", err)
	}
	if d.String() != "2028-02-29" || d.Weekday() != time.Tuesday {
		t.Fatalf("got %s %v", d, d.Weekday())
	}
}

func TestCalendarDateOfUsesJST(t *testing.T) {
	t.Parallel()
	utc := time.Date(2026, 10, 21, 15, 30, 0, 0, time.UTC) // 2026-10-22 00:30 JST
	if got := CalendarDateOf(utc).String(); got != "2026-10-22" {
		t.Fatalf("got %s, want 2026-10-22", got)
	}
}

func assertIntervals(t *testing.T, label string, got, want []TimeInterval) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %+v, want %+v", label, got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s: got %+v, want %+v", label, got, want)
		}
	}
}
