package identitylink

// realdb_selected_clinic_b_grant_a_isolation_test.go — D3 identitylink cross-clinic (7 routes)
//
// Proves identity-link search/group/history isolation through real repository +
// service + HTTP handler with a deterministic ClinicPermissionChecker.
// Offline `go test -short` SKIPs via testdb.SetupTestDB.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	realDBIDOwnerNameA = "D3xcl-realdb-idlink-owner-A"
	realDBIDOwnerNameB = "D3xcl-realdb-idlink-owner-B"
	realDBIDPetNameA   = "D3xcl-realdb-idlink-pet-A"
	realDBIDPetNameB   = "D3xcl-realdb-idlink-pet-B"
	realDBIDTxContentA = "D3xcl-realdb-idlink-tx-A"
	realDBIDTxContentB = "D3xcl-realdb-idlink-tx-B"
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
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.MedicalRecord{},
		&model.Treatment{},
		&model.OwnerIdentityGroup{},
		&model.OwnerIdentityGroupMember{},
		&model.PetIdentityGroup{},
		&model.PetIdentityGroupMember{},
	))
	testdb.Truncate(t, db,
		"pet_identity_group_members",
		"pet_identity_groups",
		"owner_identity_group_members",
		"owner_identity_groups",
		"treatments",
		"medical_records",
		"pets",
		"animal_species",
		"owners",
		"staff_clinic_assignments",
		"staffs",
	)
	return db
}

type identityClinicQueryGuardService struct {
	Service
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *identityClinicQueryGuardService) assertActor(actor ActorContext, op string) {
	s.t.Helper()
	for _, id := range actor.VerifiedClinics {
		if id == s.forbiddenClinicID {
			s.t.Fatalf("%s must not receive forbidden clinic %d in actor scope", op, id)
		}
	}
}

func (s *identityClinicQueryGuardService) SearchOwners(
	ctx context.Context,
	actor ActorContext,
	query string,
	limit int,
) ([]model.Owner, error) {
	s.assertActor(actor, "SearchOwners")
	return s.Service.SearchOwners(ctx, actor, query, limit)
}

func (s *identityClinicQueryGuardService) SearchPets(
	ctx context.Context,
	actor ActorContext,
	query string,
	limit int,
) ([]model.Pet, error) {
	s.assertActor(actor, "SearchPets")
	return s.Service.SearchPets(ctx, actor, query, limit)
}

func (s *identityClinicQueryGuardService) GetOwnerGroup(
	ctx context.Context,
	actor ActorContext,
	groupID uint64,
) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
	s.assertActor(actor, "GetOwnerGroup")
	return s.Service.GetOwnerGroup(ctx, actor, groupID)
}

func (s *identityClinicQueryGuardService) GetPetGroup(
	ctx context.Context,
	actor ActorContext,
	groupID uint64,
) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
	s.assertActor(actor, "GetPetGroup")
	return s.Service.GetPetGroup(ctx, actor, groupID)
}

func (s *identityClinicQueryGuardService) FindOwnerGroupByMember(
	ctx context.Context,
	actor ActorContext,
	clinicID, ownerID uint64,
) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
	s.assertActor(actor, "FindOwnerGroupByMember")
	return s.Service.FindOwnerGroupByMember(ctx, actor, clinicID, ownerID)
}

func (s *identityClinicQueryGuardService) FindPetGroupByMember(
	ctx context.Context,
	actor ActorContext,
	clinicID, petID uint64,
) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
	s.assertActor(actor, "FindPetGroupByMember")
	return s.Service.FindPetGroupByMember(ctx, actor, clinicID, petID)
}

func (s *identityClinicQueryGuardService) ListLinkedTreatmentHistory(
	ctx context.Context,
	actor ActorContext,
	seedClinicID, seedPetID uint64,
	includeLinked bool,
	page, limit int,
) ([]LinkedTreatmentHistoryItem, int64, error) {
	s.assertActor(actor, "ListLinkedTreatmentHistory")
	return s.Service.ListLinkedTreatmentHistory(ctx, actor, seedClinicID, seedPetID, includeLinked, page, limit)
}

type realDBIdentityFixture struct {
	fx           testdb.ClinicGrantFixture
	ownerA       *model.Owner
	ownerB       *model.Owner
	petA         *model.Pet
	petB         *model.Pet
	ownerGroupA  *model.OwnerIdentityGroup
	ownerGroupB  *model.OwnerIdentityGroup
	ownerGroupAB *model.OwnerIdentityGroup
	petGroupA    *model.PetIdentityGroup
	petGroupB    *model.PetIdentityGroup
	handler      *Handler
}

