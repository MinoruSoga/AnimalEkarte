package billing

// insurance_repository_tx_atomicity_test.go — insuranceRepository FindByID / Update /
// Delete / CountUsageByInsuranceID が ambient tx に参加することを実 DB で証明する。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

var errInsuranceAmbientTxSentinel = errors.New("simulated post-write failure in ambient tx")

func TestInsuranceRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupInsuranceRepositoryTestDB(t)
	repo := NewInsuranceRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ins := testdb.MakeInsurance(t, db, clinicA, "原子削除ロールバック対象保険")
	tx := testNewTransactor(db)
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Delete(txCtx, clinicA, ins.ID); err != nil {
			return err
		}
		return errInsuranceAmbientTxSentinel
	})
	require.ErrorIs(t, txErr, errInsuranceAmbientTxSentinel)

	got, err := repo.FindByID(ctx, clinicA, ins.ID)
	require.NoError(t, err, "Delete must join ambient tx so rollback restores the row")
	assert.Equal(t, ins.ID, got.ID)
}

func TestInsuranceRepository_UpdateAndFindByID_JoinAmbientTx(t *testing.T) {
	db := setupInsuranceRepositoryTestDB(t)
	repo := NewInsuranceRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ins := testdb.MakeInsurance(t, db, clinicA, "更新前保険名")
	tx := testNewTransactor(db)
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		updated, err := repo.Update(txCtx, clinicA, ins.ID, map[string]any{"name": "未コミット保険名"})
		if err != nil {
			return err
		}
		if updated.Name != "未コミット保険名" {
			return errors.New("Update reload did not see own write")
		}
		found, err := repo.FindByID(txCtx, clinicA, ins.ID)
		if err != nil {
			return err
		}
		if found.Name != "未コミット保険名" {
			return errors.New("FindByID did not see uncommitted Update")
		}
		return errInsuranceAmbientTxSentinel
	})
	require.ErrorIs(t, txErr, errInsuranceAmbientTxSentinel)

	got, err := repo.FindByID(ctx, clinicA, ins.ID)
	require.NoError(t, err)
	assert.Equal(t, "更新前保険名", got.Name)
}

func TestInsuranceRepository_CountUsage_SeesUncommittedPetThenRollsBack(t *testing.T) {
	db := setupInsuranceRepositoryTestDB(t)
	repo := NewInsuranceRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ins := testdb.MakeInsurance(t, db, clinicA, "使用数原子性保険")
	owner := testdb.MakeTestOwner(t, db, clinicA, "使用数原子性飼主")

	tx := testNewTransactor(db)
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		conn := persistence.DBOrTx(txCtx, db)
		testdb.MakePetWithInsurance(t, conn, clinicA, owner.ID, &ins.ID, "未コミット保険ペット")
		count, err := repo.CountUsageByInsuranceID(txCtx, clinicA, ins.ID)
		if err != nil {
			return err
		}
		if count != 1 {
			return errors.New("CountUsageByInsuranceID did not see uncommitted pet")
		}
		return errInsuranceAmbientTxSentinel
	})
	require.ErrorIs(t, txErr, errInsuranceAmbientTxSentinel)

	count, err := repo.CountUsageByInsuranceID(ctx, clinicA, ins.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
