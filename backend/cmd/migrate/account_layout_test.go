package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBundleChecksumPreservedByAccountMove(t *testing.T) {
	dir := t.TempDir()
	bundle := filepath.Join(dir, "seeds", "002_master")
	if err := os.MkdirAll(bundle, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := `{"bundle":"002_master","tables":[{"table":"permission_groups","csvFile":"permission_groups.csv"}]}`
	if err := os.WriteFile(filepath.Join(bundle, "manifest.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundle, "permission_groups.csv"), []byte("id,name\n1,test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := bundleChecksum(dir, "002_master")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(bundle, "accounts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(bundle, "permission_groups.csv"), filepath.Join(bundle, "accounts", "permission_groups.csv")); err != nil {
		t.Fatal(err)
	}
	after, err := bundleChecksum(dir, "002_master")
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("layout change altered checksum")
	}
}
