package lstep

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterSharedFileRoutes_RBACTuples(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var single []PermissionRequirement
	var anyRequirements []PermissionRequirement
	h := &Handler{
		requirePermission: func(resource, action string) gin.HandlerFunc {
			single = append(single, PermissionRequirement{Resource: resource, Action: action})
			return func(c *gin.Context) { c.Next() }
		},
		requireAnyPermission: func(requirements ...PermissionRequirement) gin.HandlerFunc {
			anyRequirements = append(anyRequirements, requirements...)
			return func(c *gin.Context) { c.Next() }
		},
	}
	r := gin.New()
	h.RegisterSharedFileRoutes(r.Group("/api/v1"))

	assert.Equal(t, []PermissionRequirement{
		{Resource: "hospital-settings", Action: "view"},
		{Resource: "hospital-settings", Action: "view"},
		{Resource: "hospital-settings", Action: "delete"},
	}, single)
	assert.Equal(t, []PermissionRequirement{
		{Resource: "owners", Action: "edit"},
		{Resource: "medical-records", Action: "create"},
		{Resource: "medical-records", Action: "edit"},
	}, anyRequirements)

	routes := map[string]string{}
	for _, route := range r.Routes() {
		if route.Path == "/api/v1/shared-files" || route.Path == "/api/v1/shared-files/:id/signed-url" || route.Path == "/api/v1/shared-files/:id" {
			routes[route.Method+" "+route.Path] = route.Handler
		}
	}
	require.Len(t, routes, 4)
	assert.Contains(t, routes, http.MethodGet+" /api/v1/shared-files")
	assert.Contains(t, routes, http.MethodPost+" /api/v1/shared-files")
	assert.Contains(t, routes, http.MethodGet+" /api/v1/shared-files/:id/signed-url")
	assert.Contains(t, routes, http.MethodDelete+" /api/v1/shared-files/:id")
}
