package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/animal-ekarte/backend/internal/seedbundle"
)

// dumpAllBundles walks bundleOrder and dumps every table in bundleTables for
// each bundle, writing one manifest.json per bundle directory.
func dumpAllBundles(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	dumped := 0
	for _, bundle := range bundleOrder {
		tables := bundleTables[bundle]
		if err := dumpBundle(ctx, pool, logger, bundle, tables); err != nil {
			return fmt.Errorf("bundle %s: %w", bundle, err)
		}
		dumped += len(tables)
	}
	if dumped != totalSeedTableCount {
		return fmt.Errorf(
			"table count mismatch: dumped %d tables across all bundles, expected %d — bundleTables was edited without updating totalSeedTableCount (or vice versa)",
			dumped, totalSeedTableCount,
		)
	}
	logger.Info("All bundles dumped", slog.Int("tables", dumped))
	return nil
}

func dumpBundle(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger, bundle string, tables []string) error {
	dir := filepath.Join(seedsRootDir, bundle)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create bundle dir %s: %w", dir, err)
	}

	manifest := seedbundle.Manifest{Bundle: bundle}
	for _, table := range tables {
		csvFile := table + ".csv"
		rows, err := dumpTable(ctx, pool, filepath.Join(dir, csvFile), table)
		if err != nil {
			return fmt.Errorf("table %s: %w", table, err)
		}
		manifest.Tables = append(manifest.Tables, seedbundle.ManifestEntry{
			Table:   table,
			CSVFile: csvFile,
		})
		logger.Info("Dumped table", slog.String("bundle", bundle), slog.String("table", table), slog.Int64("rows", rows))
	}

	manifestPath := filepath.Join(dir, "manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", manifestPath, err)
	}
	return nil
}

// dumpTable COPY-dumps one table to a CSV file, ordered by its primary key
// columns for deterministic, diffable, checksum-stable output across
// re-exports of identical data. Falls back to unordered COPY only if the
// table genuinely has no primary key (none of the 90 seeded tables currently
// hit this fallback — verified — but it's kept non-fatal rather than assumed
// away, since a future schema change could add one).
func dumpTable(ctx context.Context, pool *pgxpool.Pool, csvPath, table string) (int64, error) {
	pkCols, err := primaryKeyColumns(ctx, pool, table)
	if err != nil {
		return 0, fmt.Errorf("failed to introspect primary key: %w", err)
	}

	var copySQL string
	if len(pkCols) > 0 {
		quoted := make([]string, len(pkCols))
		for i, c := range pkCols {
			quoted[i] = quoteIdent(c)
		}
		copySQL = fmt.Sprintf(
			"COPY (SELECT * FROM %s ORDER BY %s) TO STDOUT WITH (FORMAT csv, HEADER true)",
			quoteIdent(table), strings.Join(quoted, ", "),
		)
	} else {
		copySQL = fmt.Sprintf("COPY %s TO STDOUT WITH (FORMAT csv, HEADER true)", quoteIdent(table))
	}

	f, err := os.Create(csvPath) //nolint:gosec // csvPath is built from our own hardcoded bundle/table list, not external input
	if err != nil {
		return 0, fmt.Errorf("failed to create %s: %w", csvPath, err)
	}
	defer f.Close() //nolint:errcheck // dump connection close failure is not actionable here

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	tag, err := conn.Conn().PgConn().CopyTo(ctx, f, copySQL)
	if err != nil {
		return 0, fmt.Errorf("COPY TO failed: %w", err)
	}
	return tag.RowsAffected(), nil
}

// primaryKeyColumns returns a table's primary key column names in ordinal
// position order, or an empty slice if the table has no primary key.
func primaryKeyColumns(ctx context.Context, pool *pgxpool.Pool, table string) ([]string, error) {
	rows, err := pool.Query(ctx, `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = $1::regclass AND i.indisprimary
		ORDER BY array_position(i.indkey, a.attnum)
	`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, err
		}
		cols = append(cols, col)
	}
	return cols, rows.Err()
}

// quoteIdent double-quotes a Postgres identifier. Inputs here are always
// drawn from our own hardcoded bundleTables list or a pg_attname introspection
// result, never from external/user input, but quoting is kept cheap insurance
// against a table/column name that needs it (e.g. a reserved word).
func quoteIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}
