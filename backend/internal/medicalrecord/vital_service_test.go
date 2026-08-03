package medicalrecord

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
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

// mockVitalAuditLogger は vitalAuditLogger (service_deps.go) の narrow test double。BE9-2D
// sub-batch④a 移設前は service 側 mocks_shared_test.go の 19メソッド mockAuditService を使っていたが、
// vitalService は LogVitalChange しか呼ばないため、その1メソッドと calls 記録だけを写す。
type mockVitalAuditLogger struct {
	logVitalChangeFn func(ctx context.Context, clinicID uint64, actorID *uint64, action string, vitalID, medicalRecordID uint64, oldValue, newValue map[string]any) error
	calls            []string
}

func (m *mockVitalAuditLogger) LogVitalChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, vitalID, medicalRecordID uint64, oldValue, newValue map[string]any) error {
	m.calls = append(m.calls, action)
	if m.logVitalChangeFn != nil {
		return m.logVitalChangeFn(ctx, clinicID, actorID, action, vitalID, medicalRecordID, oldValue, newValue)
	}
	return nil
}

type vitalClinicalRelationVerifierStub struct {
	petOwners map[uint64]uint64
}

func (s *vitalClinicalRelationVerifierStub) AssertOwnerInClinic(
	_ context.Context,
	_ uint64,
	ownerID uint64,
) error {
	for _, scopedOwnerID := range s.petOwners {
		if scopedOwnerID == ownerID {
			return nil
		}
	}
	return apperrors.WrapNotFound("owner", "scoped")
}

func (s *vitalClinicalRelationVerifierStub) FindPetOwnerInClinic(
	_ context.Context,
	_ uint64,
	petID uint64,
) (uint64, error) {
	ownerID, ok := s.petOwners[petID]
	if !ok {
		return 0, apperrors.WrapNotFound("pet", "scoped")
	}
	return ownerID, nil
}

func (*vitalClinicalRelationVerifierStub) AssertMedicalRecordDoctorInClinic(
	context.Context,
	uint64,
	uint64,
) error {
	return nil
}

