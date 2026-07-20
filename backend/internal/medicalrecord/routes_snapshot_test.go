package medicalrecord

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRegisterRoutes_Snapshot is medicalrecord's half of the BE9-2B/2C route-snapshot
// regression check (see internal/handler/handler_routes_snapshot_test.go's file header for
// the pattern and internal/manualarticle/routes_snapshot_test.go for the precedent this
// mirrors). Before this batch these 25 routes were captured inside
// internal/handler/testdata/route_snapshot.golden as part of *handler.Handler.RegisterRoutes
// (via RegisterMasterRoutes); that golden file was updated to drop them (they moved here —
// same routes, same methods, same handler names, just registered by
// medicalrecord.Handler.RegisterRoutes instead of *handler.Handler).
func TestRegisterRoutes_Snapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	noopPermission := func(_, _ string) gin.HandlerFunc {
		return func(c *gin.Context) {}
	}
	h := NewHandler(
		NewDiagnosisHandler(nil, nil),
		NewExamTypeHandler(nil),
		NewChiefComplaintHandler(nil),
		noopPermission,
	)

	r := gin.New()
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	lines := make([]string, 0, len(r.Routes()))
	for _, route := range r.Routes() {
		lines = append(lines, fmt.Sprintf("%s %s %s", route.Method, route.Path, lastHandlerSegment(route.Handler)))
	}
	sort.Strings(lines)
	got := strings.Join(lines, "\n") + "\n"

	want := "" +
		"DELETE /api/v1/masters/chief-complaint-types/:id DeleteChiefComplaint\n" +
		"DELETE /api/v1/masters/diagnosis-names/:id DeleteDiagnosisName\n" +
		"DELETE /api/v1/masters/diagnosis-types/:id DeleteDiagnosisType\n" +
		"DELETE /api/v1/masters/examination-types/:id DeleteExaminationType\n" +
		"GET /api/v1/masters/chief-complaint-types ListChiefComplaints\n" +
		"GET /api/v1/masters/chief-complaint-types/:id GetChiefComplaint\n" +
		"GET /api/v1/masters/diagnosis-names ListDiagnosisNames\n" +
		"GET /api/v1/masters/diagnosis-names/:id GetDiagnosisName\n" +
		"GET /api/v1/masters/diagnosis-names/all ListDiagnosisNamesAll\n" +
		"GET /api/v1/masters/diagnosis-types ListDiagnosisTypes\n" +
		"GET /api/v1/masters/diagnosis-types/:id GetDiagnosisType\n" +
		"GET /api/v1/masters/examination-types ListExaminationTypes\n" +
		"GET /api/v1/masters/examination-types/:id GetExaminationType\n" +
		"PATCH /api/v1/masters/chief-complaint-types/:id UpdateChiefComplaint\n" +
		"PATCH /api/v1/masters/chief-complaint-types/reorder ReorderChiefComplaints\n" +
		"PATCH /api/v1/masters/diagnosis-names/:id UpdateDiagnosisName\n" +
		"PATCH /api/v1/masters/diagnosis-names/reorder ReorderDiagnosisNames\n" +
		"PATCH /api/v1/masters/diagnosis-types/:id UpdateDiagnosisType\n" +
		"PATCH /api/v1/masters/diagnosis-types/reorder ReorderDiagnosisTypes\n" +
		"PATCH /api/v1/masters/examination-types/:id UpdateExaminationType\n" +
		"PATCH /api/v1/masters/examination-types/reorder ReorderExaminationTypes\n" +
		"POST /api/v1/masters/chief-complaint-types CreateChiefComplaint\n" +
		"POST /api/v1/masters/diagnosis-names CreateDiagnosisName\n" +
		"POST /api/v1/masters/diagnosis-types CreateDiagnosisType\n" +
		"POST /api/v1/masters/examination-types CreateExaminationType\n"

	assert.Equal(t, want, got, "medicalrecord master-CRUD route snapshot drifted from the pre-move baseline "+
		"(internal/handler/testdata/route_snapshot.golden, before these 25 master lines were removed)")
}

// lastHandlerSegment mirrors internal/handler/handler_routes_snapshot_test.go's helper of the
// same name: gin's RouteInfo.Handler returns a fully-qualified function name
// (".../medicalrecord.(*DiagnosisHandler).GetDiagnosisType-fm"); only the trailing method
// name is stable across environments (GOPATH/module cache absolute paths are not).
func lastHandlerSegment(fullName string) string {
	name := strings.TrimSuffix(fullName, "-fm")
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return name
}
