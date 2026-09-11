package owner

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
	setOwnersPermissionOnlyClinic(c, 1, "view")
}

func TestListOwners_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("defaults to selected clinic and rejects B without grant", func(t *testing.T) {
		h := newHandlerWithOwnerSvc(&mockOwnerService{
			listFn: func(context.Context, []uint64, int, int, string) ([]model.Owner, int64, error) {
				t.Fatal("must not list selected clinic B without owners:view")
				return nil, 0, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10", http.NoBody)
		membershipABGrantASelectedB(c)
		h.ListOwners(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), `"total":0`)
	})

	t.Run("filters mixed clinic_ids to A and returns an empty A list as 200", func(t *testing.T) {
		h := newHandlerWithOwnerSvc(&mockOwnerService{
			listFn: func(_ context.Context, clinicIDs []uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
				assert.Equal(t, []uint64{1}, clinicIDs)
				return []model.Owner{}, 0, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10&clinic_ids=1,2", http.NoBody)
		membershipABGrantASelectedB(c)
		h.ListOwners(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"total":0`)
	})

	t.Run("rejects explicit B clinic_ids", func(t *testing.T) {
		h := newHandlerWithOwnerSvc(&mockOwnerService{
			listFn: func(context.Context, []uint64, int, int, string) ([]model.Owner, int64, error) {
				t.Fatal("must not list B")
				return nil, 0, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10&clinic_ids=2", http.NoBody)
		membershipABGrantASelectedB(c)
		h.ListOwners(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestGetOwner_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithOwnerSvc(&mockOwnerService{
		getByIDForClinicsFn: func(_ context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error) {
			require.Equal(t, []uint64{1}, clinicIDs)
			assert.Equal(t, uint64(42), id)
			return &model.Owner{ID: 42, ClinicID: 1, Name: "A飼主"}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "42"}}
	membershipABGrantASelectedB(c)
	h.GetOwner(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"owner_name":"A飼主"`)
	assert.NotContains(t, w.Body.String(), `"clinic_id":2`)
}
