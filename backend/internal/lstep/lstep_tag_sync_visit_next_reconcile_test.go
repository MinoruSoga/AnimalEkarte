package lstep

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func nextVisitReconcileService(cache LstepTagCacheRepository, client lstepapi.Client, source tagSyncNextVisitAppointmentRepo) *lstepTagSyncService {
	uid := "mock-only"
	return &lstepTagSyncService{settingsSvc: &mockLstepSettingsService{}, ownerRepo: &mockOwnerRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.Owner, error) {
		return &model.Owner{ID: 2, LineUserID: &uid}, nil
	}}, tagCacheRepo: cache, nextVisitAppointmentRepo: source, buildClientFn: func(context.Context, uint64) (lstepapi.Client, error) { return client, nil }}
}

func TestNextVisitReconcile_CacheFailureAndRemoteTruth(t *testing.T) {
	date := time.Date(2026, 12, 20, 0, 0, 0, 0, time.UTC)
	tag := "next_visit_2026-12-20"
	for _, tt := range []struct {
		name          string
		remotePresent bool
		wantAdds      int
	}{{"missing remote must readd", false, 1}, {"present remote must not repeat add", true, 0}} {
		t.Run(tt.name, func(t *testing.T) {
			desired := false
			present := true
			deleteFailure := errors.New("cache delete failed")
			failDelete := true
			adds := 0
			cache := &mockLstepTagCacheRepository{findByOwnerFn: func(context.Context, uint64, uint64) ([]*model.LstepTagCache, error) {
				return []*model.LstepTagCache{{TagName: tag}}, nil
			}, deleteTagFn: func(context.Context, uint64, uint64, string) error {
				if failDelete {
					return deleteFailure
				}
				return nil
			}}
			client := &mockLstepAPIClient{getUserTagsFn: func(context.Context, string) ([]string, error) {
				if present {
					return []string{tag}, nil
				}
				return nil, nil
			}, removeTagFn: func(context.Context, string, string) error { present = false; return nil }, addTagFn: func(context.Context, string, string) error { adds++; present = true; return nil }}
			svc := nextVisitReconcileService(cache, client, &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
				if desired {
					return &model.Reservation{Status: model.ReservationStatusTentative, IntendedDate: &date}, nil
				}
				return nil, nil
			}})
			assert.ErrorIs(t, svc.SyncNextVisitTag(context.Background(), 1, 2), deleteFailure)
			desired = true
			failDelete = false
			present = tt.remotePresent
			require.NoError(t, svc.SyncNextVisitTag(context.Background(), 1, 2))
			assert.Equal(t, tt.wantAdds, adds)
			assert.True(t, present)
		})
	}
}

func TestNextVisitReconcile_SourceReadUnderOwnerLock(t *testing.T) {
	locked := false
	lockCalls := 0
	cache := &mockLstepTagCacheRepository{withNextVisitOwnerLockFn: func(ctx context.Context, clinic, owner uint64, fn func(context.Context) error) error {
		assert.Equal(t, uint64(7), clinic)
		assert.Equal(t, uint64(9), owner)
		lockCalls++
		locked = true
		defer func() { locked = false }()
		return fn(ctx)
	}}
	svc := nextVisitReconcileService(cache, &mockLstepAPIClient{}, &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
		assert.True(t, locked, "snapshot must be read while owner lock held")
		return nil, nil
	}})
	require.NoError(t, svc.SyncNextVisitTag(context.Background(), 7, 9))
	assert.Equal(t, 1, lockCalls)
}

