package pet

// realdb_selected_clinic_b_grant_a_isolation_test.go — D3 pet package (4 routes)
//
// Proves clinic-fixed pet nested list/detail isolation through real repository +
// service + HTTP handler paths. Offline `go test -short` SKIPs via testdb.SetupTestDB.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
	ownerdomain "github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	realDBPetSeedNameA            = "D3pet-realdb-pet-A"
	realDBPetSeedNameB            = "D3pet-realdb-pet-B"
	realDBOwnerPrimarySeedNameA   = "D3pet-realdb-primary-A"
	realDBOwnerPrimarySeedNameB   = "D3pet-realdb-primary-B"
	realDBOwnerSecondarySeedNameA = "D3pet-realdb-secondary-A"
	realDBOwnerSecondarySeedNameB = "D3pet-realdb-secondary-B"
	realDBConditionSeedNameA      = "D3pet-realdb-condition-A"
	realDBConditionSeedNameB      = "D3pet-realdb-condition-B"
	realDBFirstVisitDateB         = "2026-03-15"
)

type petQueryGuardService struct {
	Service
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *petQueryGuardService) GetFirstVisitDate(ctx context.Context, clinicID, petID uint64) (*time.Time, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("GetFirstVisitDate must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.Service.GetFirstVisitDate(ctx, clinicID, petID)
}

type chronicQueryGuardService struct {
	ChronicConditionService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *chronicQueryGuardService) List(ctx context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("chronic List must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.ChronicConditionService.List(ctx, clinicID, petID)
}

type petOwnerQueryGuardService struct {
	PetOwnerService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *petOwnerQueryGuardService) GetByPetID(ctx context.Context, clinicID, petID uint64) ([]model.PetOwner, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("GetByPetID must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.PetOwnerService.GetByPetID(ctx, clinicID, petID)
}

func (s *petOwnerQueryGuardService) GetSharedPetsByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]SharedPet, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("GetSharedPetsByOwnerID must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.PetOwnerService.GetSharedPetsByOwnerID(ctx, clinicID, ownerID)
}

type realDBPetFixture struct {
	fx            testdb.ClinicGrantFixture
	petA          *model.Pet
	petB          *model.Pet
	primaryOwnerA *model.Owner
	primaryOwnerB *model.Owner
	secondaryA    *model.Owner
	secondaryB    *model.Owner
	handler       *Handler
}

func setupRealDBPetIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.PetOwner{},
		&model.PetChronicCondition{},
		&model.MedicalRecord{},
	))
	testdb.Truncate(t, db,
		"pet_owners",
		"pet_chronic_conditions",
		"medical_records",
		"pets",
		"owners",
		"animal_species",
		"staff_clinic_assignments",
		"staffs",
	)
	return db
}

func newRealDBPetHandler(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) *Handler {
	t.Helper()
	petRepo := NewRepository(db)
	ownerRepo := ownerdomain.NewRepository(db, nil)
	mrRepo := medicalrecord.NewMedicalRecordRepository(db)
	chronicRepo := NewChronicConditionRepository(db)
	petOwnerRepo := NewPetOwnerRepository(db)

	petSvc := Service(NewService(petRepo, ownerRepo, nil, mrRepo, nil))
	chronicSvc := ChronicConditionService(NewChronicConditionService(chronicRepo, petRepo, nil))
	petOwnerSvc := PetOwnerService(NewPetOwnerService(petRepo, ownerRepo, petOwnerRepo, nil, nil))

	if forbiddenClinicID != 0 {
		petSvc = &petQueryGuardService{Service: petSvc, t: t, forbiddenClinicID: forbiddenClinicID}
		chronicSvc = &chronicQueryGuardService{ChronicConditionService: chronicSvc, t: t, forbiddenClinicID: forbiddenClinicID}
		petOwnerSvc = &petOwnerQueryGuardService{PetOwnerService: petOwnerSvc, t: t, forbiddenClinicID: forbiddenClinicID}
	}

	return NewHandlerWithPetOwners(petSvc, nil, chronicSvc, petOwnerSvc, ownerRepo, nil)
}

