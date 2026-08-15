package reservation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func makePetForReservationIntentTest(
	t *testing.T,
	db *gorm.DB,
	clinicID, ownerID uint64,
	name string,
) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: "予約intentテスト犬"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: species.ID,
		Name:            name,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

func markNoShowInReservationRepoTest(
	ctx context.Context,
	db *gorm.DB,
	repo ReservationIntentRepository,
	clinicID, appointmentID uint64,
) (NoShowTransition, error) {
	var transition NoShowTransition
	err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var err error
		transition, err = repo.MarkNoShow(txCtx, clinicID, appointmentID)
		return err
	})
	return transition, err
}

func createForTrimmingInReservationRepoTest(
	ctx context.Context,
	db *gorm.DB,
	repo ReservationIntentRepository,
	clinicID uint64,
	input CreateTrimmingReservationInput,
) (*model.Reservation, error) {
	var created *model.Reservation
	err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var err error
		created, err = repo.CreateForTrimming(txCtx, clinicID, input)
		return err
	})
	return created, err
}

func updateForTrimmingInReservationRepoTest(
	ctx context.Context,
	db *gorm.DB,
	repo ReservationIntentRepository,
	clinicID, appointmentID uint64,
	input UpdateTrimmingReservationInput,
) (*model.Reservation, error) {
	var updated *model.Reservation
	err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var err error
		updated, err = repo.UpdateForTrimming(txCtx, clinicID, appointmentID, input)
		return err
	})
	return updated, err
}

func deleteForTrimmingInReservationRepoTest(
	ctx context.Context,
	db *gorm.DB,
	repo ReservationIntentRepository,
	clinicID, appointmentID uint64,
) error {
	return testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		return repo.DeleteForTrimming(txCtx, clinicID, appointmentID)
	})
}

func TestReservationRepository_BackfillForMedicalRecord(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "カルテ連携飼主A")
	petA := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "カルテ連携ペットA")
	doctorA := makeDoctor(t, db, clinicA, "カルテ連携獣医A")
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID:  doctorA.ID,
		ClinicID: clinicA,
		IsMain:   true,
	}).Error)

	t.Run("fills only missing context inside the ambient transaction", func(t *testing.T) {
		appt := makeReservation(t, db, clinicA)

		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			return repo.BackfillForMedicalRecord(
				txCtx,
				clinicA,
				appt.ID,
				&ownerA.ID,
				&petA.ID,
				&doctorA.ID,
			)
		})
		require.NoError(t, err)

		var got model.Reservation
		require.NoError(t, db.First(&got, appt.ID).Error)
		require.NotNil(t, got.OwnerID)
		require.NotNil(t, got.PetID)
		require.NotNil(t, got.DoctorID)
		assert.Equal(t, ownerA.ID, *got.OwnerID)
		assert.Equal(t, petA.ID, *got.PetID)
		assert.Equal(t, doctorA.ID, *got.DoctorID)
	})

	t.Run("rejects a foreign clinic owner before persisting", func(t *testing.T) {
		appt := makeReservation(t, db, clinicA)
		ownerB := testdb.MakeTestOwner(t, db, clinicB, "カルテ連携飼主B")

		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			return repo.BackfillForMedicalRecord(txCtx, clinicA, appt.ID, &ownerB.ID, nil, nil)
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var got model.Reservation
		require.NoError(t, db.First(&got, appt.ID).Error)
		assert.Nil(t, got.OwnerID)
	})

	t.Run("rejects a doctor assigned only to another clinic", func(t *testing.T) {
		appt := makeReservation(t, db, clinicA)
		doctorB := makeDoctor(t, db, clinicB, "カルテ連携獣医B")
		require.NoError(t, db.Create(&model.StaffClinicAssignment{
			StaffID:  doctorB.ID,
			ClinicID: clinicB,
			IsMain:   true,
		}).Error)

		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			return repo.BackfillForMedicalRecord(txCtx, clinicA, appt.ID, nil, nil, &doctorB.ID)
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var got model.Reservation
		require.NoError(t, db.First(&got, appt.ID).Error)
		assert.Nil(t, got.DoctorID)
	})

	t.Run("rejects a concurrent owner mismatch instead of overwriting it", func(t *testing.T) {
		appt := makeReservation(t, db, clinicA)
		otherOwner := testdb.MakeTestOwner(t, db, clinicA, "カルテ連携別飼主")
		require.NoError(t, db.Model(&model.Reservation{}).
			Where("id = ?", appt.ID).
			Update("owner_id", otherOwner.ID).Error)

		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			return repo.BackfillForMedicalRecord(txCtx, clinicA, appt.ID, &ownerA.ID, nil, nil)
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))

		var got model.Reservation
		require.NoError(t, db.First(&got, appt.ID).Error)
		require.NotNil(t, got.OwnerID)
		assert.Equal(t, otherOwner.ID, *got.OwnerID)
	})

	t.Run("rejects a trimming appointment for a general medical record", func(t *testing.T) {
		reservationType := &model.ReservationType{
			ClinicID: clinicA,
			Name:     "通常カルテへ紐付け不可のトリミング区分",
			Category: model.ReservationTypeCategoryTrimming,
			IsActive: true,
		}
		require.NoError(t, db.Create(reservationType).Error)
		appt := makeReservationForReservationRepoTest(
			t, db, clinicA, reservationType.ID, time.Now().UTC().Add(time.Hour),
			model.ReservationStatusPending, model.ReservationSourceManual, nil, nil,
		)

		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			return repo.BackfillForMedicalRecord(txCtx, clinicA, appt.ID, &ownerA.ID, &petA.ID, nil)
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))

		var got model.Reservation
		require.NoError(t, db.First(&got, appt.ID).Error)
		assert.Nil(t, got.OwnerID)
		assert.Nil(t, got.PetID)
	})

	t.Run("rolls back with the caller transaction", func(t *testing.T) {
		appt := makeReservation(t, db, clinicA)
		sentinel := errors.New("rollback reservation context")

		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			if err := repo.BackfillForMedicalRecord(txCtx, clinicA, appt.ID, &ownerA.ID, &petA.ID, nil); err != nil {
				return err
			}
			return sentinel
		})
		require.ErrorIs(t, err, sentinel)

		var got model.Reservation
		require.NoError(t, db.First(&got, appt.ID).Error)
		assert.Nil(t, got.OwnerID)
		assert.Nil(t, got.PetID)
	})
}

