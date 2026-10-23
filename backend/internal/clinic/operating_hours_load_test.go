package clinic

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const operatingHoursSevenClosedDays = `[{"weekday":0,"intervals":[]},{"weekday":1,"intervals":[]},{"weekday":2,"intervals":[]},{"weekday":3,"intervals":[]},{"weekday":4,"intervals":[]},{"weekday":5,"intervals":[]},{"weekday":6,"intervals":[]}]`

func operatingHoursStoredRow(weekly, closed string) model.ClinicOperatingHours {
	return model.ClinicOperatingHours{
		ClinicID:              7,
		SchemaVersion:         model.ClinicOperatingHoursSchemaV1,
		Version:               3,
		WeeklyHours:           []byte(weekly),
		ClosedDates:           []byte(closed),
		NationalHolidayClosed: true,
	}
}

// weeklyWithDay replaces the element of weekday w with raw (the other days stay closed).
func weeklyWithDay(w int, raw string) string {
	days := []string{
		`{"weekday":0,"intervals":[]}`, `{"weekday":1,"intervals":[]}`, `{"weekday":2,"intervals":[]}`,
		`{"weekday":3,"intervals":[]}`, `{"weekday":4,"intervals":[]}`, `{"weekday":5,"intervals":[]}`,
		`{"weekday":6,"intervals":[]}`,
	}
	days[w] = raw
	return "[" + strings.Join(days, ",") + "]"
}

func requireCorrupt(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrOperatingHoursStoredDataCorrupt)
	// A corrupt row is a server-side defect: it must never chain to the 400 input sentinel.
	assert.NotErrorIs(t, err, apperrors.ErrInvalidInput)
}

func TestLoadOperatingHoursSettings_Valid(t *testing.T) {
	weekly := weeklyWithDay(1, `{"weekday":1,"intervals":[{"start":"09:00","end":"12:00"},{"start":"16:00","end":"24:00"}]}`)
	closed := `[{"date":"2026-11-03","reason":"臨時休診"},{"date":"2026-12-31","reason":""}]`

	settings, err := LoadOperatingHoursSettings(operatingHoursStoredRow(weekly, closed))
	require.NoError(t, err)
	require.True(t, settings.Valid())

	monday, err := settings.WeeklyIntervals(1)
	require.NoError(t, err)
	require.Len(t, monday, 2)
	assert.Equal(t, "16:00", monday[1].Start.String())
	assert.Equal(t, "24:00", monday[1].End.String())
	sunday, err := settings.WeeklyIntervals(0)
	require.NoError(t, err)
	assert.Empty(t, sunday, "empty intervals are an explicit closed day")

	dates, err := settings.ClosedDates()
	require.NoError(t, err)
	require.Len(t, dates, 2)
	assert.Equal(t, "臨時休診", dates[0].Reason)
	assert.Equal(t, "", dates[1].Reason)
	holiday, err := settings.NationalHolidayClosed()
	require.NoError(t, err)
	assert.True(t, holiday)

	// Equal to what the validated constructor builds from the same fields (single source of truth).
	wireA, err := settings.MarshalJSON()
	require.NoError(t, err)
	weeklyInput := make([]WeeklyHoursInput, 0, 7)
	for w := range 7 {
		in := WeeklyHoursInput{Weekday: w, Intervals: []IntervalInput{}}
		if w == 1 {
			in.Intervals = []IntervalInput{{Start: "09:00", End: "12:00"}, {Start: "16:00", End: "24:00"}}
		}
		weeklyInput = append(weeklyInput, in)
	}
	closedInput := []ClosedDateInput{{Date: "2026-11-03", Reason: "臨時休診"}, {Date: "2026-12-31"}}
	holidayClosed := true
	want, err := NewOperatingHoursSettings(OperatingHoursFields{
		WeeklyHours: &weeklyInput, ClosedDates: &closedInput, NationalHolidayClosed: &holidayClosed,
	})
	require.NoError(t, err)
	wireB, err := want.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, string(wireB), string(wireA))
	assert.True(t, want.Equal(settings))
}

