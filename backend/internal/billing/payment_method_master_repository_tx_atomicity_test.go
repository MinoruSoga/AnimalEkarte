package billing

// payment_method_master_repository_tx_atomicity_test.go — paymentMethodMasterRepository
// FindByID / Update / Delete / CountUsageByPaymentMethodID の ambient-tx 参加証明。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/persistence"
)

var errPaymentMethodAmbientTxSentinel = errors.New("simulated post-write failure in ambient tx")

func TestPaymentMethodMasterRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupPaymentMethodMasterRepoTestDB(t)
	repo := NewPaymentMethodMasterRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	m := makeCustomPaymentMethodMaster(t, db, clinicA, "原子削除ロールバック支払方法")
	tx := testNewTransactor(db)
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Delete(txCtx, clinicA, m.ID); err != nil {
			return err
		}
		return errPaymentMethodAmbientTxSentinel
	})
	require.ErrorIs(t, txErr, errPaymentMethodAmbientTxSentinel)

	got, err := repo.FindByID(ctx, clinicA, m.ID)
	require.NoError(t, err, "Delete must join ambient tx so rollback restores the row")
	assert.Equal(t, m.ID, got.ID)
}

func TestPaymentMethodMasterRepository_UpdateAndFindByID_JoinAmbientTx(t *testing.T) {
	db := setupPaymentMethodMasterRepoTestDB(t)
	repo := NewPaymentMethodMasterRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	m := makeCustomPaymentMethodMaster(t, db, clinicA, "更新前支払方法名")
	tx := testNewTransactor(db)
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		updated, err := repo.Update(txCtx, clinicA, m.ID, map[string]any{"name": "未コミット支払方法名"})
		if err != nil {
			return err
		}
		if updated.Name != "未コミット支払方法名" {
			return errors.New("Update reload did not see own write")
		}
		found, err := repo.FindByID(txCtx, clinicA, m.ID)
		if err != nil {
			return err
		}
		if found.Name != "未コミット支払方法名" {
			return errors.New("FindByID did not see uncommitted Update")
		}
		return errPaymentMethodAmbientTxSentinel
	})
	require.ErrorIs(t, txErr, errPaymentMethodAmbientTxSentinel)

	got, err := repo.FindByID(ctx, clinicA, m.ID)
	require.NoError(t, err)
	assert.Equal(t, "更新前支払方法名", got.Name)
}

func TestPaymentMethodMasterRepository_CountUsage_SeesUncommittedPaymentThenRollsBack(t *testing.T) {
	db := setupPaymentMethodMasterRepoTestDB(t)
	repo := NewPaymentMethodMasterRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	m := makePaymentMethodMaster(t, db, clinicA, "使用数原子性支払方法")
	tx := testNewTransactor(db)
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		conn := persistence.DBOrTx(txCtx, db)
		billing := makePaymentMethodBilling(t, conn, clinicA)
		makePaymentForBilling(t, conn, billing.ID, m.ID)
		count, err := repo.CountUsageByPaymentMethodID(txCtx, clinicA, m.ID)
		if err != nil {
			return err
		}
		if count != 1 {
			return errors.New("CountUsageByPaymentMethodID did not see uncommitted payment")
		}
		return errPaymentMethodAmbientTxSentinel
	})
	require.ErrorIs(t, txErr, errPaymentMethodAmbientTxSentinel)

	count, err := repo.CountUsageByPaymentMethodID(ctx, clinicA, m.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
