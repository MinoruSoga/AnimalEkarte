package owner

import (
	"net/http"
	"slices"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ownerPermissionCall struct {
	resource string
	action   string
}

func TestHandler_RegisterRoutes_PreservesOwnerContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var permissions []ownerPermissionCall
	requirePermission := func(resource, action string) gin.HandlerFunc {
		permissions = append(permissions, ownerPermissionCall{resource: resource, action: action})
		return func(c *gin.Context) { c.Next() }
	}

	router := gin.New()
	group := router.Group("/api/v1")
	NewHandler(nil, nil, requirePermission, nil).RegisterRoutes(group)

	gotRoutes := make([]string, 0, len(router.Routes()))
	for _, route := range router.Routes() {
		gotRoutes = append(gotRoutes, route.Method+" "+route.Path)
	}
	slices.Sort(gotRoutes)

	wantRoutes := []string{
		http.MethodDelete + " /api/v1/owners/:id",
		http.MethodGet + " /api/v1/owners",
		http.MethodGet + " /api/v1/owners/:id",
		http.MethodPatch + " /api/v1/owners/:id",
		http.MethodPatch + " /api/v1/owners/:id/delivery-caution",
		http.MethodPatch + " /api/v1/owners/:id/delivery-exclusion",
		http.MethodPatch + " /api/v1/owners/:id/line",
		http.MethodPatch + " /api/v1/owners/:id/line-id-confirm",
		http.MethodPatch + " /api/v1/owners/:id/line-user-id",
		http.MethodPatch + " /api/v1/owners/:id/transfer-status",
		http.MethodPost + " /api/v1/owners",
	}
	slices.Sort(wantRoutes)

	assert.Equal(t, wantRoutes, gotRoutes)
	assert.Equal(t, []ownerPermissionCall{
		{resource: "owners", action: "view"},
		{resource: "owners", action: "view"},
		{resource: "owners", action: "create"},
		{resource: "owners", action: "edit"},
		{resource: "owners", action: "delete"},
		{resource: "owners", action: "edit"},
		{resource: "owners", action: "edit"},
		{resource: "owners", action: "edit"},
		{resource: "owners", action: "edit"},
		{resource: "owners", action: "edit"},
		{resource: "owners", action: "edit"},
	}, permissions)
	require.Len(t, gotRoutes, 11)
}
