package auth

import "testing"

// TestAuthUseCaseSurface keeps the domain-owned constructors visible while
// internal/service was a compatibility surface during the BE9 migration and has been retired (BE10, 2026-07-26).
func TestAuthUseCaseSurface(t *testing.T) {
	t.Helper()

	_ = NewAccountService
	_ = NewService
	_ = NewTokenService
	_ = NewTokenBlacklistService
	_ = NewPasswordResetService
	_ = NewPermissionGroupService
}
