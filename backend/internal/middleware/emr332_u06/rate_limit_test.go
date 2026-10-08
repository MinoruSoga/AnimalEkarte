// Package emr332_u06 — EMR-332 C16 evidence tests
// (docs/work/emr-332-security/u06/c16-rate-limit.md).
//
// Scope: deterministic, single-threaded, in-memory verification of
// middleware.RateLimit / FailureRateLimit / FailureLimiter / LiffRateLimit /
// LimitRequestBody. httptest only — no network listener, no DB, no external
// call. Token buckets are exercised through sequential requests; rps is set
// low enough (rate.Every(time.Hour)) that refill cannot fire during the test,
// making every 429 deterministic.
//
// Side-effect accounting: the only "effect" of a rejected request is the
// recorded 429 write on the httptest.ResponseRecorder — downstream handler
// invocation counters prove refusal paths never reach the handler.
package emr332_u06

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/animal-ekarte/backend/internal/middleware"
)

const testClientIP = "203.0.113.5" // TEST-NET-3 (RFC 5737 documentation range)

func newStore(t *testing.T) *middleware.RateLimitStore {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return middleware.NewRateLimitStore(ctx)
}

func newEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Deterministic ClientIP: RemoteAddr only (production pins trusted proxies
	// via TRUSTED_PROXY_CIDR; trusting nothing here keeps the test hermetic).
	_ = r.SetTrustedProxies(nil)
	return r
}

