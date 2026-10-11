package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type lifecycleCommitObserver struct {
	inside    bool
	committed bool
	commitErr error
}

func (tx *lifecycleCommitObserver) WithTx(ctx context.Context, fn func(context.Context) error) error {
	tx.inside = true
	err := fn(ctx)
	tx.inside = false
	if err == nil {
		err = tx.commitErr
	}
	tx.committed = err == nil
	return err
}

func TestLifecycle_FailedTransitionDoesNotSync(t *testing.T) {
	for _, revival := range []bool{false, true} {
		name := "death"
		status := model.PetStatusAlive
		if revival {
			name = "revival"
			status = model.PetStatusDeceased
		}
		for _, failure := range []string{"lock", "nil pet", "state changed", "write", "audit", "commit"} {
			t.Run(name+"/"+failure, func(t *testing.T) {
				tx := &lifecycleCommitObserver{}
				writes := 0
				repo := &mockPetRepository{
					lockByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
						require.True(t, tx.inside)
						if failure == "lock" {
							return nil, errors.New("lock failed")
						}
						if failure == "nil pet" {
							return nil, nil
						}
						lockedStatus := status
						if failure == "state changed" {
							if revival {
								lockedStatus = model.PetStatusAlive
							} else {
								lockedStatus = model.PetStatusDeceased
							}
						}
						return &model.Pet{ID: 2, ClinicID: 1, OwnerID: 20, Status: lockedStatus}, nil
					},
					updateFn: func(context.Context, uint64, uint64, map[string]any) error {
						require.True(t, tx.inside)
						writes++
						if failure == "write" {
							return errors.New("write failed")
						}
						return nil
					},
				}
				audit := &mockAuditService{}
				if failure == "audit" {
					audit.logEntryTxErr = errors.New("audit failed")
				}
				if failure == "commit" {
					tx.commitErr = errors.New("commit failed")
				}
				synced := 0
				syncer := &mockLstepTagSyncService{syncNextVisitTagFn: func(context.Context, uint64, uint64) error { synced++; return nil }}
				svc := NewLstepLifecycleService(defaultLstepSettingsSvc(), &mockOwnerRepository{}, repo,
					&mockLstepTagCacheRepository{}, syncer, audit, nil, tx, audit)
				var err error
				if revival {
					err = svc.HandlePetRevival(context.Background(), 1, 2, nil)
				} else {
					err = svc.HandlePetDeath(context.Background(), 1, 2, time.Now(), "", nil)
				}
				require.Error(t, err)
				assert.False(t, tx.committed)
				assert.Zero(t, synced)
				if failure == "lock" || failure == "nil pet" || failure == "state changed" {
					assert.Zero(t, writes)
				}
			})
		}
	}
}

func TestLifecycle_DeathAllPetsDeadChecksLockedOwnerAndDoesNotReaddNextVisit(t *testing.T) {
	tx := &lifecycleCommitObserver{}
	repo := &mockPetRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
			return &model.Pet{ID: 2, ClinicID: 1, OwnerID: 10, Status: model.PetStatusAlive}, nil
		},
		lockByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
			return &model.Pet{ID: 2, ClinicID: 1, OwnerID: 20, Status: model.PetStatusAlive}, nil
		},
		findLivingByOwnerFn: func(_ context.Context, _, ownerID uint64) ([]model.Pet, error) {
			require.True(t, tx.committed)
			assert.Equal(t, uint64(20), ownerID)
			return nil, nil
		},
	}
	checkedOwners := []uint64{}
	owners := &mockOwnerRepository{findByIDFn: func(_ context.Context, _, ownerID uint64) (*model.Owner, error) {
		checkedOwners = append(checkedOwners, ownerID)
		return &model.Owner{ID: ownerID}, nil // unlinked: no external client is created
	}}
	synced := 0
	syncer := &mockLstepTagSyncService{syncNextVisitTagFn: func(context.Context, uint64, uint64) error { synced++; return nil }}
	audit := &mockAuditService{}
	svc := NewLstepLifecycleService(defaultLstepSettingsSvc(), owners, repo, &mockLstepTagCacheRepository{}, syncer, audit, nil, tx, audit)
	require.NoError(t, svc.HandlePetDeath(context.Background(), 1, 2, time.Now(), "", nil))
	assert.Equal(t, []uint64{20}, checkedOwners)
	assert.Zero(t, synced, "all-dead clear path must not add next-visit tags again")
}

func TestLifecycle_NextVisitUsesLockedOwnerAfterTransfer(t *testing.T) {
	for _, revival := range []bool{false, true} {
		name := "death"
		status := model.PetStatusAlive
		if revival {
			name = "revival"
			status = model.PetStatusDeceased
		}
		t.Run(name, func(t *testing.T) {
			tx := &lifecycleCommitObserver{}
			repo := &mockPetRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
					return &model.Pet{ID: 2, ClinicID: 1, OwnerID: 10, Status: status}, nil
				},
				lockByIDFn: func(_ context.Context, clinicID, petID uint64) (*model.Pet, error) {
					require.True(t, tx.inside)
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(2), petID)
					// Transfer A -> B completed before this operation acquired the pet lock.
					return &model.Pet{ID: 2, ClinicID: 1, OwnerID: 20, Status: status}, nil
				},
				findLivingByOwnerFn: func(_ context.Context, _, ownerID uint64) ([]model.Pet, error) {
					assert.Equal(t, uint64(20), ownerID)
					return []model.Pet{{ID: 3, OwnerID: 20, Status: model.PetStatusAlive}}, nil
				},
			}
			var owners []uint64
			syncer := &mockLstepTagSyncService{syncNextVisitTagFn: func(_ context.Context, clinicID, ownerID uint64) error {
				require.True(t, tx.committed)
				require.False(t, tx.inside)
				assert.Equal(t, uint64(1), clinicID)
				owners = append(owners, ownerID)
				return nil
			}}
			audit := &mockAuditService{}
			svc := NewLstepLifecycleService(defaultLstepSettingsSvc(), &mockOwnerRepository{}, repo,
				&mockLstepTagCacheRepository{}, syncer, audit, nil, tx, audit)
			var err error
			if revival {
				err = svc.HandlePetRevival(context.Background(), 1, 2, nil)
			} else {
				err = svc.HandlePetDeath(context.Background(), 1, 2, time.Now(), "", nil)
			}
			require.NoError(t, err)
			assert.Equal(t, []uint64{20}, owners)
		})
	}
}
