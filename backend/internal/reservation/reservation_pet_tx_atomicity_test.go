package reservation

// reservation_pet_tx_atomicity_test.go — ambient-tx participation for
// reservationRepository.FindPetByIDInClinic (SD-10 deceased write guard).
//
// FindPetByIDInClinic uses persistence.DBOrTx and, when ambient tx is present,
// takes a SHARE row lock so deceased_at cannot change between check and write.
//
// temp-revert RED: change DBOrTx → r.db.WithContext(ctx) → SeesUncommitted fails
// (ambient UPDATE invisible) and/or SHARE lock no longer blocks concurrent writers.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupReservationPetTxAtomicityTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
	))
	require.NoError(t, db.Exec(`
		TRUNCATE TABLE pets, animal_species, owners CASCADE
	`).Error)
	return db
}

func makePetForFindPetTxTest(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Pet {
	t.Helper()
	owner := testdb.MakeTestOwner(t, db, clinicID, name+"-owner")
	species := &model.AnimalSpecies{Name: name + "-species"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         owner.ID,
		AnimalSpeciesID: species.ID,
		Name:            name,
		Status:          model.PetStatusAlive,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

// TestReservationRepository_FindPetByIDInClinic_SeesUncommittedDeceasedUpdate
// proves ambient UPDATE of deceased_at/status is visible to FindPetByIDInClinic
// and rolls back (would RED if the method used r.db instead of DBOrTx).
func TestReservationRepository_FindPetByIDInClinic_SeesUncommittedDeceasedUpdate(t *testing.T) {
	db := setupReservationPetTxAtomicityTestDB(t)
	repo := NewReservationRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	pet := makePetForFindPetTxTest(t, db, clinicA, "find-pet ambient")
	forced := errors.New("forced find-pet ambient rollback")
	deceasedAt := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	// Precondition: alive outside ambient.
	got, err := repo.FindPetByIDInClinic(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusAlive, got.Status)
	assert.Nil(t, got.DeceasedAt)

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := persistence.DBOrTx(txCtx, db).Model(&model.Pet{}).
			Where("id = ? AND clinic_id = ?", pet.ID, clinicA).
			Updates(map[string]any{
				"status":      model.PetStatusDeceased,
				"deceased_at": deceasedAt,
			}).Error; err != nil {
			return err
		}
		seen, err := repo.FindPetByIDInClinic(txCtx, clinicA, pet.ID)
		if err != nil {
			return err
		}
		if seen.Status != model.PetStatusDeceased {
			return errors.New("FindPetByIDInClinic did not see uncommitted status")
		}
		if seen.DeceasedAt == nil || !seen.DeceasedAt.Equal(deceasedAt) {
			return errors.New("FindPetByIDInClinic did not see uncommitted deceased_at")
		}
		return forced
	})
	require.ErrorIs(t, txErr, forced)

	// Outside: still alive (rolled back).
	got, err = repo.FindPetByIDInClinic(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PetStatusAlive, got.Status)
	assert.Nil(t, got.DeceasedAt)
}

// TestReservationRepository_FindPetByIDInClinic_ShareLockBlocksConcurrentWriter
// proves that while ambient tx holds the SHARE lock from FindPetByIDInClinic,
// a concurrent UPDATE on the same pet row times out (lock_timeout).
func TestReservationRepository_FindPetByIDInClinic_ShareLockBlocksConcurrentWriter(t *testing.T) {
	db := setupReservationPetTxAtomicityTestDB(t)
	repo := NewReservationRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	pet := makePetForFindPetTxTest(t, db, clinicA, "find-pet share-lock")

	err := tx.WithTx(ctx, func(txCtx context.Context) error {
		// Acquire SHARE lock via FindPetByIDInClinic under ambient tx.
		_, findErr := repo.FindPetByIDInClinic(txCtx, clinicA, pet.ID)
		require.NoError(t, findErr)

		competingTx := db.WithContext(ctx).Begin()
		require.NoError(t, competingTx.Error)
		defer competingTx.Rollback()
		require.NoError(t, competingTx.Exec("SET LOCAL lock_timeout = '200ms'").Error)

		updateErr := competingTx.Model(&model.Pet{}).
			Where("id = ?", pet.ID).
			Update("status", model.PetStatusDeceased).Error
		require.Error(t, updateErr,
			"concurrent UPDATE must wait/fail while FindPetByIDInClinic holds SHARE lock")
		return nil
	})
	require.NoError(t, err)

	// After ambient commit, UPDATE succeeds.
	require.NoError(t, db.WithContext(ctx).Model(&model.Pet{}).
		Where("id = ?", pet.ID).
		Update("status", model.PetStatusDeceased).Error)
}
