package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func makeOwnedReservationForAmbientTx(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, status model.ReservationStatus) *model.Reservation {
	t.Helper()
	res := makeReservationForReservationRepoTest(t, db, clinicID,
		makeReservationType(t, db, clinicID).ID,
		time.Now().UTC().Truncate(time.Minute), status, model.ReservationSourceManual, nil, nil)
	require.NoError(t, db.Model(&model.Reservation{}).Where("id = ?", res.ID).Update("owner_id", ownerID).Error)
	res.OwnerID = &ownerID
	return res
}

func TestReservationRepository_OwnerMutationsAmbientTx(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db).(*reservationRepository)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "ambient-owner-a")
	ownerB := testdb.MakeTestOwner(t, db, clinicA, "ambient-owner-b")
	forced := errors.New("forced ambient rollback")

	cases := []struct {
		name   string
		status model.ReservationStatus
		call   func(t *testing.T, txCtx context.Context, id uint64) *uint64
	}{
		{
			name:   "updateWithPreviousOwner",
			status: model.ReservationStatusPending,
			call: func(t *testing.T, txCtx context.Context, id uint64) *uint64 {
				got, prev, err := repo.updateWithPreviousOwner(txCtx, clinicA, id, map[string]any{"notes": "ambient-notes"})
				require.NoError(t, err)
				assert.Equal(t, "ambient-notes", got.Notes)
				seen, findErr := repo.FindByID(txCtx, clinicA, id)
				require.NoError(t, findErr)
				assert.Equal(t, "ambient-notes", seen.Notes)
				return prev
			},
		},
		{
			name:   "deleteWithPreviousOwner",
			status: model.ReservationStatusPending,
			call: func(t *testing.T, txCtx context.Context, id uint64) *uint64 {
				prev, err := repo.deleteWithPreviousOwner(txCtx, clinicA, id)
				require.NoError(t, err)
				_, findErr := repo.FindByID(txCtx, clinicA, id)
				require.True(t, apperrors.IsNotFound(findErr))
				return prev
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := makeOwnedReservationForAmbientTx(t, db, clinicA, ownerA.ID, tc.status)
			bctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			err := tx.WithTx(bctx, func(txCtx context.Context) error {
				require.NoError(t, persistence.DBOrTx(txCtx, db).Model(&model.Reservation{}).
					Where("id = ?", res.ID).Update("owner_id", ownerB.ID).Error)
				prev := tc.call(t, txCtx, res.ID)
				require.NotNil(t, prev)
				assert.Equal(t, ownerB.ID, *prev)
				return forced
			})
			require.ErrorIs(t, err, forced)
			after, findErr := repo.FindByID(ctx, clinicA, res.ID)
			require.NoError(t, findErr)
			require.NotNil(t, after.OwnerID)
			assert.Equal(t, ownerA.ID, *after.OwnerID)
			assert.False(t, after.DeletedAt.Valid)
			assert.Equal(t, "", after.Notes)
		})
	}
}

func TestReservationRepository_CompleteForAccountingWithOwnersAmbientTx(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db).(*reservationRepository)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "acct-owner-a")
	ownerB := testdb.MakeTestOwner(t, db, clinicA, "acct-owner-b")
	pet := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "acct-pet")
	forced := errors.New("forced ambient rollback")

	res := makeOwnedReservationForAmbientTx(t, db, clinicA, ownerA.ID, model.ReservationStatusAccounting)
	require.NoError(t, db.Model(&model.Reservation{}).Where("id = ?", res.ID).Update("pet_id", pet.ID).Error)

	bctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err := tx.WithTx(bctx, func(txCtx context.Context) error {
		require.NoError(t, persistence.DBOrTx(txCtx, db).Model(&model.Pet{}).
			Where("id = ?", pet.ID).Update("owner_id", ownerB.ID).Error)
		require.NoError(t, persistence.DBOrTx(txCtx, db).Model(&model.Reservation{}).
			Where("id = ?", res.ID).Update("owner_id", ownerB.ID).Error)
		affected, owners, callErr := repo.CompleteForAccountingWithOwners(txCtx, clinicA, nil, &ownerB.ID, &pet.ID, res.StartTime)
		require.NoError(t, callErr)
		assert.Equal(t, int64(1), affected)
		assert.Equal(t, []uint64{ownerB.ID}, owners)
		seen, findErr := repo.FindByID(txCtx, clinicA, res.ID)
		require.NoError(t, findErr)
		assert.Equal(t, model.ReservationStatusCompleted, seen.Status)
		return forced
	})
	require.ErrorIs(t, err, forced)

	after, findErr := repo.FindByID(ctx, clinicA, res.ID)
	require.NoError(t, findErr)
	assert.Equal(t, model.ReservationStatusAccounting, after.Status)
	require.NotNil(t, after.OwnerID)
	assert.Equal(t, ownerA.ID, *after.OwnerID)

	affected, owners, err := repo.CompleteForAccountingWithOwners(ctx, clinicA, nil, &ownerB.ID, &pet.ID, res.StartTime)
	require.Error(t, err)
	assert.Equal(t, int64(0), affected)
	assert.Nil(t, owners)
}

