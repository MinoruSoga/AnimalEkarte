package pet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func membershipABGrantASelectedB(c *gin.Context) {
	c.Set("clinic_id", "2")
	c.Set("clinic_ids", []uint64{1, 2})
	c.Set("is_system_admin", false)
	c.Set("user_id", "17")
	setOwnersViewOnlyClinic(c, 1)
}

func TestListPets_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("defaults to selected clinic and rejects B without grant", func(t *testing.T) {
		h := newHandlerWithPetSvcHandler(&mockPetServiceHandler{
			listFn: func(context.Context, []uint64, PetListFilters, int, int) ([]model.Pet, int64, error) {
				t.Fatal("must not list selected clinic B without owners:view")
				return nil, 0, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10", http.NoBody)
		membershipABGrantASelectedB(c)
		h.ListPets(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), `"total":0`)
	})

	t.Run("filters mixed clinic_ids to A and returns an empty A list as 200", func(t *testing.T) {
		h := newHandlerWithPetSvcHandler(&mockPetServiceHandler{
			listFn: func(_ context.Context, clinicIDs []uint64, _ PetListFilters, _, _ int) ([]model.Pet, int64, error) {
				assert.Equal(t, []uint64{1}, clinicIDs)
				return []model.Pet{}, 0, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10&clinic_ids=1,2", http.NoBody)
		membershipABGrantASelectedB(c)
		h.ListPets(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"total":0`)
	})
}

func TestGetPet_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithPetSvcHandler(&mockPetServiceHandler{
		getByIDForClinicsFn: func(_ context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error) {
			require.Equal(t, []uint64{1}, clinicIDs)
			return &model.Pet{ID: id, ClinicID: 1, Name: "Aペット"}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	membershipABGrantASelectedB(c)
	h.GetPet(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"name":"Aペット"`)
}

func TestListOwnerReportPets_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithPetSvcHandler(&ownerReportPetServiceStub{
		listFn: func(_ context.Context, clinicIDs []uint64, ownerID uint64) ([]model.Pet, error) {
			require.Equal(t, []uint64{1}, clinicIDs)
			assert.Equal(t, uint64(41), ownerID)
			return []model.Pet{}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "41"}}
	membershipABGrantASelectedB(c)
	h.ListOwnerReportPets(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), `"clinic_id":2`)
}
