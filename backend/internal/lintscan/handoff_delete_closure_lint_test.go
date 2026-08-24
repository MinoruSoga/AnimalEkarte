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

// scripts/import-old-db-handoffs-on-reset.sh clears the existing demo rows for a
// clinic before importing a staged old_db handoff. That DELETE block runs inside a
// single transaction with ON_ERROR_STOP=1, so one missing table aborts the whole
// `make reset`.
//
// 2026-08-24: `make reset` failed with
//
//	ERROR: update or delete on table "pets" violates RESTRICT setting of foreign key
//	constraint "hospitalizations_pet_id_fkey" on table "hospitalizations"
//
// because the block deleted pets without deleting hospitalizations first. Ten tables
// were missing at once, so fixing them one error at a time would have taken ten
// reset cycles. This gate derives the required set from the schema instead.

const (
	handoffDeleteScriptRelPath = "../scripts/import-old-db-handoffs-on-reset.sh"
	handoffDeleteMigrationFile = "migrations/001_init.sql"
)

type handoffRestrictEdge struct {
	child  string
	parent string
	column string
}

func TestHandoffDeleteBlock_IsClosedUnderRestrictForeignKeys(t *testing.T) {
	moduleRoot := mustModuleRoot(t)
	edges := mustLoadHandoffRestrictEdges(t, moduleRoot)
	deleted := mustLoadHandoffDeletedTables(t, moduleRoot)

	missing := findHandoffDeleteClosureGaps(edges, deleted)
	if len(missing) > 0 {
		t.Fatalf(
			"%s deletes rows whose RESTRICT children are never deleted; `make reset` will abort.\n"+
				"Add these tables to the DELETE block before their parents: %s",
			handoffDeleteScriptRelPath, strings.Join(missing, ", "),
		)
	}
}

func TestHandoffDeleteBlock_OrdersChildrenBeforeParents(t *testing.T) {
	moduleRoot := mustModuleRoot(t)
	edges := mustLoadHandoffRestrictEdges(t, moduleRoot)
	deleted := mustLoadHandoffDeletedTables(t, moduleRoot)

	inversions := findHandoffDeleteOrderInversions(edges, deleted)
	if len(inversions) > 0 {
		t.Fatalf(
			"%s deletes a parent before its RESTRICT child; `make reset` will abort.\n%s",
			handoffDeleteScriptRelPath, strings.Join(inversions, "\n"),
		)
	}
}

func TestFindHandoffDeleteClosureGaps_DetectsMissingChild(t *testing.T) {
	edges := []handoffRestrictEdge{{child: "hospitalizations", parent: "pets", column: "pet_id"}}
	missing := findHandoffDeleteClosureGaps(edges, []string{"pets"})
	if len(missing) != 1 || missing[0] != "hospitalizations" {
		t.Fatalf("expected hospitalizations to be reported missing, got %v", missing)
	}
}

func TestFindHandoffDeleteClosureGaps_FollowsTransitiveChildren(t *testing.T) {
	// prescriptions -> medical_records is only reachable once medical_records is
	// known to be deleted; a single-level check would miss it.
	edges := []handoffRestrictEdge{
		{child: "medical_record_addenda", parent: "medical_records", column: "medical_record_id"},
		{child: "medical_records", parent: "pets", column: "pet_id"},
	}
	missing := findHandoffDeleteClosureGaps(edges, []string{"pets"})
	want := []string{"medical_record_addenda", "medical_records"}
	if strings.Join(missing, ",") != strings.Join(want, ",") {
		t.Fatalf("expected transitive closure %v, got %v", want, missing)
	}
}

func TestFindHandoffDeleteOrderInversions_DetectsParentBeforeChild(t *testing.T) {
	edges := []handoffRestrictEdge{{child: "hospitalizations", parent: "pets", column: "pet_id"}}
	inversions := findHandoffDeleteOrderInversions(edges, []string{"pets", "hospitalizations"})
	if len(inversions) != 1 {
		t.Fatalf("expected one inversion, got %v", inversions)
	}
	if !strings.Contains(inversions[0], "hospitalizations") || !strings.Contains(inversions[0], "pets") {
		t.Fatalf("inversion message should name both tables, got %q", inversions[0])
	}
}

func TestFindHandoffDeleteOrderInversions_AllowsChildBeforeParent(t *testing.T) {
	edges := []handoffRestrictEdge{{child: "hospitalizations", parent: "pets", column: "pet_id"}}
	if inversions := findHandoffDeleteOrderInversions(edges, []string{"hospitalizations", "pets"}); len(inversions) != 0 {
		t.Fatalf("correct order must not be reported, got %v", inversions)
	}
}

