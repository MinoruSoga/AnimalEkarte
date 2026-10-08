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
// This gate extracts every "CREATE TYPE ... AS ENUM (...)" statement from migrations/*.sql
// (all files, sorted — later files overwrite same-named types) and compares it,
// name-for-name and value-for-value (order-sensitive — PostgreSQL enum value order affects
// comparison operators), against sharedTestSchemaEnumTypes. A new migration ENUM type, an
// edited value list, or a stale/removed test-double entry fails this gate until
// sharedTestSchemaEnumTypes is updated to match.
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

// alterTypeAddValueRe extracts "ALTER TYPE <name> ADD VALUE [IF NOT EXISTS] '<value>'"
// statements (EMR-308: 028_appointments_tentative_status.sql adds 'tentative' to
// reservation_status this way because an enum value cannot be referenced in the same
// transaction that declares it). PostgreSQL appends the value at the end of the type's
// ordered list, so the fold below appends it to the canonical CREATE TYPE definition.
// BEFORE/AFTER positioning clauses are intentionally unsupported — a migration that
// needs them must extend this extractor first.
var alterTypeAddValueRe = regexp.MustCompile(`ALTER TYPE\s+(\w+)\s+ADD VALUE\s+(IF NOT EXISTS\s+)?('[^']*')\s*;`)

// enumLiteralRe extracts the quoted members of a folded CREATE TYPE definition so an
// ALTER TYPE ... ADD VALUE IF NOT EXISTS on an already-present value can be treated as
// the no-op PostgreSQL performs (e.g. 001_init.sql's folded 007 amendment, whose base
// definition already contains 'reverted').
var enumLiteralRe = regexp.MustCompile(`'[^']*'`)

// testSchemaEnumParityAllowlist maps an ENUM type name to a documented reason it is
// intentionally absent from sharedTestSchemaEnumTypes. Empty today — add entries only with a
// reason a future reader can verify (e.g. "no setupTestDB-based test AutoMigrates a model
// using this type").
var testSchemaEnumParityAllowlist = map[string]string{}

// extractSQLEnumTypes parses every CREATE TYPE ... AS ENUM statement out of migration SQL raw
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

// alterAddValue is one ALTER TYPE ... ADD VALUE literal plus its IF NOT EXISTS flag.
type alterAddValue struct {
	literal     string
	ifNotExists bool
}

// extractAlterTypeAddValues parses every ALTER TYPE ... ADD VALUE statement out of
// migration SQL raw text, returning name -> ordered list of added values.
// Per-file order is preserved; the caller concatenates across files in filename order.
func extractAlterTypeAddValues(sql string) map[string][]alterAddValue {
	found := make(map[string][]alterAddValue)
	for _, m := range alterTypeAddValueRe.FindAllStringSubmatch(sql, -1) {
		found[m[1]] = append(found[m[1]], alterAddValue{
			literal:     m[3],
			ifNotExists: m[2] != "",
		})
	}
	return found
}

// applyAlterTypeAddValues folds ALTER TYPE ... ADD VALUE literals into the extracted
// CREATE TYPE definitions (append-only, matching PostgreSQL enum ordering). An ALTER
// whose target type has no CREATE TYPE in migrations/*.sql is itself a drift and is
// reported as a violation. A plain ADD VALUE for a value the type already contains is
// also a violation — PostgreSQL rejects it — while IF NOT EXISTS on an existing value
// is the no-op PostgreSQL performs (001_init.sql's folded amendments rely on this).
func applyAlterTypeAddValues(sqlTypes map[string]string, alterAdds map[string][]alterAddValue) []string {
	var violations []string
	for name, values := range alterAdds {
		def, ok := sqlTypes[name]
		if !ok {
			literals := make([]string, 0, len(values))
			for _, v := range values {
				literals = append(literals, v.literal)
			}
			violations = append(violations, "ALTER TYPE "+name+" ADD VALUE "+strings.Join(literals, ", ")+
				" has no CREATE TYPE definition in migrations/*.sql — orphan enum amendment.")
			continue
		}
		present := map[string]struct{}{}
		for _, lit := range enumLiteralRe.FindAllString(def, -1) {
			present[lit] = struct{}{}
		}
		for _, v := range values {
			if _, exists := present[v.literal]; exists {
				if v.ifNotExists {
					continue
				}
				violations = append(violations, "ALTER TYPE "+name+" ADD VALUE "+v.literal+
					" re-adds a value already present in the CREATE TYPE definition — PostgreSQL would reject it.")
				continue
			}
			def = strings.TrimSuffix(def, ")") + ", " + v.literal + ")"
			present[v.literal] = struct{}{}
		}
		sqlTypes[name] = def
	}
	return violations
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
// migrations/*.sql) against goTypes (sharedTestSchemaEnumTypes) and returns a human-readable
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
			violations = append(violations, "ENUM type "+name+" exists in migrations/*.sql but is "+
				"missing from sharedTestSchemaEnumTypes (ltv_repository_test.go) — add it, or add "+
				"a documented exception to testSchemaEnumParityAllowlist. SQL definition: "+sqlDef)
		case goDef != sqlDef:
			violations = append(violations, "ENUM type "+name+" drifted: sharedTestSchemaEnumTypes "+
				"has \""+goDef+"\" but migrations/*.sql defines \""+sqlDef+"\" — sync the test double.")
		}
	}

	for name, goDef := range goTypes {
		if _, ok := sqlTypes[name]; !ok {
			violations = append(violations, "sharedTestSchemaEnumTypes has ENUM type "+name+
				" (\""+goDef+"\") that does not exist in migrations/*.sql — remove it (typo or stale "+
				"entry from a since-removed migration).")
		}
	}

	for name := range allowlist {
		if _, ok := sqlTypes[name]; !ok {
			violations = append(violations, "testSchemaEnumParityAllowlist entry "+name+" no "+
				"longer matches an existing ENUM type in migrations/*.sql — remove the stale entry.")
		}
	}

	return violations
}

