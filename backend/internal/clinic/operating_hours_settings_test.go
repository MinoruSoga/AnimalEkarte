package clinic

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// 共通医院時間 Settings の検証・不変性・wire（emr-clinic-line-time-selected-contract-20261022 §2.1）。
// DB・HTTP には依存しない。

func ohBoolPtr(v bool) *bool { return &v }

func ohWeeklyPtr(v []WeeklyHoursInput) *[]WeeklyHoursInput { return &v }

func ohClosedPtr(v []ClosedDateInput) *[]ClosedDateInput { return &v }

// ohWeek は日曜休診・月〜金 09:00-12:00/13:00-18:00・土 09:00-12:00 の週。
func ohWeek() []WeeklyHoursInput {
	am := IntervalInput{Start: "09:00", End: "12:00"}
	pm := IntervalInput{Start: "13:00", End: "18:00"}
	return []WeeklyHoursInput{
		{Weekday: 0, Intervals: []IntervalInput{}},
		{Weekday: 1, Intervals: []IntervalInput{am, pm}},
		{Weekday: 2, Intervals: []IntervalInput{am, pm}},
		{Weekday: 3, Intervals: []IntervalInput{am, pm}},
		{Weekday: 4, Intervals: []IntervalInput{am, pm}},
		{Weekday: 5, Intervals: []IntervalInput{am, pm}},
		{Weekday: 6, Intervals: []IntervalInput{am}},
	}
}

func ohClosedList() []ClosedDateInput {
	return []ClosedDateInput{
		{Date: "2026-01-12", Reason: "棚卸し"},
		{Date: "2026-03-01", Reason: ""},
	}
}

func ohFullFields() OperatingHoursFields {
	return OperatingHoursFields{
		WeeklyHours:           ohWeeklyPtr(ohWeek()),
		ClosedDates:           ohClosedPtr(ohClosedList()),
		NationalHolidayClosed: ohBoolPtr(true),
	}
}

func ohMustSettings(t *testing.T, f OperatingHoursFields) OperatingHoursSettings {
	t.Helper()
	s, err := NewOperatingHoursSettings(f)
	if err != nil {
		t.Fatalf("NewOperatingHoursSettings: %v", err)
	}
	return s
}

func ohMustDate(t *testing.T, raw string) CalendarDate {
	t.Helper()
	d, err := ParseCalendarDate(raw)
	if err != nil {
		t.Fatalf("ParseCalendarDate(%q): %v", raw, err)
	}
	return d
}

func ohWeeklyIntervals(t *testing.T, s OperatingHoursSettings, w time.Weekday) []TimeInterval {
	t.Helper()
	got, err := s.WeeklyIntervals(w)
	if err != nil {
		t.Fatalf("WeeklyIntervals: %v", err)
	}
	return got
}

func ohClosedDates(t *testing.T, s OperatingHoursSettings) []ClosedDate {
	t.Helper()
	got, err := s.ClosedDates()
	if err != nil {
		t.Fatalf("ClosedDates: %v", err)
	}
	return got
}

func ohNationalHolidayClosed(t *testing.T, s OperatingHoursSettings) bool {
	t.Helper()
	got, err := s.NationalHolidayClosed()
	if err != nil {
		t.Fatalf("NationalHolidayClosed: %v", err)
	}
	return got
}

func ohClosedDateOn(t *testing.T, s OperatingHoursSettings, d CalendarDate) (ClosedDate, bool) {
	t.Helper()
	cd, found, err := s.ClosedDateOn(d)
	if err != nil {
		t.Fatalf("ClosedDateOn: %v", err)
	}
	return cd, found
}

func ohIntervalsOn(t *testing.T, s OperatingHoursSettings, d CalendarDate, isNationalHoliday bool) []TimeInterval {
	t.Helper()
	got, err := s.IntervalsOn(d, isNationalHoliday)
	if err != nil {
		t.Fatalf("IntervalsOn: %v", err)
	}
	return got
}

// ohAssertMisuse は err が bare sentinel（誤用の内部エラー）であり、入力不正でも C06 の確定拒否でもないことを検査する。
func ohAssertMisuse(t *testing.T, label string, err, sentinel error) {
	t.Helper()
	if !errors.Is(err, sentinel) {
		t.Fatalf("%s: got %v, want sentinel %v", label, err, sentinel)
	}
	if apperrors.IsInvalidInput(err) {
		t.Fatalf("%s: misuse must not be an invalid-input error", label)
	}
	if rej, ok := OperatingHoursRejection(err); ok || rej != nil {
		t.Fatalf("%s: misuse must not map to a C06 rejection, got %+v", label, rej)
	}
}

func ohMustJSON(t *testing.T, s OperatingHoursSettings) string {
	t.Helper()
	b, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	return string(b)
}

