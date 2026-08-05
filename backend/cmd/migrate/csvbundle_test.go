package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/animal-ekarte/backend/internal/dbconn"
)

func writeBundleFixture(t *testing.T, migrationsDir, bundleDir, csvContent string) {
	t.Helper()
	dir := filepath.Join(migrationsDir, "seeds", bundleDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create fixture dir: %v", err)
	}
	manifest := `{"bundle":"` + bundleDir + `","tables":[{"table":"widgets","csvFile":"widgets.csv"}]}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("failed to write manifest fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "widgets.csv"), []byte(csvContent), 0o644); err != nil {
		t.Fatalf("failed to write csv fixture: %v", err)
	}
}

func TestBundleChecksumChangesWhenCSVContentChanges(t *testing.T) {
	dir := t.TempDir()

	writeBundleFixture(t, dir, "002_master", "id,name\n1,widget-a\n")
	sum1, err := bundleChecksum(dir, "002_master")
	if err != nil {
		t.Fatalf("bundleChecksum (first) returned error: %v", err)
	}

	// Same manifest, but the CSV data changed — checksum must change too, since
	// the CSV is the actual seed content (there is no stub SQL body anymore).
	writeBundleFixture(t, dir, "002_master", "id,name\n1,widget-b\n")
	sum2, err := bundleChecksum(dir, "002_master")
	if err != nil {
		t.Fatalf("bundleChecksum (second) returned error: %v", err)
	}

	if sum1 == sum2 {
		t.Fatalf("bundleChecksum did not change after CSV content changed (both = %q) — a CSV-only edit would be silently skipped on an already-migrated DB", sum1)
	}
}

func TestBundleChecksumChangesWhenManifestChanges(t *testing.T) {
	dir := t.TempDir()
	seedDir := filepath.Join(dir, "seeds", "003_demo")
	if err := os.MkdirAll(seedDir, 0o755); err != nil {
		t.Fatalf("failed to create fixture dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(seedDir, "a.csv"), []byte("id,name\n1,a\n"), 0o644); err != nil {
		t.Fatalf("failed to write csv fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(seedDir, "b.csv"), []byte("id,name\n1,b\n"), 0o644); err != nil {
		t.Fatalf("failed to write csv fixture: %v", err)
	}

	manifestForward := `{"bundle":"003_demo","tables":[{"table":"a","csvFile":"a.csv"},{"table":"b","csvFile":"b.csv"}]}`
	if err := os.WriteFile(filepath.Join(seedDir, "manifest.json"), []byte(manifestForward), 0o644); err != nil {
		t.Fatalf("failed to write manifest fixture: %v", err)
	}
	sum1, err := bundleChecksum(dir, "003_demo")
	if err != nil {
		t.Fatalf("bundleChecksum (forward order) returned error: %v", err)
	}

	// Same CSV files, reordered manifest — the checksum must still change,
	// since manifest.json's own bytes are folded into the hash.
	manifestReordered := `{"bundle":"003_demo","tables":[{"table":"b","csvFile":"b.csv"},{"table":"a","csvFile":"a.csv"}]}`
	if err := os.WriteFile(filepath.Join(seedDir, "manifest.json"), []byte(manifestReordered), 0o644); err != nil {
		t.Fatalf("failed to write reordered manifest fixture: %v", err)
	}
	sum2, err := bundleChecksum(dir, "003_demo")
	if err != nil {
		t.Fatalf("bundleChecksum (reordered) returned error: %v", err)
	}

	if sum1 == sum2 {
		t.Fatalf("bundleChecksum did not change after manifest.json table order changed (both = %q)", sum1)
	}
}

func TestBundleChecksumStableWhenNothingChanges(t *testing.T) {
	dir := t.TempDir()
	writeBundleFixture(t, dir, "003_demo", "id,name\n1,widget-a\n")

	sum1, err := bundleChecksum(dir, "003_demo")
	if err != nil {
		t.Fatalf("bundleChecksum (first) returned error: %v", err)
	}
	sum2, err := bundleChecksum(dir, "003_demo")
	if err != nil {
		t.Fatalf("bundleChecksum (second) returned error: %v", err)
	}
	if sum1 != sum2 {
		t.Fatalf("bundleChecksum is not deterministic for identical inputs: %q vs %q", sum1, sum2)
	}
}

func TestBundleChecksumMissingManifestErrors(t *testing.T) {
	dir := t.TempDir() // no seeds/ subdirectory created
	_, err := bundleChecksum(dir, "004_staging")
	if err == nil {
		t.Fatal("expected an error when a bundle's manifest.json is missing, got nil")
	}
}

func TestBuildCopyFromCSVSQL_ForceNotNullQuoting(t *testing.T) {
	got := buildCopyFromCSVSQL("hospitalizations", []string{"owner_request", `weird"col`})
	want := `COPY "hospitalizations" FROM STDIN WITH (FORMAT csv, HEADER true, FORCE_NOT_NULL ("owner_request", "weird""col"))`
	if got != want {
		t.Fatalf("buildCopyFromCSVSQL = %q, want %q", got, want)
	}
	plain := buildCopyFromCSVSQL("widgets", nil)
	if strings.Contains(plain, "FORCE_NOT_NULL") {
		t.Fatalf("empty force list must omit FORCE_NOT_NULL: %q", plain)
	}
}

