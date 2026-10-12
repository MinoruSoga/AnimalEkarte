package lstep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	ownerdomain "github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/persistence"
	petdomain "github.com/animal-ekarte/backend/internal/pet"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupNextVisitCleanupRealDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LstepTagCache{}, &model.LstepSyncErrorCounter{}))
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS lstep_next_visit_sync_retries (
		clinic_id BIGINT NOT NULL,
		owner_id BIGINT NOT NULL,
		revision BIGSERIAL NOT NULL,
		cleanup_line_user_id VARCHAR(64),
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (clinic_id, owner_id)
	)`).Error)
	require.NoError(t, db.Exec(`ALTER TABLE lstep_next_visit_sync_retries
		ADD COLUMN IF NOT EXISTS cleanup_line_user_id VARCHAR(64)`).Error)
	require.NoError(t, db.Exec(`DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_lstep_next_visit_sync_retries_cleanup_line_user_id') THEN
		ALTER TABLE lstep_next_visit_sync_retries
			ADD CONSTRAINT chk_lstep_next_visit_sync_retries_cleanup_line_user_id
			CHECK (cleanup_line_user_id IS NULL OR cleanup_line_user_id ~ '^[A-Za-z0-9_-]{1,64}$');
	END IF;
END $$`).Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE lstep_next_visit_sync_retries").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE lstep_tag_cache").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE lstep_sync_error_counters").Error)
	testdb.SeedClinicsForFK(t, db, 1, 2)
	require.NoError(t, db.Exec(`DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uq_owners_clinic_id_id') THEN
		ALTER TABLE owners ADD CONSTRAINT uq_owners_clinic_id_id UNIQUE (clinic_id, id);
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'lstep_next_visit_sync_retries_clinic_id_fkey') THEN
		ALTER TABLE lstep_next_visit_sync_retries
			ADD CONSTRAINT lstep_next_visit_sync_retries_clinic_id_fkey
			FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_lstep_next_visit_sync_retries_owner_clinic') THEN
		ALTER TABLE lstep_next_visit_sync_retries
			ADD CONSTRAINT fk_lstep_next_visit_sync_retries_owner_clinic
			FOREIGN KEY (clinic_id, owner_id) REFERENCES owners(clinic_id, id) ON DELETE RESTRICT;
	END IF;
END $$`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_test_lstep_tag_cache_conflict
		ON lstep_tag_cache (clinic_id, owner_id, tag_name)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_test_lstep_sync_error_counter_conflict
		ON lstep_sync_error_counters (clinic_id, owner_id)`).Error)
	return db
}

func newCleanupRealRepo(db *gorm.DB) *lstepTagCacheRepository {
	return &lstepTagCacheRepository{db: db}
}

func makeCleanupOwner(t *testing.T, db *gorm.DB, clinicID uint64, name, lineUserID string, deleted bool) *model.Owner {
	t.Helper()
	owner := testdb.MakeTestOwner(t, db, clinicID, name)
	if lineUserID != "" {
		require.NoError(t, db.Exec("UPDATE owners SET line_user_id = ? WHERE id = ?", lineUserID, owner.ID).Error)
		owner.LineUserID = &lineUserID
	}
	if deleted {
		require.NoError(t, db.Delete(owner).Error)
	}
	return owner
}

func deleteWithCleanupIntent(t *testing.T, db *gorm.DB, repo *lstepTagCacheRepository, clinicID uint64, owner *model.Owner, uid string) {
	t.Helper()
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(context.Background(), tx)
		if err := repo.EnqueueNextVisitCleanup(txCtx, clinicID, owner.ID, uid); err != nil {
			return err
		}
		return tx.Delete(owner).Error
	}))
}

func retryRowExists(t *testing.T, db *gorm.DB, clinicID, ownerID uint64) bool {
	t.Helper()
	var exists bool
	require.NoError(t, db.Raw(`SELECT EXISTS(
		SELECT 1 FROM lstep_next_visit_sync_retries WHERE clinic_id = ? AND owner_id = ?)`,
		clinicID, ownerID).Scan(&exists).Error)
	return exists
}

func cacheTagExists(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, tag string) bool {
	t.Helper()
	var exists bool
	require.NoError(t, db.Raw(`SELECT EXISTS(
		SELECT 1 FROM lstep_tag_cache WHERE clinic_id = ? AND owner_id = ? AND tag_name = ?)`,
		clinicID, ownerID, tag).Scan(&exists).Error)
	return exists
}

func lineRecipientLockKey(clinicID uint64, lineUserID string) string {
	return fmt.Sprintf("lstep-line-recipient:%d:%s", clinicID, lineUserID)
}

func waitForExactAdvisoryWaiter(t *testing.T, db *gorm.DB, key string) int64 {
	t.Helper()
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	poll := time.NewTicker(10 * time.Millisecond)
	defer poll.Stop()
	for {
		var pid int64
		require.NoError(t, db.Raw(`SELECT pid FROM pg_locks
			WHERE locktype = 'advisory' AND NOT granted
			  AND (classid::bigint << 32 | objid::bigint) = hashtextextended(?, 0)
			LIMIT 1`, key).Scan(&pid).Error)
		if pid != 0 {
			return pid
		}
		select {
		case <-poll.C:
		case <-deadline.C:
			require.Fail(t, "no session waited on the exact advisory key before the deadline")
			return 0
		}
	}
}

