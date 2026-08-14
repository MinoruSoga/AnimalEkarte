package main

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/seedbundle"
)

func TestSeedBundlesForEnv_ProductionPlanExcludesDemo(t *testing.T) {
	t.Parallel()

	got := seedBundlesForEnv("production")
	if !slices.Equal(got, []string{"002_master"}) {
		t.Fatalf("production plan = %v, want master only", got)
	}
	for _, banned := range []string{"003_demo", "004_staging"} {
		if slices.Contains(got, banned) {
			t.Fatalf("production plan must not include %s: %v", banned, got)
		}
	}
}

// SEC-CS2-F01: staging must not receive privileged demo credentials via migrate.
func TestSeedBundlesForEnv_StagingIsMasterOnly(t *testing.T) {
	t.Parallel()

	for _, env := range []string{"staging", "STAGING", " staging "} {
		got := seedBundlesForEnv(env)
		if !slices.Equal(got, []string{"002_master"}) {
			t.Fatalf("env %q plan = %v, want master only", env, got)
		}
		for _, banned := range []string{"003_demo", "004_staging"} {
			if slices.Contains(got, banned) {
				t.Fatalf("staging plan must not include %s: %v", banned, got)
			}
		}
	}
}

func TestSeedBundlesForEnv_LocalDevAndTestIncludesDemo(t *testing.T) {
	t.Parallel()

	for _, env := range []string{"development", "local", "test", "dev"} {
		got := seedBundlesForEnv(env)
		want := []string{"002_master", "003_demo", "004_staging"}
		if !slices.Equal(got, want) {
			t.Fatalf("env %q plan = %v, want %v", env, got, want)
		}
	}
}

func TestSeedBundlesForEnv_EmptyAndUnknownFailClosed(t *testing.T) {
	t.Parallel()

	for _, env := range []string{"", "preview", "prod", "PRODUCTION"} {
		got := seedBundlesForEnv(env)
		if !slices.Equal(got, []string{"002_master"}) {
			t.Fatalf("env %q must fail-closed to master only, got %v", env, got)
		}
	}
}

func TestSeedBundlesForCurrentEnv_ReadsAPP_ENV(t *testing.T) {
	// Not parallel: mutates process environment.
	t.Setenv("APP_ENV", "production")
	got := seedBundlesForCurrentEnv()
	if !slices.Equal(got, []string{"002_master"}) {
		t.Fatalf("APP_ENV=production plan = %v, want master only", got)
	}

	t.Setenv("APP_ENV", "staging")
	got = seedBundlesForCurrentEnv()
	if !slices.Equal(got, []string{"002_master"}) {
		t.Fatalf("APP_ENV=staging plan = %v, want master only", got)
	}

	t.Setenv("APP_ENV", "development")
	got = seedBundlesForCurrentEnv()
	want := []string{"002_master", "003_demo", "004_staging"}
	if !slices.Equal(got, want) {
		t.Fatalf("APP_ENV=development plan = %v, want %v", got, want)
	}

	// Unset → fail-closed.
	if err := os.Unsetenv("APP_ENV"); err != nil {
		t.Fatalf("Unsetenv: %v", err)
	}
	got = seedBundlesForCurrentEnv()
	if !slices.Equal(got, []string{"002_master"}) {
		t.Fatalf("unset APP_ENV plan = %v, want master only", got)
	}
}

// TestDemoBundleHasActiveSystemAdminDocumentsIntentionalDemoOnly asserts that
// the demo seed still contains active system-admin accounts. That is intentional
// for local development/test only; production and staging paths must not load
// this bundle (see TestSeedBundlesForEnv_ProductionPlanExcludesDemo and
// TestSeedBundlesForEnv_StagingIsMasterOnly). This test does not read or
// assert any credential material.
func TestDemoBundleHasActiveSystemAdminDocumentsIntentionalDemoOnly(t *testing.T) {
	t.Parallel()

	// Resolve seeds relative to this source file (cwd-independent for CI working-directory=backend).
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	accountsPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "seeds", "003_demo", "accounts.csv")
	raw, err := os.ReadFile(accountsPath)
	if err != nil {
		t.Fatalf("read demo accounts.csv: %v", err)
	}
	// Header contract: id,email,password_hash,is_active,is_system_admin,...
	// ",t,t," matches "...hash,t,t,timestamp..." without decoding credentials.
	if !strings.Contains(string(raw), ",t,t,") {
		t.Fatal("003_demo accounts.csv is expected to contain at least one active system admin (is_active=t, is_system_admin=t); demo-only intentional")
	}

	// Production and staging plans must still exclude the demo bundle that carries those accounts.
	for _, env := range []string{"production", "staging"} {
		if slices.Contains(seedBundlesForEnv(env), "003_demo") {
			t.Fatalf("%s seed plan must not load 003_demo despite demo CSV containing system admins", env)
		}
	}
}

func TestBundleOrderForEnv_MatchesSeedBundlesForEnv(t *testing.T) {
	t.Parallel()

	// migrate helper must stay a thin wrapper over seedbundle (single source of truth).
	for _, env := range []string{"", "production", "staging", "development", "weird"} {
		if !slices.Equal(seedBundlesForEnv(env), seedbundle.BundleOrderForEnv(env)) {
			t.Fatalf("seedBundlesForEnv(%q) diverged from seedbundle.BundleOrderForEnv", env)
		}
	}
}
