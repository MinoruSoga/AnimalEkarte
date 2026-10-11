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

const nextVisitRetryTestTag = "next_visit_2026-12-20"

func TestNextVisitRetry_DailyBatchRepairsRemoteOnlyOwner(t *testing.T) {
	for _, failure := range []string{"upsert", "commit"} {
		t.Run(failure, func(t *testing.T) {
			f := newNextVisitRetryFixture(t)
			f.failAt = failure
			require.Error(t, f.service.SyncNextVisitTag(context.Background(), 1, 2))
			require.Empty(t, f.committed)
			require.True(t, f.remote[nextVisitRetryTestTag])
			require.NotZero(t, f.pending)
			f.failAt, f.wanted = "", false
			repo := &expiryBatchRepository{findFn: func(_ context.Context, clinic uint64, _ time.Time, after uint64, limit int) ([]uint64, error) {
				require.Equal(t, uint64(1), clinic)
				require.Equal(t, 200, limit)
				// The SQL contract independently proves that pending work is a
				// candidate even when this owner has no cached next_visit row.
				if f.pending != 0 && after < 2 {
					return []uint64{2}, nil
				}
				return nil, nil
			}}
			batch := &lstepBatchService{reservationRepo: repo, tagSyncSvc: &batchMockTagSyncSvc{syncNextVisitTagFn: f.service.SyncNextVisitTag}}
			count, errs := batch.reconcileExpiredTentativeTags(context.Background(), 1, time.Now())
			require.Empty(t, errs)
			require.Equal(t, 1, count)
			require.Empty(t, f.remote)
			require.Empty(t, f.committed)
			require.Zero(t, f.pending)
		})
	}
}

// Models the service boundary, not PostgreSQL: enqueue commits independently;
// cache writes and acknowledgement are staged until the owned transaction commits.
// Provider state is deliberately never rolled back with the local transaction.
type nextVisitRetryFixture struct {
	cache     *mockLstepTagCacheRepository
	client    *mockLstepAPIClient
	service   *lstepTagSyncService
	committed map[string]bool
	remote    map[string]bool
	pending   uint64
	revision  uint64
	wanted    bool
	failAt    string
	failure   error
	locked    bool
	tx        *gorm.DB
	onAdd     func()
	reads     int
	adds      int
	removes   int
	locks     int
	enqueues  int
	acks      int
}

