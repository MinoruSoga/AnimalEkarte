package auth

// realdb_selected_clinic_b_grant_a_isolation_test.go — D3-4 auth package (2 routes)
//
// Proves clinic-fixed permission-group list/detail isolation through the real
// repository + application service + HTTP handler path. Middleware deny for
// selected-clinic grants is already covered by D3 middleware suites; this file
// keeps the deterministic ClinicPermissionChecker callback and focuses on
// response payloads from real DB reads (or fail-closed before those reads).
//
// Offline `go test -short` compiles and SKIPs via testdb.SetupTestDB. Runtime
// PostgreSQL execution requires a separately approved disposable database and
// remains NOT_RUN without that authorization.

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
	realDBAuthGroupSeedNameA = "D3auth-realdb-permission-group-A"
	realDBAuthGroupSeedNameB = "D3auth-realdb-permission-group-B"
)

// setupRealDBSelectedClinicBGrantAIsolationTestDB prepares clinics/staff/groups tables.
// Distinct from setupPermissionGroupRepositoryTestDB / setupPermissionGroupStaffIsolationTestDB.
func setupRealDBSelectedClinicBGrantAIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.PermissionGroup{},
		&model.PermissionGroupRule{},
		&model.StaffPermissionGroup{},
	))
	ensureStaffPermissionGroupsCreatedAt(t, db)
	testdb.Truncate(t, db,
		"staff_permission_groups",
		"staff_clinic_assignments",
		"permission_group_rules",
		"permission_groups",
		"staffs",
	)
	return db
}

type realDBAuthPermissionFixture struct {
	clinicA uint64
	clinicB uint64
	staffID uint64
	groupA  *model.PermissionGroup
	groupB  *model.PermissionGroup
	handler *HTTPHandler
}

// clinicQueryGuardService embeds the real PermissionGroupService and fails closed
// if List/GetByID would query a forbidden selected clinic (grant-A deny cases).
type clinicQueryGuardService struct {
	PermissionGroupService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *clinicQueryGuardService) List(
	ctx context.Context,
	clinicID uint64,
) ([]model.PermissionGroup, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("List must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.PermissionGroupService.List(ctx, clinicID)
}

func (s *clinicQueryGuardService) GetByID(
	ctx context.Context,
	clinicID, id uint64,
) (*model.PermissionGroup, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf(
			"GetByID must not query clinic %d (id=%d) without selected-clinic grant",
			clinicID,
			id,
		)
	}
	return s.PermissionGroupService.GetByID(ctx, clinicID, id)
}

func newRealDBAuthPermissionHandler(
	t *testing.T,
	db *gorm.DB,
	forbiddenClinicID uint64,
) *HTTPHandler {
	t.Helper()
	repo := NewPermissionGroupRepository(db)
	realSvc := NewPermissionGroupService(repo, nil, nil)
	var svc PermissionGroupService = realSvc
	if forbiddenClinicID != 0 {
		svc = &clinicQueryGuardService{
			PermissionGroupService: realSvc,
			t:                      t,
			forbiddenClinicID:      forbiddenClinicID,
		}
	}
	return NewHTTPHandler(HTTPDependencies{
		PermissionGroups: svc,
	}, CookieConfigForProduction(false))
}

func seedRealDBAuthPermissionFixture(
	t *testing.T,
	db *gorm.DB,
	forbiddenClinicID uint64,
) realDBAuthPermissionFixture {
	t.Helper()
	clinicA := makePermissionGroupTestClinic(t, db, "D3auth realDB clinic A").ID
	clinicB := makePermissionGroupTestClinic(t, db, "D3auth realDB clinic B").ID

	staff := makeDoctorAssignedToClinic(t, db, clinicA, "D3auth dual-member non-admin")
	require.NoError(t, db.WithContext(context.Background()).Create(&model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicB,
		IsMain:   false,
	}).Error)

	groupA := makePermissionGroup(t, db, clinicA, realDBAuthGroupSeedNameA)
	groupB := makePermissionGroup(t, db, clinicB, realDBAuthGroupSeedNameB)

	return realDBAuthPermissionFixture{
		clinicA: clinicA,
		clinicB: clinicB,
		staffID: staff.ID,
		groupA:  groupA,
		groupB:  groupB,
		handler: newRealDBAuthPermissionHandler(t, db, forbiddenClinicID),
	}
}

