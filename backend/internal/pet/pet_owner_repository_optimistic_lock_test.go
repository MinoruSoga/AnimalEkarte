package pet

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func loadPetOwnerState(
	t *testing.T,
	db *gorm.DB,
	clinicID, petID uint64,
) (model.Pet, []model.PetOwner) {
	t.Helper()

	var pet model.Pet
	require.NoError(t, db.Unscoped().
		Where("clinic_id = ? AND id = ?", clinicID, petID).
		First(&pet).Error)

	var links []model.PetOwner
	require.NoError(t, db.
		Where("clinic_id = ? AND pet_id = ?", clinicID, petID).
		Order("id ASC").
		Find(&links).Error)
	return pet, links
}

func petWithVersion(pet model.Pet, version int) model.Pet {
	pet.Version = version
	return pet
}

func TestPetOwnerRepository_ReplaceForPet_Version(t *testing.T) {
	t.Run("matching version replaces links and increments only version", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		repo := NewPetOwnerRepository(db)
		ctx := context.Background()
		const clinicID = uint64(1)

		primaryOwner := makeTestOwner(t, db, clinicID, "version一致主飼主")
		oldOwner := makeTestOwner(t, db, clinicID, "version一致旧副飼主")
		newOwner := makeTestOwner(t, db, clinicID, "version一致新副飼主")
		pet := makeSpeciesAndPet(t, db, clinicID, primaryOwner.ID, "version一致ペット")
		makePetOwnerLink(t, db, clinicID, pet.ID, oldOwner.ID, "旧")
		beforePet, _ := loadPetOwnerState(t, db, clinicID, pet.ID)

		err := repo.ReplaceForPet(ctx, clinicID, pet.ID, []model.PetOwner{{
			OwnerID:      newOwner.ID,
			Relationship: "新",
		}}, &beforePet.Version)
		require.NoError(t, err)

		afterPet, afterLinks := loadPetOwnerState(t, db, clinicID, pet.ID)
		assert.Equal(t, petWithVersion(beforePet, beforePet.Version+1), afterPet)
		require.Len(t, afterLinks, 1)
		assert.Equal(t, clinicID, afterLinks[0].ClinicID)
		assert.Equal(t, pet.ID, afterLinks[0].PetID)
		assert.Equal(t, newOwner.ID, afterLinks[0].OwnerID)
		assert.Equal(t, "新", afterLinks[0].Relationship)
	})

	t.Run("stale version conflicts without changing pet or links", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		repo := NewPetOwnerRepository(db)
		ctx := context.Background()
		const clinicID = uint64(1)

		primaryOwner := makeTestOwner(t, db, clinicID, "stale主飼主")
		oldOwner := makeTestOwner(t, db, clinicID, "stale旧副飼主")
		newOwner := makeTestOwner(t, db, clinicID, "stale新副飼主")
		pet := makeSpeciesAndPet(t, db, clinicID, primaryOwner.ID, "staleペット")
		makePetOwnerLink(t, db, clinicID, pet.ID, oldOwner.ID, "変更前")
		beforePet, beforeLinks := loadPetOwnerState(t, db, clinicID, pet.ID)
		staleVersion := beforePet.Version - 1

		err := repo.ReplaceForPet(ctx, clinicID, pet.ID, []model.PetOwner{{
			OwnerID:      newOwner.ID,
			Relationship: "反映禁止",
		}}, &staleVersion)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))

		afterPet, afterLinks := loadPetOwnerState(t, db, clinicID, pet.ID)
		assert.Equal(t, beforePet, afterPet)
		assert.Equal(t, beforeLinks, afterLinks)
	})

	t.Run("nil expected version preserves compatibility and increments version", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		repo := NewPetOwnerRepository(db)
		ctx := context.Background()
		const clinicID = uint64(1)

		primaryOwner := makeTestOwner(t, db, clinicID, "nil主飼主")
		oldOwner := makeTestOwner(t, db, clinicID, "nil旧副飼主")
		newOwner := makeTestOwner(t, db, clinicID, "nil新副飼主")
		pet := makeSpeciesAndPet(t, db, clinicID, primaryOwner.ID, "nilペット")
		require.NoError(t, db.Model(&model.Pet{}).
			Where("clinic_id = ? AND id = ?", clinicID, pet.ID).
			UpdateColumn("version", 7).Error)
		makePetOwnerLink(t, db, clinicID, pet.ID, oldOwner.ID, "旧")
		beforePet, _ := loadPetOwnerState(t, db, clinicID, pet.ID)

		err := repo.ReplaceForPet(ctx, clinicID, pet.ID, []model.PetOwner{{
			OwnerID:      newOwner.ID,
			Relationship: "新",
		}}, nil)
		require.NoError(t, err)

		afterPet, afterLinks := loadPetOwnerState(t, db, clinicID, pet.ID)
		assert.Equal(t, petWithVersion(beforePet, beforePet.Version+1), afterPet)
		require.Len(t, afterLinks, 1)
		assert.Equal(t, newOwner.ID, afterLinks[0].OwnerID)
	})

	t.Run("empty replacement increments version", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		repo := NewPetOwnerRepository(db)
		ctx := context.Background()
		const clinicID = uint64(1)

		owner := makeTestOwner(t, db, clinicID, "空置換飼主")
		pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "空置換ペット")
		makePetOwnerLink(t, db, clinicID, pet.ID, owner.ID, "解除対象")
		beforePet, _ := loadPetOwnerState(t, db, clinicID, pet.ID)

		err := repo.ReplaceForPet(ctx, clinicID, pet.ID, []model.PetOwner{}, &beforePet.Version)
		require.NoError(t, err)

		afterPet, afterLinks := loadPetOwnerState(t, db, clinicID, pet.ID)
		assert.Equal(t, petWithVersion(beforePet, beforePet.Version+1), afterPet)
		assert.Empty(t, afterLinks)
	})

	t.Run("cross clinic pet rejection leaves foreign version unchanged", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		repo := NewPetOwnerRepository(db)
		ctx := context.Background()
		const clinicA, clinicB = uint64(1), uint64(2)

		ownerA := makeTestOwner(t, db, clinicA, "越境元飼主")
		ownerB := makeTestOwner(t, db, clinicB, "越境先飼主")
		petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "越境先ペット")
		makePetOwnerLink(t, db, clinicB, petB.ID, ownerB.ID, "保護対象")
		beforePet, beforeLinks := loadPetOwnerState(t, db, clinicB, petB.ID)

		err := repo.ReplaceForPet(ctx, clinicA, petB.ID, []model.PetOwner{{
			OwnerID: ownerA.ID,
		}}, &beforePet.Version)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		afterPet, afterLinks := loadPetOwnerState(t, db, clinicB, petB.ID)
		assert.Equal(t, beforePet, afterPet)
		assert.Equal(t, beforeLinks, afterLinks)
	})

	t.Run("cross clinic owner rejection leaves own version unchanged", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		repo := NewPetOwnerRepository(db)
		ctx := context.Background()
		const clinicA, clinicB = uint64(1), uint64(2)

		ownerA := makeTestOwner(t, db, clinicA, "自院飼主")
		ownerB := makeTestOwner(t, db, clinicB, "他院飼主")
		petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "自院ペット")
		makePetOwnerLink(t, db, clinicA, petA.ID, ownerA.ID, "変更前")
		beforePet, beforeLinks := loadPetOwnerState(t, db, clinicA, petA.ID)

		err := repo.ReplaceForPet(ctx, clinicA, petA.ID, []model.PetOwner{{
			OwnerID: ownerB.ID,
		}}, &beforePet.Version)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		afterPet, afterLinks := loadPetOwnerState(t, db, clinicA, petA.ID)
		assert.Equal(t, beforePet, afterPet)
		assert.Equal(t, beforeLinks, afterLinks)
	})

	t.Run("ambient rollback restores links and version", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		repo := NewPetOwnerRepository(db)
		const clinicID = uint64(1)

		primaryOwner := makeTestOwner(t, db, clinicID, "ambient version主飼主")
		oldOwner := makeTestOwner(t, db, clinicID, "ambient version旧副飼主")
		newOwner := makeTestOwner(t, db, clinicID, "ambient version新副飼主")
		pet := makeSpeciesAndPet(t, db, clinicID, primaryOwner.ID, "ambient versionペット")
		makePetOwnerLink(t, db, clinicID, pet.ID, oldOwner.ID, "変更前")
		beforePet, beforeLinks := loadPetOwnerState(t, db, clinicID, pet.ID)

		tx, txCtx := beginAmbientPetOwnerTransaction(t, db)
		err := repo.ReplaceForPet(txCtx, clinicID, pet.ID, []model.PetOwner{{
			OwnerID:      newOwner.ID,
			Relationship: "rollback対象",
		}}, &beforePet.Version)
		require.NoError(t, err)
		inTxPet, inTxLinks := loadPetOwnerState(t, tx, clinicID, pet.ID)
		assert.Equal(t, beforePet.Version+1, inTxPet.Version)
		require.Len(t, inTxLinks, 1)
		assert.Equal(t, newOwner.ID, inTxLinks[0].OwnerID)
		require.NoError(t, tx.Rollback().Error)

		afterPet, afterLinks := loadPetOwnerState(t, db, clinicID, pet.ID)
		assert.Equal(t, beforePet, afterPet)
		assert.Equal(t, beforeLinks, afterLinks)
	})

	t.Run("insert failure rolls back links and version", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		repo := NewPetOwnerRepository(db)
		ctx := context.Background()
		const clinicID = uint64(1)

		owner := makeTestOwner(t, db, clinicID, "version rollback飼主")
		pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "version rollbackペット")
		makePetOwnerLink(t, db, clinicID, pet.ID, owner.ID, "変更前")
		beforePet, beforeLinks := loadPetOwnerState(t, db, clinicID, pet.ID)

		err := repo.ReplaceForPet(ctx, clinicID, pet.ID, []model.PetOwner{
			{OwnerID: owner.ID, Relationship: "重複1"},
			{OwnerID: owner.ID, Relationship: "重複2"},
		}, &beforePet.Version)
		require.Error(t, err)

		afterPet, afterLinks := loadPetOwnerState(t, db, clinicID, pet.ID)
		assert.Equal(t, beforePet, afterPet)
		assert.Equal(t, beforeLinks, afterLinks)
	})

	t.Run("concurrent matching versions yield one success and one conflict", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		repo := NewPetOwnerRepository(db)
		ctx := context.Background()
		const clinicID = uint64(1)

		primaryOwner := makeTestOwner(t, db, clinicID, "並行主飼主")
		oldOwner := makeTestOwner(t, db, clinicID, "並行旧副飼主")
		newOwnerA := makeTestOwner(t, db, clinicID, "並行新副飼主A")
		newOwnerB := makeTestOwner(t, db, clinicID, "並行新副飼主B")
		pet := makeSpeciesAndPet(t, db, clinicID, primaryOwner.ID, "並行ペット")
		makePetOwnerLink(t, db, clinicID, pet.ID, oldOwner.ID, "変更前")
		beforePet, _ := loadPetOwnerState(t, db, clinicID, pet.ID)

		errs := make(chan error, 2)
		for _, ownerID := range []uint64{newOwnerA.ID, newOwnerB.ID} {
			go func(replacementOwnerID uint64) {
				errs <- repo.ReplaceForPet(ctx, clinicID, pet.ID, []model.PetOwner{{
					OwnerID:      replacementOwnerID,
					Relationship: "並行更新",
				}}, &beforePet.Version)
			}(ownerID)
		}

		successes := 0
		conflicts := 0
		for range 2 {
			err := <-errs
			switch {
			case err == nil:
				successes++
			case apperrors.IsConflict(err):
				conflicts++
			default:
				t.Errorf("unexpected concurrent error: %v", err)
			}
		}
		assert.Equal(t, 1, successes)
		assert.Equal(t, 1, conflicts)

		afterPet, afterLinks := loadPetOwnerState(t, db, clinicID, pet.ID)
		assert.Equal(t, petWithVersion(beforePet, beforePet.Version+1), afterPet)
		require.Len(t, afterLinks, 1)
		assert.Contains(t, []uint64{newOwnerA.ID, newOwnerB.ID}, afterLinks[0].OwnerID)
	})
}