func TestParseHandoffDeletedTables_KeepsScriptOrderAndIgnoresSubqueries(t *testing.T) {
	script := `BEGIN;
DELETE FROM billing_items WHERE billing_id IN (SELECT id FROM billings WHERE clinic_id = 1);
DELETE FROM billings WHERE clinic_id = 1;
-- DELETE FROM commented_out WHERE clinic_id = 1;
COMMIT;
`
	got := parseHandoffDeletedTables(script)
	want := []string{"billing_items", "billings"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

// findHandoffDeleteClosureGaps returns the tables that must also be deleted, in
// alphabetical order. A table belongs to the closure when it references any
// deleted table through an ON DELETE RESTRICT foreign key, transitively.
func findHandoffDeleteClosureGaps(edges []handoffRestrictEdge, deleted []string) []string {
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

// findHandoffDeleteOrderInversions reports RESTRICT edges whose child is deleted
// after its parent. Both tables must already be present in the delete list;
// absent tables are the closure check's responsibility.
func findHandoffDeleteOrderInversions(edges []handoffRestrictEdge, deleted []string) []string {
	position := make(map[string]int, len(deleted))
	for index, table := range deleted {
		if _, seen := position[table]; !seen {
			position[table] = index
		}
	}
	var inversions []string
	for _, edge := range edges {
		childAt, childListed := position[edge.child]
		parentAt, parentListed := position[edge.parent]
		if !childListed || !parentListed || childAt <= parentAt {
			continue
		}
		inversions = append(inversions, fmt.Sprintf(
			"  %s.%s REFERENCES %s(id) ON DELETE RESTRICT, but %s is deleted at line %d and %s only at line %d",
			edge.child, edge.column, edge.parent, edge.parent, parentAt+1, edge.child, childAt+1,
		))
	}
	sort.Strings(inversions)
	return inversions
}

var (
	handoffCreateTableRe = regexp.MustCompile(`(?s)CREATE TABLE (\w+)\s*\((.*?)\n\);`)
	handoffReferencesRe  = regexp.MustCompile(`REFERENCES\s+(\w+)\s*\(\s*id\s*\)`)
	handoffDeleteFromRe  = regexp.MustCompile(`^\s*DELETE FROM\s+(\w+)\b`)
)

// parseHandoffRestrictEdges collects every ON DELETE RESTRICT foreign key that
// targets a table's id column.
func parseHandoffRestrictEdges(schema string) []handoffRestrictEdge {
	var edges []handoffRestrictEdge
	for _, table := range handoffCreateTableRe.FindAllStringSubmatch(schema, -1) {
		child, body := table[1], table[2]
		for line := range strings.SplitSeq(body, "\n") {
			if !strings.Contains(line, "ON DELETE RESTRICT") {
				continue
			}
			reference := handoffReferencesRe.FindStringSubmatch(line)
			if reference == nil {
				continue
			}
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) == 0 {
				continue
			}
			edges = append(edges, handoffRestrictEdge{child: child, parent: reference[1], column: fields[0]})
		}
	}
	return edges
}

// parseHandoffDeletedTables returns the DELETE targets in script order. Tables
// named only inside a subquery are not deleted and are skipped.
func parseHandoffDeletedTables(script string) []string {
	var tables []string
	for line := range strings.SplitSeq(script, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		if match := handoffDeleteFromRe.FindStringSubmatch(line); match != nil {
			tables = append(tables, match[1])
		}
	}
	return tables
}

func mustLoadHandoffRestrictEdges(t *testing.T, moduleRoot string) []handoffRestrictEdge {
	t.Helper()
	schema, err := os.ReadFile(filepath.Join(moduleRoot, handoffDeleteMigrationFile))
	if err != nil {
		t.Fatalf("read %s: %v", handoffDeleteMigrationFile, err)
	}
	edges := parseHandoffRestrictEdges(string(schema))
	if len(edges) == 0 {
		t.Fatalf("no ON DELETE RESTRICT foreign keys parsed from %s; the parser is broken", handoffDeleteMigrationFile)
	}
	return edges
}

func mustLoadHandoffDeletedTables(t *testing.T, moduleRoot string) []string {
	t.Helper()
	script, err := os.ReadFile(filepath.Join(moduleRoot, handoffDeleteScriptRelPath))
	if err != nil {
		t.Fatalf("read %s: %v", handoffDeleteScriptRelPath, err)
	}
	tables := parseHandoffDeletedTables(string(script))
	if len(tables) == 0 {
		t.Fatalf("no DELETE FROM statements parsed from %s; the parser is broken", handoffDeleteScriptRelPath)
	}
	return tables
}
