package lintscan

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// DeleteSyntheticClosingFixture (internal/billing/synthetic_closing_fixture.go)
// tears down an S09 synthetic clinic inside one transaction. The UAT bug EMR-210
// was a 500 caused by a scoped delete list that was incomplete relative to the
// real FK graph: cash_register_close_adjustments (RESTRICT → billings /
// cash_register_closes / staffs) was missing, so deleting billings aborted the
// whole teardown.
//
// These gates derive the blocking-FK closure straight from migrations/*.sql —
// including ALTER TABLE composite foreign keys, which the handoff gate's
// inline-only parser cannot see — and require the fixture's delete series to
// cover it, children before parents. audit_logs is the single documented
// exemption: its handling is an EMR-211 product decision, so teardown stays
// fail-closed on audit rows by default.
//
// exams ↔ examination_revisions is a genuine RESTRICT NOT DEFERRABLE cycle
// (fk_exams_current_revision points at the revision row). The fixture breaks it
// with UPDATE exams SET current_revision_version = NULL before deleting, so
// that single edge is exempt from the ordering check — and the UPDATE itself is
// pinned so it cannot be dropped silently.

const syntheticClosingFixtureRelPath = "../internal/billing/synthetic_closing_fixture.go"

// teardownTailRoots are the tables deleted by Go code after the literal DELETE
// series, in execution order. The scoped list plus these roots is the
// "deleted" set the closure runs over.
var teardownTailRoots = []string{
	"staffs",         // staff.UnscopedDeleteSyntheticClosingStaffs
	"accounts",       // by collected staff account ids
	"animal_species", // by s09-species-<clinicID> name
	"clinics",        // the clinic row itself
	"companies",      // the synthetic company
}

// teardownClosureExemptions are blocking children that teardown deliberately
// does NOT delete. audit_logs is preserved pending the EMR-211
// delete-vs-anonymize product decision — its RESTRICT FKs then keep teardown
// fail-closed, which is the accepted BLOCKED behavior, not a gap to fix here.
var teardownClosureExemptions = map[string]bool{
	"audit_logs": true,
}

// teardownBrokenCycleEdges are blocking edges the fixture resolves without a
// child-before-parent delete (the parent pointer is nulled first).
var teardownBrokenCycleEdges = map[string]bool{
	"exams->examination_revisions": true,
}

type teardownFKEdge struct {
	child      string
	parent     string
	column     string
	constraint string
}

func TestSyntheticClosingTeardown_IsClosedUnderBlockingFKs(t *testing.T) {
	moduleRoot := mustModuleRoot(t)
	edges := mustLoadTeardownBlockingEdges(t, moduleRoot)
	deleted := mustLoadTeardownDeletedTables(t, moduleRoot)

	missing := findTeardownClosureGaps(edges, deleted)
	kept := missing[:0]
	for _, table := range missing {
		if teardownClosureExemptions[table] {
			continue
		}
		kept = append(kept, table)
	}
	if len(kept) > 0 {
		t.Fatalf(
			"%s deletes rows whose blocking-FK children are never deleted; the UAT teardown will 500.\n"+
				"Add these tables to syntheticClosingDeleteStatements before their parents: %s",
			syntheticClosingFixtureRelPath, strings.Join(kept, ", "),
		)
	}
}

func TestSyntheticClosingTeardown_OrdersChildrenBeforeParents(t *testing.T) {
	moduleRoot := mustModuleRoot(t)
	edges := mustLoadTeardownBlockingEdges(t, moduleRoot)
	deleted := mustLoadTeardownDeletedTables(t, moduleRoot)

	var inversions []string
	for _, edge := range edges {
		if teardownBrokenCycleEdges[edge.child+"->"+edge.parent] {
			continue
		}
		childAt, childListed := indexOf(deleted, edge.child)
		parentAt, parentListed := indexOf(deleted, edge.parent)
		if !childListed || !parentListed || childAt <= parentAt {
			continue
		}
		inversions = append(inversions, fmt.Sprintf(
			"  %s.%s REFERENCES %s(id) blocks the parent delete, but %s is deleted at step %d and %s only at step %d",
			edge.child, edge.column, edge.parent, edge.parent, parentAt+1, edge.child, childAt+1,
		))
	}
	if len(inversions) > 0 {
		sort.Strings(inversions)
		t.Fatalf(
			"%s deletes a parent before its blocking-FK child; the UAT teardown will 500.\n%s",
			syntheticClosingFixtureRelPath, strings.Join(inversions, "\n"),
		)
	}
}