func waitForBackendBlockedBy(t *testing.T, db *gorm.DB, blockerPID int64, queryLike string, done <-chan error) {
	t.Helper()
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	poll := time.NewTicker(10 * time.Millisecond)
	defer poll.Stop()
	for {
		var blocked bool
		require.NoError(t, db.Raw(`SELECT EXISTS(
			SELECT 1 FROM pg_stat_activity
			WHERE datname = current_database()
			  AND pid <> pg_backend_pid()
			  AND query LIKE ?
			  AND ? = ANY (pg_blocking_pids(pid)))`, queryLike, blockerPID).Scan(&blocked).Error)
		if blocked {
			return
		}
		select {
		case err := <-done:
			require.Failf(t, "the operation finished without blocking on the expected backend", "err=%v", err)
		case <-poll.C:
		case <-deadline.C:
			require.Fail(t, "no backend blocked by the expected PID before the deadline")
		}
	}
}

type cleanupProviderSpy struct {
	getCalls int
	addCalls int
	removed  []string
	tags     []string
	getErr   error
	getFn    func() ([]string, error)
	removeFn func(ctx context.Context, tag string) error
}

func (s *cleanupProviderSpy) client() lstepapi.Client {
	return &mockLstepAPIClient{
		getUserTagsFn: func(context.Context, string) ([]string, error) {
			s.getCalls++
			if s.getFn != nil {
				return s.getFn()
			}
			if s.getErr != nil {
				return nil, s.getErr
			}
			return append([]string(nil), s.tags...), nil
		},
		addTagFn: func(context.Context, string, string) error {
			s.addCalls++
			return errors.New("cleanup must never add tags")
		},
		removeTagFn: func(ctx context.Context, _ string, tag string) error {
			if s.removeFn != nil {
				if err := s.removeFn(ctx, tag); err != nil {
					return err
				}
			}
			s.removed = append(s.removed, tag)
			for i, existing := range s.tags {
				if existing == tag {
					s.tags = append(s.tags[:i], s.tags[i+1:]...)
				}
			}
			return nil
		},
	}
}

func newCleanupRealService(repo *lstepTagCacheRepository, spy *cleanupProviderSpy) *lstepTagSyncService {
	return &lstepTagSyncService{
		settingsSvc:  &mockLstepSettingsService{},
		ownerRepo:    ownerdomain.NewRepository(repo.db, nil),
		tagCacheRepo: repo,
		nextVisitAppointmentRepo: &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
			return nil, errors.New("cleanup must never read appointments")
		}},
		buildClientFn: func(context.Context, uint64) (lstepapi.Client, error) { return spy.client(), nil },
	}
}

type cleanupStoreHook struct {
	*lstepTagCacheRepository
	deleteTagHook   func(tag string) error
	postCallbackErr error
}

func (h *cleanupStoreHook) DeleteTag(ctx context.Context, clinicID, ownerID uint64, tagName string) error {
	if h.deleteTagHook != nil {
		if err := h.deleteTagHook(tagName); err != nil {
			return err
		}
	}
	return h.lstepTagCacheRepository.DeleteTag(ctx, clinicID, ownerID, tagName)
}

func (h *cleanupStoreHook) WithNextVisitOwnerLock(ctx context.Context, clinicID, ownerID uint64, fn func(context.Context) error) error {
	inner := fn
	if h.postCallbackErr != nil {
		inner = func(syncCtx context.Context) error {
			if err := fn(syncCtx); err != nil {
				return err
			}
			return h.postCallbackErr
		}
	}
	return h.lstepTagCacheRepository.WithNextVisitOwnerLock(ctx, clinicID, ownerID, inner)
}

func TestNextVisitCleanupEnqueue_RepositoryContract(t *testing.T) {
	db := setupNextVisitCleanupRealDB(t)
	repo := newCleanupRealRepo(db)
	const clinicID = uint64(1)

	t.Run("requires an ambient transaction", func(t *testing.T) {
		owner := makeCleanupOwner(t, db, clinicID, "no tx", "Uno-tx", false)
		err := repo.EnqueueNextVisitCleanup(context.Background(), clinicID, owner.ID, "Uno-tx")
		require.Error(t, err)
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, "INTERNAL", appErr.Code)
	})

	t.Run("rejects malformed recipients before persistence", func(t *testing.T) {
		owner := makeCleanupOwner(t, db, clinicID, "bad recipient", "Ubad-target", false)
		for _, bad := range []string{"has space", "exclaim!!", " Upad-left", "Upad-right ", fmt.Sprintf("%065d", 0)} {
			require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
				err := repo.EnqueueNextVisitCleanup(
					persistence.WithTxValue(context.Background(), tx), clinicID, owner.ID, bad)
				require.Error(t, err)
				var appErr *apperrors.AppError
				require.ErrorAs(t, err, &appErr)
				assert.Equal(t, "INVALID_INPUT", appErr.Code, "raw recipient %q must not be trimmed into validity", bad)
				return nil
			}))
		}
		assert.False(t, retryRowExists(t, db, clinicID, owner.ID))
	})

	t.Run("revalidates the locked owner and current recipient", func(t *testing.T) {
		owner := makeCleanupOwner(t, db, clinicID, "mismatch", "Uactual", false)
		other := makeCleanupOwner(t, db, uint64(2), "foreign", "Uother", false)
		require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
			txCtx := persistence.WithTxValue(context.Background(), tx)
			err := repo.EnqueueNextVisitCleanup(txCtx, clinicID, owner.ID, "Udifferent")
			assert.True(t, apperrors.IsConflict(err), "stale recipient must conflict, got %v", err)
			err = repo.EnqueueNextVisitCleanup(txCtx, clinicID, other.ID, "Uother")
			require.Error(t, err)
			var appErr *apperrors.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, "NOT_FOUND", appErr.Code, "cross-clinic owner is invisible to the scoped lock, got %v", err)
			return nil
		}))
		assert.False(t, retryRowExists(t, db, clinicID, owner.ID))
	})

	t.Run("rollback removes the intent and commit persists it", func(t *testing.T) {
		owner := makeCleanupOwner(t, db, clinicID, "tx semantics", "Utx-semantics", false)
		require.Error(t, db.Transaction(func(tx *gorm.DB) error {
			require.NoError(t, repo.EnqueueNextVisitCleanup(
				persistence.WithTxValue(context.Background(), tx), clinicID, owner.ID, "Utx-semantics"))
			return errors.New("business rollback")
		}))
		assert.False(t, retryRowExists(t, db, clinicID, owner.ID), "rolled-back enqueue leaves no intent")

		require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
			return repo.EnqueueNextVisitCleanup(
				persistence.WithTxValue(context.Background(), tx), clinicID, owner.ID, "Utx-semantics")
		}))
		intent, err := repo.FindNextVisitCleanupIntent(context.Background(), clinicID, owner.ID)
		require.NoError(t, err)
		require.NotNil(t, intent)
		assert.Equal(t, "Utx-semantics", intent.LineUserID)
		assert.NotZero(t, intent.Revision)
	})

	t.Run("normal enqueue never clears the cleanup intent", func(t *testing.T) {
		owner := makeCleanupOwner(t, db, clinicID, "preserve intent", "Upreserve", false)
		revision, err := repo.EnqueueNextVisitSync(context.Background(), clinicID, owner.ID)
		require.NoError(t, err)
		require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
			return repo.EnqueueNextVisitCleanup(
				persistence.WithTxValue(context.Background(), tx), clinicID, owner.ID, "Upreserve")
		}))
		laterRevision, err := repo.EnqueueNextVisitSync(context.Background(), clinicID, owner.ID)
		require.NoError(t, err)
		assert.Greater(t, laterRevision, revision)
		intent, err := repo.FindNextVisitCleanupIntent(context.Background(), clinicID, owner.ID)
		require.NoError(t, err)
		require.NotNil(t, intent, "a later normal enqueue must not clear the deletion intent")
		assert.Equal(t, "Upreserve", intent.LineUserID)
	})
}

