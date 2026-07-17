package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

const testSecret = "test-secret-key"

func testTokenSvc() service.TokenService {
	return service.NewTokenService(testSecret, nil)
}

func makeToken(t *testing.T, method jwt.SigningMethod, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	s, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func validClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"user_id":         "user-uuid-123",
		"clinic_id":       "clinic-uuid-456",
		"is_system_admin": false,
		"exp":             time.Now().Add(time.Hour).Unix(),
	}
}

// runAuthMiddleware sets up a test router with the Auth middleware and a
// downstream handler that records the context values, then fires the request.
func runAuthMiddleware(t *testing.T, authHeader string) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	return runAuthMiddlewareWithAudit(t, authHeader, nil, nil)
}

// runAuthMiddlewareWithAudit は auditSvc と追加リクエスト設定を受け取る拡張ヘルパー。
func runAuthMiddlewareWithAudit(t *testing.T, authHeader string, auditSvc service.AuditService, setupReq func(*http.Request)) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var captured *gin.Context
	w := httptest.NewRecorder()
	router := gin.New()
	router.Use(Auth(testTokenSvc(), false, auditSvc, nil))
	router.GET("/test", func(c *gin.Context) {
		captured = c
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	if setupReq != nil {
		setupReq(req)
	}
	router.ServeHTTP(w, req)

	return w, captured
}

// ---- mock AuditService for middleware tests ----

type mockMiddlewareAuditService struct {
	logClinicSwitchFn   func(ctx context.Context, actorID *uint64, from, to uint64, ip, ua string) error
	logClinicSwitchCall *struct {
		actorID      *uint64
		fromClinicID uint64
		toClinicID   uint64
	}
}

func (m *mockMiddlewareAuditService) Log(_ context.Context, _ *model.AuditLog) error { return nil }
func (m *mockMiddlewareAuditService) LogEntry(_ context.Context, _ *service.AuditLogInput) error {
	return nil
}
func (m *mockMiddlewareAuditService) LogAuthLogin(_ context.Context, _, _ *uint64, _, _, _ string) error {
	return nil
}
func (m *mockMiddlewareAuditService) LogLstepOperation(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64) error {
	return nil
}
func (m *mockMiddlewareAuditService) LogLstepOperationWithMetadata(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64, _ any) error {
	return nil
}
func (m *mockMiddlewareAuditService) LogMedicalRecordChange(_ context.Context, _ uint64, _ *uint64, _ string, _ uint64, _, _ map[string]any) error {
	return nil
}
func (m *mockMiddlewareAuditService) LogVitalChange(_ context.Context, _ uint64, _ *uint64, _ string, _, _ uint64, _, _ map[string]any) error {
	return nil
}
func (m *mockMiddlewareAuditService) LogAddendumCreate(_ context.Context, _ uint64, _ *uint64, _, _ uint64, _ *model.MedicalRecordAddendum) error {
	return nil
}
func (m *mockMiddlewareAuditService) LogClinicSwitch(ctx context.Context, actorID *uint64, from, to uint64, ip, ua string) error {
	m.logClinicSwitchCall = &struct {
		actorID      *uint64
		fromClinicID uint64
		toClinicID   uint64
	}{actorID: actorID, fromClinicID: from, toClinicID: to}
	if m.logClinicSwitchFn != nil {
		return m.logClinicSwitchFn(ctx, actorID, from, to, ip, ua)
	}
	return nil
}

func TestAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing Authorization header returns 401", func(t *testing.T) {
		w, _ := runAuthMiddleware(t, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "authorization required")
	})

	t.Run("malformed header without Bearer scheme returns 401", func(t *testing.T) {
		w, _ := runAuthMiddleware(t, "Token some-token")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token string returns 401", func(t *testing.T) {
		w, _ := runAuthMiddleware(t, "Bearer not-a-valid-token")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid or expired token")
	})

	t.Run("expired token returns 401", func(t *testing.T) {
		expiredClaims := jwt.MapClaims{
			"user_id":         "user-uuid-123",
			"clinic_id":       "clinic-uuid-456",
			"is_system_admin": false,
			"exp":             time.Now().Add(-time.Hour).Unix(),
		}
		token := makeToken(t, jwt.SigningMethodHS256, expiredClaims)
		w, _ := runAuthMiddleware(t, "Bearer "+token)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("token signed with HS384 is rejected", func(t *testing.T) {
		// TokenService.VerifyAccessToken は HS256 のみ許可する（HMAC 他 alg は拒否）。
		token := makeToken(t, jwt.SigningMethodHS384, validClaims())
		w, _ := runAuthMiddleware(t, "Bearer "+token)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("token signed with wrong secret returns 401", func(t *testing.T) {
		// Manually sign with a different key
		claims := validClaims()
		rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := rawToken.SignedString([]byte("wrong-secret"))
		assert.NoError(t, err)

		w, _ := runAuthMiddleware(t, "Bearer "+tokenStr)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid HS256 token passes through with correct context values", func(t *testing.T) {
		token := makeToken(t, jwt.SigningMethodHS256, validClaims())
		w, captured := runAuthMiddleware(t, "Bearer "+token)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotNil(t, captured)
		assert.Equal(t, "user-uuid-123", captured.GetString("user_id"))
		assert.Equal(t, "clinic-uuid-456", captured.GetString("clinic_id"))
		assert.Equal(t, false, captured.GetBool("is_system_admin"))
	})

	t.Run("Bearer scheme is case-insensitive", func(t *testing.T) {
		token := makeToken(t, jwt.SigningMethodHS256, validClaims())
		w, _ := runAuthMiddleware(t, "bearer "+token)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// clinicSwitchClaims は FEAT-374 Phase 2 audit テスト用の JWT claim を生成する。
// 全 caller で user_id="1" / main clinic_id="1" 固定のため、両方を引数から除外している
// (unparam 解消)。可変なのは clinicIDs (caller の所属クリニック範囲) のみ。
func clinicSwitchClaims(clinicIDs []uint64) jwt.MapClaims {
	return jwt.MapClaims{
		"user_id":         "1",
		"clinic_id":       "1",
		"is_system_admin": false,
		"clinic_ids":      clinicIDs,
		"exp":             time.Now().Add(time.Hour).Unix(),
	}
}

// TestAuth_ClinicSwitch_AuditLog は FEAT-374 Phase 2 のクリニック切替 audit log 動作を検証する。
func TestAuth_ClinicSwitch_AuditLog(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("logs clinic switch when prev cookie differs from header", func(t *testing.T) {
		spy := &mockMiddlewareAuditService{}
		token := makeToken(t, jwt.SigningMethodHS256, clinicSwitchClaims([]uint64{1, 2}))

		w, _ := runAuthMiddlewareWithAudit(t, "Bearer "+token, spy, func(req *http.Request) {
			req.Header.Set("X-Clinic-ID", "2")
			req.AddCookie(&http.Cookie{Name: "prev_clinic_id", Value: "1"})
		})

		assert.Equal(t, http.StatusOK, w.Code)
		if assert.NotNil(t, spy.logClinicSwitchCall, "LogClinicSwitch must be called once") {
			assert.Equal(t, uint64(1), spy.logClinicSwitchCall.fromClinicID)
			assert.Equal(t, uint64(2), spy.logClinicSwitchCall.toClinicID)
		}
		// cookie が新しい clinic_id に更新される
		found := false
		for _, c := range w.Result().Cookies() {
			if c.Name == "prev_clinic_id" && c.Value == "2" {
				found = true
			}
		}
		assert.True(t, found, "prev_clinic_id cookie must be updated to 2")
	})

	t.Run("skips audit when prev cookie matches header", func(t *testing.T) {
		spy := &mockMiddlewareAuditService{}
		token := makeToken(t, jwt.SigningMethodHS256, clinicSwitchClaims([]uint64{1, 2}))

		w, _ := runAuthMiddlewareWithAudit(t, "Bearer "+token, spy, func(req *http.Request) {
			req.Header.Set("X-Clinic-ID", "2")
			req.AddCookie(&http.Cookie{Name: "prev_clinic_id", Value: "2"})
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Nil(t, spy.logClinicSwitchCall, "LogClinicSwitch must NOT be called when cookie matches")
	})

	t.Run("skips audit on first access (no prev cookie), sets cookie", func(t *testing.T) {
		spy := &mockMiddlewareAuditService{}
		token := makeToken(t, jwt.SigningMethodHS256, clinicSwitchClaims([]uint64{1, 2}))

		w, _ := runAuthMiddlewareWithAudit(t, "Bearer "+token, spy, func(req *http.Request) {
			req.Header.Set("X-Clinic-ID", "2")
			// no prev_clinic_id cookie
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Nil(t, spy.logClinicSwitchCall, "LogClinicSwitch must NOT be called on first access")
		found := false
		for _, c := range w.Result().Cookies() {
			if c.Name == "prev_clinic_id" && c.Value == "2" {
				found = true
			}
		}
		assert.True(t, found, "prev_clinic_id cookie must be set on first access")
	})

	t.Run("skips audit when no X-Clinic-ID header", func(t *testing.T) {
		spy := &mockMiddlewareAuditService{}
		token := makeToken(t, jwt.SigningMethodHS256, clinicSwitchClaims([]uint64{1}))

		w, _ := runAuthMiddlewareWithAudit(t, "Bearer "+token, spy, func(req *http.Request) {
			req.AddCookie(&http.Cookie{Name: "prev_clinic_id", Value: "1"})
			// no X-Clinic-ID header
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Nil(t, spy.logClinicSwitchCall, "LogClinicSwitch must NOT be called when no X-Clinic-ID")
		for _, c := range w.Result().Cookies() {
			assert.NotEqual(t, "prev_clinic_id", c.Name, "prev_clinic_id cookie must NOT be set")
		}
	})

	t.Run("continues request even when audit fails (best-effort)", func(t *testing.T) {
		spy := &mockMiddlewareAuditService{
			logClinicSwitchFn: func(_ context.Context, _ *uint64, _, _ uint64, _, _ string) error {
				return errors.New("db down")
			},
		}
		token := makeToken(t, jwt.SigningMethodHS256, clinicSwitchClaims([]uint64{1, 2}))

		w, _ := runAuthMiddlewareWithAudit(t, "Bearer "+token, spy, func(req *http.Request) {
			req.Header.Set("X-Clinic-ID", "2")
			req.AddCookie(&http.Cookie{Name: "prev_clinic_id", Value: "1"})
		})

		assert.Equal(t, http.StatusOK, w.Code, "request must proceed even if audit fails")
		assert.NotNil(t, spy.logClinicSwitchCall, "LogClinicSwitch was called (but returned error)")
	})
}

func TestAuth_AccessTokenCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	router := gin.New()
	router.Use(Auth(testTokenSvc(), false, nil, nil))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	token := makeToken(t, jwt.SigningMethodHS256, validClaims())
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuth_AuthTokenCookieFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	router := gin.New()
	router.Use(Auth(testTokenSvc(), false, nil, nil))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	token := makeToken(t, jwt.SigningMethodHS256, validClaims())
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuth_SigningMethodMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	router := gin.New()
	router.Use(Auth(testTokenSvc(), false, nil, nil))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	token := jwt.NewWithClaims(jwt.SigningMethodNone, validClaims())
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidClinicIDHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	router := gin.New()
	router.Use(Auth(testTokenSvc(), false, nil, nil))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	token := makeToken(t, jwt.SigningMethodHS256, validClaims())
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Clinic-ID", "invalid-int")

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuth_NotAssignedClinicIDHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	router := gin.New()
	claims := clinicSwitchClaims([]uint64{1, 2})
	token := makeToken(t, jwt.SigningMethodHS256, claims)

	router.Use(Auth(testTokenSvc(), false, nil, nil))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Clinic-ID", "3")

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

type mockStaffService struct {
	service.StaffService
	getByIDFn func(ctx context.Context, id uint64) (*model.Staff, error)
}

func (m *mockStaffService) GetByID(ctx context.Context, id uint64) (*model.Staff, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &model.Staff{ID: id, IsActive: true}, nil
}

func TestAuth_StaffValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	claims := jwt.MapClaims{
		"user_id":   "123",
		"clinic_id": "1",
		"exp":       time.Now().Add(time.Hour).Unix(),
	}
	token := makeToken(t, jwt.SigningMethodHS256, claims)

	t.Run("allows active staff", func(t *testing.T) {
		w := httptest.NewRecorder()
		router := gin.New()
		staffSvc := &mockStaffService{}
		router.Use(Auth(testTokenSvc(), false, nil, staffSvc))
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("blocks inactive staff", func(t *testing.T) {
		w := httptest.NewRecorder()
		router := gin.New()
		staffSvc := &mockStaffService{
			getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
				return &model.Staff{IsActive: false}, nil
			},
		}
		router.Use(Auth(testTokenSvc(), false, nil, staffSvc))
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("allows staff check to proceed even if DB query fails", func(t *testing.T) {
		w := httptest.NewRecorder()
		router := gin.New()
		staffSvc := &mockStaffService{
			getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
				return nil, errors.New("db error")
			},
		}
		router.Use(Auth(testTokenSvc(), false, nil, staffSvc))
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAuth_RejectsRefreshTokenSubject(t *testing.T) {
	claims := validClaims()
	claims["sub"] = "refresh"
	token := makeToken(t, jwt.SigningMethodHS256, claims)

	w, _ := runAuthMiddleware(t, "Bearer "+token)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}
