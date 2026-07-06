package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeBundleFixture(t *testing.T, migrationsDir, bundleDir, csvContent string) {
	t.Helper()
	dir := filepath.Join(migrationsDir, "seeds", bundleDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create fixture dir: %v", err)
	}
	manifest := `{"bundle":"` + bundleDir + `","tables":[{"table":"widgets","csvFile":"widgets.csv"}]}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("failed to write manifest fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "widgets.csv"), []byte(csvContent), 0o644); err != nil {
		t.Fatalf("failed to write csv fixture: %v", err)
	}
}

func TestBundleChecksumChangesWhenCSVContentChanges(t *testing.T) {
	dir := t.TempDir()

	writeBundleFixture(t, dir, "002_master", "id,name\n1,widget-a\n")
	sum1, err := bundleChecksum(dir, "002_master")
	if err != nil {
		t.Fatalf("bundleChecksum (first) returned error: %v", err)
	}

	// Same manifest, but the CSV data changed — checksum must change too, since
	// the CSV is the actual seed content (there is no stub SQL body anymore).
	writeBundleFixture(t, dir, "002_master", "id,name\n1,widget-b\n")
	sum2, err := bundleChecksum(dir, "002_master")
	if err != nil {
		t.Fatalf("bundleChecksum (second) returned error: %v", err)
	}

	if sum1 == sum2 {
		t.Fatalf("bundleChecksum did not change after CSV content changed (both = %q) — a CSV-only edit would be silently skipped on an already-migrated DB", sum1)
	}
}

func TestBundleChecksumChangesWhenManifestChanges(t *testing.T) {
	dir := t.TempDir()
	seedDir := filepath.Join(dir, "seeds", "003_demo")
	if err := os.MkdirAll(seedDir, 0o755); err != nil {
		t.Fatalf("failed to create fixture dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(seedDir, "a.csv"), []byte("id,name\n1,a\n"), 0o644); err != nil {
		t.Fatalf("failed to write csv fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(seedDir, "b.csv"), []byte("id,name\n1,b\n"), 0o644); err != nil {
		t.Fatalf("failed to write csv fixture: %v", err)
	}

	manifestForward := `{"bundle":"003_demo","tables":[{"table":"a","csvFile":"a.csv"},{"table":"b","csvFile":"b.csv"}]}`
	if err := os.WriteFile(filepath.Join(seedDir, "manifest.json"), []byte(manifestForward), 0o644); err != nil {
		t.Fatalf("failed to write manifest fixture: %v", err)
	}
	sum1, err := bundleChecksum(dir, "003_demo")
	if err != nil {
		t.Fatalf("bundleChecksum (forward order) returned error: %v", err)
	}

	// Same CSV files, reordered manifest — the checksum must still change,
	// since manifest.json's own bytes are folded into the hash.
	manifestReordered := `{"bundle":"003_demo","tables":[{"table":"b","csvFile":"b.csv"},{"table":"a","csvFile":"a.csv"}]}`
	if err := os.WriteFile(filepath.Join(seedDir, "manifest.json"), []byte(manifestReordered), 0o644); err != nil {
		t.Fatalf("failed to write reordered manifest fixture: %v", err)
	}
	sum2, err := bundleChecksum(dir, "003_demo")
	if err != nil {
		t.Fatalf("bundleChecksum (reordered) returned error: %v", err)
	}

	if sum1 == sum2 {
		t.Fatalf("bundleChecksum did not change after manifest.json table order changed (both = %q)", sum1)
	}
}

func TestBundleChecksumStableWhenNothingChanges(t *testing.T) {
	dir := t.TempDir()
	writeBundleFixture(t, dir, "003_demo", "id,name\n1,widget-a\n")

	sum1, err := bundleChecksum(dir, "003_demo")
	if err != nil {
		t.Fatalf("bundleChecksum (first) returned error: %v", err)
	}
	sum2, err := bundleChecksum(dir, "003_demo")
	if err != nil {
		t.Fatalf("bundleChecksum (second) returned error: %v", err)
	}
	if sum1 != sum2 {
		t.Fatalf("bundleChecksum is not deterministic for identical inputs: %q vs %q", sum1, sum2)
	}
}

func TestBundleChecksumMissingManifestErrors(t *testing.T) {
	dir := t.TempDir() // no seeds/ subdirectory created
	_, err := bundleChecksum(dir, "004_staging")
	if err == nil {
		t.Fatal("expected an error when a bundle's manifest.json is missing, got nil")
	}
}