func TestAcquireLineRecipientLock_CancelledWaitIsBounded(t *testing.T) {
	db := setupNextVisitCleanupRealDB(t)
	const clinicID = uint64(1)
	uid := "Uheld-recipient"

	holder := db.Begin()
	require.NoError(t, holder.Error)
	t.Cleanup(func() { _ = holder.Rollback().Error })
	require.NoError(t, holder.Exec(
		"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", lineRecipientLockKey(clinicID, uid)).Error)

	waiter := db.Begin()
	require.NoError(t, waiter.Error)
	t.Cleanup(func() { _ = waiter.Rollback().Error })

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := persistence.AcquireLineRecipientLock(persistence.WithTxValue(ctx, waiter), clinicID, uid)
	require.Error(t, err)
	assert.Less(t, time.Since(start), 10*time.Second,
		"a cancelled recipient-lock wait must return through tx.WithContext instead of hanging")
	assert.True(t,
		errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled),
		"expected a context error, got %v", err)

	cancelled, cancelNow := context.WithCancel(context.Background())
	cancelNow()
	waiter2 := db.Begin()
	require.NoError(t, waiter2.Error)
	t.Cleanup(func() { _ = waiter2.Rollback().Error })
	err = persistence.AcquireLineRecipientLock(persistence.WithTxValue(cancelled, waiter2), clinicID, uid)
	require.Error(t, err)
}

func TestNextVisitCleanup_LockTransactionAtomicity(t *testing.T) {
	db := setupNextVisitCleanupRealDB(t)
	repo := newCleanupRealRepo(db)
	const clinicID = uint64(1)

	t.Run("callback failure rolls back cache and acknowledgement together", func(t *testing.T) {
		owner := makeCleanupOwner(t, db, clinicID, "atomicity", "Uatomic", false)
		require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
			return repo.EnqueueNextVisitCleanup(
				persistence.WithTxValue(context.Background(), tx), clinicID, owner.ID, "Uatomic")
		}))
		require.NoError(t, repo.UpsertTag(context.Background(), clinicID, owner.ID, "next_visit_2026-11-01", "auto", ""))
		intent, err := repo.FindNextVisitCleanupIntent(context.Background(), clinicID, owner.ID)
		require.NoError(t, err)
		require.NotNil(t, intent)

		sentinel := errors.New("post acknowledgement failure")
		err = repo.WithNextVisitOwnerLock(context.Background(), clinicID, owner.ID, func(syncCtx context.Context) error {
			require.NoError(t, repo.DeleteTag(syncCtx, clinicID, owner.ID, "next_visit_2026-11-01"))
			require.NoError(t, repo.AcknowledgeNextVisitSync(syncCtx, clinicID, owner.ID, intent.Revision))
			return sentinel
		})
		require.ErrorIs(t, err, sentinel)
		assert.True(t, cacheTagExists(t, db, clinicID, owner.ID, "next_visit_2026-11-01"),
			"cache deletion must roll back with the acknowledgement")
		assert.True(t, retryRowExists(t, db, clinicID, owner.ID), "the intent must survive the rollback")

		require.NoError(t, repo.WithNextVisitOwnerLock(context.Background(), clinicID, owner.ID, func(syncCtx context.Context) error {
			require.NoError(t, repo.DeleteTag(syncCtx, clinicID, owner.ID, "next_visit_2026-11-01"))
			return repo.AcknowledgeNextVisitSync(syncCtx, clinicID, owner.ID, intent.Revision)
		}))
		assert.False(t, cacheTagExists(t, db, clinicID, owner.ID, "next_visit_2026-11-01"))
		assert.False(t, retryRowExists(t, db, clinicID, owner.ID))
	})

	t.Run("a newer enqueue survives an older acknowledgement", func(t *testing.T) {
		owner := makeCleanupOwner(t, db, clinicID, "cas survivor", "Ucas", false)
		oldRevision, err := repo.EnqueueNextVisitSync(context.Background(), clinicID, owner.ID)
		require.NoError(t, err)
		newRevision, err := repo.EnqueueNextVisitSync(context.Background(), clinicID, owner.ID)
		require.NoError(t, err)
		require.Greater(t, newRevision, oldRevision)
		require.NoError(t, repo.WithNextVisitOwnerLock(context.Background(), clinicID, owner.ID, func(syncCtx context.Context) error {
			return repo.AcknowledgeNextVisitSync(syncCtx, clinicID, owner.ID, oldRevision)
		}))
		var currentRevision uint64
		require.NoError(t, db.Raw(`SELECT revision FROM lstep_next_visit_sync_retries
			WHERE clinic_id = ? AND owner_id = ?`, clinicID, owner.ID).Scan(&currentRevision).Error)
		assert.Equal(t, newRevision, currentRevision, "CAS acknowledgement must not delete a newer marker")
	})
}

