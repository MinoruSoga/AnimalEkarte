package main

import "testing"

// assertPanics fails the test unless fn panics — used for the fail-closed
// "unknown table" defaults in tables.go, which must never silently no-op on a
// programmer error (a new table added to one list but not another).
func assertPanics(t *testing.T, label string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s: expected a panic for an unknown table, got none", label)
		}
	}()
	fn()
}

func TestDeleteScopeUnknownTablePanics(t *testing.T) {
	assertPanics(t, "deleteScope", func() { deleteScope("not_a_real_table") })
}

func TestSelectSQLUnknownTablePanics(t *testing.T) {
	assertPanics(t, "selectSQL", func() { selectSQL("not_a_real_table", "") })
}

func TestInsertColumnsUnknownTablePanics(t *testing.T) {
	assertPanics(t, "insertColumns", func() { insertColumns("not_a_real_table") })
}

func TestSelectArgsUnknownTablePanics(t *testing.T) {
	assertPanics(t, "selectArgs", func() { selectArgs("not_a_real_table", targetRefs{}) })
}

// --- prefixRunID: correlated-subquery alias prefixing ---

func TestPrefixRunIDEmptyRunIDYieldsNoClause(t *testing.T) {
	if got := prefixRunID("p", ""); got != "" {
		t.Fatalf("prefixRunID with empty run-id must yield no clause, got %q", got)
	}
}

func TestPrefixRunIDPrefixesAliasAndEscapes(t *testing.T) {
	got := prefixRunID("p", "run'42")
	want := " AND p.migration_run_id = 'run''42'"
	if got != want {
		t.Fatalf("prefixRunID(%q, %q) = %q, want %q", "p", "run'42", got, want)
	}
}

// --- insertColumns / selectArgs: every table in importOrder resolves ---

func TestInsertColumnsResolvesEveryImportOrderTable(t *testing.T) {
	for _, tbl := range importOrder {
		cols := insertColumns(tbl)
		if len(cols) == 0 {
			t.Errorf("insertColumns(%q) returned no columns", tbl)
		}
	}
}

func TestSelectArgsResolvesEveryImportOrderTable(t *testing.T) {
	refs := targetRefs{clinicID: 1, fallbackAnimalSpeciesID: 2, fallbackExamTypeID: 3}
	for _, tbl := range importOrder {
		// selectArgs must not panic for any real importOrder table.
		_ = selectArgs(tbl, refs)
	}
}