func TestSyntheticClosingTeardown_BreaksExamRevisionCycle(t *testing.T) {
	moduleRoot := mustModuleRoot(t)
	source := mustReadTeardownFixtureSource(t, moduleRoot)

	// Without clearing exams.current_revision_version first, exams and
	// examination_revisions RESTRICT each other and neither can be deleted.
	if !strings.Contains(source, "UPDATE exams SET current_revision_version = NULL WHERE clinic_id = ?") {
		t.Fatalf(
			"%s must clear exams.current_revision_version before deleting exams/"+
				"examination_revisions — the two tables hold a mutual RESTRICT NOT DEFERRABLE cycle",
			syntheticClosingFixtureRelPath,
		)
	}
}

func TestSyntheticClosingTeardown_DisablesEveryAppendOnlyTable(t *testing.T) {
	moduleRoot := mustModuleRoot(t)
	appendOnly := mustLoadAppendOnlyTables(t, moduleRoot)
	source := mustReadTeardownFixtureSource(t, moduleRoot)
	disabled := parseTeardownStringSlice(source, "syntheticClosingAppendOnlyTables")
	deleted := mustLoadTeardownDeletedTables(t, moduleRoot)

	for table := range appendOnly {
		if !disabled[table] {
			t.Errorf("append-only table %s is missing from syntheticClosingAppendOnlyTables; its DELETE trigger will abort teardown", table)
		}
		if _, ok := indexOf(deleted, table); !ok {
			t.Errorf("append-only table %s is never deleted", table)
		}
	}
	for table := range disabled {
		if !appendOnly[table] {
			t.Errorf("syntheticClosingAppendOnlyTables lists %s but no DELETE-blocking trigger exists for it in migrations/", table)
		}
	}
}

func TestSyntheticClosingTeardown_PreservesAuditLogs(t *testing.T) {
	moduleRoot := mustModuleRoot(t)
	source := mustReadTeardownFixtureSource(t, moduleRoot)
	for line := range strings.Lines(source) {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if regexp.MustCompile(`DELETE FROM\s+audit_logs\b`).MatchString(trimmed) {
			t.Fatalf("audit_logs must not be physically deleted by teardown — EMR-211 owns that decision: %s", trimmed)
		}
	}
}

// --- helpers ----------------------------------------------------------------

func indexOf(list []string, item string) (int, bool) {
	for i, v := range list {
		if v == item {
			return i, true
		}
	}
	return -1, false
}

func findTeardownClosureGaps(edges []teardownFKEdge, deleted []string) []string {
	closure := make(map[string]bool, len(deleted))
	for _, table := range deleted {
		closure[table] = true
	}
	for grown := true; grown; {
		grown = false
		for _, edge := range edges {
			if closure[edge.parent] && !closure[edge.child] {
				closure[edge.child] = true
				grown = true
			}
		}
	}
	seed := make(map[string]bool, len(deleted))
	for _, table := range deleted {
		seed[table] = true
	}
	missing := make([]string, 0, len(closure))
	for table := range closure {
		if !seed[table] {
			missing = append(missing, table)
		}
	}
	sort.Strings(missing)
	return missing
}

var (
	teardownCreateTableRe = regexp.MustCompile(`(?s)CREATE TABLE(?: IF NOT EXISTS)? (\w+)\s*\((.*?)\n\);`)
	teardownAlterTableRe  = regexp.MustCompile(`(?s)ALTER TABLE (\w+)\s+(.*?);`)
	teardownInlineRefRe   = regexp.MustCompile(`(?i)REFERENCES\s+(\w+)\s*\(\s*id\s*\)`)
	teardownCompositeFKRe = regexp.MustCompile(`(?is)(?:CONSTRAINT\s+(\w+)\s+)?FOREIGN KEY\s*\(([\w,\s]+)\)\s*REFERENCES\s+(\w+)\s*\(([\w,\s]+)\)\s*(ON DELETE\s+\w+(?:\s+\w+)?)?`)
	teardownConstraintRe  = regexp.MustCompile(`(?i)CONSTRAINT\s+(\w+)\s+`)
	teardownDropRe        = regexp.MustCompile(`(?i)DROP CONSTRAINT(?: IF EXISTS)?\s+(\w+)`)
	teardownNonBlockingRe = regexp.MustCompile(`(?i)ON DELETE\s+(CASCADE|SET NULL|SET DEFAULT)`)
	teardownDeleteFromRe  = regexp.MustCompile(`DELETE FROM\s+(\w+)\b`)
	teardownStringSliceRe = func(name string) *regexp.Regexp {
		return regexp.MustCompile(`(?s)` + name + `\s*=\s*\[\]string\s*\{(.*?)\}`)
	}
	teardownQuotedRe = regexp.MustCompile(`"(\w+)"`)
	// A BEFORE UPDATE OR DELETE trigger firing a prevent_*_mutation function is
	// the append-only ledger pattern — teardown must disable it first.
	teardownAppendOnlyRe = regexp.MustCompile(`(?is)BEFORE UPDATE OR DELETE ON (\w+)`)
)

