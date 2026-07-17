package model

import (
	_ "embed"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

//go:embed audit_log.go
var auditLogSrc []byte

// auditTaxonomyConstants parses audit_log.go (embedded at compile time) and returns:
//   - declaredConsts: unquoted string values of typed const declarations matching typeName
//   - mapKeys:        unquoted string values of keys in the var declaration named mapVarName
//
// Constants and map keys are discovered from source, not from a manually maintained slice.
// A developer who adds a constant without updating the map (or vice versa) will cause the
// relevant exhaustiveness test to fail.
//
// The source is embedded via go:embed so the test works correctly under -trimpath,
// in embedded test binaries, and on any OS without path-resolution logic.
func auditTaxonomyConstants(t *testing.T, typeName, mapVarName string) (declaredConsts, mapKeys map[string]struct{}) {
	t.Helper()

	fset := token.NewFileSet()
	// Pass auditLogSrc as the src argument; "audit_log.go" is used only in error messages.
	f, err := parser.ParseFile(fset, "audit_log.go", auditLogSrc, 0)
	if err != nil {
		t.Fatalf("failed to parse embedded audit_log.go: %v", err)
	}

	// --- Pass 1: collect typed constants and build name→unquoted-value map ---
	declaredConsts = make(map[string]struct{})
	nameToValue := make(map[string]string) // const name → unquoted string value

	ast.Inspect(f, func(n ast.Node) bool {
		gd, ok := n.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			return true
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			if vs.Type == nil {
				continue
			}
			ident, ok := vs.Type.(*ast.Ident)
			if !ok || ident.Name != typeName {
				continue
			}
			for i, name := range vs.Names {
				if i >= len(vs.Values) {
					continue
				}
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				raw, err := strconv.Unquote(lit.Value)
				if err != nil {
					continue
				}
				declaredConsts[raw] = struct{}{}
				nameToValue[name.Name] = raw
			}
		}
		return true
	})

	// --- Pass 2: collect map literal keys from the named var declaration ---
	mapKeys = make(map[string]struct{})

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
			for _, nm := range vs.Names {
				if nm.Name == mapVarName {
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
					switch k := kv.Key.(type) {
					case *ast.Ident:
						// Key is a constant name — resolve via name→value map from pass 1.
						if v, found := nameToValue[k.Name]; found {
							mapKeys[v] = struct{}{}
						}
					case *ast.BasicLit:
						if k.Kind == token.STRING {
							raw, err := strconv.Unquote(k.Value)
							if err == nil {
								mapKeys[raw] = struct{}{}
							}
						}
					}
				}
			}
		}
	}

	return declaredConsts, mapKeys
}

// ------------------------------------
// Phase 4A.4 — LabBlockedReason exhaustiveness
// ------------------------------------

// TestLabBlockedReason_Exhaustiveness_ConstantsInMap verifies that every declared
// LabBlockedReason constant in audit_log.go is present in validLabBlockedReasons.
// A developer who adds a constant but forgets to update the map will see this test fail.
func TestLabBlockedReason_Exhaustiveness_ConstantsInMap(t *testing.T) {
	declared, mapKs := auditTaxonomyConstants(t, "LabBlockedReason", "validLabBlockedReasons")

	if len(declared) == 0 {
		t.Fatal("no LabBlockedReason constants found — check the type name and embedded source")
	}
	for v := range declared {
		if _, ok := mapKs[v]; !ok {
			t.Errorf("LabBlockedReason constant %q is declared but missing from validLabBlockedReasons map", v)
		}
		if !ValidLabBlockedReason(LabBlockedReason(v)) {
			t.Errorf("ValidLabBlockedReason(%q) = false; want true for declared constant", v)
		}
	}
}

// TestLabBlockedReason_Exhaustiveness_MapHasNoExtras verifies that every key in
// validLabBlockedReasons corresponds to a declared LabBlockedReason constant.
// A developer who adds a raw string to the map without a constant will see this test fail.
func TestLabBlockedReason_Exhaustiveness_MapHasNoExtras(t *testing.T) {
	declared, mapKs := auditTaxonomyConstants(t, "LabBlockedReason", "validLabBlockedReasons")

	if len(mapKs) == 0 {
		t.Fatal("no keys found in validLabBlockedReasons — check the map variable name and embedded source")
	}
	for v := range mapKs {
		if _, ok := declared[v]; !ok {
			t.Errorf("validLabBlockedReasons contains %q but no matching LabBlockedReason constant is declared", v)
		}
	}
}

// TestLabBlockedReason_Exhaustiveness_SetsSizeMatch is a sanity check that both sets
// are the same size, catching simultaneous add+remove that fools the directional tests.
func TestLabBlockedReason_Exhaustiveness_SetsSizeMatch(t *testing.T) {
	declared, mapKs := auditTaxonomyConstants(t, "LabBlockedReason", "validLabBlockedReasons")

	if len(declared) != len(mapKs) {
		t.Errorf("LabBlockedReason: %d declared constants vs %d map keys — sets must be equal",
			len(declared), len(mapKs))
	}
}