func TestNextVisitReconcile_FailuresNeverBecomeSuccess(t *testing.T) {
	for _, mode := range []string{"remote read", "cache upsert", "lock acquire", "lock commit"} {
		t.Run(mode, func(t *testing.T) {
			failure := errors.New(mode)
			writes := 0
			resets := 0
			locked := false
			increments := 0
			date := time.Date(2026, 12, 20, 0, 0, 0, 0, time.UTC)
			cache := &mockLstepTagCacheRepository{withNextVisitOwnerLockFn: func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
				if mode == "lock acquire" {
					return failure
				}
				locked = true
				err := fn(ctx)
				locked = false
				if mode == "lock commit" {
					return failure
				}
				return err
			}, upsertTagFn: func(context.Context, uint64, uint64, string, string, string) error {
				if mode == "cache upsert" {
					return failure
				}
				return nil
			}}
			client := &mockLstepAPIClient{getUserTagsFn: func(context.Context, string) ([]string, error) {
				if mode == "remote read" {
					return nil, failure
				}
				return nil, nil
			}, addTagFn: func(context.Context, string, string) error { writes++; return nil }}
			svc := nextVisitReconcileService(cache, client, &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
				return &model.Reservation{Status: model.ReservationStatusTentative, IntendedDate: &date}, nil
			}})
			svc.errorCounterRepo = &mockErrorCounterRepo{incrementFn: func(ctx context.Context, _, _ uint64) (int, error) {
				assert.False(t, locked)
				assert.Nil(t, persistence.TxFromContext(ctx))
				increments++
				return 1, nil
			}, findFn: func(context.Context, uint64, uint64) (*model.LstepSyncErrorCounter, error) {
				resets++
				return &model.LstepSyncErrorCounter{}, nil
			}}
			assert.ErrorIs(t, svc.SyncNextVisitTag(context.Background(), 1, 2), failure)
			assert.Zero(t, resets)
			if mode == "remote read" {
				assert.Equal(t, 1, increments)
			}
			if mode == "remote read" || mode == "lock acquire" {
				assert.Zero(t, writes)
			}
		})
	}
}

func TestNextVisitReconcile_ConcurrentSyncRereadsLatestSource(t *testing.T) {
	var mutex sync.Mutex
	var locked atomic.Bool
	var calls atomic.Int32
	var source atomic.Pointer[model.Reservation]
	firstRead := make(chan struct{})
	releaseFirst := make(chan struct{})
	secondLock := make(chan struct{})
	done := make(chan error, 2)
	old := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	newDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	source.Store(&model.Reservation{Status: model.ReservationStatusTentative, IntendedDate: &old})
	tags := map[string]bool{}
	cache := &mockLstepTagCacheRepository{withNextVisitOwnerLockFn: func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
		if calls.Add(1) == 2 {
			close(secondLock)
		}
		mutex.Lock()
		locked.Store(true)
		defer func() { locked.Store(false); mutex.Unlock() }()
		return fn(ctx)
	}, findByOwnerFn: func(context.Context, uint64, uint64) ([]*model.LstepTagCache, error) {
		var out []*model.LstepTagCache
		for tag := range tags {
			out = append(out, &model.LstepTagCache{TagName: tag})
		}
		return out, nil
	}}
	var reads atomic.Int32
	svc := nextVisitReconcileService(cache, &mockLstepAPIClient{getUserTagsFn: func(context.Context, string) ([]string, error) {
		var out []string
		for tag := range tags {
			out = append(out, tag)
		}
		return out, nil
	}, addTagFn: func(_ context.Context, stringTag, tag string) error { tags[tag] = true; return nil }, removeTagFn: func(_ context.Context, stringTag, tag string) error { delete(tags, tag); return nil }}, &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
		assert.True(t, locked.Load())
		snapshot := source.Load()
		if reads.Add(1) == 1 {
			close(firstRead)
			<-releaseFirst
		}
		return snapshot, nil
	}})
	go func() { done <- svc.SyncNextVisitTag(context.Background(), 1, 2) }()
	<-firstRead
	source.Store(&model.Reservation{Status: model.ReservationStatusTentative, IntendedDate: &newDate})
	go func() { done <- svc.SyncNextVisitTag(context.Background(), 1, 2) }()
	select {
	case <-secondLock:
	case <-time.After(time.Second):
		close(releaseFirst)
		t.Fatal("second sync did not acquire owner lock")
	}
	assert.Equal(t, int32(1), reads.Load())
	close(releaseFirst)
	require.NoError(t, <-done)
	require.NoError(t, <-done)
	assert.Equal(t, map[string]bool{"next_visit_2026-08-01": true}, tags)
}

