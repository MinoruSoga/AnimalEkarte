package medicalrecord

import "testing"

func TestAddVitalRecordRequest_ToServiceInput(t *testing.T) {
	temperature := 38.5
	heartRate := 120
	staffID := uint64(9)
	req := addVitalRecordRequest{
		Time:        "09:30:00",
		Temperature: &temperature,
		HeartRate:   &heartRate,
		Notes:       "stable",
		StaffID:     &staffID,
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if got := input.Time.Format(dailyRecordTimeLayout); got != req.Time {
		t.Errorf("Time = %q, want %q", got, req.Time)
	}
	if input.Temperature == nil || *input.Temperature != temperature {
		t.Errorf("Temperature = %v, want %v", input.Temperature, temperature)
	}
	if input.StaffID == nil || *input.StaffID != staffID {
		t.Errorf("StaffID = %v, want %d", input.StaffID, staffID)
	}
}

func TestAddVitalRecordRequest_ToServiceInput_InvalidTime(t *testing.T) {
	req := addVitalRecordRequest{Time: "09:30"}

	if _, err := req.toServiceInput(); err == nil {
		t.Fatal("toServiceInput() error = nil, want error")
	}
}

func TestAddCareLogRequest_ToServiceInput(t *testing.T) {
	staffID := uint64(9)
	req := addCareLogRequest{
		Time:    "10:15:00",
		Type:    "food",
		Status:  "partial",
		Value:   "half",
		StaffID: &staffID,
		Notes:   "notes",
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.Time != req.Time {
		t.Errorf("Time = %q, want %q", input.Time, req.Time)
	}
	if input.Type != req.Type {
		t.Errorf("Type = %q, want %q", input.Type, req.Type)
	}
	if input.Status != req.Status {
		t.Errorf("Status = %q, want %q", input.Status, req.Status)
	}
	if input.StaffID == nil || *input.StaffID != staffID {
		t.Errorf("StaffID = %v, want %d", input.StaffID, staffID)
	}
}

func TestAddCareLogRequest_ToServiceInput_InvalidTime(t *testing.T) {
	req := addCareLogRequest{Time: "10:15"}

	if _, err := req.toServiceInput(); err == nil {
		t.Fatal("toServiceInput() error = nil, want error")
	}
}

func TestAddStaffNoteRequest_ToServiceInput(t *testing.T) {
	staffID := uint64(9)
	req := addStaffNoteRequest{
		Time:    "11:00:00",
		Content: "content",
		StaffID: &staffID,
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.Time != req.Time {
		t.Errorf("Time = %q, want %q", input.Time, req.Time)
	}
	if input.Content != req.Content {
		t.Errorf("Content = %q, want %q", input.Content, req.Content)
	}
	if input.StaffID == nil || *input.StaffID != staffID {
		t.Errorf("StaffID = %v, want %d", input.StaffID, staffID)
	}
}

func TestAddStaffNoteRequest_ToServiceInput_InvalidTime(t *testing.T) {
	req := addStaffNoteRequest{Time: "11:00"}

	if _, err := req.toServiceInput(); err == nil {
		t.Fatal("toServiceInput() error = nil, want error")
	}
}