func newNextVisitRetryFixture(t *testing.T) *nextVisitRetryFixture {
	t.Helper()
	f := &nextVisitRetryFixture{committed: map[string]bool{}, remote: map[string]bool{}, wanted: true, failure: errors.New("injected sync failure"), tx: &gorm.DB{}}
	var staged map[string]bool
	var stagedAck uint64
	owned := func(ctx context.Context, clinic, owner uint64) error {
		require.True(t, f.locked)
		require.Same(t, f.tx, persistence.TxFromContext(ctx))
		require.Equal(t, uint64(1), clinic)
		require.Equal(t, uint64(2), owner)
		return ctx.Err()
	}
	f.cache = &mockLstepTagCacheRepository{
		enqueueNextVisitSyncFn: func(ctx context.Context, clinic, owner uint64) (uint64, error) {
			require.Nil(t, persistence.TxFromContext(ctx), "enqueue must outlive any caller/cache transaction")
			require.False(t, f.locked, "enqueue precedes the owner lock")
			require.Equal(t, uint64(1), clinic)
			require.Equal(t, uint64(2), owner)
			f.enqueues++
			if err := ctx.Err(); err != nil {
				return 0, err
			}
			if f.failAt == "enqueue" {
				return 0, f.failure
			}
			f.revision++
			f.pending = f.revision
			return f.revision, nil
		},
		withNextVisitOwnerLockFn: func(ctx context.Context, clinic, owner uint64, fn func(context.Context) error) error {
			f.locks++
			require.NotZero(t, f.pending, "a durable retry must exist before the transaction opens")
			if f.failAt == "lock" {
				return f.failure
			}
			staged = make(map[string]bool, len(f.committed))
			for tag, present := range f.committed {
				staged[tag] = present
			}
			stagedAck = 0
			f.locked = true
			err := fn(persistence.WithTxValue(ctx, f.tx))
			f.locked = false
			if err != nil {
				return err
			}
			if f.failAt == "commit" {
				return f.failure
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			f.committed = staged
			if stagedAck != 0 && f.pending == stagedAck {
				f.pending = 0
			}
			return nil
		},
		findByOwnerFn: func(ctx context.Context, clinic, owner uint64) ([]*model.LstepTagCache, error) {
			if err := owned(ctx, clinic, owner); err != nil {
				return nil, err
			}
			if f.failAt == "cache read" {
				return nil, f.failure
			}
			var tags []*model.LstepTagCache
			for tag := range staged {
				tags = append(tags, &model.LstepTagCache{TagName: tag})
			}
			return tags, nil
		},
		upsertTagFn: func(ctx context.Context, clinic, owner uint64, tag, _, _ string) error {
			if err := owned(ctx, clinic, owner); err != nil {
				return err
			}
			if f.failAt == "upsert" {
				return f.failure
			}
			staged[tag] = true
			return nil
		},
		deleteTagFn: func(ctx context.Context, clinic, owner uint64, tag string) error {
			if err := owned(ctx, clinic, owner); err != nil {
				return err
			}
			if f.failAt == "cache delete" {
				return f.failure
			}
			delete(staged, tag)
			return nil
		},
		acknowledgeNextVisitSyncFn: func(ctx context.Context, clinic, owner, revision uint64) error {
			if err := owned(ctx, clinic, owner); err != nil {
				return err
			}
			f.acks++
			require.NotZero(t, revision)
			if f.failAt == "ack" {
				return f.failure
			}
			if f.pending == revision {
				stagedAck = revision
			}
			return nil
		},
	}
	f.client = &mockLstepAPIClient{
		getUserTagsFn: func(ctx context.Context, uid string) ([]string, error) {
			if err := owned(ctx, 1, 2); err != nil {
				return nil, err
			}
			require.Equal(t, "mock-only", uid)
			f.reads++
			if f.failAt == "remote read" {
				return nil, f.failure
			}
			var tags []string
			for tag := range f.remote {
				tags = append(tags, tag)
			}
			return tags, nil
		},
		addTagFn: func(ctx context.Context, uid, tag string) error {
			if err := owned(ctx, 1, 2); err != nil {
				return err
			}
			require.Equal(t, "mock-only", uid)
			f.adds++
			if f.failAt == "remote add" {
				return f.failure
			}
			f.remote[tag] = true
			if f.onAdd != nil {
				f.onAdd()
			}
			return nil
		},
		removeTagFn: func(ctx context.Context, uid, tag string) error {
			if err := owned(ctx, 1, 2); err != nil {
				return err
			}
			require.Equal(t, "mock-only", uid)
			f.removes++
			if f.failAt == "remote remove" {
				return f.failure
			}
			delete(f.remote, tag)
			return nil
		},
	}
	date := time.Date(2026, 12, 20, 0, 0, 0, 0, time.UTC)
	f.service = nextVisitReconcileService(f.cache, f.client, &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(ctx context.Context, clinic, owner uint64) (*model.Reservation, error) {
		if err := owned(ctx, clinic, owner); err != nil {
			return nil, err
		}
		if f.failAt == "source read" {
			return nil, f.failure
		}
		if !f.wanted {
			return nil, nil
		}
		return &model.Reservation{Status: model.ReservationStatusTentative, IntendedDate: &date, StartTime: time.Time{}}, nil
	}})
	return f
}

