package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	petdomain "github.com/animal-ekarte/backend/internal/pet"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestNextVisitQueries_TransferredTentativeUsesCurrentOwner(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	rt := makeReservationType(t, db, clinicID)
	oldOwner := testdb.MakeTestOwner(t, db, clinicID, "reminder previous owner")
	newOwner := testdb.MakeTestOwner(t, db, clinicID, "reminder current owner")
	pet := makeNextVisitQueryPet(t, db, clinicID, oldOwner.ID, "transferred reminder pet", model.PetStatusAlive, nil)
	day := jstMidnight(time.Now().AddDate(0, 0, 3))
	intent := makeNextVisitQueryPetReservation(t, db, clinicID, rt.ID, &oldOwner.ID, &pet.ID,
		model.ReservationStatusTentative, day, &day)
	require.NoError(t, db.Model(pet).Update("owner_id", newOwner.ID).Error)

	owners, err := repo.FindOwnerIDsWithTentativeIntendedDate(ctx, clinicID, day)
	require.NoError(t, err)
	assert.Equal(t, []uint64{newOwner.ID}, owners)
	previous, err := repo.FindNextAppointmentForOwner(ctx, clinicID, oldOwner.ID)
	require.NoError(t, err)
	assert.Nil(t, previous)
	current, err := repo.FindNextAppointmentForOwner(ctx, clinicID, newOwner.ID)
	require.NoError(t, err)
	require.NotNil(t, current)
	assert.Equal(t, intent.ID, current.ID)
	for _, ownerID := range []uint64{oldOwner.ID, newOwner.ID} {
		rows, findErr := repo.FindTentativesByOwnerAndIntendedDate(ctx, clinicID, ownerID, day)
		require.NoError(t, findErr)
		single, findErr := repo.FindTentativeByOwnerAndIntendedDate(ctx, clinicID, ownerID, day)
		require.NoError(t, findErr)
		if ownerID == oldOwner.ID {
			assert.Empty(t, rows)
			assert.Nil(t, single)
		} else {
			require.Len(t, rows, 1)
			require.NotNil(t, single)
			assert.Equal(t, intent.ID, single.ID)
		}
	}
	foreign, err := repo.FindOwnerIDsWithTentativeIntendedDate(ctx, clinicID+1, day)
	require.NoError(t, err)
	assert.Empty(t, foreign)
	// Real bookings keep their stored ownership/history contract.
	require.NoError(t, db.Model(intent).Update("status", model.ReservationStatusConfirmed).Error)
	previous, err = repo.FindNextAppointmentForOwner(ctx, clinicID, oldOwner.ID)
	require.NoError(t, err)
	require.NotNil(t, previous)
	assert.Equal(t, intent.ID, previous.ID)
}

func TestTentativeReminderSerialization_HoldsPetStateThroughCallback(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	const clinicID = uint64(1)
	owner := testdb.MakeTestOwner(t, db, clinicID, "locked reminder owner")
	pet := makeNextVisitQueryPet(t, db, clinicID, owner.ID, "locked reminder pet", model.PetStatusAlive, nil)
	deceasedAt := time.Now()
	err := repo.WithTentativeReminderSerialization(context.Background(), clinicID, []uint64{pet.ID, pet.ID}, func(_ context.Context) error {
		deathCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		deathErr := petdomain.NewRepository(db).RecordDeath(deathCtx, clinicID, pet.ID, deceasedAt, "fixture")
		assert.Error(t, deathErr, "death cannot commit while the delivery callback holds pet SHARE")
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, petdomain.NewRepository(db).RecordDeath(context.Background(), clinicID, pet.ID, deceasedAt, "fixture"))
}

func TestTentativeReminderSerialization_ConflictingDeathFailsBeforeCallback(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	const clinicID = uint64(1)
	owner := testdb.MakeTestOwner(t, db, clinicID, "dying reminder owner")
	pet := makeNextVisitQueryPet(t, db, clinicID, owner.ID, "dying reminder pet", model.PetStatusAlive, nil)
	deathTx := db.Begin()
	require.NoError(t, deathTx.Error)
	t.Cleanup(func() { deathTx.Rollback() })
	require.NoError(t, deathTx.Model(pet).Updates(map[string]any{"status": model.PetStatusDeceased, "deceased_at": time.Now()}).Error)
	called := false
	err := repo.WithTentativeReminderSerialization(context.Background(), clinicID, []uint64{pet.ID}, func(_ context.Context) error {
		called = true
		return nil
	})
	require.Error(t, err)
	assert.False(t, called, "NOWAIT must fail closed before delivery when lifecycle update owns the row")
	require.NoError(t, deathTx.Commit().Error)
}