func seedRealDBPetFixture(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) realDBPetFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3pet realDB")
	ctx := context.Background()

	primaryOwnerA := testdb.MakeTestOwner(t, db, fx.ClinicA, realDBOwnerPrimarySeedNameA)
	primaryOwnerB := testdb.MakeTestOwner(t, db, fx.ClinicB, realDBOwnerPrimarySeedNameB)
	secondaryA := testdb.MakeTestOwner(t, db, fx.ClinicA, realDBOwnerSecondarySeedNameA)
	secondaryB := testdb.MakeTestOwner(t, db, fx.ClinicB, realDBOwnerSecondarySeedNameB)

	petA := testdb.MakeSpeciesAndPet(t, db, fx.ClinicA, primaryOwnerA.ID, realDBPetSeedNameA)
	petB := testdb.MakeSpeciesAndPet(t, db, fx.ClinicB, primaryOwnerB.ID, realDBPetSeedNameB)

	require.NoError(t, db.WithContext(ctx).Create(&model.PetOwner{
		ClinicID: fx.ClinicA, PetID: petA.ID, OwnerID: secondaryA.ID, Relationship: "家族A",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.PetOwner{
		ClinicID: fx.ClinicB, PetID: petB.ID, OwnerID: secondaryB.ID, Relationship: "家族B",
	}).Error)

	diagnosed := time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)
	require.NoError(t, db.WithContext(ctx).Create(&model.PetChronicCondition{
		ClinicID: fx.ClinicA, PetID: petA.ID, ConditionCode: "CA", ConditionName: realDBConditionSeedNameA,
		DiagnosedAt: diagnosed, IsActive: true,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.PetChronicCondition{
		ClinicID: fx.ClinicB, PetID: petB.ID, ConditionCode: "CB", ConditionName: realDBConditionSeedNameB,
		DiagnosedAt: diagnosed, IsActive: true,
	}).Error)

	visitB, err := time.ParseInLocation(time.DateOnly, realDBFirstVisitDateB, time.Local)
	require.NoError(t, err)
	testdb.MakeHistoryMedicalRecord(t, db, fx.ClinicA, petA.ID, "D3PET-A-1", visitB.AddDate(-1, 0, 0))
	testdb.MakeHistoryMedicalRecord(t, db, fx.ClinicB, petB.ID, "D3PET-B-1", visitB)

	return realDBPetFixture{
		fx: fx, petA: petA, petB: petB,
		primaryOwnerA: primaryOwnerA, primaryOwnerB: primaryOwnerB,
		secondaryA: secondaryA, secondaryB: secondaryB,
		handler: newRealDBPetHandler(t, db, forbiddenClinicID),
	}
}

func configurePetGrant(fx realDBPetFixture, resource string, grantClinicID uint64) func(*gin.Context) {
	return testdb.ConfigureSelectedClinicBGrant(fx.fx, resource, grantClinicID)
}

func withPetIDParam(configure func(*gin.Context), id uint64) func(*gin.Context) {
	return func(c *gin.Context) {
		configure(c)
		c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", id)}}
	}
}

