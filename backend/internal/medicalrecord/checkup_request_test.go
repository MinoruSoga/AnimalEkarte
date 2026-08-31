package medicalrecord

import (
	"net/url"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestCreateCheckupRequest_ToServiceInput(t *testing.T) {
	nextDate := "2027-05-28"
	petID := uint64(10)
	doctorID := uint64(20)
	req := createCheckupRequest{
		CheckupTypeID: 30,
		PetID:         &petID,
		Date:          "2026-05-28",
		NextDate:      &nextDate,
		DoctorID:      &doctorID,
		Result:        "normal",
	}

	input, err := req.toServiceInput(1)
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.ClinicID != 1 {
		t.Fatalf("ClinicID = %d, want 1", input.ClinicID)
	}
	if input.CheckupTypeID != req.CheckupTypeID {
		t.Fatalf("CheckupTypeID = %d, want %d", input.CheckupTypeID, req.CheckupTypeID)
	}
	if input.PetID == nil || *input.PetID != petID {
		t.Fatalf("PetID = %v, want %d", input.PetID, petID)
	}
	if got := input.Date.Format("2006-01-02"); got != req.Date {
		t.Fatalf("Date = %q, want %q", got, req.Date)
	}
	if input.NextDate == nil || input.NextDate.Format("2006-01-02") != nextDate {
		t.Fatalf("NextDate = %v, want %q", input.NextDate, nextDate)
	}
	if input.DoctorID == nil || *input.DoctorID != doctorID {
		t.Fatalf("DoctorID = %v, want %d", input.DoctorID, doctorID)
	}
	if input.Result != req.Result {
		t.Fatalf("Result = %q, want %q", input.Result, req.Result)
	}
}

func TestCreateCheckupRequest_ToServiceInput_InvalidDate(t *testing.T) {
	req := createCheckupRequest{
		CheckupTypeID: 1,
		Date:          "2026/05/28",
	}

	input, err := req.toServiceInput(1)
	if err == nil {
		t.Fatal("toServiceInput returned nil error")
	}
	if input != nil {
		t.Fatalf("input = %#v, want nil", input)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestCreateCheckupRequest_ToServiceInput_InvalidNextDate(t *testing.T) {
	nextDate := "2027/05/28"
	req := createCheckupRequest{
		CheckupTypeID: 1,
		Date:          "2026-05-28",
		NextDate:      &nextDate,
	}

	input, err := req.toServiceInput(1)
	if err == nil {
		t.Fatal("toServiceInput returned nil error")
	}
	if input != nil {
		t.Fatalf("input = %#v, want nil", input)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestUpdateCheckupRequest_ToServiceInput(t *testing.T) {
	date := "2026-05-29"
	nextDate := "2027-05-29"
	result := ""
	doctorIDClear := true
	req := updateCheckupRequest{
		Date:          &date,
		NextDate:      &nextDate,
		DoctorIDClear: &doctorIDClear,
		Result:        &result,
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.Date == nil || input.Date.Format("2006-01-02") != date {
		t.Fatalf("Date = %v, want %q", input.Date, date)
	}
	if input.NextDate == nil || input.NextDate.Format("2006-01-02") != nextDate {
		t.Fatalf("NextDate = %v, want %q", input.NextDate, nextDate)
	}
	if input.DoctorIDClear == nil || !*input.DoctorIDClear {
		t.Fatalf("DoctorIDClear = %v, want true", input.DoctorIDClear)
	}
	if input.Result == nil || *input.Result != result {
		t.Fatalf("Result = %v, want explicit empty string", input.Result)
	}
}

func TestUpdateCheckupRequest_ToServiceInput_EmptyDates(t *testing.T) {
	date := ""
	nextDate := ""
	req := updateCheckupRequest{
		Date:     &date,
		NextDate: &nextDate,
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.Date != nil {
		t.Fatalf("Date = %v, want nil", input.Date)
	}
	if input.NextDate != nil {
		t.Fatalf("NextDate = %v, want nil", input.NextDate)
	}
}

func TestUpdateCheckupRequest_ToServiceInput_InvalidDate(t *testing.T) {
	date := "May 29, 2026"
	req := updateCheckupRequest{Date: &date}

	input, err := req.toServiceInput()
	if err == nil {
		t.Fatal("toServiceInput returned nil error")
	}
	if input != nil {
		t.Fatalf("input = %#v, want nil", input)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestListGlobalCheckupsQuery_ToServiceInput(t *testing.T) {
	startDate := "2026-01-01"
	nextEndDate := "2026-12-31"
	query := listGlobalCheckupsQuery{
		ClinicID:    1,
		StartDate:   &startDate,
		NextEndDate: &nextEndDate,
	}

	input := query.toServiceInput()

	if input.ClinicID != query.ClinicID {
		t.Errorf("ClinicID = %d, want %d", input.ClinicID, query.ClinicID)
	}
	if input.StartDate == nil || *input.StartDate != startDate {
		t.Errorf("StartDate = %v, want %q", input.StartDate, startDate)
	}
	if input.NextEndDate == nil || *input.NextEndDate != nextEndDate {
		t.Errorf("NextEndDate = %v, want %q", input.NextEndDate, nextEndDate)
	}
}

func TestNewListGlobalCheckupsQuery(t *testing.T) {
	query := newListGlobalCheckupsQuery(1, url.Values{
		"pet_id":          {"42"},
		"start_date":      {"2026-01-01"},
		"end_date":        {"2026-01-31"},
		"next_start_date": {"2026-12-01"},
		"next_end_date":   {"2026-12-31"},
	})

	if query.PetID == nil || *query.PetID != 42 {
		t.Errorf("PetID = %v, want 42", query.PetID)
	}
	if query.ClinicID != 1 {
		t.Errorf("ClinicID = %d, want 1", query.ClinicID)
	}
	if query.StartDate == nil || *query.StartDate != "2026-01-01" {
		t.Errorf("StartDate = %v, want 2026-01-01", query.StartDate)
	}
	if query.EndDate == nil || *query.EndDate != "2026-01-31" {
		t.Errorf("EndDate = %v, want 2026-01-31", query.EndDate)
	}
	if query.NextStartDate == nil || *query.NextStartDate != "2026-12-01" {
		t.Errorf("NextStartDate = %v, want 2026-12-01", query.NextStartDate)
	}
	if query.NextEndDate == nil || *query.NextEndDate != "2026-12-31" {
		t.Errorf("NextEndDate = %v, want 2026-12-31", query.NextEndDate)
	}
}
