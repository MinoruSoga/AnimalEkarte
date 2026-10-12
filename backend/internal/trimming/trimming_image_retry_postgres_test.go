package trimming

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupTrimmingRetryPostgres(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupImageLifecyclePostgres(t)
	ddl, err := os.ReadFile(filepath.Join("..", "..", "migrations", "035_trimming_image_storage_retry.sql"))
	require.NoError(t, err)
	alterDDL := string(ddl)
	alterDDL = strings.Replace(alterDDL, "ADD COLUMN storage_deleted_at", "ADD COLUMN IF NOT EXISTS storage_deleted_at", 1)
	alterDDL = strings.Replace(alterDDL, "ADD COLUMN last_attempt_at", "ADD COLUMN IF NOT EXISTS last_attempt_at", 1)
	alterDDL = strings.Replace(alterDDL, "CREATE INDEX idx_", "CREATE INDEX IF NOT EXISTS idx_", 1)
	require.NoError(t, db.Exec(alterDDL).Error)
	testdb.SeedClinicsForFK(t, db, 1, 2)
	return db
}

func insertTrimmingDeletionTombstone(t *testing.T, db *gorm.DB, clinicID uint64, key string) {
	t.Helper()
	require.NoError(t, db.Exec(
		"INSERT INTO trimming_image_deletions (clinic_id, object_key, actor_staff_id) VALUES (?, ?, ?)",
		clinicID, key, 42,
	).Error)
}

type retryRowState struct {
	LastAttemptAt    *time.Time
	StorageDeletedAt *time.Time
}

func readRetryRowState(t *testing.T, db *gorm.DB, clinicID uint64, key string) retryRowState {
	t.Helper()
	var state retryRowState
	require.NoError(t, db.Raw(
		"SELECT last_attempt_at, storage_deleted_at FROM trimming_image_deletions WHERE clinic_id = ? AND object_key = ?",
		clinicID, key,
	).Scan(&state).Error)
	return state
}

func TestImageDeletionRetry_PostgresSuccessCompletesOnce(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	store := NewImageDeletionRetryStore(db)
	objects := &spyObjectStore{errs: map[string]int{}}
	key := "trimming-images/1/style.jpg"
	insertTrimmingDeletionTombstone(t, db, 1, key)

	attempted, err := store.RetryImageDeletion(context.Background(), 1, key, objects)
	require.NoError(t, err)
	assert.True(t, attempted)
	assert.Equal(t, []string{key}, objects.calls)

	state := readRetryRowState(t, db, 1, key)
	require.NotNil(t, state.LastAttemptAt)
	require.NotNil(t, state.StorageDeletedAt, "successful delete records completion")

	attempted, err = store.RetryImageDeletion(context.Background(), 1, key, objects)
	require.NoError(t, err)
	assert.False(t, attempted, "completed rows are skipped")
	assert.Equal(t, 1, objects.callCount(), "provider must not be called for a completed row")
}

func TestImageDeletionRetry_PostgresFailureTimestampsThenConverges(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	store := NewImageDeletionRetryStore(db)
	key := "trimming-images/1/b.jpg"
	insertTrimmingDeletionTombstone(t, db, 1, key)
	objects := &spyObjectStore{errs: map[string]int{key: 1}}

	attempted, err := store.RetryImageDeletion(context.Background(), 1, key, objects)
	require.Error(t, err)
	assert.True(t, attempted)

	state := readRetryRowState(t, db, 1, key)
	require.NotNil(t, state.LastAttemptAt, "provider failure commits the attempt timestamp")
	assert.Nil(t, state.StorageDeletedAt, "no completion on failure")

	attempted, err = store.RetryImageDeletion(context.Background(), 1, key, objects)
	require.NoError(t, err)
	assert.True(t, attempted)
	state = readRetryRowState(t, db, 1, key)
	require.NotNil(t, state.StorageDeletedAt, "next attempt completes the pending row")
	assert.Equal(t, 2, objects.callCount())
}