func TestReservationRepository_MarkNoShow(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("is clinic-scoped and idempotent", func(t *testing.T) {
		rt := makeReservationType(t, db, clinicA)
		appt := makeReservationForReservationRepoTest(
			t,
			db,
			clinicA,
			rt.ID,
			time.Now().UTC().Add(-6*time.Hour),
			model.ReservationStatusConfirmed,
			model.ReservationSourceManual,
			nil,
			nil,
		)

		transition, err := markNoShowInReservationRepoTest(ctx, db, repo, clinicB, appt.ID)
		require.NoError(t, err)
		assert.False(t, transition.Changed)

		transition, err = markNoShowInReservationRepoTest(ctx, db, repo, clinicA, appt.ID)
		require.NoError(t, err)
		assert.True(t, transition.Changed)
		assert.Equal(t, model.ReservationStatusConfirmed, transition.PreviousStatus)
		assert.Equal(t, model.ReservationStatusNoShow, reloadReservationIntentStatus(t, db, appt.ID))

		transition, err = markNoShowInReservationRepoTest(ctx, db, repo, clinicA, appt.ID)
		require.NoError(t, err)
		assert.False(t, transition.Changed, "a retry must not count an already transitioned reservation")
	})

	t.Run("does not overwrite an ineligible status", func(t *testing.T) {
		rt := makeReservationType(t, db, clinicA)
		appt := makeReservationForReservationRepoTest(
			t,
			db,
			clinicA,
			rt.ID,
			time.Now().UTC().Add(-6*time.Hour),
			model.ReservationStatusCompleted,
			model.ReservationSourceManual,
			nil,
			nil,
		)

		transition, err := markNoShowInReservationRepoTest(ctx, db, repo, clinicA, appt.ID)
		require.NoError(t, err)
		assert.False(t, transition.Changed)
		assert.Equal(t, model.ReservationStatusCompleted, reloadReservationIntentStatus(t, db, appt.ID))
	})

	t.Run("does not overwrite an appointment with a finalized medical record", func(t *testing.T) {
		rt := makeReservationType(t, db, clinicA)
		appt := makeReservationForReservationRepoTest(
			t,
			db,
			clinicA,
			rt.ID,
			time.Now().UTC().Add(-6*time.Hour),
			model.ReservationStatusConfirmed,
			model.ReservationSourceManual,
			nil,
			nil,
		)
		record := &model.MedicalRecord{
			ClinicID:      clinicA,
			RecordNo:      "MR-NO-SHOW-FINALIZED",
			Date:          time.Now().UTC(),
			AppointmentID: &appt.ID,
			Status:        model.MedicalRecordStatusFinalized,
		}
		require.NoError(t, db.Create(record).Error)

		transition, err := markNoShowInReservationRepoTest(ctx, db, repo, clinicA, appt.ID)
		require.NoError(t, err)
		assert.False(t, transition.Changed)
		assert.Equal(t, model.ReservationStatusConfirmed, reloadReservationIntentStatus(t, db, appt.ID))
	})

	t.Run("caller rollback restores the transition", func(t *testing.T) {
		rt := makeReservationType(t, db, clinicA)
		appt := makeReservationForReservationRepoTest(
			t,
			db,
			clinicA,
			rt.ID,
			time.Now().UTC().Add(-6*time.Hour),
			model.ReservationStatusPending,
			model.ReservationSourceManual,
			nil,
			nil,
		)
		sentinel := errors.New("audit write failed")

		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			transition, markErr := repo.MarkNoShow(txCtx, clinicA, appt.ID)
			require.NoError(t, markErr)
			require.True(t, transition.Changed)
			return sentinel
		})
		require.ErrorIs(t, err, sentinel)
		assert.Equal(t, model.ReservationStatusPending, reloadReservationIntentStatus(t, db, appt.ID))
	})

	t.Run("only one concurrent worker performs the transition", func(t *testing.T) {
		rt := makeReservationType(t, db, clinicA)
		appt := makeReservationForReservationRepoTest(
			t,
			db,
			clinicA,
			rt.ID,
			time.Now().UTC().Add(-6*time.Hour),
			model.ReservationStatusPending,
			model.ReservationSourceManual,
			nil,
			nil,
		)

		results := make(chan NoShowTransition, 2)
		errs := make(chan error, 2)
		var wg sync.WaitGroup
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				transition, err := markNoShowInReservationRepoTest(ctx, db, repo, clinicA, appt.ID)
				results <- transition
				errs <- err
			}()
		}
		wg.Wait()
		close(results)
		close(errs)

		changedCount := 0
		for transition := range results {
			if transition.Changed {
				changedCount++
				assert.Equal(t, model.ReservationStatusPending, transition.PreviousStatus)
			}
		}
		for err := range errs {
			require.NoError(t, err)
		}
		assert.Equal(t, 1, changedCount)
		assert.Equal(t, model.ReservationStatusNoShow, reloadReservationIntentStatus(t, db, appt.ID))
	})

	t.Run("requires an ambient transaction", func(t *testing.T) {
		rt := makeReservationType(t, db, clinicA)
		appt := makeReservationForReservationRepoTest(
			t,
			db,
			clinicA,
			rt.ID,
			time.Now().UTC().Add(-6*time.Hour),
			model.ReservationStatusPending,
			model.ReservationSourceManual,
			nil,
			nil,
		)

		transition, err := repo.MarkNoShow(ctx, clinicA, appt.ID)
		require.Error(t, err)
		assert.False(t, transition.Changed)
		assert.Equal(t, model.ReservationStatusPending, reloadReservationIntentStatus(t, db, appt.ID))
	})
}

