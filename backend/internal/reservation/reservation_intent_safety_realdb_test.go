package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// Short mode never provisions or connects to a database. These cases are for an
// explicitly authorized isolated PostgreSQL environment or CI's existing DB gate.
func TestReservationIntentSafety_TransferredReadRealDB(t *testing.T) {
	if testing.Short() {
		t.Skip("requires an isolated PostgreSQL database")
	}
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db).(*reservationRepository)
	ctx := context.Background()
	const clinicID = uint64(1)
	formerOwner, pet, _ := makeNextVisitIntentFixture(t, db, clinicID, "transferred-read")
	intent, err := setNextVisitIntentInTx(ctx, db, repo, clinicID, pet.ID, SetNextVisitIntentInput{IntendedDate: futureIntentDateJST(30)})
	require.NoError(t, err)
	currentOwner := testdb.MakeTestOwner(t, db, clinicID, "current-read-owner")
	require.NoError(t, db.Model(&model.Pet{}).Where("id = ?", pet.ID).Update("owner_id", currentOwner.ID).Error)
	got, err := repo.FindByID(ctx, clinicID, intent.ID)
	require.NoError(t, err)
	response := toReservationResponse(got)
	require.NotNil(t, response.OwnerID)
	require.NotNil(t, response.Owner)
	assert.Equal(t, currentOwner.ID, *response.OwnerID)
	assert.Equal(t, currentOwner.ID, response.Owner.ID)
	rows, total, err := repo.FindAll(ctx, []uint64{clinicID}, 1, 20, nil, nil, nil, nil, nil, &pet.ID, &currentOwner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, rows, 1)
	assert.Equal(t, intent.ID, rows[0].ID)
	rows, total, err = repo.FindAll(ctx, []uint64{clinicID}, 1, 20, nil, nil, nil, nil, nil, &pet.ID, &formerOwner.ID)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, rows)
	_, err = repo.FindByID(ctx, 2, intent.ID)
	assert.True(t, apperrors.IsNotFound(err))
	require.NoError(t, db.Model(&model.Reservation{}).Where("id = ?", intent.ID).Update("status", model.ReservationStatusPending).Error)
	_, err = repo.FindByID(ctx, clinicID, intent.ID)
	assert.True(t, apperrors.IsNotFound(err), "normal bookings keep their owner mismatch guard")
	require.NoError(t, db.Model(&model.Reservation{}).Where("id = ?", intent.ID).Update("status", model.ReservationStatusTentative).Error)
	foreignOwner, _, _ := makeNextVisitIntentFixture(t, db, 2, "foreign-read")
	require.NoError(t, db.Model(&model.Pet{}).Where("id = ?", pet.ID).Update("owner_id", foreignOwner.ID).Error)
	_, err = repo.FindByID(ctx, clinicID, intent.ID)
	assert.True(t, apperrors.IsNotFound(err), "tentative does not bypass current owner's clinic")
}

func TestReservationIntentSafety_GuardedClearRealDB(t *testing.T) {
	if testing.Short() {
		t.Skip("requires an isolated PostgreSQL database")
	}
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db).(*reservationRepository)
	ctx := context.Background()
	const clinicID = uint64(1)
	_, pet, _ := makeNextVisitIntentFixture(t, db, clinicID, "guarded-clear")
	intent, err := setNextVisitIntentInTx(ctx, db, repo, clinicID, pet.ID, SetNextVisitIntentInput{IntendedDate: futureIntentDateJST(30)})
	require.NoError(t, err)
	expected := &ClearNextVisitIntentExpectation{ReservationID: intent.ID, UpdatedAt: intent.UpdatedAt}
	revision := intent.UpdatedAt.Add(time.Second)
	require.NoError(t, db.Model(&model.Reservation{}).Where("id = ?", intent.ID).Update("updated_at", revision).Error)
	tx := testNewTransactor(db)
	err = tx.WithTx(ctx, func(ctx context.Context) error {
		_, _, err := repo.clearNextVisitIntentWithOwner(ctx, clinicID, pet.ID, expected)
		return err
	})
	require.True(t, apperrors.IsConflict(err))
	got, err := repo.FindByID(ctx, clinicID, intent.ID)
	require.NoError(t, err)
	assert.Equal(t, model.ReservationStatusTentative, got.Status)
	expected.UpdatedAt = got.UpdatedAt
	require.NoError(t, tx.WithTx(ctx, func(ctx context.Context) error {
		found, _, err := repo.clearNextVisitIntentWithOwner(ctx, clinicID, pet.ID, expected)
		assert.True(t, found)
		return err
	}))
	err = tx.WithTx(ctx, func(ctx context.Context) error {
		_, _, err := repo.clearNextVisitIntentWithOwner(ctx, clinicID, pet.ID, expected)
		return err
	})
	require.True(t, apperrors.IsConflict(err))
	// Unconditional legacy clear remains idempotent after deletion.
	require.NoError(t, tx.WithTx(ctx, func(ctx context.Context) error {
		found, _, err := repo.clearNextVisitIntentWithOwner(ctx, clinicID, pet.ID)
		assert.False(t, found)
		return err
	}))
}
