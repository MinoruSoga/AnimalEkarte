package trimming

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type routePermissionTuple struct {
	resource string
	action   string
}

func TestRegisterRoutes_SnapshotAndRBAC(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var permissions []routePermissionTuple
	requirePermission := func(resource, action string) gin.HandlerFunc {
		permissions = append(permissions, routePermissionTuple{resource: resource, action: action})
		return func(*gin.Context) {}
	}

	h := NewHandlerWithPermission(nil, nil, nil, nil, requirePermission)
	router := gin.New()
	h.RegisterRoutes(router.Group("/api/v1"))

	lines := make([]string, 0, len(router.Routes()))
	for _, route := range router.Routes() {
		lines = append(lines, fmt.Sprintf("%s %s %s", route.Method, route.Path, trimmingRouteHandlerName(route.Handler)))
	}
	sort.Strings(lines)

	wantRoutes := "" +
		"DELETE /api/v1/masters/trimming-course-types/:id DeleteTrimmingCourseType\n" +
		"DELETE /api/v1/masters/trimming-courses/:id DeleteTrimmingCourse\n" +
		"DELETE /api/v1/masters/trimming-options/:id DeleteTrimmingOption\n" +
		"DELETE /api/v1/trimmings/:id DeleteTrimming\n" +
		"GET /api/v1/masters/trimming-course-types ListTrimmingCourseTypes\n" +
		"GET /api/v1/masters/trimming-course-types/:id GetTrimmingCourseType\n" +
		"GET /api/v1/masters/trimming-courses ListTrimmingCourses\n" +
		"GET /api/v1/masters/trimming-courses/:id GetTrimmingCourse\n" +
		"GET /api/v1/masters/trimming-options ListTrimmingOptions\n" +
		"GET /api/v1/masters/trimming-options/:id GetTrimmingOption\n" +
		"GET /api/v1/trimmings ListTrimmings\n" +
		"GET /api/v1/trimmings/:id GetTrimming\n" +
		"PATCH /api/v1/masters/trimming-course-types/:id UpdateTrimmingCourseType\n" +
		"PATCH /api/v1/masters/trimming-course-types/reorder ReorderTrimmingCourseTypes\n" +
		"PATCH /api/v1/masters/trimming-courses/:id UpdateTrimmingCourse\n" +
		"PATCH /api/v1/masters/trimming-courses/reorder ReorderTrimmingCourses\n" +
		"PATCH /api/v1/masters/trimming-options/:id UpdateTrimmingOption\n" +
		"PATCH /api/v1/masters/trimming-options/reorder ReorderTrimmingOptions\n" +
		"PATCH /api/v1/trimmings/:id UpdateTrimming\n" +
		"POST /api/v1/masters/trimming-course-types CreateTrimmingCourseType\n" +
		"POST /api/v1/masters/trimming-courses CreateTrimmingCourse\n" +
		"POST /api/v1/masters/trimming-options CreateTrimmingOption\n" +
		"POST /api/v1/trimmings CreateTrimming\n"

	assert.Equal(t, wantRoutes, strings.Join(lines, "\n")+"\n")
	assert.Equal(t, []routePermissionTuple{
		{resource: "trimming", action: "view"},
		{resource: "trimming", action: "view"},
		{resource: "trimming", action: "create"},
		{resource: "trimming", action: "edit"},
		{resource: "trimming", action: "delete"},
		{resource: "master-trimming", action: "view"},
		{resource: "master-trimming", action: "create"},
		{resource: "master-trimming", action: "edit"},
		{resource: "master-trimming", action: "view"},
		{resource: "master-trimming", action: "edit"},
		{resource: "master-trimming", action: "delete"},
		{resource: "master-trimming", action: "view"},
		{resource: "master-trimming", action: "create"},
		{resource: "master-trimming", action: "edit"},
		{resource: "master-trimming", action: "view"},
		{resource: "master-trimming", action: "edit"},
		{resource: "master-trimming", action: "delete"},
		{resource: "master-trimming", action: "view"},
		{resource: "master-trimming", action: "create"},
		{resource: "master-trimming", action: "edit"},
		{resource: "master-trimming", action: "view"},
		{resource: "master-trimming", action: "edit"},
		{resource: "master-trimming", action: "delete"},
	}, permissions)
}

func trimmingRouteHandlerName(fullName string) string {
	name := strings.TrimSuffix(fullName, "-fm")
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return name
}
