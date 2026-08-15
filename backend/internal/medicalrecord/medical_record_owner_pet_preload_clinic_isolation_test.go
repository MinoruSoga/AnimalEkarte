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
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupMedicalRecordOwnerPetPreloadDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupMedicalRecordListTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ReservationType{}, &model.Reservation{}))
	db.Exec("TRUNCATE TABLE medical_records, pets, animal_species, owners CASCADE")
	return db
}

func TestMedicalRecordRepository_FindByIDAndFindAllRejectPollutedParent(t *testing.T) {
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

	t.Run("FindByID rejects a parent with foreign owner/pet relations", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, contaminated.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, got)
	})

	t.Run("FindAll rejects a parent with foreign owner/pet relations", func(t *testing.T) {
		items, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{}, 1, 50)
		require.NoError(t, err)
		assert.Zero(t, total)
		assert.Empty(t, items, "polluted raw owner_id/pet_id must not reach the list response")
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

func TestDB_MedicalRecordRepositoryFindByIDForClinicsCorrelatesRelationsToParentClinic(t *testing.T) {
	type foreignRelations struct {
		owner     *model.Owner
		pet       *model.Pet
		doctor    *model.Staff
		enteredBy *model.Staff
	}
	tests := []struct {
		name   string
		mutate func(*model.MedicalRecord, foreignRelations)
	}{
		{
			name: "foreign owner",
			mutate: func(record *model.MedicalRecord, foreign foreignRelations) {
				record.OwnerID = &foreign.owner.ID
			},
		},
		{
			name: "foreign pet",
			mutate: func(record *model.MedicalRecord, foreign foreignRelations) {
				record.PetID = &foreign.pet.ID
			},
		},
		{
			name: "foreign doctor",
			mutate: func(record *model.MedicalRecord, foreign foreignRelations) {
				record.DoctorID = &foreign.doctor.ID
			},
		},
		{
			name: "foreign entered-by staff",
			mutate: func(record *model.MedicalRecord, foreign foreignRelations) {
				record.EnteredBy = &foreign.enteredBy.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupMedicalRecordOwnerPetPreloadDB(t)
			repo := NewMedicalRecordRepository(db)
			ctx := context.Background()
			const clinicA, clinicB = uint64(1), uint64(2)
			ensureVaccinationTestClinics(t, db, clinicA, clinicB)

			ownerA := makeTestOwner(t, db, clinicA, "詳細取得自院飼主")
			petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "詳細取得自院ペット")
			ownerB := makeTestOwner(t, db, clinicB, "詳細取得別院飼主")
			petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "詳細取得別院ペット")
			doctorB := makeMedicalRecordListStaff(t, db, clinicB, "詳細取得別院担当医", model.StaffTypeDoctor)
			enteredByB := makeMedicalRecordListStaff(t, db, clinicB, "詳細取得別院入力者", model.StaffTypeNurse)
			for _, staffID := range []uint64{doctorB.ID, enteredByB.ID} {
				require.NoError(t, db.Create(&model.StaffClinicAssignment{
					StaffID: staffID, ClinicID: clinicB,
				}).Error)
			}

			record := &model.MedicalRecord{
				ClinicID: clinicA, RecordNo: "DETAIL-PARENT-CORRELATION", Date: time.Now(),
				OwnerID: &ownerA.ID, PetID: &petA.ID,
			}
			tt.mutate(record, foreignRelations{
				owner: ownerB, pet: petB, doctor: doctorB, enteredBy: enteredByB,
			})
			require.NoError(t, db.WithContext(ctx).Create(record).Error)

			got, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, record.ID)

			require.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
			assert.Nil(t, got)
		})
	}

	t.Run("staff assigned to the parent clinic remains visible when its primary clinic differs", func(t *testing.T) {
		db := setupMedicalRecordOwnerPetPreloadDB(t)
		repo := NewMedicalRecordRepository(db)
		ctx := context.Background()
		const clinicA, clinicB = uint64(1), uint64(2)
		ensureVaccinationTestClinics(t, db, clinicA, clinicB)

		ownerA := makeTestOwner(t, db, clinicA, "詳細取得正常飼主")
		petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "詳細取得正常ペット")
		doctor := makeMedicalRecordListStaff(t, db, clinicB, "詳細取得兼務担当医", model.StaffTypeDoctor)
		enteredBy := makeMedicalRecordListStaff(t, db, clinicB, "詳細取得兼務入力者", model.StaffTypeNurse)
		for _, staffID := range []uint64{doctor.ID, enteredBy.ID} {
			require.NoError(t, db.Create(&model.StaffClinicAssignment{
				StaffID: staffID, ClinicID: clinicA,
			}).Error)
		}
		record := &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "DETAIL-VALID-CROSS-PRIMARY-CLINIC", Date: time.Now(),
			OwnerID: &ownerA.ID, PetID: &petA.ID, DoctorID: &doctor.ID, EnteredBy: &enteredBy.ID,
		}
		require.NoError(t, db.WithContext(ctx).Create(record).Error)

		got, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, record.ID)

		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.Owner)
		assert.Equal(t, ownerA.ID, got.Owner.ID)
		require.NotNil(t, got.Pet)
		assert.Equal(t, petA.ID, got.Pet.ID)
		require.NotNil(t, got.Doctor)
		assert.Equal(t, doctor.ID, got.Doctor.ID)
		require.NotNil(t, got.EnteredByStaff)
		assert.Equal(t, enteredBy.ID, got.EnteredByStaff.ID)
	})

	t.Run("soft-deleted same-clinic relations keep the historical parent", func(t *testing.T) {
		db := setupMedicalRecordOwnerPetPreloadDB(t)
		repo := NewMedicalRecordRepository(db)
		ctx := context.Background()
		const clinicID = uint64(1)
		ensureVaccinationTestClinics(t, db, clinicID)

		owner := makeTestOwner(t, db, clinicID, "履歴飼主")
		pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "履歴ペット")
		staff := makeMedicalRecordListStaff(t, db, clinicID, "退職済み担当者", model.StaffTypeDoctor)
		assignment := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID}
		require.NoError(t, db.Create(assignment).Error)
		record := &model.MedicalRecord{
			ClinicID: clinicID, RecordNo: "DETAIL-HISTORICAL-RELATIONS", Date: time.Now(),
			OwnerID: &owner.ID, PetID: &pet.ID, DoctorID: &staff.ID, EnteredBy: &staff.ID,
		}
		require.NoError(t, db.WithContext(ctx).Create(record).Error)

		require.NoError(t, db.Delete(assignment).Error)
		require.NoError(t, db.Delete(staff).Error)
		require.NoError(t, db.Delete(pet).Error)
		require.NoError(t, db.Delete(owner).Error)

		got, err := repo.FindByIDForClinics(ctx, []uint64{clinicID}, record.ID)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, record.ID, got.ID)
		assert.Nil(t, got.Owner)
		assert.Nil(t, got.Pet)
		assert.Nil(t, got.Doctor)
		assert.Nil(t, got.EnteredByStaff)
	})

	t.Run("vitals are correlated to the parent clinic and pet", func(t *testing.T) {
		db := setupMedicalRecordOwnerPetPreloadDB(t)
		repo := NewMedicalRecordRepository(db)
		ctx := context.Background()
		const clinicA, clinicB = uint64(1), uint64(2)
		ensureVaccinationTestClinics(t, db, clinicA, clinicB)

		ownerA := makeTestOwner(t, db, clinicA, "バイタル自院飼主")
		petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "バイタル自院ペット")
		ownerB := makeTestOwner(t, db, clinicB, "バイタル別院飼主")
		petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "バイタル別院ペット")
		record := &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "DETAIL-VITAL-CORRELATION", Date: time.Now(),
			OwnerID: &ownerA.ID, PetID: &petA.ID,
		}
		require.NoError(t, db.WithContext(ctx).Create(record).Error)
		recordID := record.ID
		validVital := &model.VitalRecord{
			ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &recordID, RecordedAt: time.Now(),
		}
		foreignVital := &model.VitalRecord{
			ClinicID: clinicB, PetID: petB.ID, MedicalRecordID: &recordID, RecordedAt: time.Now(),
		}
		require.NoError(t, db.WithContext(ctx).Create(validVital).Error)
		require.NoError(t, db.WithContext(ctx).Create(foreignVital).Error)

		got, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, record.ID)

		require.NoError(t, err)
		require.NotNil(t, got)
		require.Len(t, got.Vitals, 1)
		assert.Equal(t, validVital.ID, got.Vitals[0].ID)
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
	svc := NewMedicalRecordServiceWithTxAudit(
		recordRepo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, nil, testTransactor{db: db})
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
	svc := NewMedicalRecordServiceWithTxAudit(
		recordRepo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, nil, testTransactor{db: db})

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