func TestNextVisitReconcile_PreservesParentBudgetAndCancellation(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		t.Run(map[bool]string{false: "parent deadline", true: "cancelled parent"}[cancelled], func(t *testing.T) {
			parent, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if cancelled {
				cancel()
			}
			ambient := persistence.WithTxValue(parent, &gorm.DB{})
			cache := &mockLstepTagCacheRepository{withNextVisitOwnerLockFn: func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
				assert.Nil(t, persistence.TxFromContext(ctx))
				want, _ := parent.Deadline()
				got, ok := ctx.Deadline()
				assert.True(t, ok)
				assert.Equal(t, want, got)
				if cancelled {
					return ctx.Err()
				}
				return fn(ctx)
			}}
			svc := nextVisitReconcileService(cache, &mockLstepAPIClient{}, &mockNextVisitAppointmentRepo{})
			err := svc.SyncNextVisitTag(ambient, 1, 2)
			if cancelled {
				assert.ErrorIs(t, err, context.Canceled)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNextVisitReconcile_AllTagsCleanupSharesOwnerLock(t *testing.T) {
	var mutex sync.Mutex
	var lockCalls atomic.Int32
	firstRead := make(chan struct{})
	releaseFirst := make(chan struct{})
	clearWaiting := make(chan struct{})
	done := make(chan error, 2)
	tags := map[string]bool{}
	client := &mockLstepAPIClient{getUserTagsFn: func(context.Context, string) ([]string, error) {
		var out []string
		for tag := range tags {
			out = append(out, tag)
		}
		return out, nil
	}, addTagFn: func(_ context.Context, _, tag string) error { tags[tag] = true; return nil }, removeTagFn: func(_ context.Context, _, tag string) error { delete(tags, tag); return nil }}
	cache := &mockLstepTagCacheRepository{withNextVisitOwnerLockFn: func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
		if lockCalls.Add(1) == 2 {
			close(clearWaiting)
		}
		mutex.Lock()
		defer mutex.Unlock()
		return fn(ctx)
	}, findByOwnerFn: func(context.Context, uint64, uint64) ([]*model.LstepTagCache, error) {
		var out []*model.LstepTagCache
		for tag := range tags {
			out = append(out, &model.LstepTagCache{TagName: tag})
		}
		return out, nil
	}}
	date := time.Date(2026, 12, 20, 0, 0, 0, 0, time.UTC)
	svc := nextVisitReconcileService(cache, client, &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
		close(firstRead)
		<-releaseFirst
		return &model.Reservation{Status: model.ReservationStatusTentative, IntendedDate: &date}, nil
	}})
	lifecycle := &lstepLifecycleService{tagCacheRepo: cache}
	go func() { done <- svc.SyncNextVisitTag(context.Background(), 1, 2) }()
	<-firstRead // old alive snapshot is now paused; death commits before cleanup.
	go func() { done <- lifecycle.clearTagsWithOwnerLock(context.Background(), client, 1, 2, "mock-only") }()
	select {
	case <-clearWaiting:
	case <-time.After(time.Second):
		close(releaseFirst)
		t.Fatal("cleanup did not join owner lock")
	}
	close(releaseFirst)
	require.NoError(t, <-done)
	require.NoError(t, <-done)
	assert.Empty(t, tags)
}

// Embed only the public cache interface to deliberately omit the lock capability.
type cacheWithoutNextVisitLock struct{ LstepTagCacheRepository }

func TestNextVisitReconcile_MissingLockFailsClosed(t *testing.T) {
	writes := 0
	reads := 0
	svc := nextVisitReconcileService(&cacheWithoutNextVisitLock{&mockLstepTagCacheRepository{}}, &mockLstepAPIClient{addTagFn: func(context.Context, string, string) error { writes++; return nil }}, &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(context.Context, uint64, uint64) (*model.Reservation, error) { reads++; return nil, nil }})
	assert.Error(t, svc.SyncNextVisitTag(context.Background(), 1, 2))
	assert.Zero(t, writes)
	assert.Zero(t, reads)
}
func TestNextVisitOwnerLock_MissingDatabaseOrScopeFailsClosed(t *testing.T) {
	for _, scope := range []struct{ clinic, owner uint64 }{{1, 2}, {0, 2}, {1, 0}} {
		called := false
		repo := &lstepTagCacheRepository{}
		assert.Error(t, repo.WithNextVisitOwnerLock(context.Background(), scope.clinic, scope.owner, func(context.Context) error { called = true; return nil }))
		assert.False(t, called)
	}
}

