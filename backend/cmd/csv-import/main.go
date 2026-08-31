// Command csv-import is the fail-closed consumer for the old_db twenty-one-table
// CSV hand-off. It has no old_db database/network dependency: the source is an
// operator-selected read-only directory plus a trusted manifest SHA-256.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sys/unix"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/csvimport"
	"github.com/animal-ekarte/backend/internal/dbconn"
)

const auditReportRoot = "/migration-reports"

type options struct {
	command                   string
	sourceDir                 string
	manifestSHA256            string
	clinicCode                string
	clinicOrdinal             int64
	runID                     string
	clinicID                  int64
	animalSpeciesID           int64
	examTypeID                int64
	trimmingReservationTypeID int64
	cashPaymentMethodID       int64
	creditCardPaymentMethodID int64
	confirmTargetWrite        bool
	confirmBackupReady        bool
	confirmTargetHost         string
	confirmTargetDatabase     string
	reportPath                string
	allowLocalRehearsal       bool
}

type auditReport struct {
	Status         string                   `json:"status"`
	StartedAt      time.Time                `json:"startedAt"`
	CompletedAt    *time.Time               `json:"completedAt,omitempty"`
	ManifestSHA256 string                   `json:"manifestSha256"`
	ClinicCode     string                   `json:"clinicCode"`
	ClinicOrdinal  int64                    `json:"clinicOrdinal"`
	RunID          string                   `json:"runId"`
	TargetHost     string                   `json:"targetHost"`
	TargetDatabase string                   `json:"targetDatabase"`
	SeedIDs        csvimport.CutoverSeedIDs `json:"seedIds"`
	IDBand         csvimport.CutoverIDBand  `json:"idBand"`
	Counts         map[string]int64         `json:"counts,omitempty"`
	FailureStage   string                   `json:"failureStage,omitempty"`
}

type cutoverTarget interface {
	Ping(context.Context) error
	Close()
	Preflight(context.Context, csvimport.CutoverManifest, csvimport.CutoverSeedIDs) error
	Verify(context.Context, csvimport.CutoverManifest, csvimport.CutoverSeedIDs) error
	Apply(context.Context, csvimport.CutoverBundle, csvimport.CutoverSeedIDs) (csvimport.CutoverResult, error)
}

type pgxCutoverTarget struct {
	pool           *pgxpool.Pool
	localRehearsal bool
}

func (t *pgxCutoverTarget) Ping(ctx context.Context) error {
	return t.pool.Ping(ctx)
}

func (t *pgxCutoverTarget) Close() {
	t.pool.Close()
}

func (t *pgxCutoverTarget) Preflight(ctx context.Context, manifest csvimport.CutoverManifest, seeds csvimport.CutoverSeedIDs) error {
	return csvimport.PreflightCutoverTarget(ctx, t.pool, manifest, seeds)
}

func (t *pgxCutoverTarget) Verify(ctx context.Context, manifest csvimport.CutoverManifest, seeds csvimport.CutoverSeedIDs) error {
	tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		return fmt.Errorf("begin repeatable-read cutover verification: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	return csvimport.VerifyCutover(ctx, tx, manifest, seeds)
}

func (t *pgxCutoverTarget) Apply(ctx context.Context, bundle csvimport.CutoverBundle, seeds csvimport.CutoverSeedIDs) (csvimport.CutoverResult, error) {
	iso := pgx.Serializable
	if t.localRehearsal {
		// Avoid SSI predicate-lock exhaustion on multi-million-row local imports.
		iso = pgx.RepeatableRead
	}
	return csvimport.ApplyCutoverWithIsolation(ctx, t.pool, bundle, seeds, iso)
}

type runDependencies struct {
	configureTimeZone func() error
	preflightBundle   func(string, csvimport.ExpectedCutoverSource) (csvimport.CutoverBundle, error)
	fromEnv           func() (dbconn.ConnParams, error)
	openTarget        func(context.Context, *pgxpool.Config) (cutoverTarget, error)
	reportRoot        string
}

func productionRunDependencies() runDependencies {
	return runDependencies{
		configureTimeZone: config.ConfigureTimeZone,
		preflightBundle:   csvimport.PreflightCutoverBundle,
		fromEnv:           dbconn.FromEnv,
		openTarget: func(ctx context.Context, poolConfig *pgxpool.Config) (cutoverTarget, error) {
			pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
			if err != nil {
				return nil, err
			}
			return &pgxCutoverTarget{pool: pool}, nil
		},
		reportRoot: auditReportRoot,
	}
}

func cutoverProvenanceMode(local bool) csvimport.CutoverProvenanceMode {
	if local {
		return csvimport.CutoverProvenanceLocalRehearsal
	}
	return csvimport.CutoverProvenanceFormal
}

func openCutoverTarget(ctx context.Context, poolConfig *pgxpool.Config, localRehearsal bool) (cutoverTarget, error) {
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}
	return &pgxCutoverTarget{pool: pool, localRehearsal: localRehearsal}, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(context.Background(), os.Args[1:], logger); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, logger *slog.Logger) error {
	return runWithDependencies(ctx, args, logger, productionRunDependencies())
}

