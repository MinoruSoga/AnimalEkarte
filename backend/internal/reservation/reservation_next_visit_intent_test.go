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
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupNextVisitIntentTestDB provisions the shared test schema plus the partial
// unique index from migration 026 (testdb is a GORM double — migration indexes
// are not replayed, so the gate creates it idempotently).
func setupNextVisitIntentTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupReservationRepoTestDB(t)
	// staff_clinic_assignments.clinic_id → clinics.id の FK を満たすため、
	// 会社/医院テーブルも揃える（医院行は fixture 側で OnConflict 作成する）。
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Company{}, &model.Clinic{}))
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uq_appointments_one_active_tentative
		ON appointments (clinic_id, pet_id)
		WHERE status = 'tentative' AND deleted_at IS NULL`).Error)
	return db
}

func setNextVisitIntentInTx(
	ctx context.Context,
	db *gorm.DB,
	repo ReservationNextVisitIntentRepository,
	clinicID, petID uint64,
	input SetNextVisitIntentInput,
) (*model.Reservation, error) {
	var got *model.Reservation
	err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var err error
		got, err = repo.SetNextVisitIntent(txCtx, clinicID, petID, input)
		return err
	})
	return got, err
}

func clearNextVisitIntentInTx(
	ctx context.Context,
	db *gorm.DB,
	repo ReservationNextVisitIntentRepository,
	clinicID, petID uint64,
) (bool, error) {
	var found bool
	err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var err error
		found, err = repo.ClearNextVisitIntent(txCtx, clinicID, petID)
		return err
	})
	return found, err
}

func confirmTentativeInTx(
	ctx context.Context,
	db *gorm.DB,
	repo ReservationNextVisitIntentRepository,
	clinicID, id uint64,
	input ConfirmTentativeUpdate,
) (*model.Reservation, error) {
	var got *model.Reservation
	err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		var err error
		got, err = repo.ConfirmTentative(txCtx, clinicID, id, input)
		return err
	})
	return got, err
}

// makeNextVisitIntentFixture builds owner+pet+general type for one clinic.
// 医院行は staff_clinic_assignments→clinics FK のため fixture 呼出毎に idempotent に用意する
// （他テストの TRUNCATE 後でも自己修復する）。
func makeNextVisitIntentFixture(t *testing.T, db *gorm.DB, clinicID uint64, label string) (*model.Owner, *model.Pet, *model.ReservationType) {
	t.Helper()
	makeReservationClinicCorrelationClinic(t, db, clinicID, "次回予定"+label+"医院")
	owner := testdb.MakeTestOwner(t, db, clinicID, "次回予定"+label+"飼主")
	pet := makePetForReservationIntentTest(t, db, clinicID, owner.ID, "次回予定"+label+"ペット")
	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "次回予定" + label + "一般区分",
		Category: model.ReservationTypeCategoryGeneral,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)
	return owner, pet, reservationType
}

func makeMedicalRecordForNextVisitIntentTest(t *testing.T, db *gorm.DB, clinicID, ownerID, petID uint64, label string) *model.MedicalRecord {
	t.Helper()
	record := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: fmt.Sprintf("MR-INTENT-%s", label),
		Date:     time.Now().UTC(),
		OwnerID:  &ownerID,
		PetID:    &petID,
		Status:   model.MedicalRecordStatusDraft,
	}
	require.NoError(t, db.Create(record).Error)
	return record
}

func countTentativeForPet(t *testing.T, db *gorm.DB, clinicID, petID uint64) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&model.Reservation{}).
		Where("clinic_id = ? AND pet_id = ? AND status = ? AND deleted_at IS NULL",
			clinicID, petID, model.ReservationStatusTentative).
		Count(&count).Error)
	return count
}

func TestReservationRepository_SetNextVisitIntent(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA, petA, typeA := makeNextVisitIntentFixture(t, db, clinicA, "A")
	intendedDate := time.Date(2030, 3, 10, 0, 0, 0, 0, config.JST)

	t.Run("requires an ambient transaction", func(t *testing.T) {
		got, err := repo.SetNextVisitIntent(ctx, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Equal(t, int64(0), countTentativeForPet(t, db, clinicA, petA.ID))
	})

	t.Run("creates a tentative appointment with JST sentinel and provenance", func(t *testing.T) {
		record := makeMedicalRecordForNextVisitIntentTest(t, db, clinicA, ownerA.ID, petA.ID, "SRC")
		got, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{
			IntendedDate:          intendedDate,
			SourceMedicalRecordID: &record.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, got)

		assert.Equal(t, model.ReservationStatusTentative, got.Status)
		assert.Equal(t, clinicA, got.ClinicID)
		require.NotNil(t, got.PetID)
		assert.Equal(t, petA.ID, *got.PetID)
		require.NotNil(t, got.OwnerID)
		assert.Equal(t, ownerA.ID, *got.OwnerID)
		assert.Equal(t, typeA.ID, got.ReservationTypeID)
		assert.Nil(t, got.DoctorID)
		require.NotNil(t, got.ReservationRoute)
		assert.Equal(t, "record_intent", *got.ReservationRoute)
		require.NotNil(t, got.IntendedDate)
		assert.Equal(t, "2030-03-10", got.IntendedDate.In(config.JST).Format(time.DateOnly))
		require.NotNil(t, got.SourceMedicalRecordID)
		assert.Equal(t, record.ID, *got.SourceMedicalRecordID)

		// Sentinel: JST 00:00–00:15 of the intended date.
		startJST := got.StartTime.In(config.JST)
		endJST := got.EndTime.In(config.JST)
		assert.Equal(t, "2030-03-10 00:00", startJST.Format("2006-01-02 15:04"))
		assert.Equal(t, "2030-03-10 00:15", endJST.Format("2006-01-02 15:04"))
	})

	t.Run("updates the existing tentative instead of creating a second row", func(t *testing.T) {
		existing, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		newDate := time.Date(2030, 4, 1, 0, 0, 0, 0, config.JST)
		updated, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: newDate})
		require.NoError(t, err)
		assert.Equal(t, existing.ID, updated.ID)
		require.NotNil(t, updated.IntendedDate)
		assert.Equal(t, "2030-04-01", updated.IntendedDate.In(config.JST).Format(time.DateOnly))
		assert.Equal(t, int64(1), countTentativeForPet(t, db, clinicA, petA.ID))
	})

	t.Run("rejects a pet owned by another clinic", func(t *testing.T) {
		_, _, _ = makeNextVisitIntentFixture(t, db, clinicB, "B")
		got, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, 999999, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, got)
	})

	t.Run("rejects a source medical record owned by another pet", func(t *testing.T) {
		otherPet := makePetForReservationIntentTest(t, db, clinicA, ownerA.ID, "次回予定別ペット")
		record := makeMedicalRecordForNextVisitIntentTest(t, db, clinicA, ownerA.ID, otherPet.ID, "WRONGPET")
		got, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{
			IntendedDate:          intendedDate,
			SourceMedicalRecordID: &record.ID,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, got)
	})

	t.Run("rejects a source medical record owned by another clinic", func(t *testing.T) {
		ownerB, petB, _ := makeNextVisitIntentFixture(t, db, clinicB, "B2")
		recordB := makeMedicalRecordForNextVisitIntentTest(t, db, clinicB, ownerB.ID, petB.ID, "FOREIGN")
		got, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{
			IntendedDate:          intendedDate,
			SourceMedicalRecordID: &recordB.ID,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, got)
	})

	t.Run("rolls back with the caller transaction", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回予定rollback飼主")
		pet := makePetForReservationIntentTest(t, db, clinicA, owner.ID, "次回予定rollbackペット")
		sentinel := errors.New("rollback intent write")
		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			if _, err := repo.SetNextVisitIntent(txCtx, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate}); err != nil {
				return err
			}
			return sentinel
		})
		require.ErrorIs(t, err, sentinel)
		assert.Equal(t, int64(0), countTentativeForPet(t, db, clinicA, pet.ID))
	})

	t.Run("concurrent sets produce exactly one tentative row", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回予定競合飼主")
		pet := makePetForReservationIntentTest(t, db, clinicA, owner.ID, "次回予定競合ペット")
		var wg sync.WaitGroup
		errs := make(chan error, 2)
		for i := range 2 {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				date := intendedDate.AddDate(0, 0, i)
				_, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: date})
				errs <- err
			}(i)
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			require.NoError(t, err)
		}
		assert.Equal(t, int64(1), countTentativeForPet(t, db, clinicA, pet.ID))
	})

	t.Run("partial unique index rejects a second raw tentative row", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回予定一意飼主")
		pet := makePetForReservationIntentTest(t, db, clinicA, owner.ID, "次回予定一意ペット")
		_, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		// Bypass the service/repository lock path with a raw INSERT to prove the
		// database-level guard is live under this schema.
		dup := &model.Reservation{
			ClinicID:          clinicA,
			ReservationTypeID: typeA.ID,
			StartTime:         intendedDate,
			EndTime:           intendedDate.Add(15 * time.Minute),
			OwnerID:           &owner.ID,
			PetID:             &pet.ID,
			Status:            model.ReservationStatusTentative,
			Source:            model.ReservationSourceManual,
			CustomerFields:    []byte("{}"),
		}
		createErr := db.Create(dup).Error
		require.Error(t, createErr)
		assert.True(t, persistence.IsUniqueConstraintErr(createErr))
	})
}

func TestReservationRepository_ClearNextVisitIntent(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	intendedDate := time.Date(2030, 5, 20, 0, 0, 0, 0, config.JST)

	t.Run("requires an ambient transaction", func(t *testing.T) {
		_, petA, _ := makeNextVisitIntentFixture(t, db, clinicA, "CLR")
		found, err := repo.ClearNextVisitIntent(ctx, clinicA, petA.ID)
		require.Error(t, err)
		assert.False(t, found)
	})

	t.Run("soft-deletes the active tentative and is idempotent", func(t *testing.T) {
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "CLR1")
		created, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		found, err := clearNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, int64(0), countTentativeForPet(t, db, clinicA, pet.ID))

		var row model.Reservation
		require.NoError(t, db.Unscoped().First(&row, created.ID).Error)
		assert.NotNil(t, row.DeletedAt)

		found, err = clearNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID)
		require.NoError(t, err)
		assert.False(t, found, "second clear must be an idempotent no-op")
	})

	t.Run("does not touch another clinic's tentative", func(t *testing.T) {
		_, petA, _ := makeNextVisitIntentFixture(t, db, clinicA, "CLR2A")
		_, petB, _ := makeNextVisitIntentFixture(t, db, clinicB, "CLR2B")
		_, err := setNextVisitIntentInTx(ctx, db, repo, clinicB, petB.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		// Clearing via clinicA + petB's id: the find predicate is clinic-scoped, so nothing deletes.
		found, err := clearNextVisitIntentInTx(ctx, db, repo, clinicA, petB.ID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Equal(t, int64(1), countTentativeForPet(t, db, clinicB, petB.ID))

		// Sanity: clinicA pet unaffected on a real clear.
		_, err = setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)
		found, err = clearNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID)
		require.NoError(t, err)
		assert.True(t, found)
	})

	t.Run("leaves non-tentative appointments untouched", func(t *testing.T) {
		_, pet, _ := makeNextVisitIntentFixture(t, db, clinicA, "CLR3")
		confirmed := makeReservationForReservationRepoTest(
			t, db, clinicA, makeReservationType(t, db, clinicA).ID,
			time.Now().UTC().Add(24*time.Hour),
			model.ReservationStatusConfirmed, model.ReservationSourceManual, nil, nil,
		)
		found, err := clearNextVisitIntentInTx(ctx, db, repo, clinicA, pet.ID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Equal(t, model.ReservationStatusConfirmed, reloadReservationIntentStatus(t, db, confirmed.ID))
	})
}

func TestReservationRepository_ConfirmTentative(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	intendedDate := time.Date(2030, 6, 15, 0, 0, 0, 0, config.JST)

	_, petA, typeA := makeNextVisitIntentFixture(t, db, clinicA, "CFM")

	newStart := time.Date(2030, 6, 15, 1, 0, 0, 0, time.UTC)
	newEnd := newStart.Add(30 * time.Minute)

	t.Run("requires an ambient transaction", func(t *testing.T) {
		got, err := repo.ConfirmTentative(ctx, clinicA, 1, ConfirmTentativeUpdate{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: typeA.ID,
			Status:            model.ReservationStatusConfirmed,
		})
		require.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("transitions a tentative appointment to a real booking", func(t *testing.T) {
		record := makeMedicalRecordForNextVisitIntentTest(t, db, clinicA, petA.ID, petA.ID, "CFMSRC")
		_ = record
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, petA.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		doctor := makeDoctor(t, db, clinicA, "次回予定確定獣医")
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: doctor.ID, ClinicID: clinicA, IsMain: true}).Error)

		confirmed, err := confirmTentativeInTx(ctx, db, repo, clinicA, tentative.ID, ConfirmTentativeUpdate{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: typeA.ID,
			DoctorID:          &doctor.ID,
			Status:            model.ReservationStatusConfirmed,
		})
		require.NoError(t, err)
		require.NotNil(t, confirmed)
		assert.Equal(t, model.ReservationStatusConfirmed, confirmed.Status)
		assert.True(t, confirmed.StartTime.Equal(newStart))
		assert.True(t, confirmed.EndTime.Equal(newEnd))
		require.NotNil(t, confirmed.DoctorID)
		assert.Equal(t, doctor.ID, *confirmed.DoctorID)
		// provenance columns survive the transition.
		require.NotNil(t, confirmed.IntendedDate)
		assert.Equal(t, "2030-06-15", confirmed.IntendedDate.In(config.JST).Format(time.DateOnly))
		require.NotNil(t, confirmed.ReservationRoute)
		assert.Equal(t, "record_intent", *confirmed.ReservationRoute)
	})

	t.Run("rejects a non-tentative appointment", func(t *testing.T) {
		normal := makeReservationForReservationRepoTest(
			t, db, clinicA, typeA.ID, time.Now().UTC().Add(48*time.Hour),
			model.ReservationStatusPending, model.ReservationSourceManual, nil, nil,
		)
		got, err := confirmTentativeInTx(ctx, db, repo, clinicA, normal.ID, ConfirmTentativeUpdate{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: typeA.ID,
			Status:            model.ReservationStatusConfirmed,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Nil(t, got)
		assert.Equal(t, model.ReservationStatusPending, reloadReservationIntentStatus(t, db, normal.ID))
	})

	t.Run("rejects a tentative owned by another clinic", func(t *testing.T) {
		_, petB, typeB := makeNextVisitIntentFixture(t, db, clinicB, "CFMB")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicB, petB.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		got, err := confirmTentativeInTx(ctx, db, repo, clinicA, tentative.ID, ConfirmTentativeUpdate{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: typeB.ID,
			Status:            model.ReservationStatusConfirmed,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, got)
		assert.Equal(t, model.ReservationStatusTentative, reloadReservationIntentStatus(t, db, tentative.ID))
	})

	t.Run("rejects a doctor assigned only to another clinic", func(t *testing.T) {
		owner2 := testdb.MakeTestOwner(t, db, clinicA, "次回予定確定飼主2")
		pet2 := makePetForReservationIntentTest(t, db, clinicA, owner2.ID, "次回予定確定ペット2")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet2.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		doctorB := makeDoctor(t, db, clinicB, "次回予定確定獣医B")
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: doctorB.ID, ClinicID: clinicB, IsMain: true}).Error)

		got, err := confirmTentativeInTx(ctx, db, repo, clinicA, tentative.ID, ConfirmTentativeUpdate{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: typeA.ID,
			DoctorID:          &doctorB.ID,
			Status:            model.ReservationStatusConfirmed,
		})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Equal(t, model.ReservationStatusTentative, reloadReservationIntentStatus(t, db, tentative.ID))
	})

	t.Run("rejects a target status outside pending/confirmed", func(t *testing.T) {
		owner3 := testdb.MakeTestOwner(t, db, clinicA, "次回予定確定飼主3")
		pet3 := makePetForReservationIntentTest(t, db, clinicA, owner3.ID, "次回予定確定ペット3")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet3.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		got, err := confirmTentativeInTx(ctx, db, repo, clinicA, tentative.ID, ConfirmTentativeUpdate{
			StartTime:         newStart,
			EndTime:           newEnd,
			ReservationTypeID: typeA.ID,
			Status:            model.ReservationStatusTentative,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, got)
	})

	t.Run("rolls back with the caller transaction", func(t *testing.T) {
		owner4 := testdb.MakeTestOwner(t, db, clinicA, "次回予定確定飼主4")
		pet4 := makePetForReservationIntentTest(t, db, clinicA, owner4.ID, "次回予定確定ペット4")
		tentative, err := setNextVisitIntentInTx(ctx, db, repo, clinicA, pet4.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
		require.NoError(t, err)

		sentinel := errors.New("audit write failed")
		err = testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			if _, err := repo.ConfirmTentative(txCtx, clinicA, tentative.ID, ConfirmTentativeUpdate{
				StartTime:         newStart,
				EndTime:           newEnd,
				ReservationTypeID: typeA.ID,
				Status:            model.ReservationStatusConfirmed,
			}); err != nil {
				return err
			}
			return sentinel
		})
		require.ErrorIs(t, err, sentinel)
		assert.Equal(t, model.ReservationStatusTentative, reloadReservationIntentStatus(t, db, tentative.ID))
	})
}