func newRealDBIdentityHandler(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) *Handler {
	t.Helper()
	var svc Service = NewService(NewRepository(db), nil, nil)
	if forbiddenClinicID != 0 {
		svc = &identityClinicQueryGuardService{
			Service:           svc,
			t:                 t,
			forbiddenClinicID: forbiddenClinicID,
		}
	}
	return NewHandler(svc, nil)
}

func seedOwnerGroup(
	t *testing.T,
	db *gorm.DB,
	createdClinicID uint64,
	members ...model.OwnerIdentityGroupMember,
) *model.OwnerIdentityGroup {
	t.Helper()
	group := &model.OwnerIdentityGroup{CreatedClinicID: createdClinicID, Version: 1}
	require.NoError(t, db.Create(group).Error)
	for i := range members {
		members[i].GroupID = group.ID
		members[i].GroupCreatedClinicID = createdClinicID
		require.NoError(t, db.Create(&members[i]).Error)
	}
	return group
}

func seedPetGroup(
	t *testing.T,
	db *gorm.DB,
	createdClinicID uint64,
	ownerGroup *model.OwnerIdentityGroup,
	members ...model.PetIdentityGroupMember,
) *model.PetIdentityGroup {
	t.Helper()
	group := &model.PetIdentityGroup{
		CreatedClinicID:           createdClinicID,
		OwnerGroupCreatedClinicID: ownerGroup.CreatedClinicID,
		OwnerGroupID:              ownerGroup.ID,
		Version:                   1,
	}
	require.NoError(t, db.Create(group).Error)
	for i := range members {
		members[i].GroupID = group.ID
		members[i].GroupCreatedClinicID = createdClinicID
		require.NoError(t, db.Create(&members[i]).Error)
	}
	return group
}

func seedTreatment(
	t *testing.T,
	db *gorm.DB,
	clinicID, petID uint64,
	recordNo, content string,
) {
	t.Helper()
	pet := petID
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		PetID:    &pet,
		Status:   model.MedicalRecordStatusFinalized,
	}
	require.NoError(t, db.Create(mr).Error)
	tr := &model.Treatment{
		MedicalRecordID: mr.ID,
		Content:         content,
		ItemType:        model.TreatmentItemTypeOther,
	}
	require.NoError(t, db.Create(tr).Error)
}

func seedRealDBIdentityFixture(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) realDBIdentityFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3xcl identitylink realDB")
	ownerA := testdb.MakeTestOwner(t, db, fx.ClinicA, realDBIDOwnerNameA)
	ownerB := testdb.MakeTestOwner(t, db, fx.ClinicB, realDBIDOwnerNameB)
	// Distinct members for the mixed group so active (clinic_id, owner_id) stays unique.
	ownerMixedA := testdb.MakeTestOwner(t, db, fx.ClinicA, realDBIDOwnerNameA+"-mixed")
	ownerMixedB := testdb.MakeTestOwner(t, db, fx.ClinicB, realDBIDOwnerNameB+"-mixed")
	species := &model.AnimalSpecies{Name: "D3xcl-realdb-idlink-species", IsActive: true}
	require.NoError(t, db.Create(species).Error)
	petA := &model.Pet{
		ClinicID: fx.ClinicA, OwnerID: ownerA.ID, AnimalSpeciesID: species.ID, Name: realDBIDPetNameA,
	}
	require.NoError(t, db.Create(petA).Error)
	petB := &model.Pet{
		ClinicID: fx.ClinicB, OwnerID: ownerB.ID, AnimalSpeciesID: species.ID, Name: realDBIDPetNameB,
	}
	require.NoError(t, db.Create(petB).Error)

	ownerGroupA := seedOwnerGroup(t, db, fx.ClinicA, model.OwnerIdentityGroupMember{
		ClinicID: fx.ClinicA, OwnerID: ownerA.ID,
	})
	ownerGroupB := seedOwnerGroup(t, db, fx.ClinicB, model.OwnerIdentityGroupMember{
		ClinicID: fx.ClinicB, OwnerID: ownerB.ID,
	})
	ownerGroupAB := seedOwnerGroup(t, db, fx.ClinicA,
		model.OwnerIdentityGroupMember{ClinicID: fx.ClinicA, OwnerID: ownerMixedA.ID},
		model.OwnerIdentityGroupMember{ClinicID: fx.ClinicB, OwnerID: ownerMixedB.ID},
	)
	petGroupA := seedPetGroup(t, db, fx.ClinicA, ownerGroupA, model.PetIdentityGroupMember{
		ClinicID: fx.ClinicA, PetID: petA.ID,
	})
	petGroupB := seedPetGroup(t, db, fx.ClinicB, ownerGroupB, model.PetIdentityGroupMember{
		ClinicID: fx.ClinicB, PetID: petB.ID,
	})
	seedTreatment(t, db, fx.ClinicA, petA.ID, "D3XCL-A1", realDBIDTxContentA)
	seedTreatment(t, db, fx.ClinicB, petB.ID, "D3XCL-B1", realDBIDTxContentB)

	return realDBIdentityFixture{
		fx:           fx,
		ownerA:       ownerA,
		ownerB:       ownerB,
		petA:         petA,
		petB:         petB,
		ownerGroupA:  ownerGroupA,
		ownerGroupB:  ownerGroupB,
		ownerGroupAB: ownerGroupAB,
		petGroupA:    petGroupA,
		petGroupB:    petGroupB,
		handler:      newRealDBIdentityHandler(t, db, forbiddenClinicID),
	}
}

