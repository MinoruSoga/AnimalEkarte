package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
)

func TestRouteCompositionSmoke_TargetGraphRegistersEverySurface(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Local storage + scheduler token so base routes match docker-compose shape.
	t.Setenv("STORAGE_TYPE", "")
	t.Setenv("SCHEDULER_INTERNAL_TOKEN", "test-scheduler-internal-token-32b!!")

	composition := newRuntimeComposition(runtimeCompositionDependencies{
		Config: &config.Config{
			JWTSecret: "test-secret-for-route-registration",
		},
	})
	router := gin.New()
	require.NoError(
		t,
		composition.registerRoutes(ctx, router, nil, false),
	)

	routes := make(map[string]struct{}, len(router.Routes()))
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}
	// Runtime surface after #239 identitylink routes (11) and StaticFS GET/HEAD.
	// Measured 2026-07-30 red-sweep: 496 unique Method+Path pairs.
	// 2026-08-01: 497 — b65cf69ef added POST /api/v1/estimates/:id/successors
	// (TASK-012 estimate successor drafts), documented in api.yaml and covered
	// by the OpenAPI drift test.
	// 2026-08-04: 503 — six routes landed between b65cf69ef and this bump without
	// the constant being updated. Enumerated so the number is auditable:
	//   74aa3e2c6 (BUG-013) GET  /api/v1/billing-items/unbilled-details
	//   75f8912fc (BUG-018) POST /api/v1/accountings/complete
	//   1dd1cf04e           POST /api/v1/examinations/:id/unconfirm
	//   f609832ce (TASK-031) GET  /api/v1/examinations/:id/print-snapshot
	//   f609832ce (TASK-374) POST /api/v1/checkup-package-imports/preview
	//   f609832ce (TASK-374) POST /api/v1/checkup-package-imports
	// All six are documented in docs/api.yaml and covered by the OpenAPI drift test.
	require.Len(t, routes, 503)
	for _, expected := range []string{
		"GET /health",
		"GET /uploads/*filepath",
		"HEAD /uploads/*filepath",
		"POST /api/v1/login",
		"GET /api/v1/me",
		"GET /api/v1/masters/permission-groups",
		"GET /api/v1/owners",
		"GET /api/v1/pets",
		"GET /api/v1/masters/staffs",
		"GET /api/v1/clinics",
		"GET /api/v1/medical-records",
		"GET /api/v1/reservations",
		"GET /api/v1/accountings",
		"GET /api/v1/lstep-settings",
		"GET /api/liff/:clinicId/courses",
		"POST /api/line/webhook",
		"POST /_internal/scheduled-jobs/:jobAction",
	} {
		assert.Contains(t, routes, expected)
	}

	for _, protectedPath := range []string{
		"/api/v1/masters/permission-groups",
		"/api/v1/owners",
	} {
		request := httptest.NewRequest(http.MethodGet, protectedPath, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		assert.Equal(t, http.StatusUnauthorized, response.Code, protectedPath)
	}
}