func TestNextVisitCleanup_RealDB_RemovesOnlyStaleNextVisitTags(t *testing.T) {
	db := setupNextVisitCleanupRealDB(t)
	repo := newCleanupRealRepo(db)
	const clinicID = uint64(1)
	uid := "Ucleanup-real"
	owner := makeCleanupOwner(t, db, clinicID, "end to end", uid, false)
	deleteWithCleanupIntent(t, db, repo, clinicID, owner, uid)
	require.NoError(t, repo.UpsertTag(context.Background(), clinicID, owner.ID, "next_visit_2026-10-15", "auto", ""))
	require.NoError(t, repo.UpsertTag(context.Background(), clinicID, owner.ID, "next_visit_2026-09-30", "auto", ""))
	require.NoError(t, repo.UpsertTag(context.Background(), clinicID, owner.ID, "friend_rank_a", "auto", ""))

	spy := &cleanupProviderSpy{tags: []string{"next_visit_2026-10-15", "next_visit_2026-11-01", "friend_rank_a"}}
	svc := newCleanupRealService(repo, spy)
	require.NoError(t, svc.SyncNextVisitTag(context.Background(), clinicID, owner.ID))

	assert.Zero(t, spy.addCalls)
	assert.ElementsMatch(t, []string{"next_visit_2026-10-15", "next_visit_2026-11-01"}, spy.removed)
	assert.True(t, cacheTagExists(t, db, clinicID, owner.ID, "friend_rank_a"), "unrelated cache rows survive")
	assert.False(t, cacheTagExists(t, db, clinicID, owner.ID, "next_visit_2026-10-15"))
	assert.False(t, cacheTagExists(t, db, clinicID, owner.ID, "next_visit_2026-09-30"))
	assert.False(t, retryRowExists(t, db, clinicID, owner.ID), "the intent is acknowledged on success")
	assert.Equal(t, 1, spy.getCalls)
}

