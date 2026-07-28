package reservation

import (
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestListReservationQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&listReservationQuery{
		Date:    "2026-05-28",
		Status:  "confirmed",
		PetID:   "12",
		OwnerID: "34",
		Source:  "manual",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters() error = %v", err)
	}

	if filters.Date == nil || filters.Date.Year() != 2026 || filters.Date.Month() != time.May || filters.Date.Day() != 28 {
		t.Fatalf("Date = %v, want 2026-05-28", filters.Date)
	}
	if filters.Status == nil || *filters.Status != "confirmed" {
		t.Fatalf("Status = %v, want confirmed", filters.Status)
	}
	if filters.PetID == nil || *filters.PetID != 12 {
		t.Fatalf("PetID = %v, want 12", filters.PetID)
	}
	if filters.OwnerID == nil || *filters.OwnerID != 34 {
		t.Fatalf("OwnerID = %v, want 34", filters.OwnerID)
	}
	if filters.Source == nil || *filters.Source != "manual" {
		t.Fatalf("Source = %v, want manual", filters.Source)
	}
}

func TestListReservationQuery_ToServiceFilters_Empty(t *testing.T) {
	filters, err := (&listReservationQuery{}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters() error = %v", err)
	}

	if filters.Date != nil || filters.Status != nil || filters.PetID != nil || filters.OwnerID != nil || filters.Source != nil {
		t.Fatalf("expected nil filters, got %#v", filters)
	}
}

func TestListReservationQuery_ToServiceFilters_InvalidDate(t *testing.T) {
	_, err := (&listReservationQuery{Date: "2026/05/28"}).toServiceFilters()
	if err == nil {
		t.Fatalf("toServiceFilters() error = nil, want error")
	}
}

func TestListReservationQuery_ToServiceFilters_InvalidIDs(t *testing.T) {
	if _, err := (&listReservationQuery{PetID: "abc"}).toServiceFilters(); err == nil {
		t.Fatalf("PetID error = nil, want error")
	}
	if _, err := (&listReservationQuery{OwnerID: "abc"}).toServiceFilters(); err == nil {
		t.Fatalf("OwnerID error = nil, want error")
	}
}

func TestUpdateReservationRequest_ToServiceInput(t *testing.T) {
	startTime := time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)
	ownerID := uint64(1)
	petID := uint64(2)
	visitType := string(model.VisitTypeFirst)
	reservationTypeID := uint64(3)
	doctorID := uint64(4)
	isDesignated := true
	status := string(model.ReservationStatusCheckedIn)
	notes := "follow up"

	input, err := (&updateReservationRequest{
		StartTime:         &startTime,
		EndTime:           &endTime,
		OwnerID:           &ownerID,
		PetID:             &petID,
		VisitType:         &visitType,
		ReservationTypeID: &reservationTypeID,
		DoctorID:          &doctorID,
		IsDesignated:      &isDesignated,
		Status:            &status,
		Notes:             &notes,
	}).toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.StartTime != &startTime {
		t.Fatalf("StartTime pointer was not preserved")
	}
	if input.EndTime != &endTime {
		t.Fatalf("EndTime pointer was not preserved")
	}
	if input.OwnerID != &ownerID {
		t.Fatalf("OwnerID pointer was not preserved")
	}
	if input.PetID != &petID {
		t.Fatalf("PetID pointer was not preserved")
	}
	if input.VisitType == nil || *input.VisitType != model.VisitTypeFirst {
		t.Fatalf("VisitType = %v, want %s", input.VisitType, model.VisitTypeFirst)
	}
	if input.ReservationTypeID != &reservationTypeID {
		t.Fatalf("ReservationTypeID pointer was not preserved")
	}
	if input.DoctorID != &doctorID {
		t.Fatalf("DoctorID pointer was not preserved")
	}
	if input.IsDesignated != &isDesignated {
		t.Fatalf("IsDesignated pointer was not preserved")
	}
	if input.Status == nil || *input.Status != model.ReservationStatusCheckedIn {
		t.Fatalf("Status = %v, want %s", input.Status, model.ReservationStatusCheckedIn)
	}
	if input.Notes != &notes {
		t.Fatalf("Notes pointer was not preserved")
	}
}

func TestCreateReservationRequest_ToServiceInput(t *testing.T) {
	startTime := time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)
	ownerID := uint64(1)
	petID := uint64(2)
	doctorID := uint64(4)

	input, err := (&createReservationRequest{
		StartTime:         startTime,
		EndTime:           endTime,
		OwnerID:           &ownerID,
		PetID:             &petID,
		VisitType:         string(model.VisitTypeFirst),
		ReservationTypeID: 3,
		DoctorID:          &doctorID,
		IsDesignated:      true,
		Status:            string(model.ReservationStatusConfirmed),
		Notes:             "manual reservation",
		Source:            string(model.ReservationSourceLine),
		ReservationRoute:  "record_shortcut",
	}).toServiceInput(9, 8)
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.ClinicID != 9 {
		t.Fatalf("ClinicID = %d, want 9", input.ClinicID)
	}
	if input.CreatedBy == nil || *input.CreatedBy != 8 {
		t.Fatalf("CreatedBy = %v, want 8", input.CreatedBy)
	}
	if input.VisitType != model.VisitTypeFirst {
		t.Fatalf("VisitType = %s, want %s", input.VisitType, model.VisitTypeFirst)
	}
	if input.Status != model.ReservationStatusConfirmed {
		t.Fatalf("Status = %s, want %s", input.Status, model.ReservationStatusConfirmed)
	}
	if input.Source != model.ReservationSourceLine {
		t.Fatalf("Source = %s, want %s", input.Source, model.ReservationSourceLine)
	}
	if input.DoctorID != &doctorID {
		t.Fatalf("DoctorID pointer was not preserved")
	}
	if input.ReservationRoute == nil || *input.ReservationRoute != "record_shortcut" {
		t.Fatalf("ReservationRoute = %v, want record_shortcut", input.ReservationRoute)
	}
}

func TestCreateReservationRequest_ToServiceInput_DefaultSource(t *testing.T) {
	input, err := (&createReservationRequest{}).toServiceInput(9, 8)
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.Source != model.ReservationSourceManual {
		t.Fatalf("Source = %s, want %s", input.Source, model.ReservationSourceManual)
	}
}

func TestUpdateReservationRequest_ToServiceInput_NilFields(t *testing.T) {
	input, err := (&updateReservationRequest{}).toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.StartTime != nil || input.Status != nil || input.VisitType != nil {
		t.Fatalf("expected nil optional fields, got %#v", input)
	}
}

func TestPatchReservationReservationRouteRequest_ToServiceInput(t *testing.T) {
	input := (&patchReservationReservationRouteRequest{Route: "line"}).toServiceInput()

	if input.Route != "line" {
		t.Fatalf("Route = %q, want line", input.Route)
	}
}