// TestTestSchemaEnumParity is the gate: every ENUM type CREATE TYPE ... AS ENUM statement in
// migrations/*.sql must have a byte-for-byte matching entry in sharedTestSchemaEnumTypes, and
// vice versa (modulo documented exceptions). A floor guards against a vacuous pass if the
// extraction regex silently breaks.
func TestTestSchemaEnumParity(t *testing.T) {
	migrationsPath := "../../migrations" // B5b requires this relative path.
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	// os.ReadDir はファイル名ソート順を返すため、同名型が複数 migration に現れた場合は
	// 後の migration の定義が勝つ（CREATE TYPE の再定義は想定しないが、順序は決定論的）。
	sqlTypes := map[string]string{}
	alterAdds := map[string][]alterAddValue{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(migrationsPath, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		for name, def := range extractSQLEnumTypes(string(raw)) {
			sqlTypes[name] = def
		}
		// os.ReadDir のファイル名ソート順で ADD VALUE を収集する（同一型への複数
		// amendment は migration ファイル順＝PostgreSQL適用順のまま fold する）。
		for name, values := range extractAlterTypeAddValues(string(raw)) {
			alterAdds[name] = append(alterAdds[name], values...)
		}
	}
	for _, v := range applyAlterTypeAddValues(sqlTypes, alterAdds) {
		t.Error(v)
	}

	// 54 CREATE TYPE ... AS ENUM statements in 001_init.sql as of G12-2 (2026-07-10), plus
	// pet_neutered_status from 018_pet_neutered_status.sql. Revisit this floor if migrations
	// legitimately add/remove ENUM types.
	if len(sqlTypes) < 55 {
		t.Fatalf("extracted only %d CREATE TYPE ... AS ENUM statement(s) from migrations/*.sql; "+
			"expected 55+ — the extraction regex likely broke (would vacuously pass)", len(sqlTypes))
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
		if len(got) != 1 || !strings.Contains(got[0], "does not exist in migrations/*.sql") {
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

	t.Run("folds ALTER TYPE ADD VALUE into the CREATE TYPE definition", func(t *testing.T) {
		sqlTypes := map[string]string{
			"reservation_status": "CREATE TYPE reservation_status AS ENUM ('confirmed', 'pending', 'cancelled')",
		}
		adds := extractAlterTypeAddValues("ALTER TYPE reservation_status ADD VALUE IF NOT EXISTS 'tentative';")
		violations := applyAlterTypeAddValues(sqlTypes, adds)
		if len(violations) != 0 {
			t.Fatalf("expected 0 violations, got %v", violations)
		}
		want := "CREATE TYPE reservation_status AS ENUM ('confirmed', 'pending', 'cancelled', 'tentative')"
		if sqlTypes["reservation_status"] != want {
			t.Fatalf("got %q, want %q", sqlTypes["reservation_status"], want)
		}
	})

	t.Run("orphan ALTER TYPE ADD VALUE is a violation", func(t *testing.T) {
		sqlTypes := map[string]string{}
		adds := extractAlterTypeAddValues("ALTER TYPE ghost_type ADD VALUE 'x';")
		violations := applyAlterTypeAddValues(sqlTypes, adds)
		if len(violations) != 1 || !strings.Contains(violations[0], "no CREATE TYPE") {
			t.Fatalf("expected orphan amendment violation, got %v", violations)
		}
	})

	t.Run("ADD VALUE IF NOT EXISTS on an existing value is a no-op", func(t *testing.T) {
		// 001_init.sql's folded amendments: base CREATE TYPE already contains the value.
		sqlTypes := map[string]string{
			"lab_import_job_status": "CREATE TYPE lab_import_job_status AS ENUM ('failed', 'reverted')",
		}
		adds := extractAlterTypeAddValues("ALTER TYPE lab_import_job_status ADD VALUE IF NOT EXISTS 'reverted';")
		violations := applyAlterTypeAddValues(sqlTypes, adds)
		if len(violations) != 0 {
			t.Fatalf("expected no-op for IF NOT EXISTS re-add, got %v", violations)
		}
		if sqlTypes["lab_import_job_status"] != "CREATE TYPE lab_import_job_status AS ENUM ('failed', 'reverted')" {
			t.Fatalf("definition changed: %q", sqlTypes["lab_import_job_status"])
		}
	})

	t.Run("plain ADD VALUE on an existing value is a violation", func(t *testing.T) {
		sqlTypes := map[string]string{
			"reservation_status": "CREATE TYPE reservation_status AS ENUM ('confirmed', 'tentative')",
		}
		adds := extractAlterTypeAddValues("ALTER TYPE reservation_status ADD VALUE 'tentative';")
		violations := applyAlterTypeAddValues(sqlTypes, adds)
		if len(violations) != 1 || !strings.Contains(violations[0], "re-adds") {
			t.Fatalf("expected re-add violation, got %v", violations)
		}
	})
}
