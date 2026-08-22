package main

import (
	"database/sql"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/dbconn"
)

var disposableCSVImportDatabasePattern = regexp.MustCompile(
	`^(?:csvimport21|animalekarte_f8_g4)_[a-z0-9_]+$`,
)

// This opt-in test applies only the append-only DDL phase to a caller-provided
// disposable database. Seed bundles are deliberately excluded so migration
// compatibility can be verified without loading the large demo dataset.
func TestRunSQLMigrationsAgainstDisposablePostgres(t *testing.T) {
	if os.Getenv("MIGRATE_SQL_INTEGRATION") != "1" {
		t.Skip("set MIGRATE_SQL_INTEGRATION=1 for a disposable csvimport/F8 G4 database")
	}

	conn, err := dbconn.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	database := os.Getenv("DB_NAME")
	if !disposableCSVImportDatabasePattern.MatchString(database) {
		t.Fatalf("DB_NAME %q is not a disposable csvimport21_* or animalekarte_f8_g4_* database", database)
	}
	if confirmation := os.Getenv("MIGRATE_SQL_INTEGRATION_DB_NAME"); confirmation != database {
		t.Fatalf("MIGRATE_SQL_INTEGRATION_DB_NAME must exactly match DB_NAME")
	}
	if !dbconn.IsLocalHost(conn.Host) || conn.SSLMode != "disable" {
		t.Fatal("integration test only accepts a local/container database with SSL disabled")
	}

	db, err := sql.Open("postgres", conn.DSN(database))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	if err := ensureMigrationsTable(db); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := runSQLMigrations(db, logger); err != nil {
		t.Fatal(err)
	}

	var applied int
	if err := db.QueryRow(`
SELECT count(*)
FROM schema_migrations
WHERE filename IN (
  '001_init.sql'
)`).Scan(&applied); err != nil {
		t.Fatal(err)
	}
	if applied != 1 {
		t.Fatalf("applied DDL migrations = %d, want 1", applied)
	}

	// Topology: 001_init.sql only.
	var ddlKeys []string
	rows, err := db.Query(`
SELECT filename
FROM schema_migrations
WHERE filename NOT LIKE 'seeds/%'
ORDER BY filename`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		ddlKeys = append(ddlKeys, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(ddlKeys) != 1 || ddlKeys[0] != "001_init.sql" {
		t.Fatalf("DDL schema_migrations keys = %v, want [001_init.sql]", ddlKeys)
	}

	// Rerun is a no-op: second apply must leave history and succeed.
	if err := runSQLMigrations(db, logger); err != nil {
		t.Fatalf("rerun runSQLMigrations: %v", err)
	}
	var appliedAfterRerun int
	if err := db.QueryRow(`
SELECT count(*)
FROM schema_migrations
WHERE filename NOT LIKE 'seeds/%'`).Scan(&appliedAfterRerun); err != nil {
		t.Fatal(err)
	}
	if appliedAfterRerun != 1 {
		t.Fatalf("after rerun DDL keys = %d, want 1", appliedAfterRerun)
	}
}

// TestTopLevelDDLMigrationInventoryIncludesInitAndEnsure asserts the post-consolidation
// disk topology: 001_init.sql only under migrations/ (maxdepth 1).
func TestTopLevelDDLMigrationInventoryIncludesInitAndEnsure(t *testing.T) {
	dir := migrationsDir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// CI / host: resolve relative to this test file (backend/cmd/migrate → backend/migrations).
		_, thisFile, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatal("runtime.Caller failed")
		}
		dir = filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migrations dir %s: %v", dir, err)
	}
	var sqlFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".sql") {
			sqlFiles = append(sqlFiles, name)
		}
	}
	sort.Strings(sqlFiles)
	want := []string{"001_init.sql"}
	if len(sqlFiles) != len(want) {
		t.Fatalf("top-level DDL files = %v, want %v", sqlFiles, want)
	}
	for i := range want {
		if sqlFiles[i] != want[i] {
			t.Fatalf("top-level DDL files = %v, want %v", sqlFiles, want)
		}
	}
	if filepath.Base(dir) != "migrations" {
		t.Fatalf("migrationsDir base = %q, want migrations", filepath.Base(dir))
	}
}

