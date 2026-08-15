package main

import (
	"encoding/csv"
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

	raw, err := readDemoAccountsCSV()
	if err != nil {
		t.Fatalf("read demo accounts.csv: %v", err)
	}
	// Parse CSV properly (avoid brittle ",t,t," substring on hash lines).
	r := csv.NewReader(strings.NewReader(string(raw)))
	rows, err := r.ReadAll()
	if err != nil || len(rows) < 2 {
		t.Fatalf("parse demo accounts.csv: rows=%d err=%v", len(rows), err)
	}
	header := rows[0]
	col := map[string]int{}
	for i, h := range header {
		col[strings.TrimSpace(h)] = i
	}
	ia, okA := col["is_active"]
	isa, okS := col["is_system_admin"]
	if !okA || !okS {
		t.Fatalf("demo accounts.csv missing is_active/is_system_admin columns: %v", header)
	}
	found := false
	for _, row := range rows[1:] {
		if len(row) <= ia || len(row) <= isa {
			continue
		}
		if row[ia] == "t" && row[isa] == "t" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("003_demo accounts.csv is expected to contain at least one active system admin (is_active=t, is_system_admin=t); demo-only intentional")
	}

	// Production and staging plans must still exclude the demo bundle that carries those accounts.
	for _, env := range []string{"production", "staging"} {
		if slices.Contains(seedBundlesForEnv(env), "003_demo") {
			t.Fatalf("%s seed plan must not load 003_demo despite demo CSV containing system admins", env)
		}
	}
}

func readDemoAccountsCSV() ([]byte, error) {
	candidates := []string{}
	if _, thisFile, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates,
			filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "seeds", "003_demo", "accounts.csv"),
		)
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "migrations", "seeds", "003_demo", "accounts.csv"),
			filepath.Join(wd, "backend", "migrations", "seeds", "003_demo", "accounts.csv"),
			filepath.Join(wd, "..", "migrations", "seeds", "003_demo", "accounts.csv"),
			filepath.Join(wd, "..", "..", "migrations", "seeds", "003_demo", "accounts.csv"),
		)
	}
	if ws := os.Getenv("GITHUB_WORKSPACE"); ws != "" {
		candidates = append(candidates,
			filepath.Join(ws, "backend", "migrations", "seeds", "003_demo", "accounts.csv"),
		)
	}
	var last error
	for _, p := range candidates {
		raw, err := os.ReadFile(p)
		if err == nil {
			return raw, nil
		}
		last = err
	}
	if last == nil {
		last = os.ErrNotExist
	}
	return nil, last
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
