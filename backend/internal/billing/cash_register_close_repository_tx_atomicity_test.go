package billing

// cash_register_close_repository_tx_atomicity_test.go — ambient-tx participation for
// cashRegisterCloseRepository Create/CreateAdjustment/Find*/HasCloseOnDate.
//
// W-013 close writes and post-close reads must join the caller's ambient transaction
// so accounting correction + close adjustment stay atomic.
//
// temp-revert RED: DBOrTx → r.db.WithContext(ctx)
//   - RollsBack* leaves durable rows
//   - SeesUncommitted* misses rows written earlier in the same WithTx

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

var errSentinelCashCloseTx = errors.New("simulated post-write failure in cash close ambient tx")

func TestCashRegisterCloseRepository_Create_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	date := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)

	closeRec := &model.CashRegisterClose{
		ClinicID:          clinicA,
		CloseDate:         date,
		Period:            "am",
		CategoryBreakdown: json.RawMessage(`{}`),
	}

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Create(txCtx, closeRec); err != nil {
			return err
		}
		return errSentinelCashCloseTx
	})
	require.ErrorIs(t, txErr, errSentinelCashCloseTx)

	var count int64
	require.NoError(t, db.Model(&model.CashRegisterClose{}).
		Where("clinic_id = ? AND close_date = ?", clinicA, date.Format(time.DateOnly)).
		Count(&count).Error)
	assert.Zero(t, count, "Create must roll back with ambient tx (DBOrTx participation)")
}

func TestCashRegisterCloseRepository_Create_CommitsWithinAmbientTx(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	date := time.Date(2026, 6, 21, 0, 0, 0, 0, time.UTC)

	closeRec := &model.CashRegisterClose{
		ClinicID:          clinicA,
		CloseDate:         date,
		Period:            "pm",
		CategoryBreakdown: json.RawMessage(`{}`),
	}

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		return repo.Create(txCtx, closeRec)
	}))
	assert.NotZero(t, closeRec.ID)

	got, err := repo.FindByID(ctx, clinicA, closeRec.ID)
	require.NoError(t, err)
	assert.Equal(t, "pm", got.Period)
}

func TestCashRegisterCloseRepository_CreateAdjustment_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	// Parent close is committed outside so FK is satisfied; adjustment is ambient-only.
	parent := makeCashRegisterClose(t, db, clinicA, time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC), "am", nil)
	actorID := uint64(11)
	adj := &model.CashRegisterCloseAdjustment{
		ClinicID:           clinicA,
		CloseID:            parent.ID,
		BillingID:          99,
		AccountingDelta:    -100,
		CashMovementAmount: 0,
		Reason:             "ambient rollback probe",
		ActorID:            &actorID,
		ExecutedAt:         time.Now(),
	}

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.CreateAdjustment(txCtx, adj); err != nil {
			return err
		}
		return errSentinelCashCloseTx
	})
	require.ErrorIs(t, txErr, errSentinelCashCloseTx)

	var count int64
	require.NoError(t, db.Model(&model.CashRegisterCloseAdjustment{}).
		Where("close_id = ? AND reason = ?", parent.ID, "ambient rollback probe").
		Count(&count).Error)
	assert.Zero(t, count, "CreateAdjustment must roll back with ambient tx")
}

func TestCashRegisterCloseRepository_CreateAdjustment_CommitsWithinAmbientTx(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	parent := makeCashRegisterClose(t, db, clinicA, time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC), "pm", nil)
	actorID := uint64(12)
	adj := &model.CashRegisterCloseAdjustment{
		ClinicID:           clinicA,
		CloseID:            parent.ID,
		BillingID:          100,
		AccountingDelta:    50,
		CashMovementAmount: 0,
		Reason:             "ambient commit probe",
		ActorID:            &actorID,
		ExecutedAt:         time.Now(),
	}

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		return repo.CreateAdjustment(txCtx, adj)
	}))
	assert.NotZero(t, adj.ID)

	var got model.CashRegisterCloseAdjustment
	require.NoError(t, db.First(&got, adj.ID).Error)
	assert.Equal(t, "ambient commit probe", got.Reason)
}

