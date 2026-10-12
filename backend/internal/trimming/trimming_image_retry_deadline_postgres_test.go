package trimming

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/testdb"
)

func withImageDeletionDeadlineSeams(t *testing.T, run, provider time.Duration) {
	t.Helper()
	prevRun, prevProvider := imageDeletionRunTimeout, imageDeletionProviderTimeout
	imageDeletionRunTimeout, imageDeletionProviderTimeout = run, provider
	t.Cleanup(func() {
		imageDeletionRunTimeout = prevRun
		imageDeletionProviderTimeout = prevProvider
	})
}

type ctxDoneDeleteStore struct {
	mu          sync.Mutex
	calls       []string
	deadline    time.Time
	hasDeadline bool
	fallback    time.Duration
}

func (s *ctxDoneDeleteStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	s.calls = append(s.calls, key)
	s.deadline, s.hasDeadline = ctx.Deadline()
	fallback := s.fallback
	s.mu.Unlock()
	if fallback > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(fallback):
			return fmt.Errorf("provider fallback timeout for %s", key)
		}
	}
	<-ctx.Done()
	return ctx.Err()
}

func (s *ctxDoneDeleteStore) snapshot() (calls []string, deadline time.Time, hasDeadline bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.calls...), s.deadline, s.hasDeadline
}

type mixedDeleteStore struct {
	spy   *spyObjectStore
	block map[string]bool
}

func (m *mixedDeleteStore) Delete(ctx context.Context, key string) error {
	if m.block[key] {
		<-ctx.Done()
		return ctx.Err()
	}
	return m.spy.Delete(ctx, key)
}

func TestImageDeletionRetry_PostgresProviderDeadlineOwnedByStore(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	withImageDeletionDeadlineSeams(t, imageDeletionRunTimeout, 60*time.Millisecond)
	store := NewImageDeletionRetryStore(db)
	key := "trimming-images/1/store-deadline.jpg"
	insertTrimmingDeletionTombstone(t, db, 1, key)

	objects := &ctxDoneDeleteStore{fallback: 3 * time.Second}
	start := time.Now()
	attempted, err := store.RetryImageDeletion(context.Background(), 1, key, objects)
	require.Error(t, err)
	assert.True(t, attempted)

	calls, deadline, hasDeadline := objects.snapshot()
	assert.Equal(t, []string{key}, calls)
	require.True(t, hasDeadline, "the store must bound the provider context")
	assert.Less(t, time.Until(deadline), 5*time.Second,
		"provider deadline must be the short per-key bound, not an ambient caller deadline")
	assert.WithinDuration(t, start.Add(60*time.Millisecond), deadline, 2*time.Second)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	state := readRetryRowState(t, db, 1, key)
	require.NotNil(t, state.LastAttemptAt, "provider deadline still commits the attempt timestamp")
	assert.Nil(t, state.StorageDeletedAt, "failed deletes stay pending")
}

func TestImageDeletionRetry_PostgresParentCancelStillPersistsAttempt(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	store := NewImageDeletionRetryStore(db)
	key := "trimming-images/1/parent-cancel.jpg"
	insertTrimmingDeletionTombstone(t, db, 1, key)

	objects := &ctxDoneDeleteStore{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		attempted, err := store.RetryImageDeletion(ctx, 1, key, objects)
		assert.True(t, attempted)
		done <- err
	}()
	require.Eventually(t, func() bool {
		calls, _, _ := objects.snapshot()
		return len(calls) == 1
	}, 5*time.Second, 5*time.Millisecond, "provider must be entered before the parent cancels")

	cancel()
	select {
	case err := <-done:
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(10 * time.Second):
		t.Fatal("retry did not return after parent cancellation")
	}

	state := readRetryRowState(t, db, 1, key)
	require.NotNil(t, state.LastAttemptAt,
		"the independent transaction context persists the attempt after parent cancellation")
	assert.Nil(t, state.StorageDeletedAt, "cancelled deletes stay pending")
}

