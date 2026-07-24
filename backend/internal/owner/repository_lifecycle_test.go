package owner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestOwnerRepository_RecordAndClearLstepOptOut(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	record := makeTestOwner(t, db, clinicID, "LSTEP配信停止飼主")
	optedOutAt := time.Date(2026, 7, 24, 9, 30, 0, 0, time.UTC)
	const reason = "owner request"

	require.NoError(t, repo.RecordLstepOptOut(ctx, clinicID, record.ID, optedOutAt, reason))
	var recorded model.Owner
	require.NoError(t, db.WithContext(ctx).First(&recorded, record.ID).Error)
	assert.True(t, recorded.LstepOptOut)
	require.NotNil(t, recorded.LstepOptOutAt)
	assert.True(t, recorded.LstepOptOutAt.Equal(optedOutAt))
	require.NotNil(t, recorded.LstepOptOutReason)
	assert.Equal(t, reason, *recorded.LstepOptOutReason)

	require.NoError(t, repo.ClearLstepOptOut(ctx, clinicID, record.ID))
	var cleared model.Owner
	require.NoError(t, db.WithContext(ctx).First(&cleared, record.ID).Error)
	assert.False(t, cleared.LstepOptOut)
	assert.Nil(t, cleared.LstepOptOutAt)
	assert.Nil(t, cleared.LstepOptOutReason)
}

func TestOwnerRepository_LstepOptOutLifecycle_IsClinicScoped(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	record := makeTestOwner(t, db, clinicA, "医院分離飼主")
	at := time.Date(2026, 7, 24, 9, 45, 0, 0, time.UTC)

	err := repo.RecordLstepOptOut(ctx, clinicB, record.ID, at, "foreign clinic")
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "unexpected error: %v", err)

	require.NoError(t, repo.RecordLstepOptOut(ctx, clinicA, record.ID, at, "local clinic"))
	err = repo.ClearLstepOptOut(ctx, clinicB, record.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "unexpected error: %v", err)

	var unchanged model.Owner
	require.NoError(t, db.WithContext(ctx).First(&unchanged, record.ID).Error)
	assert.True(t, unchanged.LstepOptOut)
	require.NotNil(t, unchanged.LstepOptOutReason)
	assert.Equal(t, "local clinic", *unchanged.LstepOptOutReason)
}

func TestOwnerRepository_LstepOptOutLifecycle_JoinsAmbientTransaction(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	sentinel := errors.New("simulated downstream failure")

	t.Run("record rolls back", func(t *testing.T) {
		record := makeTestOwner(t, db, clinicID, "配信停止ロールバック飼主")
		at := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)

		txErr := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := persistence.WithTxValue(ctx, tx)
			if err := repo.RecordLstepOptOut(txCtx, clinicID, record.ID, at, "rollback"); err != nil {
				return err
			}
			var withinTx model.Owner
			if err := tx.WithContext(txCtx).First(&withinTx, record.ID).Error; err != nil {
				return err
			}
			if !withinTx.LstepOptOut {
				return errors.New("record was not visible inside ambient transaction")
			}
			return sentinel
		})
		require.ErrorIs(t, txErr, sentinel)

		var after model.Owner
		require.NoError(t, db.WithContext(ctx).First(&after, record.ID).Error)
		assert.False(t, after.LstepOptOut)
		assert.Nil(t, after.LstepOptOutAt)
		assert.Nil(t, after.LstepOptOutReason)
	})

	t.Run("clear rolls back", func(t *testing.T) {
		record := makeTestOwner(t, db, clinicID, "配信停止解除ロールバック飼主")
		at := time.Date(2026, 7, 24, 10, 15, 0, 0, time.UTC)
		require.NoError(t, repo.RecordLstepOptOut(ctx, clinicID, record.ID, at, "keep"))

		txErr := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := persistence.WithTxValue(ctx, tx)
			if err := repo.ClearLstepOptOut(txCtx, clinicID, record.ID); err != nil {
				return err
			}
			var withinTx model.Owner
			if err := tx.WithContext(txCtx).First(&withinTx, record.ID).Error; err != nil {
				return err
			}
			if withinTx.LstepOptOut {
				return errors.New("clear was not visible inside ambient transaction")
			}
			return sentinel
		})
		require.ErrorIs(t, txErr, sentinel)

		var after model.Owner
		require.NoError(t, db.WithContext(ctx).First(&after, record.ID).Error)
		assert.True(t, after.LstepOptOut)
		require.NotNil(t, after.LstepOptOutAt)
		assert.True(t, after.LstepOptOutAt.Equal(at))
		require.NotNil(t, after.LstepOptOutReason)
		assert.Equal(t, "keep", *after.LstepOptOutReason)
	})
}
