package pet

// realdb_selected_clinic_b_grant_a_isolation_test.go — D3 pet cross-clinic (3 routes)
//
// Covers GET /pets, GET /pets/:id, GET /owners/:id/report/pets only.
// Clinic-fixed pet routes are owned by a separate agent and must not be edited here.
// Offline `go test -short` SKIPs via testdb.SetupTestDB.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	realDBCrossPetSeedNameA  = "D3xcl-realdb-pet-A"
	realDBCrossPetSeedNameB  = "D3xcl-realdb-pet-B"
	realDBCrossPetOwnerNameA = "D3xcl-realdb-pet-owner-A"
	realDBCrossPetOwnerNameB = "D3xcl-realdb-pet-owner-B"
)

func setupRealDBCrossClinicPetTestDB(t *testing.T) *gorm.DB {
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
	))
	testdb.Truncate(t, db,
		"pets",
		"animal_species",
		"owners",
		"staff_clinic_assignments",
		"staffs",
	)
	return db
}

type petCrossClinicQueryGuardService struct {
	Service
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *petCrossClinicQueryGuardService) assertAllowed(clinicIDs []uint64, op string) {
	s.t.Helper()
	for _, id := range clinicIDs {
		if id == s.forbiddenClinicID {
			s.t.Fatalf("%s must not query clinic %d without grant", op, id)
		}
	}
}

func (s *petCrossClinicQueryGuardService) List(
	ctx context.Context,
	clinicIDs []uint64,
	filters PetListFilters,
	page, limit int,
) ([]model.Pet, int64, error) {
	s.assertAllowed(clinicIDs, "List")
	return s.Service.List(ctx, clinicIDs, filters, page, limit)
}

func (s *petCrossClinicQueryGuardService) GetByIDForClinics(
	ctx context.Context,
	clinicIDs []uint64,
	id uint64,
) (*model.Pet, error) {
	s.assertAllowed(clinicIDs, "GetByIDForClinics")
	return s.Service.GetByIDForClinics(ctx, clinicIDs, id)
}

func (s *petCrossClinicQueryGuardService) ListOwnerReportPets(
	ctx context.Context,
	clinicIDs []uint64,
	ownerID uint64,
) ([]model.Pet, error) {
	s.assertAllowed(clinicIDs, "ListOwnerReportPets")
	return s.Service.ListOwnerReportPets(ctx, clinicIDs, ownerID)
}

type realDBCrossClinicPetFixture struct {
	fx      testdb.ClinicGrantFixture
	ownerA  *model.Owner
	ownerB  *model.Owner
	petA    *model.Pet
	petB    *model.Pet
	handler *Handler
}

func newRealDBCrossClinicPetHandler(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) *Handler {
	t.Helper()
	var svc Service = NewService(NewRepository(db), nil, nil, nil, nil)
	if forbiddenClinicID != 0 {
		svc = &petCrossClinicQueryGuardService{
			Service:           svc,
			t:                 t,
			forbiddenClinicID: forbiddenClinicID,
		}
	}
	return NewHandler(svc, nil, nil, nil)
}

func seedRealDBCrossClinicPetFixture(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) realDBCrossClinicPetFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3xcl pet realDB")
	ownerA := testdb.MakeTestOwner(t, db, fx.ClinicA, realDBCrossPetOwnerNameA)
	ownerB := testdb.MakeTestOwner(t, db, fx.ClinicB, realDBCrossPetOwnerNameB)
	species := &model.AnimalSpecies{Name: "D3xcl-realdb-species", IsActive: true}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	petA := &model.Pet{
		ClinicID: fx.ClinicA, OwnerID: ownerA.ID, AnimalSpeciesID: species.ID, Name: realDBCrossPetSeedNameA,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(petA).Error)
	petB := &model.Pet{
		ClinicID: fx.ClinicB, OwnerID: ownerB.ID, AnimalSpeciesID: species.ID, Name: realDBCrossPetSeedNameB,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(petB).Error)
	return realDBCrossClinicPetFixture{
		fx:      fx,
		ownerA:  ownerA,
		ownerB:  ownerB,
		petA:    petA,
		petB:    petB,
		handler: newRealDBCrossClinicPetHandler(t, db, forbiddenClinicID),
	}
}

func TestRealDB_Pets_List_MembershipABGrantASelectedB_Default_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBCrossClinicPetTestDB(t)
	fx := seedRealDBCrossClinicPetFixture(t, db, 0)
	fx.handler = newRealDBCrossClinicPetHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		"/api/v1/pets?page=1&limit=50",
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicA),
	)
	fx.handler.ListPets(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	testdb.AssertBodyOmitsClinicArtifacts(
		t,
		w.Body.Bytes(),
		[]uint64{fx.fx.ClinicA, fx.fx.ClinicB},
		realDBCrossPetSeedNameA,
		realDBCrossPetSeedNameB,
	)
}

func TestRealDB_Pets_List_MembershipABGrantASelectedB_MixedClinicIDs_ReturnsAOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBCrossClinicPetTestDB(t)
	fx := seedRealDBCrossClinicPetFixture(t, db, 0)
	fx.handler = newRealDBCrossClinicPetHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/pets?page=1&limit=50&clinic_ids=%d,%d", fx.fx.ClinicA, fx.fx.ClinicB),
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicA),
	)
	fx.handler.ListPets(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var listed httpapi.PaginatedResponse[[]petListResponse]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.NotEmpty(t, listed.Data, "A seed list must be nonempty; empty list is not observation")
	foundA := false
	for _, item := range listed.Data {
		assert.Equal(t, fx.fx.ClinicA, item.ClinicID)
		assert.NotEqual(t, fx.petB.ID, item.ID)
		assert.NotEqual(t, realDBCrossPetSeedNameB, item.Name)
		if item.ID == fx.petA.ID {
			foundA = true
			assert.Equal(t, realDBCrossPetSeedNameA, item.Name)
		}
	}
	require.True(t, foundA, "A seed id/name/clinic_id must appear in filtered list")
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicB}, realDBCrossPetSeedNameB)
}

