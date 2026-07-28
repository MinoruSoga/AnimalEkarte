package pet

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestMain(m *testing.M) {
	code := m.Run()
	testdb.CloseSharedTestDB()
	os.Exit(code)
}

func makeTestOwner(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	name string,
) *model.Owner {
	t.Helper()
	return testdb.MakeTestOwner(t, db, clinicID, name)
}

func makeSpeciesAndPet(
	t *testing.T,
	db *gorm.DB,
	clinicID, ownerID uint64,
	petName string,
) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)

	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: species.ID,
		Name:            petName,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}
