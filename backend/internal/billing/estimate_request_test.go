package billing

import (
	"net/url"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestNewListEstimateQuery(t *testing.T) {
	t.Run("extracts all query parameters", func(t *testing.T) {
		values := url.Values{
			"owner_id":          []string{"10"},
			"medical_record_id": []string{"20"},
			"status":            []string{"approved"},
		}

		got := newListEstimateQuery(values)

		if got.OwnerID != "10" {
			t.Errorf("OwnerID = %q, want %q", got.OwnerID, "10")
		}
		if got.MedicalRecordID != "20" {
			t.Errorf("MedicalRecordID = %q, want %q", got.MedicalRecordID, "20")
		}
		if got.Status != "approved" {
			t.Errorf("Status = %q, want %q", got.Status, "approved")
		}
	})

	t.Run("returns zero values for empty query", func(t *testing.T) {
		got := newListEstimateQuery(url.Values{})

		if got != (listEstimateQuery{}) {
			t.Errorf("got = %#v, want zero value", got)
		}
	})
}

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
	// Callers: go test handler package. AUD-005: created_by from staffID arg, not body.
	medicalRecordID := uint64(10)
	ownerID := uint64(20)
	staffID := uint64(30)
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
	}

	input := req.toServiceInput(staffID)

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
	if input.CreatedBy == nil || *input.CreatedBy != staffID {
		t.Errorf("CreatedBy = %v, want %d", input.CreatedBy, staffID)
	}
}

func TestCreateEstimateRequest_ToServiceInput_EmptyStatus(t *testing.T) {
	req := createEstimateRequest{Title: "Estimate"}

	input := req.toServiceInput(1)

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

	input := req.toServiceInput(7)

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
	if input.ActorID == nil || *input.ActorID != 7 {
		t.Errorf("ActorID = %v, want 7", input.ActorID)
	}
}
