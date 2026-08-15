package lintscan

// preload_master_model_reconciliation_test.go — BE-refactor.md R3-2 (D9・cross-check follow-up): automated
// cross-check between the manually curated clinicScopedMasterAssoc (read side, this
// package's preload_clinic_scope_lint_test.go) and two independent sources of truth:
//
//  1. internal/model: every struct that actually declares a ClinicID field is the only
//     reliable ground truth for "does this Go type have a clinic_id column" — a hand-written
//     comment can drift out of sync with the struct definition, but the AST cannot.
//  2. internal/lintscan's clinicScopedMasterFKField (master_fk_write_inventory_lint_test.go,
//     the write-side clinic-scope registry): an INDEPENDENTLY curated list of clinic-scoped master
//     model names. Before this gate, a model newly registered on the write side had no
//     automated check forcing its read-side registration — repository/CLAUDE.md's "新マスタ
//     追加時" note previously relied on a human remembering to update both lists by hand.
//     If a Preload of that master were added later without first registering it here, it
//     would slip past TestPreloadClinicScope_RealRepositorySourceHasNoUnscopedMasterPreload's
//     `!isMaster` branch completely unscoped.
//
// This gate does NOT attempt to derive the master/non-master classification purely from
// model-layer struct shape: internal/model has 70+ structs with a ClinicID column (Pet,
// Owner, Billing, Reservation, ...) that are core business data, not configurable "masters/
// 区分". That distinction is a domain judgment call the codebase already makes explicitly
// (see clinicScopedMasterAssoc's own comment re: Account/LineCustomer being intentionally
// excluded despite having clinic_id). Instead this gate validates the judgment calls already
// made on BOTH sides of the clinic-scope-Preload read/write split against EACH OTHER and against the
// compiler-checked model layer — the audit_tx_inventory_lint_test.go bidirectional-
// reconciliation technique applied across package boundaries instead of within one file.
//
// Cross-package source access: go:embed (used by the read/write lints for their OWN package
// source) cannot embed a sibling package's directory — embed patterns may not contain ".."
// path elements. Go's test runner guarantees the working directory is the tested package's
// own source directory (a long-standing, widely relied-upon "go test" behavior — every
// package's test binary is invoked with cwd = that package's directory), so os.Getwd() plus
// a relative ".." join reliably locates internal/model and internal/lintscan from within
// internal/lintscan under `go test ./...` — the exact invocation this project's CI uses.
// This has no dependency on -trimpath (which affects embedded/debug build paths, not the OS
// working directory queried by os.Getwd()).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// masterModelReadWriteExemptions: model names present in the write-side
// clinicScopedMasterFKField (internal/lintscan) but NOT YET in the read-side
// clinicScopedMasterAssoc, because internal/model has no gorm:"foreignKey:...ID" association
// field for them yet (a Preload of the association is syntactically impossible) — mirrors
// the trailing comment already on clinicScopedMasterAssoc. When an association field is
// added to internal/model, delete the entry here AND register the association name in
// clinicScopedMasterAssoc in the same commit (read/write gates stay bidirectionally in sync).
var masterModelReadWriteExemptions = map[string]struct{}{
	"MerchandiseItem":     {},
	"TrimmingCourseType":  {},
	"PaymentMethodMaster": {},
}

// extractModelClinicScopedStructs is a pure function over (filename -> source) pairs: it
// parses each Go source file and returns the set of struct type names that declare a
// ClinicID field of type uint64 or *uint64. Pure so the real internal/model source and
// inline test fixtures exercise identical logic (same convention as the sibling lints).
func extractModelClinicScopedStructs(sources map[string][]byte) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	for filename, src := range sources {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, filename, src, 0)
		if err != nil {
			return nil, err
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok || st.Fields == nil {
					continue
				}
				if structHasClinicIDField(st) {
					result[ts.Name.Name] = struct{}{}
				}
			}
		}
	}
	return result, nil
}

func structHasClinicIDField(st *ast.StructType) bool {
	for _, field := range st.Fields.List {
		if !fieldNamed(field, "ClinicID") {
			continue
		}
		if isUint64OrPtrUint64(field.Type) {
			return true
		}
	}
	return false
}

func fieldNamed(field *ast.Field, name string) bool {
	for _, n := range field.Names {
		if n.Name == name {
			return true
		}
	}
	return false
}

func isUint64OrPtrUint64(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name == "uint64"
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name == "uint64"
		}
	}
	return false
}

