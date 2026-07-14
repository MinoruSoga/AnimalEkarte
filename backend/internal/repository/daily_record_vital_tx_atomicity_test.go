package repository

// daily_record_vital_tx_atomicity_test.go — AUD-006:
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
)

func TestDailyRecordRepository_AddVital_FindOrCreateAndCreateShareTx_Rollback(t *testing.T) {
	db := setupDailyRecordTestDB(t)
	repo := NewDailyRecordRepository(db)
	tx := NewTransactor(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	ownerA := makeTestOwner(t, db, clinicA, "飼主A-AUD006")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA-AUD006")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	date := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	err := tx.WithTx(ctx, func(txCtx context.Context) error {
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
	tx := NewTransactor(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	ownerA := makeTestOwner(t, db, clinicA, "飼主A-AUD006b")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA-AUD006b")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	date := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
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
