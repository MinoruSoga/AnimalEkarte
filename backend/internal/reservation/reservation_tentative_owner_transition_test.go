package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func newTentativeTransitionService(db *gorm.DB, syncer NextVisitTagSyncer) *reservationService {
	return newReservationService(
		NewReservationRepository(db),
		NewReservationTypeRepository(db),
		testNewTransactor(db),
		nil,
		nil,
		nil,
		nil,
		openDayHolidayFinder(),
		syncer,
	)
}

type tentativeRowSnapshot struct {
	Status    string
	OwnerID   *uint64
	DeletedAt *time.Time
}

func readTentativeRowSnapshot(t *testing.T, db *gorm.DB, id uint64) tentativeRowSnapshot {
	t.Helper()
	var snap tentativeRowSnapshot
	require.NoError(t, db.Raw(
		"SELECT status::text, owner_id, deleted_at FROM appointments WHERE id = ?", id,
	).Scan(&snap).Error)
	return snap
}

func transferPetToOwner(t *testing.T, db *gorm.DB, petID, ownerID uint64) {
	t.Helper()
	require.NoError(t, db.Model(&model.Pet{}).Where("id = ?", petID).Update("owner_id", ownerID).Error)
}

func TestTentativeTransition_TransferredOwnerCheckedInWritesCurrentOwner(t *testing.T) {
	if testing.Short() {
		t.Skip("requires an isolated PostgreSQL database")
	}
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	formerOwner, pet, _ := makeNextVisitIntentFixture(t, db, clinicID, "tr-checkin")
	intent, err := setNextVisitIntentInTx(ctx, db, repo, clinicID, pet.ID,
		SetNextVisitIntentInput{IntendedDate: futureIntentDateJST(0)})
	require.NoError(t, err)

	currentOwner := testdb.MakeTestOwner(t, db, clinicID, "checkin-current-owner")
	transferPetToOwner(t, db, pet.ID, currentOwner.ID)

	syncer := &recordingNextVisitTagSyncer{}
	svc := newTentativeTransitionService(db, syncer)
	checkedIn := model.ReservationStatusCheckedIn
	updated, err := svc.Update(ctx, clinicID, intent.ID, &UpdateReservationInput{Status: &checkedIn})
	require.NoError(t, err)
	assert.Equal(t, intent.ID, updated.ID, "the transition reuses the tentative row")
	require.NotNil(t, updated.OwnerID)
	assert.Equal(t, currentOwner.ID, *updated.OwnerID,
		"check-in writes the pet's current owner into the same CAS update")

	reloaded, err := repo.FindByID(ctx, clinicID, intent.ID)
	require.NoError(t, err)
	require.NotNil(t, reloaded.OwnerID)
	assert.Equal(t, currentOwner.ID, *reloaded.OwnerID,
		"post-get also reports the transferred owner")
	assert.ElementsMatch(t, []uint64{formerOwner.ID, currentOwner.ID}, syncer.calls,
		"post-commit tag sync covers the previous and current owners")
}

func TestTentativeTransition_TransferredOwnerCancelWritesCurrentOwner(t *testing.T) {
	if testing.Short() {
		t.Skip("requires an isolated PostgreSQL database")
	}
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	formerOwner, pet, _ := makeNextVisitIntentFixture(t, db, clinicID, "tr-cancel")
	intent, err := setNextVisitIntentInTx(ctx, db, repo, clinicID, pet.ID,
		SetNextVisitIntentInput{IntendedDate: futureIntentDateJST(30)})
	require.NoError(t, err)

	currentOwner := testdb.MakeTestOwner(t, db, clinicID, "cancel-current-owner")
	transferPetToOwner(t, db, pet.ID, currentOwner.ID)

	syncer := &recordingNextVisitTagSyncer{}
	svc := newTentativeTransitionService(db, syncer)
	cancelled := model.ReservationStatusCancelled
	updated, err := svc.Update(ctx, clinicID, intent.ID, &UpdateReservationInput{Status: &cancelled})
	require.NoError(t, err)
	require.NotNil(t, updated.OwnerID)
	assert.Equal(t, currentOwner.ID, *updated.OwnerID,
		"cancel writes the pet's current owner into the same CAS update")

	snap := readTentativeRowSnapshot(t, db, intent.ID)
	assert.Equal(t, string(model.ReservationStatusCancelled), snap.Status)
	require.NotNil(t, snap.DeletedAt, "cancel soft-deletes the tentative row")
	require.NotNil(t, snap.OwnerID)
	assert.Equal(t, currentOwner.ID, *snap.OwnerID)
	assert.ElementsMatch(t, []uint64{formerOwner.ID, currentOwner.ID}, syncer.calls,
		"post-commit tag sync covers the previous and current owners")
}

