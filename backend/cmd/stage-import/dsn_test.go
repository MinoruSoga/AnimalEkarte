package main

import (
	"strings"
	"testing"
)

func TestConnStringIncludesAllFields(t *testing.T) {
	d := dsn{host: "db", port: "5432", user: "u", password: "p", name: "n", sslMode: "disable"}
	got := d.connString()
	for _, want := range []string{"host=db", "port=5432", "user=u", "password=p", "dbname=n", "sslmode=disable"} {
		if !strings.Contains(got, want) {
			t.Fatalf("connString() = %q, want it to contain %q", got, want)
		}
	}
}

func TestConnStringReadOnlyForcesReadOnlyTransaction(t *testing.T) {
	d := dsn{host: "db", port: "5432", user: "u", password: "p", name: "n", sslMode: "disable"}
	got := d.connStringReadOnly()
	if !strings.HasPrefix(got, d.connString()) {
		t.Fatalf("connStringReadOnly() must extend connString(), got %q", got)
	}
	if !strings.Contains(got, "default_transaction_read_only=on") {
		t.Fatalf("connStringReadOnly() must set default_transaction_read_only=on, got %q", got)
	}
}

func TestEnvOrReturnsEnvValueWhenSet(t *testing.T) {
	t.Setenv("STAGE_IMPORT_TEST_KEY", "custom")
	if got := envOr("STAGE_IMPORT_TEST_KEY", "fallback"); got != "custom" {
		t.Fatalf("envOr = %q, want %q", got, "custom")
	}
}

func TestEnvOrReturnsFallbackWhenUnset(t *testing.T) {
	t.Setenv("STAGE_IMPORT_TEST_KEY_UNSET", "")
	if got := envOr("STAGE_IMPORT_TEST_KEY_UNSET", "fallback"); got != "fallback" {
		t.Fatalf("envOr = %q, want %q", got, "fallback")
	}
}

func TestTargetDSNMissingEnvReturnsError(t *testing.T) {
	for _, unset := range []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME"} {
		t.Run("missing "+unset, func(t *testing.T) {
			t.Setenv("DB_HOST", "localhost")
			t.Setenv("DB_USER", "u")
			t.Setenv("DB_PASSWORD", "p")
			t.Setenv("DB_NAME", "d")
			t.Setenv(unset, "")

			if _, err := targetDSN(); err == nil {
				t.Fatalf("targetDSN() with %s unset must return an error", unset)
			}
		})
	}
}

func TestTargetDSNRejectsNonLocalHost(t *testing.T) {
	t.Setenv("DB_HOST", "prod-db.example.com")
	t.Setenv("DB_USER", "u")
	t.Setenv("DB_PASSWORD", "p")
	t.Setenv("DB_NAME", "d")

	if _, err := targetDSN(); err == nil {
		t.Fatal("targetDSN() with a non-local host must return an error")
	}
}

func TestTargetDSNAcceptsLocalHostAndAppliesDefaults(t *testing.T) {
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_USER", "u")
	t.Setenv("DB_PASSWORD", "p")
	t.Setenv("DB_NAME", "d")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_SSL_MODE", "")

	got, err := targetDSN()
	if err != nil {
		t.Fatalf("targetDSN() unexpected error: %v", err)
	}
	if got.port != "5432" {
		t.Errorf("targetDSN().port = %q, want default 5432", got.port)
	}
	if got.sslMode != "disable" {
		t.Errorf("targetDSN().sslMode = %q, want default disable", got.sslMode)
	}
}

func TestStageDSNAppliesDefaults(t *testing.T) {
	for _, key := range []string{"STAGE_DB_HOST", "STAGE_DB_PORT", "STAGE_DB_USER", "STAGE_DB_PASSWORD", "STAGE_DB_NAME", "STAGE_DB_SSL_MODE"} {
		t.Setenv(key, "")
	}

	got := stageDSN()
	if got.host != "old-db-postgres" {
		t.Errorf("stageDSN().host = %q, want default old-db-postgres", got.host)
	}
	if got.port != "5432" {
		t.Errorf("stageDSN().port = %q, want default 5432", got.port)
	}
	if got.user != "postgres" {
		t.Errorf("stageDSN().user = %q, want default postgres", got.user)
	}
	if got.name != "ani_legacy" {
		t.Errorf("stageDSN().name = %q, want default ani_legacy", got.name)
	}
	if got.sslMode != "disable" {
		t.Errorf("stageDSN().sslMode = %q, want default disable", got.sslMode)
	}
}

func TestStageDSNHonoursEnvOverrides(t *testing.T) {
	t.Setenv("STAGE_DB_HOST", "custom-host")
	t.Setenv("STAGE_DB_PORT", "5433")
	t.Setenv("STAGE_DB_USER", "custom-user")
	t.Setenv("STAGE_DB_NAME", "custom-db")

	got := stageDSN()
	if got.host != "custom-host" || got.port != "5433" || got.user != "custom-user" || got.name != "custom-db" {
		t.Fatalf("stageDSN() = %+v, want overrides applied", got)
	}
}

// TestParseFlagsDefaults uses the shared resetFlags helper (run_test.go) to give
// parseFlags a clean global flag.CommandLine so this test does not depend on run
// order or interfere with other tests in the package.
func TestParseFlagsDefaults(t *testing.T) {
	resetFlags(t, []string{"stage-import"})

	got := parseFlags()
	if got.apply || got.confirmLocal || got.runID != "" {
		t.Fatalf("parseFlags() defaults = %+v, want all zero-value (safe dry-run)", got)
	}
}

func TestParseFlagsParsesApplyAndConfirm(t *testing.T) {
	resetFlags(t, []string{"stage-import", "-apply", "-confirm-local-destroy", "-run-id", "run-42"})

	got := parseFlags()
	if !got.apply || !got.confirmLocal || got.runID != "run-42" {
		t.Fatalf("parseFlags() = %+v, want apply=true confirmLocal=true runID=run-42", got)
	}
}
