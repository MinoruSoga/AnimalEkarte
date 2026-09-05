package main

import (
	"slices"
	"testing"

	"github.com/animal-ekarte/backend/internal/seedlogin"
)

func TestExpectedSeedBundleDirs_IncludesLoginForLocalAndStaging(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	got := expectedSeedBundleDirs()
	want := []string{"002_master", seedlogin.BundleDir}
	if !slices.Equal(got, want) {
		t.Fatalf("development plan = %v, want %v", got, want)
	}

	t.Setenv("APP_ENV", "staging")
	got = expectedSeedBundleDirs()
	if !slices.Equal(got, want) {
		t.Fatalf("staging plan = %v, want %v", got, want)
	}
}

func TestExpectedSeedBundleDirs_ProductionOmitsLogin(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	got := expectedSeedBundleDirs()
	if !slices.Equal(got, []string{"002_master"}) {
		t.Fatalf("production plan = %v, want master only", got)
	}
}

func TestSeedBundlesForEnv_StillExcludesCSVAccounts(t *testing.T) {
	t.Parallel()
	got := seedBundlesForEnv("staging")
	if slices.Contains(got, seedlogin.BundleDir) {
		t.Fatalf("CSV plan must not include %s: %v", seedlogin.BundleDir, got)
	}
}
