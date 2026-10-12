package medicalrecord

// cage_repository_tx_atomicity_test.go — EMR-306
//
// ケージ一時連結で追加された repository メソッド（SetMergedInto/ClearMergedInto/
// LockCagesForUpdate/LockCageGroupForUpdate/CountMergedMembers/CountOccupiedCages/
// ListOccupiedCageIDs）の DBOrTx 参加を実証する。writer は ambient WithTx の
// sentinel rollback で永続化されず、reader は同一 tx の未コミット write を見て、
// rollback 後は元の状態に戻る。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"

	"gorm.io/gorm"
)

var errCageRepoTxSentinel = errors.New("simulated post-merge failure in ambient tx")

func newCageMergeRepository(t *testing.T, db *gorm.DB) CageMergeRepository {
	t.Helper()
	repo, ok := NewCageRepository(db).(CageMergeRepository)
	require.True(t, ok, "cage repository must implement CageMergeRepository")
	return repo
}

func TestCageRepository_SetMergedInto_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := newCageMergeRepository(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	rep := makeCageMaster(t, db, clinicID, "代表-tx")
	m1 := makeCageMaster(t, db, clinicID, "構成1-tx")
	m2 := makeCageMaster(t, db, clinicID, "構成2-tx")

	txErr := (testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
		require.NoError(t, repo.SetMergedInto(txCtx, clinicID, rep.ID, []uint64{m1.ID, m2.ID}))

		// 同一 tx 内では未コミットの連結が見える（reader が DBOrTx 参加している証拠）。
		counts, err := repo.CountMergedMembers(txCtx, clinicID, []uint64{rep.ID})
		require.NoError(t, err)
		assert.EqualValues(t, 2, counts[rep.ID])

		locked, err := repo.LockCageGroupForUpdate(txCtx, clinicID, rep.ID)
		require.NoError(t, err)
		assert.Len(t, locked, 3)

		return errCageRepoTxSentinel
	})
	require.ErrorIs(t, txErr, errCageRepoTxSentinel)

	assert.Nil(t, cageMergedInto(t, db, m1.ID))
	assert.Nil(t, cageMergedInto(t, db, m2.ID))
	counts, err := repo.CountMergedMembers(ctx, clinicID, []uint64{rep.ID})
	require.NoError(t, err)
	assert.EqualValues(t, 0, counts[rep.ID])
}

func TestCageRepository_ClearMergedInto_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := newCageMergeRepository(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	rep := makeCageMaster(t, db, clinicID, "代表-clear-tx")
	m1 := makeCageMaster(t, db, clinicID, "構成-clear-tx")
	require.NoError(t, repo.SetMergedInto(ctx, clinicID, rep.ID, []uint64{m1.ID}))
	require.NotNil(t, cageMergedInto(t, db, m1.ID))

	txErr := (testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
		require.NoError(t, repo.ClearMergedInto(txCtx, clinicID, rep.ID))

		var cage struct{ MergedIntoCageID *uint64 }
		require.NoError(t, persistence.DBOrTx(txCtx, db).
			Table("cages").Select("merged_into_cage_id").
			Where("id = ?", m1.ID).Take(&cage).Error)
		assert.Nil(t, cage.MergedIntoCageID, "同一 tx 内では解除済みに見える")

		return errCageRepoTxSentinel
	})
	require.ErrorIs(t, txErr, errCageRepoTxSentinel)

	assert.Equal(t, rep.ID, *cageMergedInto(t, db, m1.ID))
}

func TestCageRepository_LockCagesForUpdate_JoinsAmbientTx(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := newCageMergeRepository(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	c1 := makeCageMaster(t, db, clinicID, "lock-1")
	c2 := makeCageMaster(t, db, clinicID, "lock-2")

	txErr := (testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
		// ambient tx の DBOrTx 上で FOR UPDATE を発行できることを実証する
		// （非 tx 経路に fallback すると txCtx の tx に乗らない）。
		locked, err := repo.LockCagesForUpdate(txCtx, clinicID, []uint64{c1.ID, c2.ID})
		require.NoError(t, err)
		assert.Len(t, locked, 2)
		return errCageRepoTxSentinel
	})
	require.ErrorIs(t, txErr, errCageRepoTxSentinel)
}

func TestCageRepository_CountOccupiedCages_SeesUncommittedWritesThenRollback(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := newCageMergeRepository(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "飼主 tx occupancy")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "タロウ")
	cage := makeCageMaster(t, db, clinicID, "occupancy-tx")

	txErr := (testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
		// tx 内で入院を直接 INSERT し、reader が未コミット占有を見ることを実証する。
		now := time.Now().UTC().Truncate(24 * time.Hour)
		h := &model.Hospitalization{
			ClinicID:            clinicID,
			OwnerID:             owner.ID,
			PetID:               pet.ID,
			HospitalizationType: model.HospitalizationTypeInpatient,
			StartDate:           now,
			EndDate:             now.AddDate(0, 0, 3),
			Status:              model.HospitalizationStatusAdmitted,
			CageID:              &cage.ID,
		}
		require.NoError(t, persistence.DBOrTx(txCtx, db).Create(h).Error)

		counts, err := repo.CountOccupiedCages(txCtx, clinicID, []uint64{cage.ID})
		require.NoError(t, err)
		assert.EqualValues(t, 1, counts[cage.ID])

		return errCageRepoTxSentinel
	})
	require.ErrorIs(t, txErr, errCageRepoTxSentinel)

	counts, err := repo.CountOccupiedCages(ctx, clinicID, []uint64{cage.ID})
	require.NoError(t, err)
	assert.EqualValues(t, 0, counts[cage.ID])
}

func TestCageRepository_ListOccupiedCageIDs_SeesUncommittedWritesThenRollback(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := newCageMergeRepository(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "飼主 tx occupied-ids")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "ハナ")
	cage := makeCageMaster(t, db, clinicID, "occupied-ids-tx")

	txErr := (testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
		// tx 内で入院を直接 INSERT し、reader が未コミット占有を見ることを実証する。
		now := time.Now().UTC().Truncate(24 * time.Hour)
		h := &model.Hospitalization{
			ClinicID:            clinicID,
			OwnerID:             owner.ID,
			PetID:               pet.ID,
			HospitalizationType: model.HospitalizationTypeInpatient,
			StartDate:           now,
			EndDate:             now.AddDate(0, 0, 3),
			Status:              model.HospitalizationStatusAdmitted,
			CageID:              &cage.ID,
		}
		require.NoError(t, persistence.DBOrTx(txCtx, db).Create(h).Error)

		ids, err := repo.ListOccupiedCageIDs(txCtx, clinicID)
		require.NoError(t, err)
		assert.Contains(t, ids, cage.ID)

		return errCageRepoTxSentinel
	})
	require.ErrorIs(t, txErr, errCageRepoTxSentinel)

	ids, err := repo.ListOccupiedCageIDs(ctx, clinicID)
	require.NoError(t, err)
	assert.NotContains(t, ids, cage.ID)
}
