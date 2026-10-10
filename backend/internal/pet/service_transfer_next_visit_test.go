package pet

import (
	"context"
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
