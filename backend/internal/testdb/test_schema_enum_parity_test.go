package testdb

// test_schema_enum_parity_test.go — BE-refactor.md G12-2
//
// ─── Purpose ────────────────────────────────────────────────────────────────────────
//
// setupSharedTestSchema (ltv_repository_test.go) hand-duplicates every PostgreSQL ENUM
// type from migrations/001_init.sql because the shared test schema is built via direct
// DDL execution (once per process, guarded by sharedTestSchemaOnce), not by replaying the
// migration file. Hand duplication drifts silently: G12-2 found 'trimming' missing from
// item_source (blocking any integration test that persists a trimming billing_items row —
// billing_item_repository.go:271 sets Source: model.ItemSourceTrimming) and 4 whole ENUM
// types absent from the test double (campaign_discount_type / lab_import_job_status /
// lab_import_source_type / checkup_field_type).
//
// This gate extracts every "CREATE TYPE ... AS ENUM (...)" statement from 001_init.sql and
// compares it, name-for-name and value-for-value (order-sensitive — PostgreSQL enum value
// order affects comparison operators), against sharedTestSchemaEnumTypes. A new migration
// ENUM type, an edited value list, or a stale/removed test-double entry fails this gate
// until sharedTestSchemaEnumTypes is updated to match.
//
// ─── Technique ──────────────────────────────────────────────────────────────────────
//
// migrations/*.sql is a sibling directory of internal/repository/, so go:embed's relative
// path restriction (no "..") applies here too. Reuses the same os.ReadFile relative-path
// approach as the other migration-reading tests in this repo ("../../migrations" relative path).
//
// ─── Deliberate exceptions ──────────────────────────────────────────────────────────
//
// testSchemaEnumParityAllowlist holds documented exceptions — currently none. checkup_field_type
// was evaluated for exclusion (its own real-DDL test trio DROP+CREATEs it via an isolated
// connection — see checkup_field_repository_test.go) but was added rather than skipped: see the
// judgment-call comment on sharedTestSchemaEnumTypes in ltv_repository_test.go for why it does
// not collide.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// createTypeRe extracts "CREATE TYPE <name> AS ENUM (...)" statements from 001_init.sql,
// including multi-line definitions (e.g. reservation_status, lab_import_job_status,
// checkup_field_type).
var createTypeRe = regexp.MustCompile(`(?s)CREATE TYPE\s+(\w+)\s+AS ENUM\s*\((.*?)\)\s*;`)

// testSchemaEnumParityAllowlist maps an ENUM type name to a documented reason it is
// intentionally absent from sharedTestSchemaEnumTypes. Empty today — add entries only with a
// reason a future reader can verify (e.g. "no setupTestDB-based test AutoMigrates a model
// using this type").
var testSchemaEnumParityAllowlist = map[string]string{}

// extractSQLEnumTypes parses every CREATE TYPE ... AS ENUM statement out of 001_init.sql's raw
// text, returning name -> canonical single-line definition in the same style used by
// sharedTestSchemaEnumTypes ("CREATE TYPE <name> AS ENUM ('a', 'b', ...)", no trailing ";").
func extractSQLEnumTypes(sql string) map[string]string {
	found := make(map[string]string)
	for _, m := range createTypeRe.FindAllStringSubmatch(sql, -1) {
		name, rawValues := m[1], m[2]
		values := EnumValueRe.FindAllString(rawValues, -1)
		found[name] = "CREATE TYPE " + name + " AS ENUM (" + strings.Join(values, ", ") + ")"
	}
	return found
}

// goEnumTypes returns testdb.SharedTestSchemaEnumTypes as a name -> definition map.
func goEnumTypes() map[string]string {
	out := make(map[string]string, len(SharedTestSchemaEnumTypes))
	for _, et := range SharedTestSchemaEnumTypes {
		out[et.Name] = et.Create
	}
	return out
}

// reconcileTestSchemaEnumParity is a pure function: it compares sqlTypes (extracted from
// 001_init.sql) against goTypes (sharedTestSchemaEnumTypes) and returns a human-readable
// violation for every missing, value-mismatched, stale, or orphaned entry not covered by
// allowlist.
func reconcileTestSchemaEnumParity(sqlTypes, goTypes, allowlist map[string]string) []string {
	var violations []string

	for name, sqlDef := range sqlTypes {
		if _, skipped := allowlist[name]; skipped {
			continue
		}
		goDef, ok := goTypes[name]
		switch {
		case !ok:
			violations = append(violations, "ENUM type "+name+" exists in 001_init.sql but is "+
				"missing from sharedTestSchemaEnumTypes (ltv_repository_test.go) — add it, or add "+
				"a documented exception to testSchemaEnumParityAllowlist. SQL definition: "+sqlDef)
		case goDef != sqlDef:
			violations = append(violations, "ENUM type "+name+" drifted: sharedTestSchemaEnumTypes "+
				"has \""+goDef+"\" but 001_init.sql defines \""+sqlDef+"\" — sync the test double.")
		}
	}

	for name, goDef := range goTypes {
		if _, ok := sqlTypes[name]; !ok {
			violations = append(violations, "sharedTestSchemaEnumTypes has ENUM type "+name+
				" (\""+goDef+"\") that does not exist in 001_init.sql — remove it (typo or stale "+
				"entry from a since-removed migration).")
		}
	}

	for name := range allowlist {
		if _, ok := sqlTypes[name]; !ok {
			violations = append(violations, "testSchemaEnumParityAllowlist entry "+name+" no "+
				"longer matches an existing ENUM type in 001_init.sql — remove the stale entry.")
		}
	}

	return violations
}