func TestRealDB_Pets_List_MembershipABGrantBSelectedB_ReturnsBOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBCrossClinicPetTestDB(t)
	fx := seedRealDBCrossClinicPetFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		"/api/v1/pets?page=1&limit=50",
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicB),
	)
	fx.handler.ListPets(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var listed httpapi.PaginatedResponse[[]petListResponse]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.NotEmpty(t, listed.Data, "B seed list must be nonempty; empty list is not observation")
	foundB := false
	for _, item := range listed.Data {
		assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
		assert.NotEqual(t, fx.petA.ID, item.ID)
		assert.NotEqual(t, realDBCrossPetSeedNameA, item.Name)
		if item.ID == fx.petB.ID {
			foundB = true
			assert.Equal(t, realDBCrossPetSeedNameB, item.Name)
		}
	}
	require.True(t, foundB, "B seed id/name/clinic_id must appear in list")
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBCrossPetSeedNameA)
}

func TestRealDB_Pets_Get_MembershipABGrantASelectedB_ASeed_ReturnsA(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBCrossClinicPetTestDB(t)
	fx := seedRealDBCrossClinicPetFixture(t, db, 0)
	fx.handler = newRealDBCrossClinicPetHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/pets/%d", fx.petA.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicA)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.petA.ID)}}
		},
	)
	fx.handler.GetPet(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got PetResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, fx.petA.ID, got.ID)
	assert.Equal(t, fx.fx.ClinicA, got.ClinicID)
	assert.Equal(t, realDBCrossPetSeedNameA, got.Name)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicB}, realDBCrossPetSeedNameB)
}

func TestRealDB_Pets_Get_MembershipABGrantASelectedB_BSeed_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBCrossClinicPetTestDB(t)
	fx := seedRealDBCrossClinicPetFixture(t, db, 0)
	fx.handler = newRealDBCrossClinicPetHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/pets/%d", fx.petB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicA)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.petB.ID)}}
		},
	)
	fx.handler.GetPet(c)

	require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	testdb.AssertBodyOmitsClinicArtifacts(
		t,
		w.Body.Bytes(),
		[]uint64{fx.fx.ClinicA, fx.fx.ClinicB},
		realDBCrossPetSeedNameA,
		realDBCrossPetSeedNameB,
	)
}

func TestRealDB_Pets_Get_MembershipABGrantBSelectedB_BSeed_ReturnsB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBCrossClinicPetTestDB(t)
	fx := seedRealDBCrossClinicPetFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/pets/%d", fx.petB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicB)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.petB.ID)}}
		},
	)
	fx.handler.GetPet(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got PetResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, fx.petB.ID, got.ID)
	assert.Equal(t, fx.fx.ClinicB, got.ClinicID)
	assert.Equal(t, realDBCrossPetSeedNameB, got.Name)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBCrossPetSeedNameA)
}

func TestRealDB_OwnerReportPets_MembershipABGrantASelectedB_OwnerA_ReturnsAOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBCrossClinicPetTestDB(t)
	fx := seedRealDBCrossClinicPetFixture(t, db, 0)
	fx.handler = newRealDBCrossClinicPetHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/owners/%d/report/pets", fx.ownerA.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicA)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.ownerA.ID)}}
		},
	)
	fx.handler.ListOwnerReportPets(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var listed ownerReportPetsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.NotEmpty(t, listed.Data, "A owner-report pets must be nonempty")
	foundA := false
	for _, item := range listed.Data {
		assert.NotEqual(t, fx.petB.ID, item.ID)
		assert.NotEqual(t, realDBCrossPetSeedNameB, item.Name)
		if item.ID == fx.petA.ID {
			foundA = true
			assert.Equal(t, realDBCrossPetSeedNameA, item.Name)
		}
	}
	require.True(t, foundA, "A pet must appear in owner-report list")
	assert.NotContains(t, w.Body.String(), realDBCrossPetSeedNameB)
}

func TestRealDB_OwnerReportPets_MembershipABGrantBSelectedB_OwnerB_ReturnsBOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBCrossClinicPetTestDB(t)
	fx := seedRealDBCrossClinicPetFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/owners/%d/report/pets", fx.ownerB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicB)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.ownerB.ID)}}
		},
	)
	fx.handler.ListOwnerReportPets(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var listed ownerReportPetsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.NotEmpty(t, listed.Data, "B owner-report pets must be nonempty")
	foundB := false
	for _, item := range listed.Data {
		assert.NotEqual(t, fx.petA.ID, item.ID)
		assert.NotEqual(t, realDBCrossPetSeedNameA, item.Name)
		if item.ID == fx.petB.ID {
			foundB = true
			assert.Equal(t, realDBCrossPetSeedNameB, item.Name)
		}
	}
	require.True(t, foundB, "B pet must appear in owner-report list")
	assert.NotContains(t, w.Body.String(), realDBCrossPetSeedNameA)
}