func TestReservationRepository_MarkNoShowAt_UsesSuppliedTimeInCAS(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	store := NewReservationRepository(db)
	repo, ok := store.(ReservationNoShowAtRepository)
	require.True(t, ok)
	ctx := context.Background()
	const clinicID = uint64(1)

	reservationType := makeReservationType(t, db, clinicID)
	evaluatedAt := time.Date(2030, 1, 2, 10, 0, 0, 0, time.UTC)
	eligibleEnd := evaluatedAt.Add(-4 * time.Hour)
	appointment := makeReservationForReservationRepoTest(
		t,
		db,
		clinicID,
		reservationType.ID,
		eligibleEnd.Add(-30*time.Minute),
		model.ReservationStatusConfirmed,
		model.ReservationSourceManual,
		nil,
		nil,
	)

	var beforeCutoff NoShowTransition
	err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var markErr error
		beforeCutoff, markErr = repo.MarkNoShowAt(
			txCtx,
			clinicID,
			appointment.ID,
			evaluatedAt.Add(-time.Minute),
		)
		return markErr
	})
	require.NoError(t, err)
	assert.False(t, beforeCutoff.Changed)
	assert.Equal(t, model.ReservationStatusConfirmed, reloadReservationIntentStatus(t, db, appointment.ID))

	var atCutoff NoShowTransition
	err = testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var markErr error
		atCutoff, markErr = repo.MarkNoShowAt(txCtx, clinicID, appointment.ID, evaluatedAt)
		return markErr
	})
	require.NoError(t, err)
	assert.True(t, atCutoff.Changed)
	assert.Equal(t, model.ReservationStatusConfirmed, atCutoff.PreviousStatus)
	assert.Equal(t, model.ReservationStatusNoShow, reloadReservationIntentStatus(t, db, appointment.ID))
}

