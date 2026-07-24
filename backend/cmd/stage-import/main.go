// cmd/stage-import imports validated old_db migration data from the
// `animalekarte_stage` schema (in old_db's separate `ani_legacy` Postgres) into
// the local AnimalEkarte本テーブル.
//
// This replaces the deprecated direct old-db seeder (cmd/seed-old-db, which reads
// TSV exports). The stage schema is produced by old_db's 3-layer pipeline
// (legacy_raw -> legacy_canonical -> animalekarte_stage) and is validated by
// `make migration-stage-check` BEFORE import, so mapping correctness is provable
// with a single SQL query. See old_db/docs/migration/three-layer-pipeline.md.
//
// Safety boundaries (all enforced here):
//   - The TARGET connection must be a local host (db/localhost/127.0.0.1); a
//     non-local target is refused. The STAGE connection is opened READ ONLY.
//   - Default mode is --dry-run (no writes). The destructive apply path
//     (delete old_db rows + reinsert) requires BOTH --apply AND
//     --confirm-local-destroy, so it can never run by accident.
//   - Apply runs in a SINGLE transaction: either all old_db rows are replaced or
//     nothing changes (rollback-safe; no half-imported state).
//
// Source contract: only rows with mapping_status IN ('confirmed','inferred') are
// imported. needs_review / archive_only / blocked rows are never written; their
// counts and reasons are reported.
//
// Required env (target, shared with migrate): DB_HOST, DB_PORT, DB_USER,
// DB_PASSWORD, DB_NAME. Stage env: STAGE_DB_HOST (default old-db-postgres),
// STAGE_DB_PORT (5432), STAGE_DB_USER (postgres), STAGE_DB_NAME (ani_legacy),
// STAGE_DB_PASSWORD (optional; empty for trust auth).
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/dbconn"
)

// targetEnumTypes are the AnimalEkarte enum types used by imported columns. They
// are registered on each target connection so pgx CopyFrom can encode the plain
// enum-label strings the stage SELECT produces.
var targetEnumTypes = []string{
	"membership_type",
	"pet_gender",
	"pet_status",
	"medical_record_status",
	"visit_type",
	"billing_status",
	"item_category",
	"tax_type",
}

// importableStatuses are the only mapping_status values written to本テーブル.
// Everything else (needs_review/archive_only/blocked) is reported, never inserted.
var importableStatuses = []string{"confirmed", "inferred"}

// clinicName is the single clinic every old_db row resolves to. The stage layer
// already binds every row to its synthetic clinic id, but the importer
// re-resolves clinic_id from the TARGET by NAME so a stale stage id can never
// leak a wrong/branch-code clinic into本テーブル. Mirrors the direct seeder.
const clinicName = "八王子病院"

// examTypeFallbackName is the root exam category used as the exam_type_id fallback
// for all imported old_db exams. Per-row exam type mapping is structurally impossible
// (TBL_KNS_HIST.HkKnsa_No is a field-group row marker, not a per-visit type selector).
// The root category is the least-wrong single assignment; resolved by name+clinic so
// no magic id is hardcoded.
const examTypeFallbackName = "検査"

// oldDBOwnerIDFloor scopes "old_db-derived" target rows for the destructive
// pre-import delete. Legacy owner ids are preserved as the owners PK and start at
// 300001; demo owners occupy 1-61. The wide gap is the provenance boundary
// (see memory: animalekarte-pre-olddb-cleanup). Children are scoped by descent
// from these owners, never by a raw id guess.
const oldDBOwnerIDFloor = 300000

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

type options struct {
	apply        bool
	confirmLocal bool
	runID        string // optional: import only this migration_run_id (default: all)
}

func parseFlags() options {
	var opt options
	flag.BoolVar(&opt.apply, "apply", false, "perform the destructive import (delete old_db rows + reinsert). Without this, dry-run only.")
	flag.BoolVar(&opt.confirmLocal, "confirm-local-destroy", false, "required alongside --apply: acknowledges the local DB old_db rows will be deleted and reloaded.")
	flag.StringVar(&opt.runID, "run-id", "", "optional: restrict import to a single migration_run_id (default: all staged rows).")
	flag.Parse()
	return opt
}