func validVitalRelations(petID, ownerID uint64) ClinicalRelationVerifier {
	return &vitalClinicalRelationVerifierStub{petOwners: map[uint64]uint64{petID: ownerID}}
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
			svc := NewVitalService(repo, &mockMedicalRecordRepository{}, nil, &mockCheckupTransactor{})

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
		parentErr       error
		parentStatus    model.MedicalRecordStatus
		wantErr         bool
	}{
		{
			name:            "creates vital successfully with all fields",
			medicalRecordID: 1,
			input: &CreateVitalInput{
				ClinicID:        1,
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
				ClinicID:    1,
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
				ClinicID:    1,
				PetID:       1,
				RecordedAt:  recordedAt,
				Temperature: &temperature,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:            "returns error when medical record lookup fails",
			medicalRecordID: 1,
			input: &CreateVitalInput{
				ClinicID:    1,
				PetID:       1,
				RecordedAt:  recordedAt,
				Temperature: &temperature,
			},
			parentErr: errors.New("db error"),
			wantErr:   true,
		},
		{
			name:            "returns conflict when parent medical record is finalized",
			medicalRecordID: 1,
			input: &CreateVitalInput{
				ClinicID:    1,
				PetID:       1,
				RecordedAt:  recordedAt,
				Temperature: &temperature,
			},
			parentStatus: model.MedicalRecordStatusFinalized,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVitalRepository{
				createFn: func(_ context.Context, vital *model.VitalRecord) error {
					if tt.repoErr == nil {
						assert.Equal(t, tt.input.ClinicID, vital.ClinicID)
					}
					return tt.repoErr
				},
			}
			medRecordRepo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					if tt.parentErr != nil {
						return nil, tt.parentErr
					}
					// HC-006: Return draft medical record for Create tests (finalized check)
					status := tt.parentStatus
					if status == "" {
						status = model.MedicalRecordStatusDraft
					}
					return &model.MedicalRecord{
						ID:       tt.medicalRecordID,
						ClinicID: tt.input.ClinicID,
						OwnerID:  ptrUint64(100),
						PetID:    ptrUint64(tt.input.PetID),
						Status:   status,
					}, nil
				},
			}
			svc := NewVitalServiceWithRelationValidation(
				repo,
				medRecordRepo,
				nil,
				validVitalRelations(tt.input.PetID, 100),
				&clinicalStaffLockerStub{staff: &model.Staff{ID: staffID, IsActive: true}},
				&clinicalStaffAssignmentLockerStub{
					assignment: &model.StaffClinicAssignment{StaffID: staffID, ClinicID: 1},
				},
				&mockCheckupTransactor{},
			)

			vital, err := svc.Create(context.Background(), tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, vital)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vital)
				assert.Equal(t, tt.input.ClinicID, vital.ClinicID)
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
				ClinicID:        1,
				PetID:           10,
				MedicalRecordID: ptrUint64(1),
				Temperature:     &updatedTemperature,
				HeartRate:       &updatedHeartRate,
			},
			parentRecord: &model.MedicalRecord{
				ID: 1, ClinicID: 1, OwnerID: ptrUint64(100), PetID: ptrUint64(10),
				Status: model.MedicalRecordStatusDraft,
			},
			wantErr: false,
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
				ClinicID:        1,
				PetID:           10,
				MedicalRecordID: ptrUint64(1),
			},
			parentRecord: &model.MedicalRecord{
				ID: 1, ClinicID: 1, OwnerID: ptrUint64(100), PetID: ptrUint64(10),
				Status: model.MedicalRecordStatusFinalized,
			},
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
				ClinicID:        1,
				PetID:           10,
				MedicalRecordID: ptrUint64(1),
			},
			parentRecord: &model.MedicalRecord{
				ID: 1, ClinicID: 1, OwnerID: ptrUint64(100), PetID: ptrUint64(10),
				Status: model.MedicalRecordStatusDraft,
			},
			wantErr: true,
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
				ClinicID:        1,
				PetID:           10,
				MedicalRecordID: ptrUint64(1),
			},
			parentRecord: &model.MedicalRecord{
				ID: 1, ClinicID: 1, OwnerID: ptrUint64(100), PetID: ptrUint64(10),
				Status: model.MedicalRecordStatusDraft,
			},
			updateErr: errors.New("db error"),
			wantErr:   true,
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
				ClinicID:        1,
				PetID:           10,
				MedicalRecordID: ptrUint64(1),
				Notes:           updatedNotes,
			},
			parentRecord: &model.MedicalRecord{
				ID: 1, ClinicID: 1, OwnerID: ptrUint64(100), PetID: ptrUint64(10),
				Status: model.MedicalRecordStatusDraft,
			},
			wantErr: false,
		},
		{
			name:            "returns error when parent medical record lookup fails",
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
			parentErr: errors.New("db error"),
			wantErr:   true,
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
			svc := NewVitalServiceWithRelationValidation(
				repo,
				mrRepo,
				nil,
				validVitalRelations(10, 100),
				nil,
				nil,
				&mockCheckupTransactor{},
			)

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
		{
			name:            "returns error when parent medical record lookup fails",
			clinicID:        1,
			medicalRecordID: 1,
			vitalID:         1,
			repoVital: &model.VitalRecord{
				ID:              1,
				MedicalRecordID: ptrUint64(1),
			},
			parentErr: errors.New("db error"),
			wantErr:   true,
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
			svc := NewVitalService(repo, mrRepo, nil, &mockCheckupTransactor{})

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
	auditSvc := &mockVitalAuditLogger{}
	repo := &mockVitalRepository{
		createFn: func(_ context.Context, v *model.VitalRecord) error {
			v.ID = 55
			return nil
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID:       77,
				ClinicID: 1,
				OwnerID:  ptrUint64(100),
				PetID:    ptrUint64(10),
				Status:   model.MedicalRecordStatusDraft,
			}, nil
		},
	}
	svc := NewVitalServiceWithRelationValidation(
		repo,
		medRecordRepo,
		auditSvc,
		validVitalRelations(10, 100),
		nil,
		nil,
		&mockCheckupTransactor{},
	)

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
	auditSvc := &mockVitalAuditLogger{}
	existingVital := &model.VitalRecord{
		ID:              55,
		ClinicID:        1,
		PetID:           10,
		MedicalRecordID: ptrUint64(77),
		WeightUnit:      model.BodyWeightUnitKg,
		RecordedAt:      time.Now(),
		Temperature:     ptrFloat(38.5),
	}
	updatedVital := &model.VitalRecord{
		ID:              55,
		ClinicID:        1,
		PetID:           10,
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
			return &model.MedicalRecord{
				ID: 77, ClinicID: 1, OwnerID: ptrUint64(100), PetID: ptrUint64(10),
				Status: model.MedicalRecordStatusDraft,
			}, nil
		},
	}
	svc := NewVitalServiceWithRelationValidation(
		repo,
		mrRepo,
		auditSvc,
		validVitalRelations(10, 100),
		nil,
		nil,
		&mockCheckupTransactor{},
	)

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
	auditSvc := &mockVitalAuditLogger{}
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
	svc := NewVitalService(repo, mrRepo, auditSvc, &mockCheckupTransactor{})

	err := svc.Delete(context.Background(), 1, 77, 55)
	assert.NoError(t, err)
	assert.Contains(t, auditSvc.calls, "delete", "delete 操作が audit に記録されること")
}

