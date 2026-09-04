// Command csv-import-stg-uat is the fail-closed STG UAT rehearsal consumer for
// the twenty-one-table CSV hand-off. It does not pass cmd/csv-import
// --allow-local-rehearsal. Staging APP_ENV plus STG_UAT_CSV_IMPORT_ALLOW_REHEARSAL
// is required before source preflight or opening a target. REHEARSAL_ONLY
// _old_db_handoff bundles are accepted after that sentinel and an explicit
// staging target binding. Formal csv-import still requires TRUSTED_CANDIDATE.
// The import command sequences target preflight, apply, and verify in one
// process. Individual commands remain the manual fallback.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sys/unix"

	"github.com/animal-ekarte/backend/internal/csvimport"
	"github.com/animal-ekarte/backend/internal/dbconn"
)

const (
	allowRehearsalEnv      = "STG_UAT_CSV_IMPORT_ALLOW_REHEARSAL"
	allowRehearsalSentinel = "YES_I_UNDERSTAND"
	auditLane              = "stg-uat-rehearsal"
	auditReportRoot        = "/migration-reports"
)

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
}

type auditReport struct {
	Status         string                   `json:"status"`
	Lane           string                   `json:"lane"`
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
	pool *pgxpool.Pool
}

func (t *pgxCutoverTarget) Ping(ctx context.Context) error {
	return t.pool.Ping(ctx)
}

func (t *pgxCutoverTarget) Close() {
	t.pool.Close()
}

func (t *pgxCutoverTarget) Preflight(ctx context.Context, manifest csvimport.CutoverManifest, seeds csvimport.CutoverSeedIDs) error {
	return csvimport.PreflightCutoverTargetAllowingResume(ctx, t.pool, manifest, seeds)
}

func (t *pgxCutoverTarget) Verify(ctx context.Context, manifest csvimport.CutoverManifest, seeds csvimport.CutoverSeedIDs) error {
	tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		return fmt.Errorf("begin repeatable-read cutover verification: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if _, err := tx.Exec(ctx, `SELECT set_config('app.bypass_rls', 'on', true)`); err != nil {
		return fmt.Errorf("enable verification RLS bypass: %w", err)
	}
	return csvimport.VerifyCutoverWithProvenance(ctx, tx, manifest, seeds, csvimport.CutoverProvenanceContract{Mode: csvimport.CutoverProvenanceFormal})
}

