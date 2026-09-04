package medicalrecord

import (
	"strings"
	"testing"
)

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

func TestDailyRecordRequests_RejectsOverMax(t *testing.T) {
	over1000 := strings.Repeat("x", 1001)
	tests := []struct {
		name string
		dst  any
		body map[string]any
	}{
		{
			name: "vital notes over max=1000",
			dst:  &addVitalRecordRequest{},
			body: map[string]any{"time": "09:30:00", "notes": over1000},
		},
		{
			name: "care value over max=1000",
			dst:  &addCareLogRequest{},
			body: map[string]any{"time": "10:15:00", "type": "food", "value": over1000},
		},
		{
			name: "care notes over max=1000",
			dst:  &addCareLogRequest{},
			body: map[string]any{"time": "10:15:00", "type": "food", "notes": over1000},
		},
		{
			name: "staff note content over max=1000",
			dst:  &addStaffNoteRequest{},
			body: map[string]any{"time": "11:00:00", "content": over1000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := shouldBindJSON(t, tt.body, tt.dst); err == nil {
				t.Fatal("ShouldBindJSON() error = nil, want over-max rejection")
			}
		})
	}
}

func TestDailyRecordRequests_AcceptsAtMax(t *testing.T) {
	at1000 := strings.Repeat("x", 1000)
	tests := []struct {
		name string
		dst  any
		body map[string]any
	}{
		{
			name: "vital notes at max=1000",
			dst:  &addVitalRecordRequest{},
			body: map[string]any{"time": "09:30:00", "notes": at1000},
		},
		{
			name: "care value and notes at max=1000",
			dst:  &addCareLogRequest{},
			body: map[string]any{"time": "10:15:00", "type": "food", "value": at1000, "notes": at1000},
		},
		{
			name: "staff note content at max=1000",
			dst:  &addStaffNoteRequest{},
			body: map[string]any{"time": "11:00:00", "content": at1000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := shouldBindJSON(t, tt.body, tt.dst); err != nil {
				t.Fatalf("ShouldBindJSON() error = %v, want nil at max", err)
			}
		})
	}
}
