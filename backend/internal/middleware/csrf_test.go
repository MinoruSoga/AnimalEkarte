package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequireXRequestedWith(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequireXRequestedWith())
	router.GET("/get", func(c *gin.Context) {
		c.String(http.StatusOK, "get ok")
	})
	router.POST("/post", func(c *gin.Context) {
		c.String(http.StatusOK, "post ok")
	})

	t.Run("allows idempotent methods (GET) without header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/get", http.NoBody)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "get ok", w.Body.String())
	})

	t.Run("blocks state-changing methods (POST) without header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/post", http.NoBody)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.JSONEq(t, `{"error":"X-Requested-With header required for state-changing requests"}`, w.Body.String())
	})

	t.Run("allows state-changing methods (POST) with header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/post", http.NoBody)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "post ok", w.Body.String())
	})
}