func (t *pgxCutoverTarget) Apply(ctx context.Context, bundle csvimport.CutoverBundle, seeds csvimport.CutoverSeedIDs) (csvimport.CutoverResult, error) {
	return csvimport.ApplyCutoverCommittingEachTable(ctx, t.pool, bundle, seeds, applyIsolationLevel())
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
		configureTimeZone: configureTimeZone,
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

func configureTimeZone() error {
	loc, err := time.LoadLocation(dbconn.JapanTimeZone)
	if err != nil {
		return err
	}
	time.Local = loc
	return nil
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
	if err := requireStagingRehearsalEnv(); err != nil {
		return err
	}

	conn, err := deps.fromEnv()
	if err != nil {
		return err
	}
	database := os.Getenv("DB_NAME")
	if database == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if err := requireRemoteHostAllowed(conn.Host); err != nil {
		return err
	}
	if err := validateTargetConfirmations(opt, conn.Host, database); err != nil {
		return err
	}
	if opt.command == "apply" || opt.command == "import" {
		if err := validateApplyConfirmations(opt); err != nil {
			return err
		}
	}
	if err := validateImportSeedIDs(opt); err != nil {
		return err
	}
	if err := validateImportClinicIdentity(opt); err != nil {
		return err
	}

	bundle, err := deps.preflightBundle(opt.sourceDir, csvimport.ExpectedCutoverSource{
		ManifestSHA256: opt.manifestSHA256,
		ClinicCode:     opt.clinicCode,
		ClinicOrdinal:  opt.clinicOrdinal,
		RunID:          opt.runID,
		Provenance: csvimport.CutoverProvenanceContract{
			Mode:   csvimport.CutoverProvenanceStagingRehearsal,
			Target: csvimport.CutoverTargetBinding{Environment: "staging", Host: conn.Host, Database: database, ClinicID: opt.clinicID},
		},
	})
	if err != nil {
		return fmt.Errorf("source preflight failed: %w", err)
	}

	poolConfig, err := buildTargetPoolConfig(conn, database)
	if err != nil {
		return err
	}
	target, err := deps.openTarget(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("open target database: %w", err)
	}
	defer target.Close()
	if err := target.Ping(ctx); err != nil {
		return fmt.Errorf("ping target database: %w", err)
	}
	seeds := flagSeedIDs(opt)

	switch opt.command {
	case "preflight":
		if err := target.Preflight(ctx, bundle.Manifest, seeds); err != nil {
			return fmt.Errorf("target preflight failed: %w", err)
		}
		logger.Info("CSV STG UAT cutover preflight PASS", "clinic_code", bundle.Manifest.ClinicCode, "run_id", bundle.Manifest.SourceRunID, "tables", len(bundle.Manifest.Tables), "lane", auditLane)
		return nil
	case "verify":
		if err := target.Verify(ctx, bundle.Manifest, seeds); err != nil {
			return fmt.Errorf("cutover verification failed: %w", err)
		}
		logger.Info("CSV STG UAT cutover verification PASS", "clinic_code", bundle.Manifest.ClinicCode, "run_id", bundle.Manifest.SourceRunID, "lane", auditLane)
		return nil
	case "apply":
		return applyWithAudit(ctx, target, bundle, seeds, opt, conn.Host, database, deps.reportRoot, logger)
	case "import":
		return importWithAudit(ctx, target, bundle, seeds, opt, conn.Host, database, deps.reportRoot, logger)
	default:
		return fmt.Errorf("unsupported command %q", opt.command)
	}
}

func requireStagingRehearsalEnv() error {
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if appEnv == "production" {
		return fmt.Errorf("csv-import-stg-uat refuses APP_ENV=production")
	}
	if appEnv != "staging" {
		return fmt.Errorf("csv-import-stg-uat requires APP_ENV=staging")
	}
	if os.Getenv(allowRehearsalEnv) != allowRehearsalSentinel {
		return fmt.Errorf("csv-import-stg-uat refuses REHEARSAL_ONLY without %s=%s", allowRehearsalEnv, allowRehearsalSentinel)
	}
	return nil
}

func requireRemoteHostAllowed(host string) error {
	if dbconn.IsLocalHost(host) {
		return nil
	}
	if os.Getenv(allowRehearsalEnv) != allowRehearsalSentinel {
		return fmt.Errorf("csv-import-stg-uat refuses non-local DB_HOST without %s=%s", allowRehearsalEnv, allowRehearsalSentinel)
	}
	return nil
}

func applyIsolationLevel() pgx.TxIsoLevel {
	return pgx.RepeatableRead
}

func buildTargetPoolConfig(conn dbconn.ConnParams, database string) (*pgxpool.Config, error) {
	if err := validateSSLMode(conn.SSLMode, dbconn.IsLocalHost(conn.Host)); err != nil {
		return nil, err
	}
	pgxConfig, err := conn.PGXConfig(database)
	if err != nil {
		return nil, err
	}
	// ParseConfig must see a TCP host while TLS is configured. A keyword-only
	// DSN such as "sslmode=verify-full" defaults to a unix socket, which pgx
	// treats as TLS-off; overwriting Host afterwards would speak plaintext to
	// PlanetScale and fail with FATAL: SSL/TLS required.
	poolConfig, err := pgxpool.ParseConfig("postgres://placeholder.invalid/placeholder")
	if err != nil {
		return nil, fmt.Errorf("initialize target database configuration: %w", err)
	}
	poolConfig.ConnConfig = pgxConfig
	if poolConfig.ConnConfig.Host != conn.Host ||
		poolConfig.ConnConfig.Database != database ||
		len(poolConfig.ConnConfig.Fallbacks) != 0 {
		return nil, fmt.Errorf("effective target database identity failed validation")
	}
	// PlanetScale user-defined roles are not the table owner. RLS is ENABLE
	// without FORCE, so non-owner connections see zero clinic rows unless
	// app.bypass_rls is on (001_init.sql). AfterConnect applies to every pool
	// connection, including preflight QueryRow on the pool.
	poolConfig.AfterConnect = enableTargetRLSBypass
	if !dbconn.IsLocalHost(conn.Host) {
		keepAlive := 30 * time.Second
		poolConfig.ConnConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := net.Dialer{KeepAlive: keepAlive}
			return d.DialContext(ctx, network, addr)
		}
	}
	return poolConfig, nil
}

