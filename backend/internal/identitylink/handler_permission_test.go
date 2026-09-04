package identitylink

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestRegisterRoutes_PermissionViewOnlyVsManage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type call struct {
		resource, action string
	}
	var calls []call
	perm := func(resource, action string) gin.HandlerFunc {
		return func(c *gin.Context) {
			calls = append(calls, call{resource, action})
			// Abort before nil service handlers run; we only assert middleware wiring.
			c.AbortWithStatus(http.StatusNoContent)
		}
	}

	// nil service is fine — we only assert middleware registration via a probe that aborts early.
	h := NewHandler(nil, perm)
	r := gin.New()
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	// Collect route method/path/handler by executing with a custom middleware recorder
	// via route table inspection.
	routes := r.Routes()
	require.NotEmpty(t, routes)

	viewPaths := map[string]bool{
		"/api/v1/identity-links/owners/search":                             true,
		"/api/v1/identity-links/pets/search":                               true,
		"/api/v1/identity-links/owner-groups/:id":                          true,
		"/api/v1/identity-links/owners/:clinic_id/:owner_id/group":         true,
		"/api/v1/identity-links/pet-groups/:id":                            true,
		"/api/v1/identity-links/pets/:clinic_id/:pet_id/group":             true,
		"/api/v1/identity-links/pets/:clinic_id/:pet_id/treatment-history": true,
	}
	editPaths := map[string]bool{
		"/api/v1/identity-links/owner-groups":             true,
		"/api/v1/identity-links/owner-groups/:id/members": true,
		"/api/v1/identity-links/pet-groups":               true,
		"/api/v1/identity-links/pet-groups/:id/members":   true,
	}

	// Fire each registered route once to capture permission middleware resource/action.
	for _, route := range routes {
		calls = nil
		req := httptest.NewRequest(route.Method, route.Path, http.NoBody)
		// Replace path params with dummy numbers for gin matching.
		path := route.Path
		// gin route table Path still has :params; use a matching concrete path.
		switch route.Path {
		case "/api/v1/identity-links/owner-groups/:id":
			path = "/api/v1/identity-links/owner-groups/1"
		case "/api/v1/identity-links/owner-groups/:id/members":
			path = "/api/v1/identity-links/owner-groups/1/members"
		case "/api/v1/identity-links/owners/:clinic_id/:owner_id/group":
			path = "/api/v1/identity-links/owners/1/2/group"
		case "/api/v1/identity-links/pet-groups/:id":
			path = "/api/v1/identity-links/pet-groups/1"
		case "/api/v1/identity-links/pet-groups/:id/members":
			path = "/api/v1/identity-links/pet-groups/1/members"
		case "/api/v1/identity-links/pets/:clinic_id/:pet_id/group":
			path = "/api/v1/identity-links/pets/1/2/group"
		case "/api/v1/identity-links/pets/:clinic_id/:pet_id/treatment-history":
			path = "/api/v1/identity-links/pets/1/2/treatment-history"
		}
		req = httptest.NewRequest(route.Method, path, http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.NotEmpty(t, calls, "permission middleware must run for %s %s", route.Method, route.Path)
		assert.Equal(t, string(model.ResourceIdentityLinks), calls[0].resource)

		switch {
		case viewPaths[route.Path]:
			assert.Equal(t, "view", calls[0].action, "GET/read routes require view: %s", route.Path)
		case editPaths[route.Path]:
			assert.Equal(t, "edit", calls[0].action, "mutate routes require edit (manage): %s", route.Path)
		default:
			// DELETE members shares path with POST members
			if route.Method == http.MethodDelete {
				assert.Equal(t, "edit", calls[0].action)
			}
		}
	}
}

func TestRegisterRoutes_Snapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(nil, func(_, _ string) gin.HandlerFunc {
		return func(c *gin.Context) {}
	})
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1"))

	got := map[string]bool{}
	for _, route := range r.Routes() {
		got[route.Method+" "+route.Path] = true
	}
	want := []string{
		"GET /api/v1/identity-links/owners/search",
		"GET /api/v1/identity-links/pets/search",
		"GET /api/v1/identity-links/owner-groups/:id",
		"GET /api/v1/identity-links/owners/:clinic_id/:owner_id/group",
		"POST /api/v1/identity-links/owner-groups",
		"POST /api/v1/identity-links/owner-groups/:id/members",
		"DELETE /api/v1/identity-links/owner-groups/:id/members",
		"GET /api/v1/identity-links/pet-groups/:id",
		"GET /api/v1/identity-links/pets/:clinic_id/:pet_id/group",
		"POST /api/v1/identity-links/pet-groups",
		"POST /api/v1/identity-links/pet-groups/:id/members",
		"DELETE /api/v1/identity-links/pet-groups/:id/members",
		"GET /api/v1/identity-links/pets/:clinic_id/:pet_id/treatment-history",
	}
	for _, w := range want {
		assert.Truef(t, got[w], "missing route %s", w)
	}
}
