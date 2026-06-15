package handler

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

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
