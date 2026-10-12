package lstep

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

type lifecyclePostCommitTraceKey struct{}

type lifecyclePostCommitTransactor struct {
	withTxFn func(context.Context, func(context.Context) error) error
}

func (tx *lifecyclePostCommitTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return tx.withTxFn(ctx, fn)
}

type lifecyclePostCommitAudit struct {
	*mockAuditService
	checkTxFn func(context.Context)
}

func (a *lifecyclePostCommitAudit) LogEntryTx(ctx context.Context, entry *LifecycleAuditEntry) error {
	a.checkTxFn(ctx)
	return a.mockAuditService.LogEntryTx(ctx, entry)
}

func TestPetLifecycle_PostCommitContext(t *testing.T) {
	for _, tt := range []struct {
		name         string
		revival      bool
		allDead      bool
		writeErr     error
		auditErr     error
		cancelBefore bool
	}{
		{name: "death remaining pets survives cancellation"},
		{name: "death all tags cleanup survives cancellation", allDead: true},
		{name: "revival survives cancellation", revival: true},
		{name: "death write failure does not synchronize", writeErr: errors.New("write failed")},
		{name: "revival primary audit failure does not synchronize", revival: true, auditErr: errors.New("audit failed")},
		{name: "clinical transaction still follows request cancellation", cancelBefore: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.WithValue(context.Background(), lifecyclePostCommitTraceKey{}, "lifecycle"))
			defer cancel()
			ambient := &gorm.DB{}
			ctx = persistence.WithTxValue(ctx, ambient)
			if tt.cancelBefore {
				cancel()
			}
			committed := false
			tx := &lifecyclePostCommitTransactor{withTxFn: func(writeCtx context.Context, fn func(context.Context) error) error {
				assert.Same(t, ambient, persistence.TxFromContext(writeCtx))
				if err := writeCtx.Err(); err != nil {
					return err
				}
				if err := fn(writeCtx); err != nil {
					return err
				}
				committed = true
				cancel()
				return nil
			}}
			var done <-chan struct{}
			var observedDone []<-chan struct{}
			var deadline time.Time
			postCalls, syncCalls, cleanupReads := 0, 0, 0
			check := func(postCtx context.Context) {
				require.True(t, committed)
				require.NoError(t, postCtx.Err())
				assert.Nil(t, persistence.TxFromContext(postCtx))
				assert.Equal(t, "lifecycle", postCtx.Value(lifecyclePostCommitTraceKey{}))
				got, ok := postCtx.Deadline()
				require.True(t, ok)
				if postCalls == 0 {
					deadline, done = got, postCtx.Done()
					assert.WithinDuration(t, time.Now().Add(30*time.Second), deadline, time.Second)
				} else {
					assert.Equal(t, deadline, got)
				}
				// Derived contexts may have their own Done channel, but every
				// operation must retain the shared deadline and release its timer.
				observedDone = append(observedDone, postCtx.Done())
				postCalls++
			}
			status := model.PetStatusAlive
			if tt.revival {
				status = model.PetStatusDeceased
			}
			pets := &mockPetRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Pet, error) {
					return &model.Pet{ID: 2, ClinicID: 1, OwnerID: 20, Status: status}, nil
				},
				updateFn: func(writeCtx context.Context, _, _ uint64, _ map[string]any) error {
					assert.Same(t, ambient, persistence.TxFromContext(writeCtx))
					require.NoError(t, writeCtx.Err())
					return tt.writeErr
				},
				findLivingByOwnerFn: func(postCtx context.Context, clinicID, ownerID uint64) ([]model.Pet, error) {
					check(postCtx)
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(20), ownerID)
					if tt.allDead {
						return nil, nil
					}
					return []model.Pet{{ID: 3, ClinicID: 1, OwnerID: 20, Status: model.PetStatusAlive}}, nil
				},
			}
			checkSync := func(postCtx context.Context, clinicID, ownerID uint64) error {
				check(postCtx)
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(20), ownerID)
				syncCalls++
				return errors.New("nonfatal tag failure")
			}
			syncer := &mockLstepTagSyncService{
				syncOwnerAnimalClassificationTagFn: checkSync, syncPetBasicInfoTagsFn: checkSync,
				syncCPMStageTagFn: checkSync, syncNextVisitTagFn: checkSync,
			}
			audit := &lifecyclePostCommitAudit{mockAuditService: &mockAuditService{
				logEntryTxErr: tt.auditErr,
				logLstepOperationFn: func(postCtx context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64) error {
					check(postCtx)
					return errors.New("nonfatal secondary audit failure")
				},
			}, checkTxFn: func(writeCtx context.Context) {
				require.NoError(t, writeCtx.Err())
				assert.Same(t, ambient, persistence.TxFromContext(writeCtx))
			}}
			lineUserID := "synthetic-linked-owner"
			owners := &mockOwnerRepository{findByIDFn: func(postCtx context.Context, _, _ uint64) (*model.Owner, error) {
				check(postCtx)
				return &model.Owner{ID: 20, LineUserID: &lineUserID}, nil
			}}
			settings := &mockLstepSettingsService{isSyncEnabledFn: func(postCtx context.Context, _ uint64) (bool, error) {
				check(postCtx)
				return true, nil // Exercise cleanup reads without constructing an HTTP client.
			}}
			cache := &mockLstepTagCacheRepository{findByOwnerFn: func(postCtx context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
				check(postCtx)
				cleanupReads++
				return nil, nil
			}}
			svc := NewLstepLifecycleService(settings, owners, pets, cache, syncer, audit, nil, tx, audit)
			var err error
			if tt.revival {
				err = svc.HandlePetRevival(ctx, 1, 2, nil)
			} else {
				err = svc.HandlePetDeath(ctx, 1, 2, time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC), "synthetic", nil)
			}
			if tt.writeErr != nil || tt.auditErr != nil || tt.cancelBefore {
				require.Error(t, err)
				assert.False(t, committed)
				assert.Zero(t, postCalls)
				return
			}
			require.NoError(t, err, "secondary tag/audit errors cannot reverse the clinical commit")
			if tt.allDead {
				assert.Zero(t, syncCalls)
				assert.Equal(t, 1, cleanupReads)
			} else {
				assert.Equal(t, 4, syncCalls)
			}
			require.NotNil(t, done)
			select {
			case <-done:
			default:
				t.Fatal("post-commit timer must be cleaned up")
			}
			for _, childDone := range observedDone {
				select {
				case <-childDone:
				default:
					t.Fatal("every derived post-commit context must be cleaned up")
				}
			}
		})
	}
}