// TestVitalService_Create_AuditFailureRollsBack は create の audit 失敗で業務 write が error になることを固定する（BUG-015 fail-closed）。
func TestVitalService_Create_AuditFailureRollsBack(t *testing.T) {
	createCalled := false
	auditSvc := &mockVitalAuditLogger{
		logVitalChangeFn: func(_ context.Context, _ uint64, _ *uint64, _ string, _, _ uint64, _, _ map[string]any) error {
			return errors.New("audit db down")
		},
	}
	repo := &mockVitalRepository{
		createFn: func(_ context.Context, v *model.VitalRecord) error {
			createCalled = true
			v.ID = 1
			return nil
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID:       77,
				ClinicID: 1,
				OwnerID:  ptrUint64(100),
				PetID:    ptrUint64(10),
				Status:   model.MedicalRecordStatusDraft,
			}, nil
		},
	}
	svc := NewVitalServiceWithRelationValidation(
		repo,
		medRecordRepo,
		auditSvc,
		validVitalRelations(10, 100),
		nil,
		nil,
		&mockCheckupTransactor{},
	)

	input := &CreateVitalInput{
		ClinicID:    1,
		PetID:       10,
		RecordedAt:  time.Now(),
		Temperature: ptrFloat(38.5),
	}
	got, err := svc.Create(context.Background(), 77, input)
	assert.Error(t, err, "audit 失敗時は create が error を返す")
	assert.Nil(t, got)
	assert.True(t, createCalled, "create は audit 前に走るが同一 tx 内で error により rollback される契約")
}

// TestVitalService_Update_AuditFailureRollsBack は update の audit 失敗で業務 write が error になることを固定する（BUG-015）。
func TestVitalService_Update_AuditFailureRollsBack(t *testing.T) {
	updateCalled := false
	auditSvc := &mockVitalAuditLogger{
		logVitalChangeFn: func(_ context.Context, _ uint64, _ *uint64, _ string, _, _ uint64, _, _ map[string]any) error {
			return errors.New("audit write failed")
		},
	}
	existingVital := &model.VitalRecord{
		ID:              55,
		ClinicID:        1,
		PetID:           10,
		MedicalRecordID: ptrUint64(77),
		WeightUnit:      model.BodyWeightUnitKg,
		RecordedAt:      time.Now(),
		Temperature:     ptrFloat(38.5),
	}
	updatedVital := &model.VitalRecord{
		ID:              55,
		ClinicID:        1,
		PetID:           10,
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
			updateCalled = true
			return nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID: 77, ClinicID: 1, OwnerID: ptrUint64(100), PetID: ptrUint64(10),
				Status: model.MedicalRecordStatusDraft,
			}, nil
		},
	}
	svc := NewVitalServiceWithRelationValidation(
		repo, mrRepo, auditSvc, validVitalRelations(10, 100), nil, nil, &mockCheckupTransactor{},
	)
	staffID := uint64(20)
	got, err := svc.Update(context.Background(), 1, 77, 55, &UpdateVitalInput{
		Temperature: ptrFloat(39.0),
		ActorID:     &staffID,
	})
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, updateCalled)
}

