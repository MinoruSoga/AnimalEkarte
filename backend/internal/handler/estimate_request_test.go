package handler

import (
	"testing"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func TestListEstimateQuery_ToServiceFilters(t *testing.T) {
	filters, err := (listEstimateQuery{
		OwnerID:         "10",
		MedicalRecordID: "20",
		Status:          "approved",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.OwnerID == nil || *filters.OwnerID != 10 {
		t.Fatalf("OwnerID = %v, want 10", filters.OwnerID)
	}
	if filters.MedicalRecordID == nil || *filters.MedicalRecordID != 20 {
		t.Fatalf("MedicalRecordID = %v, want 20", filters.MedicalRecordID)
	}
	if filters.Status == nil || *filters.Status != "approved" {
		t.Fatalf("Status = %v, want approved", filters.Status)
	}
}

func TestListEstimateQuery_ToServiceFilters_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		query listEstimateQuery
	}{
		{name: "owner_id", query: listEstimateQuery{OwnerID: "abc"}},
		{name: "medical_record_id", query: listEstimateQuery{MedicalRecordID: "abc"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := tt.query.toServiceFilters()
			if err == nil {
				t.Fatal("toServiceFilters returned nil error")
			}
			if filters != (listEstimateFilters{}) {
				t.Fatalf("filters = %#v, want zero value", filters)
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}

func TestCreateEstimateRequest_ToServiceInput(t *testing.T) {
	medicalRecordID := uint64(10)
	ownerID := uint64(20)
	createdBy := uint64(30)
	validUntil := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	req := createEstimateRequest{
		MedicalRecordID: &medicalRecordID,
		Title:           "Estimate",
		OwnerID:         &ownerID,
		Status:          "sent",
		Subtotal:        1000,
		TaxTotal:        100,
		TotalAmount:     1100,
		InsuranceAmount: 200,
		DiscountAmount:  50,
		ValidUntil:      &validUntil,
		Comment:         "comment",
		Notes:           "notes",
		CreatedBy:       &createdBy,
	}

	input := req.toServiceInput()

	if input.MedicalRecordID == nil || *input.MedicalRecordID != medicalRecordID {
		t.Errorf("MedicalRecordID = %v, want %d", input.MedicalRecordID, medicalRecordID)
	}
	if input.Title != req.Title {
		t.Errorf("Title = %q, want %q", input.Title, req.Title)
	}
	if string(input.Status) != req.Status {
		t.Errorf("Status = %q, want %q", input.Status, req.Status)
	}
	if input.ValidUntil == nil || !input.ValidUntil.Equal(validUntil) {
		t.Errorf("ValidUntil = %v, want %v", input.ValidUntil, validUntil)
	}
	if input.CreatedBy == nil || *input.CreatedBy != createdBy {
		t.Errorf("CreatedBy = %v, want %d", input.CreatedBy, createdBy)
	}
}

func TestCreateEstimateRequest_ToServiceInput_EmptyStatus(t *testing.T) {
	req := createEstimateRequest{Title: "Estimate"}

	input := req.toServiceInput()

	if input.Status != "" {
		t.Errorf("Status = %q, want empty", input.Status)
	}
}

func TestUpdateEstimateRequest_ToServiceInput(t *testing.T) {
	title := ""
	status := "approved"
	totalAmount := int64(0)
	clearValidUntil := true
	comment := ""
	req := updateEstimateRequest{
		Title:           &title,
		Status:          &status,
		TotalAmount:     &totalAmount,
		ClearValidUntil: clearValidUntil,
		Comment:         &comment,
	}

	input := req.toServiceInput()

	if input.Title == nil || *input.Title != title {
		t.Errorf("Title = %v, want empty string pointer", input.Title)
	}
	if input.Status == nil || string(*input.Status) != status {
		t.Errorf("Status = %v, want %q", input.Status, status)
	}
	if input.TotalAmount == nil || *input.TotalAmount != totalAmount {
		t.Errorf("TotalAmount = %v, want %d", input.TotalAmount, totalAmount)
	}
	if !input.ClearValidUntil {
		t.Error("ClearValidUntil = false, want true")
	}
	if input.Comment == nil || *input.Comment != comment {
		t.Errorf("Comment = %v, want empty string pointer", input.Comment)
	}
}
