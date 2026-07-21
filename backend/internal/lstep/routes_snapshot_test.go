package lstep

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// lastHandlerSegment は internal/medicalrecord の同名ヘルパーと同一実装。
func lastHandlerSegment(fullName string) string {
	name := strings.TrimSuffix(fullName, "-fm")
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return name
}

// TestRegisterRoutes_Snapshot は lstep 側の BE9-2C route-snapshot 回帰チェック
// （medicalrecord/reservation/billing の先例を踏襲）。L① で
// internal/handler/testdata/route_snapshot.golden から 8 route（lstep-settings 4 +
// clinics エイリアス 4）を drop し、本 package の RegisterRoutes が登録する。
func TestRegisterRoutes_Snapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	noopPermission := func(_, _ string) gin.HandlerFunc {
		return func(c *gin.Context) {}
	}
	h := NewHandler(NewLstepSettingsHandler(nil, noopPermission), noopPermission)

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
		"DELETE /api/v1/clinics/:clinic_id/lstep-settings DeleteLstepSettings\n" +
		"DELETE /api/v1/lstep-settings DeleteLstepSettings\n" +
		"GET /api/v1/clinics/:clinic_id/lstep-settings GetLstepSettings\n" +
		"GET /api/v1/lstep-settings GetLstepSettings\n" +
		"PATCH /api/v1/clinics/:clinic_id/lstep-settings UpdateLstepSettings\n" +
		"PATCH /api/v1/lstep-settings UpdateLstepSettings\n" +
		"POST /api/v1/clinics/:clinic_id/lstep-settings/test-connection TestLstepConnection\n" +
		"POST /api/v1/lstep-settings/test-connection TestLstepConnection\n"

	assert.Equal(t, want, got)
}
