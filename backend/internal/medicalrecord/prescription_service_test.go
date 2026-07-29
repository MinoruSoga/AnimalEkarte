package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockPrescriptionRepository struct {
	findByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error)
	findByIDFn              func(ctx context.Context, clinicID, id uint64) (*model.Prescription, error)
	findActiveByOwnerFn     func(ctx context.Context, clinicID, ownerID uint64) ([]model.Prescription, error)
	createFn                func(ctx context.Context, prescription *model.Prescription) error
	updateFn                func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockPrescriptionRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error) {
	if m.findByMedicalRecordIDFn != nil {
		return m.findByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
	}
	return []model.Prescription{}, nil
}

func (m *mockPrescriptionRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Prescription, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Prescription{}, nil
}

func (m *mockPrescriptionRepository) FindActiveByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Prescription, error) {
	if m.findActiveByOwnerFn != nil {
		return m.findActiveByOwnerFn(ctx, clinicID, ownerID)
	}
	return []model.Prescription{}, nil
}

func (m *mockPrescriptionRepository) Create(ctx context.Context, prescription *model.Prescription) error {
	if m.createFn != nil {
		return m.createFn(ctx, prescription)
	}
	return nil
}

func (m *mockPrescriptionRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, fields)
	}
	return nil
}

func (m *mockPrescriptionRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

// ---- List テスト ----

func TestPrescriptionService_List(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		prescriptions   []model.Prescription
		repoErr         error
		wantLen         int
		wantErr         bool
	}{
		{
			name:            "returns prescriptions for medical record",
			medicalRecordID: 1,
			prescriptions: []model.Prescription{
				{ID: 1, DurationDays: 7},
				{ID: 2, DurationDays: 14},
			},
			wantLen: 2,
		},
		{
			name:            "returns empty list when no prescriptions exist",
			medicalRecordID: 999,
			prescriptions:   []model.Prescription{},
			wantLen:         0,
		},
		{
			name:            "propagates repository error",
			medicalRecordID: 1,
			repoErr:         errors.New("db error"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPrescriptionRepository{
				findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]model.Prescription, error) {
					return tt.prescriptions, tt.repoErr
				},
			}
			svc := NewPrescriptionService(repo, &mockMedicalRecordRepository{}, nil, &mockCheckupTransactor{})

			prescriptions, err := svc.List(context.Background(), 1, tt.medicalRecordID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, prescriptions)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, prescriptions, tt.wantLen)
		})
	}
}

// ---- GetByID テスト ----

func TestPrescriptionService_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Prescription, error)
		wantErr    bool
	}{
		{
			name: "returns prescription successfully",
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Prescription, error) {
				return &model.Prescription{ID: id}, nil
			},
		},
		{
			name: "propagates repository error",
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPrescriptionRepository{findByIDFn: tt.findByIDFn}
			svc := NewPrescriptionService(repo, &mockMedicalRecordRepository{}, nil, &mockCheckupTransactor{})

			prescription, err := svc.GetByID(context.Background(), 1, 5)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, prescription)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, prescription)
			assert.Equal(t, uint64(5), prescription.ID)
		})
	}
}

// ---- buildPrescriptionUpdate 直接テスト ----

func TestBuildPrescriptionUpdate(t *testing.T) {
	t.Run("フィールド未指定時は空map", func(t *testing.T) {
		fields := buildPrescriptionUpdate(&UpdatePrescriptionInput{})
		assert.Empty(t, fields)
	})

	t.Run("全フィールド指定時はすべてmapに含まれる", func(t *testing.T) {
		prescribedAt := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		durationDays := 10
		fields := buildPrescriptionUpdate(&UpdatePrescriptionInput{
			PrescribedAt: &prescribedAt,
			DurationDays: &durationDays,
		})
		assert.Equal(t, prescribedAt, fields["prescribed_at"])
		assert.Equal(t, durationDays, fields["duration_days"])
	})
}

// ---- Create 追加分岐テスト ----

func TestPrescriptionService_Create_MedicalRecordLookupError(t *testing.T) {
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewPrescriptionService(&mockPrescriptionRepository{}, medRecordRepo, nil, &mockCheckupTransactor{})

	result, err := svc.Create(context.Background(), 1, 2, &CreatePrescriptionInput{DurationDays: 7})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestPrescriptionService_Create_NoOwnerRejected(t *testing.T) {
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{OwnerID: nil}, nil
		},
	}
	svc := NewPrescriptionService(&mockPrescriptionRepository{}, medRecordRepo, nil, &mockCheckupTransactor{})

	result, err := svc.Create(context.Background(), 1, 2, &CreatePrescriptionInput{DurationDays: 7})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestPrescriptionService_Create_FinalizedRejected(t *testing.T) {
	ownerID := uint64(10)
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{OwnerID: &ownerID, Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	svc := NewPrescriptionService(&mockPrescriptionRepository{}, medRecordRepo, nil, &mockCheckupTransactor{})

	result, err := svc.Create(context.Background(), 1, 2, &CreatePrescriptionInput{DurationDays: 7})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsConflict(err))
}