// TestMedicalRecordRepository_FindByAppointmentID_CorrelatesAppointmentClinic
// SEC-SWEEP-02-MR-B1: medical_records.appointment_id 読みは appointments.clinic_id と相関必須。
// 子行 clinic のみ一致して親 appointment が他院でも返ってしまう旧 failure mode を固定する。
func TestMedicalRecordRepository_FindByAppointmentID_CorrelatesAppointmentClinic(t *testing.T) {
	db := setupMedicalRecordOwnerPetPreloadDB(t)
	require.NoError(t, db.Exec("TRUNCATE TABLE medical_records, appointments, reservation_types CASCADE").Error)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	makeAppt := func(clinicID uint64, name string) *model.Reservation {
		t.Helper()
		rt := &model.ReservationType{
			ClinicID: clinicID,
			Name:     name,
			Category: model.ReservationTypeCategoryGeneral,
			IsActive: true,
		}
		require.NoError(t, db.Create(rt).Error)
		start := time.Date(2026, 7, 28, 10, 0, 0, 0, time.UTC)
		appt := &model.Reservation{
			ClinicID:          clinicID,
			StartTime:         start,
			EndTime:           start.Add(30 * time.Minute),
			VisitType:         model.VisitTypeRevisit,
			ReservationTypeID: rt.ID,
			Status:            model.ReservationStatusConfirmed,
			Source:            model.ReservationSourceManual,
			CustomerFields:    []byte(`{}`),
		}
		require.NoError(t, db.Create(appt).Error)
		return appt
	}

	t.Run("rejects medical record linked to foreign-clinic appointment", func(t *testing.T) {
		apptB := makeAppt(clinicB, "B院予約")
		apptID := apptB.ID
		// 子は clinicA を名乗りつつ、親 appointment は clinicB — 旧実装は子 clinic だけで返す。
		polluted := &model.MedicalRecord{
			ClinicID:      clinicA,
			Date:          time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC),
			AppointmentID: &apptID,
			Status:        model.MedicalRecordStatusDraft,
			RecordNo:      "MR-B1-APPT-POLLUTED",
		}
		require.NoError(t, db.WithContext(ctx).Create(polluted).Error)

		got, err := repo.FindByAppointmentID(ctx, clinicA, apptB.ID)
		require.NoError(t, err)
		assert.Nil(t, got, "cross-tenant appointment parent must not yield a medical record")
	})

	t.Run("returns same-clinic appointment-linked medical record", func(t *testing.T) {
		apptA := makeAppt(clinicA, "A院予約")
		apptID := apptA.ID
		valid := &model.MedicalRecord{
			ClinicID:      clinicA,
			Date:          time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC),
			AppointmentID: &apptID,
			Status:        model.MedicalRecordStatusDraft,
			RecordNo:      "MR-B1-APPT-VALID",
		}
		require.NoError(t, db.WithContext(ctx).Create(valid).Error)

		got, err := repo.FindByAppointmentID(ctx, clinicA, apptA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, valid.ID, got.ID)
		assert.Equal(t, "MR-B1-APPT-VALID", got.RecordNo)
	})

	t.Run("soft-deleted same-clinic medical record stays excluded", func(t *testing.T) {
		apptA := makeAppt(clinicA, "A院予約削除")
		apptID := apptA.ID
		rec := &model.MedicalRecord{
			ClinicID:      clinicA,
			Date:          time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC),
			AppointmentID: &apptID,
			Status:        model.MedicalRecordStatusDraft,
			RecordNo:      "MR-B1-APPT-DELETED",
		}
		require.NoError(t, db.WithContext(ctx).Create(rec).Error)
		require.NoError(t, db.WithContext(ctx).Delete(rec).Error)

		got, err := repo.FindByAppointmentID(ctx, clinicA, apptA.ID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