func TestNextVisitCleanup_RealDB_RetainedIntentConverges(t *testing.T) {
	db := setupNextVisitCleanupRealDB(t)
	repo := newCleanupRealRepo(db)
	const clinicID = uint64(1)

	for _, tc := range []struct {
		name  string
		uid   string
		setup func(spy *cleanupProviderSpy, ownerID uint64) func(context.Context) error
	}{
		{"partial remote removal failure", "Upartial", func(spy *cleanupProviderSpy, ownerID uint64) func(context.Context) error {
			failed := false
			spy.removeFn = func(_ context.Context, tag string) error {
				if tag == "next_visit_2026-11-01" && !failed {
					failed = true
					return errors.New("provider removal lost mid-flight")
				}
				return nil
			}
			svc := newCleanupRealService(repo, spy)
			return func(ctx context.Context) error {
				return svc.SyncNextVisitTag(ctx, clinicID, ownerID)
			}
		}},
		{"cache delete failure", "Ucache-fail", func(spy *cleanupProviderSpy, ownerID uint64) func(context.Context) error {
			hook := &cleanupStoreHook{lstepTagCacheRepository: repo}
			failed := false
			hook.deleteTagHook = func(tag string) error {
				if tag == "next_visit_2026-10-15" && !failed {
					failed = true
					return errors.New("cache delete lost")
				}
				return nil
			}
			svc := newCleanupRealService(repo, spy)
			svc.tagCacheRepo = hook
			return func(ctx context.Context) error {
				return svc.SyncNextVisitTag(ctx, clinicID, ownerID)
			}
		}},
		{"completion failure after successful work", "Ucompletion", func(spy *cleanupProviderSpy, ownerID uint64) func(context.Context) error {
			hook := &cleanupStoreHook{lstepTagCacheRepository: repo}
			svc := newCleanupRealService(repo, spy)
			svc.tagCacheRepo = hook
			failed := false
			return func(ctx context.Context) error {
				if !failed {
					failed = true
					hook.postCallbackErr = errors.New("transaction completion failed")
				} else {
					hook.postCallbackErr = nil
				}
				return svc.SyncNextVisitTag(ctx, clinicID, ownerID)
			}
		}},
		{"cancellation during removal", "Ucancel-mid", func(spy *cleanupProviderSpy, ownerID uint64) func(context.Context) error {
			svc := newCleanupRealService(repo, spy)
			var cancel context.CancelFunc
			cancelled := false
			spy.removeFn = func(_ context.Context, _ string) error {
				if !cancelled {
					cancelled = true
					cancel()
				}
				return nil
			}
			return func(ctx context.Context) error {
				runCtx, c := context.WithCancel(ctx)
				defer c()
				cancel = c
				return svc.SyncNextVisitTag(runCtx, clinicID, ownerID)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			owner := makeCleanupOwner(t, db, clinicID, tc.name, tc.uid, false)
			deleteWithCleanupIntent(t, db, repo, clinicID, owner, tc.uid)
			require.NoError(t, repo.UpsertTag(context.Background(), clinicID, owner.ID, "next_visit_2026-10-15", "auto", ""))
			require.NoError(t, repo.UpsertTag(context.Background(), clinicID, owner.ID, "next_visit_2026-11-01", "auto", ""))
			require.NoError(t, repo.UpsertTag(context.Background(), clinicID, owner.ID, "friend_rank_a", "auto", ""))

			spy := &cleanupProviderSpy{tags: []string{"next_visit_2026-10-15", "next_visit_2026-11-01", "friend_rank_a"}}
			run := tc.setup(spy, owner.ID)

			require.Error(t, run(context.Background()), "the first attempt must fail and retain the intent")
			assert.True(t, retryRowExists(t, db, clinicID, owner.ID), "intent retained after %s", tc.name)
			assert.Zero(t, spy.addCalls)

			require.NoError(t, run(context.Background()), "the retry must converge")
			assert.False(t, retryRowExists(t, db, clinicID, owner.ID), "the retry acknowledges the intent")
			assert.ElementsMatch(t, []string{"friend_rank_a"}, spy.tags, "only next_visit tags leave the provider")
			assert.True(t, cacheTagExists(t, db, clinicID, owner.ID, "friend_rank_a"))
			assert.False(t, cacheTagExists(t, db, clinicID, owner.ID, "next_visit_2026-10-15"))
			assert.False(t, cacheTagExists(t, db, clinicID, owner.ID, "next_visit_2026-11-01"))
			assert.Zero(t, spy.addCalls, "cleanup never adds tags even across retries")
		})
	}
}

func TestNextVisitCleanup_RealDB_RelinkedRecipientConflictsThenConverges(t *testing.T) {
	db := setupNextVisitCleanupRealDB(t)
	repo := newCleanupRealRepo(db)
	ownerRepo := ownerdomain.NewRepository(db, nil)
	const clinicID = uint64(1)
	uid := "Urelink-real"
	deleted := makeCleanupOwner(t, db, clinicID, "deleted holder", uid, false)
	deleteWithCleanupIntent(t, db, repo, clinicID, deleted, uid)
	require.NoError(t, repo.UpsertTag(context.Background(), clinicID, deleted.ID, "next_visit_2026-10-15", "auto", ""))
	relinked := makeCleanupOwner(t, db, clinicID, "relink receiver", "", false)
	require.NoError(t, ownerRepo.UpdateLineUserID(context.Background(), clinicID, relinked.ID, &uid))

	spy := &cleanupProviderSpy{tags: []string{"next_visit_2026-10-15"}}
	svc := newCleanupRealService(repo, spy)
	err := svc.SyncNextVisitTag(context.Background(), clinicID, deleted.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
	assert.Zero(t, spy.getCalls, "a relinked recipient must never reach the provider")
	assert.Zero(t, spy.addCalls)
	assert.True(t, retryRowExists(t, db, clinicID, deleted.ID), "the intent survives for later convergence")
	assert.True(t, cacheTagExists(t, db, clinicID, deleted.ID, "next_visit_2026-10-15"))

	require.NoError(t, ownerRepo.UpdateLineUserID(context.Background(), clinicID, relinked.ID, nil))
	require.NoError(t, svc.SyncNextVisitTag(context.Background(), clinicID, deleted.ID))
	assert.Equal(t, 1, spy.getCalls)
	assert.Equal(t, []string{"next_visit_2026-10-15"}, spy.removed)
	assert.False(t, retryRowExists(t, db, clinicID, deleted.ID), "the retry converges once the recipient frees")
}

func TestNextVisitCleanup_RealDB_RecipientLockSerializesRelink(t *testing.T) {
	const clinicID = uint64(1)

	t.Run("cleanup holds the recipient lock while a relink waits", func(t *testing.T) {
		db := setupNextVisitCleanupRealDB(t)
		repo := newCleanupRealRepo(db)
		ownerRepo := ownerdomain.NewRepository(db, nil)
		uid := "Ucleanup-holds"
		deleted := makeCleanupOwner(t, db, clinicID, "cleanup holds lock", uid, false)
		deleteWithCleanupIntent(t, db, repo, clinicID, deleted, uid)
		receiver := makeCleanupOwner(t, db, clinicID, "blocked relink receiver", "", false)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		getEntered := make(chan struct{})
		releaseGet := make(chan struct{})
		var releaseOnce sync.Once
		release := func() { releaseOnce.Do(func() { close(releaseGet) }) }
		spy := &cleanupProviderSpy{
			getFn: func() ([]string, error) {
				close(getEntered)
				select {
				case <-releaseGet:
					return []string{"next_visit_2026-10-15"}, nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		}
		svc := newCleanupRealService(repo, spy)

		syncDone := make(chan error, 1)
		relinkDone := make(chan error, 1)
		syncJoined, relinkJoined := false, false
		t.Cleanup(func() {
			release()
			if !syncJoined {
				select {
				case <-syncDone:
				case <-time.After(15 * time.Second):
					t.Error("cleanup goroutine did not finish during test cleanup")
				}
			}
			if !relinkJoined {
				select {
				case <-relinkDone:
				case <-time.After(15 * time.Second):
					t.Error("relink goroutine did not finish during test cleanup")
				}
			}
		})

		go func() {
			syncDone <- svc.SyncNextVisitTag(ctx, clinicID, deleted.ID)
		}()
		select {
		case <-getEntered:
		case err := <-syncDone:
			syncJoined = true
			require.Failf(t, "cleanup finished before the provider barrier", "err=%v", err)
		case <-ctx.Done():
			require.Fail(t, "cleanup never reached the provider barrier")
		}

		go func() {
			relinkDone <- ownerRepo.UpdateLineUserID(ctx, clinicID, receiver.ID, &uid)
		}()
		waiterPID := waitForExactAdvisoryWaiter(t, db, lineRecipientLockKey(clinicID, uid))
		var blocked bool
		require.NoError(t, db.Raw(`SELECT cardinality(pg_blocking_pids(?)) > 0`, waiterPID).Scan(&blocked).Error)
		assert.True(t, blocked, "the relink waiter on the exact recipient key must have a blocker")
		select {
		case err := <-relinkDone:
			relinkJoined = true
			require.Failf(t, "relink completed while cleanup held the recipient lock", "err=%v", err)
		default:
		}

		release()
		select {
		case err := <-syncDone:
			syncJoined = true
			require.NoError(t, err)
		case <-ctx.Done():
			require.Fail(t, "cleanup did not finish after the provider was released")
		}
		select {
		case err := <-relinkDone:
			relinkJoined = true
			require.NoError(t, err)
		case <-ctx.Done():
			require.Fail(t, "relink did not finish after cleanup released the recipient lock")
		}
		assert.Equal(t, []string{"next_visit_2026-10-15"}, spy.removed)
		assert.Zero(t, spy.addCalls)
		assert.False(t, retryRowExists(t, db, clinicID, deleted.ID))
	})

	t.Run("relink holds the recipient lock while cleanup waits", func(t *testing.T) {
		db := setupNextVisitCleanupRealDB(t)
		repo := newCleanupRealRepo(db)
		ownerRepo := ownerdomain.NewRepository(db, nil)
		uid := "Urelink-holds"
		deleted := makeCleanupOwner(t, db, clinicID, "relink holds lock", uid, false)
		deleteWithCleanupIntent(t, db, repo, clinicID, deleted, uid)
		receiver := makeCleanupOwner(t, db, clinicID, "holding relink receiver", "", false)
		spy := &cleanupProviderSpy{tags: []string{"next_visit_2026-10-15"}}
		svc := newCleanupRealService(repo, spy)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		relinkTx := db.Begin()
		require.NoError(t, relinkTx.Error)
		var relinkPID int64
		require.NoError(t, relinkTx.Raw("SELECT pg_backend_pid()").Scan(&relinkPID).Error)
		require.NoError(t, ownerRepo.UpdateLineUserID(
			persistence.WithTxValue(context.Background(), relinkTx), clinicID, receiver.ID, &uid))

		syncDone := make(chan error, 1)
		syncJoined := false
		t.Cleanup(func() {
			_ = relinkTx.Rollback().Error
			if !syncJoined {
				select {
				case <-syncDone:
				case <-time.After(15 * time.Second):
					t.Error("cleanup goroutine did not finish during test cleanup")
				}
			}
		})

		go func() {
			syncDone <- svc.SyncNextVisitTag(ctx, clinicID, deleted.ID)
		}()
		waiterPID := waitForExactAdvisoryWaiter(t, db, lineRecipientLockKey(clinicID, uid))
		var blockedByRelink bool
		require.NoError(t, db.Raw(`SELECT ? = ANY(pg_blocking_pids(?))`, relinkPID, waiterPID).Scan(&blockedByRelink).Error)
		assert.True(t, blockedByRelink,
			"the cleanup waiter must be blocked by the relink backend on the exact recipient key")
		select {
		case err := <-syncDone:
			syncJoined = true
			require.Failf(t, "cleanup completed while the relink held the recipient lock", "err=%v", err)
		default:
		}

		require.NoError(t, relinkTx.Commit().Error)
		select {
		case err := <-syncDone:
			syncJoined = true
			require.Error(t, err)
			assert.True(t, apperrors.IsConflict(err), "committed relink must conflict, got %v", err)
		case <-ctx.Done():
			require.Fail(t, "cleanup did not finish after the relink committed")
		}
		assert.Zero(t, spy.getCalls, "provider must not be touched after the recipient moved")
		assert.Zero(t, spy.addCalls)
		assert.True(t, retryRowExists(t, db, clinicID, deleted.ID), "the intent survives for later convergence")
	})
}

type gatedTagSync struct {
	*lstepTagSyncService
	onEnqueue func(context.Context)
}

func (g *gatedTagSync) EnqueueNextVisitCleanup(ctx context.Context, clinicID, ownerID uint64, lineUserID string) error {
	if g.onEnqueue != nil {
		g.onEnqueue(ctx)
	}
	return g.lstepTagSyncService.EnqueueNextVisitCleanup(ctx, clinicID, ownerID, lineUserID)
}

func TestNextVisitSync_RealDB_DeleteSerializesNormalSync(t *testing.T) {
	db := setupNextVisitCleanupRealDB(t)
	repo := newCleanupRealRepo(db)
	ownerRepo := ownerdomain.NewRepository(db, nil)
	petOwnerRepo := petdomain.NewPetOwnerRepository(db)
	const clinicID = uint64(1)
	uid := "Udelete-vs-sync"
	owner := makeCleanupOwner(t, db, clinicID, "delete vs sync target", uid, false)
	require.NoError(t, repo.UpsertTag(context.Background(), clinicID, owner.ID, "next_visit_2026-10-15", "auto", ""))

	spy := &cleanupProviderSpy{tags: []string{"next_visit_2026-10-15"}}
	tagSync := newCleanupRealService(repo, spy)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	deleteEnqueued := make(chan struct{})
	allowCommit := make(chan struct{})
	var allowOnce sync.Once
	allow := func() { allowOnce.Do(func() { close(allowCommit) }) }
	var deletePID int64
	var deletePIDErr error
	notifier := &gatedTagSync{
		lstepTagSyncService: tagSync,
		onEnqueue: func(txCtx context.Context) {
			deletePIDErr = persistence.DBOrTx(txCtx, db).
				Raw("SELECT pg_backend_pid()").Scan(&deletePID).Error
			close(deleteEnqueued)
			select {
			case <-allowCommit:
			case <-ctx.Done():
			}
		},
	}
	ownerSvc := ownerdomain.NewServiceWithPetOwnerDeleteGuard(
		ownerRepo, nil, notifier, nil, ownerRepo, petOwnerRepo, persistence.NewTransactor(db))

	deleteDone := make(chan error, 1)
	syncDone := make(chan error, 1)
	syncExited := make(chan struct{})
	deleteJoined, syncJoined := false, false
	t.Cleanup(func() {
		allow()
		if !deleteJoined {
			select {
			case <-deleteDone:
			case <-time.After(15 * time.Second):
				t.Error("delete goroutine did not finish during test cleanup")
			}
		}
		if !syncJoined {
			select {
			case <-syncExited:
			case <-time.After(15 * time.Second):
				t.Error("sync goroutine did not finish during test cleanup")
			}
		}
	})

	go func() {
		deleteDone <- ownerSvc.Delete(ctx, clinicID, owner.ID)
	}()
	select {
	case <-deleteEnqueued:
	case err := <-deleteDone:
		deleteJoined = true
		require.Failf(t, "delete finished before the intent gate", "err=%v", err)
	case <-ctx.Done():
		require.Fail(t, "delete never reached the cleanup enqueue")
	}

	require.NoError(t, deletePIDErr)
	require.NotZero(t, deletePID, "the gated deletion backend must identify its own pid")

	go func() {
		defer close(syncExited)
		syncDone <- tagSync.SyncNextVisitTag(ctx, clinicID, owner.ID)
	}()
	waitForBackendBlockedBy(t, db, deletePID, "%lstep_next_visit_sync_retries%", syncDone)

	allow()
	select {
	case err := <-deleteDone:
		deleteJoined = true
		require.NoError(t, err)
	case <-ctx.Done():
		require.Fail(t, "owner delete did not finish after the gate opened")
	}
	select {
	case err := <-syncDone:
		syncJoined = true
		require.Error(t, err, "the normal sync re-read must see the committed deletion")
	case <-ctx.Done():
		require.Fail(t, "normal sync did not finish after the delete committed")
	}
	assert.Zero(t, spy.addCalls, "the normal path must never AddTag for a deleted owner")
	assert.Equal(t, 1, spy.getCalls, "only the post-commit cleanup reaches the provider")
	assert.Equal(t, []string{"next_visit_2026-10-15"}, spy.removed)
	assert.False(t, retryRowExists(t, db, clinicID, owner.ID), "the cleanup intent is acknowledged")
}

func TestNextVisitSync_RealDB_EnqueueSerializesBehindOwnerDelete(t *testing.T) {
	db := setupNextVisitCleanupRealDB(t)
	repo := newCleanupRealRepo(db)
	ownerRepo := ownerdomain.NewRepository(db, nil)
	const clinicID = uint64(1)
	uid := "Udelete-enqueue-race"
	owner := makeCleanupOwner(t, db, clinicID, "delete enqueue race", uid, false)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	deleteTx := db.Begin()
	require.NoError(t, deleteTx.Error)
	t.Cleanup(func() { _ = deleteTx.Rollback().Error })
	deleteTxCtx := persistence.WithTxValue(context.Background(), deleteTx)
	locked, err := ownerRepo.LockForDelete(deleteTxCtx, clinicID, owner.ID)
	require.NoError(t, err)
	require.NotNil(t, locked)
	var deletePID int64
	require.NoError(t, deleteTx.Raw("SELECT pg_backend_pid()").Scan(&deletePID).Error)

	enqueueDone := make(chan error, 1)
	enqueueExited := make(chan struct{})
	enqueueJoined := false
	go func() {
		defer close(enqueueExited)
		_, err := repo.EnqueueNextVisitSync(ctx, clinicID, owner.ID)
		enqueueDone <- err
	}()
	t.Cleanup(func() {
		_ = deleteTx.Rollback().Error
		if enqueueJoined {
			return
		}
		select {
		case <-enqueueExited:
		case <-time.After(15 * time.Second):
			t.Error("normal enqueue goroutine did not finish during test cleanup")
		}
	})

	waitForBackendBlockedBy(t, db, deletePID, "%INSERT INTO lstep_next_visit_sync_retries%", enqueueDone)
	assert.False(t, retryRowExists(t, db, clinicID, owner.ID), "the marker is initially absent")

	require.NoError(t, repo.EnqueueNextVisitCleanup(deleteTxCtx, clinicID, owner.ID, uid))
	require.NoError(t, deleteTx.Delete(owner).Error)
	require.NoError(t, deleteTx.Commit().Error)

	select {
	case err := <-enqueueDone:
		enqueueJoined = true
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err),
			"deletion-first enqueue must return NotFound after revalidation, got %v", err)
	case <-ctx.Done():
		enqueueJoined = true
		require.Fail(t, "normal enqueue did not finish after the delete committed")
	}
	intent, err := repo.FindNextVisitCleanupIntent(context.Background(), clinicID, owner.ID)
	require.NoError(t, err, "the deletion intent survives while the blocked normal enqueue resolves to NotFound")
	assert.Equal(t, uid, intent.LineUserID)
	var reloaded model.Owner
	require.NoError(t, db.Unscoped().Where("id = ?", owner.ID).First(&reloaded).Error)
	assert.True(t, reloaded.DeletedAt.Valid)
}

func TestNextVisitCleanup_RealDB_OwnerDeleteEndToEnd(t *testing.T) {
	db := setupNextVisitCleanupRealDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ReservationType{}, &model.Reservation{}))
	repo := newCleanupRealRepo(db)
	ownerRepo := ownerdomain.NewRepository(db, nil)
	petOwnerRepo := petdomain.NewPetOwnerRepository(db)
	const clinicID = uint64(1)
	uid := "Ue2e-recipient"
	owner := makeCleanupOwner(t, db, clinicID, "delete cleanup e2e", uid, false)

	resType := &model.ReservationType{ClinicID: clinicID, Name: "e2e type"}
	require.NoError(t, db.Create(resType).Error)
	future := time.Now().Add(72 * time.Hour)
	appt := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         future,
		EndTime:           future.Add(30 * time.Minute),
		OwnerID:           &owner.ID,
		PetID:             nil,
		ReservationTypeID: resType.ID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusConfirmed,
		Source:            model.ReservationSourceManual,
		CustomerFields:    json.RawMessage(`{}`),
	}
	require.NoError(t, db.Create(appt).Error)

	spy := &cleanupProviderSpy{tags: []string{"next_visit_2026-10-15", "friend_rank_a"}}
	tagSync := newCleanupRealService(repo, spy)
	require.NoError(t, repo.UpsertTag(context.Background(), clinicID, owner.ID, "next_visit_2026-10-15", "auto", ""))

	ownerSvc := ownerdomain.NewServiceWithPetOwnerDeleteGuard(
		ownerRepo, nil, tagSync, nil, ownerRepo, petOwnerRepo, persistence.NewTransactor(db))
	require.NoError(t, ownerSvc.Delete(context.Background(), clinicID, owner.ID))

	assert.Zero(t, spy.addCalls, "owner deletion cleanup must never add tags")
	assert.Equal(t, []string{"next_visit_2026-10-15"}, spy.removed)
	assert.Equal(t, 1, spy.getCalls)
	assert.False(t, retryRowExists(t, db, clinicID, owner.ID), "the cleanup intent is acknowledged")
	assert.False(t, cacheTagExists(t, db, clinicID, owner.ID, "next_visit_2026-10-15"))

	var reloaded model.Reservation
	require.NoError(t, db.Unscoped().Where("id = ?", appt.ID).First(&reloaded).Error)
	require.NotNil(t, reloaded.OwnerID)
	assert.Equal(t, owner.ID, *reloaded.OwnerID, "the appointment must stay owner-linked")
	assert.Nil(t, reloaded.PetID, "the owner-only appointment keeps pet_id NULL")
	assert.False(t, reloaded.DeletedAt.Valid, "the appointment itself is untouched")
}

