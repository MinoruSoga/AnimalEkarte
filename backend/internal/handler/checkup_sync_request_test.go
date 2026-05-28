package handler

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/service"
)

func TestCheckupSyncPreviewQuery_ToServiceInput(t *testing.T) {
	input, err := checkupSyncPreviewQuery{
		CheckupType:         "annual",
		Species:             "dog",
		CPMStage:            string(service.CPMStageCore),
		LastVisitBefore:     "2026-05-01",
		LastVisitAfter:      "2026-01-01",
		MinAgeYears:         "2",
		MaxAgeYears:         "10",
		HasChronicCondition: "false",
		MinTotalAmount:      "10000",
		MinAnnualVisitCount: "3",
		LastCheckupBefore:   "2026-04-01",
		LastCheckupAfter:    "2026-02-01",
	}.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.CheckupType != "annual" || input.Species != "dog" {
		t.Fatalf("basic fields = %q/%q", input.CheckupType, input.Species)
	}
	if input.MinAgeYears == nil || *input.MinAgeYears != 2 {
		t.Fatalf("MinAgeYears = %v, want 2", input.MinAgeYears)
	}
	if input.HasChronicCondition == nil || *input.HasChronicCondition {
		t.Fatalf("HasChronicCondition = %v, want false", input.HasChronicCondition)
	}
	if input.MinTotalAmount == nil || *input.MinTotalAmount != 10000 {
		t.Fatalf("MinTotalAmount = %v, want 10000", input.MinTotalAmount)
	}
	if input.LastVisitBefore == nil || input.LastCheckupAfter == nil {
		t.Fatalf("date filters were not parsed")
	}
}

func TestCheckupSyncPreviewQuery_ToServiceInput_InvalidBool(t *testing.T) {
	_, err := checkupSyncPreviewQuery{HasChronicCondition: "yes"}.toServiceInput()
	if err == nil {
		t.Fatalf("toServiceInput() error = nil, want error")
	}
}

func TestCheckupSyncPreviewQuery_ToServiceInput_InvalidCPMStage(t *testing.T) {
	_, err := checkupSyncPreviewQuery{CPMStage: "invalid"}.toServiceInput()
	if err == nil {
		t.Fatalf("toServiceInput() error = nil, want error")
	}
}

func TestCheckupSyncRequest_ToServiceInput(t *testing.T) {
	input, err := checkupSyncRequest{
		CheckupType: "annual",
		OwnerIDs:    []string{"1", "2"},
		TagName:     "annual_checkup",
	}.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput() error = %v", err)
	}

	if input.CheckupType != "annual" {
		t.Fatalf("CheckupType = %q, want annual", input.CheckupType)
	}
	if len(input.OwnerIDs) != 2 || input.OwnerIDs[1] != 2 {
		t.Fatalf("OwnerIDs = %#v, want [1 2]", input.OwnerIDs)
	}
	if input.TagName != "annual_checkup" {
		t.Fatalf("TagName = %q, want annual_checkup", input.TagName)
	}
}

func TestCheckupSyncRequest_ToServiceInput_InvalidOwnerID(t *testing.T) {
	_, err := checkupSyncRequest{
		CheckupType: "annual",
		OwnerIDs:    []string{"x"},
		TagName:     "annual_checkup",
	}.toServiceInput()
	if err == nil {
		t.Fatalf("toServiceInput() error = nil, want error")
	}
}
