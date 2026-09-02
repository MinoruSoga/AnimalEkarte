package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

func authPermissionContext(t *testing.T) *gin.Context {
	t.Helper()
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Set("is_system_admin", false)
	c.Set("user_id", "17")
	c.Set("clinic_id", "23")
	return c
}

func TestHTTPHandler_HasPermission_RepositoryErrorFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHTTPHandler(HTTPDependencies{
		EffectivePermissions: authServiceEffectivePermissionStub{
			getFn: func(_ context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error) {
				assert.Equal(t, uint64(17), staffID)
				assert.Equal(t, uint64(23), clinicID)
				return nil, errors.New("permission repository unavailable")
			},
		},
	}, CookieConfigForProduction(false))

	assert.False(t, handler.HasPermission(
		authPermissionContext(t),
		string(model.ResourceDiscount),
		"edit",
	))
}

func TestHTTPHandler_HasPermission_SystemAdminBypassesRepository(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHTTPHandler(HTTPDependencies{}, CookieConfigForProduction(false))
	c := authPermissionContext(t)
	c.Set("is_system_admin", true)

	assert.True(t, handler.HasPermission(c, "any-resource", "any-action"))
}

// TestHTTPHandler_HasPermission_MissingContextDoesNotWriteResponse は AUS-09:
// Extract* 相当の欠落時に HasPermission がレスポンスを書かず、RequirePermission が1回だけ 403 を書く。
func TestHTTPHandler_HasPermission_MissingContextDoesNotWriteResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHTTPHandler(HTTPDependencies{}, CookieConfigForProduction(false))

	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	// intentionally no is_system_admin / user_id / clinic_id

	assert.False(t, handler.HasPermission(c, "owners", "view"))
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Empty(t, response.Body.String())
	assert.False(t, c.Writer.Written())

	handler.RequirePermission("owners", "view")(c)
	assert.Equal(t, http.StatusForbidden, response.Code)
	assert.Contains(t, response.Body.String(), "forbidden")
	assert.True(t, c.IsAborted())
}

func TestHTTPHandler_HasPermissionInClinic_UsesDestinationClinic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var lookedUpClinic uint64
	handler := NewHTTPHandler(HTTPDependencies{
		EffectivePermissions: authServiceEffectivePermissionStub{
			getFn: func(_ context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error) {
				assert.Equal(t, uint64(17), staffID)
				lookedUpClinic = clinicID
				if clinicID != 23 {
					return nil, nil
				}
				return []model.PermissionGroupRule{{
					Resource: string(model.ResourceOwners),
					CanView:  true,
				}}, nil
			},
		},
	}, CookieConfigForProduction(false))

	c := authPermissionContext(t)
	assert.True(t, handler.HasPermissionInClinic(c, 23, string(model.ResourceOwners), "view"))
	assert.Equal(t, uint64(23), lookedUpClinic)
	assert.False(t, handler.HasPermissionInClinic(c, 99, string(model.ResourceOwners), "view"))
	assert.Equal(t, uint64(99), lookedUpClinic)
	assert.False(t, handler.HasPermissionInClinic(c, 0, string(model.ResourceOwners), "view"))
}

func TestHTTPHandler_RequirePermission_AttachesClinicChecker(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHTTPHandler(HTTPDependencies{
		EffectivePermissions: authServiceEffectivePermissionStub{
			getFn: func(_ context.Context, _, clinicID uint64) ([]model.PermissionGroupRule, error) {
				if clinicID != 23 {
					return nil, nil
				}
				return []model.PermissionGroupRule{{
					Resource:  string(model.ResourceOwners),
					CanView:   true,
					CanCreate: true,
				}}, nil
			},
		},
	}, CookieConfigForProduction(false))

	c := authPermissionContext(t)
	handler.RequirePermission(string(model.ResourceOwners), "view")(c)
	assert.False(t, c.IsAborted())
	check, ok := httpapi.PeekClinicPermissionChecker(c)
	require.True(t, ok)
	assert.True(t, check(c, 23, string(model.ResourceOwners), "view"))
	assert.False(t, check(c, 99, string(model.ResourceOwners), "view"))
}

func TestHTTPHandler_RequirePermission_AllowsGETWhenAnotherAuthorizedClinicHasGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHTTPHandler(HTTPDependencies{
		EffectivePermissions: authServiceEffectivePermissionStub{
			getFn: func(_ context.Context, _, clinicID uint64) ([]model.PermissionGroupRule, error) {
				if clinicID != 99 {
					return nil, nil
				}
				return []model.PermissionGroupRule{{
					Resource:  string(model.ResourceOwners),
					CanView:   true,
					CanCreate: true,
				}}, nil
			},
		},
	}, CookieConfigForProduction(false))

	getCtx := authPermissionContext(t)
	getCtx.Set("clinic_ids", []uint64{23, 99})
	handler.RequirePermission(string(model.ResourceOwners), "view")(getCtx)
	assert.False(t, getCtx.IsAborted())
	check, ok := httpapi.PeekClinicPermissionChecker(getCtx)
	require.True(t, ok)
	assert.False(t, check(getCtx, 23, string(model.ResourceOwners), "view"))
	assert.True(t, check(getCtx, 99, string(model.ResourceOwners), "view"))

	post := httptest.NewRecorder()
	postCtx, _ := gin.CreateTestContext(post)
	postCtx.Request = httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	postCtx.Set("is_system_admin", false)
	postCtx.Set("user_id", "17")
	postCtx.Set("clinic_id", "23")
	postCtx.Set("clinic_ids", []uint64{23, 99})
	handler.RequirePermission(string(model.ResourceOwners), "create")(postCtx)
	assert.True(t, postCtx.IsAborted())
	assert.Equal(t, http.StatusForbidden, post.Code)
}