func TestPrescriptionService_Create_RepositoryError(t *testing.T) {
	ownerID := uint64(10)
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{OwnerID: &ownerID, Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	repo := &mockPrescriptionRepository{
		createFn: func(_ context.Context, _ *model.Prescription) error {
			return errors.New("db error")
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, nil, &mockCheckupTransactor{})

	result, err := svc.Create(context.Background(), 1, 2, &CreatePrescriptionInput{DurationDays: 7})

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ---- Update 追加分岐テスト ----

func TestPrescriptionService_Update_FindByIDError(t *testing.T) {
	durationDays := 5
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewPrescriptionService(repo, &mockMedicalRecordRepository{}, nil, &mockCheckupTransactor{})

	result, err := svc.Update(context.Background(), 1, 2, 3, &UpdatePrescriptionInput{DurationDays: &durationDays})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestPrescriptionService_Update_MedicalRecordMismatch(t *testing.T) {
	durationDays := 5
	otherMedicalRecordID := uint64(99)
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return &model.Prescription{ID: 3, MedicalRecordID: &otherMedicalRecordID}, nil
		},
	}
	svc := NewPrescriptionService(repo, &mockMedicalRecordRepository{}, nil, &mockCheckupTransactor{})

	result, err := svc.Update(context.Background(), 1, 2, 3, &UpdatePrescriptionInput{DurationDays: &durationDays})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestPrescriptionService_Update_NilMedicalRecordID(t *testing.T) {
	durationDays := 5
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return &model.Prescription{ID: 3, MedicalRecordID: nil}, nil
		},
	}
	svc := NewPrescriptionService(repo, &mockMedicalRecordRepository{}, nil, &mockCheckupTransactor{})

	result, err := svc.Update(context.Background(), 1, 2, 3, &UpdatePrescriptionInput{DurationDays: &durationDays})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestPrescriptionService_Update_MedicalRecordLookupError(t *testing.T) {
	durationDays := 5
	medicalRecordID := uint64(2)
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return &model.Prescription{ID: 3, MedicalRecordID: &medicalRecordID}, nil
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, nil, &mockCheckupTransactor{})

	result, err := svc.Update(context.Background(), 1, medicalRecordID, 3, &UpdatePrescriptionInput{DurationDays: &durationDays})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestPrescriptionService_Update_FinalizedRejected(t *testing.T) {
	durationDays := 5
	medicalRecordID := uint64(2)
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return &model.Prescription{ID: 3, MedicalRecordID: &medicalRecordID}, nil
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, nil, &mockCheckupTransactor{})

	result, err := svc.Update(context.Background(), 1, medicalRecordID, 3, &UpdatePrescriptionInput{DurationDays: &durationDays})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsConflict(err))
}

func TestPrescriptionService_Update_NoFieldsProvided(t *testing.T) {
	medicalRecordID := uint64(2)
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return &model.Prescription{ID: 3, MedicalRecordID: &medicalRecordID}, nil
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, nil, &mockCheckupTransactor{})

	result, err := svc.Update(context.Background(), 1, medicalRecordID, 3, &UpdatePrescriptionInput{})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestPrescriptionService_Update_RepositoryUpdateError(t *testing.T) {
	durationDays := 5
	medicalRecordID := uint64(2)
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return &model.Prescription{ID: 3, MedicalRecordID: &medicalRecordID}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			return errors.New("db error")
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, nil, &mockCheckupTransactor{})

	result, err := svc.Update(context.Background(), 1, medicalRecordID, 3, &UpdatePrescriptionInput{DurationDays: &durationDays})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestPrescriptionService_Update_FindByIDAfterUpdateError(t *testing.T) {
	// MRC-01: reload failure is inside WithTx, so the whole update fails (no committed success + 5xx).
	durationDays := 5
	medicalRecordID := uint64(2)
	findCount := 0
	updateCalled := false
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			findCount++
			if findCount == 1 {
				return &model.Prescription{ID: 3, MedicalRecordID: &medicalRecordID, OwnerID: 9}, nil
			}
			return nil, errors.New("db error")
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			updateCalled = true
			return nil
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, nil, &mockCheckupTransactor{})

	result, err := svc.Update(context.Background(), 1, medicalRecordID, 3, &UpdatePrescriptionInput{DurationDays: &durationDays})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, updateCalled)
	assert.Equal(t, 2, findCount, "pre-check FindByID + in-tx reload FindByID")
}

// ---- Delete 追加分岐テスト ----

func TestPrescriptionService_Delete_FindByIDError(t *testing.T) {
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewPrescriptionService(repo, &mockMedicalRecordRepository{}, nil, &mockCheckupTransactor{})

	err := svc.Delete(context.Background(), 1, 2, 3)

	assert.Error(t, err)
}