type ownerSearchEnvelope struct {
	Items []OwnerSearchItem `json:"items"`
}

type petSearchEnvelope struct {
	Items []PetSearchItem `json:"items"`
}

func TestRealDB_IdentityLinks_SearchOwners_MembershipABGrantASelectedB_ReturnsAOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)
	fx.handler = newRealDBIdentityHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		"/api/v1/identity-links/owners/search?q="+url.QueryEscape(realDBIDOwnerNameA),
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicA),
	)
	fx.handler.SearchOwners(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var listed ownerSearchEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.NotEmpty(t, listed.Items, "A search must be nonempty")
	foundA := false
	for _, item := range listed.Items {
		assert.Equal(t, fx.fx.ClinicA, item.ClinicID)
		assert.NotEqual(t, fx.ownerB.ID, item.OwnerID)
		assert.NotEqual(t, realDBIDOwnerNameB, item.Name)
		if item.OwnerID == fx.ownerA.ID {
			foundA = true
		}
	}
	require.True(t, foundA)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicB}, realDBIDOwnerNameB)
}

func TestRealDB_IdentityLinks_SearchOwners_MembershipABGrantASelectedB_QueryB_OmitsB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)
	fx.handler = newRealDBIdentityHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		"/api/v1/identity-links/owners/search?q="+url.QueryEscape(realDBIDOwnerNameB),
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicA),
	)
	fx.handler.SearchOwners(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var listed ownerSearchEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	assert.Empty(t, listed.Items)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicB}, realDBIDOwnerNameB)
}

func TestRealDB_IdentityLinks_SearchOwners_MembershipABGrantBSelectedB_ReturnsB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		"/api/v1/identity-links/owners/search?q="+url.QueryEscape(realDBIDOwnerNameB),
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicB),
	)
	fx.handler.SearchOwners(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var listed ownerSearchEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.NotEmpty(t, listed.Items, "B search must be nonempty")
	foundB := false
	for _, item := range listed.Items {
		assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
		assert.NotEqual(t, fx.ownerA.ID, item.OwnerID)
		if item.OwnerID == fx.ownerB.ID {
			foundB = true
		}
	}
	require.True(t, foundB)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBIDOwnerNameA)
}

func TestRealDB_IdentityLinks_SearchPets_MembershipABGrantASelectedB_ReturnsAOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)
	fx.handler = newRealDBIdentityHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		"/api/v1/identity-links/pets/search?q="+url.QueryEscape(realDBIDPetNameA),
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicA),
	)
	fx.handler.SearchPets(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var listed petSearchEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.NotEmpty(t, listed.Items, "A pet search must be nonempty")
	foundA := false
	for _, item := range listed.Items {
		assert.Equal(t, fx.fx.ClinicA, item.ClinicID)
		assert.NotEqual(t, fx.petB.ID, item.PetID)
		if item.PetID == fx.petA.ID {
			foundA = true
		}
	}
	require.True(t, foundA)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicB}, realDBIDPetNameB)
}

func TestRealDB_IdentityLinks_SearchPets_MembershipABGrantBSelectedB_ReturnsB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		"/api/v1/identity-links/pets/search?q="+url.QueryEscape(realDBIDPetNameB),
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicB),
	)
	fx.handler.SearchPets(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var listed petSearchEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.NotEmpty(t, listed.Items, "B pet search must be nonempty")
	foundB := false
	for _, item := range listed.Items {
		assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
		if item.PetID == fx.petB.ID {
			foundB = true
		}
	}
	require.True(t, foundB)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBIDPetNameA)
}

