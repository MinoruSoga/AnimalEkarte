package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noopLoginFailureLimiter struct{}

func (noopLoginFailureLimiter) Admit(string) (int, bool) { return 0, true }
func (noopLoginFailureLimiter) RecordFailure(string)     {}

type denyingLoginFailureLimiter struct{ retryAfter int }

func (l denyingLoginFailureLimiter) Admit(string) (int, bool) { return l.retryAfter, false }
func (denyingLoginFailureLimiter) RecordFailure(string)       {}

type recordingAuthRateLimitStore struct {
	name            string
	policies        []RateLimitPolicy
	failurePolicies []RateLimitPolicy
	calls           *[]string
	failureLimiter  LoginFailureLimiter
}

func (s *recordingAuthRateLimitStore) Middleware(
	policy RateLimitPolicy,
) gin.HandlerFunc {
	s.policies = append(s.policies, policy)
	return func(c *gin.Context) {
		*s.calls = append(*s.calls, "rate:"+s.name)
		c.Next()
	}
}

func (s *recordingAuthRateLimitStore) FailureLimiter(
	policy RateLimitPolicy,
) LoginFailureLimiter {
	s.failurePolicies = append(s.failurePolicies, policy)
	if s.failureLimiter != nil {
		return s.failureLimiter
	}
	return noopLoginFailureLimiter{}
}

type recordingAuthRateLimitFactory struct {
	stores         []*recordingAuthRateLimitStore
	calls          *[]string
	failureLimiter LoginFailureLimiter
}

func (f *recordingAuthRateLimitFactory) New(
	_ context.Context,
) RateLimitStore {
	store := &recordingAuthRateLimitStore{
		name:           string(rune('1' + len(f.stores))),
		calls:          f.calls,
		failureLimiter: f.failureLimiter,
	}
	f.stores = append(f.stores, store)
	return store
}

func appendAuthMiddlewareCall(calls *[]string, name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		*calls = append(*calls, name)
		c.Next()
	}
}

func TestHTTPHandler_RegisterRoutes_HardenedSurface(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var calls []string
	rateLimitFactory := &recordingAuthRateLimitFactory{calls: &calls}
	handler := NewHTTPHandler(HTTPDependencies{}, CookieConfigForProduction(false))
	router := gin.New()
	api := router.Group("/api/v1")

	protected, err := handler.RegisterRoutes(
		context.Background(),
		api,
		RouteMiddlewarePorts{
			CSRF:             appendAuthMiddlewareCall(&calls, "csrf"),
			Authenticate:     appendAuthMiddlewareCall(&calls, "authenticate"),
			RateLimitFactory: rateLimitFactory,
		},
		DefaultAuthRateLimitConfig(),
	)

	require.NoError(t, err)
	require.NotNil(t, protected)
	assert.Len(t, rateLimitFactory.stores, 5)
	assert.Equal(t, []RateLimitPolicy{
		DefaultAuthRateLimitConfig().Login,
	}, rateLimitFactory.stores[0].policies)
	assert.Equal(t, []RateLimitPolicy{
		DefaultAuthRateLimitConfig().LoginPerAccount,
	}, rateLimitFactory.stores[0].failurePolicies)
	assert.Equal(t, []RateLimitPolicy{
		DefaultAuthRateLimitConfig().PasswordReset,
		DefaultAuthRateLimitConfig().PasswordReset,
	}, rateLimitFactory.stores[1].policies)
	assert.Equal(t, []RateLimitPolicy{
		DefaultAuthRateLimitConfig().Refresh,
	}, rateLimitFactory.stores[2].policies)
	assert.Equal(t, []RateLimitPolicy{
		DefaultAuthRateLimitConfig().Logout,
	}, rateLimitFactory.stores[3].policies)
	assert.Equal(t, []RateLimitPolicy{
		DefaultAuthRateLimitConfig().LogoutRedirect,
	}, rateLimitFactory.stores[4].policies)

	routes := make([]string, 0, len(router.Routes()))
	for _, route := range router.Routes() {
		routes = append(routes, route.Method+" "+route.Path)
	}
	for _, expected := range []string{
		"POST /api/v1/login",
		"POST /api/v1/logout",
		"POST /api/v1/auth/refresh/logout",
		"POST /api/v1/auth/refresh",
		"POST /api/v1/auth/forgot-password",
		"POST /api/v1/auth/reset-password",
		"GET /api/v1/me",
		"PUT /api/v1/users/me/password",
	} {
		assert.True(t, slices.Contains(routes, expected), "missing route %s", expected)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/logout", http.NoBody)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, http.StatusTemporaryRedirect, response.Code)
	assert.Equal(t, []string{"csrf", "rate:5"}, calls)

	calls = nil
	request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/forgot-password",
		http.NoBody,
	)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, []string{"rate:2"}, calls, "forgot-password must remain outside CSRF")

	calls = nil
	request = httptest.NewRequest(
		http.MethodPut,
		"/api/v1/users/me/password",
		http.NoBody,
	)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, []string{"authenticate", "csrf"}, calls)
}