func configureSelectedClinicBGrant(
	fx realDBAuthPermissionFixture,
	grantClinicIDs ...uint64,
) func(*gin.Context) {
	granted := make(map[uint64]struct{}, len(grantClinicIDs))
	for _, id := range grantClinicIDs {
		granted[id] = struct{}{}
	}
	resource := string(model.ResourceMasterPermission)
	return func(c *gin.Context) {
		c.Set("clinic_id", fmt.Sprintf("%d", fx.clinicB))
		c.Set("clinic_ids", []uint64{fx.clinicA, fx.clinicB})
		c.Set("user_id", fmt.Sprintf("%d", fx.staffID))
		c.Set("is_system_admin", false)
		httpapi.SetClinicPermissionChecker(c, func(
			_ *gin.Context,
			clinicID uint64,
			res, action string,
		) bool {
			if res != resource || action != "view" {
				return false
			}
			_, ok := granted[clinicID]
			return ok
		})
	}
}

func TestRealDB_PermissionGroups_List_MembershipABGrantASelectedB_Returns403(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	// Seed first, then rebuild handler with a B-query guard using known clinicB.
	fx := seedRealDBAuthPermissionFixture(t, db, 0)
	fx.handler = newRealDBAuthPermissionHandler(t, db, fx.clinicB)

	listContext, listResponse := permissionHTTPContext(
		t,
		http.MethodGet,
		"/api/v1/masters/permission-groups",
		nil,
		configureSelectedClinicBGrant(fx, fx.clinicA),
	)
	fx.handler.ListPermissionGroups(listContext)

	assert.Equal(t, http.StatusForbidden, listResponse.Code)
	// Deny evidence is status + clinicQueryGuard Fatal if List were reached.
	// Reject any accidental permission-group JSON payload (not generic error text).
	assertNoPermissionGroupPayload(t, listResponse.Body.Bytes(), fx)
}

func TestRealDB_PermissionGroups_Get_MembershipABGrantASelectedB_Returns403(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBAuthPermissionFixture(t, db, 0)
	fx.handler = newRealDBAuthPermissionHandler(t, db, fx.clinicB)

	getContext, getResponse := permissionHTTPContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/masters/permission-groups/%d", fx.groupB.ID),
		nil,
		func(c *gin.Context) {
			configureSelectedClinicBGrant(fx, fx.clinicA)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.groupB.ID)}}
		},
	)
	fx.handler.GetPermissionGroup(getContext)

	assert.Equal(t, http.StatusForbidden, getResponse.Code)
	assertNoPermissionGroupPayload(t, getResponse.Body.Bytes(), fx)
}

func TestRealDB_PermissionGroups_List_MembershipABGrantBSelectedB_ReturnsBOnly(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBAuthPermissionFixture(t, db, 0)

	listContext, listResponse := permissionHTTPContext(
		t,
		http.MethodGet,
		"/api/v1/masters/permission-groups",
		nil,
		configureSelectedClinicBGrant(fx, fx.clinicB),
	)
	fx.handler.ListPermissionGroups(listContext)

	require.Equal(t, http.StatusOK, listResponse.Code, listResponse.Body.String())
	var listed []PermissionGroupResponse
	require.NoError(t, json.Unmarshal(listResponse.Body.Bytes(), &listed))
	require.NotEmpty(t, listed, "B seed list must be nonempty; empty list is not observation")
	foundB := false
	for _, item := range listed {
		assert.Equal(t, fx.clinicB, item.ClinicID, "list element clinic_id must stay in selected B")
		assert.NotEqual(t, fx.groupA.ID, item.ID)
		assert.NotEqual(t, realDBAuthGroupSeedNameA, item.Name)
		assert.NotEqual(t, fx.clinicA, item.ClinicID)
		if item.ID == fx.groupB.ID {
			foundB = true
			assert.Equal(t, realDBAuthGroupSeedNameB, item.Name)
		}
	}
	require.True(t, foundB, "B seed id/name/clinic_id must appear in list")
}

