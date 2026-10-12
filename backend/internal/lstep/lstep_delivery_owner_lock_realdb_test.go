package lstep

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	ownerdomain "github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// deliveryOwnerLockGate wraps the real tag-cache repository and pauses the
// delivery attempt immediately before the owner lock is acquired, so tests can
// commit mutations in the claim→lock window that the in-lock fresh reads must
// observe.
type deliveryOwnerLockGate struct {
	inner          *lstepTagCacheRepository
	entered        chan struct{}
	release        chan struct{}
	once           sync.Once
	findByOwnerErr error
}

func (g *deliveryOwnerLockGate) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error) {
	if g.findByOwnerErr != nil {
		return nil, g.findByOwnerErr
	}
	return g.inner.FindByOwner(ctx, clinicID, ownerID)
}

func (g *deliveryOwnerLockGate) FindOwnerIDsByTag(ctx context.Context, clinicID uint64, tagName string) ([]uint64, error) {
	return g.inner.FindOwnerIDsByTag(ctx, clinicID, tagName)
}

func (g *deliveryOwnerLockGate) WithDeliveryOwnerLock(ctx context.Context, clinicID, ownerID uint64, fn func(context.Context) error) error {
	if g.entered != nil {
		g.once.Do(func() { close(g.entered) })
	}
	if g.release != nil {
		<-g.release
	}
	return g.inner.WithDeliveryOwnerLock(ctx, clinicID, ownerID, fn)
}

