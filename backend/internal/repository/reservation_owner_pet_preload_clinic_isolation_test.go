package repository

// BE9-2C R③: 本テストは ReservationAdminRepository（appointment_admin=R④未移行）も検証対象に
// 含むため repository 残置。R④（appointment）移動時に internal/reservation へ移す。

// reservation_owner_pet_preload_clinic_isolation_test.go — AUD-001
// 汚染された Owner/Pet/LineCustomer FK を持つ予約から、別 clinic の個人情報が Preload されないことを検証する。

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/trimming"
)

func setupReservationOwnerPetPreloadDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupReservationIsolationTestDB(t)
	require.NoError(t, ensureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{}, &model.LineCustomer{},
	))
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	db.Exec("TRUNCATE TABLE line_customers CASCADE")
	return db
}

func TestReservationRepository_FindAll_FindByID_DoesNotPreloadForeignOwnerPet(t *testing.T) {
	db := setupReservationOwnerPetPreloadDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerB := makeTestOwner(t, db, clinicB, "医院Bの飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "医院Bのペット")
	rtA := makeReservationType(t, db, clinicA)

	now := time.Now().UTC().Truncate(time.Minute)
	ownerBID, petBID := ownerB.ID, petB.ID
	contaminated := &model.Reservation{
		ClinicID:          clinicA,
		StartTime:         now,
		EndTime:           now.Add(15 * time.Minute),
		OwnerID:           &ownerBID,
		PetID:             &petBID,
		ReservationTypeID: rtA.ID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    json.RawMessage(`{}`),
	}
	require.NoError(t, db.WithContext(ctx).Create(contaminated).Error)

	t.Run("FindByID does not preload foreign owner/pet", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, contaminated.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, ownerBID, *got.OwnerID)
		assert.Equal(t, petBID, *got.PetID)
		assert.Nil(t, got.Owner, "foreign Owner must not be preloaded")
		assert.Nil(t, got.Pet, "foreign Pet must not be preloaded")
	})

	t.Run("FindAll does not preload foreign owner/pet", func(t *testing.T) {
		items, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 50, nil, nil, nil, nil, nil, nil, nil)
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(1))
		var found *model.Reservation
		for i := range items {
			if items[i].ID == contaminated.ID {
				found = &items[i]
				break
			}
		}
		require.NotNil(t, found)
		assert.Nil(t, found.Owner, "foreign Owner must not be preloaded")
		assert.Nil(t, found.Pet, "foreign Pet must not be preloaded")
	})
}