func runWithDependencies(ctx context.Context, args []string, logger *slog.Logger, deps runDependencies) error {
	if err := deps.configureTimeZone(); err != nil {
		return fmt.Errorf("timezone configuration failed: %w", err)
	}
	opt, err := parseOptions(args)
	if err != nil {
		return err
	}
	if opt.allowLocalRehearsal {
		if err := requireLocalRehearsalEnv(); err != nil {
			return err
		}
	}

	bundle, err := deps.preflightBundle(opt.sourceDir, csvimport.ExpectedCutoverSource{
		ManifestSHA256: opt.manifestSHA256,
		ClinicCode:     opt.clinicCode,
		ClinicOrdinal:  opt.clinicOrdinal,
		RunID:          opt.runID,
		Provenance:     csvimport.CutoverProvenanceContract{Mode: cutoverProvenanceMode(opt.allowLocalRehearsal)},
	})
	if err != nil {
		return fmt.Errorf("source preflight failed: %w", err)
	}

	conn, err := deps.fromEnv()
	if err != nil {
		return err
	}
	if opt.allowLocalRehearsal && !dbconn.IsLocalHost(conn.Host) {
		return fmt.Errorf("local rehearsal import requires a local DB_HOST")
	}
	database := os.Getenv("DB_NAME")
	if database == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if opt.command == "apply" {
		if err := validateApplyConfirmations(opt, conn.Host, database); err != nil {
			return err
		}
	}

	poolConfig, err := buildTargetPoolConfig(conn, database)
	if err != nil {
		return err
	}
	var target cutoverTarget
	if opt.allowLocalRehearsal {
		target, err = openCutoverTarget(ctx, poolConfig, true)
	} else {
		target, err = deps.openTarget(ctx, poolConfig)
	}
	if err != nil {
		return fmt.Errorf("open target database: %w", err)
	}
	defer target.Close()
	if err := target.Ping(ctx); err != nil {
		return fmt.Errorf("ping target database: %w", err)
	}
	seeds := csvimport.CutoverSeedIDs{
		ClinicID:                  opt.clinicID,
		AnimalSpeciesID:           opt.animalSpeciesID,
		ExamTypeID:                opt.examTypeID,
		TrimmingReservationTypeID: opt.trimmingReservationTypeID,
		CashPaymentMethodID:       opt.cashPaymentMethodID,
		CreditCardPaymentMethodID: opt.creditCardPaymentMethodID,
	}

	switch opt.command {
	case "preflight":
		if err := target.Preflight(ctx, bundle.Manifest, seeds); err != nil {
			return fmt.Errorf("target preflight failed: %w", err)
		}
		logger.Info("CSV cutover preflight PASS", "clinic_code", bundle.Manifest.ClinicCode, "run_id", bundle.Manifest.SourceRunID, "tables", len(bundle.Manifest.Tables))
		return nil
	case "verify":
		if err := target.Verify(ctx, bundle.Manifest, seeds); err != nil {
			return fmt.Errorf("cutover verification failed: %w", err)
		}
		logger.Info("CSV cutover verification PASS", "clinic_code", bundle.Manifest.ClinicCode, "run_id", bundle.Manifest.SourceRunID)
		return nil
	case "apply":
		return applyWithAudit(ctx, target, bundle, seeds, opt, conn.Host, database, deps.reportRoot, logger)
	default:
		return fmt.Errorf("unsupported command %q", opt.command)
	}
}

func buildTargetPoolConfig(conn dbconn.ConnParams, database string) (*pgxpool.Config, error) {
	if !dbconn.IsLocalHost(conn.Host) {
		return nil, fmt.Errorf("csv-import only accepts a local/container target host")
	}
	if conn.SSLMode != "disable" {
		return nil, fmt.Errorf("csv-import local/container connection requires DB_SSL_MODE=disable")
	}
	port, err := strconv.ParseUint(conn.Port, 10, 16)
	if err != nil || port == 0 {
		return nil, fmt.Errorf("DB_PORT must be an integer between 1 and 65535")
	}
	poolConfig, err := pgxpool.ParseConfig("sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("initialize target database configuration: %w", err)
	}
	// Assign fields structurally. Interpolating environment values into a
	// libpq keyword DSN would let whitespace in a password/database inject a
	// second host option after the local-host guard.
	poolConfig.ConnConfig.Host = conn.Host
	poolConfig.ConnConfig.Port = uint16(port)
	poolConfig.ConnConfig.User = conn.User
	poolConfig.ConnConfig.Password = conn.Password
	poolConfig.ConnConfig.Database = database
	// ParseConfig also reads libpq PGHOST/PGPORT/PGSERVICE environment. Never
	// retain environment-derived remote fallback hosts or runtime options (for
	// example a search_path override) for this local-only tool.
	poolConfig.ConnConfig.Fallbacks = nil
	poolConfig.ConnConfig.RuntimeParams = map[string]string{"TimeZone": config.JapanTimeZone}
	if !dbconn.IsLocalHost(poolConfig.ConnConfig.Host) ||
		poolConfig.ConnConfig.Database != database ||
		len(poolConfig.ConnConfig.Fallbacks) != 0 {
		return nil, fmt.Errorf("effective target database identity failed validation")
	}
	return poolConfig, nil
}

