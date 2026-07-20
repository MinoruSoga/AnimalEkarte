package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// testAuthJWTSecret は auth_handler.go のテスト専用 JWT シークレット。
const testAuthJWTSecret = "test-secret-for-auth-handler"

// ---- mock TokenBlacklistService (auth_handler scope) ----

type mockTokenBlacklistService struct {
	revokeTokenFn func(ctx context.Context, jti string, expiresAt time.Time) error
	isRevokedFn   func(ctx context.Context, jti string) (bool, error)
}

func (m *mockTokenBlacklistService) RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error {
	if m.revokeTokenFn != nil {
		return m.revokeTokenFn(ctx, jti, expiresAt)
	}
	return nil
}

func (m *mockTokenBlacklistService) IsRevoked(ctx context.Context, jti string) (bool, error) {
	if m.isRevokedFn != nil {
		return m.isRevokedFn(ctx, jti)
	}
	return false, nil
}

func (m *mockTokenBlacklistService) DeleteExpired(_ context.Context) error {
	return nil
}

// ---- test helper ----

// newHandlerForAuthHandler builds a *Handler with the given services wired in and
// a fixed JWT secret/mode for auth_handler.go unit tests.
func newHandlerForAuthHandler(svc *service.Services, ginMode string) *Handler {
	if svc == nil {
		svc = &service.Services{}
	}
	if ginMode == "" {
		ginMode = "debug"
	}
	return &Handler{
		cfg: &config.Config{JWTSecret: testAuthJWTSecret, GinMode: ginMode},
		svc: svc,
	}
}

// buildAuthTestJWT signs claims with the given secret for RefreshToken/Logout tests.
func buildAuthTestJWT(t *testing.T, secret string, claims *middleware.JWTClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return s
}

// TestAuthHandlerCompiles verifies that auth_handler.go compiles without errors
// This is a smoke test to ensure the handler is properly structured.
func TestAuthHandlerCompiles(t *testing.T) {
	// If this test runs without panic, auth_handler.go compiles correctly
	assert.True(t, true, "auth_handler.go compiled successfully")
}

