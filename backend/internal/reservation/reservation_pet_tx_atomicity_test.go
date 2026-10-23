package reservation

// reservation_pet_tx_atomicity_test.go — ambient-tx participation for the reservation repository's
// pet reads used by the SD-10 deceased write guard.
//
//   - FindPetByIDInClinic uses persistence.DBOrTx and, when an ambient tx is present, takes a SHARE
//     row lock so the local pet row (owner link) cannot change between check and write. It no longer
//     reads pets.status / pets.deceased_at (retired columns).
//   - ReadPetLifeState (EMR-491 D6) is wired to the pet package's death reader: it requires the
//     ambient tx (fail closed otherwise) and resolves the life state from the common subject's current
//     basic revision, never from the local pet row. The reader's own lock / re-resolve behaviour is
//     proven in internal/pet; this file proves the production repository constructor wires it.
//
// temp-revert RED: change DBOrTx → r.db.WithContext(ctx) → SeesUncommittedOwnerUpdate fails
// (ambient UPDATE invisible) and/or SHARE lock no longer blocks concurrent writers.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupReservationPetTxAtomicityTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.CommonPetSubject{}, &model.CommonPetClinicBinding{}, &model.CommonPetBasicRevision{},
	))
	testdb.Truncate(t, db,
		"common_pet_basic_revisions", "common_pet_clinic_bindings", "common_pet_subjects",
		"pets", "animal_species", "owners",
	)
	return db
}

func makePetForFindPetTxTest(t *testing.T, db *gorm.DB, clinicID uint64, name string) (*model.Pet, *model.AnimalSpecies) {
	t.Helper()
	owner := testdb.MakeTestOwner(t, db, clinicID, name+"-owner")
	species := &model.AnimalSpecies{Name: name + "-species"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	pet := &model.Pet{
		ClinicID:  clinicID,
		OwnerID:   owner.ID,
		PetNumber: name + "-no",
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet, species
}

// linkLifeStatePet は pet を common subject + binding + 現行 revision（status / deceased_at 指定）に紐付ける。
func linkLifeStatePet(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	pet *model.Pet,
	speciesID uint64,
	status *model.PetStatus,
	deceasedAt *time.Time,
) {
	t.Helper()
	commonID := uuid.New()
	require.NoError(t, db.Create(&model.CommonPetSubject{ID: commonID}).Error)
	require.NoError(t, db.Create(&model.CommonPetClinicBinding{
		ClinicID: clinicID, CommonPetID: commonID, LocalPetID: pet.ID,
		ConfirmedBy: 1, ConfirmedAt: time.Now(), OperationID: 1,
	}).Error)
	revision := &model.CommonPetBasicRevision{
		CommonPetID: commonID, Revision: 1,
		SourceClinicID: clinicID, SourceLocalPetID: pet.ID, OperationID: 1,
		Name: "reader-wiring", AnimalSpeciesID: speciesID,
		Status: status, DeceasedAt: deceasedAt,
		NeuteredStatus: model.PetNeuteredStatusUnknown,
		Provenance:     json.RawMessage(`{}`),
		RecordedBy:     1, RecordedAt: time.Now(),
	}
	require.NoError(t, db.Create(revision).Error)
	require.NoError(t, db.Model(&model.CommonPetSubject{}).
		Where("id = ?", commonID).
		Update("current_basic_revision_id", revision.ID).Error)
}

// TestReservationRepository_FindPetByIDInClinic_SeesUncommittedOwnerUpdate
// proves an ambient UPDATE of the local pet row is visible to FindPetByIDInClinic
// and rolls back (would RED if the method used r.db instead of DBOrTx).
func TestReservationRepository_FindPetByIDInClinic_SeesUncommittedOwnerUpdate(t *testing.T) {
	db := setupReservationPetTxAtomicityTestDB(t)
	repo := NewReservationRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	pet, _ := makePetForFindPetTxTest(t, db, clinicA, "find-pet ambient")
	newOwner := testdb.MakeTestOwner(t, db, clinicA, "find-pet ambient new owner")
	forced := errors.New("forced find-pet ambient rollback")

	// Precondition: original owner outside ambient.
	got, err := repo.FindPetByIDInClinic(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	originalOwnerID := got.OwnerID
	require.NotEqual(t, newOwner.ID, originalOwnerID)

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := persistence.DBOrTx(txCtx, db).Model(&model.Pet{}).
			Where("id = ? AND clinic_id = ?", pet.ID, clinicA).
			Update("owner_id", newOwner.ID).Error; err != nil {
			return err
		}
		seen, err := repo.FindPetByIDInClinic(txCtx, clinicA, pet.ID)
		if err != nil {
			return err
		}
		if seen.OwnerID != newOwner.ID {
			return errors.New("FindPetByIDInClinic did not see uncommitted owner_id")
		}
		return forced
	})
	require.ErrorIs(t, txErr, forced)

	// Outside: still the original owner (rolled back).
	got, err = repo.FindPetByIDInClinic(ctx, clinicA, pet.ID)
	require.NoError(t, err)
	assert.Equal(t, originalOwnerID, got.OwnerID)
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

	pet, _ := makePetForFindPetTxTest(t, db, clinicA, "find-pet share-lock")

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
			Update("pet_number", "changed-while-locked").Error
		require.Error(t, updateErr,
			"concurrent UPDATE must wait/fail while FindPetByIDInClinic holds SHARE lock")
		return nil
	})
	require.NoError(t, err)

	// After ambient commit, UPDATE succeeds.
	require.NoError(t, db.WithContext(ctx).Model(&model.Pet{}).
		Where("id = ?", pet.ID).
		Update("pet_number", "changed-after-commit").Error)
}