func doRequest(r *gin.Engine, method, path, remoteAddr string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	req.RemoteAddr = remoteAddr + ":54321"
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// ── C16: 429 + Retry-After + downstream handler not reached ──────────────────

func TestC16_RateLimitRejects429WithRetryAfter(t *testing.T) {
	store := newStore(t)
	handlerCalls := 0
	r := newEngine()
	r.GET("/limited", middleware.RateLimit(store, rate.Every(time.Hour), 1), func(c *gin.Context) {
		handlerCalls++
		c.Status(http.StatusOK)
	})

	first := doRequest(r, "GET", "/limited", testClientIP, nil)
	assert.Equal(t, http.StatusOK, first.Code)
	assert.Equal(t, 1, handlerCalls)

	second := doRequest(r, "GET", "/limited", testClientIP, nil)
	assert.Equal(t, http.StatusTooManyRequests, second.Code)
	assert.Equal(t, 1, handlerCalls, "rejected request must not reach the downstream handler")

	retryAfter := second.Header().Get("Retry-After")
	secs, err := strconv.Atoi(retryAfter)
	require.NoError(t, err, "Retry-After must be an integer number of seconds")
	assert.GreaterOrEqual(t, secs, 1)
	assert.Contains(t, second.Body.String(), "rate limit exceeded")
}

// ── C16: bucket granularity — per middleware instance AND per client IP ─────

func TestC16_RateLimitBucketsAreIndependent(t *testing.T) {
	store := newStore(t)
	r := newEngine()
	r.GET("/a", middleware.RateLimit(store, rate.Every(time.Hour), 1), func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/b", middleware.RateLimit(store, rate.Every(time.Hour), 1), func(c *gin.Context) { c.Status(http.StatusOK) })

	assert.Equal(t, http.StatusOK, doRequest(r, "GET", "/a", testClientIP, nil).Code)
	assert.Equal(t, http.StatusTooManyRequests, doRequest(r, "GET", "/a", testClientIP, nil).Code,
		"same bucket+IP must be exhausted")

	// A second middleware instance on the SAME store gets its own bucketID:
	// its budget is untouched by route A's exhaustion.
	assert.Equal(t, http.StatusOK, doRequest(r, "GET", "/b", testClientIP, nil).Code,
		"independent middleware bucket must not share the exhausted budget")

	// A different client IP on route A gets a separate limiter entry.
	assert.Equal(t, http.StatusOK, doRequest(r, "GET", "/a", "203.0.113.99", nil).Code,
		"different client IP must have an independent limiter")
}

// ── C16: FailureRateLimit consumes budget only on >=400 responses ────────────

func TestC16_FailureRateLimitConsumesOnlyFailures(t *testing.T) {
	store := newStore(t)
	handlerCalls := 0
	status := http.StatusOK
	r := newEngine()
	r.POST("/failgate", middleware.FailureRateLimit(store, rate.Every(time.Hour), 2), func(c *gin.Context) {
		handlerCalls++
		c.Status(status)
	})

	// 3 successes consume nothing (budget 2 stays full).
	for i := 0; i < 3; i++ {
		status = http.StatusOK
		assert.Equal(t, http.StatusOK, doRequest(r, "POST", "/failgate", testClientIP, nil).Code)
	}
	assert.Equal(t, 3, handlerCalls)

	// Two failures drain the 2-token budget.
	status = http.StatusInternalServerError
	assert.Equal(t, http.StatusInternalServerError, doRequest(r, "POST", "/failgate", testClientIP, nil).Code)
	assert.Equal(t, http.StatusInternalServerError, doRequest(r, "POST", "/failgate", testClientIP, nil).Code)

	// Budget exhausted: the next request is gated before the handler, even
	// though this one would have succeeded.
	status = http.StatusOK
	rejected := doRequest(r, "POST", "/failgate", testClientIP, nil)
	assert.Equal(t, http.StatusTooManyRequests, rejected.Code)
	assert.Equal(t, 5, handlerCalls, "gated request must not reach the handler")
	secs, err := strconv.Atoi(rejected.Header().Get("Retry-After"))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, secs, 1)
}

// ── C16: FailureLimiter key space (e.g. per-account email) ──────────────────

func TestC16_FailureLimiterAdmitAndRecordFailure(t *testing.T) {
	store := newStore(t)
	limiter := store.NewFailureLimiter(rate.Every(time.Hour), 1)

	secs, ok := limiter.Admit("SECCHK20261009ACCT01")
	assert.True(t, ok)
	assert.Equal(t, 0, secs)

	limiter.RecordFailure("SECCHK20261009ACCT01")

	secs, ok = limiter.Admit("SECCHK20261009ACCT01")
	assert.False(t, ok, "exhausted key must be refused")
	assert.GreaterOrEqual(t, secs, 1)

	// Independent key space: another account key is unaffected.
	_, ok = limiter.Admit("SECCHK20261009ACCT02")
	assert.True(t, ok)

	// Admit is non-consuming: re-admitting the fresh key still passes.
	_, ok = limiter.Admit("SECCHK20261009ACCT02")
	assert.True(t, ok, "Admit must not consume failure budget")
}

// ── C16: LiffRateLimit converts requests/minute to rps and delegates ─────────

func TestC16_LiffRateLimitPerMinuteBurst(t *testing.T) {
	store := newStore(t)
	handlerCalls := 0
	r := newEngine()
	// 1 request/minute => burst 1, refill far beyond the test window.
	r.GET("/liff", middleware.LiffRateLimit(store, 1), func(c *gin.Context) {
		handlerCalls++
		c.Status(http.StatusOK)
	})

	assert.Equal(t, http.StatusOK, doRequest(r, "GET", "/liff", testClientIP, nil).Code)
	second := doRequest(r, "GET", "/liff", testClientIP, nil)
	assert.Equal(t, http.StatusTooManyRequests, second.Code)
	assert.Equal(t, 1, handlerCalls)
	assert.NotEmpty(t, second.Header().Get("Retry-After"))
}

// ── C16: request-size limit (413 on declared length; cap on streaming read) ──

func TestC16_LimitRequestBodyRejectsOversizedDeclaredLength(t *testing.T) {
	handlerCalls := 0
	r := newEngine()
	r.POST("/small", middleware.LimitRequestBody(64), func(c *gin.Context) {
		handlerCalls++
		c.Status(http.StatusOK)
	})

	body := strings.Repeat("x", 100)
	req := httptest.NewRequest(http.MethodPost, "/small", strings.NewReader(body))
	req.RemoteAddr = testClientIP + ":54321"
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.Equal(t, 0, handlerCalls, "oversized declared length must abort before the handler")
	assert.NotEmpty(t, rec.Body.String())
}

func TestC16_LimitRequestBodyCapsStreamingRead(t *testing.T) {
	var readErr error
	var readN int
	handlerCalls := 0
	r := newEngine()
	r.POST("/stream", middleware.LimitRequestBody(64), func(c *gin.Context) {
		handlerCalls++
		data, err := io.ReadAll(c.Request.Body)
		readN = len(data)
		readErr = err
		c.Status(http.StatusOK)
	})

	// Unknown declared length (chunked semantics): the byte ceiling is enforced
	// by MaxBytesReader when the handler actually reads, not by Content-Length.
	req := httptest.NewRequest(http.MethodPost, "/stream", strings.NewReader(strings.Repeat("y", 100)))
	req.RemoteAddr = testClientIP + ":54321"
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = -1
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, 1, handlerCalls)
	require.Error(t, readErr, "MaxBytesReader must surface an error past the cap")
	assert.LessOrEqual(t, readN, 64)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestC16_LimitRequestBodyScopeIsWriteMethods(t *testing.T) {
	var readN int
	r := newEngine()
	// GET is outside the bound-method set; the body reaches the handler unwrapped.
	r.GET("/read", middleware.LimitRequestBody(64), func(c *gin.Context) {
		data, _ := io.ReadAll(c.Request.Body)
		readN = len(data)
		c.Status(http.StatusOK)
	})
	doRequest(r, "GET", "/read", testClientIP, strings.NewReader(strings.Repeat("z", 100)))
	assert.Equal(t, 100, readN, "GET bodies are documented as unbounded by LimitRequestBody")
}
