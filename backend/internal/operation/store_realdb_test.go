//go:build operationrealdb

package operation

// store_realdb_test.go — real-DB proofs for gormStore, enrolled on lintscan
// dbOrTxParticipatingMethods: claim and attempt outcomes refuse the ambient
// owner transaction, the binding lock and the success receipt require it, the
// receipt rolls back with the owner transaction, NOWAIT reports a held binding,
// and every read is clinic scoped.
//
// Explicit opt-in (needs the worktree DB stack and a compiling internal/testdb):
//   go test -tags operationrealdb -count=1 ./internal/operation/...
// The tag keeps the DB-free C06 fixtures runnable while internal/testdb does not
// compile against the PR #729 models (pre-existing on main 5c5219965). This file
// has therefore never been type-checked or run; treat it as unverified.

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupOperationStoreDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.OperationBinding{},
		&model.OperationResult{},
		&model.OperationAttemptOutcome{},
	))
	// The AutoMigrate schema double has no composite unique key; mirror the
	// 001_init.sql binding scope key so ON CONFLICT DO NOTHING behaves as in
	// the migrated database.
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS operation_bindings_scope_key_test
		ON operation_bindings (clinic_id, owner_kind, actor_kind, actor_ref, intent, operation_key)`).Error)
	require.NoError(t, db.Exec(`TRUNCATE TABLE operation_attempt_outcomes, operation_results, operation_bindings RESTART IDENTITY`).Error)
	testdb.SeedClinicsForFK(t, db, 1, 2)
	return db
}

func storeTestSum(b byte) []byte {
	sum := make([]byte, fingerprintSize)
	for i := range sum {
		sum[i] = b
	}
	return sum
}

func TestGormStore_ClaimKeepsFirstContentAndRefusesAmbientTx(t *testing.T) {
	db := setupOperationStoreDB(t)
	s := newGormStore(db)
	scope := testScope()
	scope.ClinicID = 1
	key := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")

	first, err := s.ClaimBinding(context.Background(), scope, key, storeTestSum(1), "k1", nil)
	require.NoError(t, err)
	second, err := s.ClaimBinding(context.Background(), scope, key, storeTestSum(2), "k1", nil)
	require.NoError(t, err)

	require.Equal(t, first.Binding.ID, second.Binding.ID)
	require.Equal(t, storeTestSum(1), second.Fingerprint)
	var n int64
	require.NoError(t, db.Model(&model.OperationBinding{}).Count(&n).Error)
	require.Equal(t, int64(1), n)

	err = persistence.NewTransactor(db).WithTx(context.Background(), func(txCtx context.Context) error {
		_, claimErr := s.ClaimBinding(txCtx, scope, uuid.New(), storeTestSum(3), "k1", nil)
		return claimErr
	})
	require.Error(t, err)
}

func TestGormStore_ReceiptRollsBackWithOwnerTransaction(t *testing.T) {
	db := setupOperationStoreDB(t)
	s := newGormStore(db)
	scope := testScope()
	scope.ClinicID = 1
	claimed, err := s.ClaimBinding(context.Background(), scope, uuid.New(), storeTestSum(1), "k1", nil)
	require.NoError(t, err)
	b := claimed.Binding

	require.Error(t, s.InsertResult(context.Background(), b, Result{}), "receipt outside the owner tx must fail closed")
	require.Error(t, s.LockBindingNoWait(context.Background(), b), "lock outside the owner tx must fail closed")

	forced := errors.New("forced owner rollback")
	err = persistence.NewTransactor(db).WithTx(context.Background(), func(txCtx context.Context) error {
		require.NoError(t, s.LockBindingNoWait(txCtx, b))
		require.NoError(t, s.InsertResult(txCtx, b, Result{ResultRefs: map[string]string{"billing_id": "42"}}))
		inTx, loadErr := s.LoadResult(txCtx, b)
		require.NoError(t, loadErr)
		require.NotNil(t, inTx, "LoadResult must observe the uncommitted receipt via the ambient tx")
		return forced
	})
	require.ErrorIs(t, err, forced)

	after, err := s.LoadResult(context.Background(), b)
	require.NoError(t, err)
	require.Nil(t, after, "the receipt must roll back with the owner transaction")
}

func TestGormStore_LockNoWaitReportsHeldBinding(t *testing.T) {
	db := setupOperationStoreDB(t)
	s := newGormStore(db)
	scope := testScope()
	scope.ClinicID = 1
	claimed, err := s.ClaimBinding(context.Background(), scope, uuid.New(), storeTestSum(1), "k1", nil)
	require.NoError(t, err)
	b := claimed.Binding

	holder := db.Begin()
	require.NoError(t, holder.Error)
	t.Cleanup(func() { holder.Rollback() })
	require.NoError(t, s.LockBindingNoWait(persistence.WithTxValue(context.Background(), holder), b))

	err = persistence.NewTransactor(db).WithTx(context.Background(), func(txCtx context.Context) error {
		return s.LockBindingNoWait(txCtx, b)
	})
	require.ErrorIs(t, err, errBindingBusy)
}

func TestGormStore_OutcomesAreSeparateAndClinicScoped(t *testing.T) {
	db := setupOperationStoreDB(t)
	s := newGormStore(db)
	scope := testScope()
	scope.ClinicID = 1
	key := uuid.New()
	claimed, err := s.ClaimBinding(context.Background(), scope, key, storeTestSum(1), "k1", nil)
	require.NoError(t, err)
	b := claimed.Binding

	err = persistence.NewTransactor(db).WithTx(context.Background(), func(txCtx context.Context) error {
		return s.RecordOutcome(txCtx, b, uuid.New(), OutcomeRejected, StatusVersionConflict)
	})
	require.Error(t, err, "attempt outcomes must not join the rolled-back owner transaction")

	require.NoError(t, s.RecordOutcome(context.Background(), b, uuid.New(), OutcomeRejected, StatusVersionConflict))
	unknown, err := s.HasUnknownOutcome(context.Background(), b)
	require.NoError(t, err)
	require.False(t, unknown)
	require.NoError(t, s.RecordOutcome(context.Background(), b, uuid.New(), OutcomeUnknown, StatusResultUnavailable))
	require.NoError(t, s.RecordOutcome(context.Background(), b, uuid.New(), OutcomeRetryableFailure, ""))
	unknown, err = s.HasUnknownOutcome(context.Background(), b)
	require.NoError(t, err)
	require.True(t, unknown, "a later failure row must not release a recorded unknown")

	var stored model.OperationAttemptOutcome
	require.NoError(t, db.Where("binding_id = ?", b.ID).Order("id").First(&stored).Error)
	require.JSONEq(t, `{"status":"version-conflict"}`, string(stored.Details))

	otherClinic := scope
	otherClinic.ClinicID = 2
	_, found, err := s.FindBinding(context.Background(), otherClinic, key)
	require.NoError(t, err)
	require.False(t, found, "a binding of another clinic must not be visible")
	foreign := Binding{ID: b.ID, ClinicID: 2, Key: key}
	unknown, err = s.HasUnknownOutcome(context.Background(), foreign)
	require.NoError(t, err)
	require.False(t, unknown, "attempt outcomes must be read with the binding clinic")
}

// TestGormStore_ClaimMetadataWriterInClaimTx: the writeMetadata callback runs
// inside the claim transaction — its row commits with the winning binding, and
// a callback error rolls back both the binding row and the callback row.
func TestGormStore_ClaimMetadataWriterInClaimTx(t *testing.T) {
	db := setupOperationStoreDB(t)
	require.NoError(t, db.Exec(
		`CREATE TABLE IF NOT EXISTS claim_metadata_probe (id bigserial primary key, binding_id bigint)`).Error)
	s := newGormStore(db)
	scope := testScope()
	scope.ClinicID = 1

	claimed, err := s.ClaimBinding(context.Background(), scope, uuid.New(), storeTestSum(1), "k1",
		func(txCtx context.Context, b Binding) error {
			return persistence.DBOrTx(txCtx, db).
				Exec(`INSERT INTO claim_metadata_probe (binding_id) VALUES (?)`, b.ID).Error
		})
	require.NoError(t, err)
	var probes int64
	require.NoError(t, db.Table("claim_metadata_probe").Where("binding_id = ?", claimed.Binding.ID).Count(&probes).Error)
	require.Equal(t, int64(1), probes, "winner callback row must commit with the binding")

	failKey := uuid.New()
	forced := errors.New("forced metadata failure")
	_, err = s.ClaimBinding(context.Background(), scope, failKey, storeTestSum(2), "k1",
		func(txCtx context.Context, b Binding) error {
			if e := persistence.DBOrTx(txCtx, db).
				Exec(`INSERT INTO claim_metadata_probe (binding_id) VALUES (?)`, b.ID).Error; e != nil {
				return e
			}
			return forced
		})
	require.ErrorIs(t, err, forced)
	var bindings int64
	require.NoError(t, db.Model(&model.OperationBinding{}).Where("operation_key = ?", failKey).Count(&bindings).Error)
	require.Equal(t, int64(0), bindings, "callback failure must roll back the binding row")
	var totalProbes int64
	require.NoError(t, db.Table("claim_metadata_probe").Count(&totalProbes).Error)
	require.Equal(t, int64(1), totalProbes, "callback failure must roll back the metadata row")
}

// TestGormStore_ClaimConcurrentSingleMetadataCall: two concurrent ClaimBinding
// calls for the same scope/key — exactly one invokes writeMetadata, exactly one
// binding row exists, and the loser receives the winner's binding.
func TestGormStore_ClaimConcurrentSingleMetadataCall(t *testing.T) {
	db := setupOperationStoreDB(t)
	s := newGormStore(db)
	scope := testScope()
	scope.ClinicID = 1
	key := uuid.New()

	var calls int64
	var wg sync.WaitGroup
	ids := make(chan uint64, 2)
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			claimed, err := s.ClaimBinding(context.Background(), scope, key, storeTestSum(1), "k1",
				func(context.Context, Binding) error {
					atomic.AddInt64(&calls, 1)
					return nil
				})
			if err != nil {
				errs <- err
				return
			}
			ids <- claimed.Binding.ID
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
	got := map[uint64]bool{}
	for id := range ids {
		got[id] = true
	}
	require.Len(t, got, 1, "loser must return the winner's binding")
	require.Equal(t, int64(1), atomic.LoadInt64(&calls), "writeMetadata runs once for the claim winner only")
	var n int64
	require.NoError(t, db.Model(&model.OperationBinding{}).Count(&n).Error)
	require.Equal(t, int64(1), n)
}
