package lstep

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	petdomain "github.com/animal-ekarte/backend/internal/pet"
	"github.com/animal-ekarte/backend/internal/testdb"
)

var (
	deliveryPetSourceBirthday  = time.Date(2020, time.April, 15, 0, 0, 0, 0, time.UTC)
	deliveryPetSiblingBirthday = time.Date(2019, time.October, 3, 0, 0, 0, 0, time.UTC)
	deliveryPetAsOf            = time.Date(2026, time.April, 15, 12, 0, 0, 0, time.UTC)
)

func makeOwnerWithBirthdayPet(t *testing.T, db *gorm.DB, clinicID uint64, name, uid string) (owner *model.Owner, sibling, source *model.Pet) {
	t.Helper()
	owner = makeCleanupOwner(t, db, clinicID, name, uid, false)
	sibling = testdb.MakeSpeciesAndPet(t, db, clinicID, owner.ID, name+"-sibling")
	source = testdb.MakeSpeciesAndPet(t, db, clinicID, owner.ID, name+"-source")
	set := func(petID uint64, birth time.Time) {
		require.NoError(t, db.Model(&model.Pet{}).
			Where("clinic_id = ? AND id = ?", clinicID, petID).
			Select("birth_date", "status").
			Updates(&model.Pet{BirthDate: &birth, Status: model.PetStatusAlive}).Error)
	}
	set(sibling.ID, deliveryPetSiblingBirthday)
	set(source.ID, deliveryPetSourceBirthday)
	return owner, sibling, source
}

func newBirthdayDeliveryService(db *gorm.DB, tagRepo deliveryTagCacheRepository, client lstepapi.Client) *lstepDeliveryTriggerService {
	svc := newDeliveryOwnerLockService(db, tagRepo, client)
	svc.petRepo = petdomain.NewRepository(db)
	return svc
}

func runBirthdayDeliveryAttempt(
	svc *lstepDeliveryTriggerService,
	ctx context.Context,
	client lstepapi.Client,
	clinicID, ownerID uint64,
	asOf time.Time,
) chan deliveryAttemptResult {
	done := make(chan deliveryAttemptResult, 1)
	go func() {
		fired, err := svc.processSingleOwner(ctx, client, clinicID, ownerID,
			model.TriggerTypeBirthdayMessage, model.TriggerTypeBirthdayMessage, asOf, nil,
			svc.petBirthdayEligibility(clinicID, int(asOf.Month()), asOf.Day()))
		done <- deliveryAttemptResult{fired: fired, err: err}
	}()
	return done
}

func runRecordDeathOnDedicatedConnection(ctx context.Context, db *gorm.DB, clinicID, petID uint64) (pidCh chan int64, doneCh chan error) {
	pidCh = make(chan int64, 1)
	doneCh = make(chan error, 1)
	go func() {
		doneCh <- db.WithContext(ctx).Connection(func(conn *gorm.DB) error {
			var pid int64
			if err := conn.Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
				return err
			}
			pidCh <- pid
			return petdomain.NewRepository(conn).RecordDeath(ctx, clinicID, petID, time.Now(), "test death")
		})
	}()
	return pidCh, doneCh
}

func requireBackendBlocked(t *testing.T, ctx context.Context, db *gorm.DB, pid int64) {
	t.Helper()
	require.Eventually(t, func() bool {
		var blocked bool
		if err := db.WithContext(ctx).Raw("SELECT cardinality(pg_blocking_pids(?)) > 0", pid).Scan(&blocked).Error; err != nil {
			return false
		}
		return blocked
	}, 3*time.Second, 20*time.Millisecond, "backend %d never appeared in pg_blocking_pids", pid)
}

func requirePetAlive(t *testing.T, ctx context.Context, db *gorm.DB, clinicID, petID uint64) {
	t.Helper()
	pet, err := petdomain.NewRepository(db).FindByID(ctx, clinicID, petID)
	require.NoError(t, err)
	require.NotNil(t, pet)
	assert.Equal(t, model.PetStatusAlive, pet.Status)
	assert.Nil(t, pet.DeceasedAt)
}

func requirePetDeceased(t *testing.T, ctx context.Context, db *gorm.DB, clinicID, petID uint64) {
	t.Helper()
	pet, err := petdomain.NewRepository(db).FindByID(ctx, clinicID, petID)
	require.NoError(t, err)
	require.NotNil(t, pet)
	assert.Equal(t, model.PetStatusDeceased, pet.Status)
	assert.NotNil(t, pet.DeceasedAt)
}