func newPausedLockGate(inner *lstepTagCacheRepository) *deliveryOwnerLockGate {
	return &deliveryOwnerLockGate{
		inner:   inner,
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func runDeliveryAttempt(svc *lstepDeliveryTriggerService, ctx context.Context, clientForCall lstepapi.Client, clinicID, ownerID uint64, revalidate deliveryRevalidationHook) chan deliveryAttemptResult {
	done := make(chan deliveryAttemptResult, 1)
	go func() {
		fired, err := svc.processSingleOwner(ctx, clientForCall, clinicID, ownerID, "trigger_x", "tag_x", time.Now(), revalidate)
		done <- deliveryAttemptResult{fired: fired, err: err}
	}()
	return done
}

func awaitProviderOrAttempt(t *testing.T, recorder *deliverySendRecorder, done chan deliveryAttemptResult) {
	t.Helper()
	select {
	case <-recorder.entered:
	case res := <-done:
		t.Fatalf("provider call never started: fired=%v err=%v", res.fired, res.err)
	case <-time.After(30 * time.Second):
		t.Fatal("provider call never reached")
	}
}

// A LINE relink committed in the claim→lock window must be observed by the
// in-lock fresh read: the stale recipient gets nothing and the current
// recipient gets exactly one provider call.
func TestDeliveryOwnerLock_RealDB_RelinkPicksCurrentRecipient(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	ownerA := makeCleanupOwner(t, db, clinicID, "owner-a", "U_stale", false)
	ownerB := makeCleanupOwner(t, db, clinicID, "owner-b", "", false)
	gate := newPausedLockGate(repo)
	recorder := newDeliverySendRecorder()
	svc := newDeliveryOwnerLockService(db, gate, recorder.client())
	ownerRepo := ownerdomain.NewRepository(db, nil)

	done := runDeliveryAttempt(svc, context.Background(), recorder.client(), clinicID, ownerA.ID, nil)
	<-gate.entered
	require.NoError(t, ownerRepo.UpdateLineUserID(context.Background(), clinicID, ownerA.ID, strPtr("V_current")))
	require.NoError(t, ownerRepo.UpdateLineUserID(context.Background(), clinicID, ownerB.ID, strPtr("U_stale")))
	close(gate.release)
	res := <-done

	require.NoError(t, res.err)
	assert.True(t, res.fired)
	assert.Equal(t, 0, recorder.attemptsFor("U_stale"), "stale recipient must never receive the provider call")
	assert.Equal(t, 1, recorder.attemptsFor("V_current"), "current recipient must receive exactly one provider call")
}

// Owner flag mutations committed after the claim but before the owner lock is
// acquired must be re-evaluated inside the lock and block the send.
func TestDeliveryOwnerLock_RealDB_FreshOwnerFlagBlocksSend(t *testing.T) {
	cases := []struct {
		name   string
		column string
		reason string
	}{
		{"wait-time opt-out", "lstep_opt_out", "lstep_opt_out"},
		{"delivery excluded flag", "delivery_excluded", "delivery_excluded_flag"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupDeliveryOwnerLockDB(t)
			const clinicID = uint64(1)
			repo := &lstepTagCacheRepository{db: db}
			enableDeliverySync(t, db, clinicID)
			owner := makeCleanupOwner(t, db, clinicID, "flag-owner", "U_flag", false)
			gate := newPausedLockGate(repo)
			recorder := newDeliverySendRecorder()
			svc := newDeliveryOwnerLockService(db, gate, recorder.client())

			done := runDeliveryAttempt(svc, context.Background(), recorder.client(), clinicID, owner.ID, nil)
			<-gate.entered
			require.NoError(t, db.Exec(fmt.Sprintf("UPDATE owners SET %s = true WHERE id = ?", tc.column), owner.ID).Error)
			close(gate.release)
			res := <-done

			require.NoError(t, res.err)
			assert.False(t, res.fired)
			assert.Equal(t, 0, recorder.totalAttempts(), "owner flagged before lock acquisition must not reach the provider")
			log := deliveryTriggerLogFor(t, db, clinicID, owner.ID, "trigger_x")
			assert.Equal(t, model.TriggerStatusExcluded, log.Status)
			require.NotNil(t, log.ExcludedReason)
			assert.Equal(t, tc.reason, *log.ExcludedReason)
		})
	}
}

// An EXCL tag written in the claim→lock window must stop the send — the cache
// tag set is re-read inside the owner lock.
func TestDeliveryOwnerLock_RealDB_FreshExclTagBlocksSend(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	owner := makeCleanupOwner(t, db, clinicID, "excl-owner", "U_excl", false)
	gate := newPausedLockGate(repo)
	recorder := newDeliverySendRecorder()
	svc := newDeliveryOwnerLockService(db, gate, recorder.client())

	done := runDeliveryAttempt(svc, context.Background(), recorder.client(), clinicID, owner.ID, nil)
	<-gate.entered
	require.NoError(t, repo.UpsertTag(context.Background(), clinicID, owner.ID, exclTagDeliveryStop, "EXCL", "owner requested stop"))
	close(gate.release)
	res := <-done

	require.NoError(t, res.err)
	assert.False(t, res.fired)
	assert.Equal(t, 0, recorder.totalAttempts())
	log := deliveryTriggerLogFor(t, db, clinicID, owner.ID, "trigger_x")
	assert.Equal(t, model.TriggerStatusExcluded, log.Status)
	require.NotNil(t, log.ExcludedReason)
	assert.Equal(t, "excl_tag_delivery_stop", *log.ExcludedReason)
}

// Clinic sync turned OFF in the claim→lock window must block the send — the
// settings row is re-read under a SHARE lock so the OFF write serializes
// against provider execution.
func TestDeliveryOwnerLock_RealDB_ClinicSyncOffBlocksSend(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	owner := makeCleanupOwner(t, db, clinicID, "off-owner", "U_off", false)
	gate := newPausedLockGate(repo)
	recorder := newDeliverySendRecorder()
	svc := newDeliveryOwnerLockService(db, gate, recorder.client())

	done := runDeliveryAttempt(svc, context.Background(), recorder.client(), clinicID, owner.ID, nil)
	<-gate.entered
	setDeliverySyncEnabled(t, db, clinicID, false)
	close(gate.release)
	res := <-done

	require.NoError(t, res.err)
	assert.False(t, res.fired)
	assert.Equal(t, 0, recorder.totalAttempts())
	log := deliveryTriggerLogFor(t, db, clinicID, owner.ID, "trigger_x")
	assert.Equal(t, model.TriggerStatusExcluded, log.Status)
	require.NotNil(t, log.ExcludedReason)
	assert.Equal(t, "lstep_sync_disabled", *log.ExcludedReason)
}

// Owner deleted in the claim→lock window makes the in-lock fresh read fail —
// a known-unsent (retryable) outcome, never an exclusion or a send.
func TestDeliveryOwnerLock_RealDB_OwnerDeletedFailsKnownUnsent(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	owner := makeCleanupOwner(t, db, clinicID, "deleted-owner", "U_gone", false)
	gate := newPausedLockGate(repo)
	recorder := newDeliverySendRecorder()
	svc := newDeliveryOwnerLockService(db, gate, recorder.client())

	done := runDeliveryAttempt(svc, context.Background(), recorder.client(), clinicID, owner.ID, nil)
	<-gate.entered
	require.NoError(t, db.Delete(owner).Error)
	close(gate.release)
	res := <-done

	require.Error(t, res.err)
	assert.False(t, res.fired)
	assert.Equal(t, 0, recorder.totalAttempts())
	log := deliveryTriggerLogFor(t, db, clinicID, owner.ID, "trigger_x")
	assert.Equal(t, model.TriggerStatusFailed, log.Status)
	require.NotNil(t, log.ExcludedReason)
	assert.Equal(t, preSendRevalidationFailure, *log.ExcludedReason)
}

// A fresh-read failure inside the lock is a known-unsent blocking-free error —
// nothing reaches the provider and the claim is retryable.
func TestDeliveryOwnerLock_RealDB_FreshReadFailureFailsKnownUnsent(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	owner := makeCleanupOwner(t, db, clinicID, "readfail-owner", "U_rf", false)
	gate := newPausedLockGate(repo)
	recorder := newDeliverySendRecorder()
	svc := newDeliveryOwnerLockService(db, gate, recorder.client())

	done := runDeliveryAttempt(svc, context.Background(), recorder.client(), clinicID, owner.ID, nil)
	<-gate.entered
	gate.findByOwnerErr = errors.New("tag cache read blew up")
	close(gate.release)
	res := <-done

	require.Error(t, res.err)
	assert.False(t, res.fired)
	assert.Equal(t, 0, recorder.totalAttempts())
	log := deliveryTriggerLogFor(t, db, clinicID, owner.ID, "trigger_x")
	assert.Equal(t, model.TriggerStatusFailed, log.Status)
	require.NotNil(t, log.ExcludedReason)
	assert.Equal(t, preSendRevalidationFailure, *log.ExcludedReason)
}

// The owner row lock is held across the provider call: a relink writer, an
// EXCL tag writer, and a clinic-OFF writer must all block until the provider
// returns.
func TestDeliveryOwnerLock_RealDB_ProviderCallSerializesWriters(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	owner := makeCleanupOwner(t, db, clinicID, "barrier-owner", "U_barrier", false)
	recorder := newDeliverySendRecorder()
	recorder.entered = make(chan struct{})
	recorder.release = make(chan struct{})
	svc := newDeliveryOwnerLockService(db, repo, recorder.client())
	ownerRepo := ownerdomain.NewRepository(db, nil)

	done := runDeliveryAttempt(svc, context.Background(), recorder.client(), clinicID, owner.ID, nil)
	awaitProviderOrAttempt(t, recorder, done)

	writerDone := func(name string, fn func() error) chan error {
		ch := make(chan error, 1)
		go func() { ch <- fn() }()
		return ch
	}
	relinkDone := writerDone("relink", func() error {
		return ownerRepo.UpdateLineUserID(context.Background(), clinicID, owner.ID, strPtr("V_after"))
	})
	exclDone := writerDone("excl-tag", func() error {
		return repo.UpsertTag(context.Background(), clinicID, owner.ID, exclTagDeliveryStop, "EXCL", "stop")
	})
	exclDeleteDone := writerDone("excl-delete", func() error {
		return repo.DeleteAllByOwner(context.Background(), clinicID, owner.ID)
	})
	settingsDone := writerDone("settings-off", func() error {
		_, err := NewLstepSyncSettingsRepository(db).Upsert(context.Background(), &model.LstepSettings{
			ClinicID:      clinicID,
			IsSyncEnabled: false,
		})
		return err
	})

	for name, ch := range map[string]chan error{
		"relink": relinkDone, "excl-tag": exclDone, "excl-delete": exclDeleteDone, "settings-off": settingsDone,
	} {
		select {
		case err := <-ch:
			t.Fatalf("%s writer completed while provider call in-flight: %v", name, err)
		case <-time.After(300 * time.Millisecond):
		}
	}
	close(recorder.release)

	for name, ch := range map[string]chan error{
		"relink": relinkDone, "excl-tag": exclDone, "excl-delete": exclDeleteDone, "settings-off": settingsDone,
	} {
		select {
		case err := <-ch:
			assert.NoError(t, err, name)
		case <-time.After(10 * time.Second):
			t.Fatalf("%s writer still blocked after provider return", name)
		}
	}
	res := <-done
	assert.True(t, res.fired)
	assert.Equal(t, 1, recorder.attemptsFor("U_barrier"))
}

// With MaxOpenConns=1 the lock transaction and all inner reads share one
// connection — no nested-connection hang.
func TestDeliveryOwnerLock_RealDB_PoolSizeOneDoesNotDeadlock(t *testing.T) {
	shared := setupDeliveryOwnerLockDB(t)
	pool1 := openDeliveryPoolDB(t, 1)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: pool1}
	enableDeliverySync(t, shared, clinicID)
	owner := makeCleanupOwner(t, shared, clinicID, "pool1-owner", "U_pool1", false)
	recorder := newDeliverySendRecorder()
	svc := newDeliveryOwnerLockService(pool1, repo, recorder.client())

	done := make(chan deliveryAttemptResult, 1)
	go func() {
		fired, err := svc.processSingleOwner(context.Background(), recorder.client(), clinicID, owner.ID, "trigger_x", "tag_x", time.Now(), nil)
		done <- deliveryAttemptResult{fired: fired, err: err}
	}()
	select {
	case res := <-done:
		require.NoError(t, res.err)
		assert.True(t, res.fired)
		assert.Equal(t, 1, recorder.attemptsFor("U_pool1"))
	case <-time.After(30 * time.Second):
		t.Fatal("delivery deadlocked with MaxOpenConns=1")
	}
	log := deliveryTriggerLogFor(t, shared, clinicID, owner.ID, "trigger_x")
	assert.Equal(t, model.TriggerStatusFired, log.Status)
}

