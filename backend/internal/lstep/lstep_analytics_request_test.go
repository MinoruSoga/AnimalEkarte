package lstep

import (
	"net/url"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestNewLstepMonthlyDeliveryStatsQuery(t *testing.T) {
	query, err := newLstepMonthlyDeliveryStatsQuery(url.Values{"year_month": {"2026-05"}})
	if err != nil {
		t.Fatalf("newLstepMonthlyDeliveryStatsQuery returned error: %v", err)
	}
	if query.YearMonth != "2026-05" {
		t.Errorf("YearMonth = %q, want 2026-05", query.YearMonth)
	}
}

func TestNewLstepMonthlyDeliveryStatsQuery_MissingYearMonth(t *testing.T) {
	query, err := newLstepMonthlyDeliveryStatsQuery(url.Values{})
	if err == nil {
		t.Fatal("newLstepMonthlyDeliveryStatsQuery returned nil error")
	}
	if query.YearMonth != "" {
		t.Errorf("YearMonth = %q, want empty", query.YearMonth)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestNewLstepVisitConversionQuery(t *testing.T) {
	query, err := newLstepVisitConversionQuery(url.Values{
		"year_month": {"2026-05"},
		"days":       {"60"},
	})
	if err != nil {
		t.Fatalf("newLstepVisitConversionQuery returned error: %v", err)
	}
	if query.YearMonth != "2026-05" {
		t.Errorf("YearMonth = %q, want 2026-05", query.YearMonth)
	}
	if query.Days != "60" {
		t.Errorf("Days = %q, want 60", query.Days)
	}
}

func TestNewLstepVisitConversionQuery_MissingYearMonth(t *testing.T) {
	query, err := newLstepVisitConversionQuery(url.Values{})
	if err == nil {
		t.Fatal("newLstepVisitConversionQuery returned nil error")
	}
	if query.YearMonth != "" {
		t.Errorf("YearMonth = %q, want empty", query.YearMonth)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestLstepVisitConversionQuery_ToDays(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		days, err := (lstepVisitConversionQuery{}).toDays()
		if err != nil {
			t.Fatalf("toDays returned error: %v", err)
		}
		if days != 30 {
			t.Errorf("days = %d, want 30", days)
		}
	})

	t.Run("custom", func(t *testing.T) {
		days, err := (lstepVisitConversionQuery{Days: "60"}).toDays()
		if err != nil {
			t.Fatalf("toDays returned error: %v", err)
		}
		if days != 60 {
			t.Errorf("days = %d, want 60", days)
		}
	})
}

func TestLstepVisitConversionQuery_ToDays_Invalid(t *testing.T) {
	cases := []string{"0", "-1", "366", "abc"}
	for _, daysValue := range cases {
		t.Run(daysValue, func(t *testing.T) {
			days, err := (lstepVisitConversionQuery{Days: daysValue}).toDays()
			if err == nil {
				t.Fatal("toDays returned nil error")
			}
			if days != 0 {
				t.Errorf("days = %d, want 0", days)
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}
