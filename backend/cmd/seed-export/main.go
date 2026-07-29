// cmd/seed-export dumps the final row state of the 90 seeded tables into CSV
// files under backend/migrations/seeds/{002_master,003_demo,004_staging}/, so
// cmd/migrate can load them via COPY at startup. As of the 2026-07 CSV-only
// migration there is no seed SQL left to read: backend/migrations/ contains
// DDL-only SQL migrations, and 002/003/004 seed data exists solely as directories of
// CSV + manifest.json under seeds/ — the stub SELECT-1 *.sql files that used
// to sit between them and cmd/migrate have been deleted.
//
// This tool never reads or parses seed SQL — there is none. It creates a
// disposable database, applies the DDL migrations plus the seed bundles to it via
// the existing, UNMODIFIED cmd/migrate binary (so the historical `DO $$ ...
// random() ...` reservation-generator block, now frozen into the 003_demo
// CSV bundle, is never re-executed), then COPY-dumps the resulting tables.
// Re-running this tool is safe and reproducible by construction: it only
// ever reads rows cmd/migrate already materialized via COPY from the
// committed CSVs, so repeated exports read the same rows.
//
// Safety: refuses to run against any DB_HOST that is not "db", "localhost",
// or "127.0.0.1" (same guard as cmd/seed-old-db). The disposable database
// name is a hardcoded constant, never taken from the DB_NAME environment
// variable, so there is no code path that can point a CREATE/DROP DATABASE
// at whatever the caller's real dev/staging database happens to be named.
//
// Run via: docker compose exec backend go run ./cmd/seed-export
// (requires: db running, DB_HOST/PORT/USER/PASSWORD set as usual)
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/csvimport"
	"github.com/animal-ekarte/backend/internal/dbconn"
)

// tmpDBName is the disposable database this tool always operates against.
// Never derived from user/env input — see package doc.
const tmpDBName = "seed_export_tmp"

// seedsRootDir is where bundle directories (CSV + manifest.json) are written,
// relative to the backend module root this binary is run from.
const seedsRootDir = "migrations/seeds"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(context.Background(), logger); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	if err := config.ConfigureTimeZone(); err != nil {
		return fmt.Errorf("timezone configuration failed: %w", err)
	}

	conn, err := dbconn.FromEnv()
	if err != nil {
		return err
	}

	if !dbconn.IsLocalHost(conn.Host) {
		return fmt.Errorf(
			"SAFETY: DB_HOST=%q is not a known local host — refusing to create/drop %s against a non-local DB",
			conn.Host, tmpDBName,
		)
	}

	logger.Info("Preparing disposable export database", slog.String("db", tmpDBName), slog.String("host", conn.Host))

	maintPool, err := pgxpool.New(ctx, conn.DSN("postgres"))
	if err != nil {
		return fmt.Errorf("failed to connect to maintenance db: %w", err)
	}
	defer maintPool.Close()

	if err := recreateTmpDatabase(ctx, maintPool, logger); err != nil {
		return err
	}
	// Best-effort cleanup even on failure paths below, so a crashed run
	// doesn't leave seed_export_tmp occupying the DB for the next attempt.
	defer func() {
		if err := dropTmpDatabase(ctx, maintPool, logger); err != nil {
			logger.Error("failed to drop disposable database on cleanup", slog.String("error", err.Error()))
		}
	}()

	logger.Info("Applying 001_init + seed bundles to disposable database via cmd/migrate ...")
	if err := applyMigrationsToTmpDB(ctx, conn, logger); err != nil {
		return fmt.Errorf("failed to apply migrations to %s: %w", tmpDBName, err)
	}

	tmpPool, err := pgxpool.New(ctx, conn.DSN(tmpDBName))
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", tmpDBName, err)
	}
	defer tmpPool.Close()
	if sourceDir := os.Getenv("SEED_EXPORT_CSV_SOURCE"); sourceDir != "" {
		var clinicID, speciesID, examTypeID int64
		if err := tmpPool.QueryRow(ctx, `SELECT id FROM clinics WHERE name = $1 ORDER BY id LIMIT 1`, "八王子病院").Scan(&clinicID); err != nil {
			return fmt.Errorf("resolve import clinic: %w", err)
		}
		if err := tmpPool.QueryRow(ctx, `SELECT id FROM animal_species ORDER BY id LIMIT 1`).Scan(&speciesID); err != nil {
			return fmt.Errorf("resolve import species: %w", err)
		}
		if err := tmpPool.QueryRow(ctx, `SELECT id FROM exam_types WHERE name = $1 AND clinic_id = $2 LIMIT 1`, "検査", clinicID).Scan(&examTypeID); err != nil {
			return fmt.Errorf("resolve import exam type: %w", err)
		}
		counts, err := csvimport.Import(ctx, tmpPool, sourceDir, tmpDBName, clinicID, speciesID, examTypeID)
		if err != nil {
			return fmt.Errorf("CSV source import failed: %w", err)
		}
		logger.Info("Imported read-only CSV source into disposable database", slog.Any("counts", counts))
	}

	// Optional post-migrate SQL for additive seed rows before COPY dump.
	// Enables INSERT → dump without hand-editing CSV. Unset to skip.
	if postSQL := os.Getenv("SEED_EXPORT_POST_SQL"); postSQL != "" {
		sqlBytes, readErr := os.ReadFile(postSQL) //nolint:gosec // operator-supplied local path only
		if readErr != nil {
			return fmt.Errorf("failed to read SEED_EXPORT_POST_SQL=%s: %w", postSQL, readErr)
		}
		logger.Info("Applying SEED_EXPORT_POST_SQL before dump", slog.String("path", postSQL))
		if _, execErr := tmpPool.Exec(ctx, string(sqlBytes)); execErr != nil {
			return fmt.Errorf("SEED_EXPORT_POST_SQL failed: %w", execErr)
		}
	}

	if err := dumpAllBundles(ctx, tmpPool, logger); err != nil {
		return fmt.Errorf("failed to dump seed bundles: %w", err)
	}

	logger.Info("✓ seed-export completed", slog.String("output", seedsRootDir))
	return nil
}

