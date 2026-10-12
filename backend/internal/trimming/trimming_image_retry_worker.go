package trimming

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/animal-ekarte/backend/internal/persistence"
)

const (
	imageDeletionWorkerInterval   = 5 * time.Minute
	imageDeletionRunTimeout       = 120 * time.Second
	imageDeletionKeyTimeout       = 10 * time.Second
	imageDeletionMaxClinicsPerRun = 100
	imageDeletionMaxKeysPerRun    = 100
)

type ImageDeletionObjectStore interface {
	Delete(context.Context, string) error
}

type ImageDeletionRetryStore interface {
	ListClinicIDsAfter(context.Context, uint64, int) ([]uint64, error)
	ListPendingImageKeys(context.Context, uint64, int) ([]string, error)
	RetryImageDeletion(context.Context, uint64, string, ImageDeletionObjectStore) (bool, error)
}

type ImageDeletionRunResult struct {
	Processed int
	Succeeded int
	Failed    int
	Skipped   int
}

type ImageDeletionWorker struct {
	store       ImageDeletionRetryStore
	objects     ImageDeletionObjectStore
	tick        <-chan time.Time
	mu          sync.Mutex
	afterClinic uint64
}

func NewImageDeletionWorker(store ImageDeletionRetryStore, objects ImageDeletionObjectStore) *ImageDeletionWorker {
	return &ImageDeletionWorker{store: store, objects: objects}
}

func (w *ImageDeletionWorker) Start(ctx context.Context) func() {
	tick := w.tick
	var stop func()
	if tick == nil {
		ticker := time.NewTicker(imageDeletionWorkerInterval)
		tick = ticker.C
		stop = ticker.Stop
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		if stop != nil {
			defer stop()
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick:
				result, err := w.RunOnce(ctx)
				attrs := []any{
					slog.Int("processed", result.Processed),
					slog.Int("succeeded", result.Succeeded),
					slog.Int("failed", result.Failed),
					slog.Int("skipped", result.Skipped),
					slog.Bool("errored", err != nil),
				}
				switch {
				case err != nil && result.Failed > 0:
					slog.ErrorContext(ctx, "trimming image deletion retry run completed with failed keys", attrs...)
				case err != nil:
					slog.ErrorContext(ctx, "trimming image deletion retry run failed", attrs...)
				default:
					slog.InfoContext(ctx, "trimming image deletion retry run completed", attrs...)
				}
			}
		}
	}()
	return func() { <-done }
}

func (w *ImageDeletionWorker) RunOnce(ctx context.Context) (ImageDeletionRunResult, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	var result ImageDeletionRunResult
	if w.store == nil || w.objects == nil {
		return result, errors.New("trimming image deletion worker requires a store and an object store")
	}
	ctx = persistence.WithTxValue(ctx, nil)
	runCtx, cancel := context.WithTimeout(ctx, imageDeletionRunTimeout)
	defer cancel()
	if err := runCtx.Err(); err != nil {
		return result, err
	}
	clinicIDs, err := w.store.ListClinicIDsAfter(runCtx, w.afterClinic, imageDeletionMaxClinicsPerRun)
	if err != nil {
		slog.ErrorContext(runCtx, "trimming image deletion retry clinic scan failed",
			slog.Int("processed", result.Processed))
		return result, err
	}
	if len(clinicIDs) == 0 {
		w.afterClinic = 0
		return result, nil
	}
	var errs []error
	selected := 0
	budgetExhausted := false
	for _, clinicID := range clinicIDs {
		if err := runCtx.Err(); err != nil {
			return result, errors.Join(append(errs, err)...)
		}
		if selected >= imageDeletionMaxKeysPerRun {
			budgetExhausted = true
			break
		}
		w.afterClinic = clinicID
		keys, err := w.store.ListPendingImageKeys(runCtx, clinicID, imageDeletionMaxKeysPerRun-selected)
		if err != nil {
			slog.ErrorContext(runCtx, "trimming image deletion retry key scan failed",
				slog.Uint64("clinic_id", clinicID))
			errs = append(errs, err)
			continue
		}
		for _, key := range keys {
			if err := runCtx.Err(); err != nil {
				return result, errors.Join(append(errs, err)...)
			}
			if selected >= imageDeletionMaxKeysPerRun {
				budgetExhausted = true
				break
			}
			selected++
			keyCtx, keyCancel := context.WithTimeout(runCtx, imageDeletionKeyTimeout)
			attempted, err := w.store.RetryImageDeletion(keyCtx, clinicID, key, w.objects)
			keyCancel()
			switch {
			case err != nil:
				result.Processed++
				result.Failed++
				errs = append(errs, err)
				slog.ErrorContext(runCtx, "trimming image deletion retry failed for image key",
					slog.Uint64("clinic_id", clinicID),
					slog.Int("processed", result.Processed),
					slog.Int("failed", result.Failed))
			case attempted:
				result.Processed++
				result.Succeeded++
			default:
				result.Skipped++
			}
		}
		if budgetExhausted {
			break
		}
	}
	if len(clinicIDs) < imageDeletionMaxClinicsPerRun && !budgetExhausted {
		w.afterClinic = 0
	}
	return result, errors.Join(errs...)
}
