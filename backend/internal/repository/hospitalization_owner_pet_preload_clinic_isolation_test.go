package repository

// hospitalization_owner_pet_preload_clinic_isolation_test.go — AUD-004
// 汚染された Owner/Pet FK を持つ入院から、別 clinic の個人情報が Preload されないことを検証する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

func setupHospitalizationOwnerPetPreloadDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupHospitalizationRepoTestDB(t)
	db.Exec("TRUNCATE TABLE owners CASCADE")
	return db
}

func TestHospitalizationRepository_FindByID_FindAll_Update_DoesNotPreloadForeignOwnerPet(t *testing.T) {
	db := setupHospitalizationOwnerPetPreloadDB(t)
	repo := NewHospitalizationRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerB := makeTestOwner(t, db, clinicB, "医院Bの飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "医院Bのペット")
	contaminated := makeHospitalizationFixture(t, db, clinicA, ownerB.ID, petB.ID, nil)

	t.Run("FindByID does not preload foreign owner/pet", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, contaminated.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, ownerB.ID, got.OwnerID)
		assert.Equal(t, petB.ID, got.PetID)
		assert.Nil(t, got.Owner, "foreign Owner must not be preloaded")
		assert.Nil(t, got.Pet, "foreign Pet must not be preloaded")
	})

	t.Run("FindAll does not preload foreign owner/pet", func(t *testing.T) {
		items, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, nil, 1, 50)
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(1))
		var found *model.Hospitalization
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

	t.Run("Update refetch does not preload foreign owner/pet", func(t *testing.T) {
		memo := "aud-004-preload"
		got, err := repo.Update(ctx, clinicA, contaminated.ID, map[string]any{"memo": memo})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, ownerB.ID, got.OwnerID)
		assert.Equal(t, petB.ID, got.PetID)
		assert.Nil(t, got.Owner, "foreign Owner must not be preloaded")
		assert.Nil(t, got.Pet, "foreign Pet must not be preloaded")
	})
}
