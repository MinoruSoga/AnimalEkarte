package trimming

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/persistence"
)

type spyObjectStore struct {
	mu      sync.Mutex
	calls   []string
	errs    map[string]int
	hook    func(key string)
	lastCtx context.Context
}

func (s *spyObjectStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	s.calls = append(s.calls, key)
	s.lastCtx = ctx
	remaining := s.errs[key]
	if remaining > 0 {
		s.errs[key]--
	}
	hook := s.hook
	s.mu.Unlock()
	if hook != nil {
		hook(key)
	}
	if remaining > 0 {
		return fmt.Errorf("storage delete failed for %s", key)
	}
	return nil
}

func (s *spyObjectStore) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

type fakeImageDeletionStore struct {
	mu          sync.Mutex
	clinics     []uint64
	pending     map[uint64][]string
	completed   map[string]bool
	retryHook   func(clinicID uint64, key string, objects ImageDeletionObjectStore) (bool, error)
	retried     []string
	afterSeen   []uint64
	limitSeen   []int
	listClinErr error
	listKeysErr map[uint64]error
	listCalled  chan struct{}
}

func newFakeImageDeletionStore() *fakeImageDeletionStore {
	return &fakeImageDeletionStore{
		pending:    map[uint64][]string{},
		completed:  map[string]bool{},
		listCalled: make(chan struct{}, 16),
	}
}

func (f *fakeImageDeletionStore) ListClinicIDsAfter(_ context.Context, after uint64, limit int) ([]uint64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	select {
	case f.listCalled <- struct{}{}:
	default:
	}
	if f.listClinErr != nil {
		return nil, f.listClinErr
	}
	f.afterSeen = append(f.afterSeen, after)
	f.limitSeen = append(f.limitSeen, limit)
	var ids []uint64
	for _, id := range f.clinics {
		if id > after {
			ids = append(ids, id)
		}
	}
	if len(ids) > limit {
		ids = ids[:limit]
	}
	return ids, nil
}

func (f *fakeImageDeletionStore) ListPendingImageKeys(_ context.Context, clinicID uint64, limit int) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.listKeysErr[clinicID]; err != nil {
		return nil, err
	}
	keys := append([]string(nil), f.pending[clinicID]...)
	if len(keys) > limit {
		keys = keys[:limit]
	}
	return keys, nil
}

