package seedbundle

import (
	"slices"
	"testing"
)

func TestBundleOrderForEnv_ProductionExcludesDemoAndStaging(t *testing.T) {
	t.Parallel()

	for _, env := range []string{
		"production",
		"PRODUCTION",
		" prod ",
		"prod",
		"", // empty: fail-closed (no demo)
		"unknown",
		"preview",
		"ci",
	} {
		env := env
		t.Run("env="+env, func(t *testing.T) {
			t.Parallel()
			got := BundleOrderForEnv(env)
			if !slices.Equal(got, []string{"002_master"}) {
				t.Fatalf("BundleOrderForEnv(%q) = %v, want master only", env, got)
			}
			if slices.Contains(got, "003_demo") || slices.Contains(got, "004_staging") {
				t.Fatalf("production-like env %q must not include demo/staging: %v", env, got)
			}
		})
	}
}

// SEC-CS2-F01: APP_ENV=staging must not load privileged demo/staging CSV
// bundles (active system-admin accounts). Staging is master-only; full demo
// order is restricted to local development/test only.
func TestBundleOrderForEnv_StagingIsMasterOnly(t *testing.T) {
	t.Parallel()

	for _, env := range []string{"staging", "STAGING", " staging "} {
		env := env
		t.Run("env="+env, func(t *testing.T) {
			t.Parallel()
			got := BundleOrderForEnv(env)
			if !slices.Equal(got, []string{"002_master"}) {
				t.Fatalf("BundleOrderForEnv(%q) = %v, want master only", env, got)
			}
			if slices.Contains(got, "003_demo") || slices.Contains(got, "004_staging") {
				t.Fatalf("staging env %q must not include demo/staging bundles: %v", env, got)
			}
		})
	}
}

func TestBundleOrderForEnv_LocalDevAndTestIncludesFullOrder(t *testing.T) {
	t.Parallel()

	want := []string{"002_master", "003_demo", "004_staging"}
	for _, env := range []string{
		"development",
		"DEVELOPMENT",
		"local",
		"dev",
		"test",
	} {
		env := env
		t.Run("env="+env, func(t *testing.T) {
			t.Parallel()
			got := BundleOrderForEnv(env)
			if !slices.Equal(got, want) {
				t.Fatalf("BundleOrderForEnv(%q) = %v, want full order %v", env, got, want)
			}
		})
	}
}

func TestBundleOrderForEnv_DoesNotMutateBundleOrder(t *testing.T) {
	t.Parallel()

	original := slices.Clone(BundleOrder)
	got := BundleOrderForEnv("development")
	got[0] = "mutated"
	if !slices.Equal(BundleOrder, original) {
		t.Fatalf("BundleOrder mutated via BundleOrderForEnv return value: %v", BundleOrder)
	}
}

func TestBundleOrder_IsFullFKSafeOrder(t *testing.T) {
	t.Parallel()

	// BundleOrder remains the canonical full load order for non-env-aware
	// tooling (seed-export, docs). Env gating is BundleOrderForEnv only.
	want := []string{"002_master", "003_demo", "004_staging"}
	if !slices.Equal(BundleOrder, want) {
		t.Fatalf("BundleOrder = %v, want %v", BundleOrder, want)
	}
}