// TestVitalService_Delete_AuditFailureRollsBack は delete の audit 失敗で業務 write が error になることを固定する（BUG-015）。
func TestVitalService_Delete_AuditFailureRollsBack(t *testing.T) {
	deleteCalled := false
	auditSvc := &mockVitalAuditLogger{
		logVitalChangeFn: func(_ context.Context, _ uint64, _ *uint64, _ string, _, _ uint64, _, _ map[string]any) error {
			return errors.New("audit write failed")
		},
	}
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
			deleteCalled = true
			return nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: 77, Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewVitalService(repo, mrRepo, auditSvc, &mockCheckupTransactor{})
	err := svc.Delete(context.Background(), 1, 77, 55)
	assert.Error(t, err)
	assert.True(t, deleteCalled)
}

// TestVitalService_Create_RejectsInvalidWeight は weight 構造検証（NaN/Inf/≤0/不正 unit）を 400 契約で拒否する（BUG-015）。
func TestVitalService_Create_RejectsInvalidWeight(t *testing.T) {
	createCalled := false
	repo := &mockVitalRepository{
		createFn: func(_ context.Context, _ *model.VitalRecord) error {
			createCalled = true
			return nil
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID: 77, ClinicID: 1, OwnerID: ptrUint64(100), PetID: ptrUint64(10),
				Status: model.MedicalRecordStatusDraft,
			}, nil
		},
	}
	svc := NewVitalServiceWithRelationValidation(
		repo, medRecordRepo, nil, validVitalRelations(10, 100), nil, nil, &mockCheckupTransactor{},
	)

	nan := math.NaN()
	inf := math.Inf(1)
	zero := 0.0
	neg := -1.0
	okWeight := 5.0
	badUnit := model.BodyWeightUnit("lb")

	cases := []struct {
		name  string
		input *CreateVitalInput
	}{
		{
			name: "NaN weight",
			input: &CreateVitalInput{
				ClinicID: 1, PetID: 10, RecordedAt: time.Now(), Weight: &nan,
			},
		},
		{
			name: "Inf weight",
			input: &CreateVitalInput{
				ClinicID: 1, PetID: 10, RecordedAt: time.Now(), Weight: &inf,
			},
		},
		{
			name: "zero weight",
			input: &CreateVitalInput{
				ClinicID: 1, PetID: 10, RecordedAt: time.Now(), Weight: &zero,
			},
		},
		{
			name: "negative weight",
			input: &CreateVitalInput{
				ClinicID: 1, PetID: 10, RecordedAt: time.Now(), Weight: &neg,
			},
		},
		{
			name: "invalid unit",
			input: &CreateVitalInput{
				ClinicID: 1, PetID: 10, RecordedAt: time.Now(), Weight: &okWeight, WeightUnit: &badUnit,
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			createCalled = false
			got, err := svc.Create(context.Background(), 77, tt.input)
			assert.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "want invalid input, got %v", err)
			assert.Nil(t, got)
			assert.False(t, createCalled, "invalid weight must not reach repository")
		})
	}
}