// extractStringMapValues is a pure function that parses one Go source file and returns the
// VALUE strings of every entry in the map[string]string composite literal assigned to the
// package-level var named varName (e.g. clinicScopedMasterFKField). Used to read the
// write-side registry out of internal/lintscan without importing its unexported identifier
// across packages.
func extractStringMapValues(filename string, src []byte, varName string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return nil, err
	}
	var values []string
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			matched := false
			for _, n := range vs.Names {
				if n.Name == varName {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
			for _, val := range vs.Values {
				cl, ok := val.(*ast.CompositeLit)
				if !ok {
					continue
				}
				for _, elt := range cl.Elts {
					kv, ok := elt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					bl, ok := kv.Value.(*ast.BasicLit)
					if !ok || bl.Kind != token.STRING {
						continue
					}
					s, err := strconv.Unquote(bl.Value)
					if err != nil {
						continue
					}
					values = append(values, s)
				}
			}
		}
	}
	return values, nil
}

// canonicalModelName strips a descriptive suffix (e.g. "ExamTypeField (sub-master of
// ExaminationType, #124)" -> "ExamTypeField") so a write-side registry value can be matched
// against a real internal/model struct name. A multi-word description with no single fixed
// type (e.g. the ParentID entry "self-ref master (checkup_type/consultation/exam_type/...)"
// — intentionally context-resolved to several different masters, not one) normalizes to a
// non-identifier token that will not match any real struct name; reconcileMasterModelCoverage
// treats that as "unresolvable, skip" rather than a false violation.
func canonicalModelName(raw string) string {
	if before, _, found := strings.Cut(raw, " "); found {
		return before
	}
	return raw
}

// reconcileMasterModelCoverage is the pure bidirectional check (same technique as
// audit_tx_inventory_lint_test.go's reconcileClinicalResultDeletes). Returns human-readable
// violations; empty = clean.
//
//   - modelClinicScoped: every internal/model struct name that has a ClinicID column
//     (ground truth from extractModelClinicScopedStructs against the real model/ source).
//   - writeModelNames: raw (un-normalized) values from clinicScopedMasterFKField.
//   - readModelNames: raw values from clinicScopedMasterAssoc (canonicalModelName is applied
//     uniformly even though current read-side values are already single tokens).
//   - exemptions: model names allowed to be write-only for now (masterModelReadWriteExemptions).
func reconcileMasterModelCoverage(modelClinicScoped map[string]struct{}, writeModelNames, readModelNames []string, exemptions map[string]struct{}) []string {
	var violations []string

	readSet := make(map[string]struct{}, len(readModelNames))
	for _, raw := range readModelNames {
		readSet[canonicalModelName(raw)] = struct{}{}
	}

	// (i) every read-side entry must resolve to a real internal/model struct that actually
	// has a ClinicID column — catches typos / stale renames / a model that lost clinic_id.
	for name := range readSet {
		if _, ok := modelClinicScoped[name]; !ok {
			violations = append(violations, "clinicScopedMasterAssoc references model \""+name+
				"\" which is not a struct in internal/model with a ClinicID column "+
				"(stale/renamed entry, or a model that no longer has clinic_id)")
		}
	}

	// (ii) every write-side entry that resolves to a real clinic_id-bearing model, and is not
	// documented as read-side-impossible-for-now, must also be registered on the read side.
	seenWrite := make(map[string]struct{})
	for _, raw := range writeModelNames {
		name := canonicalModelName(raw)
		if _, alreadyChecked := seenWrite[name]; alreadyChecked {
			continue
		}
		seenWrite[name] = struct{}{}
		if _, isModel := modelClinicScoped[name]; !isModel {
			continue // unresolved alias (e.g. "self-ref") or non-clinic_id sub-master (e.g. ExamTypeField) — outside this gate
		}
		if _, exempt := exemptions[name]; exempt {
			continue
		}
		if _, onReadSide := readSet[name]; !onReadSide {
			violations = append(violations,
				"model \""+name+"\" is registered in lintscan.clinicScopedMasterFKField (write-side "+
					"clinic-scope registry) and has a ClinicID column, but is missing from "+
					"lintscan.clinicScopedMasterAssoc (read-side registry). If a Preload of this "+
					"master already exists, register the association name in clinicScopedMasterAssoc. "+
					"If no association field exists in internal/model yet, add an entry to "+
					"masterModelReadWriteExemptions with the reason.")
		}
	}

	// (iii) every declared exemption must still be genuinely write-only (not stale).
	for name := range exemptions {
		if _, onReadSide := readSet[name]; onReadSide {
			violations = append(violations, "masterModelReadWriteExemptions[\""+name+
				"\"] is stale: this model is now registered in clinicScopedMasterAssoc — remove the exemption")
		}
	}

	return violations
}