func TestRealDB_PermissionGroups_Get_MembershipABGrantBSelectedB_ReturnsB(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBAuthPermissionFixture(t, db, 0)

	getContext, getResponse := permissionHTTPContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/masters/permission-groups/%d", fx.groupB.ID),
		nil,
		func(c *gin.Context) {
			configureSelectedClinicBGrant(fx, fx.clinicB)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.groupB.ID)}}
		},
	)
	fx.handler.GetPermissionGroup(getContext)

	require.Equal(t, http.StatusOK, getResponse.Code, getResponse.Body.String())
	var got PermissionGroupResponse
	require.NoError(t, json.Unmarshal(getResponse.Body.Bytes(), &got))
	assert.Equal(t, fx.groupB.ID, got.ID)
	assert.Equal(t, fx.clinicB, got.ClinicID)
	assert.Equal(t, realDBAuthGroupSeedNameB, got.Name)
	assert.NotEqual(t, fx.groupA.ID, got.ID)
	assert.NotEqual(t, fx.clinicA, got.ClinicID)
	assert.NotEqual(t, realDBAuthGroupSeedNameA, got.Name)
	assert.NotContains(t, getResponse.Body.String(), realDBAuthGroupSeedNameA)
	assert.NotContains(t, getResponse.Body.String(), fmt.Sprintf(`"clinic_id":%d`, fx.clinicA))
}

func TestRealDB_PermissionGroups_Get_SelectedB_ASeedID_Returns404NoABody(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBAuthPermissionFixture(t, db, 0)

	getContext, getResponse := permissionHTTPContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/masters/permission-groups/%d", fx.groupA.ID),
		nil,
		func(c *gin.Context) {
			configureSelectedClinicBGrant(fx, fx.clinicB)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.groupA.ID)}}
		},
	)
	fx.handler.GetPermissionGroup(getContext)

	require.Equal(t, http.StatusNotFound, getResponse.Code, getResponse.Body.String())
	assertNoPermissionGroupPayload(t, getResponse.Body.Bytes(), fx)
}

// assertNoPermissionGroupPayload fails if the body decodes as a permission-group
// object/list that exposes A/B seed identity. Generic error envelopes without
// those fields are acceptable.
func assertNoPermissionGroupPayload(
	t *testing.T,
	body []byte,
	fx realDBAuthPermissionFixture,
) {
	t.Helper()
	var one PermissionGroupResponse
	if err := json.Unmarshal(body, &one); err == nil {
		assert.NotEqual(t, fx.groupA.ID, one.ID)
		assert.NotEqual(t, fx.groupB.ID, one.ID)
		assert.NotEqual(t, realDBAuthGroupSeedNameA, one.Name)
		assert.NotEqual(t, realDBAuthGroupSeedNameB, one.Name)
		assert.NotEqual(t, fx.clinicA, one.ClinicID)
		assert.NotEqual(t, fx.clinicB, one.ClinicID)
		if one.ID != 0 || one.Name != "" || one.ClinicID != 0 {
			t.Fatalf("deny/miss response must not carry permission-group fields: %+v", one)
		}
	}
	var many []PermissionGroupResponse
	if err := json.Unmarshal(body, &many); err == nil {
		require.Empty(t, many, "deny/miss response must not return a permission-group list")
	}
}
