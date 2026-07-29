package csvimport

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const testDisposableDatabase = "seed_export_tmp"

func TestImportWithBeginReplacesAndImportsLegacyGraph(t *testing.T) {
	sourceDir := writeLegacyImportFixture(t)
	tx := &fakeLegacyTransaction{currentDatabase: testDisposableDatabase}

	counts, err := importWithBegin(
		context.Background(),
		func(context.Context) (legacyImportTransaction, error) { return tx, nil },
		sourceDir,
		testDisposableDatabase,
		1,
		2,
		3,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !tx.committed || tx.copyCalls != len(tables) || tx.deletedTables == 0 {
		t.Fatalf("committed=%v copyCalls=%d deletedTables=%d", tx.committed, tx.copyCalls, tx.deletedTables)
	}
	for _, table := range tables {
		if counts[table] != 1 {
			t.Fatalf("table %s count = %d, want 1", table, counts[table])
		}
	}
}

func TestImportRejectsRelativeSourceBeforeOpeningTransaction(t *testing.T) {
	counts, err := Import(context.Background(), nil, "relative/source", testDisposableDatabase, 1, 2, 3)
	if err == nil || counts != nil {
		t.Fatalf("Import() = (%v, %v), want absolute-path rejection", counts, err)
	}
}

func TestImportRejectsEmptyExpectedDatabase(t *testing.T) {
	tx := &fakeLegacyTransaction{currentDatabase: testDisposableDatabase}
	_, err := importWithBegin(
		context.Background(),
		func(context.Context) (legacyImportTransaction, error) { return tx, nil },
		t.TempDir(),
		"  ",
		1, 2, 3,
	)
	if err == nil || !strings.Contains(err.Error(), "expected disposable database name is required") {
		t.Fatalf("error = %v, want required expected database", err)
	}
	if tx.deletedTables != 0 {
		t.Fatalf("deletedTables=%d, want 0 before identity check", tx.deletedTables)
	}
}

func TestImportRejectsDatabaseIdentityMismatchWithoutDelete(t *testing.T) {
	tx := &fakeLegacyTransaction{currentDatabase: "ekarte_db"}
	_, err := importWithBegin(
		context.Background(),
		func(context.Context) (legacyImportTransaction, error) { return tx, nil },
		writeLegacyImportFixture(t),
		testDisposableDatabase,
		1, 2, 3,
	)
	if err == nil || !strings.Contains(err.Error(), "target database identity mismatch") {
		t.Fatalf("error = %v, want identity mismatch", err)
	}
	if tx.deletedTables != 0 {
		t.Fatalf("deletedTables=%d, want 0 on mismatch", tx.deletedTables)
	}
	if tx.committed {
		t.Fatal("must not commit after identity mismatch")
	}
}

func TestImportWithBeginReportsBeginFailure(t *testing.T) {
	const privateValue = "private-database-detail"
	_, err := importWithBegin(
		context.Background(),
		func(context.Context) (legacyImportTransaction, error) {
			return nil, fmt.Errorf("%s", privateValue)
		},
		t.TempDir(),
		testDisposableDatabase,
		1,
		2,
		3,
	)
	if err == nil || !strings.Contains(err.Error(), "begin import tx") || strings.Contains(err.Error(), privateValue) {
		t.Fatalf("error = %v, want begin failure", err)
	}
}

func TestLegacyImportErrorsDoNotExposeSourceValues(t *testing.T) {
	const privateValue = "private-owner-name"
	for _, test := range []struct {
		column string
		value  string
	}{
		{column: "id", value: privateValue},
		{column: "weight", value: privateValue},
		{column: "is_dangerous", value: privateValue},
	} {
		t.Run(test.column, func(t *testing.T) {
			_, err := typedValue(test.column, test.value)
			if err == nil || strings.Contains(err.Error(), privateValue) {
				t.Fatalf("error leaked private value: %v", err)
			}
		})
	}

	tx := &fakeLegacyTransaction{currentDatabase: testDisposableDatabase, copyError: errors.New(privateValue)}
	_, err := importWithBegin(
		context.Background(),
		func(context.Context) (legacyImportTransaction, error) { return tx, nil },
		writeLegacyImportFixture(t),
		testDisposableDatabase,
		1,
		2,
		3,
	)
	if err == nil || strings.Contains(err.Error(), privateValue) {
		t.Fatalf("COPY error leaked private value: %v", err)
	}

	for _, test := range []struct {
		name string
		tx   *fakeLegacyTransaction
	}{
		{name: "delete", tx: &fakeLegacyTransaction{currentDatabase: testDisposableDatabase, execError: errors.New(privateValue)}},
		{name: "commit", tx: &fakeLegacyTransaction{currentDatabase: testDisposableDatabase, commitError: errors.New(privateValue)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := importWithBegin(
				context.Background(),
				func(context.Context) (legacyImportTransaction, error) { return test.tx, nil },
				writeLegacyImportFixture(t),
				testDisposableDatabase,
				1,
				2,
				3,
			)
			if err == nil || strings.Contains(err.Error(), privateValue) {
				t.Fatalf("%s error leaked private value: %v", test.name, err)
			}
		})
	}
}

func writeLegacyImportFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, table := range tables {
		tableSpec, err := tableSpec(table)
		if err != nil {
			t.Fatal(err)
		}
		file, err := os.OpenFile(filepath.Join(dir, table+".csv"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			t.Fatal(err)
		}
		writer := csv.NewWriter(file)
		if err := writer.Write(tableSpec.columns); err != nil {
			t.Fatal(err)
		}
		row := make([]string, len(tableSpec.columns))
		for i, column := range tableSpec.columns {
			switch column {
			case "clinic_id":
				row[i] = "{{CLINIC_ID}}"
			case "animal_species_id":
				row[i] = "{{FALLBACK_ANIMAL_SPECIES_ID}}"
			case "exam_type_id":
				row[i] = "{{FALLBACK_EXAM_TYPE_ID}}"
			case "id", "owner_id", "pet_id", "medical_record_id", "exam_id", "billing_id", "total_amount", "unit_price", "sort_order":
				row[i] = "1"
			case "weight", "quantity", "discount_rate":
				row[i] = "1.5"
			case "is_dangerous", "is_insurance_applicable":
				row[i] = "false"
			default:
				row[i] = "fixture"
			}
		}
		if err := writer.Write(row); err != nil {
			t.Fatal(err)
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

type fakeLegacyTransaction struct {
	committed       bool
	rolledBack      bool
	copyCalls       int
	deletedTables   int
	currentDatabase string
	copyError       error
	execError       error
	commitError     error
	queryError      error
}

func (tx *fakeLegacyTransaction) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	if tx.execError != nil {
		return pgconn.CommandTag{}, tx.execError
	}
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sql)), "DELETE") {
		tx.deletedTables++
	}
	return pgconn.NewCommandTag("DELETE 1"), nil
}

