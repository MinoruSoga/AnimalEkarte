package main

import (
	"flag"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
}

// resetFlags gives run() (via parseFlags) a clean global flag set for the
// duration of the test, restoring os.Args/flag.CommandLine afterward.
func resetFlags(t *testing.T, args []string) {
	t.Helper()
	oldArgs := os.Args
	oldCmdLine := flag.CommandLine
	t.Cleanup(func() {
		os.Args = oldArgs
		flag.CommandLine = oldCmdLine
	})
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = args
}

// TestRunMissingTargetEnvReturnsErrorBeforeAnyDBConnection covers run()'s
// targetDSN() call, which must fail fast — before opening any pgx pool — when
// required target env vars are absent.
func TestRunMissingTargetEnvReturnsErrorBeforeAnyDBConnection(t *testing.T) {
	resetFlags(t, []string{"stage-import"})
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")

	err := run(newTestLogger())
	if err == nil {
		t.Fatal("run() with missing target env must return an error")
	}
	if !strings.Contains(err.Error(), "missing required target env") {
		t.Fatalf("run() error = %q, want it to mention missing target env", err.Error())
	}
}

// TestRunRejectsNonLocalTargetBeforeAnyDBConnection is the core safety-boundary
// regression: run() must refuse a non-local TARGET host before connecting.
func TestRunRejectsNonLocalTargetBeforeAnyDBConnection(t *testing.T) {
	resetFlags(t, []string{"stage-import"})
	t.Setenv("DB_HOST", "prod-db.example.com")
	t.Setenv("DB_USER", "u")
	t.Setenv("DB_PASSWORD", "p")
	t.Setenv("DB_NAME", "d")

	err := run(newTestLogger())
	if err == nil {
		t.Fatal("run() with a non-local target host must return an error")
	}
	if !strings.Contains(err.Error(), "SAFETY") {
		t.Fatalf("run() error = %q, want the local-host safety message", err.Error())
	}
}
