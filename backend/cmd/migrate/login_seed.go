package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/animal-ekarte/backend/internal/seedlogin"
)

// runLoginSeed is phase 3: upsert synthetic demo logins matching LoginForm.
// It is not a CSV bundle. The shared password is seedlogin.SharedPassword.
// Production / empty / unknown APP_ENV skip. Re-runs always upsert.
func runLoginSeed(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	appEnv := os.Getenv("APP_ENV")
	if !seedlogin.ShouldApply(appEnv) {
		logger.Info("Skipping login seed", slog.String("APP_ENV", appEnv))
		return nil
	}

	applied, err := seedlogin.Apply(ctx, db)
	if err != nil {
		return err
	}
	seedlogin.LogApplied(logger, applied)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin login seed record tx: %w", err)
	}
	if err := upsertMigrationRecord(tx, seedlogin.MigrationKey(), seedlogin.CatalogChecksum()); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record login seed: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit login seed record: %w", err)
	}
	return nil
}

// expectedSeedBundleDirs is the coverage plan: CSV bundles for APP_ENV, plus
// 003_login when the login seed applies for this APP_ENV.
func expectedSeedBundleDirs() []string {
	bundles := append([]string{}, seedBundlesForCurrentEnv()...)
	if !seedlogin.ShouldApply(os.Getenv("APP_ENV")) {
		return bundles
	}
	return append(bundles, seedlogin.BundleDir)
}

func upsertMigrationRecord(tx *sql.Tx, filename, checksum string) error {
	_, err := tx.Exec(
		`INSERT INTO schema_migrations (filename, checksum, executed_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (filename) DO UPDATE
		    SET checksum = EXCLUDED.checksum,
		        executed_at = EXCLUDED.executed_at`,
		filename, checksum, time.Now(),
	)
	return err
}