func TestRealDB_IdentityLinks_GetOwnerGroup_MembershipABGrantASelectedB_GroupA_ReturnsA(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)
	fx.handler = newRealDBIdentityHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/identity-links/owner-groups/%d", fx.ownerGroupA.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicA)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.ownerGroupA.ID)}}
		},
	)
	fx.handler.GetOwnerGroup(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got OwnerGroupResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.NotEmpty(t, got.Members)
	for _, m := range got.Members {
		assert.Equal(t, fx.fx.ClinicA, m.ClinicID)
		assert.NotEqual(t, fx.ownerB.ID, m.OwnerID)
	}
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicB}, realDBIDOwnerNameB)
}

func TestRealDB_IdentityLinks_GetOwnerGroup_MembershipABGrantASelectedB_GroupB_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)
	fx.handler = newRealDBIdentityHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/identity-links/owner-groups/%d", fx.ownerGroupB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicA)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.ownerGroupB.ID)}}
		},
	)
	fx.handler.GetOwnerGroup(c)

	require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBIDOwnerNameA, realDBIDOwnerNameB)
}

func TestRealDB_IdentityLinks_GetOwnerGroup_MembershipABGrantASelectedB_MixedGroup_ReturnsAMembersOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)
	fx.handler = newRealDBIdentityHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/identity-links/owner-groups/%d", fx.ownerGroupAB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicA)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.ownerGroupAB.ID)}}
		},
	)
	fx.handler.GetOwnerGroup(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got OwnerGroupResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Len(t, got.Members, 1)
	assert.Equal(t, fx.fx.ClinicA, got.Members[0].ClinicID)
	assert.NotEqual(t, fx.ownerB.ID, got.Members[0].OwnerID)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicB}, realDBIDOwnerNameB)
}

func TestRealDB_IdentityLinks_GetOwnerGroup_MembershipABGrantBSelectedB_GroupB_ReturnsB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/identity-links/owner-groups/%d", fx.ownerGroupB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicB)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.ownerGroupB.ID)}}
		},
	)
	fx.handler.GetOwnerGroup(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got OwnerGroupResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.NotEmpty(t, got.Members)
	foundB := false
	for _, m := range got.Members {
		assert.Equal(t, fx.fx.ClinicB, m.ClinicID)
		if m.OwnerID == fx.ownerB.ID {
			foundB = true
		}
	}
	require.True(t, foundB)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBIDOwnerNameA)
}

func TestRealDB_IdentityLinks_FindOwnerGroupByMember_MembershipABGrantASelectedB_PathB_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)
	fx.handler = newRealDBIdentityHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/identity-links/owners/%d/%d/group", fx.fx.ClinicB, fx.ownerB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicA)(c)
			c.Params = gin.Params{
				{Key: "clinic_id", Value: fmt.Sprintf("%d", fx.fx.ClinicB)},
				{Key: "owner_id", Value: fmt.Sprintf("%d", fx.ownerB.ID)},
			}
		},
	)
	fx.handler.FindOwnerGroupByMember(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBIDOwnerNameA, realDBIDOwnerNameB)
}

func TestRealDB_IdentityLinks_FindOwnerGroupByMember_MembershipABGrantBSelectedB_PathB_ReturnsB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/identity-links/owners/%d/%d/group", fx.fx.ClinicB, fx.ownerB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicB)(c)
			c.Params = gin.Params{
				{Key: "clinic_id", Value: fmt.Sprintf("%d", fx.fx.ClinicB)},
				{Key: "owner_id", Value: fmt.Sprintf("%d", fx.ownerB.ID)},
			}
		},
	)
	fx.handler.FindOwnerGroupByMember(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got OwnerGroupResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.NotEmpty(t, got.Members)
	foundB := false
	for _, m := range got.Members {
		assert.Equal(t, fx.fx.ClinicB, m.ClinicID)
		if m.OwnerID == fx.ownerB.ID {
			foundB = true
		}
	}
	require.True(t, foundB)
}

func TestRealDB_IdentityLinks_GetPetGroup_MembershipABGrantASelectedB_GroupB_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)
	fx.handler = newRealDBIdentityHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/identity-links/pet-groups/%d", fx.petGroupB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicA)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.petGroupB.ID)}}
		},
	)
	fx.handler.GetPetGroup(c)

	require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBIDPetNameA, realDBIDPetNameB)
}

func TestRealDB_IdentityLinks_GetPetGroup_MembershipABGrantBSelectedB_GroupB_ReturnsB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/identity-links/pet-groups/%d", fx.petGroupB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicB)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.petGroupB.ID)}}
		},
	)
	fx.handler.GetPetGroup(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got PetGroupResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.NotEmpty(t, got.Members)
	foundB := false
	for _, m := range got.Members {
		assert.Equal(t, fx.fx.ClinicB, m.ClinicID)
		if m.PetID == fx.petB.ID {
			foundB = true
		}
	}
	require.True(t, foundB)
}