func run(logger *slog.Logger) error {
	if err := config.ConfigureTimeZone(); err != nil {
		return fmt.Errorf("timezone configuration failed: %w", err)
	}
	opt := parseFlags()

	target, err := targetDSN()
	if err != nil {
		return err
	}
	stage := stageDSN()

	ctx := context.Background()
	targetCfg, err := pgxpool.ParseConfig(target.connString())
	if err != nil {
		return fmt.Errorf("parse target config: %w", err)
	}
	// Register the AnimalEkarte enum types on every target connection so CopyFrom
	// can encode the plain-text enum labels the stage SELECT produces (the stage DB
	// has no enum types, so the values arrive as Go strings).
	targetCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		for _, name := range targetEnumTypes {
			t, err := conn.LoadType(ctx, name)
			if err != nil {
				return fmt.Errorf("load enum type %s: %w", name, err)
			}
			conn.TypeMap().RegisterType(t)
		}
		return nil
	}
	targetPool, err := pgxpool.NewWithConfig(ctx, targetCfg)
	if err != nil {
		return fmt.Errorf("open target pool: %w", err)
	}
	defer targetPool.Close()
	if err := targetPool.Ping(ctx); err != nil {
		return fmt.Errorf("ping target: %w", err)
	}

	stagePool, err := pgxpool.New(ctx, stage.connStringReadOnly())
	if err != nil {
		return fmt.Errorf("open stage pool: %w", err)
	}
	defer stagePool.Close()
	if err := stagePool.Ping(ctx); err != nil {
		return fmt.Errorf("ping stage: %w", err)
	}

	logger.Info("connected",
		slog.String("target_host", target.Host), slog.String("target_db", target.name),
		slog.String("stage_host", stage.Host), slog.String("stage_db", stage.name),
		slog.Bool("apply", opt.apply))

	// Resolve target-side ids needed for NOT NULL columns the stage doesn't carry.
	res, err := resolveTargetRefs(ctx, targetPool)
	if err != nil {
		return fmt.Errorf("resolve target refs: %w", err)
	}
	logger.Info("resolved target refs",
		slog.Int64("clinic_id", res.clinicID),
		slog.Int64("fallback_animal_species_id", res.fallbackAnimalSpeciesID),
		slog.Int64("fallback_exam_type_id", res.fallbackExamTypeID))
	// The stage has no per-exam type/species, so EVERY imported pet gets the same
	// fallback animal_species_id and EVERY imported exam the same exam_type_id.
	// Surface this so those columns are not mistaken for analytically meaningful.
	logger.Warn("stage lacks per-row animal_species_id / exam_type_id; all imported pets/exams use the single fallback id (not analytically meaningful for old_db rows)")

	// Plan: per-table importable / skipped-by-status counts from stage, plus the
	// old_db rows the apply would delete from the target. No writes here.
	plan, err := buildPlan(ctx, stagePool, targetPool, opt.runID)
	if err != nil {
		return fmt.Errorf("build plan: %w", err)
	}
	printPlan(logger, plan, opt)

	if !opt.apply {
		logger.Info("DRY-RUN complete — no rows were written. Re-run with --apply --confirm-local-destroy to import.")
		return nil
	}
	if !opt.confirmLocal {
		return fmt.Errorf("--apply requires --confirm-local-destroy (this DELETES and reloads all old_db rows in the local target DB)")
	}

	inserted, deleted, err := applyImport(ctx, targetPool, stagePool, res, opt.runID, logger, nil)
	if err != nil {
		return fmt.Errorf("apply import (transaction rolled back, no changes): %w", err)
	}
	logger.Info("─────────────────────────────────────────")
	for _, t := range importOrder {
		logger.Info("imported", slog.String("table", t),
			slog.Int64("deleted", deleted[t]), slog.Int64("inserted", inserted[t]))
	}
	logger.Info("✓ stage-import completed successfully (single transaction committed)")
	return nil
}

// dsn holds a parsed connection target: dbconn.ConnParams (host/port/user/
// password/sslmode, shared machinery) plus the target database name, which
// stage-import alone needs to carry separately since it juggles two DSNs
// (target DB_NAME vs stage STAGE_DB_NAME) built from the same ConnParams shape.
type dsn struct {
	dbconn.ConnParams
	name string
}

func (d dsn) connString() string {
	return d.DSN(d.name)
}