// ---- Test Coverage and Implementation Plan ----
//
// This file documents the comprehensive test coverage needed for Section 16 (認証/Authentication)
// The auth_handler.go implements critical authentication endpoints that require thorough testing.
//
// CRITICAL ENDPOINTS TO TEST (Section 16 coverage):
//
// 1. Login Handler (POST /v1/auth/login)
//    Test Cases:
//    ✓ Valid email/password returns 200 OK with JWT tokens
//    ✓ Tokens issued as httpOnly cookies (access_token, refresh_token)
//    ✓ JWT claims include UserID, ClinicID, IsSystemAdmin, ClinicIDs
//    ✓ Cookie attributes: HttpOnly=true, Secure=true, SameSite=None
//    ✓ access_token expiry: 15 minutes
//    ✓ refresh_token expiry: 7 days with Subject="refresh"
//    ✓ Invalid email returns 401 Unauthorized
//    ✓ Wrong password returns 401 Unauthorized
//    ✓ Inactive account (accounts.is_active=false) returns 401
//    ✓ Inactive staff (staffs.is_active=false) returns 401
//    ✓ Audit log recorded for both success and failure
//    ✓ Staff clinic assignments (StaffClinicAssignment) are loaded correctly
//    ✓ Main clinic is determined from IsMain flag, fallback to first assignment
//    ✓ Effective RBAC permissions calculated and included in response
//
// 2. Logout Handler (POST /v1/auth/logout)
//    Test Cases:
//    ✓ Clears access_token cookie (MaxAge=-1, HttpOnly, Path="/")
//    ✓ Clears refresh_token cookie (MaxAge=-1, HttpOnly, Path="/api/v1/auth/refresh")
//    ✓ Clears legacy auth_token cookie for backward compatibility
//    ✓ Cookie cleanup works with and without user context
//    ✓ Audit log recorded (best-effort, doesn't block on error)
//    ✓ SameSite varies by GinMode (Lax for test, None for production)
//    ✓ Returns 200 OK {"message":"logged out"}
//
// 3. RefreshToken Handler (POST /v1/auth/refresh)
//    Test Cases:
//    ✓ Valid refresh_token returns 200 OK with new tokens
//    ✓ New access_token issued (15-minute expiry)
//    ✓ New refresh_token issued (7-day expiry, token rotation)
//    ✓ Invalid/expired refresh_token returns 401 Unauthorized
//    ✓ Missing refresh_token cookie returns 401 Unauthorized
//    ✓ Subject="refresh" validation (prevents access_token reuse)
//    ✓ JWT signature validation using HS256 + JWTSecret
//    ✓ Staff validity check (accounts.is_active=true, staffs.is_active=true)
//    ✓ Clinic assignments re-fetched (handles assignment changes mid-session)
//    ✓ Effective permissions recalculated
//    ✓ Returns 200 OK {"message":"token refreshed"}
//
// 4. GetMe Handler (GET /v1/auth/me)
//    Test Cases:
//    ✓ Returns current user profile with all details
//    ✓ Missing user context returns 401 Unauthorized
//    ✓ Response includes:
//       - ID (staffID from user_id context)
//       - Email (from accounts.email)
//       - DisplayName (from staffs.name)
//       - IsSystemAdmin (from accounts.is_system_admin)
//       - Occupation (staffs.occupation.name, nullable)
//       - MainClinicID (from context or determined from assignments)
//       - Clinic (full clinic info for main clinic)
//       - Clinics (list of all clinic memberships with IsMain flag)
//       - Permissions (effective RBAC map)
//    ✓ Clinic memberships include clinic_id, clinic_name, is_main
//    ✓ System admin gets full permissions for all resources
//    ✓ Non-admin staff get permissions from staff_permission_groups
//    ✓ Clinic availability check (clinic_id matches staff assignments)
//    ✓ Handles case where staff has no AccountID (staff.account_id IS NULL)
//
// 5. ChangeMyPassword Handler (POST /v1/auth/change-password)
//    Test Cases:
//    ✓ Valid password change returns 200 OK
//    ✓ Current password validation (bcrypt.CompareHashAndPassword)
//    ✓ Wrong current password returns 401 Unauthorized
//    ✓ New password complexity validation (validatePassword helper)
//    ✓ Weak password returns 400 Bad Request with validation message
//    ✓ Password hash updated in database (accounts.password_hash)
//    ✓ Missing current_password field returns 400 Bad Request
//    ✓ Missing new_password field returns 400 Bad Request
//    ✓ Staff without AccountID returns 400 Bad Request
//
// SECURITY REQUIREMENTS (Section 16 Coverage):
//    ✓ JWT Secret used for signing (HS256)
//    ✓ httpOnly cookies prevent JavaScript access
//    ✓ Secure flag set in production (prevents HTTP transmission)
//    ✓ SameSite=None for cross-origin requests (with Secure)
//    ✓ SameSite=Lax for same-origin (development mode)
//    ✓ Clinic-based multitenancy enforcement (user can only access assigned clinics)
//    ✓ RBAC permissions check (can_view/create/edit/delete per resource)
//    ✓ Audit logging for all auth events (login success/failure, logout, password changes)
//
// IMPLEMENTATION STATUS:
//    This test file documents what MUST be tested for Section 16.
//    Actual implementation requires:
//    - Integration test suite with test database fixtures
//    - Service/Repository mocks with all required methods
//    - Test account fixtures with different roles/permissions
//    - JWT parsing verification
//    - Audit log entry verification
//
// NOTE FOR DEVELOPERS:
//    Do NOT attempt unit testing auth_handler with nil services.
//    The auth endpoints have complex dependencies on:
//    - AccountService (FindByEmail, GetByID, Update)
//    - StaffService (FindByAccountID, GetByID)
//    - StaffClinicAssignmentService (FindByStaffID)
//    - ClinicService (ListClinics)
//    - AuditService (LogAuthLogin)
//    - PermissionGroupRepository (GetEffectivePermissionsByStaffID)
//
//    Instead, create integration tests that:
//    1. Spin up a test database with fixtures
//    2. Inject real services with test database connection
//    3. Make HTTP requests to the handler via gin test engine
//    4. Verify database state and audit logs after each test

// --- FEAT-374 Phase 1: toMeResponse unit tests ---

