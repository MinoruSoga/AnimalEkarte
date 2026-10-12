package pet

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type petPostCommitTraceKey struct{}

type petPostCommitTransactor struct {
	withTxFn func(context.Context, func(context.Context) error) error
}

func (tx *petPostCommitTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return tx.withTxFn(ctx, fn)
}

func TestPetTransfer_PostCommitContext(t *testing.T) {
	for _, tt := range []struct {
		name         string
		writeErr     error
		commitErr    error
		cancelBefore bool
	}{
		{name: "committed transfer survives cancellation and tag errors"},
		{name: "failed write has no post-commit work", writeErr: errors.New("write failed")},
		{name: "failed commit has no post-commit work", commitErr: errors.New("commit failed")},
		{name: "business transaction still follows request cancellation", cancelBefore: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancelRequest := context.WithCancel(context.WithValue(context.Background(), petPostCommitTraceKey{}, "transfer"))
			defer cancelRequest()
			ambient := &gorm.DB{}
			ctx = persistence.WithTxValue(ctx, ambient)
			if tt.cancelBefore {
				cancelRequest()
			}
			committed := false
			tx := &petPostCommitTransactor{withTxFn: func(txCtx context.Context, fn func(context.Context) error) error {
				assert.Same(t, ambient, persistence.TxFromContext(txCtx))
				if err := txCtx.Err(); err != nil {
					return err
				}
				if err := fn(txCtx); err != nil {
					return err
				}
				if tt.commitErr != nil {
					return tt.commitErr
				}
				committed = true
				cancelRequest()
				return nil
			}}
			current := &model.Pet{ID: 2, ClinicID: 1, OwnerID: 20}
			newOwnerID := uint64(30)
			repo := &mockPetRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) { return current, nil },
				lockByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) { return current, nil },
				updateFn: func(writeCtx context.Context, _, _ uint64, fields map[string]any) error {
					assert.Same(t, ambient, persistence.TxFromContext(writeCtx))
					if tt.writeErr != nil {
						return tt.writeErr
					}
					current = &model.Pet{ID: 2, ClinicID: 1, OwnerID: fields[colPetOwnerID].(uint64)}
					return nil
				},
			}
			var done <-chan struct{}
			var deadline time.Time
			var nextOwners []uint64
			calls := 0
			check := func(syncCtx context.Context, clinicID, ownerID uint64) error {
				require.True(t, committed)
				assert.Equal(t, uint64(1), clinicID)
				assert.Contains(t, []uint64{30, 20}, ownerID)
				require.NoError(t, syncCtx.Err())
				assert.Nil(t, persistence.TxFromContext(syncCtx))
				assert.Equal(t, "transfer", syncCtx.Value(petPostCommitTraceKey{}))
				gotDeadline, ok := syncCtx.Deadline()
				require.True(t, ok)
				if calls == 0 {
					deadline, done = gotDeadline, syncCtx.Done()
					assert.WithinDuration(t, time.Now().Add(30*time.Second), deadline, time.Second)
				} else {
					assert.Equal(t, deadline, gotDeadline, "all tag kinds and owners share one budget")
					assert.Equal(t, done, syncCtx.Done())
				}
				calls++
				return errors.New("nonfatal tag failure")
			}
			tags := &mockLstepTagSyncService{
				syncOwnerAnimalClassificationTagFn: check,
				syncPetBasicInfoTagsFn:             check,
				syncNextVisitTagFn: func(syncCtx context.Context, clinicID, ownerID uint64) error {
					nextOwners = append(nextOwners, ownerID)
					return check(syncCtx, clinicID, ownerID)
				},
			}
			svc := NewServiceWithPetOwnerReader(repo, defaultOwnerRepo(), nil, defaultMedicalRecordRepo(), tags,
				&petOwnerServiceRepositoryDouble{}, tx,
				PetOwnerTransferDependencies{Reservations: &petOwnerTransferSpy{}, Audit: &petOwnerServiceAuditDouble{}})
			got, err := svc.Update(ctx, 1, 2, &UpdatePetInput{OwnerID: &newOwnerID})
			if tt.writeErr != nil || tt.commitErr != nil || tt.cancelBefore {
				require.Error(t, err)
				assert.Nil(t, got)
				assert.False(t, committed)
				assert.Zero(t, calls)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, newOwnerID, got.OwnerID)
			assert.Equal(t, []uint64{30, 20}, nextOwners)
			assert.Equal(t, 6, calls)
			select {
			case <-done:
			default:
				t.Fatal("post-commit context must be canceled on return")
			}
		})
	}
}

func TestPetTags_PostCommitExpiredRequest(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	ctx = persistence.WithTxValue(ctx, &gorm.DB{})
	var done <-chan struct{}
	tags := &mockLstepTagSyncService{syncNextVisitTagFn: func(syncCtx context.Context, _, _ uint64) error {
		require.NoError(t, syncCtx.Err())
		assert.Nil(t, persistence.TxFromContext(syncCtx))
		_, bounded := syncCtx.Deadline()
		assert.True(t, bounded)
		done = syncCtx.Done()
		return nil
	}}
	svc := &petService{tagSyncSvc: tags}
	svc.syncLstepTags(ctx, 1, 20)
	require.NotNil(t, done)
	select {
	case <-done:
	default:
		t.Fatal("post-commit timer must be cleaned up")
	}
}
