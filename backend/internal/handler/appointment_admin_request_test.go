package handler

import (
	"testing"
	"time"
)

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
