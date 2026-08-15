package medicalrecord

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestCreatePrescriptionRequest_ToServiceInput(t *testing.T) {
	req := createPrescriptionRequest{
		Date:         "2026-05-28",
		DurationDays: 7,
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if got := input.PrescribedAt.Format("2006-01-02"); got != req.Date {
		t.Fatalf("PrescribedAt = %q, want %q", got, req.Date)
	}
	if input.DurationDays != req.DurationDays {
		t.Fatalf("DurationDays = %d, want %d", input.DurationDays, req.DurationDays)
	}
}

func TestCreatePrescriptionRequest_ToServiceInput_InvalidDate(t *testing.T) {
	req := createPrescriptionRequest{
		Date:         "2026/05/28",
		DurationDays: 7,
	}

	input, err := req.toServiceInput()
	if err == nil {
		t.Fatal("toServiceInput returned nil error")
	}
	if input != nil {
		t.Fatalf("input = %#v, want nil", input)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestUpdatePrescriptionRequest_ToServiceInput(t *testing.T) {
	date := "2026-05-29"
	durationDays := 14
	req := updatePrescriptionRequest{
		Date:         &date,
		DurationDays: &durationDays,
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.PrescribedAt == nil {
		t.Fatal("PrescribedAt = nil, want value")
	}
	if got := input.PrescribedAt.Format("2006-01-02"); got != date {
		t.Fatalf("PrescribedAt = %q, want %q", got, date)
	}
	if input.DurationDays == nil || *input.DurationDays != durationDays {
		t.Fatalf("DurationDays = %v, want %d", input.DurationDays, durationDays)
	}
}

func TestUpdatePrescriptionRequest_ToServiceInput_EmptyDate(t *testing.T) {
	date := ""
	durationDays := 3
	req := updatePrescriptionRequest{
		Date:         &date,
		DurationDays: &durationDays,
	}

	input, err := req.toServiceInput()
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.PrescribedAt != nil {
		t.Fatalf("PrescribedAt = %v, want nil", input.PrescribedAt)
	}
	if input.DurationDays == nil || *input.DurationDays != durationDays {
		t.Fatalf("DurationDays = %v, want %d", input.DurationDays, durationDays)
	}
}

func TestUpdatePrescriptionRequest_ToServiceInput_InvalidDate(t *testing.T) {
	date := "May 29, 2026"
	req := updatePrescriptionRequest{
		Date: &date,
	}

	input, err := req.toServiceInput()
	if err == nil {
		t.Fatal("toServiceInput returned nil error")
	}
	if input != nil {
		t.Fatalf("input = %#v, want nil", input)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}
