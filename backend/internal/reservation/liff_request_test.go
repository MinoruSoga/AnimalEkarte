package reservation

import (
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestLiffCourseQuery_ToCourseID(t *testing.T) {
	courseID, err := (liffCourseQuery{CourseID: "10"}).toCourseID()
	if err != nil {
		t.Fatalf("toCourseID() error = %v", err)
	}
	if courseID != 10 {
		t.Fatalf("courseID = %d, want 10", courseID)
	}
}

func TestLiffCourseQuery_ToCourseID_InvalidInput(t *testing.T) {
	tests := []liffCourseQuery{
		{},
		{CourseID: "0"},
		{CourseID: "abc"},
	}

	for _, tt := range tests {
		_, err := tt.toCourseID()
		if err == nil {
			t.Fatalf("toCourseID() error = nil for %#v", tt)
		}
		if !apperrors.IsInvalidInput(err) {
			t.Fatalf("error = %v, want invalid input", err)
		}
	}
}

func TestLiffAvailableDatesQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&liffAvailableDatesQuery{
		CourseID: "10",
		StaffID:  "20",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters() error = %v", err)
	}
	if filters.CourseID != 10 || filters.StaffID != 20 {
		t.Fatalf("filters = %#v, want courseID=10 staffID=20", filters)
	}
}

func TestLiffAvailableDatesQuery_ToServiceFilters_InvalidStaffIDAsNoPreference(t *testing.T) {
	filters, err := (&liffAvailableDatesQuery{
		CourseID: "10",
		StaffID:  "abc",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters() error = %v", err)
	}
	if filters.StaffID != 0 {
		t.Fatalf("StaffID = %d, want 0", filters.StaffID)
	}
}

func TestLiffAvailableTimesQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&liffAvailableTimesQuery{
		CourseID: "10",
		StaffID:  "20",
		Date:     "2026-05-28",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters() error = %v", err)
	}
	if filters.CourseID != 10 || filters.StaffID != 20 {
		t.Fatalf("filters = %#v, want courseID=10 staffID=20", filters)
	}
	if got := filters.Date.Format("2006-01-02"); got != "2026-05-28" {
		t.Fatalf("Date = %q, want 2026-05-28", got)
	}
}

func TestLiffAvailableTimesQuery_ToServiceFilters_InvalidDate(t *testing.T) {
	filters, err := (&liffAvailableTimesQuery{
		CourseID: "10",
		Date:     "2026/05/28",
	}).toServiceFilters()
	if err == nil {
		t.Fatal("toServiceFilters() error = nil")
	}
	if filters != (liffAvailableTimesFilters{}) {
		t.Fatalf("filters = %#v, want zero value", filters)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestLiffCreateReservationRequest_ToServiceInput(t *testing.T) {
	trimmingCourseID := uint64(10)

	input, err := (&liffCreateReservationRequest{
		TypeID:               1,
		StaffID:              2,
		Date:                 "2026-05-28",
		StartTime:            "0900",
		EndTime:              "0930",
		CustomerFields:       jsonRawOrEmpty(`{"pet":"momo"}`),
		RequestText:          "please call",
		TrimmingCourseID:     &trimmingCourseID,
		TrimmingOptionIDs:    []uint64{3, 4},
		TrimmingStyleRequest: "short",
	}).toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	wantDate := time.Date(2026, 5, 28, 0, 0, 0, 0, time.Local)
	if !input.Date.Equal(wantDate) {
		t.Fatalf("Date = %v, want %v", input.Date, wantDate)
	}
	if input.ReservationTypeID != 1 || input.StaffID != 2 {
		t.Fatalf("type/staff IDs = %d/%d, want 1/2", input.ReservationTypeID, input.StaffID)
	}
	if input.TrimmingCourseID != &trimmingCourseID {
		t.Fatalf("TrimmingCourseID pointer was not preserved")
	}
	if len(input.TrimmingOptionIDs) != 2 || input.TrimmingOptionIDs[1] != 4 {
		t.Fatalf("TrimmingOptionIDs = %#v, want [3 4]", input.TrimmingOptionIDs)
	}
}

func TestLiffCreateReservationRequest_ToServiceInput_InvalidDate(t *testing.T) {
	_, err := (&liffCreateReservationRequest{Date: "2026/05/28"}).toServiceInput()
	if err == nil {
		t.Fatalf("toServiceInput() error = nil, want error")
	}
}