func TestReservationRepository_NoShowAndMedicalRecordFinalizationSerialize(t *testing.T) {
	const clinicID = uint64(1)
	ctx := context.Background()

	newFixture := func(t *testing.T) (*gorm.DB, ReservationStore, *model.Reservation, *model.MedicalRecord) {
		t.Helper()
		db := setupReservationRepoTestDB(t)
		repo := NewReservationRepository(db)
		reservationType := makeReservationType(t, db, clinicID)
		appointment := makeReservationForReservationRepoTest(
			t,
			db,
			clinicID,
			reservationType.ID,
			time.Now().UTC().Add(-6*time.Hour),
			model.ReservationStatusConfirmed,
			model.ReservationSourceManual,
			nil,
			nil,
		)
		record := &model.MedicalRecord{
			ClinicID:      clinicID,
			RecordNo:      fmt.Sprintf("MR-LIFECYCLE-%d", appointment.ID),
			Date:          time.Now().UTC(),
			AppointmentID: &appointment.ID,
			Status:        model.MedicalRecordStatusDraft,
		}
		require.NoError(t, db.Create(record).Error)
		return db, repo, appointment, record
	}

	t.Run("finalization wins and no-show becomes a no-op", func(t *testing.T) {
		db, repo, appointment, record := newFixture(t)
		finalizationLocked := make(chan struct{})
		releaseFinalization := make(chan struct{})
		finalizationDone := make(chan error, 1)
		go func() {
			finalizationDone <- testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
				if err := repo.PrepareForMedicalRecordFinalization(txCtx, clinicID, appointment.ID); err != nil {
					return err
				}
				if err := persistence.DBOrTx(txCtx, db).
					Model(&model.MedicalRecord{}).
					Where("clinic_id = ? AND id = ?", clinicID, record.ID).
					Update("status", model.MedicalRecordStatusFinalized).
					Error; err != nil {
					return err
				}
				close(finalizationLocked)
				<-releaseFinalization
				return nil
			})
		}()
		<-finalizationLocked

		type noShowResult struct {
			transition NoShowTransition
			err        error
		}
		noShowDone := make(chan noShowResult, 1)
		go func() {
			transition, err := markNoShowInReservationRepoTest(ctx, db, repo, clinicID, appointment.ID)
			noShowDone <- noShowResult{transition: transition, err: err}
		}()

		var result noShowResult
		completedEarly := false
		select {
		case result = <-noShowDone:
			completedEarly = true
			close(releaseFinalization)
			t.Errorf("no-show did not wait for finalization: transition=%+v err=%v", result.transition, result.err)
		case <-time.After(100 * time.Millisecond):
			close(releaseFinalization)
		}
		require.NoError(t, <-finalizationDone)
		if !completedEarly {
			result = <-noShowDone
		}
		require.NoError(t, result.err)
		assert.False(t, result.transition.Changed)
		assert.Equal(t, model.ReservationStatusConfirmed, reloadReservationIntentStatus(t, db, appointment.ID))
		var finalized model.MedicalRecord
		require.NoError(t, db.First(&finalized, record.ID).Error)
		assert.Equal(t, model.MedicalRecordStatusFinalized, finalized.Status)
	})

	t.Run("no-show wins and later finalization is rejected", func(t *testing.T) {
		db, repo, appointment, record := newFixture(t)
		noShowApplied := make(chan struct{})
		releaseNoShow := make(chan struct{})
		noShowDone := make(chan error, 1)
		go func() {
			noShowDone <- testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
				transition, err := repo.MarkNoShow(txCtx, clinicID, appointment.ID)
				if err != nil {
					return err
				}
				if !transition.Changed {
					return errors.New("expected no-show transition")
				}
				close(noShowApplied)
				<-releaseNoShow
				return nil
			})
		}()
		<-noShowApplied

		finalizationDone := make(chan error, 1)
		go func() {
			finalizationDone <- testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
				return repo.PrepareForMedicalRecordFinalization(txCtx, clinicID, appointment.ID)
			})
		}()

		var finalizationErr error
		completedEarly := false
		select {
		case finalizationErr = <-finalizationDone:
			completedEarly = true
			close(releaseNoShow)
			t.Errorf("finalization did not wait for no-show: err=%v", finalizationErr)
		case <-time.After(100 * time.Millisecond):
			close(releaseNoShow)
		}
		require.NoError(t, <-noShowDone)
		if !completedEarly {
			finalizationErr = <-finalizationDone
		}
		require.Error(t, finalizationErr)
		assert.True(t, apperrors.IsConflict(finalizationErr))
		assert.Equal(t, model.ReservationStatusNoShow, reloadReservationIntentStatus(t, db, appointment.ID))
		var draft model.MedicalRecord
		require.NoError(t, db.First(&draft, record.ID).Error)
		assert.Equal(t, model.MedicalRecordStatusDraft, draft.Status)
	})
}

func TestReservationRepository_TrimmingIntentsAreClinicScoped(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	reservationType := &model.ReservationType{
		ClinicID: clinicA,
		Name:     "トリミングintentテスト区分",
		Category: model.ReservationTypeCategoryTrimming,
	}
	require.NoError(t, db.Create(reservationType).Error)
	start := time.Date(2026, 7, 22, 1, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	created, err := createForTrimmingInReservationRepoTest(ctx, db, repo, clinicA, CreateTrimmingReservationInput{
		ReservationTypeID: reservationType.ID,
		StartTime:         start,
		EndTime:           end,
		Status:            model.ReservationStatusPending,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, clinicA, created.ClinicID)
	assert.Equal(t, model.ReservationSourceManual, created.Source)

	confirmed := model.ReservationStatusConfirmed
	_, err = updateForTrimmingInReservationRepoTest(ctx, db, repo, clinicB, created.ID, UpdateTrimmingReservationInput{Status: &confirmed})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Equal(t, model.ReservationStatusPending, reloadReservationIntentStatus(t, db, created.ID))

	updated, err := updateForTrimmingInReservationRepoTest(ctx, db, repo, clinicA, created.ID, UpdateTrimmingReservationInput{Status: &confirmed})
	require.NoError(t, err)
	assert.Equal(t, model.ReservationStatusConfirmed, updated.Status)
	assert.Equal(t, reservationType.ID, updated.ReservationTypeID, "typed intent must not expose reservation_type_id updates")

	err = deleteForTrimmingInReservationRepoTest(ctx, db, repo, clinicB, created.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	require.NoError(t, deleteForTrimmingInReservationRepoTest(ctx, db, repo, clinicA, created.ID))
	var remaining int64
	require.NoError(t, db.Model(&model.Reservation{}).Where("id = ?", created.ID).Count(&remaining).Error)
	assert.Zero(t, remaining)
}

func TestReservationRepository_CreateForTrimming_PersistsOwnerForAccountingCompletion(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicID, "トリミング会計飼主")
	pet := makePetForReservationIntentTest(t, db, clinicID, owner.ID, "トリミング会計ペット")
	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "トリミング会計区分",
		Category: model.ReservationTypeCategoryTrimming,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)
	start := time.Date(2026, 7, 22, 1, 0, 0, 0, time.UTC)

	var created *model.Reservation
	require.NoError(t, testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var err error
		created, err = repo.CreateForTrimming(txCtx, clinicID, CreateTrimmingReservationInput{
			ReservationTypeID: reservationType.ID,
			StartTime:         start,
			EndTime:           start.Add(time.Hour),
			PetID:             &pet.ID,
			Status:            model.ReservationStatusAccounting,
		})
		return err
	}))
	require.NotNil(t, created)
	require.NotNil(t, created.OwnerID)
	assert.Equal(t, owner.ID, *created.OwnerID)

	var affected int64
	require.NoError(t, testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var err error
		affected, err = repo.CompleteForAccounting(txCtx, clinicID, nil, &owner.ID, &pet.ID, start)
		return err
	}))
	assert.Equal(t, int64(1), affected)
	assert.Equal(t, model.ReservationStatusCompleted, reloadReservationIntentStatus(t, db, created.ID))
}