func TestImageDeletionRetry_PostgresTransactionFailureStaysPending(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	store := NewImageDeletionRetryStore(db)
	key := "trimming-images/1/commit-fail.jpg"
	insertTrimmingDeletionTombstone(t, db, 1, key)

	require.NoError(t, db.Exec(`
		CREATE OR REPLACE FUNCTION fail_retry_commit() RETURNS trigger AS $f$
		BEGIN RAISE EXCEPTION 'forced deferred failure'; END;
		$f$ LANGUAGE plpgsql`).Error)
	require.NoError(t, db.Exec(`
		CREATE CONSTRAINT TRIGGER trg_retry_commit_fail
		AFTER UPDATE ON trimming_image_deletions
		DEFERRABLE INITIALLY DEFERRED
		FOR EACH ROW EXECUTE FUNCTION fail_retry_commit()`).Error)
	t.Cleanup(func() {
		require.NoError(t, db.Exec("DROP TRIGGER IF EXISTS trg_retry_commit_fail ON trimming_image_deletions").Error)
		require.NoError(t, db.Exec("DROP FUNCTION IF EXISTS fail_retry_commit").Error)
	})

	objects := &spyObjectStore{errs: map[string]int{}}
	_, err := store.RetryImageDeletion(context.Background(), 1, key, objects)
	require.Error(t, err, "commit ambiguity surfaces as an error")
	assert.Equal(t, []string{key}, objects.calls, "the provider delete ran before the commit failed")

	require.NoError(t, db.Exec("DROP TRIGGER trg_retry_commit_fail ON trimming_image_deletions").Error)
	state := readRetryRowState(t, db, 1, key)
	assert.Nil(t, state.StorageDeletedAt, "rollback keeps the row pending")

	attempted, err := store.RetryImageDeletion(context.Background(), 1, key, objects)
	require.NoError(t, err)
	assert.True(t, attempted)
	state = readRetryRowState(t, db, 1, key)
	require.NotNil(t, state.StorageDeletedAt, "retry converges the pending row")
	assert.Equal(t, 2, objects.callCount(), "idempotent provider delete tolerates the duplicate call")
}

func TestImageDeletionRetry_PostgresLockBusySkips(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	store := NewImageDeletionRetryStore(db)
	key := "trimming-images/1/busy.jpg"
	insertTrimmingDeletionTombstone(t, db, 1, key)

	holder := db.Begin()
	require.NoError(t, holder.Error)
	t.Cleanup(func() { holder.Rollback() })
	lockName := fmt.Sprintf("trimming-image:%d:%s", 1, key)
	require.NoError(t, holder.Exec(
		"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", lockName).Error)

	objects := &spyObjectStore{errs: map[string]int{}}
	attempted, err := store.RetryImageDeletion(context.Background(), 1, key, objects)
	require.NoError(t, err)
	assert.False(t, attempted, "a busy key lock skips without blocking")
	assert.Zero(t, objects.callCount(), "provider must not run under a busy lock")
}

func TestImageDeletionRetry_PostgresForeignAndInvalidKeysRejectBeforeProvider(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	store := NewImageDeletionRetryStore(db)
	objects := &spyObjectStore{errs: map[string]int{}}

	for _, tt := range []struct {
		name     string
		clinicID uint64
		key      string
	}{
		{name: "foreign clinic key", clinicID: 1, key: "trimming-images/2/other.jpg"},
		{name: "traversal alias", clinicID: 1, key: "trimming-images/1/../secret.jpg"},
		{name: "unprefixed key", clinicID: 1, key: "other-prefix/1/x.jpg"},
		{name: "missing name", clinicID: 1, key: "trimming-images/1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			attempted, err := store.RetryImageDeletion(context.Background(), tt.clinicID, tt.key, objects)
			require.Error(t, err)
			assert.False(t, attempted)
		})
	}
	assert.Zero(t, objects.callCount(), "invalid keys never reach the provider")
}

