package middleware

import (
	"context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	authdomain "github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/model"
)

const testSecret = "test-secret-key"
const testCurrentAccountEpoch int64 = 1_721_000_000_000_000_000

func testTokenSvc() authdomain.TokenService {
	return authdomain.NewTokenService(testSecret, nil)
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
	now := time.Now()
	return jwt.MapClaims{
		"user_id":         "123",
		"clinic_id":       "456",
		"clinic_ids":      []uint64{456},
		"is_system_admin": false,
		"account_epoch":   testCurrentAccountEpoch,
		"jti":             "middleware-valid-access-token",
		"iat":             now.Unix(),
		"exp":             now.Add(15 * time.Minute).Unix(),
	}
}

// runAuthMiddleware sets up a test router with the Auth middleware and a
// downstream handler that records the context values, then fires the request.
func runAuthMiddleware(t *testing.T, authHeader string) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	return runAuthMiddlewareWithAudit(t, authHeader, nil, nil)
}

// runAuthMiddlewareWithAudit は auditSvc と追加リクエスト設定を受け取る拡張ヘルパー。
func runAuthMiddlewareWithAudit(t *testing.T, authHeader string, auditSvc audit.Service, setupReq func(*http.Request)) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var captured *gin.Context
	w := httptest.NewRecorder()
	router := gin.New()
	router.Use(Auth(
		testTokenSvc(),
		false,
		auditSvc,
		activeMiddlewareCurrentAccessResolver(),
	))
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
func (m *mockMiddlewareAuditService) LogEntry(_ context.Context, _ *audit.Entry) error {
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
			"account_epoch":   testCurrentAccountEpoch,
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
		assert.Equal(t, "123", captured.GetString("user_id"))
		assert.Equal(t, "456", captured.GetString("clinic_id"))
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
	now := time.Now()
	return jwt.MapClaims{
		"user_id":         "1",
		"clinic_id":       "1",
		"is_system_admin": false,
		"clinic_ids":      clinicIDs,
		"account_epoch":   testCurrentAccountEpoch,
		"jti":             "middleware-clinic-switch-access-token",
		"iat":             now.Unix(),
		"exp":             now.Add(15 * time.Minute).Unix(),
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
	router.Use(Auth(
		testTokenSvc(),
		false,
		nil,
		activeMiddlewareCurrentAccessResolver(),
	))
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
	router.Use(Auth(
		testTokenSvc(),
		false,
		nil,
		activeMiddlewareCurrentAccessResolver(),
	))
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
	router.Use(Auth(
		testTokenSvc(),
		false,
		nil,
		activeMiddlewareCurrentAccessResolver(),
	))
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
	router.Use(Auth(
		testTokenSvc(),
		false,
		nil,
		activeMiddlewareCurrentAccessResolver(),
	))
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

	router.Use(Auth(
		testTokenSvc(),
		false,
		nil,
		activeMiddlewareCurrentAccessResolver(),
	))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Clinic-ID", "3")

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

type mockCurrentAccessStaffReader struct {
	findFn func(
		ctx context.Context,
		id uint64,
	) (*authdomain.CurrentAccessStaffIdentity, error)
}

func activeMiddlewareStaffIdentity(
	staffID uint64,
) *authdomain.CurrentAccessStaffIdentity {
	accountID := uint64(41)
	return &authdomain.CurrentAccessStaffIdentity{
		ID:        staffID,
		AccountID: &accountID,
		IsActive:  true,
	}
}

func (m *mockCurrentAccessStaffReader) FindCurrentAccessStaff(
	ctx context.Context,
	id uint64,
) (*authdomain.CurrentAccessStaffIdentity, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return activeMiddlewareStaffIdentity(id), nil
}

type middlewareCurrentAccessAccountReader struct{}

func (middlewareCurrentAccessAccountReader) GetByID(
	context.Context,
	uint64,
) (*model.Account, error) {
	return &model.Account{
		ID:        41,
		IsActive:  true,
		UpdatedAt: time.Unix(0, testCurrentAccountEpoch),
	}, nil
}

type middlewareCurrentAccessAssignmentReader struct{}

func (middlewareCurrentAccessAssignmentReader) FindAllByStaffID(
	_ context.Context,
	staffID uint64,
) ([]model.StaffClinicAssignment, error) {
	return []model.StaffClinicAssignment{
		{StaffID: staffID, ClinicID: 1, IsMain: true},
		{StaffID: staffID, ClinicID: 2},
		{StaffID: staffID, ClinicID: 456},
	}, nil
}

type middlewareCurrentAccessClinicReader struct{}

func (middlewareCurrentAccessClinicReader) ListClinics(
	context.Context,
) ([]model.Clinic, error) {
	return []model.Clinic{
		{ID: 1, IsActive: true},
		{ID: 2, IsActive: true},
		{ID: 456, IsActive: true},
	}, nil
}

func middlewareCurrentAccessResolverWithStaff(
	staff authdomain.CurrentAccessStaffReader,
) authdomain.CurrentAccessResolver {
	return authdomain.NewCurrentAccessResolverWithClinics(
		staff,
		middlewareCurrentAccessAccountReader{},
		middlewareCurrentAccessAssignmentReader{},
		middlewareCurrentAccessClinicReader{},
	)
}

func activeMiddlewareCurrentAccessResolver() authdomain.CurrentAccessResolver {
	return middlewareCurrentAccessResolverWithStaff(
		&mockCurrentAccessStaffReader{},
	)
}

func TestAuth_StaffValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	claims := validClaims()
	claims["clinic_id"] = "1"
	claims["clinic_ids"] = []uint64{1}
	token := makeToken(t, jwt.SigningMethodHS256, claims)

	temporaryStorageErr := driver.ErrBadConn
	notFoundErr := apperrors.WrapNotFound("staff", "123")
	genericErr := errors.New("unexpected staff repository error")
	programmingErr := &pgconn.PgError{Code: "42601", Message: "syntax error"}
	temporaryPostgresErr := &pgconn.PgError{Code: "57P03", Message: "cannot connect now"}
	notifierErr := errors.New("notifier unavailable")
	tests := []struct {
		name              string
		staff             *authdomain.CurrentAccessStaffIdentity
		staffErr          error
		notifierErr       error
		injectNotifier    bool
		wantStatus        int
		wantDownstream    bool
		wantNotifierCalls int
	}{
		{
			name:              "allows active staff without notification",
			staff:             activeMiddlewareStaffIdentity(123),
			injectNotifier:    true,
			wantStatus:        http.StatusOK,
			wantDownstream:    true,
			wantNotifierCalls: 0,
		},
		{
			name: "blocks inactive staff without notification",
			staff: &authdomain.CurrentAccessStaffIdentity{
				ID:       123,
				IsActive: false,
			},
			injectNotifier:    true,
			wantStatus:        http.StatusForbidden,
			wantDownstream:    false,
			wantNotifierCalls: 0,
		},
		{
			name:              "blocks missing staff without notification",
			injectNotifier:    true,
			wantStatus:        http.StatusForbidden,
			wantDownstream:    false,
			wantNotifierCalls: 0,
		},
		{
			name:              "notifies exactly once and denies for bad database connection",
			staffErr:          temporaryStorageErr,
			injectNotifier:    true,
			wantStatus:        http.StatusServiceUnavailable,
			wantDownstream:    false,
			wantNotifierCalls: 1,
		},
		{
			name:              "denies when temporary failure notification fails",
			staffErr:          temporaryStorageErr,
			notifierErr:       notifierErr,
			injectNotifier:    true,
			wantStatus:        http.StatusServiceUnavailable,
			wantDownstream:    false,
			wantNotifierCalls: 1,
		},
		{
			name:              "notifies and denies for PostgreSQL cannot-connect-now",
			staffErr:          temporaryPostgresErr,
			injectNotifier:    true,
			wantStatus:        http.StatusServiceUnavailable,
			wantDownstream:    false,
			wantNotifierCalls: 1,
		},
		{
			name:              "not found fails closed without notification",
			staffErr:          notFoundErr,
			injectNotifier:    true,
			wantStatus:        http.StatusForbidden,
			wantDownstream:    false,
			wantNotifierCalls: 0,
		},
		{
			name:              "generic repository error fails closed without notification",
			staffErr:          genericErr,
			injectNotifier:    true,
			wantStatus:        http.StatusServiceUnavailable,
			wantDownstream:    false,
			wantNotifierCalls: 0,
		},
		{
			name:              "PostgreSQL programming error fails closed without notification",
			staffErr:          programmingErr,
			injectNotifier:    true,
			wantStatus:        http.StatusServiceUnavailable,
			wantDownstream:    false,
			wantNotifierCalls: 0,
		},
		{
			name:              "legacy Auth fails closed without a notifier for temporary storage error",
			staffErr:          temporaryStorageErr,
			wantStatus:        http.StatusServiceUnavailable,
			wantDownstream:    false,
			wantNotifierCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lookupStaffID uint64
			staffSvc := &mockCurrentAccessStaffReader{
				findFn: func(
					_ context.Context,
					staffID uint64,
				) (*authdomain.CurrentAccessStaffIdentity, error) {
					lookupStaffID = staffID
					return tt.staff, tt.staffErr
				},
			}

			var notifiedStaffID uint64
			var notifiedErr error
			notifierCalls := 0
			notifier := StaffValidationFailureNotifier(func(_ context.Context, staffID uint64, cause error) error {
				notifierCalls++
				notifiedStaffID = staffID
				notifiedErr = cause
				return tt.notifierErr
			})

			downstreamCalled := false
			w := httptest.NewRecorder()
			router := gin.New()
			resolver := middlewareCurrentAccessResolverWithStaff(staffSvc)
			authMiddleware := Auth(testTokenSvc(), false, nil, resolver)
			if tt.injectNotifier {
				authMiddleware = AuthWithStaffValidationFailureNotifier(
					testTokenSvc(),
					false,
					nil,
					resolver,
					notifier,
				)
			}
			router.Use(authMiddleware)
			router.GET("/test", func(c *gin.Context) {
				downstreamCalled = true
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			req.Header.Set("Authorization", "Bearer "+token)
			router.ServeHTTP(w, req)

			assert.Equal(t, uint64(123), lookupStaffID)
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantDownstream, downstreamCalled)
			assert.Equal(t, tt.wantNotifierCalls, notifierCalls)
			if tt.wantNotifierCalls > 0 {
				assert.Equal(t, uint64(123), notifiedStaffID)
				assert.Same(t, tt.staffErr, notifiedErr)
			}
		})
	}
}

func TestTemporaryStaffValidationErrorClassification(t *testing.T) {
	t.Run("allows only explicit storage availability errors", func(t *testing.T) {
		for _, temporaryErr := range []error{
			driver.ErrBadConn,
			pgconn.ErrConnClosed,
			&pgconn.PgError{Code: "08006"},
			&pgconn.PgError{Code: "53300"},
			&pgconn.PgError{Code: "57P03"},
		} {
			assert.True(t, isTemporaryStaffValidationError(temporaryErr), "%T: %v", temporaryErr, temporaryErr)
		}
	})

	t.Run("rejects identity and programming errors", func(t *testing.T) {
		for _, permanentErr := range []error{
			nil,
			apperrors.WrapNotFound("staff", "123"),
			errors.New("generic repository error"),
			apperrors.WrapInvalidInput("configuration error"),
			&pgconn.PgError{Code: "23505"},
			&pgconn.PgError{Code: "42601"},
			&pgconn.PgError{Code: "08P01"},
		} {
			assert.False(t, isTemporaryStaffValidationError(permanentErr), "%T: %v", permanentErr, permanentErr)
		}
	})

	t.Run("recognizes wrapped typed errors", func(t *testing.T) {
		wrapped := apperrors.Wrap(driver.ErrBadConn, "database error")

		require.Error(t, wrapped)
		assert.True(t, isTemporaryStaffValidationError(wrapped))
	})
}

func TestAuth_StaffValidationConfigurationFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		userID     string
		resolver   authdomain.CurrentAccessResolver
		wantStatus int
	}{
		{
			name:       "nil staff validator is unavailable",
			userID:     "123",
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "malformed staff id is unauthorized",
			userID:     "not-a-number",
			resolver:   activeMiddlewareCurrentAccessResolver(),
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := validClaims()
			claims["user_id"] = test.userID
			claims["clinic_id"] = "1"
			claims["clinic_ids"] = []uint64{1}
			token := makeToken(t, jwt.SigningMethodHS256, claims)
			downstreamCalled := false
			recorder := httptest.NewRecorder()
			router := gin.New()
			router.Use(Auth(testTokenSvc(), false, nil, test.resolver))
			router.GET("/test", func(c *gin.Context) {
				downstreamCalled = true
				c.Status(http.StatusOK)
			})

			request := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			request.Header.Set("Authorization", "Bearer "+token)
			router.ServeHTTP(recorder, request)

			assert.Equal(t, test.wantStatus, recorder.Code)
			assert.False(t, downstreamCalled)
		})
	}
}

func TestAuth_RejectsRefreshTokenSubject(t *testing.T) {
	claims := validClaims()
	claims["sub"] = "refresh"
	token := makeToken(t, jwt.SigningMethodHS256, claims)

	w, _ := runAuthMiddleware(t, "Bearer "+token)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}
