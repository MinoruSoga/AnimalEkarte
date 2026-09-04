package lintscan

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// ARCH-A8: pin intentional exception-package discipline so cutover helpers and
// isolated domains do not silently become a second architecture.

// csvimport may only be imported from cmd tooling (not online API domains).
var csvImportAllowedImporterPrefixes = []string{
	"cmd/csv-import/",
	"cmd/csv-import-failure-rehearsal/",
	"cmd/csv-import-stg-uat/",
	"cmd/seed-export/",
	"cmd/stg-uat-skeleton/",
}

const csvImportPath = "github.com/animal-ekarte/backend/internal/csvimport"

var identityLinkForbiddenDomainImports = map[string]struct{}{
	"owner": {},
	"pet":   {},
}

func TestExceptionPackageDiscipline(t *testing.T) {
	t.Run("csvimport_cmd_only_real_tree", func(t *testing.T) {
		moduleRoot := mustModuleRoot(t)
		violations, err := findCSVImportConsumerViolations(moduleRoot)
		if err != nil {
			t.Fatalf("scan csvimport consumers: %v", err)
		}
		if len(violations) > 0 {
			t.Fatalf("csvimport must stay cmd/tooling-only (A8-1):\n%s", strings.Join(violations, "\n"))
		}
	})

	t.Run("identitylink_forbids_owner_pet_real_tree", func(t *testing.T) {
		files := WalkInternalTreeT(t)
		violations := findIdentityLinkOwnerPetImportViolations(files)
		if len(violations) > 0 {
			t.Fatalf("identitylink must not import owner/pet (A8-3):\n%s", strings.Join(violations, "\n"))
		}
	})

	t.Run("mutation_csvimport_importer_prefix", func(t *testing.T) {
		if csvImportImporterAllowed("internal/billing/foo.go") {
			t.Fatal("expected internal/billing importer to be rejected")
		}
		if !csvImportImporterAllowed("cmd/csv-import/main.go") {
			t.Fatal("expected cmd/csv-import to be allowed")
		}
		if !csvImportImporterAllowed("cmd/seed-export/main.go") {
			t.Fatal("expected cmd/seed-export to be allowed")
		}
		if !csvImportImporterAllowed("cmd/csv-import-failure-rehearsal/main.go") {
			t.Fatal("expected failure-rehearsal to be allowed")
		}
		if !csvImportImporterAllowed("cmd/csv-import-stg-uat/main.go") {
			t.Fatal("expected STG UAT import tooling to be allowed")
		}
		if !csvImportImporterAllowed("cmd/stg-uat-skeleton/main.go") {
			t.Fatal("expected STG UAT skeleton tooling to be allowed")
		}
	})

	t.Run("mutation_identitylink_owner_import_fails", func(t *testing.T) {
		files := map[string][]byte{
			"identitylink/bad.go": []byte(`package identitylink
import "github.com/animal-ekarte/backend/internal/owner"
`),
		}
		got := findIdentityLinkOwnerPetImportViolations(files)
		if len(got) == 0 {
			t.Fatal("expected identitylink -> owner violation")
		}
	})

	t.Run("mutation_identitylink_httpapi_passes", func(t *testing.T) {
		files := map[string][]byte{
			"identitylink/ok.go": []byte(`package identitylink
import "github.com/animal-ekarte/backend/internal/httpapi"
`),
		}
		got := findIdentityLinkOwnerPetImportViolations(files)
		if len(got) != 0 {
			t.Fatalf("expected clean identitylink imports, got %v", got)
		}
	})
}

func mustModuleRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root, err := FindModuleRoot(cwd)
	if err != nil {
		t.Fatalf("FindModuleRoot: %v", err)
	}
	return root
}

func findCSVImportConsumerViolations(moduleRoot string) ([]string, error) {
	var violations []string
	err := filepath.WalkDir(moduleRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if path != moduleRoot && (name == "testdata" || name == "vendor" || name == ".git") {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(moduleRoot, path)
		if err != nil {
			return err
		}
		slash := filepath.ToSlash(rel)
		src, err := os.ReadFile(path) //nolint:gosec // path from WalkDir over module root
		if err != nil {
			return err
		}
		if !sourceImportsPath(src, csvImportPath) {
			return nil
		}
		if !csvImportImporterAllowed(slash) {
			violations = append(violations, fmt.Sprintf(
				"%s: imports %s (allowed only under cmd csv-import / seed-export tools)",
				slash, csvImportPath,
			))
		}
		return nil
	})
	sort.Strings(violations)
	return violations, err
}

func csvImportImporterAllowed(slashRelPath string) bool {
	for _, prefix := range csvImportAllowedImporterPrefixes {
		if strings.HasPrefix(slashRelPath, prefix) {
			return true
		}
	}
	return false
}

func findIdentityLinkOwnerPetImportViolations(files map[string][]byte) []string {
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var violations []string
	fset := token.NewFileSet()
	for _, rel := range paths {
		if !strings.HasPrefix(filepath.ToSlash(rel), "identitylink/") {
			continue
		}
		file, err := parser.ParseFile(fset, rel, files[rel], parser.ImportsOnly)
		if err != nil {
			violations = append(violations, fmt.Sprintf("%s: parse: %v", rel, err))
			continue
		}
		for _, imp := range file.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				violations = append(violations, fmt.Sprintf("%s: unquote: %v", rel, err))
				continue
			}
			if !strings.HasPrefix(importPath, backendInternalImportPrefix) {
				continue
			}
			rest := strings.TrimPrefix(importPath, backendInternalImportPrefix)
			seg := rest
			if i := strings.IndexByte(rest, '/'); i >= 0 {
				seg = rest[:i]
			}
			if _, forbidden := identityLinkForbiddenDomainImports[seg]; forbidden {
				violations = append(violations, fmt.Sprintf(
					"%s: identitylink imports %q (forbidden; keep identity isolation)",
					rel, seg,
				))
			}
		}
	}
	return violations
}

func sourceImportsPath(src []byte, importPath string) bool {
	if !strings.Contains(string(src), importPath) {
		return false
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", src, parser.ImportsOnly)
	if err != nil {
		// Fail closed: string match without parse still counts as potential import.
		return true
	}
	for _, imp := range file.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err == nil && (p == importPath || strings.HasPrefix(p, importPath+"/")) {
			return true
		}
	}
	return false
}