func TestTentativeTransition_TransferredOwnerIDDeleteWritesCurrentOwner(t *testing.T) {
	if testing.Short() {
		t.Skip("requires an isolated PostgreSQL database")
	}
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	formerOwner, pet, _ := makeNextVisitIntentFixture(t, db, clinicID, "tr-delete")
	intent, err := setNextVisitIntentInTx(ctx, db, repo, clinicID, pet.ID,
		SetNextVisitIntentInput{IntendedDate: futureIntentDateJST(30)})
	require.NoError(t, err)

	currentOwner := testdb.MakeTestOwner(t, db, clinicID, "delete-current-owner")
	transferPetToOwner(t, db, pet.ID, currentOwner.ID)

	syncer := &recordingNextVisitTagSyncer{}
	svc := newTentativeTransitionService(db, syncer)
	require.NoError(t, svc.Delete(ctx, clinicID, intent.ID))

	snap := readTentativeRowSnapshot(t, db, intent.ID)
	assert.Equal(t, string(model.ReservationStatusCancelled), snap.Status,
		"ID delete marks cancelled before soft-deleting the tentative row")
	require.NotNil(t, snap.DeletedAt)
	require.NotNil(t, snap.OwnerID)
	assert.Equal(t, currentOwner.ID, *snap.OwnerID,
		"ID delete writes the pet's current owner into the same CAS update")
	assert.ElementsMatch(t, []uint64{formerOwner.ID, currentOwner.ID}, syncer.calls,
		"post-commit tag sync covers the previous and current owners")
}

func TestTentativeTransition_RowLockedUntilConcurrentWriteCommits(t *testing.T) {
	if testing.Short() {
		t.Skip("requires an isolated PostgreSQL database")
	}
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	_, pet, _ := makeNextVisitIntentFixture(t, db, clinicID, "tr-lock")
	intent, err := setNextVisitIntentInTx(ctx, db, repo, clinicID, pet.ID,
		SetNextVisitIntentInput{IntendedDate: futureIntentDateJST(0)})
	require.NoError(t, err)

	// A foreign transaction promotes the row to pending and holds it uncommitted
	// while the check-in arrives — the post-lock read must observe committed
	// state, so the CAS cannot resurrect a terminated tentative.
	holder := db.Begin()
	require.NoError(t, holder.Error)
	require.NoError(t, holder.Model(&model.Reservation{}).Where("id = ?", intent.ID).
		Update("status", model.ReservationStatusPending).Error)

	svc := newTentativeTransitionService(db, nil)
	checkedIn := model.ReservationStatusCheckedIn
	done := make(chan error, 1)
	go func() {
		_, err := svc.Update(ctx, clinicID, intent.ID, &UpdateReservationInput{Status: &checkedIn})
		done <- err
	}()

	blocked := false
	select {
	case err := <-done:
		blocked = false
		require.Error(t, err)
	case <-time.After(500 * time.Millisecond):
		blocked = true
	}
	if !blocked {
		// The update returned early: it must still have failed conflict-closed.
		snap := readTentativeRowSnapshot(t, db, intent.ID)
		assert.NotEqual(t, string(model.ReservationStatusCheckedIn), snap.Status)
	}
	require.NoError(t, holder.Commit().Error)

	if blocked {
		select {
		case err := <-done:
			require.Error(t, err)
		case <-time.After(10 * time.Second):
			t.Fatal("check-in did not return after the row lock was released")
		}
	}
	snap := readTentativeRowSnapshot(t, db, intent.ID)
	assert.Equal(t, string(model.ReservationStatusPending), snap.Status,
		"a concurrent promotion is never overwritten by the tentative transition")
}

func TestTentativeTransition_BoundariesPreserved(t *testing.T) {
	if testing.Short() {
		t.Skip("requires an isolated PostgreSQL database")
	}
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	_, petA, _ := makeNextVisitIntentFixture(t, db, clinicID, "tr-bound")
	futureIntent, err := setNextVisitIntentInTx(ctx, db, repo, clinicID, petA.ID,
		SetNextVisitIntentInput{IntendedDate: futureIntentDateJST(30)})
	require.NoError(t, err)

	svc := newTentativeTransitionService(db, nil)
	checkedIn := model.ReservationStatusCheckedIn

	t.Run("future intended_date refuses check-in", func(t *testing.T) {
		_, err := svc.Update(ctx, clinicID, futureIntent.ID, &UpdateReservationInput{Status: &checkedIn})
		require.Error(t, err)
	})

	t.Run("foreign clinic id refuses the transition", func(t *testing.T) {
		_, petB, _ := makeNextVisitIntentFixture(t, db, 2, "tr-bound-foreign")
		foreignIntent, err := setNextVisitIntentInTx(ctx, db, repo, 2, petB.ID,
			SetNextVisitIntentInput{IntendedDate: futureIntentDateJST(0)})
		require.NoError(t, err)
		_, err = svc.Update(ctx, clinicID, foreignIntent.ID, &UpdateReservationInput{Status: &checkedIn})
		require.Error(t, err)
	})

	t.Run("deceased pet refuses check-in", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicID, "deceased-bound-owner")
		pet := makePetForReservationIntentTest(t, db, clinicID, owner.ID, "deceased-bound-pet")
		intent, err := setNextVisitIntentInTx(ctx, db, repo, clinicID, pet.ID,
			SetNextVisitIntentInput{IntendedDate: futureIntentDateJST(0)})
		require.NoError(t, err)
		deceasedAt := time.Now().UTC()
		require.NoError(t, db.Model(&model.Pet{}).Where("id = ?", pet.ID).Updates(map[string]any{
			"status":      model.PetStatusDeceased,
			"deceased_at": deceasedAt,
		}).Error)
		_, err = svc.Update(ctx, clinicID, intent.ID, &UpdateReservationInput{Status: &checkedIn})
		require.Error(t, err)
	})
}