// readSideMasterModelNames returns clinicScopedMasterAssoc's values directly — no parsing
// needed since it is a native var in this very package.
func readSideMasterModelNames() []string {
	values := make([]string, 0, len(clinicScopedMasterAssoc))
	for _, v := range clinicScopedMasterAssoc {
		values = append(values, v)
	}
	return values
}

// siblingPackageDir resolves internal/<name> relative to this package's own directory (see
// the file header for why this reads the real filesystem instead of go:embed).
func siblingPackageDir(t *testing.T, name string) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := filepath.Join(cwd, "..", name)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("expected sibling package directory %s to exist (cwd=%s): %v", dir, cwd, err)
	}
	return dir
}

// loadModelClinicScopedStructs walks internal/model (sibling of internal/lintscan) and
// returns the ClinicID-bearing struct set from the REAL source tree.
func loadModelClinicScopedStructs(t *testing.T) map[string]struct{} {
	t.Helper()
	dir := siblingPackageDir(t, "model")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read internal/model dir %s: %v", dir, err)
	}
	sources := make(map[string][]byte)
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), "_test.go") || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(dir, e.Name())) //nolint:gosec // dir is derived from os.Getwd(), not untrusted input
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		sources[e.Name()] = src
	}
	result, err := extractModelClinicScopedStructs(sources)
	if err != nil {
		t.Fatalf("parse internal/model source: %v", err)
	}
	return result
}

// loadWriteSideMasterModelNames reads internal/lintscan/master_fk_write_inventory_lint_test.go
// (sibling package) and returns clinicScopedMasterFKField's values from the REAL source.
func loadWriteSideMasterModelNames(t *testing.T) []string {
	t.Helper()
	dir := siblingPackageDir(t, "lintscan")
	path := filepath.Join(dir, "master_fk_write_inventory_lint_test.go")
	src, err := os.ReadFile(path) //nolint:gosec // path is derived from os.Getwd(), not untrusted input
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	values, err := extractStringMapValues(path, src, "clinicScopedMasterFKField")
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return values
}

// TestMasterModelReconciliation_RealSourceIsConsistent is the gate: every model registered
// on the write-side clinic-scope registry (and every model referenced by the read-side registry)
// must be mutually consistent, grounded in the real internal/model struct definitions. Floor
// guards prevent a vacuous green if cross-package file discovery or AST matching breaks.
func TestMasterModelReconciliation_RealSourceIsConsistent(t *testing.T) {
	modelStructs := loadModelClinicScopedStructs(t)
	if len(modelStructs) < 40 {
		t.Fatalf("only %d ClinicID-bearing model structs found; internal/model AST matching "+
			"likely broken (would vacuously pass)", len(modelStructs))
	}

	writeNames := loadWriteSideMasterModelNames(t)
	if len(writeNames) < 15 {
		t.Fatalf("only %d clinicScopedMasterFKField values found; internal/lintscan AST "+
			"matching likely broken", len(writeNames))
	}

	readNames := readSideMasterModelNames()
	if len(readNames) < 15 {
		t.Fatalf("only %d clinicScopedMasterAssoc values found; map likely broken/emptied", len(readNames))
	}

	for _, v := range reconcileMasterModelCoverage(modelStructs, writeNames, readNames, masterModelReadWriteExemptions) {
		t.Error(v)
	}
}

// TestMasterModelReconciliation_ExemptionsAreLive proves masterModelReadWriteExemptions is
// exercised against the real write-side registry, not dead code left over after a model
// gained its read-side registration or was renamed.
func TestMasterModelReconciliation_ExemptionsAreLive(t *testing.T) {
	if len(masterModelReadWriteExemptions) == 0 {
		t.Skip("no exemptions currently declared")
	}
	writeNames := loadWriteSideMasterModelNames(t)
	writeSet := make(map[string]struct{}, len(writeNames))
	for _, raw := range writeNames {
		writeSet[canonicalModelName(raw)] = struct{}{}
	}
	for name := range masterModelReadWriteExemptions {
		if _, ok := writeSet[name]; !ok {
			t.Errorf("masterModelReadWriteExemptions[%q] does not match any current "+
				"clinicScopedMasterFKField value; the exemption may be stale (model renamed/removed "+
				"from the write-side registry) — delete it", name)
		}
	}
}

