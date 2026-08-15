package pet

import (
	"net/http"
	"slices"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type permissionCall struct {
	resource string
	action   string
}

func TestRegisterRoutes_PreservesPetSpeciesAndChronicConditionContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var permissions []permissionCall
	requirePermission := func(resource, action string) gin.HandlerFunc {
		permissions = append(permissions, permissionCall{resource: resource, action: action})
		return func(c *gin.Context) { c.Next() }
	}

	router := gin.New()
	group := router.Group("/api/v1")
	NewHandler(nil, nil, nil, requirePermission).RegisterRoutes(group)

	gotRoutes := make([]string, 0, len(router.Routes()))
	for _, route := range router.Routes() {
		gotRoutes = append(gotRoutes, route.Method+" "+route.Path)
	}
	slices.Sort(gotRoutes)

	wantRoutes := []string{
		http.MethodDelete + " /api/v1/masters/animal-species/:id",
		http.MethodDelete + " /api/v1/pets/:id",
		http.MethodDelete + " /api/v1/pets/:id/chronic-conditions/:cc_id",
		http.MethodGet + " /api/v1/masters/animal-species",
		http.MethodGet + " /api/v1/masters/animal-species/:id",
		http.MethodGet + " /api/v1/owners/:id/report/pets",
		http.MethodGet + " /api/v1/owners/:id/shared-pets",
		http.MethodGet + " /api/v1/pets",
		http.MethodGet + " /api/v1/pets/:id",
		http.MethodGet + " /api/v1/pets/:id/chronic-conditions",
		http.MethodGet + " /api/v1/pets/:id/first-visit",
		http.MethodGet + " /api/v1/pets/:id/sub-owners",
		http.MethodPatch + " /api/v1/masters/animal-species/:id",
		http.MethodPatch + " /api/v1/masters/animal-species/reorder",
		http.MethodPatch + " /api/v1/pets/:id",
		http.MethodPatch + " /api/v1/pets/:id/chronic-conditions/:cc_id",
		http.MethodPost + " /api/v1/masters/animal-species",
		http.MethodPost + " /api/v1/pets",
		http.MethodPost + " /api/v1/pets/:id/chronic-conditions",
		http.MethodPut + " /api/v1/pets/:id/sub-owners",
	}
	slices.Sort(wantRoutes)
	assert.Equal(t, wantRoutes, gotRoutes)

	assert.Equal(t, []permissionCall{
		{resource: "owners", action: "view"},
		{resource: "owners", action: "view"},
		{resource: "owners", action: "view"},
		{resource: "owners", action: "view"},
		{resource: "owners", action: "create"},
		{resource: "owners", action: "edit"},
		{resource: "owners", action: "delete"},
		{resource: "medical-records", action: "view"},
		{resource: "owners", action: "view"},
		{resource: "owners", action: "edit"},
		{resource: "owners", action: "view"},
		{resource: "owners", action: "create"},
		{resource: "owners", action: "edit"},
		{resource: "owners", action: "delete"},
		{resource: "master-animal-species", action: "view"},
		{resource: "master-animal-species", action: "create"},
		{resource: "master-animal-species", action: "edit"},
		{resource: "master-animal-species", action: "view"},
		{resource: "master-animal-species", action: "edit"},
		{resource: "master-animal-species", action: "delete"},
	}, permissions)

	require.Len(t, gotRoutes, 20)
}

func TestRegisterRoutes_OwnerSharedPetsRouteUsesOwnerViewPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var permissions []permissionCall
	requirePermission := func(resource, action string) gin.HandlerFunc {
		permissions = append(permissions, permissionCall{resource: resource, action: action})
		return func(c *gin.Context) { c.Next() }
	}

	router := gin.New()
	handler := NewHandler(nil, nil, nil, requirePermission)
	handler.registerOwnerSharedPetRoutes(router.Group("/api/v1"))

	gotRoutes := make([]string, 0, len(router.Routes()))
	for _, route := range router.Routes() {
		gotRoutes = append(gotRoutes, route.Method+" "+route.Path)
	}
	assert.Equal(t, []string{
		http.MethodGet + " /api/v1/owners/:id/shared-pets",
	}, gotRoutes)
	assert.Equal(t, []permissionCall{
		{resource: "owners", action: "view"},
	}, permissions)
}

func TestRegisterRoutes_PetOwnerRoutesUseOwnerPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var permissions []permissionCall
	requirePermission := func(resource, action string) gin.HandlerFunc {
		permissions = append(permissions, permissionCall{resource: resource, action: action})
		return func(c *gin.Context) { c.Next() }
	}

	router := gin.New()
	handler := NewHandler(nil, nil, nil, requirePermission)
	handler.registerPetOwnerRoutes(router.Group("/api/v1/pets"))

	gotRoutes := make([]string, 0, len(router.Routes()))
	for _, route := range router.Routes() {
		gotRoutes = append(gotRoutes, route.Method+" "+route.Path)
	}
	assert.ElementsMatch(t, []string{
		http.MethodGet + " /api/v1/pets/:id/sub-owners",
		http.MethodPut + " /api/v1/pets/:id/sub-owners",
	}, gotRoutes)
	assert.Equal(t, []permissionCall{
		{resource: "owners", action: "view"},
		{resource: "owners", action: "edit"},
	}, permissions)
}
