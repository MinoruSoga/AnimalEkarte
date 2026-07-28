package lintscan

// lintscan_test.go — BE9-1: RED-first self-test for the shared lint-harness discovery core
// (lintscan.go). Three groups, matching the three exported functions:
//
//  1. TestWalkInternalTree_* — pure, synthetic-fixture tests built under t.TempDir(). These
//     prove WalkInternalTree's inclusion/exclusion rules (root-level file, 1-level and 2+-level
//     nested subpackages, an "aliased package" directory, testdata/vendor exclusion, _test.go
//     exclusion, non-.go exclusion) without touching the real repo tree or any *testing.T
//     dependency inside the production function itself.
//  2. TestFindModuleRoot_* — synthetic-fixture tests proving the upward go.mod search resolves
//     from an arbitrarily deep nested directory back to the go.mod-owning directory, and returns
//     a clear error (not a panic) when no go.mod exists anywhere upward.
//  3. TestWalkInternalTreeT_RealRepo — the one integration test that runs WalkInternalTreeT
//     against the ACTUAL repository under a real `go test` invocation. This is the executable
//     proof (not a re-derived argument) that the os.Getwd()-based upward search is safe under
//     this project's real CI invocation and under -trimpath, mirroring
//     internal/repository/preload_master_model_reconciliation_test.go's siblingPackageDir /
//     loadModelClinicScopedStructs, which already relies on the same cwd guarantee and is green
//     in this project's real CI today.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFixtureFile creates parentDir (and any missing ancestors) and writes name with contents.
func writeFixtureFile(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s/%s: %v", dir, name, err)
	}
}

// buildWalkFixture lays out a synthetic internal/ tree under root covering every
// inclusion/exclusion case WalkInternalTree must handle, and returns the exact set of relative
// (forward-slash) keys that must survive the walk.
func buildWalkFixture(t *testing.T, root string) map[string]struct{} {
	t.Helper()

	// Root-level production file, directly under internalRoot.
	writeFixtureFile(t, root, "root.go", "package lintscanfixture\n\nvar Root = 1\n")

	// 1-level subpackage.
	writeFixtureFile(t, filepath.Join(root, "sub"), "file.go", "package sub\n\nvar X = 1\n")

	// 2-or-more-level nested subpackage.
	writeFixtureFile(t, filepath.Join(root, "sub", "nested", "deeper"), "file2.go",
		"package deeper\n\nvar Y = 1\n")

	// "Aliased package" directory: dir name "weirdname", declared package name "actual".
	// WalkInternalTree never reads the "package X" clause — this fixture proves the file is
	// still discovered by location alone, regardless of the mismatch.
	writeFixtureFile(t, filepath.Join(root, "sub", "weirdname"), "file3.go",
		"package actual\n\nvar Z = 1\n")

	// Content-classification is explicitly OUT of WalkInternalTree's job (see lintscan.go
	// doc comment). These two fixtures prove the walker does not content-sniff: a raw SQL
	// string literal and a bare unregistered-name uint64 scalar field are both discovered
	// exactly like any other production file. Whether any given LINT correctly classifies
	// this content is that lint's own judgment-layer concern, not proven here.
	writeFixtureFile(t, filepath.Join(root, "sqlfile"), "rawsql.go",
		"package sqlfile\n\nconst Query = \"SELECT * FROM owners WHERE id = \" + idParam\n")
	writeFixtureFile(t, filepath.Join(root, "scalarfile"), "scalar.go",
		"package scalarfile\n\ntype Widget struct {\n\tMysteryFKID uint64\n}\n")

	// Exclusions.
	writeFixtureFile(t, filepath.Join(root, "testdata"), "file4.go",
		"package testdata\n\nvar Excluded = 1\n")
	writeFixtureFile(t, root, "root_test.go",
		"package lintscanfixture\n\nimport \"testing\"\n\nfunc TestExcluded(t *testing.T) {}\n")
	writeFixtureFile(t, filepath.Join(root, "vendor", "thirdparty"), "vendored.go",
		"package thirdparty\n\nvar Excluded = 1\n")
	writeFixtureFile(t, root, "README.md", "not a go file\n")

	return map[string]struct{}{
		"root.go":                    {},
		"sub/file.go":                {},
		"sub/nested/deeper/file2.go": {},
		"sub/weirdname/file3.go":     {},
		"sqlfile/rawsql.go":          {},
		"scalarfile/scalar.go":       {},
	}
}