func TestNextVisitRetry_RemoteOnlyTagRemainsDiscoverableUntilReconciled(t *testing.T) {
	for _, mode := range []string{"upsert", "commit", "cancel after add"} {
		t.Run(mode, func(t *testing.T) {
			f := newNextVisitRetryFixture(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			wantErr := f.failure
			if mode == "cancel after add" {
				f.onAdd = cancel
				wantErr = context.Canceled
			} else {
				f.failAt = mode
			}
			// An unrelated caller transaction must never own the durable marker.
			require.ErrorIs(t, f.service.SyncNextVisitTag(persistence.WithTxValue(ctx, &gorm.DB{}), 1, 2), wantErr)
			require.True(t, f.remote[nextVisitRetryTestTag], "the provider side effect succeeded")
			require.Empty(t, f.committed, "cache/ACK transaction must have rolled back")
			require.NotZero(t, f.pending, "daily retry discovery cannot depend on a cached tag")
			if mode != "commit" {
				assert.Zero(t, f.acks, "failure must not acknowledge")
			} else {
				assert.Equal(t, 1, f.acks, "staged acknowledgement rolls back on commit failure")
			}

			// The appointment was removed before retry: remove the remote-only tag,
			// rather than replaying the stale add, and leave unrelated tags alone.
			f.failAt, f.onAdd, f.wanted = "", nil, false
			f.remote["manual_keep"] = true
			require.NoError(t, f.service.SyncNextVisitTag(context.Background(), 1, 2))
			assert.False(t, f.remote[nextVisitRetryTestTag])
			assert.True(t, f.remote["manual_keep"])
			assert.Empty(t, f.committed)
			assert.Zero(t, f.pending, "only completed reconciliation commits ACK")
			assert.Equal(t, 1, f.adds)
			assert.Equal(t, 1, f.removes)
		})
	}
}

func TestNextVisitRetry_NewerEnqueueSurvivesOlderAcknowledgement(t *testing.T) {
	f := newNextVisitRetryFixture(t)
	// Deterministic interleaving: another request durably enqueues while this
	// sync owns the cache lock, before that request can acquire the same lock.
	f.onAdd = func() { f.revision++; f.pending = f.revision }
	require.NoError(t, f.service.SyncNextVisitTag(context.Background(), 1, 2))
	assert.Equal(t, uint64(2), f.pending, "ACK of revision 1 must not delete newer work")
	assert.True(t, f.committed[nextVisitRetryTestTag])
	assert.Equal(t, 1, f.acks)
	f.onAdd, f.wanted = nil, false
	require.NoError(t, f.service.SyncNextVisitTag(context.Background(), 1, 2))
	assert.Zero(t, f.pending)
	assert.Empty(t, f.committed)
	assert.Empty(t, f.remote)
}

func TestNextVisitRetry_FailuresPreserveCommittedCacheAndPendingWork(t *testing.T) {
	for _, mode := range []string{"lock", "source read", "cache read", "remote read", "remote add", "remote remove", "cache delete", "ack"} {
		t.Run(mode, func(t *testing.T) {
			f := newNextVisitRetryFixture(t)
			oldTag := "next_visit_2026-11-01"
			f.committed[oldTag], f.remote[oldTag] = true, true
			f.failAt = mode
			if mode == "remote remove" || mode == "cache delete" {
				f.wanted = false
			}
			require.ErrorIs(t, f.service.SyncNextVisitTag(context.Background(), 1, 2), f.failure)
			assert.Equal(t, map[string]bool{oldTag: true}, f.committed, "failed reconciliation must roll back staged cache writes")
			assert.Equal(t, uint64(1), f.pending)
			if mode == "ack" {
				assert.Equal(t, 1, f.acks)
			} else {
				assert.Zero(t, f.acks)
			}
			if mode == "lock" || mode == "source read" || mode == "cache read" {
				assert.Zero(t, f.reads)
				assert.Zero(t, f.adds)
				assert.Zero(t, f.removes)
			}
			f.failAt = ""
			f.wanted = false
			require.NoError(t, f.service.SyncNextVisitTag(context.Background(), 1, 2))
			assert.Zero(t, f.pending)
			assert.Empty(t, f.committed)
			assert.Empty(t, f.remote)
		})
	}
}

func TestNextVisitRetry_EnqueueFailureMakesNoProviderCalls(t *testing.T) {
	for _, mode := range []string{"enqueue", "already cancelled"} {
		t.Run(mode, func(t *testing.T) {
			f := newNextVisitRetryFixture(t)
			f.pending, f.revision = 7, 7 // Previously queued work is not lost.
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			wantErr := f.failure
			if mode == "already cancelled" {
				cancel()
				wantErr = context.Canceled
			} else {
				f.failAt = mode
			}
			require.ErrorIs(t, f.service.SyncNextVisitTag(ctx, 1, 2), wantErr)
			assert.Equal(t, uint64(7), f.pending)
			assert.Zero(t, f.locks)
			assert.Zero(t, f.reads)
			assert.Zero(t, f.adds)
			assert.Zero(t, f.removes)
			assert.Zero(t, f.acks)
		})
	}
}

func TestNextVisitRetry_SkippedTargetsKeepPendingWork(t *testing.T) {
	for _, gate := range []string{"before lock", "under lock"} {
		for _, mode := range []string{"opted out", "unlinked"} {
			t.Run(gate+"/"+mode, func(t *testing.T) {
				f := newNextVisitRetryFixture(t)
				f.pending, f.revision = 7, 7
				uid := "mock-only"
				f.service.ownerRepo = &mockOwnerRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.Owner, error) {
					skip := gate == "before lock" || f.locked
					if skip && mode == "unlinked" {
						return &model.Owner{ID: 2, LstepOptOut: false, LineUserID: nil}, nil
					}
					return &model.Owner{ID: 2, LstepOptOut: skip && mode == "opted out", LineUserID: &uid}, nil
				}}
				require.NoError(t, f.service.SyncNextVisitTag(context.Background(), 1, 2))
				if gate == "before lock" {
					assert.Equal(t, uint64(7), f.pending)
					assert.Zero(t, f.enqueues)
				} else {
					assert.Equal(t, uint64(8), f.pending)
					assert.Equal(t, 1, f.enqueues)
				}
				assert.Zero(t, f.reads)
				assert.Zero(t, f.adds)
				assert.Zero(t, f.removes)
				assert.Zero(t, f.acks)
			})
		}
	}
}
