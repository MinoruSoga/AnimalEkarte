package owner

// realdb_selected_clinic_b_grant_a_isolation_test.go — D3 owner cross-clinic (2 routes)
//
// Proves owners list/detail isolation through real repository + service + HTTP
// handler with a deterministic ClinicPermissionChecker.
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
	realDBOwnerSeedNameA = "D3xcl-realdb-owner-A"
	realDBOwnerSeedNameB = "D3xcl-realdb-owner-B"
)

func setupRealDBSelectedClinicBGrantAIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.Owner{},
	))
	testdb.Truncate(t, db,
		"owners",
		"staff_clinic_assignments",
		"staffs",
	)
	return db
}

type ownerClinicQueryGuardService struct {
	Service
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *ownerClinicQueryGuardService) assertAllowed(clinicIDs []uint64, op string) {
	s.t.Helper()
	for _, id := range clinicIDs {
		if id == s.forbiddenClinicID {
			s.t.Fatalf("%s must not query clinic %d without grant", op, id)
		}
	}
}

func (s *ownerClinicQueryGuardService) List(
	ctx context.Context,
	clinicIDs []uint64,
	page, limit int,
	search string,
) ([]model.Owner, int64, error) {
	s.assertAllowed(clinicIDs, "List")
	return s.Service.List(ctx, clinicIDs, page, limit, search)
}

func (s *ownerClinicQueryGuardService) GetByIDForClinics(
	ctx context.Context,
	clinicIDs []uint64,
	id uint64,
) (*model.Owner, error) {
	s.assertAllowed(clinicIDs, "GetByIDForClinics")
	return s.Service.GetByIDForClinics(ctx, clinicIDs, id)
}

type realDBOwnerFixture struct {
	fx      testdb.ClinicGrantFixture
	ownerA  *model.Owner
	ownerB  *model.Owner
	handler *Handler
}

func newRealDBOwnerHandler(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) *Handler {
	t.Helper()
	var svc Service = NewService(NewRepository(db, nil), nil, nil, nil)
	if forbiddenClinicID != 0 {
		svc = &ownerClinicQueryGuardService{
			Service:           svc,
			t:                 t,
			forbiddenClinicID: forbiddenClinicID,
		}
	}
	return NewHandler(svc, nil, nil, nil)
}

func seedRealDBOwnerFixture(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) realDBOwnerFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3xcl owner realDB")
	ownerA := testdb.MakeTestOwner(t, db, fx.ClinicA, realDBOwnerSeedNameA)
	ownerB := testdb.MakeTestOwner(t, db, fx.ClinicB, realDBOwnerSeedNameB)
	return realDBOwnerFixture{
		fx:      fx,
		ownerA:  ownerA,
		ownerB:  ownerB,
		handler: newRealDBOwnerHandler(t, db, forbiddenClinicID),
	}
}

func TestRealDB_Owners_List_MembershipABGrantASelectedB_Default_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBOwnerFixture(t, db, 0)
	fx.handler = newRealDBOwnerHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		"/api/v1/owners?page=1&limit=50",
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicA),
	)
	fx.handler.ListOwners(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	testdb.AssertBodyOmitsClinicArtifacts(
		t,
		w.Body.Bytes(),
		[]uint64{fx.fx.ClinicA, fx.fx.ClinicB},
		realDBOwnerSeedNameA,
		realDBOwnerSeedNameB,
	)
}

func TestRealDB_Owners_List_MembershipABGrantASelectedB_MixedClinicIDs_ReturnsAOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBOwnerFixture(t, db, 0)
	fx.handler = newRealDBOwnerHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/owners?page=1&limit=50&clinic_ids=%d,%d", fx.fx.ClinicA, fx.fx.ClinicB),
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicA),
	)
	fx.handler.ListOwners(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var listed httpapi.PaginatedResponse[[]OwnerResponse]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.NotEmpty(t, listed.Data, "A seed list must be nonempty; empty list is not observation")
	foundA := false
	for _, item := range listed.Data {
		assert.Equal(t, fx.fx.ClinicA, item.ClinicID)
		assert.NotEqual(t, fx.ownerB.ID, item.ID)
		assert.NotEqual(t, realDBOwnerSeedNameB, item.OwnerName)
		if item.ID == fx.ownerA.ID {
			foundA = true
			assert.Equal(t, realDBOwnerSeedNameA, item.OwnerName)
		}
	}
	require.True(t, foundA, "A seed id/name/clinic_id must appear in filtered list")
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicB}, realDBOwnerSeedNameB)
}

