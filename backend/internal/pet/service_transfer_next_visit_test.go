package pet

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestPetTransfer_NextVisitSyncUsesLockedPreviousOwner(t *testing.T) {
	const clinicID, petID = uint64(1), uint64(2)
	newOwnerID := uint64(30)
	current := &model.Pet{ID: petID, ClinicID: clinicID, OwnerID: 10}
	locked := false
	repo := &mockPetRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) { return current, nil },
		lockByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
			locked = true
			current = &model.Pet{ID: petID, ClinicID: clinicID, OwnerID: 20}
			return current, nil // A concurrent transfer won after the initial read.
		},
		updateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
			require.True(t, locked)
			current = &model.Pet{ID: petID, ClinicID: clinicID, OwnerID: fields[colPetOwnerID].(uint64)}
			return nil
		},
	}
	var synced []uint64
	tags := &mockLstepTagSyncService{syncNextVisitTagFn: func(_ context.Context, gotClinic, ownerID uint64) error {
		assert.Equal(t, clinicID, gotClinic)
		synced = append(synced, ownerID)
		return nil
	}}
	svc := NewServiceWithPetOwnerReader(repo, defaultOwnerRepo(), nil, defaultMedicalRecordRepo(), tags,
		&petOwnerServiceRepositoryDouble{}, petOwnerServiceTransactorDouble{})
	updated, err := svc.Update(context.Background(), clinicID, petID, &UpdatePetInput{OwnerID: &newOwnerID})
	require.NoError(t, err)
	assert.Equal(t, newOwnerID, updated.OwnerID)
	assert.ElementsMatch(t, []uint64{20, newOwnerID}, synced, "resync the actual replaced owner and the new owner")
	assert.NotContains(t, synced, uint64(10), "do not use the stale pre-transaction owner snapshot")
}

func TestPetTransfer_SameOwnerSyncsOnceAfterCommit(t *testing.T) {
	tx := &petCommitObserver{}
	ownerID := uint64(20)
	repo := &mockPetRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
			return &model.Pet{ID: 2, ClinicID: 1, OwnerID: ownerID}, nil
		},
		lockByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
			require.True(t, tx.inside)
			return &model.Pet{ID: 2, ClinicID: 1, OwnerID: ownerID}, nil
		},
	}
	synced := 0
	tags := &mockLstepTagSyncService{syncNextVisitTagFn: func(_ context.Context, clinicID, gotOwnerID uint64) error {
		require.True(t, tx.committed)
		assert.Equal(t, uint64(1), clinicID)
		assert.Equal(t, ownerID, gotOwnerID)
		synced++
		return nil
	}}
	svc := NewServiceWithPetOwnerReader(repo, defaultOwnerRepo(), nil, defaultMedicalRecordRepo(), tags,
		&petOwnerServiceRepositoryDouble{}, tx)
	_, err := svc.Update(context.Background(), 1, 2, &UpdatePetInput{OwnerID: &ownerID})
	require.NoError(t, err)
	assert.Equal(t, 1, synced)
}

func TestPetTransfer_FailedWriteDoesNotSync(t *testing.T) {
	tx := &petCommitObserver{}
	newOwnerID := uint64(30)
	repo := &mockPetRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
			return &model.Pet{ID: 2, ClinicID: 1, OwnerID: 10}, nil
		},
		lockByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
			return &model.Pet{ID: 2, ClinicID: 1, OwnerID: 20}, nil
		},
		updateFn: func(context.Context, uint64, uint64, map[string]any) error {
			require.True(t, tx.inside)
			return errors.New("write failed")
		},
	}
	synced := 0
	tags := &mockLstepTagSyncService{syncNextVisitTagFn: func(context.Context, uint64, uint64) error {
		synced++
		return nil
	}}
	svc := NewServiceWithPetOwnerReader(repo, defaultOwnerRepo(), nil, defaultMedicalRecordRepo(), tags,
		&petOwnerServiceRepositoryDouble{}, tx)
	_, err := svc.Update(context.Background(), 1, 2, &UpdatePetInput{OwnerID: &newOwnerID})
	require.Error(t, err)
	assert.False(t, tx.committed)
	assert.Zero(t, synced)
}
