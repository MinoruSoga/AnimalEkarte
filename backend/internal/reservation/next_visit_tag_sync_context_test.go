package reservation

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/persistence"
)

type nextVisitTagSyncFunc func(context.Context, uint64, uint64) error

func (fn nextVisitTagSyncFunc) SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error {
	return fn(ctx, clinicID, ownerID)
}

type nextVisitTraceKey struct{}

func TestNextVisitTagSync_RequestCancellation(t *testing.T) {
	for _, tt := range []struct {
		name    string
		expired bool
	}{
		{name: "request already canceled"},
		{name: "request deadline already expired", expired: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			requestCtx := context.WithValue(context.Background(), nextVisitTraceKey{}, "synthetic-request")
			var cancel context.CancelFunc
			if tt.expired {
				requestCtx, cancel = context.WithDeadline(requestCtx, time.Now().Add(-time.Second))
			} else {
				requestCtx, cancel = context.WithCancel(requestCtx)
				cancel()
			}
			defer cancel()
			require.Error(t, requestCtx.Err())
			// A completed ambient transaction must not escape into post-commit reads.
			requestCtx = persistence.WithTxValue(requestCtx, &gorm.DB{})
			completed := false
			var done <-chan struct{}
			syncer := nextVisitTagSyncFunc(func(ctx context.Context, clinicID, ownerID uint64) error {
				if err := ctx.Err(); err != nil {
					return err
				}
				assert.Equal(t, uint64(3), clinicID)
				assert.Equal(t, uint64(8), ownerID)
				assert.Equal(t, "synthetic-request", ctx.Value(nextVisitTraceKey{}))
				assert.Nil(t, persistence.TxFromContext(ctx))
				deadline, ok := ctx.Deadline()
				require.True(t, ok, "post-commit work must remain bounded")
				assert.WithinDuration(t, time.Now().Add(30*time.Second), deadline, time.Second)
				done = ctx.Done()
				completed = true
				return nil
			})

			syncNextVisitTagBestEffort(requestCtx, syncer, 3, uint64Ptr(8))

			require.True(t, completed, "request cancellation must not suppress committed tag sync")
			select {
			case <-done:
			default:
				t.Fatal("post-commit context must be canceled on return")
			}
		})
	}
}

func TestNextVisitTagSync_BatchHasOneBudget(t *testing.T) {
	requestCtx, cancelRequest := context.WithCancel(context.Background())
	defer cancelRequest()
	requestCtx = context.WithValue(requestCtx, nextVisitTraceKey{}, "synthetic-batch")
	var owners []uint64
	var firstDeadline time.Time
	var firstDone <-chan struct{}
	syncer := nextVisitTagSyncFunc(func(ctx context.Context, clinicID, ownerID uint64) error {
		assert.Equal(t, uint64(5), clinicID)
		assert.Equal(t, "synthetic-batch", ctx.Value(nextVisitTraceKey{}))
		require.NoError(t, ctx.Err())
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		if len(owners) == 0 {
			firstDeadline, firstDone = deadline, ctx.Done()
			cancelRequest() // The request can also disconnect during synchronization.
			assert.WithinDuration(t, time.Now().Add(30*time.Second), deadline, time.Second)
		} else {
			assert.Equal(t, firstDeadline, deadline, "later owners must not renew the timeout")
			assert.Equal(t, firstDone, ctx.Done())
		}
		require.NoError(t, ctx.Err())
		owners = append(owners, ownerID)
		return errors.New("synthetic tag service failure")
	})

	syncNextVisitTagsBestEffort(requestCtx, syncer, 5, uint64Ptr(8), nil, uint64Ptr(0), uint64Ptr(8), uint64Ptr(9))

	assert.Equal(t, []uint64{8, 9}, owners, "best-effort must continue after a sync error and deduplicate")
	select {
	case <-firstDone:
	default:
		t.Fatal("shared post-commit context must be canceled on return")
	}
}

func TestLiffService_CancelReservation_RequestCanceledAfterCommit(t *testing.T) {
	for _, tt := range []struct {
		name     string
		writeErr error
	}{
		{name: "committed cancellation still synchronizes"},
		{name: "failed cancellation does not synchronize", writeErr: errors.New("synthetic commit failure")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			requestCtx, cancelRequest := context.WithCancel(context.Background())
			defer cancelRequest()
			committed, synced := false, false
			svc := &liffService{
				adminRepo: &mockLiffAdminRepository{cancelByIDFn: func(ctx context.Context, clinicID, customerID, id uint64) (*uint64, error) {
					require.NoError(t, ctx.Err())
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, uint64(1), customerID)
					assert.Equal(t, uint64(10), id)
					committed = tt.writeErr == nil
					cancelRequest()
					return uint64Ptr(8), tt.writeErr
				}},
				nextVisitTagSync: nextVisitTagSyncFunc(func(ctx context.Context, clinicID, ownerID uint64) error {
					require.True(t, committed)
					if err := ctx.Err(); err != nil {
						return err
					}
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, uint64(8), ownerID)
					synced = true
					return nil
				}),
			}

			err := svc.CancelReservation(requestCtx, 3, 1, 10)
			if tt.writeErr != nil {
				require.ErrorIs(t, err, tt.writeErr)
				assert.False(t, synced)
			} else {
				require.NoError(t, err)
				assert.True(t, synced)
			}
		})
	}
}

type nextVisitSyncLogHandler struct {
	handleFn func(context.Context, slog.Record) error
}

func (h *nextVisitSyncLogHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *nextVisitSyncLogHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handleFn(ctx, r)
}
func (h *nextVisitSyncLogHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *nextVisitSyncLogHandler) WithGroup(string) slog.Handler      { return h }

func TestNextVisitTagSync_FailureLogKeepsPostCommitScope(t *testing.T) {
	requestCtx, cancel := context.WithCancel(context.WithValue(context.Background(), nextVisitTraceKey{}, "synthetic-log-trace"))
	defer cancel()
	cancel()
	original := slog.Default()
	originalOutput, originalFlags := log.Writer(), log.Flags()
	t.Cleanup(func() {
		slog.SetDefault(original)
		log.SetOutput(originalOutput)
		log.SetFlags(originalFlags)
	})
	logged := false
	slog.SetDefault(slog.New(&nextVisitSyncLogHandler{handleFn: func(ctx context.Context, record slog.Record) error {
		assert.NoError(t, ctx.Err())
		assert.Equal(t, "synthetic-log-trace", ctx.Value(nextVisitTraceKey{}))
		assert.Equal(t, slog.LevelWarn, record.Level)
		assert.Equal(t, "next visit tag sync failed (best-effort)", record.Message)
		attrs := make(map[string]slog.Value)
		record.Attrs(func(attr slog.Attr) bool { attrs[attr.Key] = attr.Value; return true })
		assert.Equal(t, uint64(3), attrs["clinic_id"].Uint64())
		assert.Equal(t, uint64(8), attrs["owner_id"].Uint64())
		assert.Equal(t, "synthetic sync failure", attrs["error"].String())
		logged = true
		return nil
	}}))
	syncer := nextVisitTagSyncFunc(func(context.Context, uint64, uint64) error {
		return errors.New("synthetic sync failure")
	})

	syncNextVisitTagBestEffort(requestCtx, syncer, 3, uint64Ptr(8))

	assert.True(t, logged)
}
