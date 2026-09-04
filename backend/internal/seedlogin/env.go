package seedlogin

import (
	"crypto/subtle"
	"strings"

	"github.com/animal-ekarte/backend/internal/seedbundle"
)

// SharedPassword is the public demo login for local/STG catalog accounts.
// Production authentication must not accept it.
const SharedPassword = "password" //nolint:gosec // G101: public non-production demo credential

// MigrationKey is the schema_migrations.filename for the login upsert phase.
func MigrationKey() string {
	return seedbundle.BundleMigrationKey(BundleDir)
}

// ShouldApply reports whether APP_ENV may receive synthetic demo logins
// and the shared-password login shortcut. Production, empty, and unknown
// values stay fail-closed.
func ShouldApply(appEnv string) bool {
	switch normalizeEnv(appEnv) {
	case "development", "local", "dev", "test", "staging":
		return true
	default:
		return false
	}
}

// IsCatalogEmail reports whether email is one of the LoginForm demo logins.
func IsCatalogEmail(email string) bool {
	normalized := strings.TrimSpace(strings.ToLower(email))
	if normalized == "" {
		return false
	}
	for _, spec := range Catalog() {
		if spec.Email == normalized {
			return true
		}
	}
	return false
}

// AcceptSharedPassword is the non-production demo shortcut.
// It is true only when APP_ENV is allowlisted, email is a catalog demo
// login, and password equals SharedPassword. Operator and production
// accounts never match.
func AcceptSharedPassword(appEnv, email, password string) bool {
	if !ShouldApply(appEnv) {
		return false
	}
	if !IsCatalogEmail(email) {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(password), []byte(SharedPassword)) != 1 {
		return false
	}
	return true
}

func normalizeEnv(appEnv string) string {
	return strings.ToLower(strings.TrimSpace(appEnv))
}