func (f *fakeImageDeletionStore) RetryImageDeletion(ctx context.Context, clinicID uint64, key string, objects ImageDeletionObjectStore) (bool, error) {
	f.mu.Lock()
	if f.retryHook != nil {
		f.mu.Unlock()
		return f.retryHook(clinicID, key, objects)
	}
	f.retried = append(f.retried, key)
	f.mu.Unlock()
	if err := objects.Delete(ctx, key); err != nil {
		return true, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	keys := f.pending[clinicID]
	for i, k := range keys {
		if k == key {
			f.pending[clinicID] = append(keys[:i], keys[i+1:]...)
			break
		}
	}
	f.completed[key] = true
	return true, nil
}

func TestImageDeletionWorker_RunOnceSuccessCompletesOnce(t *testing.T) {
	store := newFakeImageDeletionStore()
	store.clinics = []uint64{1}
	store.pending[1] = []string{"trimming-images/1/a.jpg"}
	objects := &spyObjectStore{errs: map[string]int{}}
	worker := NewImageDeletionWorker(store, objects)

	result, err := worker.RunOnce(context.Background())
	require.NoError(t, err)
	assert.Equal(t, ImageDeletionRunResult{Processed: 1, Succeeded: 1}, result)
	assert.Equal(t, []string{"trimming-images/1/a.jpg"}, objects.calls)

	result, err = worker.RunOnce(context.Background())
	require.NoError(t, err)
	assert.Equal(t, ImageDeletionRunResult{}, result, "completed key must not be reprocessed")
	assert.Equal(t, 1, objects.callCount())
}

func TestImageDeletionWorker_FailureKeepsKeyRetryable(t *testing.T) {
	store := newFakeImageDeletionStore()
	store.clinics = []uint64{1}
	store.pending[1] = []string{"trimming-images/1/a.jpg"}
	objects := &spyObjectStore{errs: map[string]int{"trimming-images/1/a.jpg": 1}}
	worker := NewImageDeletionWorker(store, objects)

	result, err := worker.RunOnce(context.Background())
	require.Error(t, err)
	assert.Equal(t, ImageDeletionRunResult{Processed: 1, Failed: 1}, result)

	result, err = worker.RunOnce(context.Background())
	require.NoError(t, err)
	assert.Equal(t, ImageDeletionRunResult{Processed: 1, Succeeded: 1}, result)
	assert.Equal(t, 2, objects.callCount())
}

func TestImageDeletionWorker_BusyAndCompletedKeysCountedSkipped(t *testing.T) {
	store := newFakeImageDeletionStore()
	store.clinics = []uint64{1}
	store.pending[1] = []string{"trimming-images/1/a.jpg", "trimming-images/1/b.jpg"}
	store.retryHook = func(_ uint64, _ string, _ ImageDeletionObjectStore) (bool, error) {
		return false, nil
	}
	objects := &spyObjectStore{errs: map[string]int{}}
	worker := NewImageDeletionWorker(store, objects)

	result, err := worker.RunOnce(context.Background())
	require.NoError(t, err)
	assert.Equal(t, ImageDeletionRunResult{Skipped: 2}, result,
		"busy keys are counted as skipped only, never processed")
	assert.Zero(t, objects.callCount())
}

func TestImageDeletionWorker_SelectedBudgetCoversBusyKeys(t *testing.T) {
	store := newFakeImageDeletionStore()
	store.clinics = []uint64{1, 2}
	var busy []string
	for i := 0; i < imageDeletionMaxKeysPerRun; i++ {
		busy = append(busy, fmt.Sprintf("trimming-images/1/busy%03d.jpg", i))
	}
	store.pending[1] = busy
	store.pending[2] = []string{"trimming-images/2/late.jpg"}
	store.retryHook = func(clinicID uint64, key string, objects ImageDeletionObjectStore) (bool, error) {
		if clinicID == 1 {
			return false, nil
		}
		if err := objects.Delete(context.Background(), key); err != nil {
			return true, err
		}
		return true, nil
	}
	objects := &spyObjectStore{errs: map[string]int{}}
	worker := NewImageDeletionWorker(store, objects)

	result, err := worker.RunOnce(context.Background())
	require.NoError(t, err)
	assert.Equal(t, ImageDeletionRunResult{Skipped: imageDeletionMaxKeysPerRun}, result,
		"the 100-selected budget covers busy keys, so clinic 2 stays untouched this run")
	assert.Zero(t, objects.callCount())

	result, err = worker.RunOnce(context.Background())
	require.NoError(t, err)
	assert.Equal(t, ImageDeletionRunResult{Processed: 1, Succeeded: 1}, result,
		"the next run resumes at clinic 1's successor and reaches clinic 2")
	assert.Equal(t, []string{"trimming-images/2/late.jpg"}, objects.calls)
}

func TestImageDeletionWorker_FairnessAfterFailure(t *testing.T) {
	store := newFakeImageDeletionStore()
	store.clinics = []uint64{1, 2}
	store.pending[1] = []string{"trimming-images/1/bad.jpg", "trimming-images/1/ok.jpg"}
	store.pending[2] = []string{"trimming-images/2/late.jpg"}
	objects := &spyObjectStore{errs: map[string]int{"trimming-images/1/bad.jpg": 1}}
	worker := NewImageDeletionWorker(store, objects)

	result, err := worker.RunOnce(context.Background())
	require.Error(t, err)
	assert.Equal(t, 3, result.Processed, "a failing early key must not starve later keys or clinics")
	assert.Equal(t, 2, result.Succeeded)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t,
		[]string{"trimming-images/1/bad.jpg", "trimming-images/1/ok.jpg", "trimming-images/2/late.jpg"},
		objects.calls)
}

func TestImageDeletionWorker_BoundsKeysAndClinics(t *testing.T) {
	store := newFakeImageDeletionStore()
	objects := &spyObjectStore{errs: map[string]int{}}
	for i := 1; i <= 100; i++ {
		store.clinics = append(store.clinics, uint64(i))
	}
	var keys []string
	for i := 0; i < 150; i++ {
		keys = append(keys, fmt.Sprintf("trimming-images/1/k%03d.jpg", i))
	}
	store.pending[1] = keys
	worker := NewImageDeletionWorker(store, objects)

	result, err := worker.RunOnce(context.Background())
	require.NoError(t, err)
	assert.Equal(t, imageDeletionMaxKeysPerRun, result.Processed, "one run attempts at most 100 keys total")
	assert.Equal(t, 100, objects.callCount())
	assert.Equal(t, 100, store.limitSeen[0], "clinic listing must be capped at 100")
	assert.Equal(t, uint64(0), store.afterSeen[0])

	result, err = worker.RunOnce(context.Background())
	require.NoError(t, err)
	assert.Zero(t, result.Processed)
	assert.Equal(t, uint64(1), store.afterSeen[1], "run scans forward from the truncated clinic")

	result, err = worker.RunOnce(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 50, result.Processed, "the wrapped cursor returns to the remaining keys")
	assert.Equal(t, uint64(0), store.afterSeen[2], "a fully exhausted scan resets the cursor to 0")
}