func TestImageDeletionRetry_PostgresPendingOrderAndClinicScope(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	store := NewImageDeletionRetryStore(db)
	insertTrimmingDeletionTombstone(t, db, 1, "trimming-images/1/never.jpg")
	insertTrimmingDeletionTombstone(t, db, 1, "trimming-images/1/old.jpg")
	insertTrimmingDeletionTombstone(t, db, 2, "trimming-images/2/foreign.jpg")
	require.NoError(t, db.Exec(
		"UPDATE trimming_image_deletions SET last_attempt_at = NOW() - interval '1 hour' WHERE object_key = ?",
		"trimming-images/1/old.jpg",
	).Error)

	keys, err := store.ListPendingImageKeys(context.Background(), 1, 100)
	require.NoError(t, err)
	assert.Equal(t, []string{"trimming-images/1/never.jpg", "trimming-images/1/old.jpg"}, keys,
		"NULLS FIRST then last_attempt_at ASC")

	keys, err = store.ListPendingImageKeys(context.Background(), 2, 100)
	require.NoError(t, err)
	assert.Equal(t, []string{"trimming-images/2/foreign.jpg"}, keys, "clinic scope is strict")
}

func TestImageDeletionRetry_PostgresConcurrentSecondWorkerBusy(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	first := NewImageDeletionRetryStore(db)
	second := NewImageDeletionRetryStore(db)
	key := "trimming-images/1/concurrent.jpg"
	insertTrimmingDeletionTombstone(t, db, 1, key)

	release := make(chan struct{})
	var closeOnce sync.Once
	releaseFirst := func() { closeOnce.Do(func() { close(release) }) }
	firstErr := make(chan error, 1)
	firstFinished := make(chan struct{})
	t.Cleanup(func() {
		releaseFirst()
		select {
		case <-firstFinished:
		case <-time.After(10 * time.Second):
			t.Error("first worker goroutine leaked past cleanup")
		}
	})
	var enterOnce sync.Once
	blockingObjects := &spyObjectStore{
		errs: map[string]int{},
		hook: func(string) { enterOnce.Do(func() { <-release }) },
	}

	go func() {
		defer close(firstFinished)
		_, err := first.RetryImageDeletion(context.Background(), 1, key, blockingObjects)
		firstErr <- err
	}()
	require.Eventually(t, func() bool { return blockingObjects.callCount() == 1 },
		5*time.Second, 5*time.Millisecond, "first worker must enter the provider call")

	quietObjects := &spyObjectStore{errs: map[string]int{}}
	attempted, err := second.RetryImageDeletion(context.Background(), 1, key, quietObjects)
	require.NoError(t, err)
	assert.False(t, attempted, "second worker sees the key lock busy")
	assert.Zero(t, quietObjects.callCount(), "second worker must not invoke the provider")

	releaseFirst()
	select {
	case <-firstFinished:
	case <-time.After(10 * time.Second):
		t.Fatal("first worker did not finish after release")
	}
	require.NoError(t, <-firstErr)
	assert.Equal(t, 1, blockingObjects.callCount(), "provider invoked exactly once across overlapping workers")
}

func TestImageDeletionRetry_PostgresListLimitsCapped(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	store := NewImageDeletionRetryStore(db)
	insertTrimmingDeletionTombstone(t, db, 1, "trimming-images/1/bound.jpg")

	for _, limit := range []int{0, -5, imageDeletionMaxClinicsPerRun + 500} {
		clinics, err := store.ListClinicIDsAfter(context.Background(), 0, limit)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(clinics), imageDeletionMaxClinicsPerRun,
			"clinic listing must be capped regardless of caller limit %d", limit)
	}
	for _, limit := range []int{0, -5, imageDeletionMaxKeysPerRun + 500} {
		keys, err := store.ListPendingImageKeys(context.Background(), 1, limit)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(keys), imageDeletionMaxKeysPerRun,
			"key listing must be capped regardless of caller limit %d", limit)
		assert.Equal(t, []string{"trimming-images/1/bound.jpg"}, keys)
	}
}

