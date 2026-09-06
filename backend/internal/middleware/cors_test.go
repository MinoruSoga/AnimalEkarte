package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORS_AccountingPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		allowedOrigin string
		origin        string
		wantOrigin    string
	}{
		{
			name:       "default origin permits idempotent accounting requests",
			origin:     "http://localhost:3000",
			wantOrigin: "http://localhost:3000",
		},
		{
			name:          "configured origin permits idempotent accounting requests",
			allowedOrigin: "https://example.com, https://app.example.com",
			origin:        "https://app.example.com",
			wantOrigin:    "https://app.example.com",
		},
		{
			name:          "unconfigured origin remains denied",
			allowedOrigin: "https://app.example.com",
			origin:        "https://unauthorized.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CORS(tt.allowedOrigin))
			router.Use(func(_ *gin.Context) {
				t.Error("preflight must not reach the downstream handler")
			})
			req := httptest.NewRequest(http.MethodOptions, "/api/v1/accountings/complete", http.NoBody)
			req.Header.Set("Origin", tt.origin)
			req.Header.Set("Access-Control-Request-Method", http.MethodPost)
			req.Header.Set("Access-Control-Request-Headers", "content-type,x-requested-with,x-clinic-id,idempotency-key")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNoContent, w.Code)
			assert.Equal(t, tt.wantOrigin, w.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
			assert.Contains(t, strings.Split(w.Header().Get("Access-Control-Allow-Methods"), ", "), http.MethodPost)
			allowedHeaders := strings.Split(strings.ToLower(w.Header().Get("Access-Control-Allow-Headers")), ", ")
			for _, header := range strings.Split(req.Header.Get("Access-Control-Request-Headers"), ",") {
				assert.Contains(t, allowedHeaders, header)
			}
			assert.NotContains(t, allowedHeaders, "*")
		})
	}
}

func TestCORS_PreflightDoesNotReflectArbitraryHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS("https://app.example.com"))
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/accountings/complete", http.NoBody)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "idempotency-key,x-unapproved-header")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	allowedHeaders := strings.Split(strings.ToLower(w.Header().Get("Access-Control-Allow-Headers")), ", ")
	assert.NotContains(t, allowedHeaders, "x-unapproved-header")
	assert.NotContains(t, allowedHeaders, "*")
}

func TestCORS_DefaultOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS(""))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	t.Run("allows default origin http://localhost:3000", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Origin", "http://localhost:3000")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Timing-Allow-Origin"))
		assert.Equal(t, "Origin", w.Header().Get("Vary"))
	})

	t.Run("denies unauthorized origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Origin", "http://unauthorized.com")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Timing-Allow-Origin"))
	})

	t.Run("handles OPTIONS request with 204", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodOptions, "/test", http.NoBody)
		req.Header.Set("Origin", "http://localhost:3000")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestCORS_CustomOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS("https://example.com,https://app.example.com"))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	t.Run("allows custom origin https://example.com", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Origin", "https://example.com")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("denies default origin when overridden", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Origin", "http://localhost:3000")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}
