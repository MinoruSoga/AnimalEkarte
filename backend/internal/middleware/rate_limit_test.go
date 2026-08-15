package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRateLimitStore_Evict(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewRateLimitStore(ctx)

	store.mu.Lock()
	oldKey := rateLimitKey{bucketID: 1, ip: "192.168.1.1"}
	newKey := rateLimitKey{bucketID: 1, ip: "192.168.1.2"}
	store.limiters[oldKey] = &limiterEntry{
		limiter:  rate.NewLimiter(1, 1),
		lastSeen: time.Now().Add(-15 * time.Minute),
	}
	store.limiters[newKey] = &limiterEntry{
		limiter:  rate.NewLimiter(1, 1),
		lastSeen: time.Now(),
	}
	store.mu.Unlock()

	store.evict(10 * time.Minute)

	store.mu.RLock()
	defer store.mu.RUnlock()
	assert.Nil(t, store.limiters[oldKey], "old entry should be evict-ed")
	assert.NotNil(t, store.limiters[newKey], "new entry should remain")
}

func TestRateLimit_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewRateLimitStore(ctx)

	router := gin.New()
	router.Use(RateLimit(store, 1, 1))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	t.Run("allows first request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("X-Forwarded-For", "1.1.1.1")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("blocks second request exceeding burst limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("X-Forwarded-For", "1.1.1.1")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "rate limit exceeded")
	})
}

func TestLiffRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewRateLimitStore(ctx)

	router := gin.New()
	router.Use(LiffRateLimit(store, 60))
	router.GET("/liff", func(c *gin.Context) {
		c.String(http.StatusOK, "liff ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/liff", http.NoBody)
	req.Header.Set("X-Forwarded-For", "2.2.2.2")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "liff ok", w.Body.String())
}

func TestRateLimit_MiddlewareInstancesUseIndependentBuckets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewRateLimitStore(ctx)
	router := gin.New()
	router.GET("/loose", RateLimit(store, 1, 60), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.GET("/strict", RateLimit(store, 1, 10), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for range 11 {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/loose", http.NoBody)
		router.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusNoContent, recorder.Code)
	}

	for range 10 {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/strict", http.NoBody)
		router.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusNoContent, recorder.Code)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/strict", http.NoBody)
	router.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
}