func TestImageDeletionRetry_PostgresFailClosedDependencies(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	objects := &spyObjectStore{errs: map[string]int{}}
	key := "trimming-images/1/guard.jpg"
	insertTrimmingDeletionTombstone(t, db, 1, key)

	nilStore := NewImageDeletionRetryStore(nil)
	_, err := nilStore.ListClinicIDsAfter(context.Background(), 0, 100)
	require.Error(t, err, "nil database store must fail closed, not panic")
	_, err = nilStore.ListPendingImageKeys(context.Background(), 1, 100)
	require.Error(t, err)
	_, err = nilStore.RetryImageDeletion(context.Background(), 1, key, objects)
	require.Error(t, err)

	store := NewImageDeletionRetryStore(db)
	_, err = store.RetryImageDeletion(context.Background(), 1, key, nil)
	require.Error(t, err, "nil object store must fail closed before any transaction")
	_, err = store.RetryImageDeletion(context.Background(), 0, key, objects)
	require.Error(t, err, "clinic id zero must be rejected before key validation")
	assert.Zero(t, objects.callCount(), "guards must run before the provider is reachable")
}

func TestImageDeletionRetry_PostgresCompletedTombstoneStillBlocksWriter(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	store := NewImageDeletionRetryStore(db)
	objects := &spyObjectStore{errs: map[string]int{}}
	key := "trimming-images/1/done.jpg"
	insertTrimmingDeletionTombstone(t, db, 1, key)

	attempted, err := store.RetryImageDeletion(context.Background(), 1, key, objects)
	require.NoError(t, err)
	require.True(t, attempted)

	repo := NewAppointmentTrimmingDetailRepository(db)
	svc := &trimmingService{trimmingDetail: repo, transactor: newTestTransactor(db)}
	err = newTestTransactor(db).WithTx(context.Background(), func(ctx context.Context) error {
		return svc.lockImageReferencesInTx(ctx, 1, key)
	})
	require.ErrorIs(t, err, apperrors.ErrConflict,
		"a completed tombstone still rejects new references")
}

func TestImageDeletionRetry_PostgresWorkerRunOnceEndToEnd(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	store := NewImageDeletionRetryStore(db)
	objects := &spyObjectStore{errs: map[string]int{"trimming-images/2/x.jpg": 1}}
	insertTrimmingDeletionTombstone(t, db, 1, "trimming-images/1/e2e.jpg")
	insertTrimmingDeletionTombstone(t, db, 2, "trimming-images/2/x.jpg")

	worker := NewImageDeletionWorker(store, objects)
	result, err := worker.RunOnce(context.Background())
	require.Error(t, err)
	assert.Equal(t, 2, result.Processed)
	assert.Equal(t, 1, result.Succeeded)
	assert.Equal(t, 1, result.Failed)

	var clinicCount int64
	require.NoError(t, db.Raw("SELECT count(*) FROM clinics").Scan(&clinicCount).Error)
	for i := int64(0); i < clinicCount/100+3; i++ {
		result, err = worker.RunOnce(context.Background())
		require.NoError(t, err)
		if result.Succeeded > 0 {
			break
		}
	}
	require.Positive(t, result.Succeeded,
		"the worker cursor must paginate the remaining clinics and wrap back to the failed key")
	assert.Equal(t, ImageDeletionRunResult{Processed: 1, Succeeded: 1}, result,
		"the completed key is skipped; only the failed key retries")
	assert.Equal(t,
		[]string{"trimming-images/1/e2e.jpg", "trimming-images/2/x.jpg", "trimming-images/2/x.jpg"},
		objects.calls)

	var pending int64
	require.NoError(t, db.Raw(
		"SELECT count(*) FROM trimming_image_deletions WHERE storage_deleted_at IS NULL").Scan(&pending).Error)
	assert.Zero(t, pending)
}
