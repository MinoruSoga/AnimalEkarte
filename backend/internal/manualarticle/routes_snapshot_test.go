package manualarticle

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRegisterRoutes_Snapshot is manualarticle's half of the BE9-2B route-snapshot regression
// check (see internal/handler/handler_routes_snapshot_test.go's file header for the pattern
// and rationale). Before BE9-2B these 5 routes were captured inside
// internal/handler/testdata/route_snapshot.golden as part of *handler.Handler.RegisterRoutes;
// that golden file was updated to drop them (they moved here — same routes, same methods, same
// handler names, just registered by manualarticle.Handler.RegisterRoutes instead of
// *handler.Handler). This test proves the moved routes still exist with the exact same
// (method, path, handler) triples.
func TestRegisterRoutes_Snapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	noopPermission := func(_, _ string) gin.HandlerFunc {
		return func(c *gin.Context) {}
	}
	h := NewHandler(nil, nil, noopPermission)

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
		"DELETE /api/v1/manual/articles/:category/:slug DeleteManualArticle\n" +
		"GET /api/v1/manual/articles ListManualArticles\n" +
		"GET /api/v1/manual/articles/:category/:slug GetManualArticle\n" +
		"GET /api/v1/manual/articles/:category/:slug/versions ListManualArticleVersions\n" +
		"PUT /api/v1/manual/articles/:category/:slug UpsertManualArticle\n"

	assert.Equal(t, want, got, "manualarticle route snapshot drifted from the pre-BE9-2B baseline "+
		"(internal/handler/testdata/route_snapshot.golden, before the 5 manual/articles lines were removed)")
}

// lastHandlerSegment mirrors internal/handler/handler_routes_snapshot_test.go's helper of the
// same name: gin's RouteInfo.Handler returns a fully-qualified function name
// (".../manualarticle.(*Handler).GetManualArticle-fm"); only the trailing method name is
// stable across environments (GOPATH/module cache absolute paths are not).
func lastHandlerSegment(fullName string) string {
	name := strings.TrimSuffix(fullName, "-fm")
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return name
}