func TestNewOperatingHoursSettingsRequiresEveryField(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		mutate func(*OperatingHoursFields)
	}{
		{name: "weekly_hours missing", mutate: func(f *OperatingHoursFields) { f.WeeklyHours = nil }},
		{name: "closed_dates missing", mutate: func(f *OperatingHoursFields) { f.ClosedDates = nil }},
		{name: "national_holiday_closed missing", mutate: func(f *OperatingHoursFields) { f.NationalHolidayClosed = nil }},
		{name: "all missing", mutate: func(f *OperatingHoursFields) { *f = OperatingHoursFields{} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := ohFullFields()
			tc.mutate(&f)
			_, err := NewOperatingHoursSettings(f)
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("got %v, want invalid input", err)
			}
		})
	}
	if s := ohMustSettings(t, ohFullFields()); !s.Valid() {
		t.Fatal("validated construction must produce a valid Settings")
	}
}

func TestNewOperatingHoursSettingsWeeklyHoursShape(t *testing.T) {
	t.Parallel()
	swapped := ohWeek()
	swapped[1], swapped[2] = swapped[2], swapped[1]
	duplicated := ohWeek()
	duplicated[2].Weekday = 1
	outOfRange := ohWeek()
	outOfRange[6].Weekday = 7
	negative := ohWeek()
	negative[0].Weekday = -1
	cases := []struct {
		name    string
		weekly  []WeeklyHoursInput
		wantErr bool
	}{
		{name: "seven ascending days", weekly: ohWeek()},
		{name: "empty", weekly: []WeeklyHoursInput{}, wantErr: true},
		{name: "nil", weekly: nil, wantErr: true},
		{name: "six days (missing)", weekly: ohWeek()[:6], wantErr: true},
		{name: "eight days", weekly: append(ohWeek(), WeeklyHoursInput{Weekday: 7}), wantErr: true},
		{name: "duplicate weekday", weekly: duplicated, wantErr: true},
		{name: "unordered weekdays", weekly: swapped, wantErr: true},
		{name: "weekday out of range", weekly: outOfRange, wantErr: true},
		{name: "negative weekday", weekly: negative, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := ohFullFields()
			f.WeeklyHours = ohWeeklyPtr(tc.weekly)
			_, err := NewOperatingHoursSettings(f)
			if !tc.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("got %v, want invalid input", err)
			}
		})
	}
}