func TestDeliveryPetSerialization_RealDB_SendFirstBlocksRecordDeath(t *testing.T) {
	const clinicID = uint64(1)
	asOf := deliveryPetAsOf

	runCase := func(t *testing.T, name, uid string, providerErr bool) {
		db := setupDeliveryOwnerLockDB(t)
		repo := &lstepTagCacheRepository{db: db}
		enableDeliverySync(t, db, clinicID)
		owner, sibling, source := makeOwnerWithBirthdayPet(t, db, clinicID, name, uid)
		recorder := newDeliverySendRecorder()
		recorder.entered = make(chan struct{})
		recorder.release = make(chan struct{})
		if providerErr {
			recorder.addErr = errors.New("provider lost mid-flight")
		}
		svc := newBirthdayDeliveryService(db, repo, recorder.client())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		var releaseOnce sync.Once
		release := func() { releaseOnce.Do(func() { close(recorder.release) }) }

		done := runBirthdayDeliveryAttempt(svc, ctx, recorder.client(), clinicID, owner.ID, asOf)
		var deathDone <-chan error
		attemptJoined, deathJoined := false, false
		t.Cleanup(func() {
			release()
			if !attemptJoined {
				select {
				case <-done:
				case <-time.After(15 * time.Second):
					t.Error("delivery goroutine did not finish during cleanup")
				}
			}
			if deathDone != nil && !deathJoined {
				select {
				case <-deathDone:
				case <-time.After(15 * time.Second):
					t.Error("death goroutine did not finish during cleanup")
				}
			}
		})

		select {
		case <-recorder.entered:
		case res := <-done:
			attemptJoined = true
			t.Fatalf("provider call never started: fired=%v err=%v", res.fired, res.err)
		case <-ctx.Done():
			t.Fatal("provider call never reached before deadline")
		}

		var deathPIDCh <-chan int64
		deathPIDCh, deathDone = runRecordDeathOnDedicatedConnection(ctx, db, clinicID, source.ID)
		var deathPID int64
		select {
		case deathPID = <-deathPIDCh:
		case err := <-deathDone:
			deathJoined = true
			t.Fatalf("death connection failed before locking: %v", err)
		case <-ctx.Done():
			t.Fatal("death backend never reported its pid")
		}
		requireBackendBlocked(t, ctx, db, deathPID)
		select {
		case err := <-deathDone:
			deathJoined = true
			t.Fatalf("RecordDeath completed while the provider callback held the family pet locks: %v", err)
		default:
		}

		release()
		select {
		case res := <-done:
			attemptJoined = true
			if providerErr {
				require.Error(t, res.err)
				assert.False(t, res.fired)
			} else {
				require.NoError(t, res.err)
				assert.True(t, res.fired)
			}
		case <-ctx.Done():
			t.Fatal("delivery did not finish after the provider was released")
		}
		select {
		case err := <-deathDone:
			deathJoined = true
			require.NoError(t, err, "RecordDeath must complete once the delivery lock released")
		case <-ctx.Done():
			t.Fatal("RecordDeath still blocked after the provider callback returned")
		}

		assert.Equal(t, 1, recorder.attemptsFor(uid))
		log := deliveryTriggerLogFor(t, db, clinicID, owner.ID, model.TriggerTypeBirthdayMessage)
		if providerErr {
			require.Equal(t, model.TriggerStatusFailed, log.Status)
			require.NotNil(t, log.ExcludedReason)
			assert.NotEqual(t, preSendRevalidationFailure, *log.ExcludedReason)
		} else {
			require.Equal(t, model.TriggerStatusFired, log.Status)
			assert.NotNil(t, log.FiredAt)
		}
		requirePetDeceased(t, ctx, db, clinicID, source.ID)
		requirePetAlive(t, ctx, db, clinicID, sibling.ID)

		if providerErr {
			fired, err := svc.processSingleOwner(ctx, recorder.client(), clinicID, owner.ID,
				model.TriggerTypeBirthdayMessage, model.TriggerTypeBirthdayMessage, asOf, nil,
				svc.petBirthdayEligibility(clinicID, int(asOf.Month()), asOf.Day()))
			require.NoError(t, err)
			assert.False(t, fired)
			assert.Equal(t, 1, recorder.totalAttempts())
		}
	}

	t.Run("provider success", func(t *testing.T) {
		runCase(t, "send-first-owner", "U_petlock", false)
	})
	t.Run("provider error releases the pet lock", func(t *testing.T) {
		runCase(t, "send-first-fail-owner", "U_petlock_fail", true)
	})
}

