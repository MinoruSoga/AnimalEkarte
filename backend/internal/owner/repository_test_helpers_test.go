package owner

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestMain(m *testing.M) {
	code := m.Run()
	testdb.CloseSharedTestDB()
	os.Exit(code)
}

func setupOwnerPetIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Insurance{},
	))
	require.NoError(t, db.Exec("TRUNCATE TABLE insurances CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE animal_species CASCADE").Error)
	return db
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

func newTestRepository(db *gorm.DB) Repository {
	return NewRepository(db, &testPetRegistrar{db: db})
}

// testPetRegistrar is a target-package test double. Cross-domain integration
// with pet's real writer is covered by the owner/pet write-owner integration
// suite; this double keeps owner repository tests focused on the owner
// transaction and reload contract.
type testPetRegistrar struct {
	db *gorm.DB
}

func (r *testPetRegistrar) CreateForOwnerRegistration(
	ctx context.Context,
	intent PetRegistrationIntent,
) ([]model.Pet, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return nil, fmt.Errorf("ambient transaction is required")
	}
	created := make([]model.Pet, 0, len(intent.Pets))
	for i := range intent.Pets {
		draft := intent.Pets[i]
		pet := model.Pet{
			ClinicID:        intent.ClinicID,
			OwnerID:         intent.OwnerID,
			AnimalSpeciesID: draft.AnimalSpeciesID,
			PetNumber:       fmt.Sprintf("%d-%d", intent.OwnerID, i+1),
			Name:            draft.Name,
			NameKana:        draft.NameKana,
			Breed:           draft.Breed,
			Color:           draft.Color,
			BloodType:       draft.BloodType,
			MicrochipNumber: draft.MicrochipNumber,
			Gender:          draft.Gender,
			Status:          draft.Status,
			BirthDate:       draft.BirthDate,
			Weight:          draft.Weight,
			NeuteredDate:    draft.NeuteredDate,
			AcquisitionType: draft.AcquisitionType,
			DangerLevel:     draft.DangerLevel,
			Food:            draft.Food,
			Environment:     draft.Environment,
			Phone:           draft.Phone,
			InsuranceID:     draft.InsuranceID,
			Remarks:         draft.Remarks,
		}
		if err := tx.WithContext(ctx).Omit(clause.Associations).Create(&pet).Error; err != nil {
			return nil, err
		}
		created = append(created, pet)
	}
	return created, nil
}