func TestReservationAdminRepository_DoesNotPreloadForeignOwnerPetLineCustomer(t *testing.T) {
	db := setupReservationAdminTestDB(t)
	repo := NewReservationAdminRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerB := makeTestOwner(t, db, clinicB, "医院B飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "医院Bペット")
	lcB := makeLineCustomerForAdmin(t, db, clinicB, "line-user-b")

	day := time.Date(2026, 7, 14, 10, 0, 0, 0, time.Local)
	ownerBID, petBID, lcBID := ownerB.ID, petB.ID, lcB.ID
	contaminated := makeAdminReservationAt(t, db, clinicA, day, &ownerBID, &petBID, nil, &lcBID)

	t.Run("FindAllByDay does not preload foreign links", func(t *testing.T) {
		items, err := repo.FindAllByDay(ctx, clinicA, day)
		require.NoError(t, err)
		var found *model.Reservation
		for i := range items {
			if items[i].ID == contaminated.ID {
				found = &items[i]
				break
			}
		}
		require.NotNil(t, found)
		assert.Nil(t, found.Owner)
		assert.Nil(t, found.Pet)
		assert.Nil(t, found.LineCustomer)
	})

	t.Run("FindAllByMonth does not preload foreign line customer", func(t *testing.T) {
		items, err := repo.FindAllByMonth(ctx, clinicA, 2026, time.July)
		require.NoError(t, err)
		var found *model.Reservation
		for i := range items {
			if items[i].ID == contaminated.ID {
				found = &items[i]
				break
			}
		}
		require.NotNil(t, found)
		assert.Nil(t, found.LineCustomer)
	})

	t.Run("FindByIDForNotify does not preload foreign owner/pet", func(t *testing.T) {
		got, err := repo.FindByIDForNotify(ctx, clinicA, contaminated.ID)
		require.NoError(t, err)
		assert.Nil(t, got.Owner)
		assert.Nil(t, got.Pet)
	})
}

func TestReservationRepository_AssertOwnerPetLineCustomer_ClinicIsolation(t *testing.T) {
	db := setupReservationOwnerPetPreloadDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "医院A飼主")
	ownerB := makeTestOwner(t, db, clinicB, "医院B飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "医院Aペット")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "医院Bペット")
	lcA := &model.LineCustomer{ClinicID: clinicA, LineUserID: "line-a", AdditionalFields: []byte(`{}`)}
	lcB := &model.LineCustomer{ClinicID: clinicB, LineUserID: "line-b", AdditionalFields: []byte(`{}`)}
	require.NoError(t, db.Create(lcA).Error)
	require.NoError(t, db.Create(lcB).Error)

	t.Run("AssertOwnerInClinic accepts same clinic", func(t *testing.T) {
		require.NoError(t, repo.AssertOwnerInClinic(ctx, clinicA, ownerA.ID))
	})
	t.Run("AssertOwnerInClinic rejects other clinic", func(t *testing.T) {
		err := repo.AssertOwnerInClinic(ctx, clinicA, ownerB.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
	t.Run("FindPetOwnerInClinic accepts same clinic", func(t *testing.T) {
		got, err := repo.FindPetOwnerInClinic(ctx, clinicA, petA.ID)
		require.NoError(t, err)
		assert.Equal(t, ownerA.ID, got)
	})
	t.Run("FindPetOwnerInClinic rejects other clinic", func(t *testing.T) {
		_, err := repo.FindPetOwnerInClinic(ctx, clinicA, petB.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
	t.Run("FindPetOwnerInClinic rejects same-clinic pet linked to foreign owner", func(t *testing.T) {
		pollutedPet := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "汚染ペット")
		require.NoError(t, db.Model(&model.Pet{}).Where("id = ?", pollutedPet.ID).Update("owner_id", ownerB.ID).Error)
		_, err := repo.FindPetOwnerInClinic(ctx, clinicA, pollutedPet.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
	t.Run("AssertLineCustomerInClinic accepts same clinic", func(t *testing.T) {
		require.NoError(t, repo.AssertLineCustomerInClinic(ctx, clinicA, lcA.ID))
	})
	t.Run("AssertLineCustomerInClinic rejects other clinic", func(t *testing.T) {
		err := repo.AssertLineCustomerInClinic(ctx, clinicA, lcB.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestReservationRepository_FindAllByCategory_DoesNotPreloadForeignPet(t *testing.T) {
	db := setupReservationOwnerPetPreloadDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	ownerB := makeTestOwner(t, db, clinicB, "医院B飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "医院Bペット")
	rtA := &model.ReservationType{
		ClinicID: clinicA,
		Name:     "トリミング",
		Category: model.ReservationTypeCategoryTrimming,
	}
	require.NoError(t, db.Create(rtA).Error)

	now := time.Now().UTC().Truncate(time.Minute)
	petBID := petB.ID
	contaminated := &model.Reservation{
		ClinicID:          clinicA,
		StartTime:         now,
		EndTime:           now.Add(15 * time.Minute),
		PetID:             &petBID,
		ReservationTypeID: rtA.ID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    json.RawMessage(`{}`),
	}
	require.NoError(t, db.Create(contaminated).Error)

	items, _, err := repo.FindAllByCategory(ctx, clinicA, model.ReservationTypeCategoryTrimming, nil, nil, nil, nil, 1, 50)
	require.NoError(t, err)
	var found *model.Reservation
	for i := range items {
		if items[i].ID == contaminated.ID {
			found = &items[i]
			break
		}
	}
	require.NotNil(t, found)
	assert.Nil(t, found.Pet, "foreign Pet must not be preloaded via FindAllByCategory")
}

func TestReservationRepository_FindAllByCategory_DoesNotPreloadForeignTrimmingDetail(t *testing.T) {
	db := setupReservationOwnerPetPreloadDB(t)
	require.NoError(t, ensureAutoMigrated(db, &model.AppointmentTrimmingDetail{}))
	repo := NewReservationRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	rtA := &model.ReservationType{
		ClinicID: clinicA,
		Name:     "トリミング詳細分離",
		Category: model.ReservationTypeCategoryTrimming,
	}
	require.NoError(t, db.Create(rtA).Error)
	appointment := makeReservation(t, db, clinicA)
	require.NoError(t, db.Model(&model.Reservation{}).
		Where("id = ?", appointment.ID).
		Update("reservation_type_id", rtA.ID).Error)
	require.NoError(t, db.Create(&model.AppointmentTrimmingDetail{
		ClinicID:      clinicB,
		AppointmentID: appointment.ID,
		BWUnit:        model.BodyWeightUnitKg,
		Remarks:       "別院のセンシティブな施術メモ",
	}).Error)

	items, _, err := repo.FindAllByCategory(ctx, clinicA, model.ReservationTypeCategoryTrimming, nil, nil, nil, nil, 1, 50)
	require.NoError(t, err)
	var found *model.Reservation
	for i := range items {
		if items[i].ID == appointment.ID {
			found = &items[i]
			break
		}
	}
	require.NotNil(t, found)
	assert.Nil(t, found.TrimmingDetail, "foreign-clinic trimming detail must not be preloaded")
}

func TestTrimmingPetValidationAndWrites_RollBackTogether(t *testing.T) {
	db := setupReservationOwnerPetPreloadDB(t)
	reservationRepo := NewReservationRepository(db)
	detailRepo := trimming.NewAppointmentTrimmingDetailRepository(db)
	transactor := NewTransactor(db)
	ctx := context.Background()

	const clinicID = uint64(1)
	owner := makeTestOwner(t, db, clinicID, "医院A飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "医院Aペット")
	appointment := makeReservation(t, db, clinicID)
	sentinel := errors.New("force rollback after trimming writes")

	err := transactor.WithTx(ctx, func(txCtx context.Context) error {
		petOwnerID, err := reservationRepo.FindPetOwnerInClinic(txCtx, clinicID, pet.ID)
		require.NoError(t, err)
		require.Equal(t, owner.ID, petOwnerID)

		if err := reservationRepo.BackfillForMedicalRecord(txCtx, clinicID, appointment.ID, nil, &pet.ID, nil); err != nil {
			return err
		}
		if err := detailRepo.Create(txCtx, &model.AppointmentTrimmingDetail{
			ClinicID:      clinicID,
			AppointmentID: appointment.ID,
			BWUnit:        model.BodyWeightUnitKg,
		}); err != nil {
			return err
		}
		return sentinel
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)

	var persisted model.Reservation
	require.NoError(t, db.First(&persisted, appointment.ID).Error)
	assert.Nil(t, persisted.PetID, "appointment pet update must roll back")

	var detailCount int64
	require.NoError(t, db.Model(&model.AppointmentTrimmingDetail{}).
		Where("clinic_id = ? AND appointment_id = ?", clinicID, appointment.ID).
		Count(&detailCount).Error)
	assert.Zero(t, detailCount, "trimming detail create must roll back with appointment update")
}