func parseOptions(args []string) (options, error) {
	if len(args) == 0 {
		return options{}, fmt.Errorf("command is required: preflight, apply, or verify")
	}
	opt := options{command: args[0]}
	if opt.command != "preflight" && opt.command != "apply" && opt.command != "verify" {
		return options{}, fmt.Errorf("unsupported command %q: use preflight, apply, or verify", opt.command)
	}
	flags := flag.NewFlagSet("csv-import "+opt.command, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&opt.sourceDir, "source-dir", "", "absolute directory containing manifest.json and twenty-one CSV files")
	flags.StringVar(&opt.manifestSHA256, "expected-manifest-sha256", "", "trusted producer manifest SHA-256")
	flags.StringVar(&opt.clinicCode, "clinic-code", "", "expected producer clinic code")
	flags.Int64Var(&opt.clinicOrdinal, "clinic-ordinal", 0, "expected producer clinic ordinal (1..50)")
	flags.StringVar(&opt.runID, "run-id", "", "expected producer migration run ID")
	flags.Int64Var(&opt.clinicID, "clinic-id", 0, "explicit target clinic seed ID")
	flags.Int64Var(&opt.animalSpeciesID, "fallback-animal-species-id", 0, "explicit active target animal species ID")
	flags.Int64Var(&opt.examTypeID, "fallback-exam-type-id", 0, "explicit target clinic exam type ID named 検査")
	flags.Int64Var(&opt.trimmingReservationTypeID, "trimming-reservation-type-id", 0, "explicit active target clinic reservation type ID in category trimming")
	flags.Int64Var(&opt.cashPaymentMethodID, "cash-payment-method-id", 0, "explicit active target clinic payment method ID with system_key cash")
	flags.Int64Var(&opt.creditCardPaymentMethodID, "credit-card-payment-method-id", 0, "explicit active target clinic payment method ID with system_key credit_card")
	flags.BoolVar(&opt.confirmTargetWrite, "confirm-target-write", false, "required for apply")
	flags.BoolVar(&opt.confirmBackupReady, "confirm-backup-ready", false, "required for apply; confirms a tested pre-import backup exists")
	flags.StringVar(&opt.confirmTargetHost, "confirm-target-host", "", "required for apply; must exactly equal DB_HOST")
	flags.StringVar(&opt.confirmTargetDatabase, "confirm-target-database", "", "required for apply; must exactly equal DB_NAME")
	flags.StringVar(&opt.reportPath, "report-path", "", "required absolute path for the owner-only aggregate apply report")
	flags.BoolVar(&opt.allowLocalRehearsal, "allow-local-rehearsal", false, "local disposable only: accept REHEARSAL_ONLY/UNVERIFIED producer bundles")
	if err := flags.Parse(args[1:]); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected positional arguments")
	}
	return opt, nil
}

func requireLocalRehearsalEnv() error {
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	switch appEnv {
	case "development", "local", "dev", "test":
	default:
		return fmt.Errorf("local rehearsal import requires APP_ENV development/local/dev/test")
	}
	if os.Getenv("CSV_IMPORT_ALLOW_LOCAL_REHEARSAL") != "1" {
		return fmt.Errorf("local rehearsal import requires CSV_IMPORT_ALLOW_LOCAL_REHEARSAL=1")
	}
	return nil
}

func validateApplyConfirmations(opt options, targetHost, targetDatabase string) error {
	if !opt.confirmTargetWrite {
		return fmt.Errorf("apply requires --confirm-target-write")
	}
	if !opt.confirmBackupReady {
		return fmt.Errorf("apply requires --confirm-backup-ready")
	}
	if opt.confirmTargetHost == "" || opt.confirmTargetHost != targetHost {
		return fmt.Errorf("apply target host confirmation must exactly match DB_HOST")
	}
	if opt.confirmTargetDatabase == "" || opt.confirmTargetDatabase != targetDatabase {
		return fmt.Errorf("apply target database confirmation must exactly match DB_NAME")
	}
	if opt.reportPath == "" || !filepath.IsAbs(opt.reportPath) {
		return fmt.Errorf("apply requires an absolute --report-path")
	}
	return nil
}