// TestSeedCutoverRehearsalAgainstDisposablePostgres creates only the six
// explicit seed bindings required by csv-import. The database-name guard keeps
// this mutating helper confined to the one-shot rehearsal database.
func TestSeedCutoverRehearsalAgainstDisposablePostgres(t *testing.T) {
	if os.Getenv("CSVIMPORT_REHEARSAL_SEEDS") != "1" {
		t.Skip("set CSVIMPORT_REHEARSAL_SEEDS=1 for a disposable csvimport/F8 G4 database")
	}

	conn, err := dbconn.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	database := os.Getenv("DB_NAME")
	if !disposableCSVImportDatabasePattern.MatchString(database) {
		t.Fatalf("DB_NAME %q is not a disposable csvimport21_* or animalekarte_f8_g4_* database", database)
	}
	if confirmation := os.Getenv("CSVIMPORT_REHEARSAL_DB_NAME"); confirmation != database {
		t.Fatalf("CSVIMPORT_REHEARSAL_DB_NAME must exactly match DB_NAME")
	}
	if !dbconn.IsLocalHost(conn.Host) || conn.SSLMode != "disable" {
		t.Fatal("integration test only accepts a local/container database with SSL disabled")
	}

	db, err := sql.Open("postgres", conn.DSN(database))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback is a no-op after commit

	if _, err := tx.Exec(`
INSERT INTO companies (id, name) VALUES (1, 'CSV import rehearsal');
INSERT INTO clinics (id, company_id, name) VALUES (1, 1, 'CSV import rehearsal');
INSERT INTO animal_species (id, name) VALUES
  (1, 'Rehearsal species 1'),
  (2, 'Rehearsal species 2'),
  (3, 'Rehearsal species 3'),
  (4, 'Rehearsal species 4'),
  (5, 'Rehearsal species 5'),
  (6, 'Rehearsal species 6');
INSERT INTO exam_types (id, clinic_id, name) VALUES (11008, 1, '検査');
INSERT INTO reservation_types (id, clinic_id, name, category)
VALUES (9, 1, 'CSV import rehearsal trimming', 'trimming');
`); err != nil {
		t.Fatal(err)
	}

	var cashID, creditCardID int64
	if err := tx.QueryRow(`
SELECT
  max(id) FILTER (WHERE system_key = 'cash'),
  max(id) FILTER (WHERE system_key = 'credit_card')
FROM payment_methods
WHERE clinic_id = 1
`).Scan(&cashID, &creditCardID); err != nil {
		t.Fatal(err)
	}
	if cashID != 1 || creditCardID != 2 {
		t.Fatalf("payment method IDs = (%d, %d), want (1, 2)", cashID, creditCardID)
	}
	var activeSpecies int
	if err := tx.QueryRow(`
SELECT count(*)
FROM animal_species
WHERE id BETWEEN 1 AND 6 AND is_active = true
`).Scan(&activeSpecies); err != nil {
		t.Fatal(err)
	}
	if activeSpecies != 6 {
		t.Fatalf("active animal species IDs 1..6 = %d, want 6", activeSpecies)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func TestDisposableCSVImportDatabasePattern(t *testing.T) {
	for _, database := range []string{
		"csvimport21_rehearsal",
		"animalekarte_f8_g4_rehearsal",
	} {
		if !disposableCSVImportDatabasePattern.MatchString(database) {
			t.Fatalf("disposable database %q was rejected", database)
		}
	}
	for _, database := range []string{
		"animalekarte",
		"animalekarte_f8_g4_",
		"animalekarte_f8_g4_escape-name",
	} {
		if disposableCSVImportDatabasePattern.MatchString(database) {
			t.Fatalf("unsafe database %q was accepted", database)
		}
	}
}
