package support

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRegisterRoutes_Snapshot pins the support route surface (same pattern as
// manualarticle/routes_snapshot_test.go).
func TestRegisterRoutes_Snapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	noopPermission := func(_, _ string) gin.HandlerFunc {
		return func(c *gin.Context) {}
	}
	h := NewHandler(nil, nil, nil, noopPermission, nil, nil)

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
		"DELETE /api/v1/support/bug-reports/:id DeleteBugReport\n" +
		"DELETE /api/v1/support/chat/history ClearChatHistory\n" +
		"GET /api/v1/support/bug-reports ListBugReports\n" +
		"GET /api/v1/support/chat/history ChatHistory\n" +
		"GET /api/v1/support/chat/status ChatStatus\n" +
		"PATCH /api/v1/support/bug-reports/:id/status UpdateBugReportStatus\n" +
		"POST /api/v1/support/bug-reports CreateBugReport\n" +
		"POST /api/v1/support/bug-reports/:id/plane-ticket CreatePlaneTicket\n" +
		"POST /api/v1/support/chat Chat\n"

	assert.Equal(t, want, got, "support route snapshot drifted")
}

func lastHandlerSegment(fullName string) string {
	name := strings.TrimSuffix(fullName, "-fm")
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return name
}
