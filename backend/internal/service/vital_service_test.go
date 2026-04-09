package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Vital モック ----

type mockVitalRepository struct {
	listByMedicalRecordIDFn func(ctx context.Context, medicalRecordID uint64) ([]model.VitalRecord, error)
	findByIDFn              func(ctx context.Context, clinicID uint64, vitalID uint64) (*model.VitalRecord, error)
	createFn                func(ctx context.Context, vital *model.VitalRecord) error
	updateFn                func(ctx context.Context, vitalID uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, vitalID uint64) error
}

func (m *mockVitalRepository) ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.VitalRecord, error) {
	return m.listByMedicalRecordIDFn(ctx, medicalRecordID)
}

func (m *mockVitalRepository) FindByID(ctx context.Context, clinicID uint64, vitalID uint64) (*model.VitalRecord, error) {
	return m.findByIDFn(ctx, clinicID, vitalID)
}

func (m *mockVitalRepository) Create(ctx context.Context, vital *model.VitalRecord) error {
	return m.createFn(ctx, vital)
}

func (m *mockVitalRepository) Update(ctx context.Context, vitalID uint64, fields map[string]any) error {
	return m.updateFn(ctx, vitalID, fields)
}

func (m *mockVitalRepository) Delete(ctx context.Context, vitalID uint64) error {
	return m.deleteFn(ctx, vitalID)
}

// ---- Tests ----

func TestVitalService_List(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		repoVitals      []model.VitalRecord
		repoErr         error
		wantLen         int
		wantErr         bool
	}{
		{
			name:            "returns vitals for medical record",
			medicalRecordID: 1,
			repoVitals: []model.VitalRecord{
				{ID: 1, MedicalRecordID: ptrUint64(1), Temperature: ptrFloat(37.5), HeartRate: ptrInt(80)},
				{ID: 2, MedicalRecordID: ptrUint64(1), Temperature: ptrFloat(37.3), HeartRate: ptrInt(78)},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:            "returns empty list when no vitals exist",
			medicalRecordID: 999,
			repoVitals:      []model.VitalRecord{},
			repoErr:         nil,
			wantLen:         0,
			wantErr:         false,
		},
		{
			name:            "propagates repository error",
			medicalRecordID: 1,
			repoVitals:      nil,
			repoErr:         errors.New("db error"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVitalRepository{
				listByMedicalRecordIDFn: func(_ context.Context, _ uint64) ([]model.VitalRecord, error) {
					return tt.repoVitals, tt.repoErr
				},
			}
			svc := NewVitalService(repo)

			vitals, err := svc.List(context.Background(), tt.medicalRecordID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, vitals, tt.wantLen)
			}
		})
	}
}

