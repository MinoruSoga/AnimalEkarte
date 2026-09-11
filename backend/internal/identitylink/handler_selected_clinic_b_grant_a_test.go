package identitylink

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

func membershipABGrantASelectedBIdentityLinks(c *gin.Context) {
	c.Set("clinic_id", "2")
	c.Set("clinic_ids", []uint64{1, 2})
	c.Set("is_system_admin", false)
	c.Set("user_id", "17")
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, clinicID uint64, resource, action string) bool {
		return clinicID == 1 && resource == string(model.ResourceIdentityLinks) && action == "view"
	})
}

type identityLinkScopeStub struct {
	Service
	searchOwnersFn func(context.Context, ActorContext, string, int) ([]model.Owner, error)
	searchPetsFn   func(context.Context, ActorContext, string, int) ([]model.Pet, error)
	findOwnerFn    func(context.Context, ActorContext, uint64, uint64) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error)
	findPetFn      func(context.Context, ActorContext, uint64, uint64) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error)
	getOwnerFn     func(context.Context, ActorContext, uint64) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error)
	getPetFn       func(context.Context, ActorContext, uint64) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error)
	historyFn      func(context.Context, ActorContext, uint64, uint64, bool, int, int) ([]LinkedTreatmentHistoryItem, int64, error)
}

func (s *identityLinkScopeStub) SearchOwners(
	ctx context.Context,
	actor ActorContext,
	query string,
	limit int,
) ([]model.Owner, error) {
	return s.searchOwnersFn(ctx, actor, query, limit)
}

func (s *identityLinkScopeStub) SearchPets(
	ctx context.Context,
	actor ActorContext,
	query string,
	limit int,
) ([]model.Pet, error) {
	return s.searchPetsFn(ctx, actor, query, limit)
}

func (s *identityLinkScopeStub) FindOwnerGroupByMember(
	ctx context.Context,
	actor ActorContext,
	clinicID, ownerID uint64,
) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
	return s.findOwnerFn(ctx, actor, clinicID, ownerID)
}

func (s *identityLinkScopeStub) FindPetGroupByMember(
	ctx context.Context,
	actor ActorContext,
	clinicID, petID uint64,
) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
	return s.findPetFn(ctx, actor, clinicID, petID)
}

func (s *identityLinkScopeStub) GetOwnerGroup(
	ctx context.Context,
	actor ActorContext,
	groupID uint64,
) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
	return s.getOwnerFn(ctx, actor, groupID)
}

func (s *identityLinkScopeStub) GetPetGroup(
	ctx context.Context,
	actor ActorContext,
	groupID uint64,
) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
	return s.getPetFn(ctx, actor, groupID)
}

func (s *identityLinkScopeStub) ListLinkedTreatmentHistory(
	ctx context.Context,
	actor ActorContext,
	seedClinicID, seedPetID uint64,
	includeLinked bool,
	page, limit int,
) ([]LinkedTreatmentHistoryItem, int64, error) {
	return s.historyFn(ctx, actor, seedClinicID, seedPetID, includeLinked, page, limit)
}

func TestSearchOwners_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&identityLinkScopeStub{
		searchOwnersFn: func(_ context.Context, actor ActorContext, _ string, _ int) ([]model.Owner, error) {
			require.Equal(t, []uint64{1}, actor.VerifiedClinics)
			return []model.Owner{}, nil
		},
	}, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?q=a", http.NoBody)
	membershipABGrantASelectedBIdentityLinks(c)
	h.SearchOwners(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"items"`)
}

func TestFindOwnerGroupByMember_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&identityLinkScopeStub{
		findOwnerFn: func(_ context.Context, actor ActorContext, clinicID, _ uint64) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
			require.Equal(t, []uint64{1}, actor.VerifiedClinics)
			require.Equal(t, uint64(2), clinicID)
			return nil, nil, apperrors.WrapForbidden("owner clinic outside actor scope")
		},
	}, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{
		{Key: "clinic_id", Value: "2"},
		{Key: "owner_id", Value: "9"},
	}
	membershipABGrantASelectedBIdentityLinks(c)
	h.FindOwnerGroupByMember(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSearchPets_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&identityLinkScopeStub{
		searchPetsFn: func(_ context.Context, actor ActorContext, _ string, _ int) ([]model.Pet, error) {
			require.Equal(t, []uint64{1}, actor.VerifiedClinics)
			return []model.Pet{}, nil
		},
	}, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?q=a", http.NoBody)
	membershipABGrantASelectedBIdentityLinks(c)
	h.SearchPets(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"items"`)
}

func TestGetOwnerGroup_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&identityLinkScopeStub{
		getOwnerFn: func(_ context.Context, actor ActorContext, groupID uint64) (*model.OwnerIdentityGroup, []model.OwnerIdentityGroupMember, error) {
			require.Equal(t, []uint64{1}, actor.VerifiedClinics)
			return &model.OwnerIdentityGroup{ID: groupID, CreatedClinicID: 1}, nil, nil
		},
	}, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "3"}}
	membershipABGrantASelectedBIdentityLinks(c)
	h.GetOwnerGroup(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), `"clinic_id":2`)
}

func TestGetPetGroup_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&identityLinkScopeStub{
		getPetFn: func(_ context.Context, actor ActorContext, groupID uint64) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
			require.Equal(t, []uint64{1}, actor.VerifiedClinics)
			return &model.PetIdentityGroup{ID: groupID, CreatedClinicID: 1}, nil, nil
		},
	}, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "4"}}
	membershipABGrantASelectedBIdentityLinks(c)
	h.GetPetGroup(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), `"clinic_id":2`)
}

func TestFindPetGroupByMember_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&identityLinkScopeStub{
		findPetFn: func(_ context.Context, actor ActorContext, clinicID, _ uint64) (*model.PetIdentityGroup, []model.PetIdentityGroupMember, error) {
			require.Equal(t, []uint64{1}, actor.VerifiedClinics)
			require.Equal(t, uint64(2), clinicID)
			return nil, nil, apperrors.WrapForbidden("pet clinic outside actor scope")
		},
	}, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{
		{Key: "clinic_id", Value: "2"},
		{Key: "pet_id", Value: "9"},
	}
	membershipABGrantASelectedBIdentityLinks(c)
	h.FindPetGroupByMember(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestListLinkedTreatmentHistory_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&identityLinkScopeStub{
		historyFn: func(_ context.Context, actor ActorContext, clinicID, _ uint64, _ bool, _, _ int) ([]LinkedTreatmentHistoryItem, int64, error) {
			require.Equal(t, []uint64{1}, actor.VerifiedClinics)
			require.Equal(t, uint64(2), clinicID)
			return nil, 0, apperrors.WrapForbidden("seed pet clinic outside actor scope")
		},
	}, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{
		{Key: "clinic_id", Value: "2"},
		{Key: "pet_id", Value: "9"},
	}
	membershipABGrantASelectedBIdentityLinks(c)
	h.ListLinkedTreatmentHistory(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
