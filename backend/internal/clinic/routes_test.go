package clinic

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_RegisterRoutes_UsesInjectedPermissionMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var permissions []string
	requirePermission := func(resource, action string) gin.HandlerFunc {
		permissions = append(permissions, resource+":"+action)
		return func(c *gin.Context) { c.Next() }
	}

	handler := NewHandler(nil, nil, nil, nil, requirePermission)
	router := gin.New()
	api := router.Group("/api/v1")
	handler.RegisterClinicRoutes(api)
	handler.RegisterClinicHolidayRoutes(api)
	handler.RegisterCompanyRoutes(api)
	handler.RegisterClosingSettingsRoutes(api)

	routes := make(map[string]struct{})
	for _, route := range router.Routes() {
		routes[fmt.Sprintf("%s %s", route.Method, route.Path)] = struct{}{}
	}

	expectedRoutes := []string{
		"GET /api/v1/clinics",
		"GET /api/v1/clinics/:clinic_id",
		"POST /api/v1/clinics",
		"PATCH /api/v1/clinics/:clinic_id",
		"DELETE /api/v1/clinics/:clinic_id",
		"GET /api/v1/clinic-holidays",
		"POST /api/v1/clinic-holidays",
		"DELETE /api/v1/clinic-holidays/:date",
		"GET /api/v1/company",
		"PATCH /api/v1/company",
		"GET /api/v1/closing-settings",
		"PATCH /api/v1/closing-settings",
		"GET /api/v1/closing-settings/special-periods",
		"POST /api/v1/closing-settings/special-periods",
		"PATCH /api/v1/closing-settings/special-periods/:id",
		"DELETE /api/v1/closing-settings/special-periods/:id",
		"GET /api/v1/closing-settings/holidays",
		"POST /api/v1/closing-settings/holidays",
		"DELETE /api/v1/closing-settings/holidays/:date",
	}
	for _, expected := range expectedRoutes {
		assert.Contains(t, routes, expected)
	}
	assert.Len(t, routes, len(expectedRoutes))

	expectedPermissions := []string{
		"hospital-settings:view",
		"hospital-settings:view",
		"hospital-settings:create",
		"hospital-settings:edit",
		"hospital-settings:delete",
		"shifts:view",
		"shifts:create",
		"shifts:delete",
		"hospital-settings:view",
		"hospital-settings:edit",
		"closing-settings:view",
		"closing-settings:edit",
		"closing-settings:view",
		"closing-settings:create",
		"closing-settings:edit",
		"closing-settings:delete",
		"closing-settings:view",
		"closing-settings:create",
		"closing-settings:delete",
	}
	require.Equal(t, expectedPermissions, permissions)
}
