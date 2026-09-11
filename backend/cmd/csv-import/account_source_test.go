package main

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/animal-ekarte/backend/internal/csvimport"
)

func TestAccountSourceFlagReachesPreflight(t *testing.T) {
	deps := testRunDependencies(t, testCLIBundle(), &fakeCutoverTarget{})
	stop := errors.New("stop after source binding")
	called := false
	deps.preflightBundle = func(_ string, expected csvimport.ExpectedCutoverSource) (csvimport.CutoverBundle, error) {
		called = true
		if expected.AccountSourceDir != "/migration-accounts" {
			t.Fatal("account source flag was lost")
		}
		return csvimport.CutoverBundle{}, stop
	}
	args := append(testCLIArgs("preflight"), "--account-source-dir", "/migration-accounts")
	if err := runWithDependencies(context.Background(), args, slog.Default(), deps); !errors.Is(err, stop) {
		t.Fatalf("source preflight error = %v", err)
	}
	if !called {
		t.Fatal("source preflight was not called")
	}
}