func TestRealDB_IdentityLinks_FindPetGroupByMember_MembershipABGrantASelectedB_PathB_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)
	fx.handler = newRealDBIdentityHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/identity-links/pets/%d/%d/group", fx.fx.ClinicB, fx.petB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicA)(c)
			c.Params = gin.Params{
				{Key: "clinic_id", Value: fmt.Sprintf("%d", fx.fx.ClinicB)},
				{Key: "pet_id", Value: fmt.Sprintf("%d", fx.petB.ID)},
			}
		},
	)
	fx.handler.FindPetGroupByMember(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBIDPetNameA, realDBIDPetNameB)
}

func TestRealDB_IdentityLinks_FindPetGroupByMember_MembershipABGrantBSelectedB_PathB_ReturnsB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/identity-links/pets/%d/%d/group", fx.fx.ClinicB, fx.petB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicB)(c)
			c.Params = gin.Params{
				{Key: "clinic_id", Value: fmt.Sprintf("%d", fx.fx.ClinicB)},
				{Key: "pet_id", Value: fmt.Sprintf("%d", fx.petB.ID)},
			}
		},
	)
	fx.handler.FindPetGroupByMember(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got PetGroupResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.NotEmpty(t, got.Members)
	foundB := false
	for _, m := range got.Members {
		assert.Equal(t, fx.fx.ClinicB, m.ClinicID)
		if m.PetID == fx.petB.ID {
			foundB = true
		}
	}
	require.True(t, foundB)
}

func TestRealDB_IdentityLinks_TreatmentHistory_MembershipABGrantASelectedB_SeedB_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)
	fx.handler = newRealDBIdentityHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/identity-links/pets/%d/%d/treatment-history", fx.fx.ClinicB, fx.petB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicA)(c)
			c.Params = gin.Params{
				{Key: "clinic_id", Value: fmt.Sprintf("%d", fx.fx.ClinicB)},
				{Key: "pet_id", Value: fmt.Sprintf("%d", fx.petB.ID)},
			}
		},
	)
	fx.handler.ListLinkedTreatmentHistory(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBIDTxContentA, realDBIDTxContentB)
}

func TestRealDB_IdentityLinks_TreatmentHistory_MembershipABGrantASelectedB_SeedA_ReturnsAOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)
	fx.handler = newRealDBIdentityHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/identity-links/pets/%d/%d/treatment-history?include_linked=true", fx.fx.ClinicA, fx.petA.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicA)(c)
			c.Params = gin.Params{
				{Key: "clinic_id", Value: fmt.Sprintf("%d", fx.fx.ClinicA)},
				{Key: "pet_id", Value: fmt.Sprintf("%d", fx.petA.ID)},
			}
		},
	)
	fx.handler.ListLinkedTreatmentHistory(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got LinkedTreatmentHistoryResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.NotEmpty(t, got.Items, "A treatment history must be nonempty")
	foundA := false
	for _, item := range got.Items {
		assert.Equal(t, fx.fx.ClinicA, item.ClinicID)
		assert.NotEqual(t, fx.petB.ID, item.PetID)
		assert.NotEqual(t, realDBIDTxContentB, item.Content)
		if item.Content == realDBIDTxContentA {
			foundA = true
		}
	}
	require.True(t, foundA)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicB}, realDBIDTxContentB)
}

func TestRealDB_IdentityLinks_TreatmentHistory_MembershipABGrantBSelectedB_SeedB_ReturnsB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBIdentityFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/identity-links/pets/%d/%d/treatment-history", fx.fx.ClinicB, fx.petB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceIdentityLinks), fx.fx.ClinicB)(c)
			c.Params = gin.Params{
				{Key: "clinic_id", Value: fmt.Sprintf("%d", fx.fx.ClinicB)},
				{Key: "pet_id", Value: fmt.Sprintf("%d", fx.petB.ID)},
			}
		},
	)
	fx.handler.ListLinkedTreatmentHistory(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got LinkedTreatmentHistoryResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.NotEmpty(t, got.Items, "B treatment history must be nonempty")
	foundB := false
	for _, item := range got.Items {
		assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
		assert.NotEqual(t, fx.petA.ID, item.PetID)
		if item.Content == realDBIDTxContentB {
			foundB = true
		}
	}
	require.True(t, foundB)
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBIDTxContentA)
}
