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

// Demo/staging CSV bundles were retired. Local and staging both load 002_master
// only (clinic skeleton + reference masters, no demo accounts or clinical rows).
func TestBundleOrderForEnv_LocalDevAndTestIsMasterOnly(t *testing.T) {
	t.Parallel()

	want := []string{"002_master"}
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
				t.Fatalf("BundleOrderForEnv(%q) = %v, want master only %v", env, got, want)
			}
			if slices.Contains(got, "003_demo") || slices.Contains(got, "004_staging") {
				t.Fatalf("env %q must not include retired demo/staging bundles: %v", env, got)
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
	want := []string{"002_master"}
	if !slices.Equal(BundleOrder, want) {
		t.Fatalf("BundleOrder = %v, want %v", BundleOrder, want)
	}
}
