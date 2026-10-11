package pet

import (
	"context"
	"errors"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type petCommitObserver struct {
	inside, committed bool
	commitErr         error
}

func (tx *petCommitObserver) WithTx(ctx context.Context, fn func(context.Context) error) error {
	tx.inside = true
	err := fn(ctx)
	tx.inside = false
	if err == nil {
		err = tx.commitErr
	}
	tx.committed = err == nil
	return err
}

func TestPetDelete_NextVisitSyncUsesLockedOwnerAfterCommit(t *testing.T) {
	tx := &petCommitObserver{}
	repo := &mockPetRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
			return &model.Pet{ID: 2, ClinicID: 1, OwnerID: 10}, nil
		},
		lockByIDFn: func(_ context.Context, clinicID, petID uint64) (*model.Pet, error) {
			require.True(t, tx.inside)
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(2), petID)
			return &model.Pet{ID: 2, ClinicID: 1, OwnerID: 20}, nil
		},
		deleteFn: func(context.Context, uint64, uint64) error { require.True(t, tx.inside); return nil },
	}
	var owners []uint64
	tags := &mockLstepTagSyncService{syncNextVisitTagFn: func(_ context.Context, clinicID, ownerID uint64) error {
		require.True(t, tx.committed)
		require.False(t, tx.inside)
		assert.Equal(t, uint64(1), clinicID)
		owners = append(owners, ownerID)
		return nil
	}}
	records := &mockMedicalRecordRepository{countByPetIDFn: func(context.Context, uint64, uint64) (int64, error) {
		require.True(t, tx.inside)
		return 0, nil
	}}
	svc := NewServiceWithPetOwnerReader(repo, defaultOwnerRepo(), nil, records, tags, nil, tx)
	require.NoError(t, svc.Delete(context.Background(), 1, 2))
	assert.Equal(t, []uint64{20}, owners)
}

func TestPetDelete_FailuresDoNotSync(t *testing.T) {
	for _, failure := range []string{"lock", "nil pet", "records", "record read", "delete", "commit", "no transaction"} {
		t.Run(failure, func(t *testing.T) {
			tx := &petCommitObserver{}
			deletes := 0
			repo := &mockPetRepository{
				lockByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
					require.True(t, tx.inside)
					if failure == "lock" {
						return nil, errors.New("lock failed")
					}
					if failure == "nil pet" {
						return nil, nil
					}
					return &model.Pet{ID: 2, ClinicID: 1, OwnerID: 20}, nil
				},
				deleteFn: func(context.Context, uint64, uint64) error {
					require.True(t, tx.inside)
					deletes++
					if failure == "delete" {
						return errors.New("delete failed")
					}
					return nil
				},
			}
			records := &mockMedicalRecordRepository{countByPetIDFn: func(context.Context, uint64, uint64) (int64, error) {
				require.True(t, tx.inside)
				if failure == "records" {
					return 1, nil
				}
				if failure == "record read" {
					return 0, errors.New("record read failed")
				}
				return 0, nil
			}}
			if failure == "commit" {
				tx.commitErr = errors.New("commit failed")
			}
			synced := 0
			tags := &mockLstepTagSyncService{syncNextVisitTagFn: func(context.Context, uint64, uint64) error { synced++; return nil }}
			var transactor PetOwnerTransactor = tx
			if failure == "no transaction" {
				transactor = nil
			}
			svc := NewServiceWithPetOwnerReader(repo, defaultOwnerRepo(), nil, records, tags, nil, transactor)
			require.Error(t, svc.Delete(context.Background(), 1, 2))
			assert.Zero(t, synced)
			assert.False(t, tx.committed)
			if failure != "delete" && failure != "commit" {
				assert.Zero(t, deletes)
			}
		})
	}
}

func TestPetDelete_TagFailureDoesNotInvertCommittedDelete(t *testing.T) {
	tx := &petCommitObserver{}
	repo := &mockPetRepository{
		lockByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) { return &model.Pet{ID: 2, OwnerID: 20}, nil },
		deleteFn:   func(context.Context, uint64, uint64) error { return nil },
	}
	synced := 0
	tags := &mockLstepTagSyncService{syncNextVisitTagFn: func(context.Context, uint64, uint64) error {
		require.True(t, tx.committed)
		synced++
		return errors.New("tag sync failed")
	}}
	svc := NewServiceWithPetOwnerReader(repo, defaultOwnerRepo(), nil, defaultMedicalRecordRepo(), tags, nil, tx)
	require.NoError(t, svc.Delete(context.Background(), 1, 2))
	assert.Equal(t, 1, synced)
}
