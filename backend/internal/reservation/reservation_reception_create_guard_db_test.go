package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestReceptionCreationGuardConcurrentTransactions(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	owner, pet, typ := makeNextVisitIntentFixture(t, db, 1, "guard-concurrent")
	repo := &reservationRepository{db: db}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	ready := make(chan int, 2)
	start := make(chan struct{})
	type outcome struct {
		id  uint64
		err error
	}
	results := make(chan outcome, 2)
	for range 2 {
		go func() {
			var id uint64
			err := db.WithContext(ctx).Connection(func(conn *gorm.DB) error {
				return conn.Transaction(func(tx *gorm.DB) error {
					var pid int
					if err := tx.Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
						return err
					}
					ready <- pid
					select {
					case <-start:
					case <-ctx.Done():
						return ctx.Err()
					}
					input := receptionCreateInput()
					booking := &model.Reservation{ClinicID: 1, OwnerID: &owner.ID, PetID: &pet.ID, ReservationTypeID: typ.ID, StartTime: input.StartTime, EndTime: input.EndTime, ReservationRoute: input.ReservationRoute, Status: model.ReservationStatusCheckedIn, Source: model.ReservationSourceManual, VisitType: model.VisitTypeRevisit}
					got, err := repo.createGuardedReception(persistence.WithTxValue(ctx, tx), booking)
					if got != nil {
						id = got.ID
					}
					return err
				})
			})
			results <- outcome{id: id, err: err}
		}()
	}
	pids := make([]int, 0, 2)
	for range 2 {
		select {
		case pid := <-ready:
			pids = append(pids, pid)
		case <-ctx.Done():
			close(start)
			t.Fatal("two independent database transactions did not become ready")
		}
	}
	require.NotEqual(t, pids[0], pids[1])
	close(start)
	successes, conflicts := 0, 0
	for range 2 {
		result := <-results
		if result.err == nil {
			successes++
			require.NotZero(t, result.id)
		} else {
			require.True(t, apperrors.IsConflict(result.err), result.err.Error())
			conflicts++
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, conflicts)
	var count int64
	require.NoError(t, db.Model(&model.Reservation{}).Where("clinic_id = ? AND pet_id = ? AND deleted_at IS NULL", 1, pet.ID).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestReceptionCreationGuardRepositoryBoundaries(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := &reservationRepository{db: db}
	for _, status := range []model.ReservationStatus{model.ReservationStatusPending, model.ReservationStatusConfirmed, model.ReservationStatusCheckedIn, model.ReservationStatusInConsultation, model.ReservationStatusAccounting, model.ReservationStatusCompleted, model.ReservationStatusCancelled, model.ReservationStatusNoShow} {
		t.Run(string(status), func(t *testing.T) {
			owner, pet, typ := makeNextVisitIntentFixture(t, db, 1, "guard-"+string(status))
			input := receptionCreateInput()
			old := makeAbsorbBooking(owner.ID, pet.ID, typ.ID, input.StartTime, input.EndTime)
			old.Status = status
			require.NoError(t, db.Create(old).Error)
			booking := makeAbsorbBooking(owner.ID, pet.ID, typ.ID, input.StartTime, input.EndTime)
			booking.Status = model.ReservationStatusCheckedIn
			booking.ReservationRoute = input.ReservationRoute
			var got *model.Reservation
			err := testNewTransactor(db).WithTx(context.Background(), func(ctx context.Context) error {
				var err error
				got, err = repo.createGuardedReception(ctx, booking)
				return err
			})
			if status == model.ReservationStatusCompleted || status == model.ReservationStatusCancelled || status == model.ReservationStatusNoShow {
				require.NoError(t, err)
				require.NotEqual(t, old.ID, got.ID)
			} else {
				require.True(t, apperrors.IsConflict(err))
				require.Nil(t, got)
			}
		})
	}
	t.Run("tentative is promoted and wrong clinic/death are rejected", func(t *testing.T) {
		owner, pet, typ := makeNextVisitIntentFixture(t, db, 2, "guard-tentative")
		input := receptionCreateInput()
		tentative, err := setNextVisitIntentInTx(context.Background(), db, repo, 2, pet.ID, SetNextVisitIntentInput{IntendedDate: input.StartTime})
		require.NoError(t, err)
		booking := makeAbsorbBooking(owner.ID, pet.ID, typ.ID, input.StartTime, input.EndTime)
		booking.ClinicID = 2
		booking.Status = model.ReservationStatusCheckedIn
		booking.ReservationRoute = input.ReservationRoute
		var got *model.Reservation
		call := func() error {
			return testNewTransactor(db).WithTx(context.Background(), func(ctx context.Context) error {
				var err error
				got, err = repo.createGuardedReception(ctx, booking)
				return err
			})
		}
		booking.ClinicID = 1
		require.Error(t, call())
		booking.ClinicID = 2
		require.NoError(t, db.Model(&model.Pet{}).Where("id = ? AND clinic_id = ?", pet.ID, 2).Update("status", model.PetStatusDeceased).Error)
		require.Error(t, call())
		require.NoError(t, db.Model(&model.Pet{}).Where("id = ? AND clinic_id = ?", pet.ID, 2).Update("status", model.PetStatusAlive).Error)
		require.NoError(t, call())
		require.Equal(t, tentative.ID, got.ID)
		require.Equal(t, model.ReservationStatusCheckedIn, got.Status)
	})
}

func TestReceptionCreationGuardIgnoresOtherDayClinicAndDeleted(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := &reservationRepository{db: db}
	for _, kind := range []string{"another day", "another clinic", "deleted"} {
		t.Run(kind, func(t *testing.T) {
			owner, pet, typ := makeNextVisitIntentFixture(t, db, 1, "guard-exclude-"+kind)
			input := receptionCreateInput()
			old := makeAbsorbBooking(owner.ID, pet.ID, typ.ID, input.StartTime, input.EndTime)
			if kind == "another day" {
				old.StartTime = old.StartTime.AddDate(0, 0, -1)
				old.EndTime = old.EndTime.AddDate(0, 0, -1)
			}
			if kind == "another clinic" {
				makeReservationClinicCorrelationClinic(t, db, 2, "guard-other-clinic")
				old.ClinicID = 2
			}
			require.NoError(t, db.Create(old).Error)
			if kind == "deleted" {
				require.NoError(t, db.Delete(old).Error)
			}
			booking := makeAbsorbBooking(owner.ID, pet.ID, typ.ID, input.StartTime, input.EndTime)
			booking.Status = model.ReservationStatusCheckedIn
			booking.ReservationRoute = input.ReservationRoute
			err := testNewTransactor(db).WithTx(context.Background(), func(ctx context.Context) error { _, err := repo.createGuardedReception(ctx, booking); return err })
			require.NoError(t, err)
		})
	}
}

func TestReceptionCreationGuardBatchConflictRollsBackEarlierInsert(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := &reservationRepository{db: db}
	owner, pet, typ := makeNextVisitIntentFixture(t, db, 1, "guard-rollback")
	pet2 := makePetForReservationIntentTest(t, db, 1, owner.ID, "guard-second-pet")
	input := receptionCreateInput()
	input.ReservationTypeID = typ.ID
	existing := makeAbsorbBooking(owner.ID, pet2.ID, typ.ID, input.StartTime, input.EndTime)
	require.NoError(t, db.Create(existing).Error)
	svc := &reservationService{repo: repo, tx: testNewTransactor(db)}
	got, err := svc.CreateBatch(context.Background(), input, []ReservationBatchPet{{OwnerID: owner.ID, PetID: pet2.ID}, {OwnerID: owner.ID, PetID: pet.ID}})
	require.True(t, apperrors.IsConflict(err))
	require.Nil(t, got)
	var count int64
	require.NoError(t, db.Model(&model.Reservation{}).Where("clinic_id = ? AND pet_id = ? AND deleted_at IS NULL", 1, pet.ID).Count(&count).Error)
	require.Zero(t, count)
}
