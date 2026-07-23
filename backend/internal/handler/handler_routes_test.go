package handler

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

// TestRegisterRoutes_NoPanic はルート登録時に panic が発生しないことを保証する。
//
// Gin のワイルドカード名衝突（例: :id vs :clinicId）は実行時 panic になるため、
// go build / golangci-lint では検出できない。このテストで CI 上で早期検出する。
//
// 実際に発生した障害: BUG-246 (2026-04-09)
//
//	reservation_line_routes.go の :clinicId/:staffId が既存の :id と衝突して
//	起動時 panic → backend コンテナ unhealthy。
func TestRegisterRoutes_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{
		cfg: &config.Config{
			JWTSecret: "test-secret-for-route-registration",
		},
	}

	assert.NotPanics(t, func() {
		r := gin.New()
		h.RegisterRoutes(context.Background(), r)
	})
}

func TestRegisterRoutes_AuthCookieMutationsRequireRequestedWith(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{
		cfg: &config.Config{
			JWTSecret: "test-secret-for-auth-route-middleware",
		},
	}
	r := gin.New()
	h.RegisterRoutes(context.Background(), r)

	for _, route := range []string{
		"/api/v1/login",
		"/api/v1/logout",
		"/api/v1/auth/refresh",
		"/api/v1/auth/refresh/logout",
	} {
		t.Run(route, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, route, http.NoBody)

			r.ServeHTTP(recorder, request)

			assert.Equal(t, http.StatusForbidden, recorder.Code)
		})
	}
}

func TestRegisterRoutes_RefreshTokenIsRateLimited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	h := &Handler{
		cfg: &config.Config{
			JWTSecret: "test-secret-for-auth-route-rate-limit",
		},
	}
	r := gin.New()
	h.RegisterRoutes(ctx, r)

	for range 30 {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", http.NoBody)
		request.Header.Set("X-Requested-With", "XMLHttpRequest")
		r.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", http.NoBody)
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	r.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
}

func TestRegisterRoutes_CSRFRejectionsDoNotConsumeLoginRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	h := &Handler{
		cfg: &config.Config{
			JWTSecret: "test-secret-for-auth-csrf-rate-limit-order",
		},
	}
	r := gin.New()
	h.RegisterRoutes(ctx, r)

	for range 5 {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/login", http.NoBody)
		r.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusForbidden, recorder.Code)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/login", http.NoBody)
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	r.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestRegisterRoutes_LegacyLogoutRedirectConsumesSingleActionBudget(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	h := &Handler{
		cfg: &config.Config{
			JWTSecret: "test-secret-for-shared-logout-rate-limit",
		},
	}
	r := gin.New()
	h.RegisterRoutes(ctx, r)
	server := httptest.NewServer(r)
	t.Cleanup(server.Close)

	for range 30 {
		request, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/logout", http.NoBody)
		require.NoError(t, err)
		request.Header.Set("X-Requested-With", "XMLHttpRequest")
		response, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		require.NoError(t, response.Body.Close())
		assert.Equal(t, http.StatusOK, response.StatusCode)
	}

	request, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/auth/refresh/logout", http.NoBody)
	require.NoError(t, err)
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, response.Body.Close())
	})
	assert.Equal(t, http.StatusTooManyRequests, response.StatusCode)
}
