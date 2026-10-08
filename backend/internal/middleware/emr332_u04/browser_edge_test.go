// Package emr332_u04 holds EMR-332 U04 (browser/edge defense) memory-only
// verification tests for cases C08 (CSRF custom-header), C11 (transport
// headers), C12 (CORS origin judgment), and C13 (cache prevention).
// Every test exercises the real middleware through httptest only — no DB,
// no network, no persistence. Downstream handler reach is proven with a
// spy counter (precedent: cors_test.go:44-46).
package emr332_u04

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// stagingConfiguredOrigins mirrors backend/wrangler.jsonc vars
// CORS_ALLOWED_ORIGIN so the C12 case binds the shipped allowlist shape
// to the real middleware (value itself is non-secret config).
const stagingConfiguredOrigins = "https://stg.noah-karte.com,https://api.stg.noah-karte.com"

// spyHandler counts downstream invocations so a rejected request is proven
// to have produced zero side-effect-bearing handler calls.
func spyHandler(calls *int) gin.HandlerFunc {
	return func(c *gin.Context) {
		*calls++
		c.String(http.StatusOK, "ok")
	}
}

// TestEMR332C08_CSRFBlocksStateChangingRequestsWithoutHeader verifies the
// real RequireXRequestedWith middleware: non-safe methods without
// X-Requested-With are rejected 403 before the handler runs.
func TestEMR332C08_CSRFBlocksStateChangingRequestsWithoutHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	calls := 0
	router := gin.New()
	router.Use(middleware.RequireXRequestedWith())
	for _, method := range []string{
		http.MethodGet, http.MethodHead, http.MethodOptions,
		http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete,
	} {
		router.Handle(method, "/resource", spyHandler(&calls))
	}

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		t.Run(method+" without X-Requested-With is rejected before handler", func(t *testing.T) {
			calls = 0
			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/resource", http.NoBody)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.Equal(t, 0, calls, "rejected request must not invoke the downstream handler")
		})
	}

	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		t.Run(method+" without X-Requested-With passes through", func(t *testing.T) {
			calls = 0
			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/resource", http.NoBody)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, 1, calls, "safe method must reach the downstream handler")
		})
	}

	t.Run("POST with X-Requested-With passes through", func(t *testing.T) {
		calls = 0
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/resource", http.NoBody)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, calls)
	})
}

// TestEMR332C12_CORSOriginJudgment verifies the real CORS middleware against
// the staging-configured allowlist shape: a disallowed Origin is never granted
// (no ACAO/Vary/Timing-Allow-Origin), a configured Origin receives the echo +
// Vary: Origin, and a preflight never reaches the downstream handler.
func TestEMR332C12_CORSOriginJudgment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	calls := 0
	router := gin.New()
	router.Use(middleware.CORS(stagingConfiguredOrigins))
	router.OPTIONS("/resource", spyHandler(&calls))
	router.GET("/resource", spyHandler(&calls))

	t.Run("disallowed Origin preflight is not granted", func(t *testing.T) {
		calls = 0
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/resource", http.NoBody)
		req.Header.Set("Origin", "https://disallowed.example.com")
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		req.Header.Set("Access-Control-Request-Headers", "x-requested-with")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, 0, calls, "preflight must not reach the downstream handler")
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Timing-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Vary"))
	})

	for _, origin := range []string{"https://stg.noah-karte.com", "https://api.stg.noah-karte.com"} {
		t.Run("configured Origin "+origin+" preflight gets ACAO and Vary", func(t *testing.T) {
			calls = 0
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodOptions, "/resource", http.NoBody)
			req.Header.Set("Origin", origin)
			req.Header.Set("Access-Control-Request-Method", http.MethodPost)
			req.Header.Set("Access-Control-Request-Headers", "content-type,x-requested-with")

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNoContent, w.Code)
			assert.Equal(t, 0, calls, "preflight must not reach the downstream handler")
			assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, origin, w.Header().Get("Timing-Allow-Origin"))
			assert.Equal(t, "Origin", w.Header().Get("Vary"))
			assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
			assert.NotContains(t, strings.ToLower(w.Header().Get("Access-Control-Allow-Headers")), "*")
		})
	}

	t.Run("disallowed Origin on actual request gets no ACAO grant", func(t *testing.T) {
		calls = 0
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/resource", http.NoBody)
		req.Header.Set("Origin", "https://disallowed.example.com")

		router.ServeHTTP(w, req)

		// CORS denial is a browser-enforced read grant: the middleware still
		// answers the request but withholds ACAO/Timing-Allow-Origin/Vary.
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, calls)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Timing-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Vary"))
	})
}

// TestEMR332C11C13_SecurityHeaderChain verifies the real SecurityHeaders
// middleware in production mode: transport/frame/cache directives are emitted
// on normal responses AND still present on a CSRF-rejected (aborted) response,
// matching the global chain order in cmd/api/main.go (SecurityHeaders runs
// before CORS and before any route-level CSRF rejection).
func TestEMR332C11C13_SecurityHeaderChain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	calls := 0
	router := gin.New()
	router.Use(middleware.SecurityHeaders(true))
	router.Use(middleware.CORS(stagingConfiguredOrigins))
	router.Use(middleware.RequireXRequestedWith())
	router.GET("/resource", spyHandler(&calls))
	router.POST("/resource", spyHandler(&calls))

	assertSecurityAndCacheHeaders := func(t *testing.T, h http.Header) {
		assert.Equal(t, "nosniff", h.Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", h.Get("X-Frame-Options"))
		assert.Equal(t, "strict-origin-when-cross-origin", h.Get("Referrer-Policy"))
		assert.Equal(t, "default-src 'none'", h.Get("Content-Security-Policy"))
		assert.Equal(t, "geolocation=(), microphone=(), camera=()", h.Get("Permissions-Policy"))
		assert.Equal(t, "max-age=31536000; includeSubDomains; preload", h.Get("Strict-Transport-Security"))
		assert.Equal(t, "no-store, no-cache, must-revalidate", h.Get("Cache-Control"))
		assert.Equal(t, "no-cache", h.Get("Pragma"))
	}

	t.Run("GET response carries transport and cache-prevention headers", func(t *testing.T) {
		calls = 0
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/resource", http.NoBody)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, calls)
		assertSecurityAndCacheHeaders(t, w.Header())
	})

	t.Run("CSRF-rejected POST still carries security and cache headers", func(t *testing.T) {
		calls = 0
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/resource", http.NoBody)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, 0, calls, "CSRF rejection must precede the handler")
		assertSecurityAndCacheHeaders(t, w.Header())
	})
}
