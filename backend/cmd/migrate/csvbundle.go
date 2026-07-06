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

	for _, entry := range manifest.Tables {
		csvPath := filepath.Join(migrationsDir, "seeds", bundleDir, entry.CSVFile)
		rows, err := copyTableFromCSV(ctx, tx.Conn(), csvPath, entry.Table)
		if err != nil {
			return fmt.Errorf("failed to load %s from %s: %w", entry.Table, csvPath, err)
		}
		logger.Info("Loaded seed table", slog.String("bundle", bundleDir), slog.String("table", entry.Table), slog.Int64("rows", rows))

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

// copyTableFromCSV streams one CSV file into one table via COPY FROM STDIN.
// No explicit column list is given — COPY defaults to the table's declared
// column order, which matches cmd/seed-export's `SELECT *` dump order exactly
// since both sides read the same, unmodified 001_init.sql schema.
func copyTableFromCSV(ctx context.Context, conn *pgx.Conn, csvPath, table string) (int64, error) {
	f, err := os.Open(csvPath) //nolint:gosec // path built from our own manifest, under the fixed migrations dir
	if err != nil {
		return 0, fmt.Errorf("failed to open %s: %w", csvPath, err)
	}
	defer f.Close() //nolint:errcheck // read-only fd; close failure is not actionable here

	copySQL := fmt.Sprintf("COPY %s FROM STDIN WITH (FORMAT csv, HEADER true)", quoteIdent(table))
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