func TestDeliveryPetSerialization_RealDB_DeathFirstSuppressesBirthdaySend(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	asOf := deliveryPetAsOf
	owner, sibling, source := makeOwnerWithBirthdayPet(t, db, clinicID, "death-first-owner", "U_deathfirst")
	gate := newPausedLockGate(repo)
	recorder := newDeliverySendRecorder()
	svc := newBirthdayDeliveryService(db, gate, recorder.client())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var releaseOnce sync.Once
	releaseGate := func() { releaseOnce.Do(func() { close(gate.release) }) }

	done := runBirthdayDeliveryAttempt(svc, ctx, recorder.client(), clinicID, owner.ID, asOf)
	attemptJoined := false
	t.Cleanup(func() {
		releaseGate()
		if !attemptJoined {
			select {
			case <-done:
			case <-time.After(15 * time.Second):
				t.Error("delivery goroutine did not finish during cleanup")
			}
		}
	})
	select {
	case <-gate.entered:
	case res := <-done:
		attemptJoined = true
		t.Fatalf("delivery finished before the lock gate: fired=%v err=%v", res.fired, res.err)
	case <-ctx.Done():
		t.Fatal("delivery never reached the lock gate")
	}

	require.NoError(t, petdomain.NewRepository(db).RecordDeath(ctx, clinicID, source.ID, time.Now(), "test death"))
	releaseGate()
	select {
	case res := <-done:
		attemptJoined = true
		require.NoError(t, res.err)
		assert.False(t, res.fired)
	case <-ctx.Done():
		t.Fatal("delivery did not finish after the lock gate was released")
	}

	assert.Equal(t, 0, recorder.totalAttempts())
	log := deliveryTriggerLogFor(t, db, clinicID, owner.ID, model.TriggerTypeBirthdayMessage)
	require.Equal(t, model.TriggerStatusExcluded, log.Status)
	require.NotNil(t, log.ExcludedReason)
	assert.Equal(t, "stale_birthday_source", *log.ExcludedReason)
	assert.NotEqual(t, "all_pets_deceased", *log.ExcludedReason)
	requirePetAlive(t, ctx, db, clinicID, sibling.ID)
}

