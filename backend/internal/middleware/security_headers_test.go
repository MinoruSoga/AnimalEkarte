package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		isProduction   bool
		wantHSTSHeader string
	}{
		{
			name:           "production adds HSTS preload",
			isProduction:   true,
			wantHSTSHeader: "max-age=31536000; includeSubDomains; preload",
		},
		{
			name:           "non-production omits HSTS",
			isProduction:   false,
			wantHSTSHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(SecurityHeaders(tt.isProduction))
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			router.ServeHTTP(w, req)

			assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
			assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
			assert.Equal(t, "default-src 'none'", w.Header().Get("Content-Security-Policy"))
			assert.Equal(t, tt.wantHSTSHeader, w.Header().Get("Strict-Transport-Security"))
		})
	}
}
