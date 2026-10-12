package trimming

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

type trimmingTagSyncFunc func(context.Context, uint64, uint64) error

func (fn trimmingTagSyncFunc) SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error {
	return fn(ctx, clinicID, ownerID)
}

type trimmingSyncTraceKey struct{}

func TestTrimmingNextVisitTagSync_RequestCancellation(t *testing.T) {
	for _, tt := range []struct {
		name    string
		expired bool
	}{
		{name: "request already canceled"},
		{name: "request deadline already expired", expired: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			requestCtx := context.WithValue(context.Background(), trimmingSyncTraceKey{}, "synthetic-request")
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
			syncer := trimmingTagSyncFunc(func(ctx context.Context, clinicID, ownerID uint64) error {
				if err := ctx.Err(); err != nil {
					return err
				}
				assert.Equal(t, uint64(3), clinicID)
				assert.Equal(t, uint64(8), ownerID)
				assert.Equal(t, "synthetic-request", ctx.Value(trimmingSyncTraceKey{}))
				assert.Nil(t, persistence.TxFromContext(ctx))
				deadline, ok := ctx.Deadline()
				require.True(t, ok, "post-commit work must remain bounded")
				assert.WithinDuration(t, time.Now().Add(30*time.Second), deadline, time.Second)
				done = ctx.Done()
				completed = true
				return nil
			})

			(&trimmingService{nextVisitTagSync: syncer}).syncNextVisitTagsBestEffort(requestCtx, 3, ptrUint64(8))

			require.True(t, completed, "request cancellation must not suppress committed tag sync")
			select {
			case <-done:
			default:
				t.Fatal("post-commit context must be canceled on return")
			}
		})
	}
}

func TestTrimmingNextVisitTagSync_BatchHasOneBudget(t *testing.T) {
	requestCtx, cancelRequest := context.WithCancel(context.Background())
	defer cancelRequest()
	requestCtx = context.WithValue(requestCtx, trimmingSyncTraceKey{}, "synthetic-batch")
	var owners []uint64
	var firstDeadline time.Time
	var firstDone <-chan struct{}
	syncer := trimmingTagSyncFunc(func(ctx context.Context, clinicID, ownerID uint64) error {
		assert.Equal(t, uint64(5), clinicID)
		assert.Equal(t, "synthetic-batch", ctx.Value(trimmingSyncTraceKey{}))
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

	(&trimmingService{nextVisitTagSync: syncer}).syncNextVisitTagsBestEffort(requestCtx, 5, ptrUint64(8), nil, ptrUint64(0), ptrUint64(8), ptrUint64(9))

	assert.Equal(t, []uint64{8, 9}, owners, "best-effort must continue after a sync error and deduplicate")
	select {
	case <-firstDone:
	default:
		t.Fatal("shared post-commit context must be canceled on return")
	}
}

type trimmingSyncLogHandler struct {
	handleFn func(context.Context, slog.Record) error
}

func (h *trimmingSyncLogHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *trimmingSyncLogHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handleFn(ctx, r)
}
func (h *trimmingSyncLogHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *trimmingSyncLogHandler) WithGroup(string) slog.Handler      { return h }

func TestTrimmingNextVisitTagSync_FailureLogKeepsPostCommitScope(t *testing.T) {
	requestCtx, cancel := context.WithCancel(context.WithValue(context.Background(), trimmingSyncTraceKey{}, "synthetic-log-trace"))
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
	slog.SetDefault(slog.New(&trimmingSyncLogHandler{handleFn: func(ctx context.Context, record slog.Record) error {
		assert.NoError(t, ctx.Err())
		assert.Equal(t, "synthetic-log-trace", ctx.Value(trimmingSyncTraceKey{}))
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
	syncer := trimmingTagSyncFunc(func(context.Context, uint64, uint64) error {
		return errors.New("synthetic sync failure")
	})

	(&trimmingService{nextVisitTagSync: syncer}).syncNextVisitTagsBestEffort(requestCtx, 3, ptrUint64(8))

	assert.True(t, logged)
}

func TestTrimmingNextVisitTagSync_MutationCanceledAfterCommit(t *testing.T) {
	for _, operation := range []string{"update", "delete"} {
		for _, failCommit := range []bool{false, true} {
			name := operation + " committed"
			if failCommit {
				name = operation + " commit failed"
			}
			t.Run(name, func(t *testing.T) {
				f := newTrimmingTagMutationFixture(t)
				if failCommit {
					f.failure = "commit"
				}
				requestCtx, cancel := context.WithCancel(context.WithValue(context.Background(), trimmingSyncTraceKey{}, "synthetic-mutation"))
				defer cancel()
				requestCtx = persistence.WithTxValue(requestCtx, &gorm.DB{})
				f.afterCommit = cancel
				var syncDone <-chan struct{}
				f.syncer.checkContext = func(ctx context.Context) {
					assert.NoError(t, ctx.Err(), "committed mutation must not inherit canceled request")
					assert.Nil(t, persistence.TxFromContext(ctx))
					assert.Equal(t, "synthetic-mutation", ctx.Value(trimmingSyncTraceKey{}))
					_, ok := ctx.Deadline()
					assert.True(t, ok)
					syncDone = ctx.Done()
				}
				// A tag service failure must leave the committed mutation successful and
				// must not suppress the next owner following a transfer.
				f.syncer.err = errors.New("synthetic tag service failure")
				svc := f.service(true)
				var err error
				var wantOwners []uint64
				if operation == "update" {
					f.after.OwnerID = ptrUint64(8)
					_, _, err = svc.Update(requestCtx, 3, 99, &UpdateTrimmingInput{ActorID: ptrUint64(1), PetID: ptrUint64(6)})
					wantOwners = []uint64{7, 8}
				} else {
					err = svc.Delete(requestCtx, 3, 99, ptrUint64(1))
					wantOwners = []uint64{7}
				}
				if failCommit {
					require.Error(t, err)
					assert.Empty(t, f.syncer.calls)
					assert.Nil(t, syncDone)
				} else {
					require.NoError(t, err)
					require.ErrorIs(t, requestCtx.Err(), context.Canceled)
					assert.Equal(t, wantOwners, f.syncer.calls)
					select {
					case <-syncDone:
					default:
						t.Fatal("post-commit timeout must be canceled on return")
					}
				}
			})
		}
	}
}

func TestTrimmingNextVisitTagSync_MissingTargets(t *testing.T) {
	for _, tt := range []struct {
		name          string
		missingSyncer bool
	}{
		{name: "nil syncer", missingSyncer: true},
		{name: "nil and zero owners"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			svc := &trimmingService{}
			owners := []*uint64{nil, ptrUint64(0)}
			if !tt.missingSyncer {
				svc.nextVisitTagSync = trimmingTagSyncFunc(func(context.Context, uint64, uint64) error { called = true; return nil })
			} else {
				owners = append(owners, ptrUint64(8))
			}
			svc.syncNextVisitTagsBestEffort(context.Background(), 3, owners...)
			assert.False(t, called)
		})
	}
}