func enableTargetRLSBypass(ctx context.Context, conn *pgx.Conn) error {
	if _, err := conn.Exec(ctx, `SELECT set_config('app.bypass_rls', 'on', false)`); err != nil {
		return fmt.Errorf("enable app.bypass_rls for non-owner STG role: %w", err)
	}
	return nil
}

func validateSSLMode(mode string, local bool) error {
	if local {
		if mode != "disable" {
			return fmt.Errorf("csv-import-stg-uat local/container connection requires DB_SSL_MODE=disable")
		}
		return nil
	}
	switch mode {
	case "require", "verify-ca", "verify-full":
		return nil
	default:
		return fmt.Errorf("csv-import-stg-uat remote DB_SSL_MODE must be require, verify-ca, or verify-full")
	}
}

func parseOptions(args []string) (options, error) {
	if len(args) == 0 {
		return options{}, fmt.Errorf("command is required: preflight, apply, verify, or import")
	}
	opt := options{command: args[0]}
	if opt.command != "preflight" && opt.command != "apply" && opt.command != "verify" && opt.command != "import" {
		return options{}, fmt.Errorf("unsupported command %q: use preflight, apply, verify, or import", opt.command)
	}
	flags := flag.NewFlagSet("csv-import-stg-uat "+opt.command, flag.ContinueOnError)
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
	flags.StringVar(&opt.confirmTargetHost, "confirm-target-host", "", "required before connect; must exactly equal DB_HOST")
	flags.StringVar(&opt.confirmTargetDatabase, "confirm-target-database", "", "required before connect; must exactly equal DB_NAME")
	flags.StringVar(&opt.reportPath, "report-path", "", "required absolute path for the owner-only aggregate apply report")
	if err := flags.Parse(args[1:]); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected positional arguments")
	}
	return opt, nil
}

func validateTargetConfirmations(opt options, targetHost, targetDatabase string) error {
	if opt.confirmTargetHost == "" || opt.confirmTargetHost != targetHost {
		return fmt.Errorf("target host confirmation must exactly match DB_HOST")
	}
	if opt.confirmTargetDatabase == "" || opt.confirmTargetDatabase != targetDatabase {
		return fmt.Errorf("target database confirmation must exactly match DB_NAME")
	}
	return nil
}

func validateApplyConfirmations(opt options) error {
	if !opt.confirmTargetWrite {
		return fmt.Errorf("apply requires --confirm-target-write")
	}
	if !opt.confirmBackupReady {
		return fmt.Errorf("apply requires --confirm-backup-ready")
	}
	if opt.reportPath == "" || !filepath.IsAbs(opt.reportPath) {
		return fmt.Errorf("apply requires an absolute --report-path")
	}
	return nil
}

func flagSeedIDs(opt options) csvimport.CutoverSeedIDs {
	return csvimport.CutoverSeedIDs{
		ClinicID:                  opt.clinicID,
		AnimalSpeciesID:           opt.animalSpeciesID,
		ExamTypeID:                opt.examTypeID,
		TrimmingReservationTypeID: opt.trimmingReservationTypeID,
		CashPaymentMethodID:       opt.cashPaymentMethodID,
		CreditCardPaymentMethodID: opt.creditCardPaymentMethodID,
	}
}

func validateImportSeedIDs(opt options) error {
	if opt.command != "import" {
		return nil
	}
	for _, id := range []int64{
		opt.clinicID,
		opt.animalSpeciesID,
		opt.examTypeID,
		opt.trimmingReservationTypeID,
		opt.cashPaymentMethodID,
		opt.creditCardPaymentMethodID,
	} {
		if id <= 0 {
			return fmt.Errorf("import requires six explicit seed IDs")
		}
	}
	return nil
}

const (
	stgUATClinicHachiojiID   int64 = 1
	stgUATClinicJoutoID      int64 = 2
	stgUATClinicShikishimaID int64 = 3
	stgUATClinicHakobunecoID int64 = 4
)

func expectedSTGUATClinic(code string, ordinal int64) (int64, error) {
	switch {
	case code == "hachioji" && ordinal == 1:
		return stgUATClinicHachiojiID, nil
	case code == "jouto" && ordinal == 2:
		return stgUATClinicJoutoID, nil
	case code == "shikishima" && ordinal == 3:
		return stgUATClinicShikishimaID, nil
	case code == "hakobuneco" && ordinal == 4:
		return stgUATClinicHakobunecoID, nil
	default:
		return 0, fmt.Errorf("import clinic-code/ordinal is not a STG UAT clinic binding")
	}
}

