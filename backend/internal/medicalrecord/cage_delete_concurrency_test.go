package medicalrecord

// cage_delete_concurrency_test.go — SEC-CS-F13: cage soft-delete serializes with
// hospitalization cage assignment (FOR UPDATE vs ambient FOR SHARE + usage re-check).

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
	"github.com/animal-ekarte/backend/internal/persistence"
)

var errForceRollback = errors.New("force rollback")

// TestCageService_Delete_ConcurrentAssignFirstYieldsConflict proves assign-first:
// hospitalization holds FOR SHARE on cage (via ambient FindByID), concurrent Delete waits,
// then after assign commits CountUsage sees the hospitalization and returns Conflict.
func TestCageService_Delete_ConcurrentAssignFirstYieldsConflict(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := NewCageRepository(db)
	svc := NewCageService(repo, persistence.NewTransactor(db))
	ctx := context.Background()
	const clinicID = uint64(1)

	cage := makeCageMaster(t, db, clinicID, "assign-first cage")
	owner := makeTestOwner(t, db, clinicID, "飼主 assign-first")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "ポチ")

	// Hold FOR SHARE on cage (same path as hospitalization cage FK validation) and insert hospitalization.
	locked := make(chan struct{})
	release := make(chan struct{})
	holderDone := make(chan error, 1)
	go func() {
		holderDone <- db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := persistence.WithTxValue(ctx, tx)
			if _, err := repo.FindByID(txCtx, clinicID, cage.ID); err != nil {
				return err
			}
			now := time.Now().UTC().Truncate(24 * time.Hour)
			if err := tx.Create(&model.Hospitalization{
				ClinicID:            clinicID,
				OwnerID:             owner.ID,
				PetID:               pet.ID,
				HospitalizationType: model.HospitalizationTypeInpatient,
				StartDate:           now,
				EndDate:             now.AddDate(0, 0, 3),
				Status:              model.HospitalizationStatusAdmitted,
				CageID:              &cage.ID,
			}).Error; err != nil {
				return err
			}
			close(locked)
			<-release
			return nil
		})
	}()
	<-locked

	deleteDone := make(chan error, 1)
	go func() {
		deleteDone <- svc.Delete(ctx, clinicID, cage.ID)
	}()

	// Soft-delete exclusive lock must wait behind the assigner's FOR SHARE.
	select {
	case err := <-deleteDone:
		close(release)
		require.Failf(t, "cage delete was not serialized behind hospitalization assign share-lock", "err=%v", err)
	case <-time.After(100 * time.Millisecond):
		// still waiting — good
	}

	close(release)
	require.NoError(t, <-holderDone)

	deleteErr := <-deleteDone
	require.Error(t, deleteErr)
	assert.True(t, apperrors.IsConflict(deleteErr), "assign-first must yield Conflict after serialization, got %v", deleteErr)
	assert.False(t, apperrors.IsNotFound(deleteErr))

	// Cage row must remain (delete rolled back / never committed).
	got, err := repo.FindByID(ctx, clinicID, cage.ID)
	require.NoError(t, err)
	assert.Equal(t, cage.ID, got.ID)
}

// TestCageService_Delete_ConcurrentDeleteFirstRejectsLaterAssign proves delete-first:
// atomic Delete soft-deletes with zero usage; later ambient FindByID rejects soft-deleted cage.
func TestCageService_Delete_ConcurrentDeleteFirstRejectsLaterAssign(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := NewCageRepository(db)
	svc := NewCageService(repo, persistence.NewTransactor(db))
	ctx := context.Background()
	const clinicID = uint64(1)

	cage := makeCageMaster(t, db, clinicID, "delete-first cage")
	require.NoError(t, svc.Delete(ctx, clinicID, cage.ID))

	// Later assign path: ambient FindByID (FOR SHARE) must NotFound soft-deleted cage.
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		_, findErr := repo.FindByID(txCtx, clinicID, cage.ID)
		return findErr
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "delete-first must make later assign reject inactive cage as NotFound")
}

// TestCageRepository_CountUsage_AmbientTxSeesUncommittedHospitalization proves CountUsage
// joins ambient tx via DBOrTx and observes uncommitted hospitalization rows.
func TestCageRepository_CountUsage_AmbientTxSeesUncommittedHospitalization(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := NewCageRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	cage := makeCageMaster(t, db, clinicID, "ambient usage cage")
	owner := makeTestOwner(t, db, clinicID, "飼主 ambient")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "タマ")

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		now := time.Now().UTC().Truncate(24 * time.Hour)
		require.NoError(t, tx.Create(&model.Hospitalization{
			ClinicID:            clinicID,
			OwnerID:             owner.ID,
			PetID:               pet.ID,
			HospitalizationType: model.HospitalizationTypeInpatient,
			StartDate:           now,
			EndDate:             now.AddDate(0, 0, 1),
			Status:              model.HospitalizationStatusAdmitted,
			CageID:              &cage.ID,
		}).Error)

		count, countErr := repo.CountUsageByCageID(txCtx, clinicID, cage.ID)
		require.NoError(t, countErr)
		assert.Equal(t, int64(1), count, "ambient CountUsage must see uncommitted hospitalization")
		return errForceRollback
	})
	require.ErrorIs(t, err, errForceRollback)

	count, err := repo.CountUsageByCageID(ctx, clinicID, cage.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "rolled-back hospitalization must not count")
}

// TestCageRepository_LockByIDForUpdate_RequiresAmbientTransaction fail-closes without ambient tx.
func TestCageRepository_LockByIDForUpdate_RequiresAmbientTransaction(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := NewCageRepository(db)
	cage := makeCageMaster(t, db, 1, "lock ambient required")

	_, err := repo.LockByIDForUpdate(context.Background(), 1, cage.ID)
	require.Error(t, err)
	assert.False(t, apperrors.IsNotFound(err))
	assert.False(t, apperrors.IsConflict(err))
}