func TestVitalService_Create(t *testing.T) {
	recordedAt := time.Now()
	staffID := uint64(1)
	temperature := 37.5
	heartRate := 80
	respirationRate := 18
	weight := 25.5

	tests := []struct {
		name            string
		medicalRecordID uint64
		input           *CreateVitalInput
		repoErr         error
		wantErr         bool
	}{
		{
			name:            "creates vital successfully with all fields",
			medicalRecordID: 1,
			input: &CreateVitalInput{
				PetID:           1,
				RecordedAt:      recordedAt,
				StaffID:         &staffID,
				Temperature:     &temperature,
				HeartRate:       &heartRate,
				RespirationRate: &respirationRate,
				Weight:          &weight,
				Notes:           "Normal vital signs",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "creates vital with temperature only",
			medicalRecordID: 1,
			input: &CreateVitalInput{
				PetID:       1,
				RecordedAt:  recordedAt,
				Temperature: &temperature,
				Notes:       "Temperature only",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "returns error when all vital values are nil",
			medicalRecordID: 1,
			input: &CreateVitalInput{
				PetID:      1,
				RecordedAt: recordedAt,
				Notes:      "No vital values",
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:            "returns error when all vital values are nil and notes is empty",
			medicalRecordID: 1,
			input: &CreateVitalInput{
				PetID:      1,
				RecordedAt: recordedAt,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:            "returns error when pet_id is zero",
			medicalRecordID: 1,
			input: &CreateVitalInput{
				PetID:      0,
				RecordedAt: recordedAt,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:            "returns error when repository fails",
			medicalRecordID: 1,
			input: &CreateVitalInput{
				PetID:       1,
				RecordedAt:  recordedAt,
				Temperature: &temperature,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVitalRepository{
				createFn: func(_ context.Context, _ *model.VitalRecord) error {
					return tt.repoErr
				},
			}
			svc := NewVitalService(repo)

			vital, err := svc.Create(context.Background(), tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, vital)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vital)
				assert.Equal(t, ptrUint64(tt.medicalRecordID), vital.MedicalRecordID)
			}
		})
	}
}

func TestVitalService_Update(t *testing.T) {
	updatedTemperature := 38.0
	updatedHeartRate := 85
	updatedNotes := "Updated vital record"

	tests := []struct {
		name            string
		clinicID        uint64
		medicalRecordID uint64
		vitalID         uint64
		input           *UpdateVitalInput
		repoVital       *model.VitalRecord
		findByIDErr     error
		updateErr       error
		wantErr         bool
	}{
		{
			name:            "updates vital successfully",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         1,
			input: &UpdateVitalInput{
				Temperature: &updatedTemperature,
				HeartRate:   &updatedHeartRate,
			},
			repoVital: &model.VitalRecord{
				ID:              1,
				MedicalRecordID: ptrUint64(1),
				Temperature:     &updatedTemperature,
				HeartRate:       &updatedHeartRate,
			},
			findByIDErr: nil,
			updateErr:   nil,
			wantErr:     false,
		},
		{
			name:            "returns error when no fields provided",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         1,
			input:           &UpdateVitalInput{},
			repoVital: &model.VitalRecord{
				ID:              1,
				MedicalRecordID: ptrUint64(1),
			},
			findByIDErr: nil,
			updateErr:   nil,
			wantErr:     true,
		},
		{
			name:            "returns not found error when vital does not belong to medical record",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         999,
			input: &UpdateVitalInput{
				Notes: &updatedNotes,
			},
			repoVital: &model.VitalRecord{
				ID:              999,
				MedicalRecordID: ptrUint64(2), // Different medical record
			},
			findByIDErr: nil,
			updateErr:   nil,
			wantErr:     true,
		},
		{
			name:            "returns error when vital not found",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         999,
			input: &UpdateVitalInput{
				Temperature: &updatedTemperature,
			},
			repoVital:   nil,
			findByIDErr: apperrors.WrapNotFound("vital", "999"),
			updateErr:   nil,
			wantErr:     true,
		},
		{
			name:            "returns error when update fails",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         1,
			input: &UpdateVitalInput{
				Temperature: &updatedTemperature,
			},
			repoVital: &model.VitalRecord{
				ID:              1,
				MedicalRecordID: ptrUint64(1),
			},
			findByIDErr: nil,
			updateErr:   errors.New("db error"),
			wantErr:     true,
		},
		{
			name:            "updates only notes field",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         1,
			input: &UpdateVitalInput{
				Notes: &updatedNotes,
			},
			repoVital: &model.VitalRecord{
				ID:              1,
				MedicalRecordID: ptrUint64(1),
				Notes:           updatedNotes,
			},
			findByIDErr: nil,
			updateErr:   nil,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVitalRepository{
				findByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.VitalRecord, error) {
					return tt.repoVital, tt.findByIDErr
				},
				updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
					return tt.updateErr
				},
			}
			svc := NewVitalService(repo)

			vital, err := svc.Update(context.Background(), tt.clinicID, tt.medicalRecordID, tt.vitalID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vital)
			}
		})
	}
}

func TestVitalService_Delete(t *testing.T) {
	tests := []struct {
		name            string
		clinicID        uint64
		medicalRecordID uint64
		vitalID         uint64
		repoVital       *model.VitalRecord
		findByIDErr     error
		deleteErr       error
		wantErr         bool
	}{
		{
			name:            "deletes vital successfully",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         1,
			repoVital: &model.VitalRecord{
				ID:              1,
				MedicalRecordID: ptrUint64(1),
			},
			findByIDErr: nil,
			deleteErr:   nil,
			wantErr:     false,
		},
		{
			name:            "returns not found error when vital does not belong to medical record",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         999,
			repoVital: &model.VitalRecord{
				ID:              999,
				MedicalRecordID: ptrUint64(2), // Different medical record
			},
			findByIDErr: nil,
			deleteErr:   nil,
			wantErr:     true,
		},
		{
			name:            "returns error when vital not found",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         999,
			repoVital:       nil,
			findByIDErr:     apperrors.WrapNotFound("vital", "999"),
			deleteErr:       nil,
			wantErr:         true,
		},
		{
			name:            "returns error when delete fails",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         1,
			repoVital: &model.VitalRecord{
				ID:              1,
				MedicalRecordID: ptrUint64(1),
			},
			findByIDErr: nil,
			deleteErr:   errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVitalRepository{
				findByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.VitalRecord, error) {
					return tt.repoVital, tt.findByIDErr
				},
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := NewVitalService(repo)

			err := svc.Delete(context.Background(), tt.clinicID, tt.medicalRecordID, tt.vitalID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions to create pointers
func ptrFloat(f float64) *float64 {
	return &f
}

func ptrInt(i int) *int {
	return &i
}