// TestNotNullTextColumns_PublicMixedSchema asserts information_schema derivation
// on a public table: includes NOT NULL text, excludes nullable text and NOT NULL non-text.
func TestNotNullTextColumns_PublicMixedSchema(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := openMigrateTestPool(t, ctx)
	defer pool.Close()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire conn: %v", err)
	}
	defer conn.Release()

	table := fmt.Sprintf("seed_nn_text_pub_%d", time.Now().UnixNano())
	createSQL := fmt.Sprintf(`
CREATE TABLE %s (
  id bigint PRIMARY KEY,
  note text NOT NULL DEFAULT '',
  label character varying(50) NOT NULL DEFAULT '',
  optional_note text,
  amount numeric NOT NULL DEFAULT 0,
  flag boolean NOT NULL DEFAULT false,
  when_ts timestamptz NOT NULL DEFAULT now()
)`, quoteIdent(table))
	if _, err := conn.Exec(ctx, createSQL); err != nil {
		t.Fatalf("create public fixture table: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), fmt.Sprintf("DROP TABLE IF EXISTS %s", quoteIdent(table)))
	})

	got, err := notNullTextColumns(ctx, conn.Conn(), table)
	if err != nil {
		t.Fatalf("notNullTextColumns: %v", err)
	}
	want := []string{"note", "label"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("notNullTextColumns = %v, want %v", got, want)
	}

	// SQL assembly must list only the derived text columns.
	copySQL := buildCopyFromCSVSQL(table, got)
	if !strings.Contains(copySQL, `FORCE_NOT_NULL ("note", "label")`) {
		t.Fatalf("copy SQL missing expected FORCE_NOT_NULL: %s", copySQL)
	}
	for _, forbidden := range []string{"amount", "flag", "when_ts", "optional_note"} {
		if strings.Contains(copySQL, quoteIdent(forbidden)) {
			t.Fatalf("copy SQL must not force %s: %s", forbidden, copySQL)
		}
	}
}

// TestCopyTableFromCSV_ForceNotNullTextOnly exercises copyTableFromCSV end-to-end:
// NOT NULL text unquoted empty -> ”, NULL-allowed text unquoted empty stays NULL,
// NOT NULL non-text is not in FORCE_NOT_NULL (and still accepts a real value).
func TestCopyTableFromCSV_ForceNotNullTextOnly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := openMigrateTestPool(t, ctx)
	defer pool.Close()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire conn: %v", err)
	}
	defer conn.Release()

	table := fmt.Sprintf("seed_copy_force_nn_%d", time.Now().UnixNano())
	createSQL := fmt.Sprintf(`
CREATE TABLE %s (
  id bigint PRIMARY KEY,
  owner_request text NOT NULL DEFAULT '',
  staff_notes text NOT NULL DEFAULT '',
  insurance_company_name varchar(100),
  amount numeric NOT NULL DEFAULT 0
)`, quoteIdent(table))
	if _, err := conn.Exec(ctx, createSQL); err != nil {
		t.Fatalf("create table: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), fmt.Sprintf("DROP TABLE IF EXISTS %s", quoteIdent(table)))
	})

	forceCols, err := notNullTextColumns(ctx, conn.Conn(), table)
	if err != nil {
		t.Fatalf("notNullTextColumns: %v", err)
	}
	wantForce := []string{"owner_request", "staff_notes"}
	if !reflect.DeepEqual(forceCols, wantForce) {
		t.Fatalf("force-not-null columns = %v, want %v", forceCols, wantForce)
	}

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "t.csv")
	// hospitalizations-like row: unquoted empty owner_request/staff_notes/insurance
	csvBody := "id,owner_request,staff_notes,insurance_company_name,amount\n" +
		"1,,,,10\n" +
		"2,has-request,,ACME,20\n"
	if err := os.WriteFile(csvPath, []byte(csvBody), 0o600); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	rows, err := copyTableFromCSV(ctx, conn.Conn(), csvPath, table)
	if err != nil {
		t.Fatalf("copyTableFromCSV: %v", err)
	}
	if rows != 2 {
		t.Fatalf("rows = %d, want 2", rows)
	}

	var nullOwner, nullStaff, emptyOwner, emptyStaff, nullIns int
	q := fmt.Sprintf(`
SELECT
  count(*) FILTER (WHERE owner_request IS NULL),
  count(*) FILTER (WHERE staff_notes IS NULL),
  count(*) FILTER (WHERE owner_request = ''),
  count(*) FILTER (WHERE staff_notes = ''),
  count(*) FILTER (WHERE insurance_company_name IS NULL)
FROM %s`, quoteIdent(table))
	if err := conn.QueryRow(ctx, q).Scan(&nullOwner, &nullStaff, &emptyOwner, &emptyStaff, &nullIns); err != nil {
		t.Fatalf("select: %v", err)
	}
	if nullOwner != 0 || nullStaff != 0 {
		t.Fatalf("NOT NULL text nulls owner=%d staff=%d, want 0", nullOwner, nullStaff)
	}
	if emptyOwner != 1 {
		t.Fatalf("empty owner_request rows = %d, want 1", emptyOwner)
	}
	if emptyStaff != 2 {
		t.Fatalf("empty staff_notes rows = %d, want 2", emptyStaff)
	}
	if nullIns != 1 {
		t.Fatalf("nullable insurance_company_name NULL count = %d, want 1", nullIns)
	}
}

func openMigrateTestPool(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	params, err := dbconn.FromEnv()
	if err != nil {
		t.Skipf("DB env unavailable: %v", err)
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		t.Skip("DB_NAME unset")
	}
	if !dbconn.IsLocalHost(params.Host) {
		t.Fatalf("refusing non-local DB host %q", params.Host)
	}
	pool, err := pgxpool.New(ctx, params.DSN(dbName))
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping: %v", err)
	}
	return pool
}
