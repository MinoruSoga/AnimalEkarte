// Command csv-import-failure-rehearsal proves one fixed synthetic G4
// transaction rollback against a disposable localhost-only database.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/csvimport"
	"github.com/animal-ekarte/backend/internal/dbconn"
)

const disposableConfirmation = "F8_G4_DISPOSABLE_ONLY"
const executionTimeout = 5 * time.Minute

var disposableDatabasePattern = regexp.MustCompile(`^animalekarte_f8_g4_[a-z0-9_]+$`)

type options struct {
	command                      string
	clinicCode                   string
	clinicOrdinal                int64
	runID                        string
	targetReleaseCommit          string
	targetDatabaseIdentitySHA256 string
	fixtureManifestSHA256        string
	failureRuntimeReportSHA256   string
	clinicID                     int64
	animalSpeciesID              int64
	examTypeID                   int64
	trimmingReservationTypeID    int64
	cashPaymentMethodID          int64
	creditCardPaymentMethodID    int64
	confirmTargetDatabase        string
	confirmDisposableRehearsal   string
}

type executionReceipt struct {
	SchemaVersion                int                           `json:"schemaVersion"`
	Status                       string                        `json:"status"`
	ClinicCode                   string                        `json:"clinicCode"`
	ClinicOrdinal                int64                         `json:"clinicOrdinal"`
	RunID                        string                        `json:"runId"`
	TargetReleaseCommit          string                        `json:"targetReleaseCommit"`
	TargetDatabaseIdentitySHA256 string                        `json:"targetDatabaseIdentitySha256"`
	FixtureManifestSHA256        string                        `json:"fixtureManifestSha256"`
	FailureRuntimeReportSHA256   string                        `json:"failureRuntimeReportSha256"`
	ExecutionMode                string                        `json:"executionMode"`
	FailureCheckpoint            string                        `json:"failureCheckpoint"`
	FailureStage                 string                        `json:"failureStage"`
	InjectionMarker              string                        `json:"injectionMarker"`
	StartedAt                    string                        `json:"startedAt"`
	FailureInjectedAt            string                        `json:"failureInjectedAt"`
	CompletedAt                  string                        `json:"completedAt"`
	TransactionStarted           bool                          `json:"transactionStarted"`
	TransactionRolledBack        bool                          `json:"transactionRolledBack"`
	BeforeBandCountsSHA256       string                        `json:"beforeBandCountsSha256"`
	AfterBandCountsSHA256        string                        `json:"afterBandCountsSha256"`
	BandRowCountBefore           int64                         `json:"bandRowCountBefore"`
	BandRowCountAfter            int64                         `json:"bandRowCountAfter"`
	TransactionEvidenceSHA256    string                        `json:"transactionEvidenceSha256"`
	BeforeBandCounts             csvimport.BandCountsEvidence  `json:"beforeBandCounts"`
	AfterBandCounts              csvimport.BandCountsEvidence  `json:"afterBandCounts"`
	TransactionEvidence          csvimport.TransactionEvidence `json:"transactionEvidence"`
	ProductionEligible           bool                          `json:"productionEligible"`
}

type failureTarget interface {
	Ping(context.Context) error
	Close()
}

type pgxFailureTarget struct {
	pool *pgxpool.Pool
}

func (target *pgxFailureTarget) Ping(ctx context.Context) error {
	return target.pool.Ping(ctx)
}

func (target *pgxFailureTarget) Close() {
	target.pool.Close()
}

type runDependencies struct {
	configureTimeZone func() error
	fromEnv           func() (dbconn.ConnParams, error)
	openTarget        func(context.Context, *pgxpool.Config) (failureTarget, error)
	runRehearsal      func(context.Context, failureTarget, csvimport.SyntheticFailureInput) (csvimport.SyntheticFailureResult, error)
}

func productionRunDependencies() runDependencies {
	return runDependencies{
		configureTimeZone: config.ConfigureTimeZone,
		fromEnv:           dbconn.FromEnv,
		openTarget: func(ctx context.Context, poolConfig *pgxpool.Config) (failureTarget, error) {
			pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
			if err != nil {
				return nil, err
			}
			return &pgxFailureTarget{pool: pool}, nil
		},
		runRehearsal: func(
			ctx context.Context,
			target failureTarget,
			input csvimport.SyntheticFailureInput,
		) (csvimport.SyntheticFailureResult, error) {
			pgxTarget, ok := target.(*pgxFailureTarget)
			if !ok {
				return csvimport.SyntheticFailureResult{}, fmt.Errorf("invalid disposable target")
			}
			return csvimport.RunSyntheticFailureRehearsal(ctx, pgxTarget.pool, input)
		},
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), executionTimeout)
	defer cancel()
	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, output io.Writer) error {
	return runWithDependencies(ctx, args, output, productionRunDependencies())
}

