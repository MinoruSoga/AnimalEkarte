package medicalrecord

// hospitalization_owner_pet_preload_clinic_isolation_test.go — AUD-004
// 汚染された Owner/Pet FK を持つ入院から、別 clinic の個人情報が Preload されないことを検証する。
//
// SEC-SWEEP-01（c16c011f2）で FindAll に親相関述語
// EXISTS (SELECT 1 FROM pets p WHERE p.id = hospitalizations.pet_id AND p.clinic_id = hospitalizations.clinic_id)
// を追加したため、一覧 read の契約は「汚染行を sanitize して返す」から「汚染行を返さない」へ強化された。
// 詳細 read（FindByID / Update refetch）は行自体を返しつつ他院 Owner/Pet の Preload を抑止する
// 従来契約を維持する。billing 側（0736cd6f9）と同じ非対称契約である。

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

func TestHospitalizationRepository_FindByID_Update_DoesNotPreloadForeignOwnerPet(t *testing.T) {
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

	t.Run("FindAll excludes hospitalization whose pet belongs to another clinic", func(t *testing.T) {
		items, _, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, nil, 1, 50)
		require.NoError(t, err)
		var found *model.Hospitalization
		for i := range items {
			if items[i].ID == contaminated.ID {
				found = &items[i]
				break
			}
		}
		assert.Nil(t, found, "list must exclude a hospitalization whose pet belongs to another clinic")
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
