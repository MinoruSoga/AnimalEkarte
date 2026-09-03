package billing

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestNewListCashRegisterClosesQuery(t *testing.T) {
	query := newListCashRegisterClosesQuery(url.Values{
		"start_date": {"2026-05-01"},
		"end_date":   {"2026-05-31"},
	})

	if query.StartDate != "2026-05-01" {
		t.Errorf("StartDate = %q, want 2026-05-01", query.StartDate)
	}
	if query.EndDate != "2026-05-31" {
		t.Errorf("EndDate = %q, want 2026-05-31", query.EndDate)
	}
}

func TestNewListCashRegisterClosesQuery_Empty(t *testing.T) {
	query := newListCashRegisterClosesQuery(url.Values{})

	if query.StartDate != "" {
		t.Errorf("StartDate = %q, want empty", query.StartDate)
	}
	if query.EndDate != "" {
		t.Errorf("EndDate = %q, want empty", query.EndDate)
	}
}

func TestNewCashRegisterPreviewQuery(t *testing.T) {
	query := newCashRegisterPreviewQuery(url.Values{
		"date":   {"2026-05-28"},
		"period": {"am"},
	})

	if query.Date != "2026-05-28" {
		t.Errorf("Date = %q, want 2026-05-28", query.Date)
	}
	if query.Period != "am" {
		t.Errorf("Period = %q, want am", query.Period)
	}
}

func TestNewCashRegisterPreviewQuery_Empty(t *testing.T) {
	query := newCashRegisterPreviewQuery(url.Values{})

	if query.Date != "" {
		t.Errorf("Date = %q, want empty", query.Date)
	}
	if query.Period != "" {
		t.Errorf("Period = %q, want empty", query.Period)
	}
}

func TestListCashRegisterClosesQuery_ToServiceFilters(t *testing.T) {
	filters, err := (listCashRegisterClosesQuery{
		StartDate: "2026-05-01",
		EndDate:   "2026-05-31",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.StartDate == nil || filters.StartDate.Format(time.DateOnly) != "2026-05-01" {
		t.Fatalf("StartDate = %v, want 2026-05-01", filters.StartDate)
	}
	if filters.EndDate == nil || filters.EndDate.Format(time.DateOnly) != "2026-05-31" {
		t.Fatalf("EndDate = %v, want 2026-05-31", filters.EndDate)
	}
}

func TestListCashRegisterClosesQuery_ToServiceFilters_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		query listCashRegisterClosesQuery
	}{
		{name: "start_date", query: listCashRegisterClosesQuery{StartDate: "2026/05/01"}},
		{name: "end_date", query: listCashRegisterClosesQuery{EndDate: "2026/05/31"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := tt.query.toServiceFilters()
			if err == nil {
				t.Fatal("toServiceFilters returned nil error")
			}
			if filters != (listCashRegisterClosesFilters{}) {
				t.Fatalf("filters = %#v, want zero value", filters)
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}

func TestCloseCashRegisterRequest_ToServiceInput(t *testing.T) {
	req := closeCashRegisterRequest{
		Date:       "2026-05-28",
		Period:     "am",
		ActualCash: 12000,
		Memo:       "memo",
	}

	input, err := req.toServiceInput(9)
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if got := input.Date.Format(time.DateOnly); got != req.Date {
		t.Errorf("Date = %q, want %q", got, req.Date)
	}
	if input.Period != req.Period {
		t.Errorf("Period = %q, want %q", input.Period, req.Period)
	}
	if input.ClosedBy == nil || *input.ClosedBy != 9 {
		t.Errorf("ClosedBy = %v, want 9", input.ClosedBy)
	}
}

func TestCloseCashRegisterRequest_ToServiceInput_InvalidDate(t *testing.T) {
	req := closeCashRegisterRequest{Date: "2026/05/28"}

	_, err := req.toServiceInput(9)
	if err == nil {
		t.Fatal("toServiceInput() error = nil, want error")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
	if strings.Contains(err.Error(), "parsing time") {
		t.Fatalf("error leaked parse internals: %v", err)
	}
}

func TestCloseCashRegisterRequest_ToServiceInput_NegativeActualCash(t *testing.T) {
	req := closeCashRegisterRequest{
		Date:       "2026-05-28",
		Period:     "am",
		ActualCash: -1,
	}

	_, err := req.toServiceInput(9)
	if err == nil {
		t.Fatal("toServiceInput() error = nil, want error for negative actual_cash")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
	if got := err.Error(); !strings.Contains(got, "actual_cash は 0 以上で指定してください") {
		t.Fatalf("error = %q, want actual_cash は 0 以上で指定してください", got)
	}
}

func TestCloseCashRegisterRequest_ToServiceInput_ZeroActualCashAllowed(t *testing.T) {
	req := closeCashRegisterRequest{
		Date:       "2026-05-28",
		Period:     "am",
		ActualCash: 0,
	}

	input, err := req.toServiceInput(9)
	if err != nil {
		t.Fatalf("toServiceInput() error = %v, want nil for actual_cash=0", err)
	}
	if input.ActualCash != 0 {
		t.Fatalf("ActualCash = %d, want 0", input.ActualCash)
	}
}
