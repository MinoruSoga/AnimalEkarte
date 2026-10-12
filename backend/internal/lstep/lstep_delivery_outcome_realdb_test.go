package lstep

import (
	"context"
	"errors"
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
	"github.com/animal-ekarte/backend/internal/testdb"
)

// ---- shared real-DB helpers for delivery owner-lock / outcome tests ----

func setupDeliveryOwnerLockDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.LstepTagCache{}, &model.LstepSettings{}, &model.LstepDeliveryTriggerLog{}, &model.PetOwner{}))
	testdb.Truncate(t, db, "lstep_tag_cache", "lstep_settings", "lstep_delivery_trigger_log", "pet_owners")
	testdb.SeedClinicsForFK(t, db, 1, 2)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_test_lstep_tag_cache_conflict
		ON lstep_tag_cache (clinic_id, owner_id, tag_name)`).Error)
	return db
}

func enableDeliverySync(t *testing.T, db *gorm.DB, clinicID uint64) {
	t.Helper()
	_, err := NewLstepSyncSettingsRepository(db).Upsert(context.Background(), &model.LstepSettings{
		ClinicID:      clinicID,
		IsSyncEnabled: true,
	})
	require.NoError(t, err)
}

func setDeliverySyncEnabled(t *testing.T, db *gorm.DB, clinicID uint64, enabled bool) {
	t.Helper()
	_, err := NewLstepSyncSettingsRepository(db).Upsert(context.Background(), &model.LstepSettings{
		ClinicID:      clinicID,
		IsSyncEnabled: enabled,
	})
	require.NoError(t, err)
}

// newDeliveryOwnerLockService wires real repositories; the only fakes are the
// provider client and the settings-service wrapper (delegating to the real
// settings repository so inside-lock re-reads see committed writes).
func newDeliveryOwnerLockService(db *gorm.DB, tagRepo deliveryTagCacheRepository, client lstepapi.Client) *lstepDeliveryTriggerService {
	settingsRepo := NewLstepSyncSettingsRepository(db)
	return &lstepDeliveryTriggerService{
		ownerRepo:      ownerdomain.NewRepository(db, nil),
		tagCacheRepo:   tagRepo,
		triggerLogRepo: NewLstepDeliveryTriggerLogRepository(db),
		settingsSvc: &mockLstepSettingsService{
			isSyncEnabledFn: func(ctx context.Context, clinicID uint64) (bool, error) {
				settings, err := settingsRepo.FindByClinicID(ctx, clinicID)
				if err != nil {
					if apperrors.IsNotFound(err) {
						return false, nil
					}
					return false, err
				}
				return settings.IsSyncEnabled, nil
			},
		},
		clientBuilderFn: func(context.Context, uint64) (lstepapi.Client, error) { return client, nil },
	}
}

type deliverySendRecorder struct {
	mu       sync.Mutex
	attempts map[string]int
	sent     map[string]int
	addErr   error
	entered  chan struct{}
	release  chan struct{}
	once     sync.Once
}

func newDeliverySendRecorder() *deliverySendRecorder {
	return &deliverySendRecorder{
		attempts: make(map[string]int),
		sent:     make(map[string]int),
	}
}

func (r *deliverySendRecorder) client() lstepapi.Client {
	return &mockLstepClientForDelivery{
		addTagFn: func(ctx context.Context, uid, tag string) error {
			if r.entered != nil {
				r.once.Do(func() { close(r.entered) })
			}
			if r.release != nil {
				<-r.release
			}
			r.mu.Lock()
			defer r.mu.Unlock()
			r.attempts[uid]++
			if r.addErr != nil {
				return r.addErr
			}
			r.sent[uid]++
			return nil
		},
	}
}

func (r *deliverySendRecorder) attemptsFor(uid string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.attempts[uid]
}

func (r *deliverySendRecorder) totalAttempts() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	total := 0
	for _, n := range r.attempts {
		total += n
	}
	return total
}

type deliveryAttemptResult struct {
	fired bool
	err   error
}

func deliveryTriggerLogFor(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, triggerType string) *model.LstepDeliveryTriggerLog {
	t.Helper()
	var row model.LstepDeliveryTriggerLog
	require.NoError(t, db.
		Where("clinic_id = ? AND owner_id = ? AND trigger_type = ?", clinicID, ownerID, triggerType).
		Order("id DESC").
		First(&row).Error)
	return &row
}

// ---- D2 outcome durability ----

// A definitely-unsent failure (caller cancellation while the revalidation
// boundary is still open) must leave a durable failed marker on a bounded
// detached context — the same-day slot is then reclaimable and a retry sends
// exactly once.
func TestDeliveryOutcome_RealDB_PreSendCancelMarksRetryableUnsentAndRetrySendsOnce(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	owner := makeCleanupOwner(t, db, clinicID, "cancel-before-send", "U_premark", false)
	recorder := newDeliverySendRecorder()
	svc := newDeliveryOwnerLockService(db, repo, recorder.client())

	ctx, cancel := context.WithCancel(context.Background())
	revalidate := func(_ context.Context, _ uint64, _ func(context.Context) error) (bool, string, error) {
		cancel()
		return false, "", context.Canceled
	}
	fired, err := svc.processSingleOwner(ctx, recorder.client(), clinicID, owner.ID, "trigger_x", "tag_x", time.Now(), revalidate)
	require.Error(t, err)
	assert.False(t, fired)
	assert.Equal(t, 0, recorder.totalAttempts(), "a cancelled pre-send attempt must never reach the provider")

	log := deliveryTriggerLogFor(t, db, clinicID, owner.ID, "trigger_x")
	require.Equal(t, model.TriggerStatusFailed, log.Status, "known-unsent claim must persist failed on a detached context")
	require.NotNil(t, log.ExcludedReason)
	assert.Equal(t, preSendRevalidationFailure, *log.ExcludedReason)
	assert.Nil(t, log.FiredAt)

	fired, err = svc.processSingleOwner(context.Background(), recorder.client(), clinicID, owner.ID, "trigger_x", "tag_x", time.Now(), nil)
	require.NoError(t, err)
	assert.True(t, fired, "the pre_send failed marker must be reclaimable same-day")
	assert.Equal(t, 1, recorder.attemptsFor("U_premark"))
	log = deliveryTriggerLogFor(t, db, clinicID, owner.ID, "trigger_x")
	assert.Equal(t, model.TriggerStatusFired, log.Status)
	assert.NotNil(t, log.FiredAt)
}

// The provider succeeding while the caller context is cancelled (serialization
// commit then fails) must still persist the fired outcome on a detached
// context — the slot stays claimed and is never resent.
func TestDeliveryOutcome_RealDB_PostSendCancelStillPersistsFired(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	owner := makeCleanupOwner(t, db, clinicID, "cancel-after-send", "U_postmark", false)
	recorder := newDeliverySendRecorder()
	recorder.entered = make(chan struct{})
	recorder.release = make(chan struct{})
	svc := newDeliveryOwnerLockService(db, repo, recorder.client())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan deliveryAttemptResult, 1)
	go func() {
		fired, err := svc.processSingleOwner(ctx, recorder.client(), clinicID, owner.ID, "trigger_x", "tag_x", time.Now(), nil)
		done <- deliveryAttemptResult{fired: fired, err: err}
	}()
	select {
	case <-recorder.entered:
	case res := <-done:
		t.Fatalf("provider call never started: fired=%v err=%v", res.fired, res.err)
	case <-time.After(30 * time.Second):
		t.Fatal("provider call never reached")
	}
	cancel()
	close(recorder.release)
	res := <-done

	require.Error(t, res.err, "cancelled serialization must surface, not pretend a clean send")
	assert.False(t, res.fired)
	assert.Equal(t, 1, recorder.attemptsFor("U_postmark"))
	log := deliveryTriggerLogFor(t, db, clinicID, owner.ID, "trigger_x")
	assert.Equal(t, model.TriggerStatusFired, log.Status, "provider success must persist fired despite caller cancellation")
	assert.NotNil(t, log.FiredAt)
}

// Two workers racing the same-day retry of a definitely-unsent claim send
// exactly once — the conditional reclaim hands the slot to one worker only.
func TestDeliveryOutcome_RealDB_TwoRetryWorkersSendOnce(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	owner := makeCleanupOwner(t, db, clinicID, "retry-race", "U_race", false)
	recorder := newDeliverySendRecorder()
	svc := newDeliveryOwnerLockService(db, repo, recorder.client())

	asOf := time.Now()
	reason := preSendRevalidationFailure
	require.NoError(t, db.Create(&model.LstepDeliveryTriggerLog{
		OwnerID:        owner.ID,
		ClinicID:       clinicID,
		TriggerType:    "trigger_x",
		ScheduledAt:    asOf,
		Status:         model.TriggerStatusFailed,
		ExcludedReason: &reason,
	}).Error)

	var wg sync.WaitGroup
	results := make(chan deliveryAttemptResult, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fired, err := svc.processSingleOwner(context.Background(), recorder.client(), clinicID, owner.ID, "trigger_x", "tag_x", asOf, nil)
			results <- deliveryAttemptResult{fired: fired, err: err}
		}()
	}
	wg.Wait()
	close(results)

	errs := 0
	for res := range results {
		if res.err != nil {
			errs++
		}
	}
	assert.Zero(t, errs, "both workers should resolve without surfacing an error")
	assert.Equal(t, 1, recorder.totalAttempts(), "racing retry workers must not double-send")
	log := deliveryTriggerLogFor(t, db, clinicID, owner.ID, "trigger_x")
	assert.Equal(t, model.TriggerStatusFired, log.Status)
	assert.NotNil(t, log.FiredAt)
}

// A provider failure after send started is a blocking outcome — never the
// retryable pre_send marker — and the same-day slot stays claimed.
func TestDeliveryOutcome_RealDB_ProviderFailureIsBlockingNotRetryable(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	owner := makeCleanupOwner(t, db, clinicID, "provider-fail", "U_pfail", false)
	recorder := newDeliverySendRecorder()
	recorder.addErr = errors.New("provider lost mid-flight")
	svc := newDeliveryOwnerLockService(db, repo, recorder.client())

	fired, err := svc.processSingleOwner(context.Background(), recorder.client(), clinicID, owner.ID, "trigger_x", "tag_x", time.Now(), nil)
	require.Error(t, err)
	assert.False(t, fired)
	assert.Equal(t, 1, recorder.totalAttempts())
	log := deliveryTriggerLogFor(t, db, clinicID, owner.ID, "trigger_x")
	require.Equal(t, model.TriggerStatusFailed, log.Status)
	require.NotNil(t, log.ExcludedReason)
	assert.NotEqual(t, preSendRevalidationFailure, *log.ExcludedReason,
		"an attempt that reached the provider must never become a retryable pre-send marker")

	fired, err = svc.processSingleOwner(context.Background(), recorder.client(), clinicID, owner.ID, "trigger_x", "tag_x", time.Now(), nil)
	require.NoError(t, err)
	assert.False(t, fired, "a blocking failure keeps the day slot claimed")
	assert.Equal(t, 1, recorder.totalAttempts(), "no retry must fire after a provider-attempt failure")
}