// parseTeardownBlockingEdges collects every foreign key that blocks deleting the
// referenced row: ON DELETE RESTRICT, ON DELETE NO ACTION, and omitted clauses
// (which default to NO ACTION). Unlike the handoff gate's parser this also sees
// ALTER TABLE composite foreign keys — the cash_register_close_adjustments
// tenant-correlated FKs that EMR-210 tripped on live exclusively in that form.
// DROP CONSTRAINT entries remove edges so re-shaped constraints do not
// double-count.
func parseTeardownBlockingEdges(schema string) []teardownFKEdge {
	var edges []teardownFKEdge
	dropped := make(map[string]bool)

	addEdge := func(child, parent, column, constraint, onDelete string) {
		if teardownNonBlockingRe.MatchString(onDelete) {
			return
		}
		if constraint == "" {
			constraint = fmt.Sprintf("%s_%s_fkey", child, column)
		}
		edges = append(edges, teardownFKEdge{
			child: child, parent: parent, column: column, constraint: constraint,
		})
	}

	for _, table := range teardownCreateTableRe.FindAllStringSubmatch(schema, -1) {
		child, body := table[1], table[2]
		for line := range strings.Lines(body) {
			reference := teardownInlineRefRe.FindStringSubmatchIndex(line)
			if reference == nil {
				continue
			}
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) == 0 {
				continue
			}
			col := fields[0]
			if col == "FOREIGN" || col == "CONSTRAINT" {
				continue // table-level clause — handled by the composite pass below
			}
			onDelete := line[reference[1]:]
			if teardownNonBlockingRe.MatchString(onDelete) {
				continue
			}
			parent := line[reference[2]:reference[3]]
			constraint := ""
			if cm := teardownConstraintRe.FindStringSubmatch(line); cm != nil {
				constraint = cm[1]
			}
			edges = append(edges, teardownFKEdge{child: child, parent: parent, column: col, constraint: constraint})
		}
		for _, fk := range teardownCompositeFKRe.FindAllStringSubmatch(body, -1) {
			firstCol := strings.TrimSpace(strings.Split(fk[2], ",")[0])
			addEdge(child, fk[3], firstCol, fk[1], fk[5])
		}
	}

	for _, alter := range teardownAlterTableRe.FindAllStringSubmatch(schema, -1) {
		child, segment := alter[1], alter[2]
		for _, d := range teardownDropRe.FindAllStringSubmatch(segment, -1) {
			dropped[child+"."+d[1]] = true
		}
		for _, fk := range teardownCompositeFKRe.FindAllStringSubmatch(segment, -1) {
			firstCol := strings.TrimSpace(strings.Split(fk[2], ",")[0])
			addEdge(child, fk[3], firstCol, fk[1], fk[5])
		}
	}

	if len(dropped) > 0 {
		kept := edges[:0]
		for _, edge := range edges {
			if dropped[edge.child+"."+edge.constraint] {
				continue
			}
			kept = append(kept, edge)
		}
		edges = kept
	}
	return edges
}

// parseTeardownDeletedTables extracts the fixture's delete order: every
// "DELETE FROM <table>" literal in syntheticClosingDeleteStatements order,
// then the Go-side tail (staffs → accounts → species → clinic → company) plus
// the audit_logs exemption entry so closure treats it as handled.
func parseTeardownDeletedTables(source string) []string {
	var tables []string
	for line := range strings.Lines(source) {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if match := teardownDeleteFromRe.FindStringSubmatch(trimmed); match != nil {
			tables = append(tables, match[1])
		}
	}
	return append(tables, teardownTailRoots...)
}

// parseTeardownStringSlice extracts quoted identifiers from a `var <name> =
// []string{ ... }` literal in Go source.
func parseTeardownStringSlice(source, name string) map[string]bool {
	out := make(map[string]bool)
	block := teardownStringSliceRe(name).FindStringSubmatch(source)
	if block == nil {
		return out
	}
	for _, q := range teardownQuotedRe.FindAllStringSubmatch(block[1], -1) {
		out[q[1]] = true
	}
	return out
}

