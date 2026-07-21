package billing

// accounting_owner_pet_preload_clinic_isolation_test.go — AUD-002
// 汚染された Owner/Pet FK を持つ billing から、別 clinic の個人情報が Preload されないことを検証する。
// Preload から clinic_id 述語を外すと本テストは RED になる。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repotest"
)

func setupAccountingOwnerPetPreloadDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupAccountingIsolationTestDB(t)
	require.NoError(t, repotest.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}))
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func TestAccountingRepository_FindByID_FindAll_Update_DoesNotPreloadForeignOwnerPet(t *testing.T) {
	db := setupAccountingOwnerPetPreloadDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerB := repotest.MakeTestOwner(t, db, clinicB, "医院Bの飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "医院Bのペット")
	ownerBID, petBID := ownerB.ID, petB.ID

	contaminated := makeBillingWith(t, db, billingFixtureOpts{
		ClinicID:      clinicA,
		OwnerID:       &ownerBID,
		PetID:         &petBID,
		TotalAmount:   1500,
		Status:        model.BillingStatusWaiting,
		ScheduledDate: time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC),
	})

	t.Run("FindByID does not preload foreign owner/pet", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, contaminated.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, ownerBID, *got.OwnerID)
		assert.Equal(t, petBID, *got.PetID)
		assert.Nil(t, got.Owner, "foreign Owner must not be preloaded")
		assert.Nil(t, got.Pet, "foreign Pet must not be preloaded")
	})

	t.Run("FindAll does not preload foreign owner/pet", func(t *testing.T) {
		items, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, nil, 1, 50)
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(1))
		var found *model.Billing
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
		memo := "aud-002-preload"
		got, err := repo.Update(ctx, clinicA, contaminated.ID, map[string]any{"memo": memo})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, ownerBID, *got.OwnerID)
		assert.Equal(t, petBID, *got.PetID)
		assert.Nil(t, got.Owner, "foreign Owner must not be preloaded on Update refetch")
		assert.Nil(t, got.Pet, "foreign Pet must not be preloaded on Update refetch")
	})
}
