package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/animal-ekarte/backend/internal/seedlogin"
)

// runLoginSeed is phase 3: upsert synthetic demo logins matching LoginForm,
// then optionally one operator system-admin from SEEDLOGIN_OPERATOR_* env.
// It is not a CSV bundle. The shared password is seedlogin.SharedPassword
// and applies only to catalog emails. Production / empty / unknown APP_ENV skip.
// When schema_migrations already records the current catalog checksum, skip.
// Catalog changes (checksum drift) re-upsert and refresh the record.
func runLoginSeed(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	appEnv := os.Getenv("APP_ENV")
	if !seedlogin.ShouldApply(appEnv) {
		logger.Info("Skipping login seed", slog.String("APP_ENV", appEnv))
		return nil
	}

	key := seedlogin.MigrationKey()
	checksum := seedlogin.CatalogChecksum()
	needsApply, err := loginSeedNeedsApply(db, key, checksum)
	if err != nil {
		return err
	}
	if !needsApply {
		logger.Info("⏭ Skipping login seed (already applied)",
			slog.String("bundle", seedlogin.BundleDir),
			slog.String("APP_ENV", appEnv),
		)
		return nil
	}

	logger.Info("Applying login seed",
		slog.String("bundle", seedlogin.BundleDir),
		slog.String("APP_ENV", appEnv),
	)
	applied, err := seedlogin.Apply(ctx, db)
	if err != nil {
		return err
	}
	seedlogin.LogApplied(logger, applied)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin login seed record tx: %w", err)
	}
	if err := upsertMigrationRecord(tx, key, checksum); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record login seed: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit login seed record: %w", err)
	}
	return nil
}

// loginSeedNeedsApply reports whether catalog upsert should run.
// Missing row or checksum change → apply. Matching checksum → skip.
// Unlike DDL isAlreadyApplied, checksum change is not fail-closed: login seed is upsertable.
func loginSeedNeedsApply(db *sql.DB, filename, checksum string) (bool, error) {
	var storedChecksum string
	err := db.QueryRow(
		"SELECT checksum FROM schema_migrations WHERE filename = $1",
		filename,
	).Scan(&storedChecksum)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to query schema_migrations for login seed: %w", err)
	}
	return storedChecksum != checksum, nil
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
