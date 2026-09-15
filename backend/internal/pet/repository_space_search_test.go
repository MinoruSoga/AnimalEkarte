package pet

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPetRepository_FindAll_PetNameIdeographicSpaceFourWay(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	tests := []struct {
		name       string
		storedName string
		query      string
	}{
		{name: "DB fullwidth × query fullwidth", storedName: "ペット全角全角　甲", query: "ペット全角全角　甲"},
		{name: "DB fullwidth × query halfwidth", storedName: "ペット全角半角　乙", query: "ペット全角半角 乙"},
		{name: "DB halfwidth × query fullwidth", storedName: "ペット半角全角 丙", query: "ペット半角全角　丙"},
		{name: "DB halfwidth × query halfwidth", storedName: "ペット半角半角 丁", query: "ペット半角半角 丁"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ownerA := makeTestOwner(t, db, clinicA, "pet-name-space-owner-a-"+tt.name)
			ownerB := makeTestOwner(t, db, clinicB, "pet-name-space-owner-b-"+tt.name)
			pet := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, tt.storedName)
			foreign := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, tt.storedName)

			got, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: tt.query}, 1, 10)
			require.NoError(t, err)
			require.Equal(t, int64(1), total)
			require.Len(t, got, 1)
			require.Equal(t, pet.ID, got[0].ID)
			require.NotEqual(t, foreign.ID, got[0].ID)
		})
	}
}

func TestPetRepository_FindAll_MultiTokenSearchAND(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerPochi := makeTestOwner(t, db, clinicA, "小林太郎")
	petPochi := makeSpeciesAndPet(t, db, clinicA, ownerPochi.ID, "ポチ")
	ownerShiro := makeTestOwner(t, db, clinicA, "小林")
	petShiro := makeSpeciesAndPet(t, db, clinicA, ownerShiro.ID, "シロ")
	ownerForeign := makeTestOwner(t, db, clinicB, "小林太郎")
	petForeign := makeSpeciesAndPet(t, db, clinicB, ownerForeign.ID, "ポチ")

	findIDs := func(t *testing.T, clinicIDs []uint64, search string) (ids []uint64, total int64) {
		t.Helper()
		pets, total, err := repo.FindAll(ctx, clinicIDs, PetListFilters{Search: search}, 1, 10)
		require.NoError(t, err)
		ids = make([]uint64, len(pets))
		for i, p := range pets {
			ids[i] = p.ID
		}
		return ids, total
	}

	t.Run("two-token half-width AND returns only the matching pet", func(t *testing.T) {
		ids, total := findIDs(t, []uint64{clinicA}, "小林 ポチ")
		require.Equal(t, int64(1), total)
		require.Equal(t, []uint64{petPochi.ID}, ids)
		assert.NotContains(t, ids, petShiro.ID)
		assert.NotContains(t, ids, petForeign.ID)
	})
	t.Run("two-token full-width space AND returns only the matching pet", func(t *testing.T) {
		ids, total := findIDs(t, []uint64{clinicA}, "小林　ポチ")
		require.Equal(t, int64(1), total)
		require.Equal(t, []uint64{petPochi.ID}, ids)
	})
	t.Run("two-token repeated whitespace AND returns only the matching pet", func(t *testing.T) {
		ids, total := findIDs(t, []uint64{clinicA}, "小林  ポチ")
		require.Equal(t, int64(1), total)
		require.Equal(t, []uint64{petPochi.ID}, ids)
	})
	t.Run("one-token surname OR returns both clinic rows", func(t *testing.T) {
		ids, total := findIDs(t, []uint64{clinicA}, "小林")
		require.Equal(t, int64(2), total)
		assert.Contains(t, ids, petPochi.ID)
		assert.Contains(t, ids, petShiro.ID)
		assert.NotContains(t, ids, petForeign.ID)
	})
	t.Run("same names in another clinic do not match", func(t *testing.T) {
		ids, total := findIDs(t, []uint64{clinicA}, "小林 ポチ")
		require.Equal(t, int64(1), total)
		assert.NotContains(t, ids, petForeign.ID)
		foreignIDs, foreignTotal := findIDs(t, []uint64{clinicB}, "小林 ポチ")
		require.Equal(t, int64(1), foreignTotal)
		require.Equal(t, []uint64{petForeign.ID}, foreignIDs)
		assert.NotContains(t, foreignIDs, petPochi.ID)
	})
}
