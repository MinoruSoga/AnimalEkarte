package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestLogger returns a logger that discards output so these tests do not
// spam stdout with expected-error log lines.
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
}

// TestRunMissingRequiredEnvVars covers run()'s first safety gate: without the
// DB_HOST/USER/PASSWORD/NAME quartet it must fail fast, before ever attempting a
// database connection.
func TestRunMissingRequiredEnvVars(t *testing.T) {
	for _, unset := range []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME"} {
		t.Run("missing "+unset, func(t *testing.T) {
			t.Setenv("DB_HOST", "localhost")
			t.Setenv("DB_USER", "u")
			t.Setenv("DB_PASSWORD", "p")
			t.Setenv("DB_NAME", "d")
			t.Setenv(unset, "")

			err := run(newTestLogger())
			if err == nil {
				t.Fatalf("run() with %s unset must return an error", unset)
			}
			if !strings.Contains(err.Error(), "missing required DB env vars") {
				t.Fatalf("run() error = %q, want it to mention missing DB env vars", err.Error())
			}
		})
	}
}

// TestRunRejectsNonLocalHost is the core safety-boundary regression: this seeder
// must never be pointed at a non-local DB_HOST, however the caller configures it.
func TestRunRejectsNonLocalHost(t *testing.T) {
	t.Setenv("DB_HOST", "prod-db.example.com")
	t.Setenv("DB_USER", "u")
	t.Setenv("DB_PASSWORD", "p")
	t.Setenv("DB_NAME", "d")

	err := run(newTestLogger())
	if err == nil {
		t.Fatal("run() with a non-local DB_HOST must return an error")
	}
	if !strings.Contains(err.Error(), "SAFETY") || !strings.Contains(err.Error(), "not a known local host") {
		t.Fatalf("run() error = %q, want the local-host safety message", err.Error())
	}
}

// TestRunMissingOutputDirReturnsError covers the OLD_DB_MIGRATION_OUTPUT_DIR
// existence guard, reached only after the env-var and host checks pass.
func TestRunMissingOutputDirReturnsError(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_USER", "u")
	t.Setenv("DB_PASSWORD", "p")
	t.Setenv("DB_NAME", "d")
	t.Setenv("OLD_DB_MIGRATION_OUTPUT_DIR", filepath.Join(t.TempDir(), "does-not-exist"))

	err := run(newTestLogger())
	if err == nil {
		t.Fatal("run() with a missing OLD_DB_MIGRATION_OUTPUT_DIR must return an error")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("run() error = %q, want it to mention the missing output dir", err.Error())
	}
}

// TestRunManifestLoadFailureReturnsError covers the loadManifest error-wrapping
// branch: outputDir exists, but the manifest path is missing/invalid.
func TestRunManifestLoadFailureReturnsError(t *testing.T) {
	outputDir := t.TempDir()
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_USER", "u")
	t.Setenv("DB_PASSWORD", "p")
	t.Setenv("DB_NAME", "d")
	t.Setenv("OLD_DB_MIGRATION_OUTPUT_DIR", outputDir)
	t.Setenv("OLD_DB_MANIFEST_PATH", filepath.Join(outputDir, "missing-manifest.json"))

	err := run(newTestLogger())
	if err == nil {
		t.Fatal("run() with a missing manifest file must return an error")
	}
	if !strings.Contains(err.Error(), "failed to load manifest") {
		t.Fatalf("run() error = %q, want it to mention the manifest load failure", err.Error())
	}
}