func TestImageDeletionRetry_PostgresCancelledBeforeProviderNeverCalls(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	store := NewImageDeletionRetryStore(db)
	key := "trimming-images/1/pre-cancelled.jpg"
	insertTrimmingDeletionTombstone(t, db, 1, key)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	objects := &spyObjectStore{errs: map[string]int{}}
	attempted, err := store.RetryImageDeletion(ctx, 1, key, objects)
	require.Error(t, err)
	assert.False(t, attempted, "a cancelled caller context aborts before the provider")
	assert.Zero(t, objects.callCount(), "provider must not run once the caller context is done")

	state := readRetryRowState(t, db, 1, key)
	assert.Nil(t, state.LastAttemptAt)
	assert.Nil(t, state.StorageDeletedAt)
}

func TestImageDeletionRetry_PostgresFailedAttemptWriteFailureJoinsError(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	store := NewImageDeletionRetryStore(db)
	key := "trimming-images/1/write-fail.jpg"
	insertTrimmingDeletionTombstone(t, db, 1, key)

	require.NoError(t, db.Exec(`
		CREATE OR REPLACE FUNCTION fail_retry_attempt_write() RETURNS trigger AS $f$
		BEGIN RAISE EXCEPTION 'forced attempt write failure'; END;
		$f$ LANGUAGE plpgsql`).Error)
	require.NoError(t, db.Exec(`
		CREATE CONSTRAINT TRIGGER trg_retry_attempt_fail
		AFTER UPDATE ON trimming_image_deletions
		DEFERRABLE INITIALLY DEFERRED
		FOR EACH ROW EXECUTE FUNCTION fail_retry_attempt_write()`).Error)
	t.Cleanup(func() {
		require.NoError(t, db.Exec("DROP TRIGGER IF EXISTS trg_retry_attempt_fail ON trimming_image_deletions").Error)
		require.NoError(t, db.Exec("DROP FUNCTION IF EXISTS fail_retry_attempt_write").Error)
	})

	objects := &spyObjectStore{errs: map[string]int{key: 1}}
	attempted, err := store.RetryImageDeletion(context.Background(), 1, key, objects)
	require.Error(t, err)
	assert.True(t, attempted, "the provider ran, so the attempt is reported even when persistence fails")
	assert.Error(t, err, "the provider error stays joined and observable")
	assert.True(t, errors.Is(err, context.Canceled) || err.Error() != "",
		"joined error must carry the provider failure")

	require.NoError(t, db.Exec("DROP TRIGGER trg_retry_attempt_fail ON trimming_image_deletions").Error)
	state := readRetryRowState(t, db, 1, key)
	assert.Nil(t, state.LastAttemptAt, "rolled back attempt writes leave the prior state intact")
	assert.Nil(t, state.StorageDeletedAt)
}

func TestImageDeletionRetry_PostgresWorkerRunBudgetAndRecovery(t *testing.T) {
	db := setupTrimmingRetryPostgres(t)
	testdb.SeedClinicsForFK(t, db, 1, 2)
	withImageDeletionDeadlineSeams(t, 350*time.Millisecond, 50*time.Millisecond)
	store := NewImageDeletionRetryStore(db)
	blocking := map[string]bool{
		"trimming-images/1/a-timeout.jpg": true,
		"trimming-images/1/b-timeout.jpg": true,
		"trimming-images/1/c-timeout.jpg": true,
	}
	for key := range blocking {
		insertTrimmingDeletionTombstone(t, db, 1, key)
	}
	healthy := "trimming-images/2/healthy.jpg"
	insertTrimmingDeletionTombstone(t, db, 2, healthy)

	spy := &spyObjectStore{errs: map[string]int{}}
	objects := &mixedDeleteStore{spy: spy, block: blocking}
	worker := NewImageDeletionWorker(store, objects)

	healthyDone := false
	for run := 0; run < 6 && !healthyDone; run++ {
		_, _ = worker.RunOnce(context.Background())
		state := readRetryRowState(t, db, 2, healthy)
		healthyDone = state.StorageDeletedAt != nil
	}
	assert.True(t, healthyDone,
		"repeated runs must reach the healthy key once timeout keys are bounded per key")

	for key := range blocking {
		state := readRetryRowState(t, db, 1, key)
		require.NotNil(t, state.LastAttemptAt,
			"timeout keys keep a durable attempt timestamp for %s", key)
		assert.Nil(t, state.StorageDeletedAt, "timeout keys stay pending")
	}
}
