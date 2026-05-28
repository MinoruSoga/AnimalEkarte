package handler

import (
	"testing"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func TestListCashRegisterClosesQuery_ToServiceFilters(t *testing.T) {
	filters, err := (listCashRegisterClosesQuery{
		StartDate: "2026-05-01",
		EndDate:   "2026-05-31",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.StartDate == nil || filters.StartDate.Format(cashRegisterDateLayout) != "2026-05-01" {
		t.Fatalf("StartDate = %v, want 2026-05-01", filters.StartDate)
	}
	if filters.EndDate == nil || filters.EndDate.Format(cashRegisterDateLayout) != "2026-05-31" {
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

	if got := input.Date.Format(cashRegisterDateLayout); got != req.Date {
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

	if _, err := req.toServiceInput(9); err == nil {
		t.Fatal("toServiceInput() error = nil, want error")
	}
}
