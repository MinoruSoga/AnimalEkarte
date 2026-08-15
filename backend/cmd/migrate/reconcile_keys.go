package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/animal-ekarte/backend/internal/seedbundle"
)

// reconcileMigrationKeys is the pure post-apply coverage check.
//
// Inputs:
//   - recorded: schema_migrations.filename values actually present
//   - ddlFilenames: top-level *.sql basenames currently on disk
//   - seedBundles: seed bundle directory names (e.g. BundleOrder entries)
//
// Expected keys = ddlFilenames ∪ {"seeds/" + bundle for each seedBundles entry}
// (via seedbundle.BundleMigrationKey so key shape stays shared with phase 2).
//
// missing = expected − recorded  → caller must fail-closed when non-empty
// extra   = recorded − expected  → informational only (merged/deleted files)
//
// Count equality is never used: extra keys from consolidated migrations are
// legitimate and must not fail a healthy environment.
func reconcileMigrationKeys(recorded, ddlFilenames, seedBundles []string) (missing, extra []string) {
	expected := make(map[string]struct{}, len(ddlFilenames)+len(seedBundles))
	for _, name := range ddlFilenames {
		expected[name] = struct{}{}
	}
	for _, bundle := range seedBundles {
		expected[seedbundle.BundleMigrationKey(bundle)] = struct{}{}
	}

	recordedSet := make(map[string]struct{}, len(recorded))
	for _, key := range recorded {
		recordedSet[key] = struct{}{}
	}

	for key := range expected {
		if _, ok := recordedSet[key]; !ok {
			missing = append(missing, key)
		}
	}
	for key := range recordedSet {
		if _, ok := expected[key]; !ok {
			extra = append(extra, key)
		}
	}

	sort.Strings(missing)
	sort.Strings(extra)
	return missing, extra
}

// listTopLevelDDLFilenames returns basenames of non-directory *.sql files under
// dir, matching the selection rules of runSQLMigrations (directories skipped;
// only the migrations root, not seeds/).
func listTopLevelDDLFilenames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory %s: %w", dir, err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".sql" {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// loadRecordedMigrationKeys reads every schema_migrations.filename value.
func loadRecordedMigrationKeys(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT filename FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read schema_migrations for key coverage: %w", err)
	}
	defer rows.Close() //nolint:errcheck // close error is not actionable after scan path

	var keys []string
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, fmt.Errorf("failed to scan schema_migrations filename: %w", err)
		}
		keys = append(keys, filename)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate schema_migrations: %w", err)
	}
	return keys, nil
}

// verifyMigrationKeyCoverage runs after both migrate phases. It fails closed
// when any expected on-disk DDL or seed key is absent from schema_migrations.
// Extra recorded keys (consolidated/deleted files) are logged only.
//
// Summary line (one-line operator gate):
//
//	msg="Migration key coverage" missing=0 extra=N expected=E recorded=R
func verifyMigrationKeyCoverage(db *sql.DB, logger *slog.Logger) error {
	recorded, err := loadRecordedMigrationKeys(db)
	if err != nil {
		return err
	}
	ddlFiles, err := listTopLevelDDLFilenames(migrationsDir)
	if err != nil {
		return err
	}

	// Expected seed keys must match the same APP_ENV gate as runSeedBundles.
	// Production plans exclude demo/staging; previously-applied demo keys on a
	// production DB surface as informational "extra", not as missing failures.
	seedBundles := seedBundlesForCurrentEnv()
	missing, extra := reconcileMigrationKeys(recorded, ddlFiles, seedBundles)
	expectedCount := len(ddlFiles) + len(seedBundles)

	logger.Info("Migration key coverage",
		slog.Int("missing", len(missing)),
		slog.Int("extra", len(extra)),
		slog.Int("expected", expectedCount),
		slog.Int("recorded", len(recorded)),
		slog.String("APP_ENV", os.Getenv("APP_ENV")),
		slog.Any("seed_bundles", seedBundles))

	if len(extra) > 0 {
		logger.Info("Migration key coverage has extra recorded keys (informational)",
			slog.Any("extra", extra))
	}
	if len(missing) > 0 {
		return fmt.Errorf("migration key coverage check failed: missing keys %v", missing)
	}
	return nil
}
