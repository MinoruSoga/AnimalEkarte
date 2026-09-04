package reservation

// reservation_type_delete_tx_atomicity_test.go — unused reservation type / group
// Delete joins ambient tx via DBOrTx.

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errReservationTypeDeleteAmbientTxSentinel = errors.New("simulated post-delete failure in ambient tx")

func TestReservationTypeRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupReservationTypeRepoTestDB(t)
	repo := NewReservationTypeRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rt := makeReservationTypeLinked(t, db, clinicA, "原子削除ロールバック対象区分", nil, nil)
	tx := testNewTransactor(db)
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Delete(txCtx, clinicA, rt.ID); err != nil {
			return err
		}
		return errReservationTypeDeleteAmbientTxSentinel
	})
	require.ErrorIs(t, txErr, errReservationTypeDeleteAmbientTxSentinel)

	got, err := repo.FindByID(ctx, clinicA, rt.ID)
	require.NoError(t, err, "Delete must join ambient tx so rollback restores the row")
	assert.Equal(t, rt.ID, got.ID)
}

func TestReservationTypeGroupRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupGroupRepoTestDB(t)
	repo := NewReservationTypeGroupRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	g := makeReservationTypeGroup(t, db, clinicA, "原子削除ロールバック対象グループ")
	tx := testNewTransactor(db)
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Delete(txCtx, clinicA, g.ID); err != nil {
			return err
		}
		return errReservationTypeDeleteAmbientTxSentinel
	})
	require.ErrorIs(t, txErr, errReservationTypeDeleteAmbientTxSentinel)

	got, err := repo.FindByID(ctx, clinicA, g.ID)
	require.NoError(t, err, "Delete must join ambient tx so rollback restores the row")
	assert.Equal(t, g.ID, got.ID)
}