func runWithDependencies(
	ctx context.Context,
	args []string,
	output io.Writer,
	deps runDependencies,
) error {
	if err := deps.configureTimeZone(); err != nil {
		return fmt.Errorf("timezone configuration failed: %w", err)
	}
	opt, err := parseOptions(args)
	if err != nil {
		return err
	}
	conn, err := deps.fromEnv()
	if err != nil {
		return err
	}
	database := os.Getenv("DB_NAME")
	if err := validateDisposableTarget(opt, conn, database); err != nil {
		return err
	}
	poolConfig, err := buildDisposablePoolConfig(conn, database)
	if err != nil {
		return err
	}
	target, err := deps.openTarget(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("open disposable target database: %w", err)
	}
	defer target.Close()
	if err := target.Ping(ctx); err != nil {
		return fmt.Errorf("ping disposable target database: %w", err)
	}

	input := syntheticFailureInput(opt)
	result, err := deps.runRehearsal(ctx, target, input)
	if err != nil {
		return fmt.Errorf("synthetic G4 rehearsal failed: %w", err)
	}
	receipt := executionReceipt{
		SchemaVersion:                1,
		Status:                       "FAILED_DATA_ROLLED_BACK",
		ClinicCode:                   input.ClinicCode,
		ClinicOrdinal:                input.ClinicOrdinal,
		RunID:                        input.RunID,
		TargetReleaseCommit:          input.TargetReleaseCommit,
		TargetDatabaseIdentitySHA256: input.TargetDatabaseIdentitySHA256,
		FixtureManifestSHA256:        input.FixtureManifestSHA256,
		FailureRuntimeReportSHA256:   input.FailureRuntimeReportSHA256,
		ExecutionMode:                "SYNTHETIC_FAILURE_REHEARSAL",
		FailureCheckpoint:            csvimport.SyntheticFailureCheckpoint,
		FailureStage:                 csvimport.SyntheticFailureStage,
		InjectionMarker:              csvimport.SyntheticFailureMarker,
		StartedAt:                    evidenceTimestamp(result.StartedAt),
		FailureInjectedAt:            evidenceTimestamp(result.FailureInjectedAt),
		CompletedAt:                  evidenceTimestamp(result.CompletedAt),
		TransactionStarted:           true,
		TransactionRolledBack:        true,
		BeforeBandCountsSHA256:       result.BeforeBandCountsSHA256,
		AfterBandCountsSHA256:        result.AfterBandCountsSHA256,
		BandRowCountBefore:           result.BandRowCountBefore,
		BandRowCountAfter:            result.BandRowCountAfter,
		TransactionEvidenceSHA256:    result.TransactionEvidenceSHA256,
		BeforeBandCounts:             result.BeforeBandCounts,
		AfterBandCounts:              result.AfterBandCounts,
		TransactionEvidence:          result.TransactionEvidence,
		ProductionEligible:           false,
	}
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(receipt); err != nil {
		return fmt.Errorf("write aggregate execution receipt: %w", err)
	}
	return nil
}

func evidenceTimestamp(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05.000Z")
}

