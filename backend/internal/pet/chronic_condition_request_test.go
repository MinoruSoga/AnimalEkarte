package pet

import (
	"testing"
	"time"
)

func TestCreateChronicConditionRequest_ToServiceInput(t *testing.T) {
	notes := "requires monitoring"
	isActive := false
	req := createChronicConditionRequest{
		ConditionCode: "heart_disease",
		ConditionName: "Heart disease",
		DiagnosedAt:   "2026-05-28",
		Notes:         &notes,
		IsActive:      &isActive,
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.ConditionCode != req.ConditionCode {
		t.Errorf("ConditionCode = %q, want %q", input.ConditionCode, req.ConditionCode)
	}
	if input.ConditionName != req.ConditionName {
		t.Errorf("ConditionName = %q, want %q", input.ConditionName, req.ConditionName)
	}
	if got := input.DiagnosedAt.Format(time.DateOnly); got != req.DiagnosedAt {
		t.Errorf("DiagnosedAt = %q, want %q", got, req.DiagnosedAt)
	}
	if input.Notes == nil || *input.Notes != notes {
		t.Errorf("Notes = %v, want %q", input.Notes, notes)
	}
	if input.IsActive {
		t.Error("IsActive = true, want false")
	}
}

func TestCreateChronicConditionRequest_ToServiceInput_DefaultActive(t *testing.T) {
	req := createChronicConditionRequest{
		ConditionCode: "kidney_disease",
		ConditionName: "Kidney disease",
		DiagnosedAt:   "2026-05-28",
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if !input.IsActive {
		t.Error("IsActive = false, want true")
	}
}

func TestCreateChronicConditionRequest_ToServiceInput_InvalidDate(t *testing.T) {
	req := createChronicConditionRequest{
		ConditionCode: "kidney_disease",
		ConditionName: "Kidney disease",
		DiagnosedAt:   "2026/05/28",
	}

	if _, err := req.toServiceInput(); err == nil {
		t.Fatal("toServiceInput() error = nil, want error")
	}
}

func TestUpdateChronicConditionRequest_ToServiceInput(t *testing.T) {
	conditionCode := "heart_disease"
	conditionName := "Heart disease"
	diagnosedAt := "2026-05-28"
	notes := ""
	isActive := false
	req := updateChronicConditionRequest{
		ConditionCode: &conditionCode,
		ConditionName: &conditionName,
		DiagnosedAt:   &diagnosedAt,
		Notes:         &notes,
		IsActive:      &isActive,
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.ConditionCode == nil || *input.ConditionCode != conditionCode {
		t.Errorf("ConditionCode = %v, want %q", input.ConditionCode, conditionCode)
	}
	if input.ConditionName == nil || *input.ConditionName != conditionName {
		t.Errorf("ConditionName = %v, want %q", input.ConditionName, conditionName)
	}
	if input.DiagnosedAt == nil || input.DiagnosedAt.Format(time.DateOnly) != diagnosedAt {
		t.Errorf("DiagnosedAt = %v, want %q", input.DiagnosedAt, diagnosedAt)
	}
	if input.Notes == nil || *input.Notes != notes {
		t.Errorf("Notes = %v, want empty string pointer", input.Notes)
	}
	if input.IsActive == nil || *input.IsActive {
		t.Errorf("IsActive = %v, want false pointer", input.IsActive)
	}
}

func TestUpdateChronicConditionRequest_ToServiceInput_NilFields(t *testing.T) {
	req := updateChronicConditionRequest{}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.ConditionCode != nil {
		t.Errorf("ConditionCode = %v, want nil", input.ConditionCode)
	}
	if input.ConditionName != nil {
		t.Errorf("ConditionName = %v, want nil", input.ConditionName)
	}
	if input.DiagnosedAt != nil {
		t.Errorf("DiagnosedAt = %v, want nil", input.DiagnosedAt)
	}
	if input.Notes != nil {
		t.Errorf("Notes = %v, want nil", input.Notes)
	}
	if input.IsActive != nil {
		t.Errorf("IsActive = %v, want nil", input.IsActive)
	}
}

func TestUpdateChronicConditionRequest_ToServiceInput_InvalidDate(t *testing.T) {
	diagnosedAt := "2026/05/28"
	req := updateChronicConditionRequest{DiagnosedAt: &diagnosedAt}

	if _, err := req.toServiceInput(); err == nil {
		t.Fatal("toServiceInput() error = nil, want error")
	}
}