func (tx *fakeLegacyTransaction) QueryRow(_ context.Context, sql string, _ ...any) pgx.Row {
	return fakeLegacyRow{tx: tx, sql: sql}
}

type fakeLegacyRow struct {
	tx  *fakeLegacyTransaction
	sql string
}

func (r fakeLegacyRow) Scan(dest ...any) error {
	if r.tx.queryError != nil {
		return r.tx.queryError
	}
	if !strings.Contains(strings.ToLower(r.sql), "current_database") {
		return fmt.Errorf("unexpected QueryRow sql")
	}
	if len(dest) != 1 {
		return fmt.Errorf("unexpected Scan arity")
	}
	ptr, ok := dest[0].(*string)
	if !ok {
		return fmt.Errorf("unexpected Scan type")
	}
	*ptr = r.tx.currentDatabase
	return nil
}

func (tx *fakeLegacyTransaction) CopyFrom(_ context.Context, _ pgx.Identifier, _ []string, source pgx.CopyFromSource) (int64, error) {
	if tx.copyError != nil {
		return 0, tx.copyError
	}
	var count int64
	for source.Next() {
		if _, err := source.Values(); err != nil {
			return 0, err
		}
		count++
	}
	if err := source.Err(); err != nil {
		return 0, err
	}
	tx.copyCalls++
	return count, nil
}

func (tx *fakeLegacyTransaction) Commit(context.Context) error {
	if tx.commitError != nil {
		return tx.commitError
	}
	tx.committed = true
	return nil
}

func (tx *fakeLegacyTransaction) Rollback(context.Context) error {
	tx.rolledBack = true
	return nil
}

func TestTableSpecResolvesTargetOnlyReferences(t *testing.T) {
	s, err := tableSpec("pets")
	if err != nil {
		t.Fatal(err)
	}
	row := make([]string, len(s.columns))
	row[0], row[1], row[2], row[5] = "10", "{{CLINIC_ID}}", "20", ""
	values, err := s.values(row, 1, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	if values[1] != int64(1) {
		t.Fatalf("clinic placeholder was not resolved: %#v", values[1])
	}
	if values[5] != int64(2) {
		t.Fatalf("species fallback was not resolved: %#v", values[5])
	}
}

func TestTypedValueConvertsBoundaryTypes(t *testing.T) {
	if got, err := typedValue("id", "42"); err != nil || got != int64(42) {
		t.Fatalf("id conversion: %#v %v", got, err)
	}
	if got, err := typedValue("is_dangerous", "t"); err != nil || got != true {
		t.Fatalf("bool conversion: %#v %v", got, err)
	}
}