func TestWalkInternalTree_ExactExpectedKeySet(t *testing.T) {
	root := t.TempDir()
	want := buildWalkFixture(t, root)

	got, err := WalkInternalTree(root)
	if err != nil {
		t.Fatalf("WalkInternalTree: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("got %d keys, want %d\ngot:  %v\nwant: %v", len(got), len(want), keysOf(got), want)
	}
	for k := range want {
		if _, ok := got[k]; !ok {
			t.Errorf("expected key %q missing from result (got: %v)", k, keysOf(got))
		}
	}
	for k := range got {
		if _, ok := want[k]; !ok {
			t.Errorf("unexpected key %q present in result (must be excluded)", k)
		}
	}
}

func TestWalkInternalTree_ReturnsFileContents(t *testing.T) {
	root := t.TempDir()
	buildWalkFixture(t, root)

	got, err := WalkInternalTree(root)
	if err != nil {
		t.Fatalf("WalkInternalTree: %v", err)
	}
	src, ok := got["root.go"]
	if !ok {
		t.Fatal("expected root.go in result")
	}
	if string(src) != "package lintscanfixture\n\nvar Root = 1\n" {
		t.Errorf("root.go contents mismatch: %q", string(src))
	}
}

func TestWalkInternalTree_EmptyTreeReturnsEmptyMapNotError(t *testing.T) {
	root := t.TempDir()
	got, err := WalkInternalTree(root)
	if err != nil {
		t.Fatalf("WalkInternalTree on empty dir: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", keysOf(got))
	}
}

func TestWalkInternalTree_NonexistentRootReturnsError(t *testing.T) {
	root := filepath.Join(t.TempDir(), "does-not-exist")
	if _, err := WalkInternalTree(root); err == nil {
		t.Error("expected an error for a nonexistent internalRoot, got nil")
	}
}

func keysOf(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestFindModuleRoot_ResolvesFromDeeplyNestedDir(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "go.mod", "module example.com/fixture\n\ngo 1.25.0\n")

	nested := filepath.Join(root, "internal", "lintscan", "sub", "deeper", "evendeeper")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", nested, err)
	}

	got, err := FindModuleRoot(nested)
	if err != nil {
		t.Fatalf("FindModuleRoot(%s): %v", nested, err)
	}

	wantAbs, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("resolve want: %v", err)
	}
	gotAbs, err := filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatalf("resolve got: %v", err)
	}
	if gotAbs != wantAbs {
		t.Errorf("FindModuleRoot = %q, want %q", got, root)
	}
}

func TestFindModuleRoot_ResolvesFromModuleRootItself(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "go.mod", "module example.com/fixture\n\ngo 1.25.0\n")

	got, err := FindModuleRoot(root)
	if err != nil {
		t.Fatalf("FindModuleRoot(%s): %v", root, err)
	}
	gotAbs, _ := filepath.EvalSymlinks(got)
	wantAbs, _ := filepath.EvalSymlinks(root)
	if gotAbs != wantAbs {
		t.Errorf("FindModuleRoot = %q, want %q", got, root)
	}
}

func TestFindModuleRoot_NoGoModAnywhereUpwardReturnsClearError(t *testing.T) {
	// A tree with no go.mod at all: FindModuleRoot must walk all the way to the filesystem
	// root and return a descriptive error, never panic.
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", nested, err)
	}

	_, err := FindModuleRoot(nested)
	if err == nil {
		t.Fatal("expected an error when no go.mod exists anywhere upward, got nil")
	}
}

// TestWalkInternalTreeT_RealRepo is the one real-repo integration test: it runs
// WalkInternalTreeT against the ACTUAL repository under this project's real `go test`
// invocation, proving the os.Getwd()-based upward search (see FindModuleRoot's doc comment,
// which cites internal/repository/preload_master_model_reconciliation_test.go directly) is
// -trimpath-safe in practice, not just in theory.
func TestWalkInternalTreeT_RealRepo(t *testing.T) {
	got := WalkInternalTreeT(t)

	if len(got) == 0 {
		t.Fatal("expected a non-empty file set from the real internal/ tree")
	}

	if _, ok := got["lintscan/lintscan.go"]; !ok {
		t.Errorf(`expected "lintscan/lintscan.go" in result (this very package's production file); got %d keys`, len(got))
	}
	if _, ok := got["infra/crypto/aes_gcm.go"]; !ok {
		t.Errorf(`expected "infra/crypto/aes_gcm.go" in result (a known nested production file); got %d keys`, len(got))
	}

	for k := range got {
		if filepath.Ext(k) != ".go" {
			t.Errorf("unexpected non-.go key %q", k)
		}
		if strings.HasSuffix(k, "_test.go") {
			t.Errorf("unexpected _test.go key %q (must be excluded)", k)
		}
		if strings.Contains(k, "/testdata/") || strings.HasPrefix(k, "testdata/") {
			t.Errorf("unexpected /testdata/ key %q (must be excluded)", k)
		}
	}
}
