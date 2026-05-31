package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	svc := NewPrescriptionService(repo, medRecordRepo, tagSync)

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
	svc := NewPrescriptionService(repo, &mockMedicalRecordRepository{}, tagSync)
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
	svc := NewPrescriptionService(repo, &mockMedicalRecordRepository{}, tagSync)

	err := svc.Delete(context.Background(), 1, medicalRecordID, 30)

	require.NoError(t, err)
	assert.True(t, syncedAfterDelete)
}
