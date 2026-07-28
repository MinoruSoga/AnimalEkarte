package owner_test

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	ownerdomain "github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/persistence"
	petdomain "github.com/animal-ekarte/backend/internal/pet"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type reportingOwnerDeleteLocker struct {
	delegate ownerdomain.OwnerDeleteLocker
	db       *gorm.DB
	pid      chan<- int
}

func (l reportingOwnerDeleteLocker) LockForDelete(
	ctx context.Context,
	clinicID, ownerID uint64,
) (*model.Owner, error) {
	var pid int
	if err := persistence.DBOrTx(ctx, l.db).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
		return nil, err
	}
	select {
	case l.pid <- pid:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return l.delegate.LockForDelete(ctx, clinicID, ownerID)
}

type pausingPetOwnerCounter struct {
	delegate ownerdomain.PetOwnerCounter
	counted  chan<- struct{}
	release  <-chan struct{}
}

func (c pausingPetOwnerCounter) CountByOwnerID(
	ctx context.Context,
	clinicID, ownerID uint64,
) (int64, error) {
	count, err := c.delegate.CountByOwnerID(ctx, clinicID, ownerID)
	if err != nil {
		return 0, err
	}
	select {
	case c.counted <- struct{}{}:
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	select {
	case <-c.release:
		return count, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

type failingPetOwnerCounter struct {
	err error
}

func (c failingPetOwnerCounter) CountByOwnerID(
	context.Context,
	uint64,
	uint64,
) (int64, error) {
	return 0, c.err
}

func setupOwnerDeletePetOwnerGuardTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupOwnerCreateWithPetsTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.PetOwner{}))
	require.NoError(t, db.Exec("TRUNCATE TABLE pet_owners RESTART IDENTITY CASCADE").Error)
	return db
}

func makeOwnerDeleteGuardPet(
	t *testing.T,
	db *gorm.DB,
	clinicID, primaryOwnerID uint64,
	name string,
) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: name + "種"}
	require.NoError(t, db.Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         primaryOwnerID,
		AnimalSpeciesID: species.ID,
		Name:            name,
	}
	require.NoError(t, db.Create(pet).Error)
	return pet
}

func newGuardedOwnerService(
	db *gorm.DB,
	repo ownerdomain.Repository,
	locker ownerdomain.OwnerDeleteLocker,
	counter ownerdomain.PetOwnerCounter,
) ownerdomain.Service {
	return ownerdomain.NewServiceWithPetOwnerDeleteGuard(
		repo,
		nil,
		nil,
		nil,
		locker,
		counter,
		persistence.NewTransactor(db),
	)
}

func requireOwnerDeleteInvariant(
	t *testing.T,
	db *gorm.DB,
	clinicID, ownerID uint64,
) (model.Owner, int64) {
	t.Helper()
	var owner model.Owner
	require.NoError(t, db.Unscoped().
		Where("clinic_id = ? AND id = ?", clinicID, ownerID).
		First(&owner).Error)
	var linkCount int64
	require.NoError(t, db.Model(&model.PetOwner{}).
		Where("clinic_id = ? AND owner_id = ?", clinicID, ownerID).
		Count(&linkCount).Error)
	assert.False(t, owner.DeletedAt.Valid && linkCount > 0,
		"deleted owner must never retain a committed pet_owners row")
	return owner, linkCount
}

func requireBackendPIDWaiting(
	t *testing.T,
	ctx context.Context,
	db *gorm.DB,
	pid int,
	operationDone <-chan error,
) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-operationDone:
			require.Failf(t, "operation completed before entering the expected lock wait", "err=%v", err)
		case <-ticker.C:
			var waiting bool
			require.NoError(t, db.WithContext(ctx).
				Raw("SELECT EXISTS (SELECT 1 FROM pg_locks WHERE pid = ? AND NOT granted)", pid).
				Scan(&waiting).Error)
			if waiting {
				return
			}
		case <-ctx.Done():
			require.Failf(t, "operation did not enter a database lock wait", "pid=%d err=%v", pid, ctx.Err())
		}
	}
}

func receiveOwnerDeleteGuardResult(
	t *testing.T,
	ctx context.Context,
	result <-chan error,
) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		require.Failf(t, "operation did not complete before deadline", "err=%v", ctx.Err())
		return ctx.Err()
	}
}

func cleanupOwnerDeleteGuardTransaction(t *testing.T, tx *gorm.DB) {
	t.Helper()
	if err := tx.Rollback().Error; err != nil && !errors.Is(err, sql.ErrTxDone) {
		t.Errorf("cleanup rollback failed: %v", err)
	}
}