func TestCashRegisterCloseRepository_Void_RollsBackWithAmbientTx(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	closeRec := makeCashRegisterClose(
		t, db, clinicA, time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC), "am", nil,
	)

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Void(txCtx, clinicA, closeRec.ID); err != nil {
			return err
		}
		return errSentinelCashCloseTx
	})
	require.ErrorIs(t, txErr, errSentinelCashCloseTx)

	got, err := repo.FindByID(ctx, clinicA, closeRec.ID)
	require.NoError(t, err)
	assert.Equal(t, closeRec.ID, got.ID, "Void must roll back with the ambient transaction")

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		return repo.Void(txCtx, clinicA, closeRec.ID)
	}))
	_, err = repo.FindByID(ctx, clinicA, closeRec.ID)
	require.Error(t, err, "committed Void must hide the close")
}

// TestCashRegisterCloseRepository_Reads_SeeUncommittedCreateInAmbientTx proves
// FindAll / FindByID / HasCloseOnDate / FindByDateAndPeriod observe Create in the
// same ambient tx and miss the row outside until commit (or after rollback).
func TestCashRegisterCloseRepository_Reads_SeeUncommittedCreateInAmbientTx(t *testing.T) {
	db := setupCashRegisterCloseTestDB(t)
	repo := NewCashRegisterCloseRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	date := time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC)

	// Outside: no close yet.
	has, err := repo.HasCloseOnDate(ctx, clinicA, date)
	require.NoError(t, err)
	assert.False(t, has)

	forced := errors.New("forced cash-close read probe rollback")
	var createdID uint64

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if persistence.TxFromContext(txCtx) == nil {
			return errors.New("expected ambient tx installed")
		}
		closeRec := &model.CashRegisterClose{
			ClinicID:          clinicA,
			CloseDate:         date,
			Period:            "emg",
			CategoryBreakdown: json.RawMessage(`{}`),
		}
		if err := repo.Create(txCtx, closeRec); err != nil {
			return err
		}
		createdID = closeRec.ID

		// FindByID
		got, err := repo.FindByID(txCtx, clinicA, closeRec.ID)
		if err != nil {
			return err
		}
		if got.Period != "emg" {
			return errors.New("FindByID did not see uncommitted Create")
		}

		// FindByDateAndPeriod
		byDate, err := repo.FindByDateAndPeriod(txCtx, clinicA, date, "emg")
		if err != nil {
			return err
		}
		if byDate == nil || byDate.ID != closeRec.ID {
			return errors.New("FindByDateAndPeriod did not see uncommitted Create")
		}

		// HasCloseOnDate
		has, err := repo.HasCloseOnDate(txCtx, clinicA, date)
		if err != nil {
			return err
		}
		if !has {
			return errors.New("HasCloseOnDate did not see uncommitted Create")
		}

		// FindAll
		all, total, err := repo.FindAll(txCtx, clinicA, nil, nil, 1, 20)
		if err != nil {
			return err
		}
		if total < 1 {
			return errors.New("FindAll total did not count uncommitted Create")
		}
		found := false
		for _, c := range all {
			if c.ID == closeRec.ID {
				found = true
				break
			}
		}
		if !found {
			return errors.New("FindAll did not list uncommitted Create")
		}

		// Outside connection must NOT see the uncommitted row yet.
		outsideHas, err := repo.HasCloseOnDate(ctx, clinicA, date)
		if err != nil {
			return err
		}
		if outsideHas {
			return errors.New("outside HasCloseOnDate saw uncommitted Create (isolation broken)")
		}

		return forced
	})
	require.ErrorIs(t, txErr, forced)

	// After rollback: still invisible.
	has, err = repo.HasCloseOnDate(ctx, clinicA, date)
	require.NoError(t, err)
	assert.False(t, has)

	_, err = repo.FindByID(ctx, clinicA, createdID)
	require.Error(t, err)

	byDate, err := repo.FindByDateAndPeriod(ctx, clinicA, date, "emg")
	require.NoError(t, err)
	assert.Nil(t, byDate)

	// Commit path: Create then outside reads succeed.
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		return repo.Create(txCtx, &model.CashRegisterClose{
			ClinicID:          clinicA,
			CloseDate:         date,
			Period:            "emg",
			CategoryBreakdown: json.RawMessage(`{}`),
		})
	}))
	has, err = repo.HasCloseOnDate(ctx, clinicA, date)
	require.NoError(t, err)
	assert.True(t, has, "after commit HasCloseOnDate must see the close")
}