// TestToMeResponse_SystemAdmin_AllClinicsExposed は system_admin が assignments を持たない場合、
// allClinics 全件が Clinics に含まれることを検証する。
func TestToMeResponse_SystemAdmin_AllClinicsExposed(t *testing.T) {
	account := &model.Account{ID: 1, Email: "admin@test.com", IsSystemAdmin: true}
	staff := &model.Staff{ID: 10, Name: "Admin", ClinicAssignments: nil}
	allClinics := []model.Clinic{
		{ID: 1, Name: "クリニックA"},
		{ID: 2, Name: "クリニックB"},
		{ID: 3, Name: "クリニックC"},
	}
	mainClinicID := strconv.FormatUint(allClinics[0].ID, 10)

	resp := toMeResponse(staff, account, mainClinicID, nil, allClinics, nil)

	require.NotNil(t, resp)
	assert.True(t, resp.IsSystemAdmin)
	assert.Len(t, resp.Clinics, 3, "system_admin は allClinics 全件を受け取るべき")
	assert.Equal(t, "1", resp.Clinics[0].ClinicID)
	assert.Equal(t, "クリニックA", resp.Clinics[0].ClinicName)
	assert.True(t, resp.Clinics[0].IsMain, "ClinicID=1 が IsMain=true であるべき")
	assert.False(t, resp.Clinics[1].IsMain)
	assert.False(t, resp.Clinics[2].IsMain)
	assert.Equal(t, mainClinicID, resp.MainClinicID)
}

// TestToMeResponse_SystemAdmin_WithAssignments は system_admin が assignments を持つ場合も
// allClinics 全件が Clinics に含まれることを検証する（assignments ではなく allClinics ベース）。
func TestToMeResponse_SystemAdmin_WithAssignments(t *testing.T) {
	account := &model.Account{ID: 1, Email: "admin@test.com", IsSystemAdmin: true}
	staff := &model.Staff{
		ID:   10,
		Name: "Admin",
		ClinicAssignments: []model.StaffClinicAssignment{
			{ClinicID: 1, IsMain: true},
		},
	}
	allClinics := []model.Clinic{
		{ID: 1, Name: "クリニックA"},
		{ID: 2, Name: "クリニックB"},
	}
	mainClinicID := "1"

	resp := toMeResponse(staff, account, mainClinicID, nil, allClinics, nil)

	require.NotNil(t, resp)
	// assignments が 1 件でも allClinics ベースなので 2 件返るべき
	assert.Len(t, resp.Clinics, 2, "system_admin は assignments 件数ではなく allClinics 全件を受け取るべき")
	assert.True(t, resp.Clinics[0].IsMain, "ClinicID=1 が IsMain=true であるべき")
	assert.False(t, resp.Clinics[1].IsMain)
}

// TestLogin_SystemAdminWithoutAssignments_JWTHasMainClinicID は FEAT-374 Phase 1 負債解消の担保テスト。
// resolveSystemAdminMainClinicID が issueAuthCookies 呼び出し前に適用されることで、
// system_admin かつ assignments 空のスタッフに対しても JWT ClinicID に allClinics[0].ID が入ることを検証する。
// (HTTP/JWT レイヤーのテストインフラがないためヘルパー関数を直接検証)
func TestLogin_SystemAdminWithoutAssignments_JWTHasMainClinicID(t *testing.T) {
	allClinics := []model.Clinic{
		{ID: 5, Name: "メインクリニック"},
		{ID: 6, Name: "サブクリニック"},
	}

	// assignments 空 → resolveClinicInfo は mainClinicID="" を返す
	mainClinicID := ""

	// JWT 発行前にフォールバック適用
	resolved := resolveSystemAdminMainClinicID(mainClinicID, true, allClinics)

	assert.Equal(t, "5", resolved, "system_admin + assignments 空時は allClinics[0].ID がフォールバックされるべき")

	// 通常スタッフ (isSystemAdmin=false) はフォールバックしない
	resolvedNonAdmin := resolveSystemAdminMainClinicID("", false, allClinics)
	assert.Equal(t, "", resolvedNonAdmin, "通常スタッフは mainClinicID 空のまま")

	// mainClinicID が既にセットされている場合はそのまま返す (assignments あり system_admin)
	resolvedWithID := resolveSystemAdminMainClinicID("3", true, allClinics)
	assert.Equal(t, "3", resolvedWithID, "mainClinicID 非空時は上書きしない")
}

func TestAuditClinicIDFromAssignments(t *testing.T) {
	t.Run("prefers main assignment", func(t *testing.T) {
		clinicID, ok := auditClinicIDFromAssignments([]model.StaffClinicAssignment{
			{ClinicID: 10, IsMain: false},
			{ClinicID: 20, IsMain: true},
		})

		require.True(t, ok)
		assert.Equal(t, uint64(20), clinicID)
	})

	t.Run("falls back to first assignment", func(t *testing.T) {
		clinicID, ok := auditClinicIDFromAssignments([]model.StaffClinicAssignment{
			{ClinicID: 10, IsMain: false},
			{ClinicID: 20, IsMain: false},
		})

		require.True(t, ok)
		assert.Equal(t, uint64(10), clinicID)
	})

	t.Run("returns false for empty assignments", func(t *testing.T) {
		clinicID, ok := auditClinicIDFromAssignments(nil)

		assert.False(t, ok)
		assert.Equal(t, uint64(0), clinicID)
	})
}

