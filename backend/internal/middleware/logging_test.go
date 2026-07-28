package middleware

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		reqID, _ := c.Get("request_id")
		c.String(http.StatusOK, reqID.(string))
	})

	t.Run("generates request id if not provided", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Body.String())
		assert.Equal(t, w.Body.String(), w.Header().Get("X-Request-ID"))
	})

	t.Run("uses provided valid request id", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("X-Request-ID", "valid-id-123_abc-ABC")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "valid-id-123_abc-ABC", w.Body.String())
	})

	t.Run("regenerates request id if invalid characters present", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("X-Request-ID", "invalid id!")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEqual(t, "invalid id!", w.Body.String())
		assert.NotEmpty(t, w.Body.String())
	})

	t.Run("regenerates request id if too long", func(t *testing.T) {
		tooLongID := strings.Repeat("a", 65)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("X-Request-ID", tooLongID)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEqual(t, tooLongID, w.Body.String())
		assert.NotEmpty(t, w.Body.String())
	})
}

func TestRequestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLoggingMiddleware())

	router.GET("/200", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	router.GET("/400", func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "client error"})
	})
	router.GET("/500", func(c *gin.Context) {
		httpapi.RespondError(c, fmt.Errorf("database error: %w", &pgconn.PgError{
			Code:    "42703",
			Message: "undefined column",
		}))
	})

	t.Run("logs status 200", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/200?query=1", http.NoBody)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("does not log query values", func(t *testing.T) {
		var logBuffer bytes.Buffer
		previousLogger := slog.Default()
		slog.SetDefault(slog.New(slog.NewJSONHandler(&logBuffer, nil)))
		t.Cleanup(func() {
			slog.SetDefault(previousLogger)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/200?access_token=secret-value", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, logBuffer.String(), `"path":"/200"`)
		assert.NotContains(t, logBuffer.String(), "access_token")
		assert.NotContains(t, logBuffer.String(), "secret-value")
	})

	t.Run("logs status 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/400", http.NoBody)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("logs status 500 with error", func(t *testing.T) {
		var logBuffer bytes.Buffer
		previousLogger := slog.Default()
		slog.SetDefault(slog.New(slog.NewJSONHandler(&logBuffer, nil)))
		t.Cleanup(func() {
			slog.SetDefault(previousLogger)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/500", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, logBuffer.String(), `"msg":"server error"`)
		assert.Contains(t, logBuffer.String(), `"status":500`)
		assert.Contains(t, logBuffer.String(), `"error":`)
		assert.Contains(t, logBuffer.String(), "SQLSTATE 42703")
	})
}
