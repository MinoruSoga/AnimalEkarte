package medicalrecord

import (
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestListVaccinationQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&listVaccinationQuery{
		PetID:     "10",
		OwnerID:   "20",
		StartDate: "2026-05-01",
		EndDate:   "2026-05-31",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.PetID == nil || *filters.PetID != 10 {
		t.Fatalf("PetID = %v, want 10", filters.PetID)
	}
	if filters.OwnerID == nil || *filters.OwnerID != 20 {
		t.Fatalf("OwnerID = %v, want 20", filters.OwnerID)
	}
	if filters.StartDate == nil || *filters.StartDate != "2026-05-01" {
		t.Fatalf("StartDate = %v, want 2026-05-01", filters.StartDate)
	}
	if filters.EndDate == nil || *filters.EndDate != "2026-05-31" {
		t.Fatalf("EndDate = %v, want 2026-05-31", filters.EndDate)
	}
}

func TestListVaccinationQuery_ToServiceFilters_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		query listVaccinationQuery
	}{
		{name: "pet_id", query: listVaccinationQuery{PetID: "abc"}},
		{name: "owner_id", query: listVaccinationQuery{OwnerID: "abc"}},
		{name: "start_date", query: listVaccinationQuery{StartDate: "2026/05/01"}},
		{name: "end_date", query: listVaccinationQuery{EndDate: "2026/05/31"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := tt.query.toServiceFilters()
			if err == nil {
				t.Fatal("toServiceFilters returned nil error")
			}
			if filters != (listVaccinationFilters{}) {
				t.Fatalf("filters = %#v, want zero value", filters)
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}

func TestCreateVaccinationRequest_ToServiceInput(t *testing.T) {
	date := "2026-05-28"
	nextDate := "2027-05-28"
	medicalRecordID := uint64(10)
	petID := uint64(20)
	doctorID := uint64(30)
	req := createVaccinationRequest{
		MedicalRecordID:  &medicalRecordID,
		PetID:            &petID,
		VaccineID:        40,
		Date:             &date,
		DoctorID:         &doctorID,
		NextDate:         &nextDate,
		NextScheduleType: string(model.NextScheduleType1Year),
		Supplemental:     "supplemental",
		Lot1:             "lot1",
		Lot2:             "lot2",
		Lot3:             "lot3",
		Lot4:             "lot4",
		Remarks:          "remarks",
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.MedicalRecordID == nil || *input.MedicalRecordID != medicalRecordID {
		t.Fatalf("MedicalRecordID = %v, want %d", input.MedicalRecordID, medicalRecordID)
	}
	if input.PetID == nil || *input.PetID != petID {
		t.Fatalf("PetID = %v, want %d", input.PetID, petID)
	}
	if input.VaccineID != req.VaccineID {
		t.Fatalf("VaccineID = %d, want %d", input.VaccineID, req.VaccineID)
	}
	if got := input.Date.Format("2006-01-02"); got != date {
		t.Fatalf("Date = %q, want %q", got, date)
	}
	if input.DoctorID == nil || *input.DoctorID != doctorID {
		t.Fatalf("DoctorID = %v, want %d", input.DoctorID, doctorID)
	}
	if input.NextDate == nil || input.NextDate.Format("2006-01-02") != nextDate {
		t.Fatalf("NextDate = %v, want %q", input.NextDate, nextDate)
	}
	if input.NextScheduleType == nil || *input.NextScheduleType != model.NextScheduleType1Year {
		t.Fatalf("NextScheduleType = %v, want %q", input.NextScheduleType, model.NextScheduleType1Year)
	}
	if input.Remarks != req.Remarks {
		t.Fatalf("Remarks = %q, want %q", input.Remarks, req.Remarks)
	}
}

func TestCreateVaccinationRequest_ToServiceInput_DateRequired(t *testing.T) {
	req := createVaccinationRequest{VaccineID: 1}

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

func TestVaccinationRequest_ToServiceInput_InvalidDate(t *testing.T) {
	validDate := "2026-05-28"
	invalidDate := "2027/05/28"

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "create/date",
			call: func() error {
				_, err := (&createVaccinationRequest{VaccineID: 1, Date: &invalidDate}).toServiceInput()
				return err
			},
		},
		{
			name: "create/next_date",
			call: func() error {
				_, err := (&createVaccinationRequest{VaccineID: 1, Date: &validDate, NextDate: &invalidDate}).toServiceInput()
				return err
			},
		},
		{
			name: "update/date",
			call: func() error {
				_, err := (&updateVaccinationRequest{Date: &invalidDate}).toServiceInput()
				return err
			},
		},
		{
			name: "update/next_date",
			call: func() error {
				_, err := (&updateVaccinationRequest{NextDate: &invalidDate}).toServiceInput()
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if err == nil {
				t.Fatal("toServiceInput returned nil error")
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
			if strings.Contains(err.Error(), invalidDate) {
				t.Fatalf("error = %v, must not leak raw input %q", err, invalidDate)
			}
		})
	}
}

func TestUpdateVaccinationRequest_ToServiceInput(t *testing.T) {
	date := "2026-05-29"
	nextScheduleType := string(model.NextScheduleTypeOther)
	remarks := ""
	vaccineID := uint64(2)
	req := updateVaccinationRequest{
		VaccineID:        &vaccineID,
		Date:             &date,
		NextScheduleType: &nextScheduleType,
		Remarks:          &remarks,
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.VaccineID == nil || *input.VaccineID != vaccineID {
		t.Fatalf("VaccineID = %v, want %d", input.VaccineID, vaccineID)
	}
	if input.Date == nil || input.Date.Format("2006-01-02") != date {
		t.Fatalf("Date = %v, want %q", input.Date, date)
	}
	if input.NextScheduleType == nil || *input.NextScheduleType != model.NextScheduleTypeOther {
		t.Fatalf("NextScheduleType = %v, want %q", input.NextScheduleType, model.NextScheduleTypeOther)
	}
	if input.Remarks == nil || *input.Remarks != remarks {
		t.Fatalf("Remarks = %v, want explicit empty string", input.Remarks)
	}
}