func TestNewOperatingHoursSettingsIntervals(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		intervals []IntervalInput
		want      []TimeInterval
		wantErr   bool
	}{
		{name: "empty is an explicit closed day", intervals: []IntervalInput{}, want: []TimeInterval{}},
		{
			name:      "end of day 24:00 accepted",
			intervals: []IntervalInput{{Start: "22:00", End: "24:00"}},
			want:      []TimeInterval{{Start: 1320, End: 1440}},
		},
		{
			name:      "adjacent intervals accepted as given (no merge)",
			intervals: []IntervalInput{{Start: "09:00", End: "12:00"}, {Start: "12:00", End: "15:00"}},
			want:      []TimeInterval{{Start: 540, End: 720}, {Start: 720, End: 900}},
		},
		{
			name:      "gap between intervals accepted",
			intervals: []IntervalInput{{Start: "09:00", End: "12:00"}, {Start: "13:00", End: "15:00"}},
			want:      []TimeInterval{{Start: 540, End: 720}, {Start: 780, End: 900}},
		},
		{
			name:      "unsorted rejected",
			intervals: []IntervalInput{{Start: "13:00", End: "15:00"}, {Start: "09:00", End: "12:00"}},
			wantErr:   true,
		},
		{
			name:      "overlapping rejected",
			intervals: []IntervalInput{{Start: "09:00", End: "13:00"}, {Start: "12:00", End: "15:00"}},
			wantErr:   true,
		},
		{
			name:      "identical duplicate rejected",
			intervals: []IntervalInput{{Start: "09:00", End: "12:00"}, {Start: "09:00", End: "12:00"}},
			wantErr:   true,
		},
		{name: "start equals end rejected", intervals: []IntervalInput{{Start: "09:00", End: "09:00"}}, wantErr: true},
		{name: "reversed rejected", intervals: []IntervalInput{{Start: "18:00", End: "09:00"}}, wantErr: true},
		{name: "legacy HHMM rejected", intervals: []IntervalInput{{Start: "0900", End: "1800"}}, wantErr: true},
		{name: "start 24:00 rejected", intervals: []IntervalInput{{Start: "24:00", End: "24:00"}}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			week := ohWeek()
			week[3].Intervals = tc.intervals
			f := ohFullFields()
			f.WeeklyHours = ohWeeklyPtr(week)
			s, err := NewOperatingHoursSettings(f)
			if tc.wantErr {
				if !apperrors.IsInvalidInput(err) {
					t.Fatalf("got %v, want invalid input", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertIntervals(t, "wednesday", ohWeeklyIntervals(t, s, time.Wednesday), tc.want)
		})
	}
}

func TestAdjacentIntervalsAreNotNormalizedAndEqualIsStructural(t *testing.T) {
	t.Parallel()
	adjacent := ohWeek()
	adjacent[1].Intervals = []IntervalInput{{Start: "09:00", End: "12:00"}, {Start: "12:00", End: "15:00"}}
	merged := ohWeek()
	merged[1].Intervals = []IntervalInput{{Start: "09:00", End: "15:00"}}
	fa, fm := ohFullFields(), ohFullFields()
	fa.WeeklyHours = ohWeeklyPtr(adjacent)
	fm.WeeklyHours = ohWeeklyPtr(merged)
	sa, sm := ohMustSettings(t, fa), ohMustSettings(t, fm)

	if got := len(ohWeeklyIntervals(t, sa, time.Monday)); got != 2 {
		t.Fatalf("adjacent intervals were normalized: %d intervals", got)
	}
	if sa.Equal(sm) || sm.Equal(sa) {
		t.Fatal("[09:00,15:00) must not be Equal to the two adjacent intervals (structural Equal)")
	}
	if !sa.Equal(ohMustSettings(t, fa)) {
		t.Fatal("identical adjacent input must be Equal")
	}
	wantJSON := `{"start":"09:00","end":"12:00"},{"start":"12:00","end":"15:00"}`
	if got := ohMustJSON(t, sa); !strings.Contains(got, wantJSON) {
		t.Fatalf("adjacent intervals not preserved on the wire: %s", got)
	}
}

func TestNewOperatingHoursSettingsClosedDates(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		closed  []ClosedDateInput
		wantLen int
		wantErr bool
	}{
		{name: "nil list accepted as empty", closed: nil, wantLen: 0},
		{name: "empty list", closed: []ClosedDateInput{}, wantLen: 0},
		{name: "ascending unique", closed: ohClosedList(), wantLen: 2},
		{name: "empty reason accepted", closed: []ClosedDateInput{{Date: "2026-01-12", Reason: ""}}, wantLen: 1},
		{name: "newline in reason accepted", closed: []ClosedDateInput{{Date: "2026-01-12", Reason: "a\nb"}}, wantLen: 1},
		{name: "tab in reason accepted", closed: []ClosedDateInput{{Date: "2026-01-12", Reason: "a\tb"}}, wantLen: 1},
		{name: "NUL in reason rejected", closed: []ClosedDateInput{{Date: "2026-01-12", Reason: "a\x00b"}}, wantErr: true},
		{
			name:    "500 runes accepted",
			closed:  []ClosedDateInput{{Date: "2026-01-12", Reason: strings.Repeat("休", 500)}},
			wantLen: 1,
		},
		{
			name:    "501 runes rejected",
			closed:  []ClosedDateInput{{Date: "2026-01-12", Reason: strings.Repeat("休", 501)}},
			wantErr: true,
		},
		{
			name:    "invalid UTF-8 rejected",
			closed:  []ClosedDateInput{{Date: "2026-01-12", Reason: "bad\xff"}},
			wantErr: true,
		},
		{
			name:    "duplicate date rejected",
			closed:  []ClosedDateInput{{Date: "2026-01-12"}, {Date: "2026-01-12", Reason: "x"}},
			wantErr: true,
		},
		{
			name:    "descending rejected",
			closed:  []ClosedDateInput{{Date: "2026-03-01"}, {Date: "2026-01-12"}},
			wantErr: true,
		},
		{name: "nonexistent date rejected", closed: []ClosedDateInput{{Date: "2026-02-30"}}, wantErr: true},
		{name: "unpadded date rejected", closed: []ClosedDateInput{{Date: "2026-1-05"}}, wantErr: true},
		{name: "slash date rejected", closed: []ClosedDateInput{{Date: "2026/01/05"}}, wantErr: true},
		{name: "empty date rejected", closed: []ClosedDateInput{{Date: ""}}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := ohFullFields()
			f.ClosedDates = ohClosedPtr(tc.closed)
			s, err := NewOperatingHoursSettings(f)
			if tc.wantErr {
				if !apperrors.IsInvalidInput(err) {
					t.Fatalf("got %v, want invalid input", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := len(ohClosedDates(t, s)); got != tc.wantLen {
				t.Fatalf("got %d closed dates, want %d", got, tc.wantLen)
			}
		})
	}
}

func TestApplyChanges(t *testing.T) {
	t.Parallel()
	base := ohMustSettings(t, ohFullFields())
	baseJSON := ohMustJSON(t, base)

	t.Run("empty changes rejected", func(t *testing.T) {
		t.Parallel()
		_, err := base.ApplyChanges(OperatingHoursFields{})
		if !apperrors.IsInvalidInput(err) {
			t.Fatalf("got %v, want invalid input", err)
		}
	})

	t.Run("replacing one field keeps the others", func(t *testing.T) {
		t.Parallel()
		next, err := base.ApplyChanges(OperatingHoursFields{NationalHolidayClosed: ohBoolPtr(false)})
		if err != nil {
			t.Fatal(err)
		}
		if ohNationalHolidayClosed(t, next) {
			t.Fatal("national_holiday_closed was not replaced")
		}
		want := strings.Replace(baseJSON, `"national_holiday_closed":true`, `"national_holiday_closed":false`, 1)
		if got := ohMustJSON(t, next); got != want {
			t.Fatalf("other fields changed:\n got %s\nwant %s", got, want)
		}
	})

	t.Run("set field is replaced wholesale", func(t *testing.T) {
		t.Parallel()
		next, err := base.ApplyChanges(OperatingHoursFields{ClosedDates: ohClosedPtr([]ClosedDateInput{{Date: "2026-05-05", Reason: "新規"}})})
		if err != nil {
			t.Fatal(err)
		}
		got := ohClosedDates(t, next)
		if len(got) != 1 || got[0].Date.String() != "2026-05-05" || got[0].Reason != "新規" {
			t.Fatalf("closed_dates not replaced wholesale: %+v", got)
		}
		if len(ohWeeklyIntervals(t, next, time.Monday)) != 2 {
			t.Fatal("weekly_hours must be kept when only closed_dates is replaced")
		}
	})

	t.Run("weekly_hours replaced wholesale to all closed", func(t *testing.T) {
		t.Parallel()
		closedWeek := make([]WeeklyHoursInput, daysPerWeek)
		for i := range closedWeek {
			closedWeek[i] = WeeklyHoursInput{Weekday: i, Intervals: []IntervalInput{}}
		}
		next, err := base.ApplyChanges(OperatingHoursFields{WeeklyHours: ohWeeklyPtr(closedWeek)})
		if err != nil {
			t.Fatal(err)
		}
		for w := time.Sunday; w <= time.Saturday; w++ {
			if got := ohWeeklyIntervals(t, next, w); len(got) != 0 {
				t.Fatalf("weekday %d not replaced: %+v", w, got)
			}
		}
		if len(ohClosedDates(t, next)) != 2 || !ohNationalHolidayClosed(t, next) {
			t.Fatal("closed_dates and national_holiday_closed must be kept")
		}
	})

	t.Run("invalid replacement rejected and receiver untouched", func(t *testing.T) {
		t.Parallel()
		_, err := base.ApplyChanges(OperatingHoursFields{WeeklyHours: ohWeeklyPtr(ohWeek()[:3])})
		if !apperrors.IsInvalidInput(err) {
			t.Fatalf("got %v, want invalid input", err)
		}
		if got := ohMustJSON(t, base); got != baseJSON {
			t.Fatal("receiver mutated by a rejected change")
		}
	})

	t.Run("invalid receiver rejected", func(t *testing.T) {
		t.Parallel()
		got, err := OperatingHoursSettings{}.ApplyChanges(OperatingHoursFields{NationalHolidayClosed: ohBoolPtr(true)})
		ohAssertMisuse(t, "ApplyChanges on zero", err, errInvalidOperatingHoursSettings)
		if got.Valid() {
			t.Fatal("a rejected change must not return a valid Settings")
		}
	})
}

func TestOperatingHoursSettingsZeroValueIsInvalid(t *testing.T) {
	t.Parallel()
	var zero OperatingHoursSettings
	valid := ohMustSettings(t, ohFullFields())

	if zero.Valid() {
		t.Fatal("zero value must not be valid")
	}
	if _, err := zero.MarshalJSON(); !errors.Is(err, errInvalidOperatingHoursSettings) {
		t.Fatalf("MarshalJSON on zero: got %v", err)
	}
	if _, err := json.Marshal(zero); err == nil {
		t.Fatal("json.Marshal must fail for the zero value")
	}
	if _, err := json.Marshal(struct{ S OperatingHoursSettings }{S: zero}); err == nil {
		t.Fatal("json.Marshal must fail for a zero value embedded in a struct")
	}
	_, _, err := zero.WithClosedDate(ohMustDate(t, "2026-01-05"), "x")
	ohAssertMisuse(t, "WithClosedDate on zero", err, errInvalidOperatingHoursSettings)
	got, changed, err := zero.WithoutClosedDate(ohMustDate(t, "2026-01-05"))
	ohAssertMisuse(t, "WithoutClosedDate on zero", err, errInvalidOperatingHoursSettings)
	if changed || got.Valid() {
		t.Fatalf("WithoutClosedDate on zero: got valid=%v changed=%v", got.Valid(), changed)
	}
	if !zero.Equal(OperatingHoursSettings{}) {
		t.Fatal("two invalid values must be equal")
	}
	if zero.Equal(valid) || valid.Equal(zero) {
		t.Fatal("invalid and valid must be unequal")
	}
	if zero.clone().Valid() {
		t.Fatal("clone of an invalid value must stay invalid")
	}
	if !valid.clone().Valid() {
		t.Fatal("clone of a valid value must stay valid")
	}

	// 全日休診を明示した有効な Settings は、ゼロ値とは別物（Equal は不等）。
	closedWeek := make([]WeeklyHoursInput, daysPerWeek)
	for i := range closedWeek {
		closedWeek[i] = WeeklyHoursInput{Weekday: i, Intervals: []IntervalInput{}}
	}
	explicit := ohMustSettings(t, OperatingHoursFields{
		WeeklyHours:           ohWeeklyPtr(closedWeek),
		ClosedDates:           ohClosedPtr(nil),
		NationalHolidayClosed: ohBoolPtr(false),
	})
	if !explicit.Valid() || explicit.Equal(zero) {
		t.Fatal("an explicit all-closed Settings must be valid and unequal to the zero value")
	}
}

func TestOperatingHoursSettingsZeroValueReadsReturnErrors(t *testing.T) {
	t.Parallel()
	var zero OperatingHoursSettings
	date := ohMustDate(t, "2026-01-05")

	weekly, err := zero.WeeklyIntervals(time.Monday)
	ohAssertMisuse(t, "WeeklyIntervals", err, errInvalidOperatingHoursSettings)
	if weekly != nil {
		t.Fatalf("WeeklyIntervals must not return a silent value, got %+v", weekly)
	}
	closed, err := zero.ClosedDates()
	ohAssertMisuse(t, "ClosedDates", err, errInvalidOperatingHoursSettings)
	if closed != nil {
		t.Fatalf("ClosedDates must not return a silent value, got %+v", closed)
	}
	policy, err := zero.NationalHolidayClosed()
	ohAssertMisuse(t, "NationalHolidayClosed", err, errInvalidOperatingHoursSettings)
	if policy {
		t.Fatal("NationalHolidayClosed must not claim a policy")
	}
	cd, found, err := zero.ClosedDateOn(date)
	ohAssertMisuse(t, "ClosedDateOn", err, errInvalidOperatingHoursSettings)
	if found || cd != (ClosedDate{}) {
		t.Fatalf("ClosedDateOn must not return a value, got %+v %v", cd, found)
	}
	on, err := zero.IntervalsOn(date, false)
	ohAssertMisuse(t, "IntervalsOn", err, errInvalidOperatingHoursSettings)
	if on != nil {
		t.Fatalf("IntervalsOn must not return a silent [] for invalid Settings, got %+v", on)
	}
}

func TestZeroCalendarDateIsRejected(t *testing.T) {
	t.Parallel()
	s := ohMustSettings(t, ohFullFields())
	var zeroDate CalendarDate

	if !zeroDate.IsZero() || zeroDate.String() != "" {
		t.Fatalf("zero date: IsZero=%v String=%q", zeroDate.IsZero(), zeroDate.String())
	}
	if ohMustDate(t, "2026-01-05").IsZero() || CalendarDateOf(time.Now()).IsZero() {
		t.Fatal("dates from ParseCalendarDate/CalendarDateOf must not be zero")
	}

	_, _, err := s.ClosedDateOn(zeroDate)
	ohAssertMisuse(t, "ClosedDateOn", err, errZeroCalendarDate)
	_, err = s.IntervalsOn(zeroDate, false)
	ohAssertMisuse(t, "IntervalsOn", err, errZeroCalendarDate)
	_, changed, err := s.WithClosedDate(zeroDate, "x")
	ohAssertMisuse(t, "WithClosedDate", err, errZeroCalendarDate)
	if changed {
		t.Fatal("WithClosedDate with a zero date must not report a change")
	}
	_, changed, err = s.WithoutClosedDate(zeroDate)
	ohAssertMisuse(t, "WithoutClosedDate", err, errZeroCalendarDate)
	if changed {
		t.Fatal("WithoutClosedDate with a zero date must not report a change")
	}
	// 受け手が invalid の場合は Settings 側の sentinel が先に返る。
	_, _, err = OperatingHoursSettings{}.ClosedDateOn(zeroDate)
	ohAssertMisuse(t, "ClosedDateOn on zero settings", err, errInvalidOperatingHoursSettings)
}

func TestOperatingHoursSettingsUnmarshalJSONAlwaysFails(t *testing.T) {
	t.Parallel()
	valid := ohMustSettings(t, ohFullFields())
	payloads := []string{ohMustJSON(t, valid), `{}`, `null`, ``}
	for _, raw := range payloads {
		var got OperatingHoursSettings
		err := json.Unmarshal([]byte(raw), &got)
		if err == nil {
			t.Fatalf("json.Unmarshal(%q) must fail", raw)
		}
		if got.Valid() {
			t.Fatalf("json.Unmarshal(%q) produced a valid Settings", raw)
		}
	}
	var direct OperatingHoursSettings
	if err := direct.UnmarshalJSON([]byte(`{}`)); !errors.Is(err, errOperatingHoursSettingsDecode) {
		t.Fatalf("got %v", err)
	}
	var wrapper struct{ S OperatingHoursSettings }
	if err := json.Unmarshal([]byte(`{"S":{}}`), &wrapper); err == nil {
		t.Fatal("decoding through a wrapping struct must fail")
	}
}

func TestOperatingHoursSettingsMarshalJSON(t *testing.T) {
	t.Parallel()
	const weekdays = `[{"weekday":0,"intervals":[]},` +
		`{"weekday":1,"intervals":[{"start":"09:00","end":"12:00"},{"start":"13:00","end":"18:00"}]},` +
		`{"weekday":2,"intervals":[{"start":"09:00","end":"12:00"},{"start":"13:00","end":"18:00"}]},` +
		`{"weekday":3,"intervals":[{"start":"09:00","end":"12:00"},{"start":"13:00","end":"18:00"}]},` +
		`{"weekday":4,"intervals":[{"start":"09:00","end":"12:00"},{"start":"13:00","end":"18:00"}]},` +
		`{"weekday":5,"intervals":[{"start":"09:00","end":"12:00"},{"start":"13:00","end":"18:00"}]},` +
		`{"weekday":6,"intervals":[{"start":"09:00","end":"12:00"}]}]`

	t.Run("full settings exact bytes", func(t *testing.T) {
		t.Parallel()
		want := `{"weekly_hours":` + weekdays +
			`,"closed_dates":[{"date":"2026-01-12","reason":"棚卸し"},{"date":"2026-03-01","reason":""}]` +
			`,"national_holiday_closed":true}`
		s := ohMustSettings(t, ohFullFields())
		if got := ohMustJSON(t, s); got != want {
			t.Fatalf("\n got %s\nwant %s", got, want)
		}
		b, err := json.Marshal(s)
		if err != nil || string(b) != want {
			t.Fatalf("json.Marshal differs from MarshalJSON: %s (%v)", b, err)
		}
	})

	t.Run("empty closed_dates is [] and false policy is explicit", func(t *testing.T) {
		t.Parallel()
		f := ohFullFields()
		f.ClosedDates = ohClosedPtr(nil)
		f.NationalHolidayClosed = ohBoolPtr(false)
		want := `{"weekly_hours":` + weekdays + `,"closed_dates":[],"national_holiday_closed":false}`
		if got := ohMustJSON(t, ohMustSettings(t, f)); got != want {
			t.Fatalf("\n got %s\nwant %s", got, want)
		}
	})

	t.Run("24:00 end is emitted verbatim", func(t *testing.T) {
		t.Parallel()
		week := ohWeek()
		week[0].Intervals = []IntervalInput{{Start: "00:00", End: "24:00"}}
		f := ohFullFields()
		f.WeeklyHours = ohWeeklyPtr(week)
		want := `{"weekday":0,"intervals":[{"start":"00:00","end":"24:00"}]}`
		if got := ohMustJSON(t, ohMustSettings(t, f)); !strings.Contains(got, want) {
			t.Fatalf("24:00 not preserved: %s", got)
		}
	})

	t.Run("seven weekday entries in order", func(t *testing.T) {
		t.Parallel()
		var wire struct {
			WeeklyHours []struct {
				Weekday   int               `json:"weekday"`
				Intervals []json.RawMessage `json:"intervals"`
			} `json:"weekly_hours"`
		}
		if err := json.Unmarshal([]byte(ohMustJSON(t, ohMustSettings(t, ohFullFields()))), &wire); err != nil {
			t.Fatal(err)
		}
		if len(wire.WeeklyHours) != daysPerWeek {
			t.Fatalf("got %d weekday entries", len(wire.WeeklyHours))
		}
		for i, day := range wire.WeeklyHours {
			if day.Weekday != i || day.Intervals == nil {
				t.Fatalf("entry %d: weekday=%d intervals-nil=%v", i, day.Weekday, day.Intervals == nil)
			}
		}
	})

	t.Run("canonical bytes pin json.Marshal HTML escaping of the reason", func(t *testing.T) {
		t.Parallel()
		f := ohFullFields()
		f.ClosedDates = ohClosedPtr([]ClosedDateInput{{Date: "2026-01-12", Reason: `<&>"`}})
		want := `"closed_dates":[{"date":"2026-01-12","reason":"\u003c\u0026\u003e\""}]`
		s := ohMustSettings(t, f)
		if got := ohMustJSON(t, s); !strings.Contains(got, want) {
			t.Fatalf("reason bytes changed:\n got %s\nwant to contain %s", got, want)
		}
		b, err := json.Marshal(s)
		if err != nil || !strings.Contains(string(b), want) {
			t.Fatalf("json.Marshal differs: %s (%v)", b, err)
		}
	})

	t.Run("zero value returns an error", func(t *testing.T) {
		t.Parallel()
		if b, err := (OperatingHoursSettings{}).MarshalJSON(); err == nil {
			t.Fatalf("expected an error, got %s", b)
		}
	})
}

func TestOperatingHoursSettingsIntervalsOn(t *testing.T) {
	t.Parallel()
	// 2026-01-05 は月曜、2026-01-12 は月曜（個別休診）、2026-01-06 は火曜。
	monday := []TimeInterval{{Start: 540, End: 720}, {Start: 780, End: 1080}}
	cases := []struct {
		name            string
		holidayPolicy   bool
		date            string
		isNationalHoday bool
		want            []TimeInterval
	}{
		{name: "ordinary day returns the weekday intervals", holidayPolicy: true, date: "2026-01-05", want: monday},
		{name: "closed date is empty", holidayPolicy: true, date: "2026-01-12", want: []TimeInterval{}},
		{name: "closed date is empty even when not a holiday and policy false", holidayPolicy: false, date: "2026-01-12", want: []TimeInterval{}},
		{name: "national holiday with closing policy is empty", holidayPolicy: true, date: "2026-01-05", isNationalHoday: true, want: []TimeInterval{}},
		{name: "national holiday without closing policy keeps the weekday intervals", holidayPolicy: false, date: "2026-01-05", isNationalHoday: true, want: monday},
		{name: "closed weekday is empty", holidayPolicy: false, date: "2026-01-04", want: []TimeInterval{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := ohFullFields()
			f.NationalHolidayClosed = ohBoolPtr(tc.holidayPolicy)
			s := ohMustSettings(t, f)
			got := ohIntervalsOn(t, s, ohMustDate(t, tc.date), tc.isNationalHoday)
			if got == nil {
				t.Fatal("IntervalsOn must return a non-nil (possibly empty) slice")
			}
			assertIntervals(t, tc.name, got, tc.want)
		})
	}
}

func TestOperatingHoursSettingsGettersReturnCopies(t *testing.T) {
	t.Parallel()
	s := ohMustSettings(t, ohFullFields())
	before := ohMustJSON(t, s)
	monday := ohMustDate(t, "2026-01-05")

	on := ohIntervalsOn(t, s, monday, false)
	on[0] = TimeInterval{Start: 0, End: 1}
	weekly := ohWeeklyIntervals(t, s, time.Monday)
	weekly[0] = TimeInterval{Start: 0, End: 1}
	closed := ohClosedDates(t, s)
	closed[0] = ClosedDate{Date: monday, Reason: "tampered"}

	if got := ohMustJSON(t, s); got != before {
		t.Fatalf("Settings changed through a returned slice:\n got %s\nwant %s", got, before)
	}
}

func TestWeeklyIntervalsRejectsOutOfRangeWeekday(t *testing.T) {
	t.Parallel()
	s := ohMustSettings(t, ohFullFields())
	for _, w := range []time.Weekday{-1, 7, 9} {
		got, err := s.WeeklyIntervals(w)
		ohAssertMisuse(t, "WeeklyIntervals", err, errInvalidWeekday)
		if got != nil {
			t.Fatalf("weekday %d: must not return a silent value, got %+v", w, got)
		}
	}
	// IntervalsOn は CalendarDate.Weekday() の範囲内の値しか渡さないので、有効な全曜日が通る。
	for w := time.Sunday; w <= time.Saturday; w++ {
		if _, err := s.WeeklyIntervals(w); err != nil {
			t.Fatalf("weekday %d: %v", w, err)
		}
	}
}

func TestClosedDateOn(t *testing.T) {
	t.Parallel()
	s := ohMustSettings(t, ohFullFields())
	got, ok := ohClosedDateOn(t, s, ohMustDate(t, "2026-01-12"))
	if !ok || got.Reason != "棚卸し" {
		t.Fatalf("got %+v ok=%v", got, ok)
	}
	if _, ok := ohClosedDateOn(t, s, ohMustDate(t, "2026-01-13")); ok {
		t.Fatal("non-closed date reported as closed")
	}
}

func TestWithClosedDate(t *testing.T) {
	t.Parallel()
	base := ohMustSettings(t, ohFullFields()) // 2026-01-12 棚卸し, 2026-03-01 ""
	baseJSON := ohMustJSON(t, base)

	insertCases := []struct {
		name string
		date string
		want []string
	}{
		{name: "before the first", date: "2026-01-01", want: []string{"2026-01-01", "2026-01-12", "2026-03-01"}},
		{name: "between", date: "2026-02-01", want: []string{"2026-01-12", "2026-02-01", "2026-03-01"}},
		{name: "after the last", date: "2026-12-31", want: []string{"2026-01-12", "2026-03-01", "2026-12-31"}},
	}
	for _, tc := range insertCases {
		t.Run("insert "+tc.name+" keeps ascending order", func(t *testing.T) {
			t.Parallel()
			next, changed, err := base.WithClosedDate(ohMustDate(t, tc.date), "臨時")
			if err != nil || !changed {
				t.Fatalf("changed=%v err=%v", changed, err)
			}
			var got []string
			for _, cd := range ohClosedDates(t, next) {
				got = append(got, cd.Date.String())
			}
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			if cd, _ := ohClosedDateOn(t, next, ohMustDate(t, tc.date)); cd.Reason != "臨時" {
				t.Fatalf("reason not stored: %+v", cd)
			}
			if !next.Valid() {
				t.Fatal("result must stay valid")
			}
			if got := ohMustJSON(t, base); got != baseJSON {
				t.Fatal("receiver mutated")
			}
		})
	}

	t.Run("same date same reason is a no-op", func(t *testing.T) {
		t.Parallel()
		next, changed, err := base.WithClosedDate(ohMustDate(t, "2026-01-12"), "棚卸し")
		if err != nil || changed {
			t.Fatalf("changed=%v err=%v", changed, err)
		}
		if !next.Equal(base) || !next.Valid() {
			t.Fatal("no-op must return an equal valid copy")
		}
	})

	t.Run("same date empty reason is a no-op", func(t *testing.T) {
		t.Parallel()
		_, changed, err := base.WithClosedDate(ohMustDate(t, "2026-03-01"), "")
		if err != nil || changed {
			t.Fatalf("changed=%v err=%v", changed, err)
		}
	})

	t.Run("same date different reason is a conflict", func(t *testing.T) {
		t.Parallel()
		next, changed, err := base.WithClosedDate(ohMustDate(t, "2026-01-12"), "別理由")
		if !errors.Is(err, ErrClosedDateReasonConflict) || changed || next.Valid() {
			t.Fatalf("valid=%v changed=%v err=%v", next.Valid(), changed, err)
		}
		if got := ohMustJSON(t, base); got != baseJSON {
			t.Fatal("receiver mutated")
		}
	})

	reasonCases := []struct {
		name   string
		reason string
	}{
		{name: "501 runes", reason: strings.Repeat("休", 501)},
		{name: "invalid UTF-8", reason: "bad\xff"},
		{name: "NUL", reason: "a\x00b"},
	}
	for _, tc := range reasonCases {
		t.Run("invalid reason "+tc.name+" rejected", func(t *testing.T) {
			t.Parallel()
			_, changed, err := base.WithClosedDate(ohMustDate(t, "2026-02-01"), tc.reason)
			if !apperrors.IsInvalidInput(err) || changed {
				t.Fatalf("changed=%v err=%v", changed, err)
			}
		})
	}
}

func TestWithoutClosedDate(t *testing.T) {
	t.Parallel()
	base := ohMustSettings(t, ohFullFields())
	baseJSON := ohMustJSON(t, base)

	t.Run("missing date is a no-op", func(t *testing.T) {
		t.Parallel()
		next, changed, err := base.WithoutClosedDate(ohMustDate(t, "2026-02-01"))
		if err != nil || changed {
			t.Fatalf("changed=%v err=%v", changed, err)
		}
		if !next.Equal(base) || !next.Valid() {
			t.Fatal("no-op must return an equal valid copy")
		}
	})

	t.Run("existing date is removed", func(t *testing.T) {
		t.Parallel()
		next, changed, err := base.WithoutClosedDate(ohMustDate(t, "2026-01-12"))
		if err != nil || !changed {
			t.Fatalf("changed=%v err=%v", changed, err)
		}
		if _, ok := ohClosedDateOn(t, next, ohMustDate(t, "2026-01-12")); ok {
			t.Fatal("date still closed")
		}
		if got := ohClosedDates(t, next); len(got) != 1 || got[0].Date.String() != "2026-03-01" {
			t.Fatalf("remaining closed dates: %+v", got)
		}
		if got := ohMustJSON(t, base); got != baseJSON {
			t.Fatal("receiver mutated")
		}
	})

	t.Run("removing the last date yields an empty valid list", func(t *testing.T) {
		t.Parallel()
		f := ohFullFields()
		f.ClosedDates = ohClosedPtr([]ClosedDateInput{{Date: "2026-01-12", Reason: "x"}})
		next, changed, err := ohMustSettings(t, f).WithoutClosedDate(ohMustDate(t, "2026-01-12"))
		if err != nil || !changed || !next.Valid() || len(ohClosedDates(t, next)) != 0 {
			t.Fatalf("valid=%v changed=%v err=%v closed=%+v", next.Valid(), changed, err, ohClosedDates(t, next))
		}
		if !strings.Contains(ohMustJSON(t, next), `"closed_dates":[]`) {
			t.Fatal("empty closed_dates must marshal as []")
		}
	})
}