func TestReservationRepository_ClearNextVisitIntentWithOwnerAmbientTx(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db).(*reservationRepository)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	ownerA, petA, _ := makeNextVisitIntentFixture(t, db, clinicA, "clear-ambient")
	ownerB := testdb.MakeTestOwner(t, db, clinicA, "clear-ambient-owner-b")
	forced := errors.New("forced ambient rollback")

	intent, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: futureIntentDateJST(30)})
	require.NoError(t, err)
	require.NotNil(t, intent)
	require.Equal(t, model.ReservationStatusTentative, intent.Status)

	bctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err = tx.WithTx(bctx, func(txCtx context.Context) error {
		require.NoError(t, persistence.DBOrTx(txCtx, db).Model(&model.Reservation{}).
			Where("id = ?", intent.ID).Update("owner_id", ownerB.ID).Error)
		found, prevOwner, callErr := repo.clearNextVisitIntentWithOwner(txCtx, clinicA, petA.ID)
		require.NoError(t, callErr)
		assert.True(t, found)
		require.NotNil(t, prevOwner)
		assert.Equal(t, ownerB.ID, *prevOwner)
		var stored model.Reservation
		require.NoError(t, persistence.DBOrTx(txCtx, db).Unscoped().
			Select("status", "deleted_at").Where("id = ?", intent.ID).Take(&stored).Error)
		assert.Equal(t, model.ReservationStatusCancelled, stored.Status)
		assert.True(t, stored.DeletedAt.Valid)
		return forced
	})
	require.ErrorIs(t, err, forced)

	var stored model.Reservation
	require.NoError(t, db.Unscoped().Where("id = ?", intent.ID).Take(&stored).Error)
	assert.Equal(t, model.ReservationStatusTentative, stored.Status)
	assert.False(t, stored.DeletedAt.Valid)
	require.NotNil(t, stored.OwnerID)
	assert.Equal(t, ownerA.ID, *stored.OwnerID)

	found, prevOwner, err := repo.clearNextVisitIntentWithOwner(ctx, clinicA, petA.ID)
	require.Error(t, err)
	assert.False(t, found)
	assert.Nil(t, prevOwner)

	otherPet := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "clear-ambient-pet-none")
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		found, prevOwner, callErr := repo.clearNextVisitIntentWithOwner(txCtx, clinicA, otherPet.ID)
		require.NoError(t, callErr)
		assert.False(t, found)
		assert.Nil(t, prevOwner)
		return nil
	}))
}

func TestReservationRepository_AssertLineCustomerReservationOwnerAmbientTx(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db).(*reservationRepository)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "line-owner-a")
	ownerB := testdb.MakeTestOwner(t, db, clinicA, "line-owner-b")
	customer := makeLineCustomerForAdmin(t, db, clinicA, "U-ambient-owner")
	require.NoError(t, db.Model(&model.LineCustomer{}).Where("id = ?", customer.ID).Update("owner_id", ownerA.ID).Error)
	forced := errors.New("forced ambient rollback")

	bctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err := tx.WithTx(bctx, func(txCtx context.Context) error {
		require.NoError(t, persistence.DBOrTx(txCtx, db).Model(&model.LineCustomer{}).
			Where("id = ?", customer.ID).Update("owner_id", ownerB.ID).Error)
		require.NoError(t, repo.assertLineCustomerReservationOwner(txCtx, clinicA, customer.ID, &ownerB.ID))
		assert.True(t, apperrors.IsConflict(repo.assertLineCustomerReservationOwner(txCtx, clinicA, customer.ID, &ownerA.ID)))
		assert.True(t, apperrors.IsNotFound(repo.assertLineCustomerReservationOwner(txCtx, clinicB, customer.ID, &ownerB.ID)))
		return forced
	})
	require.ErrorIs(t, err, forced)

	var stored model.LineCustomer
	require.NoError(t, db.Select("owner_id").Where("id = ?", customer.ID).Take(&stored).Error)
	require.NotNil(t, stored.OwnerID)
	assert.Equal(t, ownerA.ID, *stored.OwnerID)
}

func TestReservationRepository_AssertLineCustomerReservationOwnerShareLock(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db).(*reservationRepository)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "line-lock-owner-a")
	ownerB := testdb.MakeTestOwner(t, db, clinicA, "line-lock-owner-b")
	customer := makeLineCustomerForAdmin(t, db, clinicA, "U-ambient-lock")
	require.NoError(t, db.Model(&model.LineCustomer{}).Where("id = ?", customer.ID).Update("owner_id", ownerA.ID).Error)

	bctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	require.NoError(t, tx.WithTx(bctx, func(txCtx context.Context) error {
		require.NoError(t, repo.assertLineCustomerReservationOwner(txCtx, clinicA, customer.ID, &ownerA.ID))
		competing := db.WithContext(bctx).Begin()
		require.NoError(t, competing.Error)
		defer competing.Rollback()
		require.NoError(t, competing.Exec("SET LOCAL lock_timeout = '200ms'").Error)
		updateErr := competing.Model(&model.LineCustomer{}).
			Where("id = ?", customer.ID).
			Update("owner_id", ownerB.ID).Error
		var pgErr *pgconn.PgError
		require.ErrorAs(t, updateErr, &pgErr)
		assert.Equal(t, "55P03", pgErr.Code)
		return nil
	}))

	bctx2, cancel2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancel2()
	require.NoError(t, db.WithContext(bctx2).Model(&model.LineCustomer{}).
		Where("id = ?", customer.ID).Update("owner_id", ownerB.ID).Error)
}