func TestNextVisitReconcile_CleanupRepairsRemoteWithoutCache(t *testing.T) {
	for _, mode := range []string{"uncached successful add", "cache deletion failure", "remote read failure"} {
		t.Run(mode, func(t *testing.T) {
			tag := "next_visit_2026-12-20"
			remote := map[string]bool{tag: true, "manual_remote_only": true}
			failure := errors.New(mode)
			failCache := mode == "cache deletion failure"
			removals := 0
			cache := &mockLstepTagCacheRepository{deleteTagFn: func(context.Context, uint64, uint64, string) error {
				if failCache {
					return failure
				}
				return nil
			}}
			client := &mockLstepAPIClient{getUserTagsFn: func(context.Context, string) ([]string, error) {
				if mode == "remote read failure" {
					return nil, failure
				}
				var out []string
				for tag := range remote {
					out = append(out, tag)
				}
				return out, nil
			}, removeTagFn: func(_ context.Context, _, tag string) error { removals++; delete(remote, tag); return nil }}
			svc := &lstepLifecycleService{tagCacheRepo: cache}
			err := svc.clearTagsWithOwnerLock(context.Background(), client, 1, 2, "mock-only")
			if mode == "remote read failure" {
				assert.ErrorIs(t, err, failure)
				assert.Zero(t, removals)
				assert.True(t, remote[tag])
				return
			}
			if failCache {
				assert.ErrorIs(t, err, failure)
				failCache = false
				require.NoError(t, svc.clearTagsWithOwnerLock(context.Background(), client, 1, 2, "mock-only"))
			} else {
				require.NoError(t, err)
			}
			assert.False(t, remote[tag])
			assert.True(t, remote["manual_remote_only"], "untracked categories retain existing policy")
		})
	}
}

func TestNextVisitReconcile_RemoteAddSurvivesPersistenceFailure(t *testing.T) {
	for _, mode := range []string{"upsert", "commit"} {
		t.Run(mode, func(t *testing.T) {
			failure := errors.New(mode)
			fail := true
			remote := false
			adds := 0
			upserts := 0
			date := time.Date(2026, 12, 20, 0, 0, 0, 0, time.UTC)
			cache := &mockLstepTagCacheRepository{withNextVisitOwnerLockFn: func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
				err := fn(ctx)
				if err == nil && fail && mode == "commit" {
					return failure
				}
				return err
			}, upsertTagFn: func(context.Context, uint64, uint64, string, string, string) error {
				upserts++
				if fail && mode == "upsert" {
					return failure
				}
				return nil
			}}
			client := &mockLstepAPIClient{getUserTagsFn: func(context.Context, string) ([]string, error) {
				if remote {
					return []string{"next_visit_2026-12-20"}, nil
				}
				return nil, nil
			}, addTagFn: func(context.Context, string, string) error { adds++; remote = true; return nil }}
			svc := nextVisitReconcileService(cache, client, &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
				return &model.Reservation{Status: model.ReservationStatusTentative, IntendedDate: &date}, nil
			}})
			require.ErrorIs(t, svc.SyncNextVisitTag(context.Background(), 1, 2), failure)
			assert.True(t, remote)
			fail = false
			require.NoError(t, svc.SyncNextVisitTag(context.Background(), 1, 2))
			assert.Equal(t, 1, adds, "provider read prevents repeated add after remote success")
			assert.Equal(t, 2, upserts, "retry repairs cache persistence")
		})
	}
}

func TestNextVisitReconcile_CleanupGetFailureKeepsKnownCacheBestEffort(t *testing.T) {
	readFailure := errors.New("provider tag read failed")
	known := []string{"cpm_core", "next_visit_2026-12-20"}
	remote := map[string]bool{known[0]: true, known[1]: true, "manual_remote_only": true}
	var removed, deleted []string
	cache := &mockLstepTagCacheRepository{findByOwnerFn: func(context.Context, uint64, uint64) ([]*model.LstepTagCache, error) {
		return []*model.LstepTagCache{{TagName: known[0]}, {TagName: known[1]}}, nil
	}, deleteTagFn: func(_ context.Context, _, _ uint64, tag string) error { deleted = append(deleted, tag); return nil }}
	client := &mockLstepAPIClient{getUserTagsFn: func(context.Context, string) ([]string, error) { return nil, readFailure }, removeTagFn: func(_ context.Context, _, tag string) error {
		removed = append(removed, tag)
		delete(remote, tag)
		return nil
	}}
	svc := &lstepLifecycleService{tagCacheRepo: cache}
	assert.ErrorIs(t, svc.clearTagsWithOwnerLock(context.Background(), client, 1, 2, "mock-only"), readFailure, "incomplete provider discovery remains observable")
	assert.ElementsMatch(t, known, removed, "known cache cleanup must continue when remote discovery fails")
	assert.ElementsMatch(t, known, deleted)
	assert.False(t, remote[known[0]])
	assert.False(t, remote[known[1]])
	assert.True(t, remote["manual_remote_only"], "unknown remote categories are untouched")
}

