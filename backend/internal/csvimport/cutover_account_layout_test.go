package csvimport

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCutoverAccountLayout(t *testing.T) {
	dir, digest := writeCutoverFixture(t, nil)
	accountDir := filepath.Join(dir, "accounts")
	if err := os.Mkdir(accountDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(dir, "staffs.csv"), filepath.Join(accountDir, "staffs.csv")); err != nil {
		t.Fatal(err)
	}
	_, err := PreflightCutoverBundle(dir, stagingExpectedCutoverSource(digest))
	if err != nil {
		t.Fatal(err)
	}
}

func TestCutoverCentralAccountLayout(t *testing.T) {
	dir, digest := writeCutoverFixture(t, nil)
	accountDir := filepath.Join(t.TempDir(), "002_master", "accounts", "_old_db_handoff", "fixture")
	if err := os.MkdirAll(accountDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(dir, "staffs.csv"), filepath.Join(accountDir, "staffs.csv")); err != nil {
		t.Fatal(err)
	}
	expected := stagingExpectedCutoverSource(digest)
	expected.AccountSourceDir = accountDir
	if _, err := PreflightCutoverBundle(dir, expected); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(accountDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := PreflightCutoverBundle(dir, expected); err == nil {
		t.Fatal("public account directory accepted")
	}
}

func TestCutoverCentralAccountAutoResolution(t *testing.T) {
	dir, digest := writeCutoverFixture(t, nil)
	root := t.TempDir()
	sourceDir := filepath.Join(root, "_old_db_handoff", "fixture")
	if err := os.MkdirAll(filepath.Dir(sourceDir), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(dir, sourceDir); err != nil {
		t.Fatal(err)
	}
	accountDir := filepath.Join(root, "002_master", "accounts", "_old_db_handoff", "fixture")
	if err := os.MkdirAll(accountDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(sourceDir, "staffs.csv"), filepath.Join(accountDir, "staffs.csv")); err != nil {
		t.Fatal(err)
	}
	if _, err := PreflightCutoverBundle(sourceDir, stagingExpectedCutoverSource(digest)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountDir, "staffs.csv"), []byte("id\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := PreflightCutoverBundle(sourceDir, stagingExpectedCutoverSource(digest)); err == nil {
		t.Fatal("changed staff CSV accepted")
	}
}

func TestCutoverAccountLayoutRejectsUnsafeDirectories(t *testing.T) {
	for _, scenario := range []string{"symlink", "duplicate", "extra", "permissions", "missing"} {
		t.Run(scenario, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, nil)
			accountDir := filepath.Join(dir, "accounts")
			if scenario == "symlink" {
				if err := os.Symlink(t.TempDir(), accountDir); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Mkdir(accountDir, 0o700); err != nil {
					t.Fatal(err)
				}
				if scenario != "duplicate" && scenario != "missing" {
					if err := os.Rename(filepath.Join(dir, "staffs.csv"), filepath.Join(accountDir, "staffs.csv")); err != nil {
						t.Fatal(err)
					}
				}
				if scenario == "extra" {
					if err := os.WriteFile(filepath.Join(accountDir, "unexpected.csv"), nil, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if scenario == "permissions" {
					if err := os.Chmod(accountDir, 0o755); err != nil {
						t.Fatal(err)
					}
				}
				if scenario == "missing" {
					if err := os.Remove(filepath.Join(dir, "staffs.csv")); err != nil {
						t.Fatal(err)
					}
				}
			}
			if _, err := PreflightCutoverBundle(dir, stagingExpectedCutoverSource(digest)); err == nil {
				t.Fatal("unsafe layout accepted")
			}
		})
	}
}