func TestOwnerRepository_DeleteAndCountPets_AmbientTransaction(t *testing.T) {
	t.Run("Delete rollback leaves owner active", func(t *testing.T) {
		db := setupOwnerDeletePetOwnerGuardTestDB(t)
		repo := newOwnerRepository(db)
		ctx := context.Background()
		const clinicID = uint64(1)
		owner := testdb.MakeTestOwner(t, db, clinicID, "ambient delete rollback")

		tx := db.WithContext(ctx).Begin()
		require.NoError(t, tx.Error)
		t.Cleanup(func() { cleanupOwnerDeleteGuardTransaction(t, tx) })
		txCtx := persistence.WithTxValue(ctx, tx)

		require.NoError(t, repo.Delete(txCtx, clinicID, owner.ID))
		var deletedInsideTx model.Owner
		require.NoError(t, tx.Unscoped().First(&deletedInsideTx, owner.ID).Error)
		assert.True(t, deletedInsideTx.DeletedAt.Valid)
		require.NoError(t, tx.Rollback().Error)

		reloaded, err := repo.FindByID(ctx, clinicID, owner.ID)
		require.NoError(t, err)
		assert.False(t, reloaded.DeletedAt.Valid)
	})

	t.Run("CountPets observes uncommitted insert and rollback removes it", func(t *testing.T) {
		db := setupOwnerDeletePetOwnerGuardTestDB(t)
		repo := newOwnerRepository(db)
		ctx := context.Background()
		const clinicID = uint64(1)
		owner := testdb.MakeTestOwner(t, db, clinicID, "ambient pet count")
		species := &model.AnimalSpecies{Name: "ambient pet count species"}
		require.NoError(t, db.Create(species).Error)

		tx := db.WithContext(ctx).Begin()
		require.NoError(t, tx.Error)
		t.Cleanup(func() { cleanupOwnerDeleteGuardTransaction(t, tx) })
		txCtx := persistence.WithTxValue(ctx, tx)
		pet := &model.Pet{
			ClinicID:        clinicID,
			OwnerID:         owner.ID,
			AnimalSpeciesID: species.ID,
			Name:            "uncommitted pet",
		}
		require.NoError(t, persistence.DBOrTx(txCtx, db).Create(pet).Error)

		count, err := repo.CountPetsByOwnerID(txCtx, clinicID, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
		require.NoError(t, tx.Rollback().Error)

		var committedCount int64
		require.NoError(t, db.Model(&model.Pet{}).
			Where("id = ?", pet.ID).
			Count(&committedCount).Error)
		assert.Zero(t, committedCount)
	})
}

func TestOwnerService_Delete_PetOwnerGuards(t *testing.T) {
	t.Run("linked secondary owner returns Conflict and remains active", func(t *testing.T) {
		db := setupOwnerDeletePetOwnerGuardTestDB(t)
		repo := newOwnerRepository(db)
		petOwnerRepo := petdomain.NewPetOwnerRepository(db)
		const clinicID = uint64(1)
		primaryOwner := testdb.MakeTestOwner(t, db, clinicID, "secondary guard primary")
		secondaryOwner := testdb.MakeTestOwner(t, db, clinicID, "secondary guard target")
		pet := makeOwnerDeleteGuardPet(t, db, clinicID, primaryOwner.ID, "secondary guard pet")
		require.NoError(t, petOwnerRepo.ReplaceForPet(context.Background(), clinicID, pet.ID, []model.PetOwner{{
			OwnerID:      secondaryOwner.ID,
			Relationship: "家族",
		}}, nil))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := newGuardedOwnerService(db, repo, repo, petOwnerRepo).Delete(ctx, clinicID, secondaryOwner.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Contains(t, err.Error(), "副飼主として紐付いているため削除できません。先に紐付けを解除してください")

		owner, linkCount := requireOwnerDeleteInvariant(t, db, clinicID, secondaryOwner.ID)
		assert.False(t, owner.DeletedAt.Valid)
		assert.Equal(t, int64(1), linkCount)
	})

	t.Run("owner without dependencies is soft deleted", func(t *testing.T) {
		db := setupOwnerDeletePetOwnerGuardTestDB(t)
		repo := newOwnerRepository(db)
		petOwnerRepo := petdomain.NewPetOwnerRepository(db)
		const clinicID = uint64(1)
		target := testdb.MakeTestOwner(t, db, clinicID, "unlinked delete target")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, newGuardedOwnerService(db, repo, repo, petOwnerRepo).
			Delete(ctx, clinicID, target.ID))

		_, err := repo.FindByID(context.Background(), clinicID, target.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "expected NotFound, got %v", err)
		owner, linkCount := requireOwnerDeleteInvariant(t, db, clinicID, target.ID)
		assert.True(t, owner.DeletedAt.Valid)
		assert.Zero(t, linkCount)
	})

	t.Run("primary pet dependency remains Conflict", func(t *testing.T) {
		db := setupOwnerDeletePetOwnerGuardTestDB(t)
		repo := newOwnerRepository(db)
		petOwnerRepo := petdomain.NewPetOwnerRepository(db)
		const clinicID = uint64(1)
		target := testdb.MakeTestOwner(t, db, clinicID, "primary dependency target")
		makeOwnerDeleteGuardPet(t, db, clinicID, target.ID, "primary dependency pet")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := newGuardedOwnerService(db, repo, repo, petOwnerRepo).Delete(ctx, clinicID, target.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Contains(t, err.Error(), "ペットが紐付いているため削除できません。先にペットを削除してください")

		owner, linkCount := requireOwnerDeleteInvariant(t, db, clinicID, target.ID)
		assert.False(t, owner.DeletedAt.Valid)
		assert.Zero(t, linkCount)
	})

	t.Run("cross clinic request is NotFound without disclosing dependency", func(t *testing.T) {
		db := setupOwnerDeletePetOwnerGuardTestDB(t)
		repo := newOwnerRepository(db)
		petOwnerRepo := petdomain.NewPetOwnerRepository(db)
		const clinicA, clinicB = uint64(1), uint64(2)
		primaryOwnerB := testdb.MakeTestOwner(t, db, clinicB, "clinic B primary")
		targetB := testdb.MakeTestOwner(t, db, clinicB, "clinic B secondary target")
		petB := makeOwnerDeleteGuardPet(t, db, clinicB, primaryOwnerB.ID, "clinic B pet")
		require.NoError(t, petOwnerRepo.ReplaceForPet(context.Background(), clinicB, petB.ID, []model.PetOwner{{
			OwnerID:      targetB.ID,
			Relationship: "医院B家族",
		}}, nil))

		err := newGuardedOwnerService(db, repo, repo, petOwnerRepo).
			Delete(context.Background(), clinicA, targetB.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "expected NotFound, got %v", err)
		assert.False(t, apperrors.IsConflict(err), "cross-clinic dependency must not be disclosed")

		owner, linkCount := requireOwnerDeleteInvariant(t, db, clinicB, targetB.ID)
		assert.False(t, owner.DeletedAt.Valid)
		assert.Equal(t, int64(1), linkCount)
	})
}

func TestOwnerService_Delete_PetOwnerCounterErrorFailsClosed(t *testing.T) {
	db := setupOwnerDeletePetOwnerGuardTestDB(t)
	repo := newOwnerRepository(db)
	const clinicID = uint64(1)
	target := testdb.MakeTestOwner(t, db, clinicID, "counter error target")
	counterErr := errors.New("pet owner count unavailable")

	err := newGuardedOwnerService(db, repo, repo, failingPetOwnerCounter{err: counterErr}).
		Delete(context.Background(), clinicID, target.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, counterErr)

	owner, linkCount := requireOwnerDeleteInvariant(t, db, clinicID, target.ID)
	assert.False(t, owner.DeletedAt.Valid)
	assert.Zero(t, linkCount)
}

func TestOwnerService_Delete_ConcurrentPetOwnerReplacementSerializes(t *testing.T) {
	t.Run("link commits first and delete returns Conflict", func(t *testing.T) {
		db := setupOwnerDeletePetOwnerGuardTestDB(t)
		repo := newOwnerRepository(db)
		petOwnerRepo := petdomain.NewPetOwnerRepository(db)
		const clinicID = uint64(1)
		primaryOwner := testdb.MakeTestOwner(t, db, clinicID, "link first primary")
		target := testdb.MakeTestOwner(t, db, clinicID, "link first target")
		pet := makeOwnerDeleteGuardPet(t, db, clinicID, primaryOwner.ID, "link first pet")

		linkTx := db.WithContext(context.Background()).Begin()
		require.NoError(t, linkTx.Error)
		t.Cleanup(func() { cleanupOwnerDeleteGuardTransaction(t, linkTx) })
		linkCtx := persistence.WithTxValue(context.Background(), linkTx)
		require.NoError(t, petOwnerRepo.ReplaceForPet(linkCtx, clinicID, pet.ID, []model.PetOwner{{
			OwnerID:      target.ID,
			Relationship: "link first",
		}}, nil))

		deletePID := make(chan int, 1)
		reportingLocker := reportingOwnerDeleteLocker{delegate: repo, db: db, pid: deletePID}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		deleteDone := make(chan error, 1)
		go func() {
			deleteDone <- newGuardedOwnerService(db, repo, reportingLocker, petOwnerRepo).
				Delete(ctx, clinicID, target.ID)
		}()

		select {
		case pid := <-deletePID:
			waitCtx, cancelWait := context.WithTimeout(context.Background(), time.Second)
			requireBackendPIDWaiting(t, waitCtx, db, pid, deleteDone)
			cancelWait()
		case err := <-deleteDone:
			require.Failf(t, "delete completed before lock acquisition was reported", "err=%v", err)
		case <-ctx.Done():
			require.Failf(t, "delete did not attempt owner lock", "err=%v", ctx.Err())
		}

		require.NoError(t, linkTx.Commit().Error)
		deleteErr := receiveOwnerDeleteGuardResult(t, ctx, deleteDone)
		require.Error(t, deleteErr)
		assert.True(t, apperrors.IsConflict(deleteErr), "expected Conflict, got %v", deleteErr)

		owner, linkCount := requireOwnerDeleteInvariant(t, db, clinicID, target.ID)
		assert.False(t, owner.DeletedAt.Valid)
		assert.Equal(t, int64(1), linkCount)
	})

	t.Run("delete commits first and link returns NotFound", func(t *testing.T) {
		db := setupOwnerDeletePetOwnerGuardTestDB(t)
		repo := newOwnerRepository(db)
		petOwnerRepo := petdomain.NewPetOwnerRepository(db)
		const clinicID = uint64(1)
		primaryOwner := testdb.MakeTestOwner(t, db, clinicID, "delete first primary")
		target := testdb.MakeTestOwner(t, db, clinicID, "delete first target")
		pet := makeOwnerDeleteGuardPet(t, db, clinicID, primaryOwner.ID, "delete first pet")

		counted := make(chan struct{}, 1)
		releaseDelete := make(chan struct{})
		var releaseDeleteOnce sync.Once
		releaseDeleteBarrier := func() {
			releaseDeleteOnce.Do(func() { close(releaseDelete) })
		}
		t.Cleanup(releaseDeleteBarrier)
		counter := pausingPetOwnerCounter{
			delegate: petOwnerRepo,
			counted:  counted,
			release:  releaseDelete,
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		deleteDone := make(chan error, 1)
		go func() {
			deleteDone <- newGuardedOwnerService(db, repo, repo, counter).
				Delete(ctx, clinicID, target.ID)
		}()

		select {
		case <-counted:
		case err := <-deleteDone:
			require.Failf(t, "delete completed before secondary count barrier", "err=%v", err)
		case <-ctx.Done():
			require.Failf(t, "delete did not reach secondary count barrier", "err=%v", ctx.Err())
		}

		linkTx := db.WithContext(context.Background()).Begin()
		require.NoError(t, linkTx.Error)
		t.Cleanup(func() { cleanupOwnerDeleteGuardTransaction(t, linkTx) })
		var linkPID int
		require.NoError(t, linkTx.Raw("SELECT pg_backend_pid()").Scan(&linkPID).Error)
		linkCtx := persistence.WithTxValue(ctx, linkTx)
		linkDone := make(chan error, 1)
		go func() {
			linkDone <- petOwnerRepo.ReplaceForPet(linkCtx, clinicID, pet.ID, []model.PetOwner{{
				OwnerID:      target.ID,
				Relationship: "delete first",
			}}, nil)
		}()

		waitCtx, cancelWait := context.WithTimeout(context.Background(), time.Second)
		requireBackendPIDWaiting(t, waitCtx, db, linkPID, linkDone)
		cancelWait()
		releaseDeleteBarrier()

		require.NoError(t, receiveOwnerDeleteGuardResult(t, ctx, deleteDone))
		linkErr := receiveOwnerDeleteGuardResult(t, ctx, linkDone)
		require.Error(t, linkErr)
		assert.True(t, apperrors.IsNotFound(linkErr), "expected NotFound, got %v", linkErr)
		require.NoError(t, linkTx.Rollback().Error)

		owner, linkCount := requireOwnerDeleteInvariant(t, db, clinicID, target.ID)
		assert.True(t, owner.DeletedAt.Valid)
		assert.Zero(t, linkCount)
	})
}

func TestOwnerRepository_LockForDelete_RequiresAmbientTransaction(t *testing.T) {
	db := setupOwnerDeletePetOwnerGuardTestDB(t)
	repo := newOwnerRepository(db)
	target := testdb.MakeTestOwner(t, db, 1, "missing transaction target")

	locked, err := repo.LockForDelete(context.Background(), 1, target.ID)
	require.Error(t, err)
	assert.Nil(t, locked)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr, "expected AppError, got %v", err)
	assert.Equal(t, "INTERNAL", appErr.Code)
}