func TestImageDeletionWorker_ExhaustedScanResetsCursor(t *testing.T) {
	store := newFakeImageDeletionStore()
	store.clinics = []uint64{1}
	store.pending[1] = []string{"trimming-images/1/a.jpg"}
	objects := &spyObjectStore{errs: map[string]int{}}
	worker := NewImageDeletionWorker(store, objects)

	_, err := worker.RunOnce(context.Background())
	require.NoError(t, err)
	_, err = worker.RunOnce(context.Background())
	require.NoError(t, err)
	require.Len(t, store.afterSeen, 2)
	assert.Equal(t, uint64(0), store.afterSeen[0])
	assert.Equal(t, uint64(0), store.afterSeen[1], "cursor wraps to 0 after the scan reaches the end")
}

func TestImageDeletionWorker_MissingDependenciesFailClosed(t *testing.T) {
	for _, tt := range []struct {
		name    string
		store   ImageDeletionRetryStore
		objects ImageDeletionObjectStore
	}{
		{name: "nil store", store: nil, objects: &spyObjectStore{errs: map[string]int{}}},
		{name: "nil objects", store: newFakeImageDeletionStore(), objects: nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			worker := NewImageDeletionWorker(tt.store, tt.objects)
			_, err := worker.RunOnce(context.Background())
			require.Error(t, err)
		})
	}
}

func TestImageDeletionWorker_CancelledContextAbortsBeforeQueries(t *testing.T) {
	store := newFakeImageDeletionStore()
	store.clinics = []uint64{1}
	store.pending[1] = []string{"trimming-images/1/a.jpg"}
	objects := &spyObjectStore{errs: map[string]int{}}
	worker := NewImageDeletionWorker(store, objects)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := worker.RunOnce(ctx)
	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, store.afterSeen, "cancellation must abort before querying")
	assert.Zero(t, objects.callCount())
}

func TestImageDeletionWorker_CancellationInterruptsBetweenKeys(t *testing.T) {
	store := newFakeImageDeletionStore()
	store.clinics = []uint64{1}
	store.pending[1] = []string{"trimming-images/1/a.jpg", "trimming-images/1/b.jpg"}
	ctx, cancel := context.WithCancel(context.Background())
	objects := &spyObjectStore{
		errs: map[string]int{},
		hook: func(key string) {
			if key == "trimming-images/1/a.jpg" {
				cancel()
			}
		},
	}
	worker := NewImageDeletionWorker(store, objects)

	_, err := worker.RunOnce(ctx)
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, []string{"trimming-images/1/a.jpg"}, objects.calls, "no key may be attempted after cancellation")
}

func TestImageDeletionWorker_ListClinicsFailureReturnsError(t *testing.T) {
	store := newFakeImageDeletionStore()
	store.listClinErr = errors.New("clinic scan failed")
	objects := &spyObjectStore{errs: map[string]int{}}
	worker := NewImageDeletionWorker(store, objects)

	_, err := worker.RunOnce(context.Background())
	require.ErrorIs(t, err, store.listClinErr)
	assert.Zero(t, objects.callCount())
}

func TestImageDeletionWorker_StartRunsOnTickAndDrains(t *testing.T) {
	store := newFakeImageDeletionStore()
	store.clinics = []uint64{1}
	store.pending[1] = []string{"trimming-images/1/a.jpg"}
	objects := &spyObjectStore{errs: map[string]int{}}
	worker := NewImageDeletionWorker(store, objects)
	tick := make(chan time.Time, 1)
	worker.tick = tick

	ctx, cancel := context.WithCancel(context.Background())
	drain := worker.Start(ctx)

	assert.Zero(t, len(store.listCalled), "no store query may run before the first tick")

	tick <- time.Now()
	select {
	case <-store.listCalled:
	case <-time.After(2 * time.Second):
		t.Fatal("run did not start on the injected tick")
	}
	require.Eventually(t, func() bool { return objects.callCount() == 1 }, 2*time.Second, 10*time.Millisecond)

	cancel()
	drained := make(chan struct{})
	go func() { drain(); close(drained) }()
	select {
	case <-drained:
	case <-time.After(2 * time.Second):
		t.Fatal("drain did not return after context cancellation")
	}
}

