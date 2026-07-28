package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

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