func TestReservationRepository_CompleteForAccountingRequiresAmbientTransaction(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)

	_, err := repo.CompleteForAccounting(context.Background(), 1, nil, nil, nil, time.Time{})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
}

func TestReservationRepository_TrimmingWriteIntentsRequireAmbientTransaction(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "ambient tx 必須区分",
		Category: model.ReservationTypeCategoryTrimming,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)
	appointment := makeReservationForReservationRepoTest(
		t, db, clinicID, reservationType.ID, time.Now().UTC().Add(time.Hour),
		model.ReservationStatusPending, model.ReservationSourceManual, nil, nil,
	)

	assertInternal := func(t *testing.T, err error) {
		t.Helper()
		require.Error(t, err)
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, "INTERNAL", appErr.Code)
	}

	_, err := repo.CreateForTrimming(ctx, clinicID, CreateTrimmingReservationInput{
		ReservationTypeID: reservationType.ID,
		StartTime:         appointment.StartTime.Add(time.Hour),
		EndTime:           appointment.EndTime.Add(time.Hour),
		Status:            model.ReservationStatusPending,
	})
	assertInternal(t, err)

	_, err = repo.LockTrimmingByID(ctx, clinicID, appointment.ID)
	assertInternal(t, err)

	confirmed := model.ReservationStatusConfirmed
	_, err = repo.UpdateForTrimming(ctx, clinicID, appointment.ID, UpdateTrimmingReservationInput{Status: &confirmed})
	assertInternal(t, err)

	err = repo.DeleteForTrimming(ctx, clinicID, appointment.ID)
	assertInternal(t, err)

	err = repo.AcquireBookingLock(ctx, clinicID)
	assertInternal(t, err)

	_, err = repo.LockAndFindByID(ctx, clinicID, appointment.ID)
	assertInternal(t, err)

	_, err = repo.HasDoctorConflict(ctx, clinicID, 1, appointment.StartTime, appointment.EndTime, nil)
	assertInternal(t, err)

	_, err = repo.CountConflicts(ctx, clinicID, appointment.StartTime, appointment.EndTime, nil)
	assertInternal(t, err)
}

func TestReservationRepository_TrimmingReservationTypeShareLocksSerializeMasterMutation(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "共有ロック対象区分",
		Category: model.ReservationTypeCategoryTrimming,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)

	locked := make(chan struct{})
	release := make(chan struct{})
	createDone := make(chan error, 1)
	go func() {
		createDone <- testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			start := time.Now().UTC().Add(time.Hour)
			if _, err := repo.CreateForTrimming(txCtx, clinicID, CreateTrimmingReservationInput{
				ReservationTypeID: reservationType.ID,
				StartTime:         start,
				EndTime:           start.Add(time.Hour),
				Status:            model.ReservationStatusPending,
			}); err != nil {
				return err
			}
			close(locked)
			<-release
			return nil
		})
	}()
	<-locked

	updateStarted := make(chan struct{})
	updateDone := make(chan error, 1)
	go func() {
		close(updateStarted)
		updateDone <- db.Model(&model.ReservationType{}).
			Where("id = ? AND clinic_id = ?", reservationType.ID, clinicID).
			Update("category", model.ReservationTypeCategoryGeneral).Error
	}()
	<-updateStarted

	select {
	case err := <-updateDone:
		close(release)
		require.Failf(t, "reservation type update was not serialized", "err=%v", err)
	case <-time.After(100 * time.Millisecond):
		close(release)
	}
	require.NoError(t, <-createDone)
	require.NoError(t, <-updateDone)
}

func TestReservationRepository_ClinicScopedFKAssertionsHoldShareLocks(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	assertMutationBlocks := func(t *testing.T, assertFK func(context.Context) error, mutate func() error) {
		t.Helper()
		locked := make(chan struct{})
		release := make(chan struct{})
		holderDone := make(chan error, 1)
		go func() {
			holderDone <- testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
				if err := assertFK(txCtx); err != nil {
					return err
				}
				close(locked)
				<-release
				return nil
			})
		}()
		<-locked

		mutationDone := make(chan error, 1)
		go func() { mutationDone <- mutate() }()
		var mutationErr error
		completedEarly := false
		select {
		case mutationErr = <-mutationDone:
			completedEarly = true
			close(release)
			t.Errorf("clinic-scoped FK mutation was not serialized: err=%v", mutationErr)
		case <-time.After(100 * time.Millisecond):
			close(release)
		}
		require.NoError(t, <-holderDone)
		if !completedEarly {
			mutationErr = <-mutationDone
		}
		require.NoError(t, mutationErr)
	}

	t.Run("owner", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicID, "共有ロック対象飼主")
		assertMutationBlocks(t,
			func(txCtx context.Context) error { return repo.AssertOwnerInClinic(txCtx, clinicID, owner.ID) },
			func() error { return db.Delete(&model.Owner{}, owner.ID).Error },
		)
	})

	t.Run("line customer", func(t *testing.T) {
		customer := makeLineCustomerForAdmin(t, db, clinicID, "share-lock-line-customer")
		assertMutationBlocks(t,
			func(txCtx context.Context) error {
				return repo.AssertLineCustomerInClinic(txCtx, clinicID, customer.ID)
			},
			func() error { return db.Delete(&model.LineCustomer{}, customer.ID).Error },
		)
	})
}

