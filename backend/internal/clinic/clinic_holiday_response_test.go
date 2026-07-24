package clinic

import (
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToClinicHolidayResponse(t *testing.T) {
	date := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 1, 9, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 4, 2, 10, 0, 0, 0, time.UTC)

	holiday := &model.ClinicHoliday{
		ID:        7,
		ClinicID:  3,
		Date:      date,
		Reason:    "こどもの日",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	got := ToClinicHolidayResponse(holiday)

	if got.ID != 7 {
		t.Errorf("ID = %d, want 7", got.ID)
	}
	if got.ClinicID != 3 {
		t.Errorf("ClinicID = %d, want 3", got.ClinicID)
	}
	if got.Date != date.In(time.Local).Format("2006-01-02") {
		t.Errorf("Date = %q, want %q", got.Date, date.In(time.Local).Format("2006-01-02"))
	}
	if got.Reason != "こどもの日" {
		t.Errorf("Reason = %q, want %q", got.Reason, "こどもの日")
	}
	if got.CreatedAt != createdAt.In(time.Local).Format(time.RFC3339) {
		t.Errorf("CreatedAt = %q, want %q", got.CreatedAt, createdAt.In(time.Local).Format(time.RFC3339))
	}
	if got.UpdatedAt != updatedAt.In(time.Local).Format(time.RFC3339) {
		t.Errorf("UpdatedAt = %q, want %q", got.UpdatedAt, updatedAt.In(time.Local).Format(time.RFC3339))
	}
}

func TestToClinicHolidayResponse_ZeroValueReason(t *testing.T) {
	holiday := &model.ClinicHoliday{}

	got := ToClinicHolidayResponse(holiday)

	if got.Reason != "" {
		t.Errorf("Reason = %q, want empty", got.Reason)
	}
	if got.ID != 0 {
		t.Errorf("ID = %d, want 0", got.ID)
	}
}
