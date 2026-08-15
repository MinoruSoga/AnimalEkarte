package staff

import (
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

type routePermissionTuple struct {
	Method   string
	Path     string
	Resource string
	Action   string
}

func TestHandler_RegisterRoutesPinsAllLegacyRoutesAndPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	protected := router.Group("/api/v1")
	permissionCalls := make([]routePermissionTuple, 0, 31)
	handler := NewHandler(nil, nil, nil, nil, nil, func(resource, action string) gin.HandlerFunc {
		permissionCalls = append(permissionCalls, routePermissionTuple{
			Resource: resource,
			Action:   action,
		})
		return func(c *gin.Context) { c.Next() }
	})

	handler.RegisterRoutes(protected)

	expectedPermissions := []routePermissionTuple{
		{Method: "GET", Path: "/api/v1/masters/staffs", Resource: string(model.ResourceMasterStaff), Action: "view"},
		{Method: "POST", Path: "/api/v1/masters/staffs", Resource: string(model.ResourceMasterStaff), Action: "create"},
		{Method: "PATCH", Path: "/api/v1/masters/staffs/reorder", Resource: string(model.ResourceMasterStaff), Action: "edit"},
		{Method: "GET", Path: "/api/v1/masters/staffs/:id", Resource: string(model.ResourceMasterStaff), Action: "view"},
		{Method: "PATCH", Path: "/api/v1/masters/staffs/:id", Resource: string(model.ResourceMasterStaff), Action: "edit"},
		{Method: "DELETE", Path: "/api/v1/masters/staffs/:id", Resource: string(model.ResourceMasterStaff), Action: "delete"},
		{Method: "GET", Path: "/api/v1/masters/staffs/:id/permission-groups", Resource: string(model.ResourceMasterStaff), Action: "view"},
		{Method: "PUT", Path: "/api/v1/masters/staffs/:id/permission-groups", Resource: string(model.ResourceMasterStaff), Action: "edit"},
		{Method: "PUT", Path: "/api/v1/masters/staffs/:id/permission-groups", Resource: string(model.ResourceMasterPermission), Action: "edit"},
		{Method: "GET", Path: "/api/v1/masters/staffs/:id/clinics", Resource: string(model.ResourceMasterStaff), Action: "view"},
		{Method: "PUT", Path: "/api/v1/masters/staffs/:id/clinics", Resource: string(model.ResourceMasterStaff), Action: "edit"},
		{Method: "GET", Path: "/api/v1/masters/staffs/:id/excluded-reservation-types", Resource: string(model.ResourceMasterStaff), Action: "view"},
		{Method: "PUT", Path: "/api/v1/masters/staffs/:id/excluded-reservation-types", Resource: string(model.ResourceMasterStaff), Action: "edit"},
		{Method: "GET", Path: "/api/v1/masters/staffs/:id/capable-reservation-types", Resource: string(model.ResourceMasterStaff), Action: "view"},
		{Method: "PUT", Path: "/api/v1/masters/staffs/:id/capable-reservation-types", Resource: string(model.ResourceMasterStaff), Action: "edit"},
		{Method: "GET", Path: "/api/v1/masters/occupations", Resource: string(model.ResourceMasterStaff), Action: "view"},
		{Method: "POST", Path: "/api/v1/masters/occupations", Resource: string(model.ResourceMasterStaff), Action: "create"},
		{Method: "PATCH", Path: "/api/v1/masters/occupations/reorder", Resource: string(model.ResourceMasterStaff), Action: "edit"},
		{Method: "GET", Path: "/api/v1/masters/occupations/:id", Resource: string(model.ResourceMasterStaff), Action: "view"},
		{Method: "PATCH", Path: "/api/v1/masters/occupations/:id", Resource: string(model.ResourceMasterStaff), Action: "edit"},
		{Method: "DELETE", Path: "/api/v1/masters/occupations/:id", Resource: string(model.ResourceMasterStaff), Action: "delete"},
		{Method: "GET", Path: "/api/v1/shifts", Resource: string(model.ResourceShifts), Action: "view"},
		{Method: "GET", Path: "/api/v1/shifts/on-duty-staffs", Resource: string(model.ResourceShifts), Action: "view"},
		{Method: "POST", Path: "/api/v1/shifts", Resource: string(model.ResourceShifts), Action: "create"},
		{Method: "PATCH", Path: "/api/v1/shifts/:id", Resource: string(model.ResourceShifts), Action: "edit"},
		{Method: "DELETE", Path: "/api/v1/shifts/:id", Resource: string(model.ResourceShifts), Action: "delete"},
		{Method: "GET", Path: "/api/v1/shift-templates", Resource: string(model.ResourceShifts), Action: "view"},
		{Method: "POST", Path: "/api/v1/shift-templates", Resource: string(model.ResourceShifts), Action: "create"},
		{Method: "PATCH", Path: "/api/v1/shift-templates/reorder", Resource: string(model.ResourceShifts), Action: "edit"},
		{Method: "GET", Path: "/api/v1/shift-templates/:id", Resource: string(model.ResourceShifts), Action: "view"},
		{Method: "PATCH", Path: "/api/v1/shift-templates/:id", Resource: string(model.ResourceShifts), Action: "edit"},
		{Method: "DELETE", Path: "/api/v1/shift-templates/:id", Resource: string(model.ResourceShifts), Action: "delete"},
	}

	require.Len(t, permissionCalls, len(expectedPermissions))
	for index := range expectedPermissions {
		assert.Equal(t, expectedPermissions[index].Resource, permissionCalls[index].Resource)
		assert.Equal(t, expectedPermissions[index].Action, permissionCalls[index].Action)
	}

	gotRoutes := make([]string, 0, len(router.Routes()))
	for _, route := range router.Routes() {
		gotRoutes = append(gotRoutes, route.Method+" "+route.Path)
	}
	wantRouteSet := make(map[string]struct{}, len(expectedPermissions))
	for _, tuple := range expectedPermissions {
		wantRouteSet[tuple.Method+" "+tuple.Path] = struct{}{}
	}
	wantRoutes := make([]string, 0, len(wantRouteSet))
	for route := range wantRouteSet {
		wantRoutes = append(wantRoutes, route)
	}
	sort.Strings(gotRoutes)
	sort.Strings(wantRoutes)
	assert.Equal(t, wantRoutes, gotRoutes)
}
