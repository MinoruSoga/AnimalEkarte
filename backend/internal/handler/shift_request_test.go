package handler

import (
	"testing"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func TestListShiftEntriesQuery_ToServiceFilters(t *testing.T) {
	filters, err := (listShiftEntriesQuery{
		Date:    "2026-05",
		StaffID: "10",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.YearMonth != "2026-05" {
		t.Fatalf("YearMonth = %q, want 2026-05", filters.YearMonth)
	}
	if filters.StaffID == nil || *filters.StaffID != 10 {
		t.Fatalf("StaffID = %v, want 10", filters.StaffID)
	}
}

func TestListShiftEntriesQuery_ToServiceFilters_InvalidStaffID(t *testing.T) {
	filters, err := (listShiftEntriesQuery{StaffID: "abc"}).toServiceFilters()
	if err == nil {
		t.Fatal("toServiceFilters returned nil error")
	}
	if filters != (listShiftEntriesFilters{}) {
		t.Fatalf("filters = %#v, want zero value", filters)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestCreateShiftRequest_ToServiceInput(t *testing.T) {
	req := createShiftRequest{
		StaffID:   10,
		Date:      "2026-05-28",
		ShiftType: "full",
		StartTime: "09:00",
		EndTime:   "18:00",
		Notes:     "notes",
		Breaks: []shiftBreakRequest{
			{BreakStart: "12:00", BreakEnd: "13:00"},
		},
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.StaffID != req.StaffID {
		t.Errorf("StaffID = %d, want %d", input.StaffID, req.StaffID)
	}
	if got := input.Date.Format(shiftDateLayout); got != req.Date {
		t.Errorf("Date = %q, want %q", got, req.Date)
	}
	if input.StartTime == nil || *input.StartTime != req.StartTime {
		t.Errorf("StartTime = %v, want %q", input.StartTime, req.StartTime)
	}
	if len(input.Breaks) != 1 || input.Breaks[0].BreakEnd != "13:00" {
		t.Errorf("Breaks = %+v, want one break ending 13:00", input.Breaks)
	}
}

func TestCreateShiftRequest_ToServiceInput_InvalidDate(t *testing.T) {
	req := createShiftRequest{Date: "2026/05/28"}

	if _, err := req.toServiceInput(); err == nil {
		t.Fatal("toServiceInput() error = nil, want error")
	}
}

func TestUpdateShiftRequest_ToServiceInput(t *testing.T) {
	shiftType := "off"
	startTime := ""
	notes := ""
	breaksReq := []shiftBreakRequest{{BreakStart: "10:00", BreakEnd: "10:15"}}
	req := updateShiftRequest{
		ShiftType: &shiftType,
		StartTime: &startTime,
		Notes:     &notes,
		Breaks:    &breaksReq,
	}

	input := req.toServiceInput()

	if input.ShiftType == nil || *input.ShiftType != shiftType {
		t.Errorf("ShiftType = %v, want %q", input.ShiftType, shiftType)
	}
	if input.StartTime == nil || *input.StartTime != startTime {
		t.Errorf("StartTime = %v, want empty string pointer", input.StartTime)
	}
	if input.Notes == nil || *input.Notes != notes {
		t.Errorf("Notes = %v, want empty string pointer", input.Notes)
	}
	if input.Breaks == nil || len(*input.Breaks) != 1 {
		t.Fatalf("Breaks = %v, want one item", input.Breaks)
	}
	if (*input.Breaks)[0].BreakStart != "10:00" {
		t.Errorf("BreakStart = %q, want 10:00", (*input.Breaks)[0].BreakStart)
	}
}