func TestHTTPHandler_RegisterRoutes_RejectsIncompleteSecurityPorts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHTTPHandler(HTTPDependencies{}, CookieConfigForProduction(false))
	router := gin.New()

	protected, err := handler.RegisterRoutes(
		context.Background(),
		router.Group("/api/v1"),
		RouteMiddlewarePorts{},
		DefaultAuthRateLimitConfig(),
	)

	assert.Nil(t, protected)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CSRF middleware is required")
}

func TestHTTPHandler_RegisterRoutes_RejectsInvalidRateLimitConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var calls []string
	factory := &recordingAuthRateLimitFactory{calls: &calls}
	handler := NewHTTPHandler(HTTPDependencies{}, CookieConfigForProduction(false))
	router := gin.New()
	config := DefaultAuthRateLimitConfig()
	config.Login.Requests = 0

	protected, err := handler.RegisterRoutes(
		context.Background(),
		router.Group("/api/v1"),
		RouteMiddlewarePorts{
			CSRF:             appendAuthMiddlewareCall(&calls, "csrf"),
			Authenticate:     appendAuthMiddlewareCall(&calls, "authenticate"),
			RateLimitFactory: factory,
		},
		config,
	)

	assert.Nil(t, protected)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid auth rate-limit policy: login")
	assert.Empty(t, factory.stores)
}

func TestRateLimitPolicy_RequestsPerSecond(t *testing.T) {
	policy := DefaultAuthRateLimitConfig().Login
	assert.InDelta(t, 30.0/60.0, policy.RequestsPerSecond(), 0.000001)
	assert.Zero(t, (RateLimitPolicy{}).RequestsPerSecond())
}

func TestHTTPHandler_Login_PerAccountThrottleResponds429(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var calls []string
	factory := &recordingAuthRateLimitFactory{
		calls:          &calls,
		failureLimiter: denyingLoginFailureLimiter{retryAfter: 7},
	}
	handler := NewHTTPHandler(HTTPDependencies{}, CookieConfigForProduction(false))
	router := gin.New()

	_, err := handler.RegisterRoutes(
		context.Background(),
		router.Group("/api/v1"),
		RouteMiddlewarePorts{
			CSRF:             appendAuthMiddlewareCall(&calls, "csrf"),
			Authenticate:     appendAuthMiddlewareCall(&calls, "authenticate"),
			RateLimitFactory: factory,
		},
		DefaultAuthRateLimitConfig(),
	)
	require.NoError(t, err)

	body := strings.NewReader(`{"email":"staff@example.com","password":"password123"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/login", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusTooManyRequests, response.Code)
	assert.Equal(t, "7", response.Header().Get("Retry-After"))
	assert.Contains(t, response.Body.String(), "ログイン試行が集中しています")
}
