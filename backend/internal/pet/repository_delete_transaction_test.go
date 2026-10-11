package pet

import (
	"context"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This needs an explicitly approved isolated PostgreSQL test database. -short skips it.
func TestPetRepository_DeleteParticipatesInAmbientTransaction(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewRepository(db)
	owner := makeTestOwner(t, db, 1, "削除transaction飼主")
	pet := makeSpeciesAndPet(t, db, 1, owner.ID, "削除transactionペット")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tx, _ := beginAmbientPetTransaction(t, db)
	txCtx := persistence.WithTxValue(ctx, tx)
	t.Cleanup(func() { _ = tx.Rollback().Error })
	locked, err := repo.LockByIDForUpdate(txCtx, 1, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, owner.ID, locked.OwnerID)
	require.NoError(t, repo.Delete(txCtx, 1, pet.ID))
	// Query the transaction directly: FindByID is a request read, not an ambient read.
	var active int64
	require.NoError(t, tx.Model(pet).Where("id = ? AND clinic_id = ?", pet.ID, 1).Count(&active).Error)
	assert.Zero(t, active)
	require.NoError(t, tx.Rollback().Error)
	_, err = repo.FindByID(context.Background(), 1, pet.ID)
	require.NoError(t, err, "rollback must restore the active pet")
	require.NoError(t, repo.Delete(context.Background(), 1, pet.ID))
	_, err = repo.FindByID(context.Background(), 1, pet.ID)
	assert.True(t, apperrors.IsNotFound(err))
}