func TestReservationRepository_LockTrimmingByID_SerializesReservationTypeMutation(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "既存トリミング共有ロック区分",
		Category: model.ReservationTypeCategoryTrimming,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)
	appointment := makeReservationForReservationRepoTest(
		t, db, clinicID, reservationType.ID, time.Now().UTC().Add(time.Hour),
		model.ReservationStatusPending, model.ReservationSourceManual, nil, nil,
	)

	locked := make(chan struct{})
	release := make(chan struct{})
	lockDone := make(chan error, 1)
	go func() {
		lockDone <- testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			if _, err := repo.LockTrimmingByID(txCtx, clinicID, appointment.ID); err != nil {
				return err
			}
			close(locked)
			<-release
			return nil
		})
	}()
	<-locked

	updateStarted := make(chan struct{})
	updateDone := make(chan error, 1)
	go func() {
		close(updateStarted)
		updateDone <- db.Model(&model.ReservationType{}).
			Where("id = ? AND clinic_id = ?", reservationType.ID, clinicID).
			Update("category", model.ReservationTypeCategoryGeneral).Error
	}()
	<-updateStarted
	select {
	case err := <-updateDone:
		close(release)
		require.Failf(t, "reservation type update was not serialized", "err=%v", err)
	case <-time.After(100 * time.Millisecond):
		close(release)
	}
	require.NoError(t, <-lockDone)
	require.NoError(t, <-updateDone)
}

func TestReservationRepository_CreateForTrimming_SerializesStaffAssignmentMutation(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "担当者所属共有ロック区分",
		Category: model.ReservationTypeCategoryTrimming,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)
	doctor := makeDoctor(t, db, clinicID, "所属共有ロック獣医")
	assignment := &model.StaffClinicAssignment{StaffID: doctor.ID, ClinicID: clinicID, IsMain: true}
	require.NoError(t, db.Create(assignment).Error)

	created := make(chan struct{})
	release := make(chan struct{})
	createDone := make(chan error, 1)
	go func() {
		createDone <- testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			start := time.Now().UTC().Add(time.Hour)
			if _, err := repo.CreateForTrimming(txCtx, clinicID, CreateTrimmingReservationInput{
				ReservationTypeID: reservationType.ID,
				StartTime:         start,
				EndTime:           start.Add(time.Hour),
				DoctorID:          &doctor.ID,
				Status:            model.ReservationStatusPending,
			}); err != nil {
				return err
			}
			close(created)
			<-release
			return nil
		})
	}()
	<-created

	deleteStarted := make(chan struct{})
	deleteDone := make(chan error, 1)
	go func() {
		close(deleteStarted)
		deleteDone <- db.Delete(&model.StaffClinicAssignment{}, assignment.ID).Error
	}()
	<-deleteStarted
	select {
	case err := <-deleteDone:
		close(release)
		require.Failf(t, "staff assignment delete was not serialized", "err=%v", err)
	case <-time.After(100 * time.Millisecond):
		close(release)
	}
	require.NoError(t, <-createDone)
	require.NoError(t, <-deleteDone)
}

func TestReservationRepository_FindAllByCategoryScopesJoinedTablesToAppointmentClinic(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	typeA := &model.ReservationType{
		ClinicID: clinicA,
		Name:     "院Aトリミング区分",
		Category: model.ReservationTypeCategoryTrimming,
	}
	typeB := &model.ReservationType{
		ClinicID: clinicB,
		Name:     "院Bトリミング区分",
		Category: model.ReservationTypeCategoryTrimming,
	}
	require.NoError(t, db.Create(typeA).Error)
	require.NoError(t, db.Create(typeB).Error)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "一覧JOIN飼主A")
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "一覧JOIN飼主B")
	petA := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "一覧JOINペットA")
	petB := makePetForReservationIntentTest(t, db, clinicB, ownerB.ID, "一覧JOINペットB")
	start := time.Now().UTC().Add(time.Hour)

	valid := makeReservationForReservationRepoTest(
		t, db, clinicA, typeA.ID, start, model.ReservationStatusPending, model.ReservationSourceManual, nil, nil,
	)
	require.NoError(t, db.Model(&model.Reservation{}).Where("id = ?", valid.ID).Update("pet_id", petA.ID).Error)
	foreignType := makeReservationForReservationRepoTest(
		t, db, clinicA, typeB.ID, start.Add(time.Hour), model.ReservationStatusPending, model.ReservationSourceManual, nil, nil,
	)
	foreignPet := makeReservationForReservationRepoTest(
		t, db, clinicA, typeA.ID, start.Add(2*time.Hour), model.ReservationStatusPending, model.ReservationSourceManual, nil, nil,
	)
	require.NoError(t, db.Model(&model.Reservation{}).Where("id = ?", foreignPet.ID).Update("pet_id", petB.ID).Error)

	items, _, err := repo.FindAllByCategory(
		ctx, clinicA, model.ReservationTypeCategoryTrimming, nil, nil, nil, nil, 1, 50,
	)
	require.NoError(t, err)
	ids := make([]uint64, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
	}
	assert.Contains(t, ids, valid.ID)
	assert.NotContains(t, ids, foreignType.ID, "a foreign-clinic reservation type must not classify an appointment")

	items, total, err := repo.FindAllByCategory(
		ctx, clinicA, model.ReservationTypeCategoryTrimming, nil, &ownerB.ID, nil, nil, 1, 50,
	)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, items, "a foreign-clinic pet must not satisfy an owner filter")

	items, total, err = repo.FindAllByCategory(
		ctx, clinicA, model.ReservationTypeCategoryTrimming, &petB.ID, nil, nil, nil, 1, 50,
	)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, items, "a foreign-clinic pet must not satisfy a pet filter")

	// Simulate legacy pollution where a same-clinic pet points at another clinic's owner.
	require.NoError(t, db.Model(&model.Pet{}).Where("id = ?", petA.ID).Update("owner_id", ownerB.ID).Error)
	items, total, err = repo.FindAllByCategory(
		ctx, clinicA, model.ReservationTypeCategoryTrimming, nil, &ownerB.ID, nil, nil, 1, 50,
	)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, items, "a same-clinic pet with a foreign owner must fail closed")
}