func TestRealDB_SelectedClinicBGrantAIsolation_Pet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("shared_pets_grantA_403", func(t *testing.T) {
		db := setupRealDBPetIsolationTestDB(t)
		fx := seedRealDBPetFixture(t, db, 0)
		fx.handler = newRealDBPetHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/owners/%d/shared-pets", fx.secondaryB.ID),
			withPetIDParam(configurePetGrant(fx, string(model.ResourceOwners), fx.fx.ClinicA), fx.secondaryB.ID))
		fx.handler.ListOwnerSharedPets(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), nil, realDBPetSeedNameA, realDBPetSeedNameB)
	})

	t.Run("shared_pets_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBPetIsolationTestDB(t)
		fx := seedRealDBPetFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/owners/%d/shared-pets", fx.secondaryB.ID),
			withPetIDParam(configurePetGrant(fx, string(model.ResourceOwners), fx.fx.ClinicB), fx.secondaryB.ID))
		fx.handler.ListOwnerSharedPets(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got ownerSharedPetsResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.NotEmpty(t, got.SharedPets)
		foundB := false
		for _, item := range got.SharedPets {
			assert.NotEqual(t, realDBPetSeedNameA, item.Name)
			if item.ID == fx.petB.ID {
				foundB = true
				assert.Equal(t, realDBPetSeedNameB, item.Name)
			}
		}
		require.True(t, foundB)
	})

	t.Run("shared_pets_A_owner_404", func(t *testing.T) {
		db := setupRealDBPetIsolationTestDB(t)
		fx := seedRealDBPetFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/owners/%d/shared-pets", fx.secondaryA.ID),
			withPetIDParam(configurePetGrant(fx, string(model.ResourceOwners), fx.fx.ClinicB), fx.secondaryA.ID))
		fx.handler.ListOwnerSharedPets(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), nil, realDBPetSeedNameA, realDBPetSeedNameB)
	})

	t.Run("sub_owners_grantA_403", func(t *testing.T) {
		db := setupRealDBPetIsolationTestDB(t)
		fx := seedRealDBPetFixture(t, db, 0)
		fx.handler = newRealDBPetHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/pets/%d/sub-owners", fx.petB.ID),
			withPetIDParam(configurePetGrant(fx, string(model.ResourceOwners), fx.fx.ClinicA), fx.petB.ID))
		fx.handler.ListPetOwners(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), nil, realDBOwnerSecondarySeedNameA, realDBOwnerSecondarySeedNameB)
	})

	t.Run("sub_owners_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBPetIsolationTestDB(t)
		fx := seedRealDBPetFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/pets/%d/sub-owners", fx.petB.ID),
			withPetIDParam(configurePetGrant(fx, string(model.ResourceOwners), fx.fx.ClinicB), fx.petB.ID))
		fx.handler.ListPetOwners(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got petOwnersResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.NotEmpty(t, got.SubOwners)
		foundB := false
		for _, item := range got.SubOwners {
			assert.NotEqual(t, realDBOwnerSecondarySeedNameA, item.Name)
			if item.OwnerID == fx.secondaryB.ID {
				foundB = true
				assert.Equal(t, realDBOwnerSecondarySeedNameB, item.Name)
			}
		}
		require.True(t, foundB)
	})

	t.Run("sub_owners_A_pet_404", func(t *testing.T) {
		db := setupRealDBPetIsolationTestDB(t)
		fx := seedRealDBPetFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/pets/%d/sub-owners", fx.petA.ID),
			withPetIDParam(configurePetGrant(fx, string(model.ResourceOwners), fx.fx.ClinicB), fx.petA.ID))
		fx.handler.ListPetOwners(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), nil, realDBOwnerSecondarySeedNameA, realDBOwnerSecondarySeedNameB)
	})

	t.Run("chronic_grantA_403", func(t *testing.T) {
		db := setupRealDBPetIsolationTestDB(t)
		fx := seedRealDBPetFixture(t, db, 0)
		fx.handler = newRealDBPetHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/pets/%d/chronic-conditions", fx.petB.ID),
			withPetIDParam(configurePetGrant(fx, string(model.ResourceOwners), fx.fx.ClinicA), fx.petB.ID))
		fx.handler.ListChronicConditions(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBConditionSeedNameA, realDBConditionSeedNameB)
	})

	t.Run("chronic_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBPetIsolationTestDB(t)
		fx := seedRealDBPetFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/pets/%d/chronic-conditions", fx.petB.ID),
			withPetIDParam(configurePetGrant(fx, string(model.ResourceOwners), fx.fx.ClinicB), fx.petB.ID))
		fx.handler.ListChronicConditions(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []chronicConditionResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBConditionSeedNameA, item.ConditionName)
			if item.ConditionName == realDBConditionSeedNameB {
				foundB = true
			}
		}
		require.True(t, foundB)
	})

	t.Run("chronic_A_pet_404", func(t *testing.T) {
		db := setupRealDBPetIsolationTestDB(t)
		fx := seedRealDBPetFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/pets/%d/chronic-conditions", fx.petA.ID),
			withPetIDParam(configurePetGrant(fx, string(model.ResourceOwners), fx.fx.ClinicB), fx.petA.ID))
		fx.handler.ListChronicConditions(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBConditionSeedNameA)
	})

	t.Run("first_visit_grantA_403", func(t *testing.T) {
		db := setupRealDBPetIsolationTestDB(t)
		fx := seedRealDBPetFixture(t, db, 0)
		fx.handler = newRealDBPetHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/pets/%d/first-visit", fx.petB.ID),
			withPetIDParam(configurePetGrant(fx, string(model.ResourceMedicalRecords), fx.fx.ClinicA), fx.petB.ID))
		fx.handler.GetPetFirstVisit(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), realDBFirstVisitDateB)
	})

	t.Run("first_visit_grantB_ok", func(t *testing.T) {
		db := setupRealDBPetIsolationTestDB(t)
		fx := seedRealDBPetFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/pets/%d/first-visit", fx.petB.ID),
			withPetIDParam(configurePetGrant(fx, string(model.ResourceMedicalRecords), fx.fx.ClinicB), fx.petB.ID))
		fx.handler.GetPetFirstVisit(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBFirstVisitDateB)
		assert.NotContains(t, w.Body.String(), `"first_visit_date":null`)
	})

	t.Run("first_visit_A_pet_404", func(t *testing.T) {
		db := setupRealDBPetIsolationTestDB(t)
		fx := seedRealDBPetFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/pets/%d/first-visit", fx.petA.ID),
			withPetIDParam(configurePetGrant(fx, string(model.ResourceMedicalRecords), fx.fx.ClinicB), fx.petA.ID))
		fx.handler.GetPetFirstVisit(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		assert.NotContains(t, w.Body.String(), realDBFirstVisitDateB)
	})
}
