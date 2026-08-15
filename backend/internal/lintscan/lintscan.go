// Package lintscan provides the shared, package-structure-independent discovery core for this
// backend's safety lints (clinic-scope Preload, audit-tx inventory, DB-or-tx inventory,
// master-FK write inventory, N+1, and future lints under BE9-1's package-restructuring work).
//
// BE9-0 lesson (see internal/testdb/testdb.go for the current shared test-kernel precedent of
// this exact move): helpers defined inside a "_test.go" file cannot be imported from another
// package's "_test.go" files — the Go toolchain excludes "_test.go" files from a package's
// normal build graph, so only files ending in a plain ".go" extension are importable across
// package boundaries. Several existing lints (internal/repository's
// preload_clinic_scope_lint_test.go / audit_tx_inventory_lint_test.go / dbortx_inventory_lint_test.go
// and internal/service's master_fk_write_inventory_lint_test.go / n1_lint_test.go) each
// currently re-implement their own "walk the internal/ tree and find production .go files"
// discovery step, using two different techniques (go:embed with a hardcoded one-level
// "*/*.go" pattern in some, os.Getwd()+".." sibling-package traversal in others — see
// preload_master_model_reconciliation_test.go's file header for that second technique's
// rationale). Both techniques are correct but ad hoc. This package factors the discovery step
// out once so it is importable from BOTH internal/repository/*_test.go and
// internal/service/*_test.go (and, as BE9-1's target packages come online, from every other
// internal/<domain>/*_test.go package) without re-deriving the walk logic or its trimpath
// safety argument in each call site — which is exactly why lintscan.go itself must NOT be a
// "_test.go" file, even though every one of its current callers is a test file.
//
// Explicit division of responsibility (do not blur this boundary): this package is a pure
// DISCOVERY layer — "which production .go files exist, and what are their bytes". It has no
// opinion on package boundaries, dependency direction, cyclic imports, or any future
// per-package split BE9-1 might produce; walking arbitrarily deep is what makes it agnostic to
// however many subpackage levels exist today or are added later (see WalkInternalTree's own
// comment for why a fixed-depth glob is deliberately not used here). Classifying what a
// discovered file's CONTENT means (e.g. is this raw-SQL string literal a violation of the
// tenant-scope rule, does this bare scalar field represent an unregistered clinic-scoped
// master FK) is each individual lint's own judgment-layer job, layered on top of the byte
// slices this package returns — lintscan.go intentionally provides no content-level guarantee
// of any kind, and never inspects a file's own "package X" clause, which is exactly why a
// directory whose declared package name differs from its directory name ("aliased package")
// is fully transparent to WalkInternalTree: discovery here is by filesystem location only.
package lintscan

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FindModuleRoot walks upward from startDir — via its parent directory, repeatedly — looking
// for a "go.mod" file, and returns the directory containing it.
//
// This deliberately does NOT hard-code a fixed number of ".." hops: it is called today from
// internal/repository (2 levels below backend/), internal/service (2 levels below backend/),
// and internal/lintscan itself (2 levels below backend/), and BE9-1's future target packages
// (internal/<domain>/, internal/<domain>/<subpackage>/, ...) will call it from a range of
// nesting depths that are not fixed in advance. An upward search is the only approach that
// stays correct for all current AND future callers regardless of nesting depth — see the
// package doc comment's "package-structure-independent" framing, which BE9-1 requires of this
// discovery layer.
//
// Precedent for why searching from os.Getwd() (the usual caller, via WalkInternalTreeT) is
// trimpath-safe: internal/repository/preload_master_model_reconciliation_test.go's
// siblingPackageDir already relies on os.Getwd() returning the tested package's own source
// directory under `go test` — a long-standing, widely relied-upon Go test runner guarantee
// that every package's test binary is invoked with cwd = that package's own directory — and is
// green in this project's real CI today. That guarantee is about the OS working directory, not
// about embedded/debug build paths, so it has no dependency on -trimpath. This function cites
// that file directly rather than re-deriving the argument, per this task's precedent that
// executable, currently-green precedent outranks a fresh re-derivation.
func FindModuleRoot(startDir string) (string, error) {
	dir := startDir
	for {
		candidate := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(candidate); err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("lintscan: stat %s: %w", candidate, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the filesystem root without finding go.mod anywhere upward.
			return "", fmt.Errorf("lintscan: no go.mod found walking upward from %s", startDir)
		}
		dir = parent
	}
}