func TestLoadOperatingHoursSettings_AcceptsInsignificantWhitespaceAndEmptyClosedDates(t *testing.T) {
	spaced := "\n [ " + strings.TrimPrefix(operatingHoursSevenClosedDays, "[") + " \n"
	settings, err := LoadOperatingHoursSettings(operatingHoursStoredRow(spaced, " [ ] "))
	require.NoError(t, err)
	dates, err := settings.ClosedDates()
	require.NoError(t, err)
	assert.Empty(t, dates)
}

func TestLoadOperatingHoursSettings_RejectsCorruptWeeklyHours(t *testing.T) {
	cases := []struct {
		name   string
		weekly string
	}{
		{"empty raw", ``},
		{"whitespace only", "  \n"},
		{"JSON null", `null`},
		{"object instead of array", `{}`},
		{"unknown key in weekday element", weeklyWithDay(2, `{"weekday":2,"intervals":[],"note":"x"}`)},
		{"missing intervals key", weeklyWithDay(2, `{"weekday":2}`)},
		{"missing weekday key", weeklyWithDay(2, `{"intervals":[]}`)},
		{"duplicate weekday key", weeklyWithDay(2, `{"weekday":2,"weekday":2,"intervals":[]}`)},
		{"duplicate intervals key", weeklyWithDay(2, `{"weekday":2,"intervals":[],"intervals":[]}`)},
		{"duplicate start key in interval", weeklyWithDay(2, `{"weekday":2,"intervals":[{"start":"09:00","start":"09:00","end":"12:00"}]}`)},
		{"unknown key in interval", weeklyWithDay(2, `{"weekday":2,"intervals":[{"start":"09:00","end":"12:00","x":1}]}`)},
		{"null intervals", weeklyWithDay(2, `{"weekday":2,"intervals":null}`)},
		{"null element", strings.Replace(operatingHoursSevenClosedDays, `{"weekday":2,"intervals":[]}`, `null`, 1)},
		{"null interval", weeklyWithDay(2, `{"weekday":2,"intervals":[null]}`)},
		{"null start", weeklyWithDay(2, `{"weekday":2,"intervals":[{"start":null,"end":"12:00"}]}`)},
		{"number end", weeklyWithDay(2, `{"weekday":2,"intervals":[{"start":"09:00","end":1200}]}`)},
		{"string weekday", weeklyWithDay(2, `{"weekday":"2","intervals":[]}`)},
		{"fractional weekday", weeklyWithDay(2, `{"weekday":2.0,"intervals":[]}`)},
		{"exponent weekday", weeklyWithDay(2, `{"weekday":2e0,"intervals":[]}`)},
		{"negative zero weekday", weeklyWithDay(0, `{"weekday":-0,"intervals":[]}`)},
		{"weekday out of order", weeklyWithDay(2, `{"weekday":3,"intervals":[]}`)},
		{"six weekdays", `[{"weekday":0,"intervals":[]},{"weekday":1,"intervals":[]},{"weekday":2,"intervals":[]},{"weekday":3,"intervals":[]},{"weekday":4,"intervals":[]},{"weekday":5,"intervals":[]}]`},
		{"eight weekdays", strings.TrimSuffix(operatingHoursSevenClosedDays, "]") + `,{"weekday":7,"intervals":[]}]`},
		{"end before start", weeklyWithDay(2, `{"weekday":2,"intervals":[{"start":"12:00","end":"09:00"}]}`)},
		{"empty interval", weeklyWithDay(2, `{"weekday":2,"intervals":[{"start":"09:00","end":"09:00"}]}`)},
		{"start 24:00", weeklyWithDay(2, `{"weekday":2,"intervals":[{"start":"24:00","end":"24:00"}]}`)},
		{"end 00:00", weeklyWithDay(2, `{"weekday":2,"intervals":[{"start":"00:00","end":"00:00"}]}`)},
		{"non-padded time", weeklyWithDay(2, `{"weekday":2,"intervals":[{"start":"9:00","end":"12:00"}]}`)},
		{"seconds in time", weeklyWithDay(2, `{"weekday":2,"intervals":[{"start":"09:00:00","end":"12:00"}]}`)},
		{"overlapping intervals", weeklyWithDay(2, `{"weekday":2,"intervals":[{"start":"09:00","end":"12:00"},{"start":"11:00","end":"13:00"}]}`)},
		{"descending intervals", weeklyWithDay(2, `{"weekday":2,"intervals":[{"start":"16:00","end":"18:00"},{"start":"09:00","end":"12:00"}]}`)},
		{"trailing second value", operatingHoursSevenClosedDays + ` []`},
		{"trailing garbage", operatingHoursSevenClosedDays + `x`},
		{"trailing closing bracket", operatingHoursSevenClosedDays + `]`},
		{"truncated", operatingHoursSevenClosedDays[:len(operatingHoursSevenClosedDays)-1]},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadOperatingHoursSettings(operatingHoursStoredRow(tc.weekly, `[]`))
			requireCorrupt(t, err)
		})
	}
}