func TestPrescriptionService_Delete_MedicalRecordMismatch(t *testing.T) {
	otherMedicalRecordID := uint64(99)
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return &model.Prescription{ID: 3, MedicalRecordID: &otherMedicalRecordID}, nil
		},
	}
	svc := NewPrescriptionService(repo, &mockMedicalRecordRepository{}, nil, &mockCheckupTransactor{})

	err := svc.Delete(context.Background(), 1, 2, 3)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestPrescriptionService_Delete_MedicalRecordLookupError(t *testing.T) {
	medicalRecordID := uint64(2)
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return &model.Prescription{ID: 3, MedicalRecordID: &medicalRecordID}, nil
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, nil, &mockCheckupTransactor{})

	err := svc.Delete(context.Background(), 1, medicalRecordID, 3)

	assert.Error(t, err)
}

func TestPrescriptionService_Delete_FinalizedRejected(t *testing.T) {
	medicalRecordID := uint64(2)
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return &model.Prescription{ID: 3, MedicalRecordID: &medicalRecordID}, nil
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, nil, &mockCheckupTransactor{})

	err := svc.Delete(context.Background(), 1, medicalRecordID, 3)

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
}

func TestPrescriptionService_Delete_RepositoryError(t *testing.T) {
	medicalRecordID := uint64(2)
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return &model.Prescription{ID: 3, MedicalRecordID: &medicalRecordID}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("db error")
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, nil, &mockCheckupTransactor{})

	err := svc.Delete(context.Background(), 1, medicalRecordID, 3)

	assert.Error(t, err)
}

// ---- syncPrescriptionTag 直接分岐テスト ----

func TestPrescriptionService_syncPrescriptionTag_NilTagSyncSvc(t *testing.T) {
	// tagSyncSvc=nil の場合は何もせず早期returnする（Create経由でパニックしないことを確認）。
	ownerID := uint64(10)
	repo := &mockPrescriptionRepository{
		createFn: func(_ context.Context, p *model.Prescription) error {
			p.ID = 1
			return nil
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{OwnerID: &ownerID, Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, nil, &mockCheckupTransactor{})

	result, err := svc.Create(context.Background(), 1, 2, &CreatePrescriptionInput{DurationDays: 7})

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestPrescriptionService_Create_SyncsPrescriptionTagBestEffort(t *testing.T) {
	ownerID := uint64(10)
	petID := uint64(20)
	var syncedClinicID, syncedOwnerID uint64

	repo := &mockPrescriptionRepository{
		createFn: func(_ context.Context, prescription *model.Prescription) error {
			prescription.ID = 30
			return nil
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{OwnerID: &ownerID, PetID: &petID}, nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncPrescriptionTagFn: func(_ context.Context, clinicID, ownerID uint64) error {
			syncedClinicID = clinicID
			syncedOwnerID = ownerID
			return errors.New("sync failed")
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, tagSync, &mockCheckupTransactor{})

	result, err := svc.Create(context.Background(), 1, 2, &CreatePrescriptionInput{
		PrescribedAt: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		DurationDays: 14,
	})

	require.NoError(t, err)
	assert.Equal(t, uint64(30), result.ID)
	assert.Equal(t, uint64(1), syncedClinicID)
	assert.Equal(t, ownerID, syncedOwnerID)
}

func TestPrescriptionService_Update_SyncsPrescriptionTagAfterUpdate(t *testing.T) {
	ownerID := uint64(10)
	medicalRecordID := uint64(2)
	findCount := 0
	var syncedOwnerID uint64

	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			findCount++
			if findCount == 1 {
				return &model.Prescription{OwnerID: ownerID, MedicalRecordID: &medicalRecordID}, nil
			}
			return &model.Prescription{ID: 30, OwnerID: ownerID, MedicalRecordID: &medicalRecordID, DurationDays: 21}, nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncPrescriptionTagFn: func(_ context.Context, _, ownerID uint64) error {
			syncedOwnerID = ownerID
			return nil
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: medicalRecordID, Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, tagSync, &mockCheckupTransactor{})
	durationDays := 21

	result, err := svc.Update(context.Background(), 1, medicalRecordID, 30, &UpdatePrescriptionInput{DurationDays: &durationDays})

	require.NoError(t, err)
	assert.Equal(t, 21, result.DurationDays)
	assert.Equal(t, ownerID, syncedOwnerID)
}

func TestPrescriptionService_Delete_SyncsPrescriptionTagAfterDelete(t *testing.T) {
	ownerID := uint64(10)
	medicalRecordID := uint64(2)
	deleted := false
	syncedAfterDelete := false

	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return &model.Prescription{OwnerID: ownerID, MedicalRecordID: &medicalRecordID}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleted = true
			return nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncPrescriptionTagFn: func(_ context.Context, _, syncedOwnerID uint64) error {
			syncedAfterDelete = deleted
			assert.Equal(t, ownerID, syncedOwnerID)
			return nil
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: medicalRecordID, Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, tagSync, &mockCheckupTransactor{})

	err := svc.Delete(context.Background(), 1, medicalRecordID, 30)

	require.NoError(t, err)
	assert.True(t, syncedAfterDelete)
}