// ------------------------------------
// Phase 4A.4 — LabAuditErrorCategory exhaustiveness
// ------------------------------------

// TestLabAuditErrorCategory_Exhaustiveness_ConstantsInMap verifies that every declared
// LabAuditErrorCategory constant in audit_log.go is present in validLabAuditErrorCategories.
func TestLabAuditErrorCategory_Exhaustiveness_ConstantsInMap(t *testing.T) {
	declared, mapKs := auditTaxonomyConstants(t, "LabAuditErrorCategory", "validLabAuditErrorCategories")

	if len(declared) == 0 {
		t.Fatal("no LabAuditErrorCategory constants found — check the type name and embedded source")
	}
	for v := range declared {
		if _, ok := mapKs[v]; !ok {
			t.Errorf("LabAuditErrorCategory constant %q is declared but missing from validLabAuditErrorCategories map", v)
		}
		if !ValidLabAuditErrorCategory(LabAuditErrorCategory(v)) {
			t.Errorf("ValidLabAuditErrorCategory(%q) = false; want true for declared constant", v)
		}
	}
}

// TestLabAuditErrorCategory_Exhaustiveness_MapHasNoExtras verifies that every key in
// validLabAuditErrorCategories corresponds to a declared LabAuditErrorCategory constant.
func TestLabAuditErrorCategory_Exhaustiveness_MapHasNoExtras(t *testing.T) {
	declared, mapKs := auditTaxonomyConstants(t, "LabAuditErrorCategory", "validLabAuditErrorCategories")

	if len(mapKs) == 0 {
		t.Fatal("no keys found in validLabAuditErrorCategories — check the map variable name and embedded source")
	}
	for v := range mapKs {
		if _, ok := declared[v]; !ok {
			t.Errorf("validLabAuditErrorCategories contains %q but no matching LabAuditErrorCategory constant is declared", v)
		}
	}
}

// TestLabAuditErrorCategory_Exhaustiveness_SetsSizeMatch is a sanity check that both
// sets are the same size.
func TestLabAuditErrorCategory_Exhaustiveness_SetsSizeMatch(t *testing.T) {
	declared, mapKs := auditTaxonomyConstants(t, "LabAuditErrorCategory", "validLabAuditErrorCategories")

	if len(declared) != len(mapKs) {
		t.Errorf("LabAuditErrorCategory: %d declared constants vs %d map keys — sets must be equal",
			len(declared), len(mapKs))
	}
}

// ------------------------------------
// Phase 4A.4 — Parser self-verification
// ------------------------------------

// TestAuditTaxonomyParser_DetectsConstantsAndMapKeys proves the parser discovers
// the expected counts from the embedded source. This is a regression anchor: if the
// parser were silently broken, these counts would be zero and every other
// exhaustiveness test would vacuously pass (the Fatal guards prevent that, but this
// test pins the actual parsed sizes independently).
//
// When a new constant is added to either taxonomy, update the count here too.
func TestAuditTaxonomyParser_DetectsConstantsAndMapKeys(t *testing.T) {
	// LabBlockedReason: 3 constants, 3 map keys.
	brDecl, brMap := auditTaxonomyConstants(t, "LabBlockedReason", "validLabBlockedReasons")
	if len(brDecl) != 3 {
		t.Errorf("expected 3 LabBlockedReason constants, got %d", len(brDecl))
	}
	if len(brMap) != 3 {
		t.Errorf("expected 3 validLabBlockedReasons map keys, got %d", len(brMap))
	}

	// LabAuditErrorCategory: 5 constants, 5 map keys.
	ecDecl, ecMap := auditTaxonomyConstants(t, "LabAuditErrorCategory", "validLabAuditErrorCategories")
	if len(ecDecl) != 5 {
		t.Errorf("expected 5 LabAuditErrorCategory constants, got %d", len(ecDecl))
	}
	if len(ecMap) != 5 {
		t.Errorf("expected 5 validLabAuditErrorCategories map keys, got %d", len(ecMap))
	}
}

// TestAuditTaxonomyParser_WrongMapNameYieldsEmptyKeys proves the failure-detection path:
// a nonexistent map name produces zero map keys, confirming that ConstantsInMap would
// report every constant as missing rather than vacuously passing.
func TestAuditTaxonomyParser_WrongMapNameYieldsEmptyKeys(t *testing.T) {
	declared, mapKs := auditTaxonomyConstants(t, "LabBlockedReason", "nonExistentMap")

	if len(declared) == 0 {
		t.Error("parser must still find LabBlockedReason constants even with wrong map name")
	}
	if len(mapKs) != 0 {
		t.Errorf("nonexistent map name should produce 0 keys, got %d", len(mapKs))
	}
}