// TestReservationRepository_ReadPetLifeState_UsesCommonRevisionInsideAmbientTx は本番 constructor が
// EMR-491 D6 の reader を配線していることを固定する（tx 外は fail closed・共通 revision から生死を解決）。
func TestReservationRepository_ReadPetLifeState_UsesCommonRevisionInsideAmbientTx(t *testing.T) {
	db := setupReservationPetTxAtomicityTestDB(t)
	repo := NewReservationRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	alive := model.PetStatusAlive
	deceased := model.PetStatusDeceased
	deceasedAt := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	alivePet, species := makePetForFindPetTxTest(t, db, clinicA, "life alive")
	linkLifeStatePet(t, db, clinicA, alivePet, species.ID, &alive, nil)
	deadPet, _ := makePetForFindPetTxTest(t, db, clinicA, "life dead")
	linkLifeStatePet(t, db, clinicA, deadPet, species.ID, &deceased, &deceasedAt)
	unknownPet, _ := makePetForFindPetTxTest(t, db, clinicA, "life unknown")
	linkLifeStatePet(t, db, clinicA, unknownPet, species.ID, nil, nil)
	unlinkedPet, _ := makePetForFindPetTxTest(t, db, clinicA, "life unlinked")

	t.Run("outside an ambient transaction fails closed", func(t *testing.T) {
		state, err := repo.ReadPetLifeState(ctx, clinicA, alivePet.ID)
		require.Error(t, err)
		assert.Empty(t, state)
	})

	t.Run("inside an ambient transaction", func(t *testing.T) {
		require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
			state, err := repo.ReadPetLifeState(txCtx, clinicA, alivePet.ID)
			require.NoError(t, err)
			assert.Equal(t, sharedkernel.PetLifeStateAlive, state)

			state, err = repo.ReadPetLifeState(txCtx, clinicA, deadPet.ID)
			require.NoError(t, err)
			assert.Equal(t, sharedkernel.PetLifeStateDeceased, state)

			state, err = repo.ReadPetLifeState(txCtx, clinicA, unknownPet.ID)
			require.ErrorIs(t, err, sharedkernel.ErrPetLifeStateUnresolved)
			assert.Empty(t, state)

			state, err = repo.ReadPetLifeState(txCtx, clinicA, unlinkedPet.ID)
			require.ErrorIs(t, err, sharedkernel.ErrPetLifeStateUnresolved)
			assert.Empty(t, state)
			return nil
		}))
	})

	t.Run("reservation guard rejects deceased and unknown, allows alive", func(t *testing.T) {
		require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
			aliveID, deadID, unknownID := alivePet.ID, deadPet.ID, unknownPet.ID
			require.NoError(t, ValidateReservationPetNotDeceased(txCtx, repo, clinicA, &aliveID))
			require.Error(t, ValidateReservationPetNotDeceased(txCtx, repo, clinicA, &deadID))
			require.Error(t, ValidateReservationPetNotDeceased(txCtx, repo, clinicA, &unknownID))
			return nil
		}))
	})
}
