package inventory

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type permissionTuple struct {
	resource string
	action   string
}

func TestRegisterRoutes_SnapshotAndRBAC(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var permissions []permissionTuple
	requirePermission := func(resource, action string) gin.HandlerFunc {
		permissions = append(permissions, permissionTuple{resource: resource, action: action})
		return func(*gin.Context) {}
	}

	h := NewHandler(nil, nil, requirePermission)
	router := gin.New()
	h.RegisterRoutes(router.Group("/api/v1"))

	lines := make([]string, 0, len(router.Routes()))
	for _, route := range router.Routes() {
		lines = append(lines, fmt.Sprintf("%s %s %s", route.Method, route.Path, inventoryRouteHandlerName(route.Handler)))
	}
	sort.Strings(lines)

	wantRoutes := "" +
		"DELETE /api/v1/inventory/:id DeleteInventory\n" +
		"DELETE /api/v1/masters/merchandise-items/:id DeleteMerchandiseItem\n" +
		"GET /api/v1/inventory ListInventory\n" +
		"GET /api/v1/inventory/:id GetInventory\n" +
		"GET /api/v1/masters/merchandise-items ListMerchandiseItems\n" +
		"GET /api/v1/masters/merchandise-items/:id GetMerchandiseItem\n" +
		"PATCH /api/v1/inventory/:id UpdateInventory\n" +
		"PATCH /api/v1/masters/merchandise-items/:id UpdateMerchandiseItem\n" +
		"PATCH /api/v1/masters/merchandise-items/reorder ReorderMerchandiseItems\n" +
		"POST /api/v1/inventory CreateInventory\n" +
		"POST /api/v1/masters/merchandise-items CreateMerchandiseItem\n"

	assert.Equal(t, wantRoutes, strings.Join(lines, "\n")+"\n")
	assert.Equal(t, []permissionTuple{
		{resource: "inventory", action: "view"},
		{resource: "inventory", action: "view"},
		{resource: "inventory", action: "create"},
		{resource: "inventory", action: "edit"},
		{resource: "inventory", action: "delete"},
		{resource: "master-merchandise", action: "view"},
		{resource: "master-merchandise", action: "create"},
		{resource: "master-merchandise", action: "edit"},
		{resource: "master-merchandise", action: "view"},
		{resource: "master-merchandise", action: "edit"},
		{resource: "master-merchandise", action: "delete"},
	}, permissions)
}

func inventoryRouteHandlerName(fullName string) string {
	name := strings.TrimSuffix(fullName, "-fm")
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return name
}
