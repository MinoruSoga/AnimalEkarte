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

func TestRateLimit_RejectedRequestSetsRetryAfter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewRateLimitStore(ctx)
	router := gin.New()
	router.Use(RateLimit(store, rate.Limit(1.0/60.0), 1))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	first := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("X-Forwarded-For", "3.3.3.3")
	router.ServeHTTP(first, req)
	assert.Equal(t, http.StatusOK, first.Code)

	second := httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("X-Forwarded-For", "3.3.3.3")
	router.ServeHTTP(second, req)

	assert.Equal(t, http.StatusTooManyRequests, second.Code)
	assert.NotEmpty(t, second.Header().Get("Retry-After"))
}

func TestFailureRateLimit_SuccessfulRequestsAreFree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewRateLimitStore(ctx)
	router := gin.New()
	router.Use(FailureRateLimit(store, rate.Limit(1.0/60.0), 2))
	router.POST("/login", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// burst=2 を大きく超える成功 request がすべて通ることを確認する
	for range 5 {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/login", http.NoBody)
		req.RemoteAddr = "4.4.4.4:12345"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestFailureRateLimit_CountsFailuresOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewRateLimitStore(ctx)
	router := gin.New()
	router.Use(FailureRateLimit(store, rate.Limit(1.0/60.0), 2))
	router.POST("/login", func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid"})
	})

	newRequest := func(remoteAddr string) (*httptest.ResponseRecorder, *http.Request) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/login", http.NoBody)
		req.RemoteAddr = remoteAddr
		return w, req
	}

	// burst=2 の失敗でバケットを使い切る
	for range 2 {
		w, req := newRequest("5.5.5.5:12345")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}

	w, req := newRequest("5.5.5.5:12345")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.NotEmpty(t, w.Header().Get("Retry-After"))

	// 別 IP は独立バケットのため影響を受けない
	other, otherReq := newRequest("6.6.6.6:12345")
	router.ServeHTTP(other, otherReq)
	assert.Equal(t, http.StatusUnauthorized, other.Code)
}

func TestRateLimitStore_FailureLimiter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewRateLimitStore(ctx)
	limiter := store.NewFailureLimiter(rate.Limit(1.0/60.0), 2)

	retryAfter, ok := limiter.Admit("a@example.com")
	assert.True(t, ok)
	assert.Zero(t, retryAfter)

	limiter.RecordFailure("a@example.com")
	limiter.RecordFailure("a@example.com")

	retryAfter, ok = limiter.Admit("a@example.com")
	assert.False(t, ok)
	assert.Positive(t, retryAfter)

	// 別 key は独立したバケットを持つ
	_, ok = limiter.Admit("b@example.com")
	assert.True(t, ok)
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
