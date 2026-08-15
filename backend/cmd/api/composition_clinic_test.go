package main

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/clinic"
)

func TestNewClinicCompositionBuildsDomainHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var permissionCalls int
	var requirePermission clinic.PermissionMiddleware = func(_, _ string) gin.HandlerFunc {
		permissionCalls++
		return func(c *gin.Context) {
			c.Next()
		}
	}

	repositories := newClinicRepositories(nil)
	composition := newClinicComposition(repositories, nil, nil, nil)

	require.Equal(t, repositories, composition.Repositories)
	require.NotNil(t, composition.Clinic)
	require.NotNil(t, composition.ClosingSettings)
	require.Zero(t, permissionCalls)

	handler := composition.newHandler(requirePermission)
	require.NotNil(t, handler)

	router := gin.New()
	protected := router.Group("/api/v1")
	handler.RegisterClinicRoutes(protected)
	handler.RegisterClinicHolidayRoutes(protected)
	handler.RegisterCompanyRoutes(protected)
	handler.RegisterClosingSettingsRoutes(protected)

	require.Len(t, router.Routes(), 19)
	require.Equal(t, 19, permissionCalls)
}
