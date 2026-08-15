package main

import (
	"fmt"
	"go/ast"
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

// ARCH-A5: keep cmd/api a composition root — no god Services/Repositories aggregator,
// no retired layer imports, no domain wiring dumped into main.go.

var retiredLayerImportPrefixes = []string{
	"github.com/animal-ekarte/backend/internal/handler",
	"github.com/animal-ekarte/backend/internal/service",
	"github.com/animal-ekarte/backend/internal/repository",
}

// Exact package-level type names that recreate the pre-BE9 god aggregators.
// Domain-scoped names (authRepositories, runtimeRepositories, …) remain allowed.
var forbiddenGodTypeNames = map[string]struct{}{
	"Services":     {},
	"Repositories": {},
}

// Required composition entry files — new domains should add composition_<domain>.go
// rather than expanding main.go.
var requiredCompositionFiles = []string{
	"composition_runtime.go",
	"composition_auth.go",
	"composition_staff.go",
	"composition_clinic.go",
	"composition_owner_pet.go",
	"composition_reservation.go",
	"composition_billing.go",
	"composition_medicalrecord.go",
}

func TestCompositionRootConventions(t *testing.T) {
	apiDir := mustAPIDir(t)

	t.Run("required_composition_files_exist", func(t *testing.T) {
		for _, name := range requiredCompositionFiles {
			path := filepath.Join(apiDir, name)
			if _, err := os.Stat(path); err != nil {
				t.Errorf("missing required composition file %s: %v", name, err)
			}
		}
	})

	t.Run("production_sources", func(t *testing.T) {
		files, err := listAPIProductionGoFiles(apiDir)
		if err != nil {
			t.Fatalf("list cmd/api production sources: %v", err)
		}
		if len(files) < 10 {
			t.Fatalf("expected multiple composition sources, found %d", len(files))
		}

		var violations []string
		fset := token.NewFileSet()
		for _, path := range files {
			base := filepath.Base(path)
			src, err := os.ReadFile(path) //nolint:gosec // path from WalkDir of package dir
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
			if err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			violations = append(violations, scanCompositionConventions(base, file)...)
		}
		if len(violations) > 0 {
			sort.Strings(violations)
			t.Fatalf("composition root convention violations:\n%s", strings.Join(violations, "\n"))
		}
	})

	t.Run("mutation_god_type_fails", func(t *testing.T) {
		src := `package main
type Services struct { A int }
type Repositories struct { B int }
`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "synthetic.go", src, 0)
		if err != nil {
			t.Fatal(err)
		}
		got := scanCompositionConventions("synthetic.go", file)
		if len(got) < 2 {
			t.Fatalf("expected god type violations, got %v", got)
		}
	})

	t.Run("mutation_retired_import_fails", func(t *testing.T) {
		src := `package main
import _ "github.com/animal-ekarte/backend/internal/service/foo"
`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "synthetic.go", src, 0)
		if err != nil {
			t.Fatal(err)
		}
		got := scanCompositionConventions("main.go", file)
		if len(got) == 0 {
			t.Fatal("expected retired layer import violation")
		}
	})

	t.Run("mutation_main_domain_repo_ctor_fails", func(t *testing.T) {
		src := `package main
func wire() {
	_ = newAuthRepositories(nil)
}
`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "main.go", src, 0)
		if err != nil {
			t.Fatal(err)
		}
		got := scanCompositionConventions("main.go", file)
		if len(got) == 0 {
			t.Fatal("expected main.go domain wiring call violation")
		}
	})

	t.Run("mutation_allowed_domain_scoped_type_passes", func(t *testing.T) {
		src := `package main
type authRepositories struct{}
type runtimeComposition struct{}
func newAuthRepositories() authRepositories { return authRepositories{} }
`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "composition_auth.go", src, 0)
		if err != nil {
			t.Fatal(err)
		}
		got := scanCompositionConventions("composition_auth.go", file)
		if len(got) != 0 {
			t.Fatalf("expected clean domain-scoped types, got %v", got)
		}
	})
}

func mustAPIDir(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// go test cwd is the package directory (cmd/api).
	if filepath.Base(cwd) != "api" {
		// Allow running from module root in some harnesses.
		candidate := filepath.Join(cwd, "cmd", "api")
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			return candidate
		}
	}
	return cwd
}

func listAPIProductionGoFiles(apiDir string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(apiDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path != apiDir {
				return filepath.SkipDir // cmd/api has no production subpackages
			}
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	sort.Strings(paths)
	return paths, err
}

func scanCompositionConventions(baseName string, file *ast.File) []string {
	var violations []string

	for _, imp := range file.Imports {
		importPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			violations = append(violations, fmt.Sprintf("%s: bad import literal", baseName))
			continue
		}
		for _, prefix := range retiredLayerImportPrefixes {
			if importPath == prefix || strings.HasPrefix(importPath, prefix+"/") {
				violations = append(violations, fmt.Sprintf(
					"%s: imports retired layer package %q (use domain packages / composition_*.go)",
					baseName, importPath,
				))
			}
		}
	}

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if _, forbidden := forbiddenGodTypeNames[ts.Name.Name]; forbidden {
				violations = append(violations, fmt.Sprintf(
					"%s: forbidden god aggregator type %q (use domain-scoped *Composition / *Repositories)",
					baseName, ts.Name.Name,
				))
			}
		}
	}

	if baseName == "main.go" {
		violations = append(violations, scanMainGoForDomainWiring(file)...)
	}

	return violations
}

// Domain wiring helpers belong in composition_*.go. main.go may only bootstrap.
var mainForbiddenCallNames = map[string]struct{}{
	"newAuthRepositories":           {},
	"newStaffRepositories":          {},
	"newClinicRepositories":         {},
	"newOwnerPetRepositories":       {},
	"newReservationRepositories":    {},
	"newBillingRepositories":        {},
	"newMedicalRecordRepositories":  {},
	"newRuntimeRepositories":        {},
	"newRuntimeComposition":         {},
	"newAuthComposition":            {},
	"newStaffComposition":           {},
	"newClinicComposition":          {},
	"newOwnerPetComposition":        {},
	"newReservationComposition":     {},
	"newBillingComposition":         {},
	"newMedicalRecordComposition":   {},
}

func scanMainGoForDomainWiring(file *ast.File) []string {
	var violations []string
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := call.Fun.(*ast.Ident)
		if !ok {
			return true
		}
		if _, forbidden := mainForbiddenCallNames[ident.Name]; forbidden {
			violations = append(violations, fmt.Sprintf(
				"main.go: must not call domain wiring helper %s (keep bootstrap-only; use composition_*.go / runtime_execution)",
				ident.Name,
			))
		}
		return true
	})
	return violations
}