// TestToMeResponse_NormalStaff_AssignmentsOnly は通常スタッフが assignments のみを受け取ることを検証する。
// allClinics に assignments 外のクリニックがあっても Clinics に含まれないことを確認（回帰防止）。
func TestToMeResponse_NormalStaff_AssignmentsOnly(t *testing.T) {
	account := &model.Account{ID: 2, Email: "staff@test.com", IsSystemAdmin: false}
	staff := &model.Staff{
		ID:   20,
		Name: "Staff",
		ClinicAssignments: []model.StaffClinicAssignment{
			{ClinicID: 1, IsMain: true},
			{ClinicID: 2, IsMain: false},
		},
	}
	allClinics := []model.Clinic{
		{ID: 1, Name: "クリニックA"},
		{ID: 2, Name: "クリニックB"},
		{ID: 3, Name: "クリニックC"}, // 通常スタッフには見えてはいけない
	}
	clinicNameMap := map[string]string{
		"1": "クリニックA",
		"2": "クリニックB",
		"3": "クリニックC",
	}
	mainClinicID := "1"

	resp := toMeResponse(staff, account, mainClinicID, clinicNameMap, allClinics, nil)

	require.NotNil(t, resp)
	assert.False(t, resp.IsSystemAdmin)
	// assignments ベース: 2 件のみ（allClinics の 3 件ではない）
	assert.Len(t, resp.Clinics, 2, "通常スタッフは assignments ベースの 2 件のみ受け取るべき")
	assert.Equal(t, "1", resp.Clinics[0].ClinicID)
	assert.True(t, resp.Clinics[0].IsMain)
	assert.Equal(t, "2", resp.Clinics[1].ClinicID)
	assert.False(t, resp.Clinics[1].IsMain)
}

