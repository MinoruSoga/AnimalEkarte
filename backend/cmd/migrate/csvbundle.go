package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/animal-ekarte/backend/internal/seedbundle"
)

// bundleChecksum returns the checksum tracked in schema_migrations (under
// seedbundle.BundleMigrationKey(bundleDir)) for one seed bundle. It hashes
// the bundle's manifest.json together with every CSV file it references, in
// manifest order. Since the stub SQL files that used to carry seed
// application were deleted, the manifest + CSVs are now the entire content
// that constitutes a bundle — there is no separate "file body" to fold in.
// An edit to any CSV, or to manifest.json itself (e.g. reordering tables),
// changes this checksum, so isAlreadyApplied's mismatch guard still fires on
// an already-migrated database.
func bundleChecksum(migrationsDir, bundleDir string) (string, error) {
	h := sha256.New()

	manifest, manifestBytes, err := loadBundleManifest(migrationsDir, bundleDir)
	if err != nil {
		return "", fmt.Errorf("checksum for bundle %s: %w", bundleDir, err)
	}
	h.Write(manifestBytes)

	for _, entry := range manifest.Tables {
		csvPath := filepath.Join(migrationsDir, "seeds", bundleDir, entry.CSVFile)
		csvBytes, err := os.ReadFile(csvPath) //nolint:gosec // path built from our own manifest, under the fixed migrations dir
		if err != nil {
			return "", fmt.Errorf("checksum for bundle %s: failed to read %s: %w", bundleDir, csvPath, err)
		}
		h.Write(csvBytes)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// loadBundleManifest reads and parses a bundle's manifest.json, returning
// both the parsed struct (for iterating tables) and the raw bytes (so
// bundleChecksum can fold the manifest's own content into the hash — an edit
// to manifest.json itself, e.g. reordering tables, must also bust the
// checksum).
func loadBundleManifest(migrationsDir, bundleDir string) (seedbundle.Manifest, []byte, error) {
	manifestPath := filepath.Join(migrationsDir, "seeds", bundleDir, "manifest.json")
	raw, err := os.ReadFile(manifestPath) //nolint:gosec // fixed migrations dir, not external input
	if err != nil {
		return seedbundle.Manifest{}, nil, fmt.Errorf("failed to read %s: %w", manifestPath, err)
	}
	var manifest seedbundle.Manifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return seedbundle.Manifest{}, nil, fmt.Errorf("failed to parse %s: %w", manifestPath, err)
	}
	return manifest, raw, nil
}

// clinicDefaultTrimmingCourseTypesTrigger auto-inserts default
// trimming_course_types on clinics INSERT (001_init.sql). When the seed
// bundle ships an explicit trimming_course_types.csv with stable IDs that
// trimming_courses.course_type_id references, that trigger must not fire
// during clinics COPY or the explicit CSV collides on PK/UNIQUE.
const clinicDefaultTrimmingCourseTypesTrigger = "trg_create_default_trimming_course_types"

// clinicDefaultPaymentMethodsTrigger auto-inserts default payment_methods on
// clinics INSERT (001_init.sql). When the seed bundle ships an explicit
// payment_methods.csv with stable IDs that payments/payment_splits may
// reference, that trigger must not fire during clinics COPY or the explicit
// CSV collides on PK/UNIQUE.
const clinicDefaultPaymentMethodsTrigger = "trg_create_default_payment_methods"

// applyCSVBundle loads every table in a bundle's manifest via a raw Postgres
// COPY FROM STDIN, in manifest order (already FK-safe — see
// cmd/seed-export/tables.go for how that order was derived and verified),
// inside a single transaction so the whole bundle lands atomically.
//
// This uses pgx's low-level CopyFrom (conn.PgConn().CopyFrom), which streams
// CSV bytes to Postgres and lets the server parse/type-coerce them — there is
// no per-column Go-side mapping code, mirroring cmd/seed-export's dump side
// (COPY ... TO STDOUT) so the format is symmetric by construction.
//
// connStr is the same libpq-style DSN cmd/migrate already builds for its
// lib/pq connection (host/port/user/password/dbname/sslmode/TimeZone) — pgx
// accepts the identical string, so no separate connection-string
// construction is needed here.
func applyCSVBundle(ctx context.Context, connStr, migrationsDir, bundleDir string, logger *slog.Logger) error {
	manifest, _, err := loadBundleManifest(migrationsDir, bundleDir)
	if err != nil {
		return err
	}

	// Explicit CSV owns IDs for trigger-default child masters; disable the
	// clinics default triggers only around the clinics COPY so app-time clinic
	// creation still gets defaults after seed commits.
	ownsTrimmingCourseTypes := false
	ownsPaymentMethods := false
	for _, entry := range manifest.Tables {
		switch entry.Table {
		case "trimming_course_types":
			ownsTrimmingCourseTypes = true
		case "payment_methods":
			ownsPaymentMethods = true
		}
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to open pgx pool for CSV bundle load: %w", err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin CSV bundle transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after a successful Commit; rollback is the failure-path safety net

	// Sequences do not roll back with failed seed transactions. Clinic INSERT
	// triggers (payment_methods / trimming_course_types) consume nextval during
	// a rolled-back attempt; the next retry then assigns non-1-based IDs while
	// CSVs still hardcode the original serials (course_type_id=1/3/5, …) and
	// fail FK. Re-align those sequences when the child table is empty.
	// Safety net when a trigger-owned table is empty after a failed seed, or
	// when an explicit CSV is loaded without the matching owns* flag.
	if err := realignEmptyTriggerSerials(ctx, tx); err != nil {
		return fmt.Errorf("failed to realign empty trigger serials: %w", err)
	}

	for _, entry := range manifest.Tables {
		if entry.Table == "clinics" {
			if ownsTrimmingCourseTypes {
				if _, err := tx.Exec(ctx, `ALTER TABLE clinics DISABLE TRIGGER `+clinicDefaultTrimmingCourseTypesTrigger); err != nil {
					return fmt.Errorf("failed to disable %s for seed clinics load: %w", clinicDefaultTrimmingCourseTypesTrigger, err)
				}
			}
			if ownsPaymentMethods {
				if _, err := tx.Exec(ctx, `ALTER TABLE clinics DISABLE TRIGGER `+clinicDefaultPaymentMethodsTrigger); err != nil {
					return fmt.Errorf("failed to disable %s for seed clinics load: %w", clinicDefaultPaymentMethodsTrigger, err)
				}
			}
		}

		csvPath := filepath.Join(migrationsDir, "seeds", bundleDir, entry.CSVFile)
		rows, err := copyTableFromCSV(ctx, tx.Conn(), csvPath, entry.Table)
		if err != nil {
			return fmt.Errorf("failed to load %s from %s: %w", entry.Table, csvPath, err)
		}
		logger.Info("Loaded seed table", slog.String("bundle", bundleDir), slog.String("table", entry.Table), slog.Int64("rows", rows))

		if entry.Table == "clinics" {
			// Re-enable inside the same transaction before commit so a
			// successful seed does not leave the triggers permanently off.
			if ownsTrimmingCourseTypes {
				if _, err := tx.Exec(ctx, `ALTER TABLE clinics ENABLE TRIGGER `+clinicDefaultTrimmingCourseTypesTrigger); err != nil {
					return fmt.Errorf("failed to re-enable %s after seed clinics load: %w", clinicDefaultTrimmingCourseTypesTrigger, err)
				}
			}
			if ownsPaymentMethods {
				if _, err := tx.Exec(ctx, `ALTER TABLE clinics ENABLE TRIGGER `+clinicDefaultPaymentMethodsTrigger); err != nil {
					return fmt.Errorf("failed to re-enable %s after seed clinics load: %w", clinicDefaultPaymentMethodsTrigger, err)
				}
			}
		}

		// COPY loads explicit id values without touching the backing
		// SERIAL/BIGSERIAL sequence, so the next application-level INSERT
		// (which relies on nextval()) would collide with the max id we just
		// loaded. Advance the sequence to match — a no-op for the small
		// number of seeded tables that have no "id" column at all or whose
		// "id" isn't sequence-backed (composite PK, UUID default).
		if err := advanceSerialSequence(ctx, tx, entry.Table); err != nil {
			return fmt.Errorf("failed to advance sequence for %s: %w", entry.Table, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit CSV bundle %s: %w", bundleDir, err)
	}
	return nil
}

// realignEmptyTriggerSerials resets id sequences for tables that clinic
// INSERT triggers populate when those tables are currently empty. Safe to
// call on a fresh DB (setval(1,false) is a no-op relative to the default)
// and required after a prior failed seed that advanced sequences without
// leaving rows.
func realignEmptyTriggerSerials(ctx context.Context, tx pgx.Tx) error {
	// Keep in sync with CREATE TRIGGER defaults in backend/migrations/001_init.sql.
	for _, table := range []string{"trimming_course_types", "payment_methods"} {
		var count int64
		if err := tx.QueryRow(ctx, fmt.Sprintf(
			`SELECT COUNT(*) FROM %s`, quoteIdent(table),
		)).Scan(&count); err != nil {
			return fmt.Errorf("count %s: %w", table, err)
		}
		if count > 0 {
			continue
		}
		if _, err := tx.Exec(ctx, fmt.Sprintf(`
			DO $$
			DECLARE
				seq regclass;
			BEGIN
				seq := pg_get_serial_sequence(%[1]s, 'id');
				IF seq IS NOT NULL THEN
					PERFORM setval(seq, 1, false);
				END IF;
			END $$;
		`, quoteLiteral(table))); err != nil {
			return fmt.Errorf("reset sequence for empty %s: %w", table, err)
		}
	}
	return nil
}

// advanceSerialSequence advances table's "id" sequence (if any) to match the
// max id already present, so the next application-level INSERT doesn't
// collide with a CSV-loaded row. Generic and safe across all 90 seeded
// tables without per-table metadata: it no-ops when the table has no "id"
// column (e.g. staff_permission_groups, composite PK) or when "id" isn't
// sequence-backed (e.g. lstep_csv_imports, UUID default) — pg_get_serial_
// sequence returns NULL in both cases and the DO block simply skips setval.
func advanceSerialSequence(ctx context.Context, tx pgx.Tx, table string) error {
	_, err := tx.Exec(ctx, fmt.Sprintf(`
		DO $$
		DECLARE
			seq regclass;
			max_id bigint;
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = %[1]s AND column_name = 'id'
			) THEN
				seq := pg_get_serial_sequence(%[1]s, 'id');
				IF seq IS NOT NULL THEN
					EXECUTE format('SELECT max(id) FROM %%I', %[1]s) INTO max_id;
					PERFORM setval(seq, COALESCE(max_id, 1), max_id IS NOT NULL);
				END IF;
			END IF;
		END $$;
	`, quoteLiteral(table)))
	return err
}

// quoteLiteral single-quotes a string for embedding as a SQL literal
// (table name passed to information_schema / pg_get_serial_sequence, not an
// identifier position). Doubles embedded single quotes per standard SQL
// escaping. Inputs here are always our own hardcoded/manifest-derived table
// names, never external input.
func quoteLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// textDataTypes are information_schema.data_type values that accept empty
// string as a meaningful non-NULL value. FORCE_NOT_NULL is only safe for
// these: applying it to bigint/numeric/boolean/timestamptz/enum would turn
// unquoted empty CSV fields into ” and fail type coercion.
var textDataTypes = map[string]struct{}{
	"text":              {},
	"character varying": {},
	"character":         {},
}

// notNullTextColumns returns public.<table> columns that are NOT NULL and
// text-family (text / varchar / char). Used to build COPY FORCE_NOT_NULL so
// unquoted empty CSV fields become ” instead of NULL — matching
// `text NOT NULL DEFAULT ”` semantics when COPY supplies every column.
//
// Columns are ordered by ordinal_position for stable SQL. Identifiers are
// still quoteIdent'd at SQL assembly time even though they come from
// information_schema.
func notNullTextColumns(ctx context.Context, conn *pgx.Conn, table string) ([]string, error) {
	rows, err := conn.Query(ctx, `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name = $1
		  AND is_nullable = 'NO'
		ORDER BY ordinal_position
	`, table)
	if err != nil {
		return nil, fmt.Errorf("query not-null columns for %s: %w", table, err)
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var name, dataType string
		if err := rows.Scan(&name, &dataType); err != nil {
			return nil, fmt.Errorf("scan not-null column for %s: %w", table, err)
		}
		if _, ok := textDataTypes[dataType]; !ok {
			continue
		}
		cols = append(cols, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate not-null columns for %s: %w", table, err)
	}
	return cols, nil
}

// buildCopyFromCSVSQL builds COPY ... FROM STDIN for seed load. When forceNotNull
// is non-empty it appends FORCE_NOT_NULL (col, ...) with quoteIdent-safe names.
// Column list is omitted: COPY targets the table's declared column order, which
// matches cmd/seed-export's SELECT * dump order.
func buildCopyFromCSVSQL(table string, forceNotNull []string) string {
	options := []string{"FORMAT csv", "HEADER true"}
	if len(forceNotNull) > 0 {
		quoted := make([]string, len(forceNotNull))
		for i, col := range forceNotNull {
			quoted[i] = quoteIdent(col)
		}
		options = append(options, "FORCE_NOT_NULL ("+strings.Join(quoted, ", ")+")")
	}
	return fmt.Sprintf(
		"COPY %s FROM STDIN WITH (%s)",
		quoteIdent(table),
		strings.Join(options, ", "),
	)
}

// copyTableFromCSV streams one CSV file into one table via COPY FROM STDIN.
// No explicit column list is given — COPY defaults to the table's declared
// column order, which matches cmd/seed-export's `SELECT *` dump order exactly
// since both sides read the same, unmodified 001_init.sql schema.
//
// NOT NULL text-family columns receive FORCE_NOT_NULL so unquoted empty CSV
// fields become empty strings (not NULL). NULL-allowed columns and non-text
// NOT NULL columns are left alone: the former keep NULL-from-empty semantics;
// the latter would type-coerce ” and fail.
func copyTableFromCSV(ctx context.Context, conn *pgx.Conn, csvPath, table string) (int64, error) {
	f, err := os.Open(csvPath) //nolint:gosec // path built from our own manifest, under the fixed migrations dir
	if err != nil {
		return 0, fmt.Errorf("failed to open %s: %w", csvPath, err)
	}
	defer f.Close() //nolint:errcheck // read-only fd; close failure is not actionable here

	forceNotNull, err := notNullTextColumns(ctx, conn, table)
	if err != nil {
		return 0, err
	}
	copySQL := buildCopyFromCSVSQL(table, forceNotNull)
	tag, err := conn.PgConn().CopyFrom(ctx, f, copySQL)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// quoteIdent double-quotes a Postgres identifier. Table names here always
// come from our own manifest.json (itself generated from the hardcoded,
// verified bundleTables list in cmd/seed-export), never external input.
func quoteIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}
