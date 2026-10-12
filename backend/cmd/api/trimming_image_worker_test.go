package main

import (
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/infra"
)

type fakeFileUploader struct {
	deleted []string
}

func (f *fakeFileUploader) Upload(context.Context, string, io.Reader, string) (string, error) {
	return "", nil
}

func (f *fakeFileUploader) Delete(_ context.Context, key string) error {
	f.deleted = append(f.deleted, key)
	return nil
}

func (f *fakeFileUploader) GetSignedURL(context.Context, string, time.Duration) (string, error) {
	return "", nil
}

type recordingImageDeletionRunner struct {
	started    chan context.Context
	activeDone chan struct{}
}

func (r *recordingImageDeletionRunner) Start(ctx context.Context) func() {
	select {
	case r.started <- ctx:
	default:
	}
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		<-r.activeDone
		close(done)
	}()
	return func() { <-done }
}

func overrideImageDeletionFactory(t *testing.T, factory func(*gorm.DB, infra.FileUploader) imageDeletionRunner) {
	t.Helper()
	original := newTrimmingImageDeletionWorker
	newTrimmingImageDeletionWorker = factory
	t.Cleanup(func() { newTrimmingImageDeletionWorker = original })
}

func TestStartTrimmingImageDeletionWorker_DisabledReturnsNoop(t *testing.T) {
	var factoryCalls atomic.Int32
	overrideImageDeletionFactory(t, func(*gorm.DB, infra.FileUploader) imageDeletionRunner {
		factoryCalls.Add(1)
		return &recordingImageDeletionRunner{started: make(chan context.Context, 1), activeDone: make(chan struct{})}
	})
	for _, tt := range []struct {
		name string
		cfg  *config.Config
	}{
		{name: "disabled flag", cfg: &config.Config{TrimmingImageCleanupEnabled: false}},
		{name: "nil config", cfg: nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			drain := startTrimmingImageDeletionWorker(context.Background(), tt.cfg, nil, nil)
			require.NotNil(t, drain)
			done := make(chan struct{})
			go func() { drain(); close(done) }()
			select {
			case <-done:
			case <-time.After(500 * time.Millisecond):
				t.Fatal("disabled drain must return without waiting for cancellation")
			}
		})
	}
	require.Zero(t, factoryCalls.Load(), "disabled worker must not construct a runner")
}

func TestStartTrimmingImageDeletionWorker_EnabledBuildsAndStartsRunner(t *testing.T) {
	runner := &recordingImageDeletionRunner{
		started:    make(chan context.Context, 1),
		activeDone: make(chan struct{}),
	}
	var gotDB *gorm.DB
	var gotUploader infra.FileUploader
	overrideImageDeletionFactory(t, func(db *gorm.DB, uploader infra.FileUploader) imageDeletionRunner {
		gotDB = db
		gotUploader = uploader
		return runner
	})
	db := &gorm.DB{}
	uploader := &fakeFileUploader{}

	ctx, cancel := context.WithCancel(context.Background())
	drain := startTrimmingImageDeletionWorker(
		ctx,
		&config.Config{TrimmingImageCleanupEnabled: true},
		db,
		uploader,
	)
	require.NotNil(t, drain)
	require.Same(t, db, gotDB, "enabled worker must receive the configured database handle")
	require.Same(t, uploader, gotUploader, "enabled worker must receive the configured uploader")
	select {
	case startedCtx := <-runner.started:
		require.Same(t, ctx, startedCtx, "the runner must be started with the runtime context")
	case <-time.After(time.Second):
		t.Fatal("enabled runner was never started")
	}

	cancel()
	close(runner.activeDone)
	drained := make(chan struct{})
	go func() { drain(); close(drained) }()
	select {
	case <-drained:
	case <-time.After(2 * time.Second):
		t.Fatal("enabled worker did not drain after context cancellation")
	}
}