// TestTestSchemaEnumParity is the gate: every ENUM type CREATE TYPE ... AS ENUM statement in
// 001_init.sql must have a byte-for-byte matching entry in sharedTestSchemaEnumTypes, and vice
// versa (modulo documented exceptions). A floor guards against a vacuous pass if the extraction
// regex silently breaks.
func TestTestSchemaEnumParity(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("../../migrations", "001_init.sql")) //nolint:gocritic // B5b requires this relative path.
	if err != nil {
		t.Fatalf("read 001_init.sql: %v", err)
	}

	sqlTypes := extractSQLEnumTypes(string(raw))
	// 54 CREATE TYPE ... AS ENUM statements in 001_init.sql as of G12-2 (2026-07-10). Revisit
	// this floor if migrations legitimately add/remove ENUM types.
	if len(sqlTypes) < 54 {
		t.Fatalf("extracted only %d CREATE TYPE ... AS ENUM statement(s) from 001_init.sql; "+
			"expected 54+ — the extraction regex likely broke (would vacuously pass)", len(sqlTypes))
	}

	for _, v := range reconcileTestSchemaEnumParity(sqlTypes, goEnumTypes(), testSchemaEnumParityAllowlist) {
		t.Error(v)
	}
}

// TestReconcileTestSchemaEnumParity_Analyzer pins the reconciler's behavior on inline fixtures.
func TestReconcileTestSchemaEnumParity_Analyzer(t *testing.T) {
	t.Run("clean baseline reports no violation", func(t *testing.T) {
		sqlTypes := map[string]string{"pet_status": "CREATE TYPE pet_status AS ENUM ('alive', 'deceased')"}
		goTypes := map[string]string{"pet_status": "CREATE TYPE pet_status AS ENUM ('alive', 'deceased')"}
		got := reconcileTestSchemaEnumParity(sqlTypes, goTypes, map[string]string{})
		if len(got) != 0 {
			t.Fatalf("expected 0 violations, got %v", got)
		}
	})

	t.Run("missing from go types fails", func(t *testing.T) {
		sqlTypes := map[string]string{"item_source": "CREATE TYPE item_source AS ENUM ('manual', 'trimming')"}
		got := reconcileTestSchemaEnumParity(sqlTypes, map[string]string{}, map[string]string{})
		if len(got) != 1 || !strings.Contains(got[0], "missing from sharedTestSchemaEnumTypes") {
			t.Fatalf("expected missing-type violation, got %v", got)
		}
	})

	t.Run("value drift fails", func(t *testing.T) {
		sqlTypes := map[string]string{"item_source": "CREATE TYPE item_source AS ENUM ('manual', 'trimming')"}
		goTypes := map[string]string{"item_source": "CREATE TYPE item_source AS ENUM ('manual')"}
		got := reconcileTestSchemaEnumParity(sqlTypes, goTypes, map[string]string{})
		if len(got) != 1 || !strings.Contains(got[0], "drifted") {
			t.Fatalf("expected drift violation, got %v", got)
		}
	})

	t.Run("orphaned go-only entry fails", func(t *testing.T) {
		goTypes := map[string]string{"typo_type": "CREATE TYPE typo_type AS ENUM ('x')"}
		got := reconcileTestSchemaEnumParity(map[string]string{}, goTypes, map[string]string{})
		if len(got) != 1 || !strings.Contains(got[0], "does not exist in 001_init.sql") {
			t.Fatalf("expected orphan violation, got %v", got)
		}
	})

	t.Run("allowlisted missing type does not fail", func(t *testing.T) {
		sqlTypes := map[string]string{"checkup_field_type": "CREATE TYPE checkup_field_type AS ENUM ('number')"}
		got := reconcileTestSchemaEnumParity(sqlTypes, map[string]string{}, map[string]string{"checkup_field_type": "documented reason"})
		if len(got) != 0 {
			t.Fatalf("expected 0 violations for allowlisted type, got %v", got)
		}
	})

	t.Run("stale allowlist entry fails", func(t *testing.T) {
		got := reconcileTestSchemaEnumParity(map[string]string{}, map[string]string{}, map[string]string{"removed_type": "reason"})
		if len(got) != 1 || !strings.Contains(got[0], "no longer matches") {
			t.Fatalf("expected stale allowlist violation, got %v", got)
		}
	})

	t.Run("extracts multi-line CREATE TYPE and matches order-sensitively", func(t *testing.T) {
		sql := "CREATE TYPE reservation_status AS ENUM (\n    'confirmed', 'pending',\n    'cancelled'\n);"
		got := extractSQLEnumTypes(sql)
		want := "CREATE TYPE reservation_status AS ENUM ('confirmed', 'pending', 'cancelled')"
		if got["reservation_status"] != want {
			t.Fatalf("got %q, want %q", got["reservation_status"], want)
		}
	})
}
