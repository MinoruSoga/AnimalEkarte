package testdb

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

// TestBillDiscountExcluded007StorageInvariants checks the A01 storage
// contract (the former 007_bill_discount_excluded_and_master_versions.sql,
// absorbed into 001_init.sql by the 2026-10-23 consolidation) against a
// disposable 001 database:
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

	// All 7 tables carry a partial unique index on (clinic_id, name) WHERE
	// deleted_at IS NULL, so every fixture row needs its own name.
	insertSeq := 0
	insertRow := func(table string) int64 {
		t.Helper()
		insertSeq++
		var id int64
		require.NoErrorf(t, db.Raw(
			fmt.Sprintf("INSERT INTO %s (clinic_id, name) VALUES (?, ?) RETURNING id", table),
			clinicID, fmt.Sprintf("b007-%s-%d", table, insertSeq),
		).Scan(&id).Error, "insert fixture row into %s", table)
		return id
	}

	// Fixture rows omit the column: the storage carries no DB default and no
	// backfill, so they must read NULL (and version=1 where versioned).
	preIDs := make(map[string]int64, len(flagTables))
	for _, table := range flagTables {
		preIDs[table] = insertRow(table)
	}

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
