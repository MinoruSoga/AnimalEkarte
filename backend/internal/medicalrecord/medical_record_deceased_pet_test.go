package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// BUG-002: 死亡ペットへの新規カルテ Create を BE で拒否する。

func medicalRecordDeceasedAtFixture() time.Time {
	return time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
}

func TestMedicalRecordService_Create_RejectsDeceasedPet(t *testing.T) {
	t.Parallel()
	ownerID := uint64(10)
	petID := uint64(20)
	deceased := medicalRecordDeceasedAtFixture()
	createCalled := false

	recordRepo := &mockMedicalRecordRepository{
		createFn: func(_ context.Context, _ *model.MedicalRecord) error {
			createCalled = true
			return nil
		},
	}
	reservationRepo := &mockReservationRepoForMedicalRecord{
		findPetOwnerFn: func(_ context.Context, _, id uint64) (uint64, error) {
			if id == petID {
				return ownerID, nil
			}
			return id, nil
		},
		findPetByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
			assertMedicalRecordTxContext(ctx, t)
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, petID, id)
			return &model.Pet{
				ID:         id,
				Status:     model.PetStatusDeceased,
				DeceasedAt: &deceased,
			}, nil
		},
	}
	svc := newMedicalRecordClinicIsolationService(recordRepo, reservationRepo)

	got, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
		Date:    time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerID,
		PetID:   &petID,
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsInvalidInput(err), "want invalid input, got %v", err)
	assert.Contains(t, err.Error(), medicalRecordDeceasedPetMessage)
	assert.False(t, createCalled, "deceased pet must not reach repository Create")
}

func TestMedicalRecordService_Create_AllowsLivingPet(t *testing.T) {
	t.Parallel()
	ownerID := uint64(10)
	petID := uint64(20)
	createCalled := false

	recordRepo := &mockMedicalRecordRepository{
		createFn: func(_ context.Context, r *model.MedicalRecord) error {
			createCalled = true
			r.ID = 99
			return nil
		},
	}
	reservationRepo := &mockReservationRepoForMedicalRecord{
		findPetOwnerFn: func(_ context.Context, _, id uint64) (uint64, error) {
			if id == petID {
				return ownerID, nil
			}
			return id, nil
		},
		// default FindPetByIDInClinic returns living pet (DeceasedAt nil)
	}
	svc := newMedicalRecordClinicIsolationService(recordRepo, reservationRepo)

	got, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
		Date:    time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerID,
		PetID:   &petID,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, createCalled)
	assert.Equal(t, uint64(99), got.ID)
}