func parseOptions(args []string) (options, error) {
	if len(args) == 0 {
		return options{}, fmt.Errorf("command is required: run")
	}
	opt := options{command: args[0]}
	if opt.command != "run" {
		return options{}, fmt.Errorf("unsupported command %q: use run", opt.command)
	}
	flags := flag.NewFlagSet("csv-import-failure-rehearsal run", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&opt.clinicCode, "clinic-code", "", "clinic code binding")
	flags.Int64Var(&opt.clinicOrdinal, "clinic-ordinal", 0, "fixed clinic ordinal 1")
	flags.StringVar(&opt.runID, "run-id", "", "failure rehearsal run ID")
	flags.StringVar(&opt.targetReleaseCommit, "target-release-commit", "", "exact clean target release commit")
	flags.StringVar(&opt.targetDatabaseIdentitySHA256, "target-database-identity-sha256", "", "host-attested disposable database identity SHA-256")
	flags.StringVar(&opt.fixtureManifestSHA256, "fixture-manifest-sha256", "", "host-generated fixture manifest SHA-256")
	flags.StringVar(&opt.failureRuntimeReportSHA256, "failure-runtime-report-sha256", "", "host-generated runtime report SHA-256")
	flags.Int64Var(&opt.clinicID, "clinic-id", 0, "explicit target clinic seed ID")
	flags.Int64Var(&opt.animalSpeciesID, "fallback-animal-species-id", 0, "explicit active target species seed ID")
	flags.Int64Var(&opt.examTypeID, "fallback-exam-type-id", 0, "explicit target exam type seed ID")
	flags.Int64Var(&opt.trimmingReservationTypeID, "trimming-reservation-type-id", 0, "explicit target trimming reservation type seed ID")
	flags.Int64Var(&opt.cashPaymentMethodID, "cash-payment-method-id", 0, "explicit target cash payment method seed ID")
	flags.Int64Var(&opt.creditCardPaymentMethodID, "credit-card-payment-method-id", 0, "explicit target credit-card payment method seed ID")
	flags.StringVar(&opt.confirmTargetDatabase, "confirm-target-database", "", "exact disposable DB_NAME confirmation")
	flags.StringVar(&opt.confirmDisposableRehearsal, "confirm-disposable-rehearsal", "", "fixed disposable-only confirmation")
	if err := flags.Parse(args[1:]); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected positional arguments")
	}
	if err := csvimport.ValidateSyntheticFailureInput(syntheticFailureInput(opt)); err != nil {
		return options{}, err
	}
	return opt, nil
}

func syntheticFailureInput(opt options) csvimport.SyntheticFailureInput {
	return csvimport.SyntheticFailureInput{
		ClinicCode:                   opt.clinicCode,
		ClinicOrdinal:                opt.clinicOrdinal,
		RunID:                        opt.runID,
		TargetReleaseCommit:          opt.targetReleaseCommit,
		TargetDatabaseIdentitySHA256: opt.targetDatabaseIdentitySHA256,
		FixtureManifestSHA256:        opt.fixtureManifestSHA256,
		FailureRuntimeReportSHA256:   opt.failureRuntimeReportSHA256,
		Seeds: csvimport.CutoverSeedIDs{
			ClinicID:                  opt.clinicID,
			AnimalSpeciesID:           opt.animalSpeciesID,
			ExamTypeID:                opt.examTypeID,
			TrimmingReservationTypeID: opt.trimmingReservationTypeID,
			CashPaymentMethodID:       opt.cashPaymentMethodID,
			CreditCardPaymentMethodID: opt.creditCardPaymentMethodID,
		},
	}
}

func validateDisposableTarget(opt options, conn dbconn.ConnParams, database string) error {
	if conn.Host != "db" {
		return fmt.Errorf("synthetic failure rehearsal requires DB_HOST=db")
	}
	if conn.SSLMode != "disable" {
		return fmt.Errorf("synthetic failure rehearsal requires DB_SSL_MODE=disable")
	}
	if !disposableDatabasePattern.MatchString(database) {
		return fmt.Errorf("DB_NAME must identify a disposable F8 G4 database")
	}
	if opt.confirmTargetDatabase != database {
		return fmt.Errorf("target database confirmation must exactly match DB_NAME")
	}
	if opt.confirmDisposableRehearsal != disposableConfirmation {
		return fmt.Errorf("disposable confirmation must be F8_G4_DISPOSABLE_ONLY")
	}
	return nil
}

func buildDisposablePoolConfig(
	conn dbconn.ConnParams,
	database string,
) (*pgxpool.Config, error) {
	port, err := strconv.ParseUint(conn.Port, 10, 16)
	if err != nil || port == 0 {
		return nil, fmt.Errorf("DB_PORT must be an integer between 1 and 65535")
	}
	poolConfig, err := pgxpool.ParseConfig("sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("initialize disposable database configuration: %w", err)
	}
	poolConfig.ConnConfig.Host = conn.Host
	poolConfig.ConnConfig.Port = uint16(port)
	poolConfig.ConnConfig.User = conn.User
	poolConfig.ConnConfig.Password = conn.Password
	poolConfig.ConnConfig.Database = database
	poolConfig.ConnConfig.Fallbacks = nil
	poolConfig.ConnConfig.RuntimeParams = map[string]string{"TimeZone": config.JapanTimeZone}
	if poolConfig.ConnConfig.Host != "db" ||
		poolConfig.ConnConfig.Database != database ||
		len(poolConfig.ConnConfig.Fallbacks) != 0 {
		return nil, fmt.Errorf("effective disposable target identity failed validation")
	}
	return poolConfig, nil
}