func TestDeliveryPetSerialization_RealDB_PetFirstFailsPromptlyAndRetries(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	asOf := deliveryPetAsOf
	owner, _, source := makeOwnerWithBirthdayPet(t, db, clinicID, "pet-first-owner", "U_petfirst")
	recorder := newDeliverySendRecorder()
	recorder.entered = make(chan struct{})
	svc := newBirthdayDeliveryService(db, repo, recorder.client())

	deathTx := db.Begin()
	require.NoError(t, deathTx.Error)
	t.Cleanup(func() { _ = deathTx.Rollback().Error })
	require.NoError(t, petdomain.NewRepository(db).RecordDeath(
		persistence.WithTxValue(context.Background(), deathTx),
		clinicID, source.ID, time.Now(), "test death"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	start := time.Now()
	fired, err := svc.processSingleOwner(ctx, recorder.client(), clinicID, owner.ID,
		model.TriggerTypeBirthdayMessage, model.TriggerTypeBirthdayMessage, asOf, nil,
		svc.petBirthdayEligibility(clinicID, int(asOf.Month()), asOf.Day()))
	elapsed := time.Since(start)
	require.Error(t, err)
	assert.False(t, fired)
	assert.Less(t, elapsed, 3*time.Second)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		assert.Equal(t, "55P03", pgErr.Code)
	} else {
		assert.Contains(t, err.Error(), "lock")
	}
	select {
	case <-recorder.entered:
		t.Fatal("the provider callback ran despite the NOWAIT lock failure")
	default:
	}
	assert.Equal(t, 0, recorder.totalAttempts())
	log := deliveryTriggerLogFor(t, db, clinicID, owner.ID, model.TriggerTypeBirthdayMessage)
	require.Equal(t, model.TriggerStatusFailed, log.Status)
	require.NotNil(t, log.ExcludedReason)
	assert.Equal(t, preSendRevalidationFailure, *log.ExcludedReason)
	assert.Nil(t, log.FiredAt)

	require.NoError(t, deathTx.Rollback().Error)
	retryCtx, retryCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer retryCancel()
	fired, err = svc.processSingleOwner(retryCtx, recorder.client(), clinicID, owner.ID,
		model.TriggerTypeBirthdayMessage, model.TriggerTypeBirthdayMessage, asOf, nil,
		svc.petBirthdayEligibility(clinicID, int(asOf.Month()), asOf.Day()))
	require.NoError(t, err)
	assert.True(t, fired)
	assert.Equal(t, 1, recorder.attemptsFor("U_petfirst"))
	log = deliveryTriggerLogFor(t, db, clinicID, owner.ID, model.TriggerTypeBirthdayMessage)
	assert.Equal(t, model.TriggerStatusFired, log.Status)
	assert.NotNil(t, log.FiredAt)

	fired, err = svc.processSingleOwner(retryCtx, recorder.client(), clinicID, owner.ID,
		model.TriggerTypeBirthdayMessage, model.TriggerTypeBirthdayMessage, asOf, nil,
		svc.petBirthdayEligibility(clinicID, int(asOf.Month()), asOf.Day()))
	require.NoError(t, err)
	assert.False(t, fired)
	assert.Equal(t, 1, recorder.attemptsFor("U_petfirst"))
}

func TestDeliveryPetSerialization_RealDB_ScopedWritersProceed(t *testing.T) {
	db := setupDeliveryOwnerLockDB(t)
	const clinicID = uint64(1)
	repo := &lstepTagCacheRepository{db: db}
	enableDeliverySync(t, db, clinicID)
	asOf := deliveryPetAsOf
	owner, sibling, source := makeOwnerWithBirthdayPet(t, db, clinicID, "scope-target", "U_scopetarget")

	otherOwner := makeCleanupOwner(t, db, clinicID, "scope-other-owner", "", false)
	otherOwnerPet := testdb.MakeSpeciesAndPet(t, db, clinicID, otherOwner.ID, "other-owner-pet")
	otherClinicOwner := makeCleanupOwner(t, db, 2, "scope-other-clinic-owner", "", false)
	otherClinicPet := testdb.MakeSpeciesAndPet(t, db, 2, otherClinicOwner.ID, "other-clinic-pet")

	targets := []struct {
		name     string
		clinicID uint64
		petID    uint64
	}{
		{"other owner same clinic", clinicID, otherOwnerPet.ID},
		{"other clinic", 2, otherClinicPet.ID},
	}

	recorder := newDeliverySendRecorder()
	recorder.entered = make(chan struct{})
	recorder.release = make(chan struct{})
	svc := newBirthdayDeliveryService(db, repo, recorder.client())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(recorder.release) }) }

	done := runBirthdayDeliveryAttempt(svc, ctx, recorder.client(), clinicID, owner.ID, asOf)
	deathResults := make([]chan error, 0, len(targets))
	attemptJoined := false
	t.Cleanup(func() {
		release()
		if !attemptJoined {
			select {
			case <-done:
			case <-time.After(15 * time.Second):
				t.Error("delivery goroutine did not finish during cleanup")
			}
		}
		for _, ch := range deathResults {
			if ch == nil {
				continue
			}
			select {
			case <-ch:
			case <-time.After(15 * time.Second):
				t.Error("a death goroutine did not finish during cleanup")
			}
		}
	})
	select {
	case <-recorder.entered:
	case res := <-done:
		attemptJoined = true
		t.Fatalf("provider call never started: fired=%v err=%v", res.fired, res.err)
	case <-ctx.Done():
		t.Fatal("provider call never reached before deadline")
	}

	for _, tgt := range targets {
		ch := make(chan error, 1)
		deathResults = append(deathResults, ch)
		go func(clinicID, petID uint64) {
			ch <- petdomain.NewRepository(db).RecordDeath(ctx, clinicID, petID, time.Now(), "test death")
		}(tgt.clinicID, tgt.petID)
	}
	for i, ch := range deathResults {
		select {
		case err := <-ch:
			assert.NoError(t, err, targets[i].name)
		case <-ctx.Done():
			t.Fatalf("%s RecordDeath blocked by the target owner's delivery lock", targets[i].name)
		}
		deathResults[i] = nil
	}

	release()
	select {
	case res := <-done:
		attemptJoined = true
		require.NoError(t, res.err)
		assert.True(t, res.fired)
	case <-ctx.Done():
		t.Fatal("delivery did not finish after the provider was released")
	}
	assert.Equal(t, 1, recorder.attemptsFor("U_scopetarget"))

	requirePetAlive(t, ctx, db, clinicID, sibling.ID)
	requirePetAlive(t, ctx, db, clinicID, source.ID)
	for _, tgt := range targets {
		requirePetDeceased(t, ctx, db, tgt.clinicID, tgt.petID)
	}
}