func TestImageDeletionWorker_RunOnceClearsAmbientTxKeepsCancelAndDeadline(t *testing.T) {
	store := newFakeImageDeletionStore()
	store.clinics = []uint64{1}
	store.pending[1] = []string{"trimming-images/1/a.jpg"}
	objects := &spyObjectStore{errs: map[string]int{}}
	worker := NewImageDeletionWorker(store, objects)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	callerDeadline, _ := ctx.Deadline()
	ctx = context.WithValue(ctx, struct{ k string }{"trace"}, "trace-value")
	ctx = persistence.WithTxValue(ctx, &gorm.DB{})

	result, err := worker.RunOnce(ctx)
	require.NoError(t, err)
	assert.Equal(t, ImageDeletionRunResult{Processed: 1, Succeeded: 1}, result)
	require.NotNil(t, objects.lastCtx)
	assert.Nil(t, persistence.TxFromContext(objects.lastCtx),
		"the provider must never observe an ambient caller transaction")
	assert.Equal(t, "trace-value", objects.lastCtx.Value(struct{ k string }{"trace"}),
		"caller values must still propagate to the provider")
	deadline, ok := objects.lastCtx.Deadline()
	require.True(t, ok, "the provider context must carry a deadline")
	assert.True(t, deadline.Equal(callerDeadline) || deadline.Before(callerDeadline),
		"the caller deadline remains honored inside the worker run")
}

func TestImageDeletionWorker_LogsDoNotLeakProviderErrorsOrKeys(t *testing.T) {
	const marker = "SECRET-PROVIDER-MARKER-7f2a"
	store := newFakeImageDeletionStore()
	store.clinics = []uint64{1}
	store.pending[1] = []string{"trimming-images/1/a.jpg"}
	store.retryHook = func(_ uint64, _ string, _ ImageDeletionObjectStore) (bool, error) {
		return true, errors.New("provider failure " + marker)
	}
	objects := &spyObjectStore{errs: map[string]int{}}
	worker := NewImageDeletionWorker(store, objects)

	var logBuf bytes.Buffer
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })

	result, err := worker.RunOnce(context.Background())
	require.Error(t, err)
	assert.Equal(t, ImageDeletionRunResult{Processed: 1, Failed: 1}, result)
	assert.Contains(t, err.Error(), marker, "the provider error stays observable to the caller")
	assert.NotContains(t, logBuf.String(), marker,
		"worker logs must never contain raw provider errors")
	assert.NotContains(t, logBuf.String(), "trimming-images/1/a.jpg",
		"worker logs must never contain object keys")
}

func TestImageDeletionWorker_StartDrainWaitsInFlightKey(t *testing.T) {
	store := newFakeImageDeletionStore()
	store.clinics = []uint64{1}
	store.pending[1] = []string{"trimming-images/1/a.jpg"}
	entered := make(chan struct{}, 1)
	cancelled := make(chan struct{}, 1)
	objects := &ctxCancelObjects{entered: entered, cancelled: cancelled}
	worker := NewImageDeletionWorker(store, objects)
	tick := make(chan time.Time, 1)
	worker.tick = tick

	ctx, cancel := context.WithCancel(context.Background())
	drain := worker.Start(ctx)
	tick <- time.Now()

	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("in-flight key never reached the provider")
	}
	cancel()
	select {
	case <-cancelled:
	case <-time.After(2 * time.Second):
		t.Fatal("provider context was not cancelled by worker shutdown")
	}
	drained := make(chan struct{})
	go func() { drain(); close(drained) }()
	select {
	case <-drained:
	case <-time.After(2 * time.Second):
		t.Fatal("drain returned before the in-flight key finished")
	}
}

type ctxCancelObjects struct {
	entered   chan struct{}
	cancelled chan struct{}
}

func (b *ctxCancelObjects) Delete(ctx context.Context, _ string) error {
	select {
	case b.entered <- struct{}{}:
	default:
	}
	<-ctx.Done()
	select {
	case b.cancelled <- struct{}{}:
	default:
	}
	return ctx.Err()
}