func TestLoadOperatingHoursSettings_RejectsCorruptClosedDates(t *testing.T) {
	cases := []struct {
		name   string
		closed string
	}{
		{"empty raw", ``},
		{"JSON null", `null`},
		{"object instead of array", `{}`},
		{"null element", `[null]`},
		{"unknown key", `[{"date":"2026-11-03","reason":"","extra":1}]`},
		{"missing reason", `[{"date":"2026-11-03"}]`},
		{"missing date", `[{"reason":"x"}]`},
		{"duplicate date key", `[{"date":"2026-11-03","date":"2026-11-03","reason":""}]`},
		{"duplicate reason key", `[{"date":"2026-11-03","reason":"a","reason":"a"}]`},
		{"null reason", `[{"date":"2026-11-03","reason":null}]`},
		{"number reason", `[{"date":"2026-11-03","reason":1}]`},
		{"null date", `[{"date":null,"reason":""}]`},
		{"non-existent date", `[{"date":"2026-02-30","reason":""}]`},
		{"non-padded date", `[{"date":"2026-2-3","reason":""}]`},
		{"timestamp instead of date", `[{"date":"2026-11-03T00:00:00+09:00","reason":""}]`},
		{"unsorted dates", `[{"date":"2026-12-01","reason":""},{"date":"2026-11-03","reason":""}]`},
		{"duplicate dates", `[{"date":"2026-11-03","reason":"a"},{"date":"2026-11-03","reason":"b"}]`},
		{"reason over 500 characters", `[{"date":"2026-11-03","reason":"` + strings.Repeat("あ", 501) + `"}]`},
		{"trailing second value", `[] []`},
		{"trailing garbage", `[]x`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadOperatingHoursSettings(operatingHoursStoredRow(operatingHoursSevenClosedDays, tc.closed))
			requireCorrupt(t, err)
		})
	}
}

func TestLoadOperatingHoursSettings_RejectsCorruptRowMetadata(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*model.ClinicOperatingHours)
	}{
		{"unknown schema_version", func(r *model.ClinicOperatingHours) { r.SchemaVersion = "clinic-operating-hours/v2" }},
		{"empty schema_version", func(r *model.ClinicOperatingHours) { r.SchemaVersion = "" }},
		{"schema_version with different case", func(r *model.ClinicOperatingHours) { r.SchemaVersion = "Clinic-Operating-Hours/v1" }},
		{"version zero", func(r *model.ClinicOperatingHours) { r.Version = 0 }},
		{"negative version", func(r *model.ClinicOperatingHours) { r.Version = -1 }},
		{"zero clinic_id", func(r *model.ClinicOperatingHours) { r.ClinicID = 0 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			row := operatingHoursStoredRow(operatingHoursSevenClosedDays, `[]`)
			tc.mutate(&row)
			_, err := LoadOperatingHoursSettings(row)
			requireCorrupt(t, err)
		})
	}
}

func TestLoadOperatingHoursSettings_RejectsExcessiveNesting(t *testing.T) {
	deep := strings.Repeat("[", storedJSONMaxDepth+2) + strings.Repeat("]", storedJSONMaxDepth+2)
	_, err := LoadOperatingHoursSettings(operatingHoursStoredRow(deep, `[]`))
	requireCorrupt(t, err)
}