func validateImportClinicIdentity(opt options) error {
	if opt.command != "import" {
		return nil
	}
	expectedID, err := expectedSTGUATClinic(opt.clinicCode, opt.clinicOrdinal)
	if err != nil {
		return err
	}
	if opt.clinicID != expectedID {
		return fmt.Errorf("import clinic-id must match clinic-code/ordinal binding")
	}
	return nil
}

func importWithAudit(
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
	if err := target.Preflight(ctx, bundle.Manifest, seeds); err != nil {
		return fmt.Errorf("target preflight failed: %w", err)
	}
	err := applyWithAuditFinalizer(ctx, target, bundle, seeds, opt, targetHost, targetDatabase, reportRoot, logger,
		func(reportFile *os.File, report *auditReport, result csvimport.CutoverResult) error {
			report.Status = "APPLIED_PENDING_VERIFY"
			report.CompletedAt = &result.CompletedAt
			report.Counts = result.Counts
			if err := replaceAuditReport(reportFile, *report); err != nil {
				return fmt.Errorf("cutover committed but pending-verification audit report update failed; run verify immediately: %w", err)
			}
			if err := target.Verify(ctx, bundle.Manifest, seeds); err != nil {
				report.Status = "FAILED_POST_COMMIT_VERIFY"
				report.FailureStage = "verify"
				if writeErr := replaceAuditReport(reportFile, *report); writeErr != nil {
					return fmt.Errorf("cutover verification failed after committed apply and audit report update also failed: %w", err)
				}
				return fmt.Errorf("cutover verification failed after committed apply: %w", err)
			}
			report.Status = "PASS"
			report.FailureStage = ""
			if err := replaceAuditReport(reportFile, *report); err != nil {
				return fmt.Errorf("cutover verified but final audit report update failed; inspect target before any retry: %w", err)
			}
			logger.Info("CSV STG UAT cutover import PASS", "clinic_code", result.ClinicCode, "run_id", result.RunID, "report_path", opt.reportPath, "lane", auditLane)
			return nil
		},
	)
	if err != nil && errors.Is(err, csvimport.ErrCutoverCommitOutcomeUnknown) {
		if verifyErr := target.Verify(ctx, bundle.Manifest, seeds); verifyErr != nil {
			return fmt.Errorf("%w; verify also failed: %w", err, verifyErr)
		}
	}
	return err
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
	return applyWithAuditFinalizer(ctx, target, bundle, seeds, opt, targetHost, targetDatabase, reportRoot, logger,
		func(reportFile *os.File, report *auditReport, result csvimport.CutoverResult) error {
			report.Status = "PASS"
			report.CompletedAt = &result.CompletedAt
			report.Counts = result.Counts
			if err := replaceAuditReport(reportFile, *report); err != nil {
				return fmt.Errorf("cutover committed but final audit report update failed; run verify immediately: %w", err)
			}
			logger.Info("CSV STG UAT cutover apply PASS", "clinic_code", result.ClinicCode, "run_id", result.RunID, "report_path", opt.reportPath, "lane", auditLane)
			return nil
		},
	)
}

type auditApplyFinalizer func(*os.File, *auditReport, csvimport.CutoverResult) error

func applyWithAuditFinalizer(
	ctx context.Context,
	target cutoverTarget,
	bundle csvimport.CutoverBundle,
	seeds csvimport.CutoverSeedIDs,
	opt options,
	targetHost string,
	targetDatabase string,
	reportRoot string,
	logger *slog.Logger,
	finalize auditApplyFinalizer,
) error {
	started := time.Now().UTC()
	report := auditReport{
		Status:         "STARTED",
		Lane:           auditLane,
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

	logger.Info("csv-import-stg-uat apply committing each table to bound PlanetScale transaction size", "lane", auditLane)
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
		return fmt.Errorf("cutover apply failed; uncommitted table rolled back: %w", err)
	}
	return finalize(reportFile, &report, result)
}

func cutoverApplyFailureClassification(err error) (status string, stage string) {
	if errors.Is(err, csvimport.ErrCutoverCommitOutcomeUnknown) {
		return "COMMIT_OUTCOME_UNKNOWN", "commit"
	}
	if errors.Is(err, csvimport.ErrCutoverTransactionNotStarted) {
		return "FAILED_BEFORE_TRANSACTION", "begin"
	}
	return "FAILED_TABLE_ROLLED_BACK", "table"
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