func TestReservationRepository_CreateForTrimmingRejectsForeignReservationType(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	foreignType := &model.ReservationType{
		ClinicID: clinicB,
		Name:     "別院トリミング区分",
		Category: model.ReservationTypeCategoryTrimming,
	}
	require.NoError(t, db.Create(foreignType).Error)
	start := time.Date(2026, 7, 22, 1, 0, 0, 0, time.UTC)

	created, err := createForTrimmingInReservationRepoTest(ctx, db, repo, clinicA, CreateTrimmingReservationInput{
		ReservationTypeID: foreignType.ID,
		StartTime:         start,
		EndTime:           start.Add(time.Hour),
		Status:            model.ReservationStatusPending,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, created)

	var count int64
	require.NoError(t, db.Model(&model.Reservation{}).
		Where("clinic_id = ? AND reservation_type_id = ?", clinicA, foreignType.ID).
		Count(&count).Error)
	assert.Zero(t, count)
}

func TestReservationRepository_CreateForTrimmingRejectsPetLinkedToForeignOwner(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "トリミング飼主A")
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "トリミング飼主B")
	pollutedPet := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "飼主関係汚染ペット")
	require.NoError(t, db.Model(&model.Pet{}).Where("id = ?", pollutedPet.ID).Update("owner_id", ownerB.ID).Error)
	reservationType := &model.ReservationType{
		ClinicID: clinicA,
		Name:     "飼主関係検証区分",
		Category: model.ReservationTypeCategoryTrimming,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)
	start := time.Now().UTC().Add(time.Hour)

	created, err := createForTrimmingInReservationRepoTest(ctx, db, repo, clinicA, CreateTrimmingReservationInput{
		ReservationTypeID: reservationType.ID,
		StartTime:         start,
		EndTime:           start.Add(time.Hour),
		PetID:             &pollutedPet.ID,
		Status:            model.ReservationStatusPending,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, created)

	var count int64
	require.NoError(t, db.Model(&model.Reservation{}).Where("clinic_id = ?", clinicA).Count(&count).Error)
	assert.Zero(t, count)
}

func TestReservationRepository_CreateForTrimmingRejectsInactiveReservationType(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "無効トリミング区分",
		Category: model.ReservationTypeCategoryTrimming,
	}
	require.NoError(t, db.Create(reservationType).Error)
	require.NoError(t, db.Model(reservationType).Update("is_active", false).Error)
	start := time.Now().UTC().Add(time.Hour)

	created, err := createForTrimmingInReservationRepoTest(ctx, db, repo, clinicID, CreateTrimmingReservationInput{
		ReservationTypeID: reservationType.ID,
		StartTime:         start,
		EndTime:           start.Add(time.Hour),
		Status:            model.ReservationStatusPending,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, created)
}

func TestReservationRepository_TrimmingIntentsRejectSameClinicGeneralAppointments(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	generalType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "一般診療区分",
		Category: model.ReservationTypeCategoryGeneral,
	}
	require.NoError(t, db.Create(generalType).Error)
	appointment := makeReservationForReservationRepoTest(
		t,
		db,
		clinicID,
		generalType.ID,
		time.Now().UTC().Add(time.Hour),
		model.ReservationStatusPending,
		model.ReservationSourceManual,
		nil,
		nil,
	)

	_, err := repo.FindTrimmingByID(ctx, clinicID, appointment.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	confirmed := model.ReservationStatusConfirmed
	_, err = updateForTrimmingInReservationRepoTest(ctx, db, repo, clinicID, appointment.ID, UpdateTrimmingReservationInput{Status: &confirmed})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Equal(t, model.ReservationStatusPending, reloadReservationIntentStatus(t, db, appointment.ID))

	err = testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		return repo.DeleteForTrimming(txCtx, clinicID, appointment.ID)
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Equal(t, model.ReservationStatusPending, reloadReservationIntentStatus(t, db, appointment.ID))
}