// TestMasterModelReconciliation_Extractors pins the extractor behaviour on inline fixtures
// (RED->GREEN + regression freeze, same convention as the sibling lints' _Analyzer tests).
func TestMasterModelReconciliation_Extractors(t *testing.T) {
	t.Run("extractModelClinicScopedStructs finds ClinicID structs and skips others", func(t *testing.T) {
		src := []byte(`package model

type WithClinic struct {
	ID       uint64
	ClinicID uint64
}

type WithPtrClinic struct {
	ID       uint64
	ClinicID *uint64
}

type WithoutClinic struct {
	ID   uint64
	Name string
}
`)
		got, err := extractModelClinicScopedStructs(map[string][]byte{"fixture.go": src})
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if _, ok := got["WithClinic"]; !ok {
			t.Error("expected WithClinic to be detected")
		}
		if _, ok := got["WithPtrClinic"]; !ok {
			t.Error("expected WithPtrClinic (*uint64) to be detected")
		}
		if _, ok := got["WithoutClinic"]; ok {
			t.Error("WithoutClinic must not be detected (no ClinicID field)")
		}
	})

	t.Run("extractStringMapValues reads map[string]string literal values by var name", func(t *testing.T) {
		src := []byte(`package service

var someOtherVar = map[string]string{"x": "Y"}

var clinicScopedMasterFKField = map[string]string{
	"MedicineID": "Medicine",
	"VaccineID":  "Vaccine",
}
`)
		got, err := extractStringMapValues("fixture.go", src, "clinicScopedMasterFKField")
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		want := map[string]struct{}{"Medicine": {}, "Vaccine": {}}
		if len(got) != 2 {
			t.Fatalf("got %d values, want 2: %v", len(got), got)
		}
		for _, v := range got {
			if _, ok := want[v]; !ok {
				t.Errorf("unexpected value %q (must read clinicScopedMasterFKField only, not someOtherVar)", v)
			}
		}
	})

	t.Run("canonicalModelName strips descriptive suffix", func(t *testing.T) {
		cases := map[string]string{
			"Vaccine": "Vaccine",
			"ExamTypeField (sub-master of ExaminationType, #124)": "ExamTypeField",
			"self-ref master (checkup_type/consultation/...)":     "self-ref",
		}
		for in, want := range cases {
			if got := canonicalModelName(in); got != want {
				t.Errorf("canonicalModelName(%q) = %q, want %q", in, got, want)
			}
		}
	})
}

// TestMasterModelReconciliation_GateDetectsOmission proves the reconciliation function is not
// vacuously green. This is the harness-improvement fixture BE-refactor.md R3-2 asks for: "a
// fixture that intentionally omits one model-to-association mapping and proves the check
// fails."
func TestMasterModelReconciliation_GateDetectsOmission(t *testing.T) {
	modelStructs := map[string]struct{}{"Vaccine": {}, "Medicine": {}}

	t.Run("clean baseline reports no violation", func(t *testing.T) {
		got := reconcileMasterModelCoverage(modelStructs, []string{"Vaccine", "Medicine"}, []string{"Vaccine", "Medicine"}, nil)
		if len(got) != 0 {
			t.Fatalf("expected 0 violations, got %v", got)
		}
	})

	t.Run("model intentionally omitted from the read side fails", func(t *testing.T) {
		got := reconcileMasterModelCoverage(modelStructs, []string{"Vaccine", "Medicine"}, []string{"Vaccine"}, nil)
		if len(got) != 1 || !strings.Contains(got[0], "Medicine") {
			t.Fatalf("expected omitted Medicine mapping to be flagged, got %v", got)
		}
	})

	t.Run("documented exemption suppresses the write-only gap", func(t *testing.T) {
		got := reconcileMasterModelCoverage(modelStructs, []string{"Vaccine", "Medicine"}, []string{"Vaccine"}, map[string]struct{}{"Medicine": {}})
		if len(got) != 0 {
			t.Fatalf("expected exempted Medicine to suppress the violation, got %v", got)
		}
	})

	t.Run("stale exemption (model now registered on read side) fails", func(t *testing.T) {
		got := reconcileMasterModelCoverage(modelStructs, []string{"Vaccine", "Medicine"}, []string{"Vaccine", "Medicine"}, map[string]struct{}{"Medicine": {}})
		if len(got) != 1 || !strings.Contains(got[0], "stale") {
			t.Fatalf("expected stale exemption to be flagged, got %v", got)
		}
	})

	t.Run("stale read-side entry (model not in internal/model) fails", func(t *testing.T) {
		got := reconcileMasterModelCoverage(modelStructs, nil, []string{"Vaccine", "GhostModel"}, nil)
		if len(got) != 1 || !strings.Contains(got[0], "GhostModel") {
			t.Fatalf("expected stale read-side GhostModel entry to be flagged, got %v", got)
		}
	})

	t.Run("write-only alias that does not resolve to a real model is skipped", func(t *testing.T) {
		got := reconcileMasterModelCoverage(modelStructs, []string{"self-ref master (many kinds)"}, []string{"Vaccine", "Medicine"}, nil)
		if len(got) != 0 {
			t.Fatalf("expected unresolvable write-side alias to be silently skipped, got %v", got)
		}
	})
}
