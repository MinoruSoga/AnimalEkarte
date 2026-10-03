package seedlogin

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"os"
	"strings"

	"github.com/animal-ekarte/backend/internal/seedbundle"
)

// SharedPassword is the public demo login for local/dev/test catalog accounts.
// Production and staging never use it — staging accepts only the
// SEEDLOGIN_DEMO_PASSWORD secret (a repo-public value must not authenticate on
// an Internet-reachable deployed Worker).
const SharedPassword = "password" //nolint:gosec // G101: public non-production demo credential

// DemoPasswordEnv supplies the staging shared demo password as a secret.
// `wrangler secret put` only — unset locks staging demo logins (fail-closed).
const DemoPasswordEnv = "SEEDLOGIN_DEMO_PASSWORD" //nolint:gosec // G101: env var name, not a credential

// sharedPasswordForEnv resolves the shared demo password for appEnv.
// staging uses only the injected secret (empty = demo login closed).
// development/local/dev/test keep the public constant for DX.
func sharedPasswordForEnv(appEnv string) string {
	if normalizeEnv(appEnv) == "staging" {
		return strings.TrimSpace(os.Getenv(DemoPasswordEnv))
	}
	return SharedPassword
}

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
// login, and password equals the environment's shared demo password
// (SharedPassword locally; SEEDLOGIN_DEMO_PASSWORD on staging — unset there
// means the shortcut never matches). Operator and production accounts never
// match.
func AcceptSharedPassword(appEnv, email, password string) bool {
	if !ShouldApply(appEnv) {
		return false
	}
	if !IsCatalogEmail(email) {
		return false
	}
	expected := sharedPasswordForEnv(appEnv)
	if expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(password), []byte(expected)) == 1
}

// DemoPasswordFingerprint is a non-reversible marker of the environment's
// shared demo password. Mixed into the login-seed migration checksum so
// rotating (or removing) the secret re-upserts catalog hashes — otherwise
// checksum-skipped seeds would leave stale bcrypt hashes authenticating.
// "locked" is a stable sentinel for the fail-closed unset state.
func DemoPasswordFingerprint(appEnv string) string {
	password := sharedPasswordForEnv(appEnv)
	if password == "" {
		return "locked"
	}
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func normalizeEnv(appEnv string) string {
	return strings.ToLower(strings.TrimSpace(appEnv))
}