// TestBuildVitalUpdate は buildVitalUpdate の全フィールド網羅とゼロ値挙動を検証する。
func TestBuildVitalUpdate(t *testing.T) {
	recordedAt := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	staffID := uint64(7)
	temperature := 38.2
	heartRate := 90
	respirationRate := 20
	weight := 12.5
	weightUnit := model.BodyWeightUnitG
	notes := "元気"

	t.Run("maps all provided fields", func(t *testing.T) {
		input := &UpdateVitalInput{
			RecordedAt:      &recordedAt,
			StaffID:         &staffID,
			Temperature:     &temperature,
			HeartRate:       &heartRate,
			RespirationRate: &respirationRate,
			Weight:          &weight,
			WeightUnit:      &weightUnit,
			Notes:           &notes,
		}
		fields := buildVitalUpdate(input)
		assert.Equal(t, recordedAt, fields["recorded_at"])
		assert.Equal(t, staffID, fields["staff_id"])
		assert.Equal(t, temperature, fields["temperature"])
		assert.Equal(t, heartRate, fields["heart_rate"])
		assert.Equal(t, respirationRate, fields["respiration_rate"])
		assert.Equal(t, weight, fields["weight"])
		assert.Equal(t, weightUnit, fields["weight_unit"])
		assert.Equal(t, notes, fields["notes"])
		assert.Len(t, fields, 8)
	})

	t.Run("returns empty map when all fields are nil", func(t *testing.T) {
		fields := buildVitalUpdate(&UpdateVitalInput{})
		assert.Empty(t, fields)
	})
}

// TestWeightUnitOrDefault は weightUnitOrDefault のデフォルト補完・明示指定の両分岐を検証する。
func TestWeightUnitOrDefault(t *testing.T) {
	t.Run("returns default kg when nil", func(t *testing.T) {
		assert.Equal(t, model.BodyWeightUnitKg, weightUnitOrDefault(nil))
	})
	t.Run("returns provided unit when set", func(t *testing.T) {
		g := model.BodyWeightUnitG
		assert.Equal(t, model.BodyWeightUnitG, weightUnitOrDefault(&g))
	})
}

func TestVitalService_CreateRejectsMedicalRecordPetMismatch(t *testing.T) {
	const (
		clinicID      = uint64(1)
		medicalRecord = uint64(10)
		recordPetID   = uint64(20)
		requestPetID  = uint64(21)
	)
	createCalls := 0
	repo := &mockVitalRepository{
		createFn: func(_ context.Context, _ *model.VitalRecord) error {
			createCalls++
			return nil
		},
	}
	medRecRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID: medicalRecord, ClinicID: clinicID, OwnerID: ptrUint64(100),
				PetID:  ptrUint64(recordPetID),
				Status: model.MedicalRecordStatusDraft,
			}, nil
		},
	}
	svc := NewVitalServiceWithRelationValidation(
		repo,
		medRecRepo,
		nil,
		&vitalClinicalRelationVerifierStub{petOwners: map[uint64]uint64{
			recordPetID: 100, requestPetID: 100,
		}},
		nil,
		nil,
		&mockCheckupTransactor{},
	)

	got, err := svc.Create(context.Background(), medicalRecord, &CreateVitalInput{
		ClinicID: clinicID, PetID: requestPetID, Temperature: ptrFloat(38.1),
	})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Zero(t, createCalls)
}

func TestVitalService_CreateRejectsInactiveOrForeignPetBeforeWrite(t *testing.T) {
	createCalls := 0
	repo := &mockVitalRepository{
		createFn: func(_ context.Context, _ *model.VitalRecord) error {
			createCalls++
			return nil
		},
	}
	medRecRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID: id, ClinicID: clinicID, OwnerID: ptrUint64(100), PetID: ptrUint64(20),
				Status: model.MedicalRecordStatusDraft,
			}, nil
		},
	}
	svc := NewVitalServiceWithRelationValidation(
		repo,
		medRecRepo,
		nil,
		&vitalClinicalRelationVerifierStub{petOwners: map[uint64]uint64{}},
		nil,
		nil,
		&mockCheckupTransactor{},
	)

	got, err := svc.Create(context.Background(), 10, &CreateVitalInput{
		ClinicID: 1, PetID: 20, Temperature: ptrFloat(38.1),
	})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Zero(t, createCalls)
}