func TestNextVisitNotify_RealDB_SuccessClearsCounterOnSingleConnection(t *testing.T) {
	db := setupNextVisitCleanupRealDB(t)
	repo := newCleanupRealRepo(db)
	counterRepo := NewLstepSyncErrorCounterRepository(db)
	const clinicID = uint64(1)
	uid := "Unotify-pool"
	owner := makeCleanupOwner(t, db, clinicID, "notify pool target", uid, false)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	previousMax := sqlDB.Stats().MaxOpenConnections
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.SetMaxOpenConns(previousMax) })

	for range 2 {
		_, err := counterRepo.IncrementFailure(context.Background(), clinicID, owner.ID)
		require.NoError(t, err)
	}
	require.NoError(t, repo.UpsertTag(context.Background(), clinicID, owner.ID, lstepErrorTag, "auto", ""))

	spy := &cleanupProviderSpy{tags: []string{lstepErrorTag}}
	svc := newCleanupRealService(repo, spy)
	svc.errorCounterRepo = counterRepo
	svc.nextVisitAppointmentRepo = &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
		return nil, nil
	}}

	require.NoError(t, svc.SyncNextVisitTag(context.Background(), clinicID, owner.ID),
		"the success notifier must not deadlock on a single-connection pool")
	assert.Contains(t, spy.removed, lstepErrorTag, "the EXCL error tag is removed after protected revalidation")
	assert.Equal(t, 0, spy.addCalls)
	_, err = counterRepo.FindByOwner(context.Background(), clinicID, owner.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "the counter resets only after the protected callback succeeds")
}

