package repository

// owner_pet_relationship_preload_clinic_isolation_test.go verifies that corrupt
// owner_id relationships cannot make Owner/Pet preloads cross an unauthorized
// clinic boundary. The pets.owner_id FK is not a composite FK with clinic_id,
// so the database permits these intentionally malformed fixtures.

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestOwnerRepository_PetsOwnerPreload_HasClinicScopeContract(t *testing.T) {
	source, err := os.ReadFile("owner_repository.go")
	require.NoError(t, err)
	assert.Contains(
		t,
		string(source),
		`Preload("Pets.Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs)`,
		"Pets.Ownerはouter Owner/Petsの写像に依存せずclinic scopeを明示すべき",
	)
}

func makeOwnerPetRelationshipTestPet(
	t *testing.T,
	db *gorm.DB,
	clinicID, ownerID, speciesID uint64,
	name string,
) *model.Pet {
	t.Helper()
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: speciesID,
		Name:            name,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

func makeOwnerPetRelationshipTestSpecies(t *testing.T, db *gorm.DB) *model.AnimalSpecies {
	t.Helper()
	species := &model.AnimalSpecies{Name: "owner-pet preload isolation species"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	return species
}

func TestPetRepository_OwnerPreload_ClinicIsolation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "医院Aの飼主")
	ownerB := makeTestOwner(t, db, clinicB, "医院Bの飼主")
	species := makeOwnerPetRelationshipTestSpecies(t, db)
	legitimatePet := makeOwnerPetRelationshipTestPet(
		t, db, clinicA, ownerA.ID, species.ID, "正規関連ペット",
	)
	crossClinicPet := makeOwnerPetRelationshipTestPet(
		t, db, clinicA, ownerB.ID, species.ID, "越境関連ペット",
	)

	gotLegitimate, err := repo.FindByID(ctx, clinicA, legitimatePet.ID)
	require.NoError(t, err)
	require.NotNil(t, gotLegitimate.Owner)
	assert.Equal(t, ownerA.ID, gotLegitimate.Owner.ID)

	gotCrossClinic, err := repo.FindByID(ctx, clinicA, crossClinicPet.ID)
	require.NoError(t, err)
	assert.Nil(t, gotCrossClinic.Owner, "認可外clinicのOwnerをPet.Ownerへpreloadしてはならない")

	gotForAOnly, err := repo.FindByIDForClinics(ctx, []uint64{clinicA}, crossClinicPet.ID)
	require.NoError(t, err)
	assert.Nil(t, gotForAOnly.Owner, "単一clinicの認可集合でも他clinicのOwnerを隠すべき")

	gotForBoth, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, crossClinicPet.ID)
	require.NoError(t, err)
	require.NotNil(t, gotForBoth.Owner, "明示的に両clinicを認可した#86経路は既存semanticsを維持する")
	assert.Equal(t, ownerB.ID, gotForBoth.Owner.ID)
}

func TestOwnerRepository_PetsPreload_ClinicIsolation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "医院Aの飼主")
	ownerB := makeTestOwner(t, db, clinicB, "医院Bの飼主")
	species := makeOwnerPetRelationshipTestSpecies(t, db)
	legitimatePet := makeOwnerPetRelationshipTestPet(
		t, db, clinicA, ownerA.ID, species.ID, "正規clinicのペット",
	)
	crossClinicPet := makeOwnerPetRelationshipTestPet(
		t, db, clinicB, ownerA.ID, species.ID, "破損clinic関連のペット",
	)
	legitimatePetB := makeOwnerPetRelationshipTestPet(
		t, db, clinicB, ownerB.ID, species.ID, "医院Bの正規関連ペット",
	)

	owners, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, "")
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, owners, 1)
	require.Len(t, owners[0].Pets, 1, "FindAllは認可外clinicのPetをpreloadしてはならない")
	assert.Equal(t, legitimatePet.ID, owners[0].Pets[0].ID)

	ownersForBoth, totalForBoth, err := repo.FindAll(ctx, []uint64{clinicA, clinicB}, 1, 100, "")
	require.NoError(t, err)
	require.Equal(t, int64(2), totalForBoth)
	require.Len(t, ownersForBoth, 2)
	petsByOwnerID := make(map[uint64][]model.Pet, len(ownersForBoth))
	for i := range ownersForBoth {
		petsByOwnerID[ownersForBoth[i].ID] = ownersForBoth[i].Pets
	}
	require.Len(t, petsByOwnerID[ownerA.ID], 2)
	assert.ElementsMatch(t, []uint64{legitimatePet.ID, crossClinicPet.ID}, []uint64{
		petsByOwnerID[ownerA.ID][0].ID,
		petsByOwnerID[ownerA.ID][1].ID,
	})
	require.Len(t, petsByOwnerID[ownerB.ID], 1)
	assert.Equal(t, legitimatePetB.ID, petsByOwnerID[ownerB.ID][0].ID)

	got, err := repo.FindByID(ctx, clinicA, ownerA.ID)
	require.NoError(t, err)
	require.Len(t, got.Pets, 1, "FindByIDは認可外clinicのPetをpreloadしてはならない")
	assert.Equal(t, legitimatePet.ID, got.Pets[0].ID)
	require.NotNil(t, got.Pets[0].Owner)
	assert.Equal(t, clinicA, got.Pets[0].Owner.ClinicID)

	gotForBoth, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, ownerA.ID)
	require.NoError(t, err)
	require.Len(t, gotForBoth.Pets, 2, "明示的に両clinicを認可した#86経路は既存semanticsを維持する")
	assert.ElementsMatch(t, []uint64{legitimatePet.ID, crossClinicPet.ID}, []uint64{
		gotForBoth.Pets[0].ID,
		gotForBoth.Pets[1].ID,
	})
	for i := range gotForBoth.Pets {
		require.NotNil(t, gotForBoth.Pets[i].Owner)
		assert.Equal(t, clinicA, gotForBoth.Pets[i].Owner.ClinicID)
	}

	gotOwnerB, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, ownerB.ID)
	require.NoError(t, err)
	require.Len(t, gotOwnerB.Pets, 1)
	assert.Equal(t, legitimatePetB.ID, gotOwnerB.Pets[0].ID)
	require.NotNil(t, gotOwnerB.Pets[0].Owner, "認可集合の2番目のclinicでもPets.Ownerをpreloadすべき")
	assert.Equal(t, clinicB, gotOwnerB.Pets[0].Owner.ClinicID)
}