func TestReservationRepository_TrimmingMutationSafety(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "トリミング安全性区分",
		Category: model.ReservationTypeCategoryTrimming,
	}
	require.NoError(t, db.Create(reservationType).Error)

	t.Run("terminal status cannot be reopened", func(t *testing.T) {
		appointment := makeReservationForReservationRepoTest(
			t,
			db,
			clinicID,
			reservationType.ID,
			time.Now().UTC().Add(-time.Hour),
			model.ReservationStatusCompleted,
			model.ReservationSourceManual,
			nil,
			nil,
		)
		pending := model.ReservationStatusPending

		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			_, err := repo.UpdateForTrimming(txCtx, clinicID, appointment.ID, UpdateTrimmingReservationInput{Status: &pending})
			return err
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Equal(t, model.ReservationStatusCompleted, reloadReservationIntentStatus(t, db, appointment.ID))
	})

	t.Run("terminal appointments cannot be deleted", func(t *testing.T) {
		statuses := []model.ReservationStatus{
			model.ReservationStatusCompleted,
			model.ReservationStatusCancelled,
			model.ReservationStatusNoShow,
		}
		for _, status := range statuses {
			t.Run(string(status), func(t *testing.T) {
				appointment := makeReservationForReservationRepoTest(
					t,
					db,
					clinicID,
					reservationType.ID,
					time.Now().UTC(),
					status,
					model.ReservationSourceManual,
					nil,
					nil,
				)

				err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
					return repo.DeleteForTrimming(txCtx, clinicID, appointment.ID)
				})
				require.Error(t, err)
				assert.True(t, apperrors.IsConflict(err))

				var got model.Reservation
				require.NoError(t, db.First(&got, appointment.ID).Error)
				assert.Equal(t, status, got.Status)
			})
		}
	})

	t.Run("medical record dependency blocks deletion", func(t *testing.T) {
		appointment := makeReservationForReservationRepoTest(
			t,
			db,
			clinicID,
			reservationType.ID,
			time.Now().UTC(),
			model.ReservationStatusPending,
			model.ReservationSourceManual,
			nil,
			nil,
		)
		record := &model.MedicalRecord{
			ClinicID:      clinicID,
			RecordNo:      "MR-TRIMMING-DELETE-GUARD",
			Date:          time.Now().UTC(),
			AppointmentID: &appointment.ID,
			Status:        model.MedicalRecordStatusDraft,
		}
		require.NoError(t, db.Create(record).Error)

		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			return repo.DeleteForTrimming(txCtx, clinicID, appointment.ID)
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Equal(t, model.ReservationStatusPending, reloadReservationIntentStatus(t, db, appointment.ID))
	})

	t.Run("medical record dependency blocks patient and schedule changes", func(t *testing.T) {
		appointment := makeReservationForReservationRepoTest(
			t,
			db,
			clinicID,
			reservationType.ID,
			time.Now().UTC(),
			model.ReservationStatusPending,
			model.ReservationSourceManual,
			nil,
			nil,
		)
		record := &model.MedicalRecord{
			ClinicID:      clinicID,
			RecordNo:      "MR-TRIMMING-UPDATE-GUARD",
			Date:          time.Now().UTC(),
			AppointmentID: &appointment.ID,
			Status:        model.MedicalRecordStatusDraft,
		}
		require.NoError(t, db.Create(record).Error)
		newStart := appointment.StartTime.Add(time.Hour)

		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			_, err := repo.UpdateForTrimming(txCtx, clinicID, appointment.ID, UpdateTrimmingReservationInput{StartTime: &newStart})
			return err
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
	})

	t.Run("caller rollback restores update", func(t *testing.T) {
		appointment := makeReservationForReservationRepoTest(
			t,
			db,
			clinicID,
			reservationType.ID,
			time.Now().UTC(),
			model.ReservationStatusPending,
			model.ReservationSourceManual,
			nil,
			nil,
		)
		confirmed := model.ReservationStatusConfirmed
		sentinel := errors.New("rollback trimming update")

		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			if _, err := repo.UpdateForTrimming(txCtx, clinicID, appointment.ID, UpdateTrimmingReservationInput{Status: &confirmed}); err != nil {
				return err
			}
			return sentinel
		})
		require.ErrorIs(t, err, sentinel)
		assert.Equal(t, model.ReservationStatusPending, reloadReservationIntentStatus(t, db, appointment.ID))
	})

	t.Run("caller rollback restores deletion", func(t *testing.T) {
		appointment := makeReservationForReservationRepoTest(
			t,
			db,
			clinicID,
			reservationType.ID,
			time.Now().UTC(),
			model.ReservationStatusPending,
			model.ReservationSourceManual,
			nil,
			nil,
		)
		sentinel := errors.New("rollback trimming delete")

		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			if err := repo.DeleteForTrimming(txCtx, clinicID, appointment.ID); err != nil {
				return err
			}
			return sentinel
		})
		require.ErrorIs(t, err, sentinel)
		assert.Equal(t, model.ReservationStatusPending, reloadReservationIntentStatus(t, db, appointment.ID))
	})
}

func reloadReservationIntentStatus(t *testing.T, db *gorm.DB, id uint64) model.ReservationStatus {
	t.Helper()
	var reservation model.Reservation
	require.NoError(t, db.First(&reservation, id).Error)
	return reservation.Status
}