// connStringReadOnly opens the stage connection in read-only transaction mode so
// the importer can never write to the source DB even by mistake.
func (d dsn) connStringReadOnly() string {
	return d.connString() + " default_transaction_read_only=on"
}

// targetDSN reads the target connection from env and enforces the local-host guard.
func targetDSN() (dsn, error) {
	conn, err := dbconn.FromEnv()
	name := os.Getenv("DB_NAME")
	if err != nil || name == "" {
		return dsn{}, fmt.Errorf("missing required target env (DB_HOST, DB_USER, DB_PASSWORD, DB_NAME)")
	}
	if err := guardLocalTarget(conn.Host); err != nil {
		return dsn{}, err
	}
	return dsn{ConnParams: conn, name: name}, nil
}

// guardLocalTarget refuses any non-local target host. Extracted for testability.
func guardLocalTarget(host string) error {
	if !dbconn.IsLocalHost(host) {
		return fmt.Errorf("SAFETY: target DB_HOST=%q is not a known local host — refusing to import into a non-local DB", host)
	}
	return nil
}

// stageDSN reads the stage (source) connection. Defaults match old_db's
// docker-compose.postgres.yml (trust auth, db ani_legacy).
func stageDSN() dsn {
	return dsn{
		ConnParams: dbconn.ConnParams{
			Host:        dbconn.EnvOr("STAGE_DB_HOST", "old-db-postgres"),
			Port:        dbconn.EnvOr("STAGE_DB_PORT", "5432"),
			User:        dbconn.EnvOr("STAGE_DB_USER", "postgres"),
			Password:    os.Getenv("STAGE_DB_PASSWORD"),
			SSLMode:     dbconn.EnvOr("STAGE_DB_SSL_MODE", "disable"),
			SSLRootCert: os.Getenv("STAGE_DB_SSL_ROOT_CERT"),
		},
		name: dbconn.EnvOr("STAGE_DB_NAME", "ani_legacy"),
	}
}

// statusInClause builds the SQL `mapping_status IN (...)` fragment for the
// importable set. Centralised so the plan and the apply use the exact same filter.
func statusInClause() string {
	quoted := make([]string, len(importableStatuses))
	for i, s := range importableStatuses {
		quoted[i] = "'" + s + "'"
	}
	return "mapping_status IN (" + strings.Join(quoted, ", ") + ")"
}

// runIDClause optionally restricts to a single migration_run_id.
func runIDClause(runID string) string {
	if runID == "" {
		return ""
	}
	return fmt.Sprintf(" AND migration_run_id = %s", pqLiteral(runID))
}

// pqLiteral single-quotes a string literal, doubling embedded quotes. run-id is
// operator-supplied, so it is escaped rather than interpolated raw.
func pqLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// targetRefs are TARGET-side ids the stage schema does not carry but the本テーブル
// requires (NOT NULL, no default). Resolved once at startup; fail-fast if absent.
type targetRefs struct {
	clinicID                int64
	fallbackAnimalSpeciesID int64
	fallbackExamTypeID      int64
}

func resolveTargetRefs(ctx context.Context, pool *pgxpool.Pool) (targetRefs, error) {
	var r targetRefs
	if err := pool.QueryRow(ctx,
		`SELECT id FROM public.clinics WHERE name = $1 ORDER BY id LIMIT 1`, clinicName,
	).Scan(&r.clinicID); err != nil {
		return r, fmt.Errorf("clinic %q not found in target (run base seed first): %w", clinicName, err)
	}
	if err := pool.QueryRow(ctx,
		`SELECT id FROM public.animal_species ORDER BY id LIMIT 1`,
	).Scan(&r.fallbackAnimalSpeciesID); err != nil {
		return r, fmt.Errorf("no animal_species in target for pets.animal_species_id fallback: %w", err)
	}
	if err := pool.QueryRow(ctx,
		`SELECT id FROM public.exam_types WHERE name = $1 AND clinic_id = $2`,
		examTypeFallbackName, r.clinicID,
	).Scan(&r.fallbackExamTypeID); err != nil {
		return r, fmt.Errorf("exam_type %q not found for clinic %q in target (run base seed first): %w", examTypeFallbackName, clinicName, err)
	}
	return r, nil
}