func TestNextVisitReconcile_RevalidatesTargetAfterOwnerLock(t *testing.T) {
	for _, mode := range []string{"opted out", "unlinked", "line changed", "sync disabled", "latest target failure counter"} {
		t.Run(mode, func(t *testing.T) {
			locked := false
			sourceReads := 0
			tagReads := 0
			nextAdds := 0
			counterReads := 0
			counterAdds := 0
			counterIncrements := 0
			oldUID, newUID := "old-mock-only", "new-mock-only"
			tx := &gorm.DB{}
			readFailure := errors.New("tag read failed")
			cache := &mockLstepTagCacheRepository{withNextVisitOwnerLockFn: func(ctx context.Context, _, _ uint64, fn func(context.Context) error) error {
				locked = true // model an opt-out/link/settings commit and cleanup before this waiting sync acquires the lock.
				defer func() { locked = false }()
				return fn(persistence.WithTxValue(ctx, tx))
			}}
			date := time.Date(2026, 12, 20, 0, 0, 0, 0, time.UTC)
			client := &mockLstepAPIClient{getUserTagsFn: func(_ context.Context, uid string) ([]string, error) {
				tagReads++
				assert.Equal(t, newUID, uid)
				if mode == "latest target failure counter" {
					return nil, readFailure
				}
				return nil, nil
			}, addTagFn: func(_ context.Context, uid, tag string) error {
				assert.Equal(t, newUID, uid)
				if tag == lstepErrorTag {
					counterAdds++
				} else {
					nextAdds++
				}
				return nil
			}}
			svc := nextVisitReconcileService(cache, client, &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(ctx context.Context, _, _ uint64) (*model.Reservation, error) {
				sourceReads++
				assert.Same(t, tx, persistence.TxFromContext(ctx))
				return &model.Reservation{Status: model.ReservationStatusTentative, IntendedDate: &date}, nil
			}})
			svc.settingsSvc = &mockLstepSettingsService{isSyncEnabledFn: func(ctx context.Context, _ uint64) (bool, error) {
				if locked {
					assert.Same(t, tx, persistence.TxFromContext(ctx))
					if mode == "sync disabled" {
						return false, nil
					}
				}
				return true, nil
			}}
			svc.ownerRepo = &mockOwnerRepository{findByIDFn: func(ctx context.Context, _, _ uint64) (*model.Owner, error) {
				if !locked {
					return &model.Owner{ID: 2, LineUserID: &oldUID}, nil
				}
				assert.Same(t, tx, persistence.TxFromContext(ctx))
				if mode == "opted out" {
					return &model.Owner{ID: 2, LstepOptOut: true, LineUserID: &newUID}, nil
				}
				if mode == "unlinked" {
					return &model.Owner{ID: 2}, nil
				}
				return &model.Owner{ID: 2, LineUserID: &newUID}, nil
			}}
			svc.errorCounterRepo = &mockErrorCounterRepo{findFn: func(ctx context.Context, _, _ uint64) (*model.LstepSyncErrorCounter, error) {
				counterReads++
				assert.False(t, locked)
				assert.Nil(t, persistence.TxFromContext(ctx))
				return &model.LstepSyncErrorCounter{}, nil
			}, incrementFn: func(ctx context.Context, _, _ uint64) (int, error) {
				counterIncrements++
				assert.False(t, locked)
				assert.Nil(t, persistence.TxFromContext(ctx))
				return lstepSyncErrorThreshold, nil
			}}
			err := svc.SyncNextVisitTag(context.Background(), 1, 2)
			if mode == "latest target failure counter" {
				assert.ErrorIs(t, err, readFailure)
				assert.Equal(t, 1, counterIncrements)
				assert.Equal(t, 1, counterAdds)
				assert.Zero(t, nextAdds)
				return
			}
			require.NoError(t, err)
			if mode == "line changed" {
				assert.Equal(t, 1, sourceReads)
				assert.Equal(t, 1, tagReads)
				assert.Equal(t, 1, nextAdds)
				assert.Equal(t, 1, counterReads)
			} else {
				assert.Zero(t, sourceReads)
				assert.Zero(t, tagReads)
				assert.Zero(t, nextAdds)
				assert.Zero(t, counterReads)
			}
		})
	}
}
