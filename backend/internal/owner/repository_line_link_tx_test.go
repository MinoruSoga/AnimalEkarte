package owner

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestOwnerRepository_LockLineLinkOwnerRequiresAmbientTransaction(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	owner := makeTestOwner(t, db, 1, "LINE row lock owner")

	got, err := repo.LockLineLinkOwner(context.Background(), 1, owner.ID)

	assert.Nil(t, got)
	require.Error(t, err)
}

func TestOwnerRepository_LineLinkLockAndUpdateParticipateInAmbientTransaction(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	owner := makeTestOwner(t, db, 1, "LINE rollback owner")
	rollbackErr := errors.New("force rollback")
	lineUserID := "U-rollback"

	err := persistence.NewTransactor(db).WithTx(context.Background(), func(ctx context.Context) error {
		locked, err := repo.LockLineLinkOwner(ctx, 1, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, owner.ID, locked.ID)
		require.NoError(t, repo.UpdateLineUserID(ctx, 1, owner.ID, &lineUserID))
		return rollbackErr
	})

	require.ErrorIs(t, err, rollbackErr)
	var persisted model.Owner
	require.NoError(t, db.First(&persisted, owner.ID).Error)
	assert.Nil(t, persisted.LineUserID)
}

func TestOwnerRepository_LockLineLinkOwnerIsClinicScoped(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	owner := makeTestOwner(t, db, 1, "LINE clinic scoped owner")

	err := persistence.NewTransactor(db).WithTx(context.Background(), func(ctx context.Context) error {
		got, err := repo.LockLineLinkOwner(ctx, 2, owner.ID)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		return nil
	})

	require.NoError(t, err)
}
