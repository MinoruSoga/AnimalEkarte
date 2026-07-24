package auth

import "testing"

// TestAuthUseCaseSurface keeps the domain-owned constructors visible while
// internal/service remains a compatibility surface during BE9 migration.
func TestAuthUseCaseSurface(t *testing.T) {
	t.Helper()

	_ = NewAccountService
	_ = NewAuthService
	_ = NewTokenService
	_ = NewTokenBlacklistService
	_ = NewPasswordResetService
	_ = NewPermissionGroupService
}