// recreateTmpDatabase drops (if present, WITH FORCE so lingering connections
// from a crashed prior run don't block it) and creates tmpDBName. CREATE/DROP
// DATABASE cannot run inside a transaction block; pgxpool.Exec issues a plain
// simple-query, which is what Postgres requires here.
func recreateTmpDatabase(ctx context.Context, maintPool *pgxpool.Pool, logger *slog.Logger) error {
	if _, err := maintPool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", tmpDBName)); err != nil {
		return fmt.Errorf("failed to drop pre-existing %s: %w", tmpDBName, err)
	}
	if _, err := maintPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", tmpDBName)); err != nil {
		return fmt.Errorf("failed to create %s: %w", tmpDBName, err)
	}
	logger.Info("Disposable database created", slog.String("db", tmpDBName))
	return nil
}

func dropTmpDatabase(ctx context.Context, maintPool *pgxpool.Pool, logger *slog.Logger) error {
	if _, err := maintPool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", tmpDBName)); err != nil {
		return err
	}
	logger.Info("Disposable database dropped", slog.String("db", tmpDBName))
	return nil
}

// applyMigrationsToTmpDB shells out to the existing, unmodified cmd/migrate
// binary with DB_NAME overridden to the disposable database. This is
// deliberate: seed-export must never reimplement "apply DDL + seed bundles", it must
// reuse the exact same code path every other environment uses, so the
// disposable DB ends up in a state indistinguishable from a real fresh apply.
func applyMigrationsToTmpDB(ctx context.Context, conn dbconn.ConnParams, logger *slog.Logger) error {
	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/migrate")
	cmd.Env = append(os.Environ(),
		"DB_HOST="+conn.Host,
		"DB_PORT="+conn.Port,
		"DB_USER="+conn.User,
		"DB_PASSWORD="+conn.Password,
		"DB_SSL_MODE="+conn.SSLMode,
		"DB_SSL_ROOT_CERT="+conn.SSLRootCert,
		"DB_NAME="+tmpDBName,
		"DB_RESET=false",
	)
	cmd.Stdout = logStreamWriter{logger}
	cmd.Stderr = logStreamWriter{logger}
	return cmd.Run()
}

// logStreamWriter adapts a *slog.Logger to io.Writer so the migrate
// subprocess's output is folded into this tool's structured log stream
// instead of writing to a separate fd.
type logStreamWriter struct{ logger *slog.Logger }

func (w logStreamWriter) Write(p []byte) (int, error) {
	w.logger.Info(string(p))
	return len(p), nil
}
