package medicalrecord

import (
	"strconv"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToPrescriptionResponse(t *testing.T) {
	prescribedAt := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	petID := uint64(4)
	medicalRecordID := uint64(8)

	tests := []struct {
		name string
		p    *model.Prescription
	}{
		{
			name: "converts prescription with pet_id and medical_record_id set",
			p: &model.Prescription{
				ID:              1,
				ClinicID:        2,
				OwnerID:         3,
				PetID:           &petID,
				MedicalRecordID: &medicalRecordID,
				PrescribedAt:    prescribedAt,
				DurationDays:    7,
				CreatedAt:       createdAt,
				UpdatedAt:       updatedAt,
			},
		},
		{
			name: "converts prescription with nil pet_id and medical_record_id",
			p: &model.Prescription{
				ID:           5,
				ClinicID:     2,
				OwnerID:      3,
				PrescribedAt: prescribedAt,
				DurationDays: 0,
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toPrescriptionResponse(tt.p)

			if got.ID != strconv.FormatUint(tt.p.ID, 10) {
				t.Errorf("ID = %q, want %q", got.ID, strconv.FormatUint(tt.p.ID, 10))
			}
			if got.ClinicID != strconv.FormatUint(tt.p.ClinicID, 10) {
				t.Errorf("ClinicID = %q, want %q", got.ClinicID, strconv.FormatUint(tt.p.ClinicID, 10))
			}
			if got.OwnerID != strconv.FormatUint(tt.p.OwnerID, 10) {
				t.Errorf("OwnerID = %q, want %q", got.OwnerID, strconv.FormatUint(tt.p.OwnerID, 10))
			}
			if got.PrescribedAt != prescribedAt.In(time.Local).Format("2006-01-02") {
				t.Errorf("PrescribedAt = %q, want %q", got.PrescribedAt, prescribedAt.In(time.Local).Format("2006-01-02"))
			}
			if got.DurationDays != tt.p.DurationDays {
				t.Errorf("DurationDays = %d, want %d", got.DurationDays, tt.p.DurationDays)
			}
			if got.CreatedAt != createdAt.In(time.Local).Format(time.RFC3339) {
				t.Errorf("CreatedAt = %q, want %q", got.CreatedAt, createdAt.In(time.Local).Format(time.RFC3339))
			}
			if got.UpdatedAt != updatedAt.In(time.Local).Format(time.RFC3339) {
				t.Errorf("UpdatedAt = %q, want %q", got.UpdatedAt, updatedAt.In(time.Local).Format(time.RFC3339))
			}

			if tt.p.PetID == nil {
				if got.PetID != nil {
					t.Errorf("PetID = %v, want nil", *got.PetID)
				}
			} else {
				if got.PetID == nil || *got.PetID != "4" {
					t.Errorf("PetID = %v, want %q", got.PetID, "4")
				}
			}

			if tt.p.MedicalRecordID == nil {
				if got.MedicalRecordID != nil {
					t.Errorf("MedicalRecordID = %v, want nil", *got.MedicalRecordID)
				}
			} else {
				if got.MedicalRecordID == nil || *got.MedicalRecordID != "8" {
					t.Errorf("MedicalRecordID = %v, want %q", got.MedicalRecordID, "8")
				}
			}
		})
	}
}
