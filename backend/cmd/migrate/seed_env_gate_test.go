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

func TestSeedBundlesForEnv_LocalDevAndTestIsMasterOnly(t *testing.T) {
	t.Parallel()

	for _, env := range []string{"development", "local", "test", "dev"} {
		got := seedBundlesForEnv(env)
		want := []string{"002_master"}
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
	want := []string{"002_master"}
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

func TestMasterBundleHasClinicSkeletonAndNoAccounts(t *testing.T) {
	t.Parallel()

	root := masterSeedDir(t)
	if _, err := os.Stat(filepath.Join(root, "accounts.csv")); !os.IsNotExist(err) {
		t.Fatalf("002_master must not contain accounts.csv (got err=%v)", err)
	}
	clinics, err := os.ReadFile(filepath.Join(root, "clinics.csv"))
	if err != nil {
		t.Fatalf("read clinics.csv: %v", err)
	}
	text := string(clinics)
	if !strings.Contains(text, "八王子病院") || !strings.Contains(text, "城東センター病院") {
		t.Fatal("002_master clinics.csv must name 八王子病院 and 城東センター病院")
	}
	if !strings.Contains(text, "ノア動物病院　敷島病院") || !strings.Contains(text, "ノア動物病院　Hako bu neco") {
		t.Fatal("002_master clinics.csv must name Jouto-group clinics 3 and 4")
	}
	if strings.Contains(text, "敷島医院") {
		t.Fatal("002_master clinics.csv must not include demo-only 敷島医院")
	}
	f, err := os.Open(filepath.Join(root, "exam_types.csv"))
	if err != nil {
		t.Fatalf("open exam_types.csv: %v", err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil || len(rows) < 2 {
		t.Fatalf("parse exam_types.csv: rows=%d err=%v", len(rows), err)
	}
	nameIdx := -1
	clinicIdx := -1
	for i, h := range rows[0] {
		switch h {
		case "name":
			nameIdx = i
		case "clinic_id":
			clinicIdx = i
		}
	}
	if nameIdx < 0 || clinicIdx < 0 {
		t.Fatalf("exam_types.csv missing name/clinic_id: %v", rows[0])
	}
	kensaClinics := map[string]bool{}
	for _, row := range rows[1:] {
		if row[nameIdx] == "検査" {
			kensaClinics[row[clinicIdx]] = true
		}
	}
	if !kensaClinics["1"] || !kensaClinics["2"] || !kensaClinics["3"] || !kensaClinics["4"] {
		t.Fatalf("002_master exam_types 検査 clinics = %v, want 1-4", kensaClinics)
	}
}

func masterSeedDir(t *testing.T) string {
	t.Helper()
	if _, thisFile, _, ok := runtime.Caller(0); ok {
		p := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "seeds", "002_master")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if wd, err := os.Getwd(); err == nil {
		for _, p := range []string{
			filepath.Join(wd, "migrations", "seeds", "002_master"),
			filepath.Join(wd, "backend", "migrations", "seeds", "002_master"),
		} {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	t.Fatal("002_master seed dir not found")
	return ""
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
