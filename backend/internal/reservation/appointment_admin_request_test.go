package reservation

import (
	"net/url"
	"testing"
	"time"
)

func TestNewListReservationsAdminQuery(t *testing.T) {
	fixedNow := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		values   url.Values
		now      time.Time
		wantView string
		wantDate string
	}{
		{
			name:     "explicit view and date are preserved",
			values:   url.Values{"view": []string{"day"}, "date": []string{"2026-06-01"}},
			now:      fixedNow,
			wantView: "day",
			wantDate: "2026-06-01",
		},
		{
			name:     "empty view defaults to month",
			values:   url.Values{"date": []string{"2026-06-01"}},
			now:      fixedNow,
			wantView: "month",
			wantDate: "2026-06-01",
		},
		{
			name:     "empty date defaults to now formatted as YYYY-MM",
			values:   url.Values{"view": []string{"month"}},
			now:      fixedNow,
			wantView: "month",
			wantDate: "2026-07",
		},
		{
			name:     "nil values defaults both view and date",
			values:   nil,
			now:      fixedNow,
			wantView: "month",
			wantDate: "2026-07",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := newListReservationsAdminQuery(tt.values, tt.now)
			if q.View != tt.wantView {
				t.Errorf("View = %q, want %q", q.View, tt.wantView)
			}
			if q.Date != tt.wantDate {
				t.Errorf("Date = %q, want %q", q.Date, tt.wantDate)
			}
		})
	}
}

func TestCreateReservationAdminRequest_ToServiceInput(t *testing.T) {
	startTime := time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)
	ownerID := uint64(1)
	doctorID := uint64(2)
	req := createReservationAdminRequest{
		StartTime:         startTime,
		EndTime:           endTime,
		OwnerID:           &ownerID,
		ReservationTypeID: 3,
		DoctorID:          &doctorID,
		IsDesignated:      true,
		CustomerFields:    jsonRawOrEmpty(`{"memo":"x"}`),
	}

	input := req.toServiceInput(9)

	if input.StartTime != startTime {
		t.Errorf("StartTime = %v, want %v", input.StartTime, startTime)
	}
	if input.OwnerID == nil || *input.OwnerID != ownerID {
		t.Errorf("OwnerID = %v, want %d", input.OwnerID, ownerID)
	}
	if input.CreatedBy == nil || *input.CreatedBy != 9 {
		t.Errorf("CreatedBy = %v, want 9", input.CreatedBy)
	}
	if string(input.CustomerFields) != string(req.CustomerFields) {
		t.Errorf("CustomerFields = %s, want %s", input.CustomerFields, req.CustomerFields)
	}
}

func TestCreateReservationAdminRequest_ToServiceInput_NilOptionalFields(t *testing.T) {
	startTime := time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)
	req := createReservationAdminRequest{
		StartTime:         startTime,
		EndTime:           endTime,
		ReservationTypeID: 3,
	}

	input := req.toServiceInput(5)

	if input.OwnerID != nil {
		t.Errorf("OwnerID = %v, want nil", input.OwnerID)
	}
	if input.PetID != nil {
		t.Errorf("PetID = %v, want nil", input.PetID)
	}
	if input.DoctorID != nil {
		t.Errorf("DoctorID = %v, want nil", input.DoctorID)
	}
	if input.LineCustomerID != nil {
		t.Errorf("LineCustomerID = %v, want nil", input.LineCustomerID)
	}
	if input.IsDesignated {
		t.Error("IsDesignated = true, want false")
	}
	if input.IsStaffDelegated {
		t.Error("IsStaffDelegated = true, want false")
	}
	if input.CreatedBy == nil || *input.CreatedBy != 5 {
		t.Errorf("CreatedBy = %v, want 5", input.CreatedBy)
	}
}
