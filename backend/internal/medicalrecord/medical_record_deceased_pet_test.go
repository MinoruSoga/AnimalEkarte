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
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// BUG-002: 死亡ペットへの新規カルテ Create を BE で拒否する。

// EMR-491 D6: deceased は業務拒否（InvalidInput）、unresolved / reader 失敗は生存扱いにせずエラーで拒否する。
func TestMedicalRecordService_Create_RejectsDeceasedOrUnresolvedPet(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name        string
		state       sharedkernel.PetLifeState
		readErr     error
		wantInvalid bool
	}{
		{name: "deceased", state: sharedkernel.PetLifeStateDeceased, wantInvalid: true},
		{name: "unresolved", readErr: apperrors.Wrap(sharedkernel.ErrPetLifeStateUnresolved, "pet status is not recorded")},
		{name: "reader failure", readErr: errors.New("db down")},
		{name: "empty state is not alive"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ownerID := uint64(10)
			petID := uint64(20)
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
				readPetLifeStateFn: func(ctx context.Context, clinicID, id uint64) (sharedkernel.PetLifeState, error) {
					assertMedicalRecordTxContext(ctx, t)
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, petID, id)
					return tc.state, tc.readErr
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
			assert.Equal(t, tc.wantInvalid, apperrors.IsInvalidInput(err), "unexpected error class: %v", err)
			if tc.wantInvalid {
				assert.Contains(t, err.Error(), medicalRecordDeceasedPetMessage)
			}
			if tc.readErr != nil {
				assert.ErrorIs(t, err, tc.readErr)
			}
			assert.False(t, createCalled, "non-alive pet must not reach repository Create")
		})
	}
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
		// default ReadPetLifeState returns alive (explicit test double)
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
