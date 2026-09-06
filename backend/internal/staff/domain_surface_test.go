package staff_test

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/staff"
)

func TestStaffVerticalSliceOwnsPersistenceAndUseCases(t *testing.T) {
	t.Parallel()

	if staff.NewRepository(nil) == nil {
		t.Fatal("staff repository constructor returned nil")
	}
	if staff.NewOccupationRepository(nil) == nil {
		t.Fatal("occupation repository constructor returned nil")
	}
	if staff.NewShiftEntryRepository(nil) == nil {
		t.Fatal("shift-entry repository constructor returned nil")
	}
	if staff.NewShiftTemplateRepository(nil) == nil {
		t.Fatal("shift-template repository constructor returned nil")
	}
	if staff.NewStaffClinicAssignmentRepository(nil) == nil {
		t.Fatal("staff-clinic-assignment repository constructor returned nil")
	}
}