func TestRealDB_Owners_List_MembershipABGrantASelectedB_ExplicitB_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBOwnerFixture(t, db, 0)
	fx.handler = newRealDBOwnerHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/owners?page=1&limit=50&clinic_ids=%d", fx.fx.ClinicB),
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicA),
	)
	fx.handler.ListOwners(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	testdb.AssertBodyOmitsClinicArtifacts(
		t,
		w.Body.Bytes(),
		[]uint64{fx.fx.ClinicA, fx.fx.ClinicB},
		realDBOwnerSeedNameA,
		realDBOwnerSeedNameB,
	)
}

func TestRealDB_Owners_List_MembershipABGrantBSelectedB_ReturnsBOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBOwnerFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		"/api/v1/owners?page=1&limit=50",
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicB),
	)
	fx.handler.ListOwners(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var listed httpapi.PaginatedResponse[[]OwnerResponse]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.NotEmpty(t, listed.Data, "B seed list must be nonempty; empty list is not observation")
	foundB := false
	for _, item := range listed.Data {
		assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
		assert.NotEqual(t, fx.ownerA.ID, item.ID)
		assert.NotEqual(t, realDBOwnerSeedNameA, item.OwnerName)
		if item.ID == fx.ownerB.ID {
			foundB = true
			assert.Equal(t, realDBOwnerSeedNameB, item.OwnerName)
		}
	}
	require.True(t, foundB, "B seed id/name/clinic_id must appear in list")
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBOwnerSeedNameA)
}

func TestRealDB_Owners_Get_MembershipABGrantASelectedB_ASeed_ReturnsA(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBOwnerFixture(t, db, 0)
	fx.handler = newRealDBOwnerHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/owners/%d", fx.ownerA.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicA)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.ownerA.ID)}}
		},
	)
	fx.handler.GetOwner(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got OwnerResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, fx.ownerA.ID, got.ID)
	assert.Equal(t, fx.fx.ClinicA, got.ClinicID)
	assert.Equal(t, realDBOwnerSeedNameA, got.OwnerName)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicB}, realDBOwnerSeedNameB)
}

func TestRealDB_Owners_Get_MembershipABGrantASelectedB_BSeed_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBOwnerFixture(t, db, 0)
	fx.handler = newRealDBOwnerHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/owners/%d", fx.ownerB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicA)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.ownerB.ID)}}
		},
	)
	fx.handler.GetOwner(c)

	require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	testdb.AssertBodyOmitsClinicArtifacts(
		t,
		w.Body.Bytes(),
		[]uint64{fx.fx.ClinicA, fx.fx.ClinicB},
		realDBOwnerSeedNameA,
		realDBOwnerSeedNameB,
	)
}

func TestRealDB_Owners_Get_MembershipABGrantBSelectedB_BSeed_ReturnsB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBOwnerFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/owners/%d", fx.ownerB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicB)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.ownerB.ID)}}
		},
	)
	fx.handler.GetOwner(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got OwnerResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, fx.ownerB.ID, got.ID)
	assert.Equal(t, fx.fx.ClinicB, got.ClinicID)
	assert.Equal(t, realDBOwnerSeedNameB, got.OwnerName)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBOwnerSeedNameA)
}

func TestRealDB_Owners_Get_MembershipABGrantBSelectedB_ASeed_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBOwnerFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/owners/%d", fx.ownerA.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceOwners), fx.fx.ClinicB)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.ownerA.ID)}}
		},
	)
	fx.handler.GetOwner(c)

	require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	testdb.AssertBodyOmitsClinicArtifacts(
		t,
		w.Body.Bytes(),
		[]uint64{fx.fx.ClinicA, fx.fx.ClinicB},
		realDBOwnerSeedNameA,
		realDBOwnerSeedNameB,
	)
}
