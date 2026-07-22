package medicalrecord

// medical_record_owner_pet_preload_clinic_isolation_test.go — AUD-008
// 汚染された Owner/Pet FK を持つカルテから、別 clinic の個人情報が Preload されないことを検証する。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	reservationdomain "github.com/animal-ekarte/backend/internal/reservation"
)

func setupMedicalRecordOwnerPetPreloadDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupMedicalRecordListTestDB(t)
	db.Exec("TRUNCATE TABLE medical_records, pets, animal_species, owners CASCADE")
	return db
}

func TestMedicalRecordRepository_FindByID_FindAll_DoesNotPreloadForeignOwnerPet(t *testing.T) {
	db := setupMedicalRecordOwnerPetPreloadDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerB := makeTestOwner(t, db, clinicB, "医院Bの飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "医院Bのペット")

	ownerBID, petBID := ownerB.ID, petB.ID
	contaminated := &model.MedicalRecord{
		ClinicID: clinicA,
		Date:     time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC),
		OwnerID:  &ownerBID,
		PetID:    &petBID,
		Status:   model.MedicalRecordStatusDraft,
		RecordNo: "AUD008-CONTAMINATED",
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
		items, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{}, 1, 50)
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(1))
		var found *model.MedicalRecord
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

	t.Run("FindAll does not use foreign owner or pet names as search predicates", func(t *testing.T) {
		for _, search := range []string{ownerB.Name, petB.Name} {
			items, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: search}, 1, 50)
			require.NoError(t, err)
			assert.Zero(t, total, "foreign relation name must not affect a clinic-local search")
			assert.Empty(t, items)
		}
	})
}

func TestMedicalRecordService_Create_RejectsSameClinicPetLinkedToForeignOwner(t *testing.T) {
	db := setupMedicalRecordOwnerPetPreloadDB(t)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "医院Aの飼主")
	ownerB := makeTestOwner(t, db, clinicB, "医院Bの飼主")
	pollutedPet := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "飼主関係汚染ペット")
	require.NoError(t, db.Model(&model.Pet{}).
		Where("id = ?", pollutedPet.ID).
		Update("owner_id", ownerB.ID).Error)

	recordRepo := NewMedicalRecordRepository(db)
	reservationRepo := reservationdomain.NewReservationRepository(db)
	svc := NewMedicalRecordService(
		recordRepo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, testTransactor{db: db},
	)
	record, err := svc.Create(ctx, clinicA, &CreateMedicalRecordInput{
		Date:  time.Now().UTC(),
		PetID: &pollutedPet.ID,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, record)

	var count int64
	require.NoError(t, db.Model(&model.MedicalRecord{}).Where("clinic_id = ?", clinicA).Count(&count).Error)
	assert.Zero(t, count)
}

func TestMedicalRecordService_Create_SerializesDuplicateAppointmentRecords(t *testing.T) {
	db := setupMedicalRecordOwnerPetPreloadDB(t)
	require.NoError(t, db.Exec("TRUNCATE TABLE medical_records, appointments, reservation_types CASCADE").Error)
	const clinicID = uint64(1)

	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "同時カルテ作成テスト",
		Category: model.ReservationTypeCategoryGeneral,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)
	appointmentStart := time.Date(2026, 7, 24, 23, 30, 0, 0, time.UTC)
	appointment := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         appointmentStart,
		EndTime:           appointmentStart.Add(30 * time.Minute),
		VisitType:         model.VisitTypeRevisit,
		ReservationTypeID: reservationType.ID,
		Status:            model.ReservationStatusConfirmed,
		Source:            model.ReservationSourceManual,
		CustomerFields:    []byte(`{}`),
	}
	require.NoError(t, db.Create(appointment).Error)

	recordRepo := NewMedicalRecordRepository(db)
	reservationRepo := reservationdomain.NewReservationRepository(db)
	svc := NewMedicalRecordService(
		recordRepo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, testTransactor{db: db},
	)

	type createResult struct {
		record *model.MedicalRecord
		err    error
	}
	ready := make(chan struct{})
	results := make(chan createResult, 2)
	recordNumbers := []string{"MR-CONCURRENT-A", "MR-CONCURRENT-B"}
	requestDates := []time.Time{
		time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	}
	for i := range recordNumbers {
		go func(index int) {
			<-ready
			record, err := svc.Create(context.Background(), clinicID, &CreateMedicalRecordInput{
				RecordNo:      recordNumbers[index],
				Date:          requestDates[index],
				AppointmentID: &appointment.ID,
			})
			results <- createResult{record: record, err: err}
		}(i)
	}
	close(ready)

	first := <-results
	second := <-results
	require.NoError(t, first.err)
	require.NoError(t, second.err)
	require.NotNil(t, first.record)
	require.NotNil(t, second.record)
	assert.Equal(t, first.record.ID, second.record.ID, "both callers must observe the same appointment-linked record")

	var count int64
	require.NoError(t, db.Model(&model.MedicalRecord{}).
		Where("clinic_id = ? AND appointment_id = ? AND deleted_at IS NULL", clinicID, appointment.ID).
		Count(&count).Error)
	assert.Equal(t, int64(1), count)
	assert.Equal(t, "2026-07-25", first.record.Date.In(time.FixedZone("JST", 9*60*60)).Format(time.DateOnly))
}

func TestMedicalRecordRepository_Create_ParticipatesInAmbientTransaction(t *testing.T) {
	db := setupMedicalRecordOwnerPetPreloadDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "自院飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "自院ペット")
	ownerID, petID := owner.ID, pet.ID

	tx := testTransactor{db: db}
	sentinel := errors.New("force medical record create rollback")
	err := tx.WithTx(ctx, func(txCtx context.Context) error {
		rec := &model.MedicalRecord{
			ClinicID: clinicA,
			Date:     time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC),
			OwnerID:  &ownerID,
			PetID:    &petID,
			Status:   model.MedicalRecordStatusDraft,
			RecordNo: "AUD008-TX-CREATE",
		}
		if e := repo.Create(txCtx, rec); e != nil {
			return e
		}
		return sentinel
	})
	require.Error(t, err)

	var count int64
	require.NoError(t, db.WithContext(ctx).Model(&model.MedicalRecord{}).
		Where("record_no = ?", "AUD008-TX-CREATE").Count(&count).Error)
	assert.Zero(t, count, "medical record create must roll back with the ambient transaction")
}
