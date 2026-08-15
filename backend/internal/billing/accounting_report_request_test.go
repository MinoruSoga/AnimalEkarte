package billing

import (
	"net/url"
	"testing"
	"time"
)

func TestMonthlyReportQuery_ToYearMonth(t *testing.T) {
	now := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)

	t.Run("explicit year and month", func(t *testing.T) {
		year, month, err := newMonthlyReportQuery(url.Values{
			"year":  {"2026"},
			"month": {"4"},
		}).toYearMonth(now)
		if err != nil {
			t.Fatalf("toYearMonth returned error: %v", err)
		}
		if year != 2026 {
			t.Errorf("year = %d, want 2026", year)
		}
		if month != 4 {
			t.Errorf("month = %d, want 4", month)
		}
	})

	t.Run("defaults to current year and month", func(t *testing.T) {
		year, month, err := newMonthlyReportQuery(url.Values{}).toYearMonth(now)
		if err != nil {
			t.Fatalf("toYearMonth returned error: %v", err)
		}
		if year != 2026 {
			t.Errorf("year = %d, want 2026", year)
		}
		if month != 5 {
			t.Errorf("month = %d, want 5", month)
		}
	})
}

func TestMonthlyReportQuery_ToYearMonth_Invalid(t *testing.T) {
	now := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name  string
		query monthlyReportQuery
	}{
		{"year below minimum", monthlyReportQuery{Year: "1999", Month: "1"}},
		{"year above maximum", monthlyReportQuery{Year: "2101", Month: "1"}},
		{"year non numeric", monthlyReportQuery{Year: "abc", Month: "1"}},
		{"month below minimum", monthlyReportQuery{Year: "2026", Month: "0"}},
		{"month above maximum", monthlyReportQuery{Year: "2026", Month: "13"}},
		{"month non numeric", monthlyReportQuery{Year: "2026", Month: "abc"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := tc.query.toYearMonth(now); err == nil {
				t.Fatal("toYearMonth returned nil error")
			}
		})
	}
}

func TestMonthlyReportQuery_HasPeriod(t *testing.T) {
	cases := []struct {
		name  string
		query monthlyReportQuery
		want  bool
	}{
		{"both start and end set", monthlyReportQuery{StartDate: "2026-04-01", EndDate: "2026-04-30"}, true},
		{"only start set", monthlyReportQuery{StartDate: "2026-04-01"}, true},
		{"only end set", monthlyReportQuery{EndDate: "2026-04-30"}, true},
		{"neither set", monthlyReportQuery{}, false},
		{"year/month set but no period", monthlyReportQuery{Year: "2026", Month: "4"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.query.hasPeriod(); got != tc.want {
				t.Errorf("hasPeriod() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMonthlyReportQuery_ToPeriod(t *testing.T) {
	start, end, err := newMonthlyReportQuery(url.Values{
		"start_date": {"2026-04-30"},
		"end_date":   {"2026-05-02"},
	}).toPeriod()
	if err != nil {
		t.Fatalf("toPeriod returned error: %v", err)
	}
	if start.Format("2006-01-02") != "2026-04-30" {
		t.Errorf("start = %s, want 2026-04-30", start.Format("2006-01-02"))
	}
	if end.Format("2006-01-02") != "2026-05-02" {
		t.Errorf("end = %s, want 2026-05-02", end.Format("2006-01-02"))
	}
}

func TestMonthlyReportQuery_ToPeriod_Invalid(t *testing.T) {
	cases := []struct {
		name  string
		query monthlyReportQuery
	}{
		{"missing start", monthlyReportQuery{EndDate: "2026-05-02"}},
		{"missing end", monthlyReportQuery{StartDate: "2026-04-30"}},
		{"invalid start", monthlyReportQuery{StartDate: "2026/04/30", EndDate: "2026-05-02"}},
		{"invalid end", monthlyReportQuery{StartDate: "2026-04-30", EndDate: "2026/05/02"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := tc.query.toPeriod(); err == nil {
				t.Fatal("toPeriod returned nil error")
			}
		})
	}
}