func TestVitalService_StaffReferenceValidation(t *testing.T) {
	const (
		clinicID      = uint64(1)
		medicalRecord = uint64(10)
		petID         = uint64(20)
		staffID       = uint64(30)
	)
	newService := func(
		repo *mockVitalRepository,
		staff *model.Staff,
		assignment *model.StaffClinicAssignment,
	) VitalService {
		medRecRepo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{
					ID: medicalRecord, ClinicID: clinicID, OwnerID: ptrUint64(100),
					PetID:  ptrUint64(petID),
					Status: model.MedicalRecordStatusDraft,
				}, nil
			},
		}
		return NewVitalServiceWithRelationValidation(
			repo,
			medRecRepo,
			nil,
			validVitalRelations(petID, 100),
			&clinicalStaffLockerStub{staff: staff},
			&clinicalStaffAssignmentLockerStub{assignment: assignment},
			&mockCheckupTransactor{},
		)
	}

	t.Run("create rejects unassigned staff before write", func(t *testing.T) {
		createCalls := 0
		repo := &mockVitalRepository{
			createFn: func(_ context.Context, _ *model.VitalRecord) error {
				createCalls++
				return nil
			},
		}
		svc := newService(
			repo,
			&model.Staff{ID: staffID, IsActive: true},
			&model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID + 1},
		)

		got, err := svc.Create(context.Background(), medicalRecord, &CreateVitalInput{
			ClinicID: clinicID, PetID: petID, StaffID: ptrUint64(staffID),
			Temperature: ptrFloat(38.1),
		})

		assert.Error(t, err)
		assert.Nil(t, got)
		assert.Zero(t, createCalls)
	})

	t.Run("update rejects inactive staff before write", func(t *testing.T) {
		updateCalls := 0
		repo := &mockVitalRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.VitalRecord, error) {
				return &model.VitalRecord{
					ID: id, ClinicID: clinicID, PetID: petID, MedicalRecordID: ptrUint64(medicalRecord),
				}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				updateCalls++
				return nil
			},
		}
		svc := newService(
			repo,
			&model.Staff{ID: staffID, IsActive: false},
			&model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID},
		)

		got, err := svc.Update(context.Background(), clinicID, medicalRecord, 40, &UpdateVitalInput{
			StaffID: ptrUint64(staffID),
		})

		assert.Error(t, err)
		assert.Nil(t, got)
		assert.Zero(t, updateCalls)
	})

	t.Run("create accepts active assigned staff", func(t *testing.T) {
		createCalls := 0
		repo := &mockVitalRepository{
			createFn: func(_ context.Context, vital *model.VitalRecord) error {
				createCalls++
				vital.ID = 1
				return nil
			},
		}
		svc := newService(
			repo,
			&model.Staff{ID: staffID, IsActive: true},
			&model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID},
		)

		got, err := svc.Create(context.Background(), medicalRecord, &CreateVitalInput{
			ClinicID: clinicID, PetID: petID, StaffID: ptrUint64(staffID),
			Temperature: ptrFloat(38.1),
		})

		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, 1, createCalls)
	})
}

func TestVitalService_CreateFailsClosedWithoutStaffValidationDependencies(t *testing.T) {
	createCalls := 0
	repo := &mockVitalRepository{
		createFn: func(_ context.Context, _ *model.VitalRecord) error {
			createCalls++
			return nil
		},
	}
	medRecRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID: id, ClinicID: clinicID, OwnerID: ptrUint64(100), PetID: ptrUint64(20),
				Status: model.MedicalRecordStatusDraft,
			}, nil
		},
	}
	svc := NewVitalServiceWithRelationValidation(
		repo,
		medRecRepo,
		nil,
		validVitalRelations(20, 100),
		nil,
		nil,
		&mockCheckupTransactor{},
	)

	got, err := svc.Create(context.Background(), 10, &CreateVitalInput{
		ClinicID: 1, PetID: 20, StaffID: ptrUint64(30), Temperature: ptrFloat(38.1),
	})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Zero(t, createCalls)
}
