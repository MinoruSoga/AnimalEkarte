package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This needs an explicitly approved isolated PostgreSQL test database. -short skips it.
func TestMedicalRecordRepository_CountByPetIDUsesAmbientTransaction(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	owner := makeTestOwner(t, db, 1, "transaction内カルテcount飼主")
	pet := makeSpeciesAndPet(t, db, 1, owner.ID, "transaction内カルテcountペット")
	tx := db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { _ = tx.Rollback().Error })
	makeMedicalRecordForPet(t, tx, 1, pet.ID, time.Now(), false)
	count, err := repo.CountByPetID(persistence.WithTxValue(context.Background(), tx), 1, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "dependency checks must see transaction-local records")
	otherClinicCount, err := repo.CountByPetID(persistence.WithTxValue(context.Background(), tx), 2, pet.ID)
	require.NoError(t, err)
	assert.Zero(t, otherClinicCount)
	require.NoError(t, tx.Rollback().Error)
	count, err = repo.CountByPetID(context.Background(), 1, pet.ID)
	require.NoError(t, err)
	assert.Zero(t, count)
}
