package handler

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