func mustReadTeardownFixtureSource(t *testing.T, moduleRoot string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(moduleRoot, "internal/billing/synthetic_closing_fixture.go"))
	if err != nil {
		t.Fatalf("read fixture source: %v", err)
	}
	return string(raw)
}

func mustLoadTeardownDeletedTables(t *testing.T, moduleRoot string) []string {
	t.Helper()
	tables := parseTeardownDeletedTables(mustReadTeardownFixtureSource(t, moduleRoot))
	if len(tables) <= len(teardownTailRoots) {
		t.Fatalf("no DELETE FROM statements parsed from %s; the parser is broken", syntheticClosingFixtureRelPath)
	}
	return tables
}

func mustLoadTeardownBlockingEdges(t *testing.T, moduleRoot string) []teardownFKEdge {
	t.Helper()
	dir := filepath.Join(moduleRoot, "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}
	var schema strings.Builder
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("read migration %s: %v", entry.Name(), err)
		}
		schema.Write(raw)
		schema.WriteString("\n")
	}
	edges := parseTeardownBlockingEdges(schema.String())
	if len(edges) == 0 {
		t.Fatal("no blocking foreign keys parsed from migrations; the parser is broken")
	}
	return edges
}

func mustLoadAppendOnlyTables(t *testing.T, moduleRoot string) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(moduleRoot, "migrations/001_init.sql"))
	if err != nil {
		t.Fatalf("read 001_init.sql: %v", err)
	}
	out := make(map[string]bool)
	for _, m := range teardownAppendOnlyRe.FindAllStringSubmatch(string(raw), -1) {
		out[m[1]] = true
	}
	if len(out) == 0 {
		t.Fatal("no append-only BEFORE UPDATE OR DELETE triggers parsed; the parser is broken")
	}
	return out
}

// --- parser self-tests --------------------------------------------------------

func TestParseTeardownBlockingEdges_SeesAlterTableCompositeFKs(t *testing.T) {
	schema := `CREATE TABLE staffs (id bigint PRIMARY KEY, clinic_id bigint NOT NULL);

CREATE TABLE billings (id bigint PRIMARY KEY, clinic_id bigint NOT NULL);

CREATE TABLE cash_register_close_adjustments (
    id BIGSERIAL PRIMARY KEY,
    clinic_id BIGINT NOT NULL REFERENCES clinics(id),
    billing_id BIGINT NOT NULL
);

ALTER TABLE cash_register_close_adjustments
    ADD CONSTRAINT fk_cash_register_close_adjustments_billing_clinic
        FOREIGN KEY (billing_id, clinic_id)
        REFERENCES billings (id, clinic_id)
        ON DELETE RESTRICT;
`
	var found bool
	for _, edge := range parseTeardownBlockingEdges(schema) {
		if edge.child == "cash_register_close_adjustments" && edge.parent == "billings" {
			found = true
		}
	}
	if !found {
		t.Fatal("ALTER TABLE composite FK to billings was not parsed — this is exactly the EMR-210 blind spot")
	}
}

func TestParseTeardownBlockingEdges_DroppedConstraintsDoNotBlock(t *testing.T) {
	schema := `CREATE TABLE owners (id bigint PRIMARY KEY);

CREATE TABLE pets (
    id bigint PRIMARY KEY,
    owner_id bigint REFERENCES owners(id)
);

ALTER TABLE pets DROP CONSTRAINT pets_owner_id_fkey;
`
	for _, edge := range parseTeardownBlockingEdges(schema) {
		if edge.child == "pets" && edge.parent == "owners" {
			t.Fatal("dropped pets_owner_id_fkey still reported as blocking")
		}
	}
}

func TestParseTeardownDeletedTables_KeepsOrderAndAddsTail(t *testing.T) {
	source := `var syntheticClosingDeleteStatements = []string{
	"DELETE FROM payment_splits WHERE clinic_id = ?",
	"DELETE FROM billings WHERE clinic_id = ?",
	// "DELETE FROM commented_out WHERE clinic_id = ?",
}`
	got := parseTeardownDeletedTables(source)
	if len(got) < 3 || got[0] != "payment_splits" || got[1] != "billings" {
		t.Fatalf("expected ordered deletes then tail roots, got %v", got)
	}
	if got[len(got)-1] != teardownTailRoots[len(teardownTailRoots)-1] {
		t.Fatalf("tail roots must be appended in order, got %v", got[len(got)-1])
	}
}