// WalkInternalTree is a PURE function (no *testing.T dependency) that recursively walks
// internalRoot to arbitrary depth via filepath.WalkDir and returns every production .go file it
// finds, keyed by its path relative to internalRoot using forward slashes (e.g.
// "repository/paymentmethod/repository.go", "service/vaccination_service.go",
// "lintscan/lintscan.go").
//
// Excluded:
//   - any file whose name ends in "_test.go"
//   - any file inside a directory literally named "testdata", at any depth
//   - any directory literally named "vendor", at any depth (and everything under it)
//   - any file whose name does not end in ".go"
//
// This function does NOT inspect file contents, and does NOT parse or read a file's own
// "package X" clause at all — discovery here is purely by filesystem location. That is exactly
// what makes a package whose declared package name differs from its containing directory name
// (an "aliased package") transparent to this function: it is still found, keyed by its real
// path, regardless of what it calls itself internally.
//
// Separately (documented here, not implemented — this is a blind spot BE9-1 requires stated
// explicitly, not solved): classifying a discovered file's CONTENT — e.g. distinguishing a raw
// SQL string literal that legitimately concatenates a clinic_id predicate from one that does
// not, or recognizing a bare scalar field as an unregistered clinic-scoped master FK versus an
// ordinary non-tenant-scoped ID — is each individual lint's own judgment-layer job, built on
// top of the byte slices this function returns. WalkInternalTree provides no content-level
// guarantee of any kind; it only proves a file is visible to whichever lint chooses to inspect
// it.
func WalkInternalTree(internalRoot string) (map[string][]byte, error) {
	result := make(map[string][]byte)

	err := filepath.WalkDir(internalRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			if path != internalRoot && (d.Name() == "testdata" || d.Name() == "vendor") {
				return filepath.SkipDir
			}
			return nil
		}

		name := d.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}

		rel, relErr := filepath.Rel(internalRoot, path)
		if relErr != nil {
			return fmt.Errorf("lintscan: relativize %s against %s: %w", path, internalRoot, relErr)
		}

		src, readErr := os.ReadFile(path) //nolint:gosec // path is produced by filepath.WalkDir over a caller-provided root, not untrusted external input
		if readErr != nil {
			return fmt.Errorf("lintscan: read %s: %w", path, readErr)
		}

		result[filepath.ToSlash(rel)] = src
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// WalkInternalTreeT is the ergonomic wrapper real lint call-sites use: it calls os.Getwd(),
// then FindModuleRoot(cwd), then WalkInternalTree(filepath.Join(moduleRoot, "internal")), and
// calls t.Fatalf with a clear message on any error instead of returning one — mirroring the
// existing walkRepositoryPreloads/walkServiceN1-style helpers' t.Fatalf ergonomics so callers
// need no extra error handling.
func WalkInternalTreeT(t *testing.T) map[string][]byte {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("lintscan: os.Getwd: %v", err)
	}

	moduleRoot, err := FindModuleRoot(cwd)
	if err != nil {
		t.Fatalf("lintscan: FindModuleRoot(%s): %v", cwd, err)
	}

	tree, err := WalkInternalTree(filepath.Join(moduleRoot, "internal"))
	if err != nil {
		t.Fatalf("lintscan: WalkInternalTree(%s): %v", filepath.Join(moduleRoot, "internal"), err)
	}

	return tree
}
