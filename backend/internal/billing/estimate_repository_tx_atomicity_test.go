package billing

// estimate_repository_tx_atomicity_test.go — ambient-tx participation for
// estimateRepository.AllocateNextEstimateNo (clinic-scoped EST-{N} numbering).
//
// AllocateNextEstimateNo takes pg_advisory_xact_lock on the ambient connection and
// scans EST-% (including soft-deleted) so concurrent Create in the same tx advances
// the sequence without leaking uncommitted numbers to other sessions.
//
// temp-revert RED: DBOrTx → r.db.WithContext(ctx)
//   - second allocate inside ambient after uncommitted insert may miss the row
//     (or use a different connection) and return a duplicate number
//   - rollback isolation still holds for committed rows, but concurrent session
//     may steal the number if the advisory lock is not on the ambient xact

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

var errSentinelEstimateNoTx = errors.New("simulated post-allocate failure in ambient tx")

// TestEstimateRepository_AllocateNextEstimateNo_SeesUncommittedInsertAndRollsBack
// allocates EST-1, inserts that number uncommitted, allocates EST-2 in the same
// ambient tx, then rolls back so an outside allocate restarts at EST-1.
func TestEstimateRepository_AllocateNextEstimateNo_SeesUncommittedInsertAndRollsBack(t *testing.T) {
	db := setupEstimateRepoTestDB(t)
	repo := NewEstimateRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicA, "estimate-no ambient owner")

	// Empty table → first number is EST-1.
	firstOutside, err := repo.AllocateNextEstimateNo(ctx, clinicA)
	require.NoError(t, err)
	assert.Equal(t, "EST-1", firstOutside)

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if persistence.TxFromContext(txCtx) == nil {
			return errors.New("expected ambient tx installed")
		}
		no1, err := repo.AllocateNextEstimateNo(txCtx, clinicA)
		if err != nil {
			return err
		}
		if no1 != "EST-1" {
			return errors.New("first ambient allocate expected EST-1, got " + no1)
		}
		// Occupy the number inside the same ambient tx (uncommitted).
		est := &model.Estimate{
			ClinicID:   clinicA,
			OwnerID:    &owner.ID,
			EstimateNo: no1,
			Status:     model.EstimateStatusDraft,
			Title:      "uncommitted number holder",
		}
		if err := repo.Create(txCtx, est); err != nil {
			return err
		}
		no2, err := repo.AllocateNextEstimateNo(txCtx, clinicA)
		if err != nil {
			return err
		}
		if no2 != "EST-2" {
			return errors.New("second ambient allocate must see uncommitted EST-1 and return EST-2, got " + no2)
		}
		return errSentinelEstimateNoTx
	})
	require.ErrorIs(t, txErr, errSentinelEstimateNoTx)

	// Rollback: uncommitted estimate gone; sequence restarts at EST-1.
	var count int64
	require.NoError(t, db.Model(&model.Estimate{}).
		Where("clinic_id = ?", clinicA).
		Count(&count).Error)
	assert.Zero(t, count, "uncommitted estimate must roll back")

	after, err := repo.AllocateNextEstimateNo(ctx, clinicA)
	require.NoError(t, err)
	assert.Equal(t, "EST-1", after,
		"rolled-back allocate must not permanently consume EST-1")
}

// TestEstimateRepository_AllocateNextEstimateNo_AdvisoryLockBlocksConcurrentSession
// holds the xact advisory lock after the first allocate so a concurrent session's
// AllocateNextEstimateNo times out (cannot steal an uncommitted sequence slot).
func TestEstimateRepository_AllocateNextEstimateNo_AdvisoryLockBlocksConcurrentSession(t *testing.T) {
	db := setupEstimateRepoTestDB(t)
	repo := NewEstimateRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicA, "estimate-no lock owner")

	err := tx.WithTx(ctx, func(txCtx context.Context) error {
		no1, allocErr := repo.AllocateNextEstimateNo(txCtx, clinicA)
		require.NoError(t, allocErr)
		assert.Equal(t, "EST-1", no1)

		// Occupy EST-1 so a competing session that somehow bypassed the lock
		// would still have evidence of the uncommitted number if it could allocate.
		require.NoError(t, repo.Create(txCtx, &model.Estimate{
			ClinicID:   clinicA,
			OwnerID:    &owner.ID,
			EstimateNo: no1,
			Status:     model.EstimateStatusDraft,
			Title:      "lock holder",
		}))

		competingTx := db.WithContext(ctx).Begin()
		require.NoError(t, competingTx.Error)
		defer competingTx.Rollback()
		require.NoError(t, competingTx.Exec("SET LOCAL lock_timeout = '200ms'").Error)

		competingCtx := persistence.WithTxValue(ctx, competingTx)
		_, competeErr := repo.AllocateNextEstimateNo(competingCtx, clinicA)
		require.Error(t, competeErr,
			"concurrent AllocateNextEstimateNo must block on pg_advisory_xact_lock held by ambient tx")

		// Outside base connection (no ambient) also cannot see the uncommitted EST-1
		// occupancy for durable state — after we return and commit or roll back.
		// Here we only assert lock contention while ambient is open.
		return nil
	})
	require.NoError(t, err)

	// After ambient commit of EST-1, next allocate is EST-2.
	// (WithTx committed because callback returned nil.)
	next, err := repo.AllocateNextEstimateNo(ctx, clinicA)
	require.NoError(t, err)
	assert.Equal(t, "EST-2", next)
}
