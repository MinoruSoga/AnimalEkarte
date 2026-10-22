package testdb

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

// TestBillDiscountExcluded007StorageInvariants applies the post-001
// incremental migrations (003–006, then
// backend/migrations/007_bill_discount_excluded_and_master_versions.sql) in
// migrate order on top of a disposable 001 database and checks the A01
// storage contract:
//
//   - rows that existed before 007 keep is_bill_discount_excluded NULL
//     (no backfill), and INSERTs that omit the column also store NULL
//     (no DB default);
//   - the 3 newly versioned tables (exam_types, medicines,
//     merchandise_items) start at version=1, keep it on a no-change UPDATE,
//     advance it on business-value changes, reject app-side version SET
//     (23514), and advance it when only is_bill_discount_excluded changes;
//   - the flag-only change also advances version on the 4 previously
//     versioned master tables.
//
// DB スタックが必要（-short / DB 不在では既存の skipShortDatabase 慣習で skip）。
func TestBillDiscountExcluded007StorageInvariants(t *testing.T) {
	db := SetupIsolated001TestDB(t)

	// Apply the intermediate migrations between 001 and 007 in the same
	// order as `make migrate`. 004 adds the version columns and
	// *_storage_version triggers on consultations/procedures/vaccines that
	// the version assertions below rely on.
	for _, name := range []string{
		"003_trimming_course_image_uploads.sql",
		"004_master_insurance_versions.sql",
		"005_manual_article_operation_targets.sql",
		"006_operation_anchor_and_lab_mapping_versions.sql",
	} {
		ddl, err := os.ReadFile(migrationSQLPath(t, name))
		require.NoErrorf(t, err, "read %s", name)
		require.NoErrorf(t, db.Exec(string(ddl)).Error, "apply %s", name)
	}

	var companyID, clinicID int64
	require.NoError(t, db.Raw(`INSERT INTO companies (name) VALUES ('b007-flag') RETURNING id`).Scan(&companyID).Error)
	require.NoError(t, db.Raw(`INSERT INTO clinics (company_id, name) VALUES (?, 'b007-flag') RETURNING id`, companyID).Scan(&clinicID).Error)

	flagTables := []string{
		"consultations",
		"exam_types",
		"procedures",
		"vaccines",
		"checkup_types",
		"medicines",
		"merchandise_items",
	}
	newlyVersioned := []string{"exam_types", "medicines", "merchandise_items"}

	insertRow := func(table string) int64 {
		t.Helper()
		var id int64
		require.NoErrorf(t, db.Raw(
			fmt.Sprintf("INSERT INTO %s (clinic_id, name) VALUES (?, ?) RETURNING id", table),
			clinicID, "b007-"+table,
		).Scan(&id).Error, "insert fixture row into %s", table)
		return id
	}

	// Rows written before 007 is applied: the migration must not rewrite them.
	preIDs := make(map[string]int64, len(flagTables))
	for _, table := range flagTables {
		preIDs[table] = insertRow(table)
	}

	ddl, err := os.ReadFile(migrationSQLPath(t, "007_bill_discount_excluded_and_master_versions.sql"))
	require.NoError(t, err, "read 007 migration")
	require.NoError(t, db.Exec(string(ddl)).Error, "apply 007 migration")

	flagOf := func(table string, id int64) *bool {
		t.Helper()
		var flag *bool
		require.NoErrorf(t, db.Raw(
			fmt.Sprintf("SELECT is_bill_discount_excluded FROM %s WHERE id = ?", table), id,
		).Scan(&flag).Error, "read is_bill_discount_excluded from %s", table)
		return flag
	}
	versionOf := func(table string, id int64) int64 {
		t.Helper()
		var version int64
		require.NoErrorf(t, db.Raw(
			fmt.Sprintf("SELECT version FROM %s WHERE id = ?", table), id,
		).Scan(&version).Error, "read version from %s", table)
		return version
	}

	t.Run("existing rows keep raw NULL and omitted INSERT stores NULL", func(t *testing.T) {
		for _, table := range flagTables {
			require.Nilf(t, flagOf(table, preIDs[table]),
				"%s pre-007 row must keep is_bill_discount_excluded NULL (no backfill)", table)
			require.Nilf(t, flagOf(table, insertRow(table)),
				"%s INSERT omitting the column must store NULL (no DB default)", table)
		}
	})

	t.Run("flag stores explicit true and false as written", func(t *testing.T) {
		for _, table := range flagTables {
			id := insertRow(table)
			require.NoErrorf(t, db.Exec(
				fmt.Sprintf("UPDATE %s SET is_bill_discount_excluded = true WHERE id = ?", table), id).Error,
				"set flag true on %s", table)
			require.NotNilf(t, flagOf(table, id), "%s flag must be non-NULL after explicit true", table)
			require.Truef(t, *flagOf(table, id), "%s flag must read back true", table)
			require.NoErrorf(t, db.Exec(
				fmt.Sprintf("UPDATE %s SET is_bill_discount_excluded = false WHERE id = ?", table), id).Error,
				"set flag false on %s", table)
			require.NotNilf(t, flagOf(table, id), "%s flag must be non-NULL after explicit false", table)
			require.Falsef(t, *flagOf(table, id), "%s flag must read back false", table)
		}
	})

	t.Run("newly versioned tables follow entity_version_guard rules", func(t *testing.T) {
		for _, table := range newlyVersioned {
			require.Equalf(t, int64(1), versionOf(table, preIDs[table]),
				"%s pre-007 row must start at version 1", table)

			id := insertRow(table)
			require.Equalf(t, int64(1), versionOf(table, id),
				"%s new row must start at version 1", table)

			require.NoErrorf(t, db.Exec(
				fmt.Sprintf("UPDATE %s SET name = name WHERE id = ?", table), id).Error,
				"no-change update on %s", table)
			require.Equalf(t, int64(1), versionOf(table, id),
				"%s no-change UPDATE must keep version", table)

			require.NoErrorf(t, db.Exec(
				fmt.Sprintf("UPDATE %s SET name = 'b007-changed' WHERE id = ?", table), id).Error,
				"business change on %s", table)
			require.Equalf(t, int64(2), versionOf(table, id),
				"%s business change must increment version", table)

			err := db.Exec(
				fmt.Sprintf("UPDATE %s SET version = 99 WHERE id = ?", table), id).Error
			require.Errorf(t, err, "%s app-side version SET must fail", table)
			var pgErr *pgconn.PgError
			require.ErrorAsf(t, err, &pgErr, "%s version SET must fail as a PostgreSQL error", table)
			require.Equalf(t, "23514", pgErr.Code,
				"%s version SET must fail with 23514 (storage owns version)", table)
		}
	})

	t.Run("flag-only change increments version on all versioned masters", func(t *testing.T) {
		for _, table := range flagTables {
			id := insertRow(table)
			before := versionOf(table, id)
			require.NoErrorf(t, db.Exec(
				fmt.Sprintf("UPDATE %s SET is_bill_discount_excluded = true WHERE id = ?", table), id).Error,
				"flag-only update on %s", table)
			require.Equalf(t, before+1, versionOf(table, id),
				"%s flag-only change must increment version (A.4 flag edits are versioned writes)", table)
		}
	})
}

// migrationSQLPath resolves backend/migrations/<name> from this package
// directory (same runtime.Caller convention as isolated_001.go).
func migrationSQLPath(t *testing.T, name string) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed resolving testdb source path")
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", name)
}