func TestNextVisitCleanup_RealDB_ClinicIsolationAndDeletedSelf(t *testing.T) {
	db := setupNextVisitCleanupRealDB(t)
	repo := newCleanupRealRepo(db)
	const clinicID = uint64(1)
	uid := "Ushared-recipient"
	deleted := makeCleanupOwner(t, db, clinicID, "clinic one deleted", uid, false)
	deleteWithCleanupIntent(t, db, repo, clinicID, deleted, uid)
	makeCleanupOwner(t, db, uint64(2), "clinic two active", uid, false)

	spy := &cleanupProviderSpy{tags: []string{"next_visit_2026-10-15"}}
	svc := newCleanupRealService(repo, spy)
	require.NoError(t, svc.SyncNextVisitTag(context.Background(), clinicID, deleted.ID),
		"same recipient in another clinic and the deleted owner itself must not block cleanup")
	assert.Equal(t, 1, spy.getCalls)
	assert.Equal(t, []string{"next_visit_2026-10-15"}, spy.removed)
	assert.False(t, retryRowExists(t, db, clinicID, deleted.ID))
}

func TestNextVisitCleanup_RealDB_DeletedOwnerWithoutIntentFailsClosed(t *testing.T) {
	db := setupNextVisitCleanupRealDB(t)
	repo := newCleanupRealRepo(db)
	const clinicID = uint64(1)
	deleted := makeCleanupOwner(t, db, clinicID, "no intent", "Uno-intent", true)

	spy := &cleanupProviderSpy{}
	svc := newCleanupRealService(repo, spy)
	err := svc.SyncNextVisitTag(context.Background(), clinicID, deleted.ID)
	require.Error(t, err)
	assert.Zero(t, spy.getCalls, "no provider work happens for a deleted owner without intent")
	assert.Zero(t, spy.addCalls)
	assert.False(t, retryRowExists(t, db, clinicID, deleted.ID))
}