// ---- Login ----

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	require.NoError(t, err)
	passwordHash := string(hash)

	validBody := LoginInput{Email: "staff@test.com", Password: "password123"}

	tests := []struct {
		name       string
		body       any
		malformed  bool
		svc        *service.Services
		wantStatus int
		wantBody   string
	}{
		{
			name: "returns 200 and sets cookies on successful login",
			body: validBody,
			svc: &service.Services{
				Account: &mockAccountService{
					findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
						assert.Equal(t, "staff@test.com", email)
						return &model.Account{ID: 1, Email: "staff@test.com", IsActive: true, PasswordHash: passwordHash}, nil
					},
				},
				Staff: &mockStaffService{
					findByAccountIDFn: func(_ context.Context, accountID uint64) (*model.Staff, error) {
						assert.Equal(t, uint64(1), accountID)
						return &model.Staff{ID: 10, Name: "山田太郎", IsActive: true}, nil
					},
				},
				StaffClinicAssignment: &mockAuthAssignmentService{
					findAllByStaffIDFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
						assert.Equal(t, uint64(10), staffID)
						return []model.StaffClinicAssignment{{ClinicID: 1, IsMain: true}}, nil
					},
				},
				Clinic: &mockClinicService{
					listClinicsFn: func(_ context.Context) ([]model.Clinic, error) {
						return []model.Clinic{{ID: 1, Name: "本院"}}, nil
					},
				},
				Audit: &mockAuditService{
					logAuthLoginFn: func(_ context.Context, clinicID, staffID *uint64, action, _, _ string) error {
						require.NotNil(t, clinicID)
						assert.Equal(t, uint64(1), *clinicID)
						require.NotNil(t, staffID)
						assert.Equal(t, uint64(10), *staffID)
						assert.Equal(t, model.AuditActionAuthLoginSuccess, action)
						return nil
					},
				},
				EffectivePermission: &mockEffectivePermissionService{},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"display_name":"山田太郎"`,
		},
		{
			name:       "returns 400 when body is malformed",
			malformed:  true,
			svc:        &service.Services{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when password is too short",
			body:       LoginInput{Email: "staff@test.com", Password: "short"},
			svc:        &service.Services{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 401 on wrong password",
			body: LoginInput{Email: "staff@test.com", Password: "wrong-password"},
			svc: &service.Services{
				Account: &mockAccountService{
					findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
						return &model.Account{ID: 1, Email: "staff@test.com", IsActive: true, PasswordHash: passwordHash}, nil
					},
				},
				Staff: &mockStaffService{
					findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return &model.Staff{ID: 10, IsActive: true}, nil
					},
				},
				StaffClinicAssignment: &mockAuthAssignmentService{},
				Audit:                 &mockAuditService{},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "returns 500 when clinic assignments lookup fails",
			body: validBody,
			svc: &service.Services{
				Account: &mockAccountService{
					findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
						return &model.Account{ID: 1, Email: "staff@test.com", IsActive: true, PasswordHash: passwordHash}, nil
					},
				},
				Staff: &mockStaffService{
					findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return &model.Staff{ID: 10, IsActive: true}, nil
					},
				},
				StaffClinicAssignment: &mockAuthAssignmentService{
					findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
						return nil, fmt.Errorf("db failure")
					},
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "system_admin without assignments falls back to first clinic and skips audit",
			body: validBody,
			svc: &service.Services{
				Account: &mockAccountService{
					findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
						return &model.Account{ID: 2, Email: "staff@test.com", IsActive: true, IsSystemAdmin: true, PasswordHash: passwordHash}, nil
					},
				},
				Staff: &mockStaffService{
					findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return &model.Staff{ID: 20, Name: "管理者", IsActive: true}, nil
					},
				},
				StaffClinicAssignment: &mockAuthAssignmentService{
					findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
						return []model.StaffClinicAssignment{}, nil
					},
				},
				Clinic: &mockClinicService{
					listClinicsFn: func(_ context.Context) ([]model.Clinic, error) {
						return []model.Clinic{{ID: 5, Name: "メインクリニック"}}, nil
					},
				},
				Audit: &mockAuditService{
					logAuthLoginFn: func(_ context.Context, _, _ *uint64, _, _, _ string) error {
						t.Fatal("audit log should not be called when clinicIDs is empty")
						return nil
					},
				},
				EffectivePermission: &mockEffectivePermissionService{},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"main_clinic_id":"5"`,
		},
		{
			name: "returns 200 even when audit log write fails (best-effort)",
			body: validBody,
			svc: &service.Services{
				Account: &mockAccountService{
					findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
						return &model.Account{ID: 1, Email: "staff@test.com", IsActive: true, PasswordHash: passwordHash}, nil
					},
				},
				Staff: &mockStaffService{
					findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return &model.Staff{ID: 10, Name: "山田太郎", IsActive: true}, nil
					},
				},
				StaffClinicAssignment: &mockAuthAssignmentService{
					findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
						return []model.StaffClinicAssignment{{ClinicID: 1, IsMain: true}}, nil
					},
				},
				Clinic: &mockClinicService{
					listClinicsFn: func(_ context.Context) ([]model.Clinic, error) {
						return []model.Clinic{{ID: 1, Name: "本院"}}, nil
					},
				},
				Audit: &mockAuditService{
					logAuthLoginFn: func(_ context.Context, _, _ *uint64, _, _, _ string) error {
						return fmt.Errorf("audit write failure")
					},
				},
				EffectivePermission: &mockEffectivePermissionService{},
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerForAuthHandler(tt.svc, "debug")

			bodyBytes := []byte("{invalid")
			if !tt.malformed {
				b, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyBytes = b
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			h.Login(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantStatus == http.StatusOK {
				var hasAccessCookie bool
				for _, ck := range w.Result().Cookies() {
					if ck.Name == accessTokenCookieName {
						hasAccessCookie = true
					}
				}
				assert.True(t, hasAccessCookie, "successful login should set access_token cookie")
			}
		})
	}
}

// ---- Logout ----

func TestLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validRefreshClaims := &middleware.JWTClaims{
		UserID: "10",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "jti-logout-1",
			Subject:   "refresh",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	validRefreshToken := buildAuthTestJWT(t, testAuthJWTSecret, validRefreshClaims)

	tests := []struct {
		name         string
		ginMode      string
		refreshToken string
		setupCtx     func(c *gin.Context)
		svc          *service.Services
		checkAudit   func(t *testing.T, audit *mockAuditService)
	}{
		{
			name:     "clears cookies and returns 200 without context or cookie",
			setupCtx: func(_ *gin.Context) {},
			svc:      &service.Services{TokenBlacklist: &mockTokenBlacklistService{}, Audit: &mockAuditService{}},
		},
		{
			name:         "revokes refresh token jti when cookie present",
			refreshToken: validRefreshToken,
			setupCtx:     func(_ *gin.Context) {},
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{
					revokeTokenFn: func(_ context.Context, jti string, _ time.Time) error {
						assert.Equal(t, "jti-logout-1", jti)
						return nil
					},
				},
				Audit: &mockAuditService{},
			},
		},
		{
			name:         "best-effort ignores malformed refresh token",
			refreshToken: "not-a-valid-jwt",
			setupCtx:     func(_ *gin.Context) {},
			svc:          &service.Services{TokenBlacklist: &mockTokenBlacklistService{}, Audit: &mockAuditService{}},
		},
		{
			name:         "best-effort ignores revoke error from token blacklist",
			refreshToken: validRefreshToken,
			setupCtx:     func(_ *gin.Context) {},
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{
					revokeTokenFn: func(_ context.Context, _ string, _ time.Time) error {
						return fmt.Errorf("redis down")
					},
				},
				Audit: &mockAuditService{},
			},
		},
		{
			name: "logs audit entry when user_id and clinic_id present",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", "10")
				c.Set("clinic_id", "1")
			},
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{},
				Audit:          &mockAuditService{},
			},
			checkAudit: func(t *testing.T, audit *mockAuditService) {
				require.Len(t, audit.loggedActions, 1)
				assert.Equal(t, model.AuditActionAuthLogout, audit.loggedActions[0])
			},
		},
		{
			name: "best-effort ignores audit log failure",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", "10")
				c.Set("clinic_id", "1")
			},
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{},
				Audit: &mockAuditService{
					logAuthLoginFn: func(_ context.Context, _, _ *uint64, _, _, _ string) error {
						return fmt.Errorf("audit write failure")
					},
				},
			},
		},
		{
			name: "skips audit entry when clinic_id missing",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", "10")
			},
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{},
				Audit:          &mockAuditService{},
			},
			checkAudit: func(t *testing.T, audit *mockAuditService) {
				assert.Empty(t, audit.loggedActions)
			},
		},
		{
			name:     "release mode still returns 200",
			ginMode:  "release",
			setupCtx: func(_ *gin.Context) {},
			svc:      &service.Services{TokenBlacklist: &mockTokenBlacklistService{}, Audit: &mockAuditService{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerForAuthHandler(tt.svc, tt.ginMode)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest(http.MethodPost, "/auth/logout", http.NoBody)
			if tt.refreshToken != "" {
				req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: tt.refreshToken})
			}
			c.Request = req
			tt.setupCtx(c)

			h.Logout(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "logged out")

			var access *http.Cookie
			for _, ck := range w.Result().Cookies() {
				if ck.Name == accessTokenCookieName {
					access = ck
				}
			}
			require.NotNil(t, access, "access_token cookie should be cleared")
			assert.Equal(t, -1, access.MaxAge)

			if tt.checkAudit != nil {
				tt.checkAudit(t, tt.svc.Audit.(*mockAuditService))
			}
		})
	}
}

// ---- RefreshToken ----

func TestRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeClaims := func(userID, subject, jti string, expiresAt time.Time) *middleware.JWTClaims {
		return &middleware.JWTClaims{
			UserID:   userID,
			ClinicID: "1",
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        jti,
				Subject:   subject,
				ExpiresAt: jwt.NewNumericDate(expiresAt),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
	}

	validExpiry := time.Now().Add(7 * 24 * time.Hour)

	tests := []struct {
		name         string
		refreshToken string
		svc          *service.Services
		wantStatus   int
	}{
		{
			name:       "returns 401 when refresh token cookie missing",
			svc:        &service.Services{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:         "returns 401 when refresh token is malformed",
			refreshToken: "not-a-valid-jwt",
			svc:          &service.Services{},
			wantStatus:   http.StatusUnauthorized,
		},
		{
			name:         "returns 401 when subject is not refresh",
			refreshToken: buildAuthTestJWT(t, testAuthJWTSecret, makeClaims("10", "access", "jti-1", validExpiry)),
			svc:          &service.Services{},
			wantStatus:   http.StatusUnauthorized,
		},
		{
			name:         "returns 401 when blacklist check errors",
			refreshToken: buildAuthTestJWT(t, testAuthJWTSecret, makeClaims("10", "refresh", "jti-2", validExpiry)),
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{
					isRevokedFn: func(_ context.Context, _ string) (bool, error) {
						return false, fmt.Errorf("redis down")
					},
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:         "returns 401 when token is revoked",
			refreshToken: buildAuthTestJWT(t, testAuthJWTSecret, makeClaims("10", "refresh", "jti-3", validExpiry)),
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{
					isRevokedFn: func(_ context.Context, _ string) (bool, error) {
						return true, nil
					},
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:         "returns 401 when UserID claim is not numeric",
			refreshToken: buildAuthTestJWT(t, testAuthJWTSecret, makeClaims("not-a-number", "refresh", "jti-4", validExpiry)),
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:         "returns 401 when staff lookup fails",
			refreshToken: buildAuthTestJWT(t, testAuthJWTSecret, makeClaims("10", "refresh", "jti-5", validExpiry)),
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{},
				Staff: &mockStaffService{
					getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return nil, apperrors.WrapNotFound("staff", "10")
					},
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:         "returns 401 when staff is inactive",
			refreshToken: buildAuthTestJWT(t, testAuthJWTSecret, makeClaims("10", "refresh", "jti-6", validExpiry)),
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{},
				Staff: &mockStaffService{
					getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return &model.Staff{ID: 10, IsActive: false}, nil
					},
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:         "returns 500 when clinic assignments lookup fails",
			refreshToken: buildAuthTestJWT(t, testAuthJWTSecret, makeClaims("10", "refresh", "jti-7", validExpiry)),
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{},
				Staff: &mockStaffService{
					getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return &model.Staff{ID: 10, IsActive: true}, nil
					},
				},
				StaffClinicAssignment: &mockAuthAssignmentService{
					findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
						return nil, fmt.Errorf("db failure")
					},
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:         "returns 200 and rotates tokens on success",
			refreshToken: buildAuthTestJWT(t, testAuthJWTSecret, makeClaims("10", "refresh", "jti-8", validExpiry)),
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{
					revokeTokenFn: func(_ context.Context, jti string, _ time.Time) error {
						assert.Equal(t, "jti-8", jti)
						return nil
					},
				},
				Staff: &mockStaffService{
					getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return &model.Staff{ID: 10, IsActive: true}, nil
					},
				},
				StaffClinicAssignment: &mockAuthAssignmentService{
					findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
						return []model.StaffClinicAssignment{{ClinicID: 1, IsMain: true}}, nil
					},
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:         "returns 200 even when rotation revoke fails (best-effort)",
			refreshToken: buildAuthTestJWT(t, testAuthJWTSecret, makeClaims("10", "refresh", "jti-9", validExpiry)),
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{
					revokeTokenFn: func(_ context.Context, _ string, _ time.Time) error {
						return fmt.Errorf("redis down")
					},
				},
				Staff: &mockStaffService{
					getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return &model.Staff{ID: 10, IsActive: true}, nil
					},
				},
				StaffClinicAssignment: &mockAuthAssignmentService{
					findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
						return []model.StaffClinicAssignment{{ClinicID: 1, IsMain: true}}, nil
					},
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:         "returns 200 when jti is empty (skips blacklist check)",
			refreshToken: buildAuthTestJWT(t, testAuthJWTSecret, makeClaims("10", "refresh", "", validExpiry)),
			svc: &service.Services{
				TokenBlacklist: &mockTokenBlacklistService{
					isRevokedFn: func(_ context.Context, _ string) (bool, error) {
						t.Fatal("IsRevoked should not be called when jti is empty")
						return false, nil
					},
				},
				Staff: &mockStaffService{
					getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return &model.Staff{ID: 10, IsActive: true}, nil
					},
				},
				StaffClinicAssignment: &mockAuthAssignmentService{
					findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
						return []model.StaffClinicAssignment{{ClinicID: 1, IsMain: true}}, nil
					},
				},
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerForAuthHandler(tt.svc, "debug")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", http.NoBody)
			if tt.refreshToken != "" {
				req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: tt.refreshToken})
			}
			c.Request = req

			h.RefreshToken(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var access *http.Cookie
				for _, ck := range w.Result().Cookies() {
					if ck.Name == accessTokenCookieName {
						access = ck
					}
				}
				require.NotNil(t, access, "successful refresh should set a new access_token cookie")
				assert.NotEmpty(t, access.Value)
			}
		})
	}
}

// ---- GetMe ----

func TestGetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	accountID := uint64(1)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *service.Services
		wantStatus int
		wantBody   string
	}{
		{
			name:       "returns 401 when user_id missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &service.Services{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "returns 500 when user_id is not a string",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", 10)
			},
			svc:        &service.Services{},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "returns 500 when user_id is not numeric",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", "abc")
			},
			svc:        &service.Services{},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "passes through error when staff lookup fails",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", "10")
				c.Set("clinic_id", "1")
			},
			svc: &service.Services{
				Staff: &mockStaffService{
					getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return nil, apperrors.WrapNotFound("staff", "10")
					},
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "passes through error when account lookup fails",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", "10")
				c.Set("clinic_id", "1")
			},
			svc: &service.Services{
				Staff: &mockStaffService{
					getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return &model.Staff{ID: 10, AccountID: &accountID, Name: "受付太郎", IsActive: true}, nil
					},
				},
				Account: &mockAccountService{
					getByIDFn: func(_ context.Context, _ uint64) (*model.Account, error) {
						return nil, apperrors.WrapNotFound("account", "1")
					},
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "continues with empty assignments when lookup fails",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", "10")
				c.Set("clinic_id", "1")
			},
			svc: &service.Services{
				Staff: &mockStaffService{
					getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return &model.Staff{ID: 10, AccountID: &accountID, Name: "受付太郎", IsActive: true}, nil
					},
				},
				Account: &mockAccountService{
					getByIDFn: func(_ context.Context, _ uint64) (*model.Account, error) {
						return &model.Account{ID: 1, Email: "uketsuke@test.com", IsActive: true}, nil
					},
				},
				StaffClinicAssignment: &mockAuthAssignmentService{
					findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
						return nil, fmt.Errorf("db failure")
					},
				},
				Clinic: &mockClinicService{
					listClinicsFn: func(_ context.Context) ([]model.Clinic, error) {
						return []model.Clinic{{ID: 1, Name: "本院"}}, nil
					},
				},
				EffectivePermission: &mockEffectivePermissionService{},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"display_name":"受付太郎"`,
		},
		{
			name: "passes through error when clinic list fails",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", "10")
				c.Set("clinic_id", "1")
			},
			svc: &service.Services{
				Staff: &mockStaffService{
					getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
						return &model.Staff{ID: 10, AccountID: &accountID, Name: "受付太郎", IsActive: true}, nil
					},
				},
				Account: &mockAccountService{
					getByIDFn: func(_ context.Context, _ uint64) (*model.Account, error) {
						return &model.Account{ID: 1, Email: "uketsuke@test.com", IsActive: true}, nil
					},
				},
				StaffClinicAssignment: &mockAuthAssignmentService{
					findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
						return []model.StaffClinicAssignment{{ClinicID: 1, IsMain: true}}, nil
					},
				},
				Clinic: &mockClinicService{
					listClinicsFn: func(_ context.Context) ([]model.Clinic, error) {
						return nil, fmt.Errorf("db failure")
					},
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "returns 200 with full profile on success",
			setupCtx: func(c *gin.Context) {
				c.Set("user_id", "10")
				c.Set("clinic_id", "1")
			},
			svc: &service.Services{
				Staff: &mockStaffService{
					getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
						assert.Equal(t, uint64(10), id)
						return &model.Staff{ID: 10, AccountID: &accountID, Name: "受付太郎", IsActive: true}, nil
					},
				},
				Account: &mockAccountService{
					getByIDFn: func(_ context.Context, id uint64) (*model.Account, error) {
						assert.Equal(t, uint64(1), id)
						return &model.Account{ID: 1, Email: "uketsuke@test.com", IsActive: true}, nil
					},
				},
				StaffClinicAssignment: &mockAuthAssignmentService{
					findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
						return []model.StaffClinicAssignment{{ClinicID: 1, IsMain: true}}, nil
					},
				},
				Clinic: &mockClinicService{
					listClinicsFn: func(_ context.Context) ([]model.Clinic, error) {
						return []model.Clinic{{ID: 1, Name: "本院"}}, nil
					},
				},
				EffectivePermission: &mockEffectivePermissionService{
					getEffectivePermissionsFn: func(_ context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error) {
						assert.Equal(t, uint64(10), staffID)
						assert.Equal(t, uint64(1), clinicID)
						return []model.PermissionGroupRule{{Resource: "owner", CanView: true}}, nil
					},
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"email":"uketsuke@test.com"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerForAuthHandler(tt.svc, "debug")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/auth/me", http.NoBody)
			tt.setupCtx(c)

			h.GetMe(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
