package pet

import (
	"context"
	"testing"

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
