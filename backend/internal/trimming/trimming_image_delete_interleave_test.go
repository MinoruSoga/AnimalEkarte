package trimming

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

type deleteGateService struct {
	Service
	mu      sync.Mutex
	events  []string
	entered chan struct{}
	release chan struct{}
	sigOnce sync.Once
}

func (g *deleteGateService) record(event string) {
	g.mu.Lock()
	g.events = append(g.events, event)
	g.mu.Unlock()
}

func (g *deleteGateService) eventLog() []string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return append([]string(nil), g.events...)
}

func (g *deleteGateService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	appt, err := g.Service.GetByID(ctx, clinicID, id)
	if err == nil && appt != nil {
		if style, completed, ok := trimmingStoredImages(appt); ok {
			g.record(fmt.Sprintf("PRE_READ_SNAPSHOT style=%q completed=%q", style, completed))
		}
	}
	return appt, err
}

func (g *deleteGateService) Delete(ctx context.Context, clinicID, id uint64, actorID *uint64) (*TrimmingReplacedImages, error) {
	g.record("DELETE_ENTERED_BEFORE_LOCK")
	g.sigOnce.Do(func() { close(g.entered) })
	select {
	case <-g.release:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	g.record("DELETE_RELEASED_DELEGATE")
	return g.Service.Delete(ctx, clinicID, id, actorID)
}

func (g *deleteGateService) bindImageObjectStore(u FileUploader) {
	if b, ok := g.Service.(imageObjectStoreBinder); ok {
		b.bindImageObjectStore(u)
	}
}

// TestDeleteTrimming_UsesKeysLockedInsideDeleteTransaction は
// pre-transaction スナップショットと delete ロックの間にコミットされた
// 更新キーをクリーンアップが必ず拾うことを確認する。
// typed Service decorator gate が Delete の pre-transaction 入口で
// ブロックし、その間に Update をコミットさせるインターリーブを作る。
// 旧実装は GetByID の pre-read 値 (K1) を使うため K2 を評価しない。
//
// oracle は cleanup service spy が記録する評価対象キーと結果:
// 正しい実装は K2 を評価し、stale pre-read 実装は K1 しか評価しない。
// A3 primary/live-parent 契約では K2 の参照行は delete で soft-delete
// された親 appointment 配下のみとなるため、評価のうえ削除する
// （pg_locks waiter 数推測には依存しない）。
func TestDeleteTrimming_UsesKeysLockedInsideDeleteTransaction(t *testing.T) {
	t.Setenv("TRIMMING_IMAGE_OBJECT_DELETE_ENABLED", "1")
	gin.SetMode(gin.TestMode)
	db := setupImageReferenceTestDB(t)
	require.NoError(t, ensureAutoMigrated(db,
		&model.StaffClinicAssignment{}, &model.LineCustomer{}))
	clinicID := uint64(1)
	keyK1 := "trimming-images/1/stale-prelock.jpg"
	keyK2 := "trimming-images/1/locked-time.jpg"

	appt := makeTrimmingLifecycleAppointment(t, db, clinicID)
	makeLifecycleDetail(t, db, clinicID, appt.ID, keyK1, "")

	svc := newLifecycleService(t, db)
	gate := &deleteGateService{
		Service: svc,
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
	spy := &cleanupEvalSpyService{Service: gate}
	uploader := &fakeTrimmingImageUploader{}
	h := newHandlerWithTrimmingSvc(spy)
	h.WithImageSupport(ImageDependencies{Uploader: uploader})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	releaseOnce := sync.Once{}
	release := func() { releaseOnce.Do(func() { close(gate.release) }) }

	deleteDone := make(chan int, 1)
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", appt.ID)}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/", nil).WithContext(ctx)
		setTrimmingWriteContext(c)
		h.DeleteTrimming(c)
		deleteDone <- c.Writer.Status()
	}()
	// LIFO: setupImageReferenceTestDB の DB cleanup より先に走り、
	// FailNow を含む全 exit で worker を join する。
	t.Cleanup(func() {
		release()
		cancel()
		select {
		case <-workerDone:
		case <-time.After(10 * time.Second):
			t.Errorf("delete worker did not join within bound")
		}
	})

	select {
	case <-gate.entered:
	case <-ctx.Done():
		require.FailNow(t, "Delete did not reach the pre-transaction gate")
	}

	_, _, err := svc.Update(ctx, clinicID, appt.ID, &UpdateTrimmingInput{
		StyleImage: strPtr(keyK2),
		ActorID:    ptrUint64(42),
	})
	require.NoError(t, err)
	gate.record("UPDATE_K2_COMMITTED")
	release()

	var status int
	select {
	case status = <-deleteDone:
	case <-ctx.Done():
		require.FailNow(t, "Delete did not finish")
	}
	require.Equal(t, http.StatusNoContent, status)

	events := gate.eventLog()
	t.Logf("event order: %v", events)
	enteredAt, committedAt, releasedAt := -1, -1, -1
	for i, ev := range events {
		switch ev {
		case "DELETE_ENTERED_BEFORE_LOCK":
			enteredAt = i
		case "UPDATE_K2_COMMITTED":
			committedAt = i
		case "DELETE_RELEASED_DELEGATE":
			releasedAt = i
		}
	}
	require.GreaterOrEqual(t, enteredAt, 0)
	require.GreaterOrEqual(t, committedAt, 0)
	require.GreaterOrEqual(t, releasedAt, 0)
	assert.Less(t, enteredAt, committedAt, "delete must reach the gate before the K2 update commits")
	assert.Less(t, committedAt, releasedAt, "the K2 update must commit before delete runs its lock")

	k2Calls := spy.callsFor(keyK2)
	require.Len(t, k2Calls, 1,
		"delete-side cleanup must evaluate the locked-time key K2 exactly once")
	assert.NoError(t, k2Calls[0].err, "cleanup must not error on K2")
	assert.True(t, k2Calls[0].deleted,
		"K2 references sit only on the just-soft-deleted appointment; A3 primary policy cleans it")
	assert.Contains(t, uploader.deletes, keyK2,
		"K2 object must be physically deleted post-commit, not merely evaluated")
	assert.Empty(t, spy.callsFor(keyK1),
		"a stale pre-lock snapshot key must not drive delete-side cleanup")
}
