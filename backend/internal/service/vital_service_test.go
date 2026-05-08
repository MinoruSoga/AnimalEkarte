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
	listByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error)
	findByIDFn              func(ctx context.Context, clinicID, vitalID uint64) (*model.VitalRecord, error)
	createFn                func(ctx context.Context, vital *model.VitalRecord) error
	updateFn                func(ctx context.Context, clinicID, vitalID uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, clinicID, vitalID uint64) error
}

func (m *mockVitalRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error) {
	return m.listByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockVitalRepository) FindByID(ctx context.Context, clinicID, vitalID uint64) (*model.VitalRecord, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, vitalID)
	}
	return nil, nil
}

func (m *mockVitalRepository) Create(ctx context.Context, vital *model.VitalRecord) error {
	return m.createFn(ctx, vital)
}

func (m *mockVitalRepository) Update(ctx context.Context, clinicID, vitalID uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, vitalID, fields)
}

func (m *mockVitalRepository) Delete(ctx context.Context, clinicID, vitalID uint64) error {
	return m.deleteFn(ctx, clinicID, vitalID)
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
				listByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]model.VitalRecord, error) {
					return tt.repoVitals, tt.repoErr
				},
			}
			svc := NewVitalService(repo, &mockMedicalRecordRepository{}, nil)

			vitals, err := svc.List(context.Background(), 1, tt.medicalRecordID)

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
			svc := NewVitalService(repo, &mockMedicalRecordRepository{}, nil)

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
		parentRecord    *model.MedicalRecord
		parentErr       error
		wantErr         bool
		wantConflict    bool
	}{
		{
			name:            "updates vital on draft medical record",
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
			parentRecord: &model.MedicalRecord{ID: 1, Status: model.MedicalRecordStatusDraft},
			wantErr:      false,
		},
		{
			name:            "rejects update on finalized medical record",
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
			parentRecord: &model.MedicalRecord{ID: 1, Status: model.MedicalRecordStatusFinalized},
			wantErr:      true,
			wantConflict: true,
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
			parentRecord: &model.MedicalRecord{ID: 1, Status: model.MedicalRecordStatusDraft},
			wantErr:      true,
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
				MedicalRecordID: ptrUint64(2),
			},
			wantErr: true,
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
			parentRecord: &model.MedicalRecord{ID: 1, Status: model.MedicalRecordStatusDraft},
			updateErr:    errors.New("db error"),
			wantErr:      true,
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
			parentRecord: &model.MedicalRecord{ID: 1, Status: model.MedicalRecordStatusDraft},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVitalRepository{
				findByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.VitalRecord, error) {
					return tt.repoVital, tt.findByIDErr
				},
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.updateErr
				},
			}
			mrRepo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return tt.parentRecord, tt.parentErr
				},
			}
			svc := NewVitalService(repo, mrRepo, nil)

			vital, err := svc.Update(context.Background(), tt.clinicID, tt.medicalRecordID, tt.vitalID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
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
		parentRecord    *model.MedicalRecord
		parentErr       error
		wantErr         bool
		wantConflict    bool
	}{
		{
			name:            "deletes vital on draft medical record",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         1,
			repoVital: &model.VitalRecord{
				ID:              1,
				MedicalRecordID: ptrUint64(1),
			},
			parentRecord: &model.MedicalRecord{ID: 1, Status: model.MedicalRecordStatusDraft},
			wantErr:      false,
		},
		{
			name:            "rejects delete on finalized medical record",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         1,
			repoVital: &model.VitalRecord{
				ID:              1,
				MedicalRecordID: ptrUint64(1),
			},
			parentRecord: &model.MedicalRecord{ID: 1, Status: model.MedicalRecordStatusFinalized},
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:            "returns not found error when vital does not belong to medical record",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         999,
			repoVital: &model.VitalRecord{
				ID:              999,
				MedicalRecordID: ptrUint64(2),
			},
			wantErr: true,
		},
		{
			name:            "returns error when vital not found",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         999,
			repoVital:       nil,
			findByIDErr:     apperrors.WrapNotFound("vital", "999"),
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
			parentRecord: &model.MedicalRecord{ID: 1, Status: model.MedicalRecordStatusDraft},
			deleteErr:    errors.New("db error"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVitalRepository{
				findByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.VitalRecord, error) {
					return tt.repoVital, tt.findByIDErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.deleteErr
				},
			}
			mrRepo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return tt.parentRecord, tt.parentErr
				},
			}
			svc := NewVitalService(repo, mrRepo, nil)

			err := svc.Delete(context.Background(), tt.clinicID, tt.medicalRecordID, tt.vitalID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
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

// TestVitalService_Create_AuditLog はバイタル作成時に audit "create" が記録されることを確認する。
func TestVitalService_Create_AuditLog(t *testing.T) {
	auditSvc := &mockMedicalRecordAuditService{}
	repo := &mockVitalRepository{
		createFn: func(_ context.Context, v *model.VitalRecord) error {
			v.ID = 55
			return nil
		},
	}
	svc := NewVitalService(repo, &mockMedicalRecordRepository{}, auditSvc)

	input := &CreateVitalInput{
		ClinicID:    1,
		PetID:       10,
		RecordedAt:  time.Now(),
		Temperature: ptrFloat(38.5),
	}
	_, err := svc.Create(context.Background(), 77, input)
	assert.NoError(t, err)
	assert.Contains(t, auditSvc.calls, "create", "create 操作が audit に記録されること")
}

// TestVitalService_Update_AuditLog はバイタル更新時に audit "update" が記録されることを確認する。
func TestVitalService_Update_AuditLog(t *testing.T) {
	auditSvc := &mockMedicalRecordAuditService{}
	existingVital := &model.VitalRecord{
		ID:              55,
		MedicalRecordID: ptrUint64(77),
		WeightUnit:      model.BodyWeightUnitKg,
		RecordedAt:      time.Now(),
		Temperature:     ptrFloat(38.5),
	}
	updatedVital := &model.VitalRecord{
		ID:              55,
		MedicalRecordID: ptrUint64(77),
		WeightUnit:      model.BodyWeightUnitKg,
		RecordedAt:      time.Now(),
		Temperature:     ptrFloat(39.0),
	}
	callCount := 0
	repo := &mockVitalRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.VitalRecord, error) {
			callCount++
			if callCount == 1 {
				return existingVital, nil
			}
			return updatedVital, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			return nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: 77, Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewVitalService(repo, mrRepo, auditSvc)

	staffID := uint64(20)
	_, err := svc.Update(context.Background(), 1, 77, 55, &UpdateVitalInput{
		Temperature: ptrFloat(39.0),
		ActorID:     &staffID,
	})
	assert.NoError(t, err)
	assert.Contains(t, auditSvc.calls, "update", "update 操作が audit に記録されること")
}

// TestVitalService_Delete_AuditLog はバイタル削除時に audit "delete" が記録されることを確認する。
func TestVitalService_Delete_AuditLog(t *testing.T) {
	auditSvc := &mockMedicalRecordAuditService{}
	existingVital := &model.VitalRecord{
		ID:              55,
		MedicalRecordID: ptrUint64(77),
		WeightUnit:      model.BodyWeightUnitKg,
		RecordedAt:      time.Now(),
	}
	repo := &mockVitalRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.VitalRecord, error) {
			return existingVital, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: 77, Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewVitalService(repo, mrRepo, auditSvc)

	err := svc.Delete(context.Background(), 1, 77, 55)
	assert.NoError(t, err)
	assert.Contains(t, auditSvc.calls, "delete", "delete 操作が audit に記録されること")
}

// TestVitalService_AuditFailure_NonFatal はバイタル監査ログ失敗がメイン処理を止めないことを確認する。
func TestVitalService_AuditFailure_NonFatal(t *testing.T) {
	auditSvc := &mockMedicalRecordAuditService{
		logVitalChangeFn: func(_ context.Context, _ uint64, _ *uint64, _ string, _, _ uint64, _, _ map[string]any) error {
			return errors.New("audit db down")
		},
	}
	repo := &mockVitalRepository{
		createFn: func(_ context.Context, v *model.VitalRecord) error {
			v.ID = 1
			return nil
		},
	}
	svc := NewVitalService(repo, &mockMedicalRecordRepository{}, auditSvc)

	input := &CreateVitalInput{
		ClinicID:    1,
		PetID:       10,
		RecordedAt:  time.Now(),
		Temperature: ptrFloat(38.5),
	}
	_, err := svc.Create(context.Background(), 77, input)
	assert.NoError(t, err, "監査ログ失敗はメイン処理のエラーを返さない（best-effort）")
}
