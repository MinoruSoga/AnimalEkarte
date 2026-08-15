package pet

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestPetRepository_FindAll_KanaNameSearch(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	petKatakanaKatakanaOwner := makeTestOwner(t, db, clinicID, "pet fixture owner 1")
	petKatakanaKatakana := makeSpeciesAndPet(t, db, clinicID, petKatakanaKatakanaOwner.ID, "アカリ")
	petKatakanaHiraganaOwner := makeTestOwner(t, db, clinicID, "pet fixture owner 2")
	petKatakanaHiragana := makeSpeciesAndPet(t, db, clinicID, petKatakanaHiraganaOwner.ID, "キナコ")
	petHiraganaKatakanaOwner := makeTestOwner(t, db, clinicID, "pet fixture owner 3")
	petHiraganaKatakana := makeSpeciesAndPet(t, db, clinicID, petHiraganaKatakanaOwner.ID, "くるみ")
	petHiraganaHiraganaOwner := makeTestOwner(t, db, clinicID, "pet fixture owner 4")
	petHiraganaHiragana := makeSpeciesAndPet(t, db, clinicID, petHiraganaHiraganaOwner.ID, "しずく")

	ownerKatakanaKatakana := makeTestOwner(t, db, clinicID, "ツバサ")
	ownerKatakanaKatakanaPet := makeSpeciesAndPet(t, db, clinicID, ownerKatakanaKatakana.ID, "owner fixture pet 1")
	ownerKatakanaHiragana := makeTestOwner(t, db, clinicID, "ナデシコ")
	ownerKatakanaHiraganaPet := makeSpeciesAndPet(t, db, clinicID, ownerKatakanaHiragana.ID, "owner fixture pet 2")
	ownerHiraganaKatakana := makeTestOwner(t, db, clinicID, "ほたる")
	ownerHiraganaKatakanaPet := makeSpeciesAndPet(t, db, clinicID, ownerHiraganaKatakana.ID, "owner fixture pet 3")
	ownerHiraganaHiragana := makeTestOwner(t, db, clinicID, "まどか")
	ownerHiraganaHiraganaPet := makeSpeciesAndPet(t, db, clinicID, ownerHiraganaHiragana.ID, "owner fixture pet 4")

	tests := []struct {
		name   string
		owner  *model.Owner
		pet    *model.Pet
		search string
	}{
		{
			name:   "pets.name katakana stored and katakana query",
			owner:  petKatakanaKatakanaOwner,
			pet:    petKatakanaKatakana,
			search: "アカリ",
		},
		{
			name:   "pets.name katakana stored and hiragana query",
			owner:  petKatakanaHiraganaOwner,
			pet:    petKatakanaHiragana,
			search: "きなこ",
		},
		{
			name:   "pets.name hiragana stored and katakana query",
			owner:  petHiraganaKatakanaOwner,
			pet:    petHiraganaKatakana,
			search: "クルミ",
		},
		{
			name:   "pets.name hiragana stored and hiragana query",
			owner:  petHiraganaHiraganaOwner,
			pet:    petHiraganaHiragana,
			search: "しずく",
		},
		{
			name:   "owners.name katakana stored and katakana query",
			owner:  ownerKatakanaKatakana,
			pet:    ownerKatakanaKatakanaPet,
			search: "ツバサ",
		},
		{
			name:   "owners.name katakana stored and hiragana query",
			owner:  ownerKatakanaHiragana,
			pet:    ownerKatakanaHiraganaPet,
			search: "なでしこ",
		},
		{
			name:   "owners.name hiragana stored and katakana query",
			owner:  ownerHiraganaKatakana,
			pet:    ownerHiraganaKatakanaPet,
			search: "ホタル",
		},
		{
			name:   "owners.name hiragana stored and hiragana query",
			owner:  ownerHiraganaHiragana,
			pet:    ownerHiraganaHiraganaPet,
			search: "まどか",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Empty(t, tt.owner.NameKana)
			require.Empty(t, tt.pet.NameKana)
			require.Nil(t, tt.pet.DeceasedAt)

			pets, total, err := repo.FindAll(
				ctx,
				[]uint64{clinicID},
				PetListFilters{Search: tt.search},
				1,
				10,
			)

			require.NoError(t, err)
			assert.Equal(t, int64(1), total)
			require.Len(t, pets, 1)
			assert.Equal(t, tt.pet.ID, pets[0].ID)
		})
	}
}
