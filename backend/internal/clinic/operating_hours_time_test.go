package clinic

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/config"
)

// fixture は emr-clinic-line-time-selected-contract-20261022 §1/§8 の時刻境界。

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

func TestParseClockRejectsMalformedFiveByteInputs(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"ab:cd", "09;00", "+9:00", "-1:00", " 9:00", "25:00", "24:01", "09:0a", "9a:00"} {
		if got, err := ParseClockStart(raw); err == nil {
			t.Errorf("ParseClockStart(%q) = %d, want error", raw, got)
		}
		if got, err := ParseClockEnd(raw); err == nil {
			t.Errorf("ParseClockEnd(%q) = %d, want error", raw, got)
		}
	}
}

func TestParseClockErrorDoesNotEchoLongInput(t *testing.T) {
	t.Parallel()
	raw := strings.Repeat("x", 4096)
	_, err := ParseClockStart(raw)
	if err == nil {
		t.Fatal("expected an error")
	}
	if len(err.Error()) > 512 {
		t.Fatalf("error message copied the client input (%d bytes)", len(err.Error()))
	}
	_, err = ParseCalendarDate(raw)
	if err == nil || len(err.Error()) > 512 {
		t.Fatalf("date error message copied the client input: %v", err)
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

func TestIntervalAlgebraEdgeCases(t *testing.T) {
	t.Parallel()
	iv := func(s, e ClockMinute) TimeInterval { return TimeInterval{Start: s, End: e} }

	assertIntervals(t, "normalize overlap+duplicate+unsorted",
		NormalizeIntervals([]TimeInterval{iv(900, 950), iv(600, 700), iv(540, 650), iv(540, 650), iv(900, 950)}),
		[]TimeInterval{iv(540, 700), iv(900, 950)})
	assertIntervals(t, "normalize drops empty and inverted", NormalizeIntervals([]TimeInterval{iv(100, 100), iv(200, 100)}), []TimeInterval{})
	assertIntervals(t, "normalize nil is empty non-nil", NormalizeIntervals(nil), []TimeInterval{})

	assertIntervals(t, "intersect unsorted overlapping inputs",
		IntersectIntervals([]TimeInterval{iv(780, 1080), iv(540, 720), iv(600, 800)}, []TimeInterval{iv(1000, 1200), iv(500, 560)}),
		[]TimeInterval{iv(540, 560), iv(1000, 1080)})
	assertIntervals(t, "intersect disjoint is empty non-nil",
		IntersectIntervals([]TimeInterval{iv(0, 10)}, []TimeInterval{iv(20, 30)}), []TimeInterval{})
	assertIntervals(t, "intersect touching boundary is empty",
		IntersectIntervals([]TimeInterval{iv(0, 10)}, []TimeInterval{iv(10, 30)}), []TimeInterval{})

	assertIntervals(t, "subtract multiple exclusions",
		SubtractIntervals([]TimeInterval{iv(540, 1080)}, []TimeInterval{iv(720, 780), iv(1000, 1080), iv(600, 660)}),
		[]TimeInterval{iv(540, 600), iv(660, 720), iv(780, 1000)})
	assertIntervals(t, "subtract exclusion spanning several base intervals",
		SubtractIntervals([]TimeInterval{iv(540, 600), iv(660, 720), iv(780, 840)}, []TimeInterval{iv(570, 800)}),
		[]TimeInterval{iv(540, 570), iv(800, 840)})
	assertIntervals(t, "subtract everything is empty non-nil",
		SubtractIntervals([]TimeInterval{iv(540, 600)}, []TimeInterval{iv(500, 700)}), []TimeInterval{})
	assertIntervals(t, "subtract nothing keeps base",
		SubtractIntervals([]TimeInterval{iv(540, 600)}, nil), []TimeInterval{iv(540, 600)})

	if !IntervalsCover([]TimeInterval{iv(540, 700), iv(600, 800)}, []TimeInterval{iv(560, 790)}) {
		t.Fatal("overlapping outer intervals must cover the union")
	}
	if IntervalsCover([]TimeInterval{iv(540, 600), iv(601, 700)}, []TimeInterval{iv(560, 650)}) {
		t.Fatal("a gap in outer must not be covered")
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

func TestParseCalendarDateLeapYears(t *testing.T) {
	t.Parallel()
	cases := []struct {
		raw     string
		wantErr bool
	}{
		{raw: "2100-02-29", wantErr: true}, // 100 の倍数だが 400 の倍数ではない
		{raw: "2000-02-29"},                // 400 の倍数
		{raw: "2026-02-29", wantErr: true},
		{raw: "2026-12-31"},
		{raw: "2026-04-31", wantErr: true},
		{raw: "2026-00-10", wantErr: true},
		{raw: "2026-01-00", wantErr: true},
	}
	for _, tc := range cases {
		d, err := ParseCalendarDate(tc.raw)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseCalendarDate(%q) accepted", tc.raw)
			}
			continue
		}
		if err != nil || d.String() != tc.raw || d.IsZero() {
			t.Errorf("ParseCalendarDate(%q) = %v, %v", tc.raw, d, err)
		}
	}
}

func TestCalendarDateCompare(t *testing.T) {
	t.Parallel()
	mk := func(raw string) CalendarDate {
		d, err := ParseCalendarDate(raw)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	cases := []struct {
		a, b string
		want int
	}{
		{a: "2026-01-05", b: "2026-01-05", want: 0},
		{a: "2026-01-05", b: "2026-01-06", want: -1},
		{a: "2026-01-06", b: "2026-01-05", want: 1},
		{a: "2026-01-31", b: "2026-02-01", want: -1},
		{a: "2026-12-31", b: "2027-01-01", want: -1},
		{a: "2027-01-01", b: "2026-12-31", want: 1},
		{a: "2026-02-01", b: "2025-12-31", want: 1},
	}
	for _, tc := range cases {
		if got := mk(tc.a).Compare(mk(tc.b)); got != tc.want {
			t.Errorf("Compare(%s, %s) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCalendarDateOfUsesJST(t *testing.T) {
	t.Parallel()
	utc := time.Date(2026, 10, 21, 15, 30, 0, 0, time.UTC) // 2026-10-22 00:30 JST
	if got := CalendarDateOf(utc).String(); got != "2026-10-22" {
		t.Fatalf("got %s, want 2026-10-22", got)
	}
}

func TestCalendarDateOfJSTBoundary(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		at   time.Time
		want string
	}{
		{name: "one nanosecond before JST midnight", at: time.Date(2026, 10, 22, 14, 59, 59, 999999999, time.UTC), want: "2026-10-22"},
		{name: "JST midnight", at: time.Date(2026, 10, 22, 15, 0, 0, 0, time.UTC), want: "2026-10-23"},
	}
	for _, tc := range cases {
		if got := CalendarDateOf(tc.at).String(); got != tc.want {
			t.Errorf("%s: got %s, want %s", tc.name, got, tc.want)
		}
	}
}

// assertIntervals は内容に加えて nil と空を区別する（want が nil なら got も nil、空なら非 nil の空）。
func assertIntervals(t *testing.T, label string, got, want []TimeInterval) {
	t.Helper()
	if (got == nil) != (want == nil) {
		t.Fatalf("%s: nil-ness differs: got %#v, want %#v", label, got, want)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("%s: got %+v, want %+v", label, got, want)
	}
}