// The tentative revalidation hook must run inside the owner serialization
// transaction so its intent/pet locks join the same canonical lock order.
func TestDeliveryOwnerLock_RealDB_RevalidateRunsInsideOwnerTx(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	owner := makeCleanupOwner(t, db, clinicID, "hook-owner", "U_hook", false)
	recorder := newDeliverySendRecorder()
	svc := newDeliveryOwnerLockService(db, repo, recorder.client())

	var ambientTx *gorm.DB
	revalidate := func(ctx context.Context, _ uint64, send func(context.Context) error) (bool, string, error) {
		ambientTx = persistence.TxFromContext(ctx)
		if err := send(ctx); err != nil {
			return false, "", err
		}
		return true, "", nil
	}
	fired, err := svc.processSingleOwner(context.Background(), recorder.client(), clinicID, owner.ID, "trigger_x", "tag_x", time.Now(), revalidate)
	require.NoError(t, err)
	assert.True(t, fired)
	require.NotNil(t, ambientTx, "revalidation hook must run inside the owner lock transaction")
	assert.Equal(t, 1, recorder.attemptsFor("U_hook"))
}

func openDeliveryPoolDB(t *testing.T, maxConns int) *gorm.DB {
	t.Helper()
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "db"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "ekarte_user"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "ekarte_password"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "ekarte_db"
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName+"_test")
	if envDSN := os.Getenv("TEST_DATABASE_URL"); envDSN != "" {
		dsn = envDSN
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(maxConns)
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

// MarkKnownUnsent flips only a genuinely-unsent claim to the retryable failed
// marker; fired, suppressed, and other completed outcomes are never rewritten.
func TestLstepDeliveryTriggerLogRepository_MarkKnownUnsent_RealDB(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := NewLstepDeliveryTriggerLogRepository(db)
	asOf := time.Now()

	newLog := func(status string, mutate func(*model.LstepDeliveryTriggerLog)) *model.LstepDeliveryTriggerLog {
		row := &model.LstepDeliveryTriggerLog{
			OwnerID:     9001,
			ClinicID:    clinicID,
			TriggerType: "trigger_x",
			ScheduledAt: asOf,
			Status:      status,
		}
		if mutate != nil {
			mutate(row)
		}
		require.NoError(t, db.Create(row).Error)
		return row
	}
	readBack := func(id uint64) *model.LstepDeliveryTriggerLog {
		var row model.LstepDeliveryTriggerLog
		require.NoError(t, db.First(&row, id).Error)
		return &row
	}

	t.Run("scheduled claim becomes retryable failed", func(t *testing.T) {
		row := newLog(model.TriggerStatusScheduled, nil)
		require.NoError(t, repo.MarkKnownUnsent(context.Background(), clinicID, row.ID, preSendRevalidationFailure))
		got := readBack(row.ID)
		assert.Equal(t, model.TriggerStatusFailed, got.Status)
		require.NotNil(t, got.ExcludedReason)
		assert.Equal(t, preSendRevalidationFailure, *got.ExcludedReason)
	})

	t.Run("fired outcome is never overwritten", func(t *testing.T) {
		now := time.Now()
		row := newLog(model.TriggerStatusFired, func(r *model.LstepDeliveryTriggerLog) { r.FiredAt = &now })
		require.Error(t, repo.MarkKnownUnsent(context.Background(), clinicID, row.ID, preSendRevalidationFailure))
		got := readBack(row.ID)
		assert.Equal(t, model.TriggerStatusFired, got.Status)
		assert.Nil(t, got.ExcludedReason)
	})

	t.Run("suppressed outcome is never overwritten", func(t *testing.T) {
		reason := "higher priority fired"
		row := newLog(model.TriggerStatusScheduled, func(r *model.LstepDeliveryTriggerLog) {
			r.SuppressedByPriority = true
			r.SuppressionReason = &reason
		})
		require.Error(t, repo.MarkKnownUnsent(context.Background(), clinicID, row.ID, preSendRevalidationFailure))
		got := readBack(row.ID)
		assert.Equal(t, model.TriggerStatusScheduled, got.Status)
		assert.True(t, got.SuppressedByPriority)
		assert.Nil(t, got.ExcludedReason)
	})

	t.Run("excluded outcome is never overwritten", func(t *testing.T) {
		reason := "lstep_opt_out"
		row := newLog(model.TriggerStatusExcluded, func(r *model.LstepDeliveryTriggerLog) { r.ExcludedReason = &reason })
		require.Error(t, repo.MarkKnownUnsent(context.Background(), clinicID, row.ID, preSendRevalidationFailure))
		got := readBack(row.ID)
		assert.Equal(t, model.TriggerStatusExcluded, got.Status)
		require.NotNil(t, got.ExcludedReason)
		assert.Equal(t, "lstep_opt_out", *got.ExcludedReason)
	})

	t.Run("missing row errors", func(t *testing.T) {
		require.Error(t, repo.MarkKnownUnsent(context.Background(), clinicID, 999999, preSendRevalidationFailure))
	})
}

// EXCL tag cache writes serialize on the owner row lock — a writer holding
// that lock (the delivery transaction) blocks EXCL upsert/delete/delete-all
// until it releases, while non-EXCL writes proceed untouched.
func TestExclTagCacheWrite_RealDB_SerializesWithOwnerLock(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	owner := makeCleanupOwner(t, db, clinicID, "lock-holder", "U_hold", false)

	holder := db.Begin()
	require.NoError(t, holder.Error)
	require.NoError(t, holder.Exec(
		"SELECT id FROM owners WHERE clinic_id = ? AND id = ? FOR UPDATE", clinicID, owner.ID).Error)
	defer func() { _ = holder.Rollback() }()

	assertBlocked := func(name string, ch chan error) {
		t.Helper()
		select {
		case err := <-ch:
			t.Fatalf("%s completed while owner row lock held: %v", name, err)
		case <-time.After(300 * time.Millisecond):
		}
	}
	run := func(fn func() error) chan error {
		ch := make(chan error, 1)
		go func() { ch <- fn() }()
		return ch
	}

	nonExcl := run(func() error {
		return repo.UpsertTag(context.Background(), clinicID, owner.ID, "NEXT_VISIT_2026-10-20", "NEXT_VISIT", "")
	})
	select {
	case err := <-nonExcl:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("non-EXCL tag write must not take the owner row lock")
	}

	exclUpsert := run(func() error {
		return repo.UpsertTag(context.Background(), clinicID, owner.ID, exclTagDeliveryStop, "EXCL", "stop")
	})
	exclDelete := run(func() error {
		return repo.DeleteTag(context.Background(), clinicID, owner.ID, exclTagDeliveryStop)
	})
	deleteAll := run(func() error {
		return repo.DeleteAllByOwner(context.Background(), clinicID, owner.ID)
	})
	assertBlocked("EXCL upsert", exclUpsert)
	assertBlocked("EXCL delete", exclDelete)
	assertBlocked("delete-all", deleteAll)

	require.NoError(t, holder.Commit().Error)
	for name, ch := range map[string]chan error{
		"EXCL upsert": exclUpsert, "EXCL delete": exclDelete, "delete-all": deleteAll,
	} {
		select {
		case err := <-ch:
			assert.NoError(t, err, name)
		case <-time.After(10 * time.Second):
			t.Fatalf("%s still blocked after holder commit", name)
		}
	}
}
