package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestCheckupRepository_CreateAndRelationReadsJoinAmbientTx(t *testing.T) {
	fixture := setupClinicalRelationWriteFixture(t)
	repo := NewCheckupRepository(fixture.db)
	ctx := context.Background()
	sentinel := errors.New("rollback checkup create")

	err := fixture.writeTransactor.WithTx(ctx, func(txCtx context.Context) error {
		petID := fixture.petA.ID
		doctorID := fixture.assignedDoctor.ID
		checkup := &model.Checkup{
			ClinicID: fixture.clinicA, MedicalRecordID: fixture.recordA.ID,
			CheckupTypeID: fixture.checkupType.ID, PetID: &petID, DoctorID: &doctorID,
			Date: time.Now(),
		}
		require.NoError(t, repo.Create(txCtx, checkup))

		byID, findErr := repo.FindByID(txCtx, fixture.clinicA, checkup.ID)
		require.NoError(t, findErr)
		assert.Equal(t, checkup.ID, byID.ID)

		byRecord, findErr := repo.FindByMedicalRecordID(txCtx, fixture.clinicA, fixture.recordA.ID)
		require.NoError(t, findErr)
		require.Len(t, byRecord, 1)
		assert.Equal(t, checkup.ID, byRecord[0].ID)

		byOwner, findErr := repo.FindByOwnerID(txCtx, fixture.clinicA, *fixture.recordA.OwnerID)
		require.NoError(t, findErr)
		require.Len(t, byOwner, 1)
		assert.Equal(t, checkup.ID, byOwner[0].ID)

		byClinic, total, findErr := repo.FindByClinicID(txCtx, fixture.clinicA, CheckupFilters{}, 1, 100)
		require.NoError(t, findErr)
		require.EqualValues(t, 1, total)
		require.Len(t, byClinic, 1)
		assert.Equal(t, checkup.ID, byClinic[0].ID)
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)

	var count int64
	require.NoError(t, fixture.db.Model(&model.Checkup{}).
		Where("clinic_id = ?", fixture.clinicA).
		Count(&count).Error)
	assert.Zero(t, count, "ambient rollback must remove the uncommitted checkup")
}

func TestCheckupRepository_UpdateAndDeleteRollbackWithAmbientTx(t *testing.T) {
	fixture := setupClinicalRelationWriteFixture(t)
	repo := NewCheckupRepository(fixture.db)
	ctx := context.Background()
	petID := fixture.petA.ID
	doctorID := fixture.assignedDoctor.ID
	seed := &model.Checkup{
		ClinicID: fixture.clinicA, MedicalRecordID: fixture.recordA.ID,
		CheckupTypeID: fixture.checkupType.ID, PetID: &petID, DoctorID: &doctorID,
		Date: time.Now(), Result: "before",
	}
	require.NoError(t, repo.Create(ctx, seed))
	sentinel := errors.New("rollback checkup update and delete")

	err := fixture.writeTransactor.WithTx(ctx, func(txCtx context.Context) error {
		result := "inside transaction"
		require.NoError(t, repo.Update(txCtx, fixture.clinicA, seed.ID, UpdateCheckupInput{Result: &result}))
		updated, findErr := repo.FindByID(txCtx, fixture.clinicA, seed.ID)
		require.NoError(t, findErr)
		assert.Equal(t, "inside transaction", updated.Result)

		require.NoError(t, repo.Delete(txCtx, fixture.clinicA, seed.ID))
		_, findErr = repo.FindByID(txCtx, fixture.clinicA, seed.ID)
		require.Error(t, findErr)
		assert.True(t, apperrors.IsNotFound(findErr))
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)

	persisted, err := repo.FindByID(ctx, fixture.clinicA, seed.ID)
	require.NoError(t, err)
	assert.Equal(t, "before", persisted.Result)
}

func TestDB_CheckupRepository_LockedDeleteRollsBackWithAmbientTx(t *testing.T) {
	fixture := setupClinicalRelationWriteFixture(t)
	repo := NewCheckupRepository(fixture.db)
	ctx := context.Background()
	petID := fixture.petA.ID
	checkup := &model.Checkup{
		ClinicID: fixture.clinicA, MedicalRecordID: fixture.recordA.ID,
		CheckupTypeID: fixture.checkupType.ID, PetID: &petID, Date: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, checkup))
	sentinel := errors.New("rollback locked checkup delete")

	err := fixture.writeTransactor.WithTx(ctx, func(txCtx context.Context) error {
		locked, lockErr := repo.LockByIDForUpdate(txCtx, fixture.clinicA, checkup.ID)
		require.NoError(t, lockErr)
		assert.Equal(t, checkup.ID, locked.ID)
		require.NoError(t, repo.Delete(txCtx, fixture.clinicA, checkup.ID))
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)

	persisted, err := repo.FindByID(ctx, fixture.clinicA, checkup.ID)
	require.NoError(t, err)
	assert.Equal(t, checkup.ID, persisted.ID)
}

func TestDB_CheckupRepository_ParentThenChildLocksSerializeMedicalRecordFinalization(t *testing.T) {
	fixture := setupClinicalRelationWriteFixture(t)
	checkups := NewCheckupRepository(fixture.db)
	records := NewMedicalRecordRepository(fixture.db)
	ctx := context.Background()
	petID := fixture.petA.ID
	checkup := &model.Checkup{
		ClinicID: fixture.clinicA, MedicalRecordID: fixture.recordA.ID,
		CheckupTypeID: fixture.checkupType.ID, PetID: &petID, Date: time.Now(),
	}
	require.NoError(t, checkups.Create(ctx, checkup))

	err := fixture.writeTransactor.WithTx(ctx, func(txCtx context.Context) error {
		parent, lockErr := records.LockByIDForUpdate(txCtx, fixture.clinicA, fixture.recordA.ID)
		require.NoError(t, lockErr)
		assert.Equal(t, model.MedicalRecordStatusDraft, parent.Status)
		_, lockErr = checkups.LockByIDForUpdate(txCtx, fixture.clinicA, checkup.ID)
		require.NoError(t, lockErr)

		competingTx := fixture.db.WithContext(ctx).Begin()
		require.NoError(t, competingTx.Error)
		defer competingTx.Rollback()
		require.NoError(t, competingTx.Exec("SET LOCAL lock_timeout = '200ms'").Error)

		competingCtx := persistence.WithTxValue(ctx, competingTx)
		_, updateErr := records.Update(competingCtx, fixture.clinicA, fixture.recordA.ID, medicalRecordUpdateStatus(model.MedicalRecordStatusFinalized), nil)
		require.Error(t, updateErr, "finalization must wait while checkup deletion holds the parent lock")
		return nil
	})
	require.NoError(t, err)

	finalized, err := records.Update(ctx, fixture.clinicA, fixture.recordA.ID, medicalRecordUpdateStatus(model.MedicalRecordStatusFinalized), nil)
	require.NoError(t, err)
	assert.Equal(t, model.MedicalRecordStatusFinalized, finalized.Status)
}