func applyWithAudit(
	ctx context.Context,
	target cutoverTarget,
	bundle csvimport.CutoverBundle,
	seeds csvimport.CutoverSeedIDs,
	opt options,
	targetHost string,
	targetDatabase string,
	reportRoot string,
	logger *slog.Logger,
) error {
	started := time.Now().UTC()
	report := auditReport{
		Status:         "STARTED",
		StartedAt:      started,
		ManifestSHA256: strings.ToLower(opt.manifestSHA256),
		ClinicCode:     bundle.Manifest.ClinicCode,
		ClinicOrdinal:  bundle.Manifest.ClinicOrdinal,
		RunID:          bundle.Manifest.SourceRunID,
		TargetHost:     targetHost,
		TargetDatabase: targetDatabase,
		SeedIDs:        seeds,
		IDBand:         bundle.Manifest.IDBand,
	}
	reportFile, err := createAuditReport(opt.reportPath, reportRoot, report)
	if err != nil {
		return err
	}
	defer func() { _ = reportFile.Close() }()

	result, err := target.Apply(ctx, bundle, seeds)
	if err != nil {
		report.Status, report.FailureStage = cutoverApplyFailureClassification(err)
		if writeErr := replaceAuditReport(reportFile, report); writeErr != nil {
			return fmt.Errorf("cutover failed and audit report update also failed; apply status=%s; report update: %w", report.Status, writeErr)
		}
		if errors.Is(err, csvimport.ErrCutoverCommitOutcomeUnknown) {
			return fmt.Errorf("cutover commit outcome is unknown; run read-only verify before any retry or restore: %w", err)
		}
		if errors.Is(err, csvimport.ErrCutoverTransactionNotStarted) {
			return fmt.Errorf("cutover apply failed before transaction began: %w", err)
		}
		return fmt.Errorf("cutover apply failed; transaction rolled back: %w", err)
	}
	report.Status = "PASS"
	report.CompletedAt = &result.CompletedAt
	report.Counts = result.Counts
	if err := replaceAuditReport(reportFile, report); err != nil {
		return fmt.Errorf("cutover committed but final audit report update failed; run verify immediately: %w", err)
	}
	logger.Info("CSV cutover apply PASS", "clinic_code", result.ClinicCode, "run_id", result.RunID, "report_path", opt.reportPath)
	return nil
}

func cutoverApplyFailureClassification(err error) (status string, stage string) {
	if errors.Is(err, csvimport.ErrCutoverCommitOutcomeUnknown) {
		return "COMMIT_OUTCOME_UNKNOWN", "commit"
	}
	if errors.Is(err, csvimport.ErrCutoverTransactionNotStarted) {
		return "FAILED_BEFORE_TRANSACTION", "begin"
	}
	return "FAILED_DATA_ROLLED_BACK", "transaction"
}

func createAuditReport(path string, root string, report auditReport) (*os.File, error) {
	if !pathWithinRoot(root, path) {
		return nil, fmt.Errorf("audit report path must stay under %s", root)
	}
	cleanRoot := filepath.Clean(root)
	resolvedRoot, err := filepath.EvalSymlinks(cleanRoot)
	if err != nil || resolvedRoot != cleanRoot {
		return nil, fmt.Errorf("audit report directory must not contain symbolic links")
	}
	info, err := os.Lstat(cleanRoot)
	if err != nil {
		return nil, fmt.Errorf("inspect audit report directory: %w", err)
	}
	if !info.IsDir() || info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("audit report directory must be owner-only")
	}
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_EXCL|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0o600) //nolint:gosec // fixed-root operator path, no-follow and exclusive owner-only creation
	if err != nil {
		return nil, fmt.Errorf("create audit report without overwrite: %w", err)
	}
	file := os.NewFile(uintptr(fd), filepath.Base(path))
	if file == nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("create audit report: invalid file descriptor")
	}
	if err := replaceAuditReport(file, report); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}

func pathWithinRoot(root string, candidate string) bool {
	if !filepath.IsAbs(root) || !filepath.IsAbs(candidate) {
		return false
	}
	cleanRoot := filepath.Clean(root)
	cleanCandidate := filepath.Clean(candidate)
	return cleanCandidate != cleanRoot && filepath.Dir(cleanCandidate) == cleanRoot
}

func replaceAuditReport(file *os.File, report auditReport) error {
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("seek audit report: %w", err)
	}
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("truncate audit report: %w", err)
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("write audit report: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync audit report: %w", err)
	}
	return nil
}
