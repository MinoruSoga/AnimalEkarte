package medicalrecord

// vital_tx_atomicity_test.go — AUD-006:
// FindOrCreateByDate と CreateVitalRecord が同一 ambient tx に参加し、
// Vital 失敗時に新規 DailyRecord も rollback されることを検証する。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestDailyRecordRepository_AddVital_FindOrCreateAndCreateShareTx_Rollback(t *testing.T) {
	db := setupDailyRecordTestDB(t)
	repo := NewDailyRecordRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "飼主A-AUD006")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA-AUD006")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	date := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	err := withTx(ctx, db, func(txCtx context.Context) error {
		daily, err := repo.FindOrCreateByDate(txCtx, clinicA, hospA.ID, date)
		if err != nil {
			return err
		}
		vr := &model.VitalRecord{
			ClinicID:      clinicA,
			PetID:         petA.ID,
			DailyRecordID: &daily.ID,
			RecordedAt:    date.Add(9 * time.Hour),
		}
		if err := repo.CreateVitalRecord(txCtx, vr); err != nil {
			return err
		}
		// Vital 成功後に意図的失敗 → FindOrCreate で作った DailyRecord も rollback されること
		return errors.New("forced rollback after vital create")
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forced rollback")

	// rollback 後、当該日の DailyRecord は存在しない
	_, findErr := repo.FindByHospitalizationIDAndDate(ctx, clinicA, hospA.ID, date)
	require.Error(t, findErr)
	assert.True(t, true)
}

func TestDailyRecordRepository_AddVital_FindOrCreateAndCreateShareTx_Commit(t *testing.T) {
	db := setupDailyRecordTestDB(t)
	repo := NewDailyRecordRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "飼主A-AUD006b")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA-AUD006b")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	date := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)

	require.NoError(t, withTx(ctx, db, func(txCtx context.Context) error {
		daily, err := repo.FindOrCreateByDate(txCtx, clinicA, hospA.ID, date)
		if err != nil {
			return err
		}
		vr := &model.VitalRecord{
			ClinicID:      clinicA,
			PetID:         petA.ID,
			DailyRecordID: &daily.ID,
			RecordedAt:    date.Add(10 * time.Hour),
		}
		return repo.CreateVitalRecord(txCtx, vr)
	}))

	got, err := repo.FindByHospitalizationIDAndDate(ctx, clinicA, hospA.ID, date)
	require.NoError(t, err)
	require.Len(t, got.VitalRecords, 1)
	assert.Equal(t, petA.ID, got.VitalRecords[0].PetID)
}

func TestDailyRecordRepository_CreateCareLogParticipatesInAmbientTx(t *testing.T) {
	db := setupDailyRecordTestDB(t)
	repo := NewDailyRecordRepository(db)
	ctx := context.Background()

	const clinicID = uint64(1)
	owner := testdb.MakeTestOwner(t, db, clinicID, "飼主A-care-log-tx")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "ポチA-care-log-tx")
	hospitalization := makeHospitalizationRec(t, db, clinicID, owner.ID, pet.ID, nil)
	daily := makeDailyRecord(
		t,
		db,
		clinicID,
		hospitalization.ID,
		time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC),
	)

	err := withTx(ctx, db, func(txCtx context.Context) error {
		if err := repo.CreateCareLog(txCtx, &model.CareLog{
			ClinicID:      clinicID,
			DailyRecordID: daily.ID,
			Time:          "09:00:00",
			Type:          model.CareLogTypeFood,
		}); err != nil {
			return err
		}
		return errors.New("forced rollback after care log create")
	})
	require.EqualError(t, err, "forced rollback after care log create")

	var count int64
	require.NoError(t, db.Model(&model.CareLog{}).
		Where("daily_record_id = ?", daily.ID).
		Count(&count).Error)
	assert.Zero(t, count)
}

func TestDailyRecordRepository_CreateStaffNoteParticipatesInAmbientTx(t *testing.T) {
	db := setupDailyRecordTestDB(t)
	repo := NewDailyRecordRepository(db)
	ctx := context.Background()

	const clinicID = uint64(1)
	owner := testdb.MakeTestOwner(t, db, clinicID, "飼主A-staff-note-tx")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "ポチA-staff-note-tx")
	hospitalization := makeHospitalizationRec(t, db, clinicID, owner.ID, pet.ID, nil)
	daily := makeDailyRecord(
		t,
		db,
		clinicID,
		hospitalization.ID,
		time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC),
	)

	err := withTx(ctx, db, func(txCtx context.Context) error {
		if err := repo.CreateStaffNote(txCtx, &model.StaffNote{
			DailyRecordID: daily.ID,
			Time:          "10:00:00",
			Content:       "ambient tx rollback",
		}); err != nil {
			return err
		}
		return errors.New("forced rollback after staff note create")
	})
	require.EqualError(t, err, "forced rollback after staff note create")

	var count int64
	require.NoError(t, db.Model(&model.StaffNote{}).
		Where("daily_record_id = ?", daily.ID).
		Count(&count).Error)
	assert.Zero(t, count)
}
