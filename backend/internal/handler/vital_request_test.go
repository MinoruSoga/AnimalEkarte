package handler

import (
	"testing"
	"time"
)

func TestCreateVitalRequest_ToServiceInput(t *testing.T) {
	recordedAt := time.Date(2026, 5, 28, 10, 30, 0, 0, time.UTC)
	staffID := uint64(11)
	temperature := 38.5
	heartRate := 120
	respirationRate := 30
	weight := 4.2
	weightUnit := "kg"

	req := createVitalRequest{
		RecordedAt:      recordedAt,
		StaffID:         &staffID,
		Temperature:     &temperature,
		HeartRate:       &heartRate,
		RespirationRate: &respirationRate,
		Weight:          &weight,
		WeightUnit:      &weightUnit,
		Notes:           "stable",
	}

	input := req.toServiceInput(1, 2)

	if input.ClinicID != 1 {
		t.Fatalf("ClinicID = %d, want 1", input.ClinicID)
	}
	if input.PetID != 2 {
		t.Fatalf("PetID = %d, want 2", input.PetID)
	}
	if !input.RecordedAt.Equal(recordedAt) {
		t.Fatalf("RecordedAt = %v, want %v", input.RecordedAt, recordedAt)
	}
	if input.StaffID != &staffID || input.Temperature != &temperature || input.HeartRate != &heartRate ||
		input.RespirationRate != &respirationRate || input.Weight != &weight {
		t.Fatalf("measurement pointers were not preserved")
	}
	if input.WeightUnit == nil || string(*input.WeightUnit) != weightUnit {
		t.Fatalf("WeightUnit = %v, want %q", input.WeightUnit, weightUnit)
	}
	if input.Notes != req.Notes {
		t.Fatalf("Notes = %q, want %q", input.Notes, req.Notes)
	}
}

func TestUpdateVitalRequest_ToServiceInput(t *testing.T) {
	recordedAt := time.Date(2026, 5, 28, 10, 30, 0, 0, time.UTC)
	staffID := uint64(0)
	temperature := 0.0
	heartRate := 0
	respirationRate := 0
	weight := 0.0
	weightUnit := "g"
	notes := ""

	req := updateVitalRequest{
		RecordedAt:      &recordedAt,
		StaffID:         &staffID,
		Temperature:     &temperature,
		HeartRate:       &heartRate,
		RespirationRate: &respirationRate,
		Weight:          &weight,
		WeightUnit:      &weightUnit,
		Notes:           &notes,
	}

	input := req.toServiceInput(9)

	if input.RecordedAt != &recordedAt || input.StaffID != &staffID || input.Temperature != &temperature ||
		input.HeartRate != &heartRate || input.RespirationRate != &respirationRate || input.Weight != &weight ||
		input.Notes != &notes {
		t.Fatalf("update pointers were not preserved")
	}
	if input.WeightUnit == nil || string(*input.WeightUnit) != weightUnit {
		t.Fatalf("WeightUnit = %v, want %q", input.WeightUnit, weightUnit)
	}
	if input.ActorID == nil || *input.ActorID != 9 {
		t.Fatalf("ActorID = %v, want 9", input.ActorID)
	}
}

func TestUpdateVitalRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateVitalRequest{}).toServiceInput(9)

	if input.RecordedAt != nil || input.StaffID != nil || input.Temperature != nil || input.HeartRate != nil ||
		input.RespirationRate != nil || input.Weight != nil || input.WeightUnit != nil || input.Notes != nil {
		t.Fatalf("input = %+v, want request fields nil", input)
	}
	if input.ActorID == nil || *input.ActorID != 9 {
		t.Fatalf("ActorID = %v, want 9", input.ActorID)
	}
}
