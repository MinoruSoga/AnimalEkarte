package trimming

import (
	"bytes"
	"testing"
	"time"

	"github.com/gin-gonic/gin/binding"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestListTrimmingQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&listTrimmingQuery{
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

func TestListTrimmingQuery_ToServiceFilters_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		query listTrimmingQuery
	}{
		{name: "pet_id", query: listTrimmingQuery{PetID: "abc"}},
		{name: "owner_id", query: listTrimmingQuery{OwnerID: "abc"}},
		{name: "start_date", query: listTrimmingQuery{StartDate: "2026/05/01"}},
		{name: "end_date", query: listTrimmingQuery{EndDate: "2026/05/31"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := tt.query.toServiceFilters()
			if err == nil {
				t.Fatal("toServiceFilters returned nil error")
			}
			if filters != (listTrimmingFilters{}) {
				t.Fatalf("filters = %#v, want zero value", filters)
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}

func TestCreateTrimmingRequest_ToServiceInput(t *testing.T) {
	startTime := time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)
	petID := uint64(10)
	staffID := uint64(20)
	bw := 5.6
	req := createTrimmingRequest{
		ReservationTypeID: 3,
		StartTime:         &startTime,
		EndTime:           &endTime,
		PetID:             &petID,
		StaffID:           &staffID,
		Status:            "confirmed",
		ReservationRoute:  "record_shortcut",
		StyleRequest:      "short",
		BW:                &bw,
		BWUnit:            "kg",
		UsedShampoo:       "shampoo",
		OptionIDs:         []uint64{1, 2},
	}

	input := req.toServiceInput()

	if input.StartTime != startTime {
		t.Errorf("StartTime = %v, want %v", input.StartTime, startTime)
	}
	if input.PetID == nil || *input.PetID != petID {
		t.Errorf("PetID = %v, want %d", input.PetID, petID)
	}
	if string(input.Status) != req.Status {
		t.Errorf("Status = %q, want %q", input.Status, req.Status)
	}
	if input.ReservationRoute == nil || *input.ReservationRoute != req.ReservationRoute {
		t.Errorf("ReservationRoute = %v, want %q", input.ReservationRoute, req.ReservationRoute)
	}
	if input.BodyWeight == nil || *input.BodyWeight != bw {
		t.Errorf("BodyWeight = %v, want %v", input.BodyWeight, bw)
	}
	if string(input.BWUnit) != req.BWUnit {
		t.Errorf("BWUnit = %q, want %q", input.BWUnit, req.BWUnit)
	}
	if len(input.OptionIDs) != 2 {
		t.Errorf("len(OptionIDs) = %d, want 2", len(input.OptionIDs))
	}
}

func TestUpdateTrimmingRequest_ToServiceInput(t *testing.T) {
	status := "accounting"
	bwUnit := "kg"
	styleRequest := ""
	optionIDs := []uint64{}
	req := updateTrimmingRequest{
		Status:       &status,
		BWUnit:       &bwUnit,
		StyleRequest: &styleRequest,
		OptionIDs:    &optionIDs,
	}

	input := req.toServiceInput()

	if input.Status == nil || string(*input.Status) != status {
		t.Errorf("Status = %v, want %q", input.Status, status)
	}
	if input.BWUnit == nil || string(*input.BWUnit) != bwUnit {
		t.Errorf("BWUnit = %v, want %q", input.BWUnit, bwUnit)
	}
	if input.StyleRequest == nil || *input.StyleRequest != styleRequest {
		t.Errorf("StyleRequest = %v, want empty string pointer", input.StyleRequest)
	}
	if input.OptionIDs == nil || len(*input.OptionIDs) != 0 {
		t.Errorf("OptionIDs = %v, want empty slice pointer", input.OptionIDs)
	}
}

func TestTrimmingRequest_BindingRejectsDirectCompletedStatus(t *testing.T) {
	completed := "completed"
	accounting := "accounting"
	petID := uint64(1)

	tests := []struct {
		name    string
		request any
		wantErr bool
	}{
		{
			name: "create rejects completed",
			request: &createTrimmingRequest{
				ReservationTypeID: 1,
				PetID:             &petID,
				Status:            completed,
			},
			wantErr: true,
		},
		{
			name: "create accepts accounting",
			request: &createTrimmingRequest{
				ReservationTypeID: 1,
				PetID:             &petID,
				Status:            accounting,
			},
		},
		{
			name:    "update rejects completed",
			request: &updateTrimmingRequest{Status: &completed},
			wantErr: true,
		},
		{
			name:    "update accepts accounting",
			request: &updateTrimmingRequest{Status: &accounting},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := binding.Validator.ValidateStruct(tt.request)
			if tt.wantErr && err == nil {
				t.Fatal("ValidateStruct() = nil, want validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateStruct() = %v, want nil", err)
			}
		})
	}
}

func TestTrimmingRequest_BindingOptionIDsLimit(t *testing.T) {
	fiftyIDs := make([]uint64, 50)
	fiftyOneIDs := make([]uint64, 51)
	petID := uint64(1)

	tests := []struct {
		name    string
		request any
		wantErr bool
	}{
		{
			name: "create accepts 50 options",
			request: &createTrimmingRequest{
				ReservationTypeID: 1,
				PetID:             &petID,
				OptionIDs:         fiftyIDs,
			},
		},
		{
			name: "create rejects 51 options",
			request: &createTrimmingRequest{
				ReservationTypeID: 1,
				PetID:             &petID,
				OptionIDs:         fiftyOneIDs,
			},
			wantErr: true,
		},
		{
			name:    "update accepts 50 options",
			request: &updateTrimmingRequest{OptionIDs: &fiftyIDs},
		},
		{
			name:    "update rejects 51 options",
			request: &updateTrimmingRequest{OptionIDs: &fiftyOneIDs},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := binding.Validator.ValidateStruct(tt.request)
			if tt.wantErr && err == nil {
				t.Fatal("ValidateStruct() = nil, want validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateStruct() = %v, want nil", err)
			}
		})
	}
}

func TestTrimmingRequest_BindingBWUnitAndFreeTextLimits(t *testing.T) {
	petID := uint64(1)
	tooLong := string(bytes.Repeat([]byte("a"), 2001))
	okUnit := "Kg"
	badUnit := "lb"

	tests := []struct {
		name    string
		request any
		wantErr bool
	}{
		{
			name: "create accepts bw_unit Kg",
			request: &createTrimmingRequest{
				ReservationTypeID: 1, PetID: &petID, BWUnit: "Kg",
			},
		},
		{
			name: "create accepts bw_unit g",
			request: &createTrimmingRequest{
				ReservationTypeID: 1, PetID: &petID, BWUnit: "g",
			},
		},
		{
			name: "create rejects invalid bw_unit",
			request: &createTrimmingRequest{
				ReservationTypeID: 1, PetID: &petID, BWUnit: "lb",
			},
			wantErr: true,
		},
		{
			name: "create rejects overlong style_request",
			request: &createTrimmingRequest{
				ReservationTypeID: 1, PetID: &petID, StyleRequest: tooLong,
			},
			wantErr: true,
		},
		{
			name:    "update rejects invalid bw_unit",
			request: &updateTrimmingRequest{BWUnit: &badUnit},
			wantErr: true,
		},
		{
			name:    "update accepts bw_unit Kg",
			request: &updateTrimmingRequest{BWUnit: &okUnit},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := binding.Validator.ValidateStruct(tt.request)
			if tt.wantErr && err == nil {
				t.Fatal("ValidateStruct() = nil, want validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateStruct() = %v, want nil", err)
			}
		})
	}
}
