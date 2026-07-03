package handler

import (
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToExaminationResponse(t *testing.T) {
	petID := uint64(20)
	doctorID := uint64(30)
	medicalRecordID := uint64(40)
	date := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	tests := []struct {
		name string
		in   *model.Examination
	}{
		{
			name: "converts examination with relations",
			in: &model.Examination{
				ID:              1,
				ClinicID:        2,
				MedicalRecordID: &medicalRecordID,
				PetID:           &petID,
				ExamTypeID:      3,
				DoctorID:        &doctorID,
				Date:            date,
				ResultSummary:   "normal",
				Machine:         "machine-a",
				Status:          model.ExaminationStatusCompleted,
				CreatedAt:       now,
				UpdatedAt:       now,
				Pet:             &model.Pet{ID: petID, Name: "Pochi"},
				Doctor:          &model.Staff{ID: doctorID, Name: "Dr. Smith"},
				ExaminationType: &model.ExaminationType{ID: 3, Name: "Blood Test"},
			},
		},
		{
			name: "converts examination without relations",
			in: &model.Examination{
				ID:         2,
				ClinicID:   2,
				ExamTypeID: 3,
				Date:       date,
				Status:     model.ExaminationStatusPending,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		},
		{
			name: "converts zero-value examination",
			in:   &model.Examination{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toExaminationResponse(tt.in)

			if got.ID != tt.in.ID {
				t.Errorf("ID = %d, want %d", got.ID, tt.in.ID)
			}
			if got.ClinicID != tt.in.ClinicID {
				t.Errorf("ClinicID = %d, want %d", got.ClinicID, tt.in.ClinicID)
			}
			if got.ExamTypeID != tt.in.ExamTypeID {
				t.Errorf("ExamTypeID = %d, want %d", got.ExamTypeID, tt.in.ExamTypeID)
			}
			if got.ResultSummary != tt.in.ResultSummary {
				t.Errorf("ResultSummary = %q, want %q", got.ResultSummary, tt.in.ResultSummary)
			}
			if got.Machine != tt.in.Machine {
				t.Errorf("Machine = %q, want %q", got.Machine, tt.in.Machine)
			}
			if got.Status != string(tt.in.Status) {
				t.Errorf("Status = %q, want %q", got.Status, string(tt.in.Status))
			}

			if tt.in.Pet == nil {
				if got.Pet != nil {
					t.Errorf("Pet = %+v, want nil", got.Pet)
				}
			} else if got.Pet == nil || got.Pet.Name != tt.in.Pet.Name {
				t.Errorf("Pet = %+v, want name %q", got.Pet, tt.in.Pet.Name)
			}

			if tt.in.Doctor == nil {
				if got.Doctor != nil {
					t.Errorf("Doctor = %+v, want nil", got.Doctor)
				}
			} else if got.Doctor == nil || got.Doctor.Name != tt.in.Doctor.Name {
				t.Errorf("Doctor = %+v, want name %q", got.Doctor, tt.in.Doctor.Name)
			}

			if tt.in.ExaminationType == nil {
				if got.ExamType != nil {
					t.Errorf("ExamType = %+v, want nil", got.ExamType)
				}
			} else {
				if got.ExamType == nil {
					t.Fatalf("ExamType = nil, want non-nil")
				}
				if got.ExamType.ID != tt.in.ExaminationType.ID {
					t.Errorf("ExamType.ID = %d, want %d", got.ExamType.ID, tt.in.ExaminationType.ID)
				}
				if got.ExamType.Name != tt.in.ExaminationType.Name {
					t.Errorf("ExamType.Name = %q, want %q", got.ExamType.Name, tt.in.ExaminationType.Name)
				}
			}
		})
	}
}
