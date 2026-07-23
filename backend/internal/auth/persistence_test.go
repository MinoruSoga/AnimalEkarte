package auth

import "testing"

func TestPersistenceConstructors(t *testing.T) {
	t.Parallel()

	if NewAccountRepository(nil) == nil {
		t.Fatal("NewAccountRepository returned nil")
	}
	if NewPasswordResetTokenRepository(nil) == nil {
		t.Fatal("NewPasswordResetTokenRepository returned nil")
	}
	if NewTokenBlacklistRepository(nil) == nil {
		t.Fatal("NewTokenBlacklistRepository returned nil")
	}
	if NewPermissionGroupRepository(nil) == nil {
		t.Fatal("NewPermissionGroupRepository returned nil")
	}
}
