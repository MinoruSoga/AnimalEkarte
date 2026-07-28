package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/animal-ekarte/backend/internal/csvimport"
	"github.com/animal-ekarte/backend/internal/dbconn"
)

func TestParseOptions(t *testing.T) {
	args := []string{
		"apply",
		"--source-dir", "/migration-input",
		"--expected-manifest-sha256", strings.Repeat("a", 64),
		"--clinic-code", "hachioji",
		"--clinic-ordinal", "1",
		"--run-id", "run-1",
		"--clinic-id", "2",
		"--fallback-animal-species-id", "3",
		"--fallback-exam-type-id", "4",
		"--trimming-reservation-type-id", "5",
		"--cash-payment-method-id", "6",
		"--credit-card-payment-method-id", "7",
		"--confirm-target-write",
		"--confirm-backup-ready",
		"--confirm-target-host", "db",
		"--confirm-target-database", "animalekarte",
		"--report-path", "/migration-reports/run.json",
	}
	opt, err := parseOptions(args)
	if err != nil {
		t.Fatal(err)
	}
	if opt.command != "apply" || opt.sourceDir != "/migration-input" || opt.clinicOrdinal != 1 ||
		opt.clinicID != 2 || opt.animalSpeciesID != 3 || opt.examTypeID != 4 ||
		opt.trimmingReservationTypeID != 5 || opt.cashPaymentMethodID != 6 ||
		opt.creditCardPaymentMethodID != 7 || !opt.confirmTargetWrite || !opt.confirmBackupReady {
		t.Fatalf("parsed options = %#v", opt)
	}

	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{"missing command", nil, "command is required"},
		{"unsupported command", []string{"destroy"}, "unsupported command"},
		{"invalid flag value", []string{"verify", "--clinic-ordinal", "not-an-integer"}, "invalid value"},
		{"extra positional", []string{"preflight", "extra"}, "unexpected positional"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseOptions(test.args)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestRunWithDependenciesRoutesCommandsAndWritesAudit(t *testing.T) {
	t.Setenv("DB_NAME", "animalekarte")
	bundle := testCLIBundle()
	commonArgs := testCLIArgs("preflight")

	for _, command := range []string{"preflight", "verify"} {
		t.Run(command, func(t *testing.T) {
			target := &fakeCutoverTarget{}
			deps := testRunDependencies(t, bundle, target)
			args := append([]string(nil), commonArgs...)
			args[0] = command
			if err := runWithDependencies(context.Background(), args, slog.Default(), deps); err != nil {
				t.Fatal(err)
			}
			if !target.pinged || !target.closed {
				t.Fatalf("target lifecycle: pinged=%v closed=%v", target.pinged, target.closed)
			}
			if command == "preflight" && target.preflightCalls != 1 {
				t.Fatalf("preflight calls = %d", target.preflightCalls)
			}
			if command == "verify" && target.verifyCalls != 1 {
				t.Fatalf("verify calls = %d", target.verifyCalls)
			}
			if target.lastSeeds.CashPaymentMethodID != 5 || target.lastSeeds.CreditCardPaymentMethodID != 6 {
				t.Fatalf("payment method seeds = %#v", target.lastSeeds)
			}
		})
	}

	t.Run("apply pass", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "reports")
		if err := os.Mkdir(root, 0o700); err != nil {
			t.Fatal(err)
		}
		reportPath := filepath.Join(root, "apply.json")
		target := &fakeCutoverTarget{applyResult: csvimport.CutoverResult{
			CompletedAt: time.Date(2026, 7, 22, 1, 2, 3, 0, time.UTC),
			ClinicCode:  "hachioji",
			RunID:       "run-1",
			Counts:      map[string]int64{"owners": 2},
		}}
		deps := testRunDependencies(t, bundle, target)
		deps.reportRoot = root
		args := append(testCLIArgs("apply"),
			"--confirm-target-write",
			"--confirm-backup-ready",
			"--confirm-target-host", "db",
			"--confirm-target-database", "animalekarte",
			"--report-path", reportPath,
		)
		var logs bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logs, nil))
		if err := runWithDependencies(context.Background(), args, logger, deps); err != nil {
			t.Fatal(err)
		}
		if target.applyCalls != 1 || !strings.Contains(logs.String(), "apply PASS") {
			t.Fatalf("apply calls=%d logs=%q", target.applyCalls, logs.String())
		}
		report := readAuditReport(t, reportPath)
		if report.Status != "PASS" || report.CompletedAt == nil || report.Counts["owners"] != 2 {
			t.Fatalf("report = %#v", report)
		}
		if report.SeedIDs != target.lastSeeds ||
			report.SeedIDs.CashPaymentMethodID != 5 ||
			report.SeedIDs.CreditCardPaymentMethodID != 6 {
			t.Fatalf("report seed IDs = %#v, target seeds = %#v", report.SeedIDs, target.lastSeeds)
		}
	})
}

func TestRunWithDependenciesRecordsApplyFailures(t *testing.T) {
	t.Setenv("DB_NAME", "animalekarte")
	for _, test := range []struct {
		name       string
		applyError error
		wantStatus string
		wantError  string
	}{
		{"rolled back", errors.New("copy failed"), "FAILED_DATA_ROLLED_BACK", "transaction rolled back"},
		{"before transaction", fmt.Errorf("wrapped: %w", csvimport.ErrCutoverTransactionNotStarted), "FAILED_BEFORE_TRANSACTION", "before transaction began"},
		{"unknown commit", fmt.Errorf("wrapped: %w", csvimport.ErrCutoverCommitOutcomeUnknown), "COMMIT_OUTCOME_UNKNOWN", "run read-only verify"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "reports")
			if err := os.Mkdir(root, 0o700); err != nil {
				t.Fatal(err)
			}
			reportPath := filepath.Join(root, "failure.json")
			target := &fakeCutoverTarget{applyError: test.applyError}
			deps := testRunDependencies(t, testCLIBundle(), target)
			deps.reportRoot = root
			args := append(testCLIArgs("apply"),
				"--confirm-target-write",
				"--confirm-backup-ready",
				"--confirm-target-host", "db",
				"--confirm-target-database", "animalekarte",
				"--report-path", reportPath,
			)
			err := runWithDependencies(context.Background(), args, slog.Default(), deps)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("error = %v, want %q", err, test.wantError)
			}
			if got := readAuditReport(t, reportPath).Status; got != test.wantStatus {
				t.Fatalf("status = %q, want %q", got, test.wantStatus)
			}
		})
	}
}

func TestRunWithDependenciesFailsClosedAtBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		command string
		dbName  string
		mutate  func(*runDependencies, *fakeCutoverTarget)
		want    string
	}{
		{
			name: "timezone", command: "preflight", dbName: "animalekarte", want: "timezone configuration failed",
			mutate: func(deps *runDependencies, _ *fakeCutoverTarget) {
				deps.configureTimeZone = func() error { return errors.New("timezone unavailable") }
			},
		},
		{
			name: "source", command: "preflight", dbName: "animalekarte", want: "source preflight failed",
			mutate: func(deps *runDependencies, _ *fakeCutoverTarget) {
				deps.preflightBundle = func(string, csvimport.ExpectedCutoverSource) (csvimport.CutoverBundle, error) {
					return csvimport.CutoverBundle{}, errors.New("untrusted source")
				}
			},
		},
		{
			name: "environment", command: "preflight", dbName: "animalekarte", want: "missing database environment",
			mutate: func(deps *runDependencies, _ *fakeCutoverTarget) {
				deps.fromEnv = func() (dbconn.ConnParams, error) {
					return dbconn.ConnParams{}, errors.New("missing database environment")
				}
			},
		},
		{name: "database name", command: "preflight", want: "DB_NAME is required", mutate: func(*runDependencies, *fakeCutoverTarget) {}},
		{
			name: "open target", command: "preflight", dbName: "animalekarte", want: "open target database",
			mutate: func(deps *runDependencies, _ *fakeCutoverTarget) {
				deps.openTarget = func(context.Context, *pgxpool.Config) (cutoverTarget, error) {
					return nil, errors.New("open failed")
				}
			},
		},
		{
			name: "ping target", command: "preflight", dbName: "animalekarte", want: "ping target database",
			mutate: func(_ *runDependencies, target *fakeCutoverTarget) { target.pingError = errors.New("ping failed") },
		},
		{
			name: "target preflight", command: "preflight", dbName: "animalekarte", want: "target preflight failed",
			mutate: func(_ *runDependencies, target *fakeCutoverTarget) {
				target.preflightError = errors.New("band occupied")
			},
		},
		{
			name: "verification", command: "verify", dbName: "animalekarte", want: "cutover verification failed",
			mutate: func(_ *runDependencies, target *fakeCutoverTarget) { target.verifyError = errors.New("count mismatch") },
		},
		{name: "apply confirmation", command: "apply", dbName: "animalekarte", want: "confirm-target-write", mutate: func(*runDependencies, *fakeCutoverTarget) {}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("DB_NAME", test.dbName)
			target := &fakeCutoverTarget{}
			deps := testRunDependencies(t, testCLIBundle(), target)
			test.mutate(&deps, target)
			err := runWithDependencies(context.Background(), testCLIArgs(test.command), slog.Default(), deps)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func testCLIBundle() csvimport.CutoverBundle {
	return csvimport.CutoverBundle{Manifest: csvimport.CutoverManifest{
		ClinicCode:    "hachioji",
		ClinicOrdinal: 1,
		SourceRunID:   "run-1",
		IDBand:        csvimport.CutoverIDBand{Base: 10_000_000, EndExclusive: 20_000_000},
		Tables:        []csvimport.CutoverManifestTable{{Table: "owners"}},
	}}
}

func testCLIArgs(command string) []string {
	return []string{
		command,
		"--source-dir", "/migration-input",
		"--expected-manifest-sha256", strings.Repeat("a", 64),
		"--clinic-code", "hachioji",
		"--clinic-ordinal", "1",
		"--run-id", "run-1",
		"--clinic-id", "1",
		"--fallback-animal-species-id", "2",
		"--fallback-exam-type-id", "3",
		"--trimming-reservation-type-id", "4",
		"--cash-payment-method-id", "5",
		"--credit-card-payment-method-id", "6",
	}
}

type fakeCutoverTarget struct {
	pinged         bool
	closed         bool
	preflightCalls int
	verifyCalls    int
	applyCalls     int
	applyResult    csvimport.CutoverResult
	applyError     error
	pingError      error
	preflightError error
	verifyError    error
	lastSeeds      csvimport.CutoverSeedIDs
}

func (f *fakeCutoverTarget) Ping(context.Context) error {
	f.pinged = true
	return f.pingError
}

func (f *fakeCutoverTarget) Close() {
	f.closed = true
}

func (f *fakeCutoverTarget) Preflight(_ context.Context, _ csvimport.CutoverManifest, seeds csvimport.CutoverSeedIDs) error {
	f.preflightCalls++
	f.lastSeeds = seeds
	return f.preflightError
}

func (f *fakeCutoverTarget) Verify(_ context.Context, _ csvimport.CutoverManifest, seeds csvimport.CutoverSeedIDs) error {
	f.verifyCalls++
	f.lastSeeds = seeds
	return f.verifyError
}

func (f *fakeCutoverTarget) Apply(_ context.Context, _ csvimport.CutoverBundle, seeds csvimport.CutoverSeedIDs) (csvimport.CutoverResult, error) {
	f.applyCalls++
	f.lastSeeds = seeds
	return f.applyResult, f.applyError
}

func testRunDependencies(t *testing.T, bundle csvimport.CutoverBundle, target cutoverTarget) runDependencies {
	t.Helper()
	return runDependencies{
		configureTimeZone: func() error { return nil },
		preflightBundle: func(string, csvimport.ExpectedCutoverSource) (csvimport.CutoverBundle, error) {
			return bundle, nil
		},
		fromEnv: func() (dbconn.ConnParams, error) {
			return dbconn.ConnParams{Host: "db", Port: "5432", User: "user", Password: "secret", SSLMode: "disable"}, nil
		},
		openTarget: func(context.Context, *pgxpool.Config) (cutoverTarget, error) {
			return target, nil
		},
		reportRoot: "/migration-reports",
	}
}

func readAuditReport(t *testing.T, path string) auditReport {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report auditReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	return report
}

func TestValidateApplyConfirmations(t *testing.T) {
	valid := options{
		command:               "apply",
		confirmTargetWrite:    true,
		confirmBackupReady:    true,
		confirmTargetHost:     "db.internal",
		confirmTargetDatabase: "animalekarte",
		reportPath:            "/migration-reports/report.json",
	}
	if err := validateApplyConfirmations(valid, "db.internal", "animalekarte"); err != nil {
		t.Fatalf("valid confirmations rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*options)
		want   string
	}{
		{"write acknowledgement", func(o *options) { o.confirmTargetWrite = false }, "write"},
		{"backup acknowledgement", func(o *options) { o.confirmBackupReady = false }, "backup"},
		{"host binding", func(o *options) { o.confirmTargetHost = "other" }, "host"},
		{"database binding", func(o *options) { o.confirmTargetDatabase = "other" }, "database"},
		{"audit report", func(o *options) { o.reportPath = "" }, "report"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := valid
			tt.mutate(&candidate)
			err := validateApplyConfirmations(candidate, "db.internal", "animalekarte")
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestPathWithinRoot(t *testing.T) {
	if !pathWithinRoot("/migration-reports", "/migration-reports/report.json") {
		t.Fatal("expected direct-child report path to be accepted")
	}
	for _, candidate := range []string{
		"relative.json",
		"/migration-reports",
		"/migration-reports/run/report.json",
		"/migration-reports/../outside.json",
		"/migration-reports-suffix/report.json",
	} {
		if pathWithinRoot("/migration-reports", candidate) {
			t.Fatalf("unsafe candidate accepted: %s", candidate)
		}
	}
}

func TestCreateAuditReportIsOwnerOnlyAndNoClobber(t *testing.T) {
	root := filepath.Join(t.TempDir(), "reports")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "report.json")
	file, err := createAuditReport(path, root, auditReport{Status: "STARTED"})
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("report mode = %#o, want 0600", info.Mode().Perm())
	}
	if _, err := createAuditReport(path, root, auditReport{}); err == nil {
		t.Fatal("existing audit report was overwritten")
	}
}

func TestCreateAuditReportRejectsUnsafeLocations(t *testing.T) {
	root := filepath.Join(t.TempDir(), "reports")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := createAuditReport(filepath.Join(t.TempDir(), "outside.json"), root, auditReport{}); err == nil {
		t.Fatal("report path outside the fixed root was accepted")
	}
	if err := os.Chmod(root, 0o750); err != nil {
		t.Fatal(err)
	}
	if _, err := createAuditReport(filepath.Join(root, "report.json"), root, auditReport{}); err == nil {
		t.Fatal("group-readable report root was accepted")
	}
}

func TestCreateAuditReportRejectsInvalidRoots(t *testing.T) {
	tempDir := t.TempDir()
	missingRoot := filepath.Join(tempDir, "missing")
	if _, err := createAuditReport(filepath.Join(missingRoot, "report.json"), missingRoot, auditReport{}); err == nil {
		t.Fatal("missing report root was accepted")
	}

	fileRoot := filepath.Join(tempDir, "file-root")
	if err := os.WriteFile(fileRoot, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := createAuditReport(filepath.Join(fileRoot, "report.json"), fileRoot, auditReport{}); err == nil {
		t.Fatal("non-directory report root was accepted")
	}

	actualRoot := filepath.Join(tempDir, "actual")
	if err := os.Mkdir(actualRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	symlinkRoot := filepath.Join(tempDir, "symlink")
	if err := os.Symlink(actualRoot, symlinkRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := createAuditReport(filepath.Join(symlinkRoot, "report.json"), symlinkRoot, auditReport{}); err == nil {
		t.Fatal("symlinked report root was accepted")
	}
}

func TestReplaceAuditReportRejectsUnwritableFiles(t *testing.T) {
	closedPath := filepath.Join(t.TempDir(), "closed.json")
	closed, err := os.Create(closedPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	if err := replaceAuditReport(closed, auditReport{}); err == nil || !strings.Contains(err.Error(), "seek") {
		t.Fatalf("closed file error = %v, want seek failure", err)
	}

	readOnlyPath := filepath.Join(t.TempDir(), "read-only.json")
	if err := os.WriteFile(readOnlyPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	readOnly, err := os.Open(readOnlyPath)
	if err != nil {
		t.Fatal(err)
	}
	defer readOnly.Close() //nolint:errcheck // test cleanup
	if err := replaceAuditReport(readOnly, auditReport{}); err == nil || !strings.Contains(err.Error(), "truncate") {
		t.Fatalf("read-only file error = %v, want truncate failure", err)
	}
}

func TestBuildTargetPoolConfigUsesStructuredValues(t *testing.T) {
	t.Setenv("PGHOST", "db,evil.invalid")
	t.Setenv("PGPORT", "5432,5432")
	t.Setenv("PGOPTIONS", "-c search_path=attacker")
	conn := dbconn.ConnParams{
		Host:     "db",
		Port:     "5432",
		User:     "operator host=evil.invalid",
		Password: "secret host=evil.invalid",
		SSLMode:  "disable",
	}
	config, err := buildTargetPoolConfig(conn, "animalekarte host=evil.invalid")
	if err != nil {
		t.Fatal(err)
	}
	if config.ConnConfig.Host != "db" || config.ConnConfig.Password != conn.Password || config.ConnConfig.Database != "animalekarte host=evil.invalid" {
		t.Fatalf("structured config changed identities: host=%q database=%q", config.ConnConfig.Host, config.ConnConfig.Database)
	}
	if len(config.ConnConfig.Fallbacks) != 0 {
		t.Fatalf("environment-derived fallbacks retained: %d", len(config.ConnConfig.Fallbacks))
	}
	if len(config.ConnConfig.RuntimeParams) != 1 || config.ConnConfig.RuntimeParams["TimeZone"] == "" {
		t.Fatalf("environment-derived runtime params retained: %#v", config.ConnConfig.RuntimeParams)
	}
	if _, err := buildTargetPoolConfig(dbconn.ConnParams{Host: "evil.invalid", Port: "5432", SSLMode: "disable"}, "animalekarte"); err == nil {
		t.Fatal("remote host was accepted")
	}
	if _, err := buildTargetPoolConfig(dbconn.ConnParams{Host: "db", Port: "bad", SSLMode: "disable"}, "animalekarte"); err == nil {
		t.Fatal("invalid port was accepted")
	}
	if _, err := buildTargetPoolConfig(dbconn.ConnParams{Host: "db", Port: "5432", SSLMode: "require"}, "animalekarte"); err == nil {
		t.Fatal("non-disabled SSL mode was accepted")
	}
	if _, err := buildTargetPoolConfig(dbconn.ConnParams{Host: "db", Port: "0", SSLMode: "disable"}, "animalekarte"); err == nil {
		t.Fatal("zero port was accepted")
	}
}

func TestCutoverApplyFailureClassification(t *testing.T) {
	status, stage := cutoverApplyFailureClassification(errors.New("pre-commit failure"))
	if status != "FAILED_DATA_ROLLED_BACK" || stage != "transaction" {
		t.Fatalf("ordinary failure = (%q, %q)", status, stage)
	}

	status, stage = cutoverApplyFailureClassification(fmt.Errorf("wrapped: %w", csvimport.ErrCutoverCommitOutcomeUnknown))
	if status != "COMMIT_OUTCOME_UNKNOWN" || stage != "commit" {
		t.Fatalf("unknown commit = (%q, %q)", status, stage)
	}

	status, stage = cutoverApplyFailureClassification(fmt.Errorf("wrapped: %w", csvimport.ErrCutoverTransactionNotStarted))
	if status != "FAILED_BEFORE_TRANSACTION" || stage != "begin" {
		t.Fatalf("begin failure = (%q, %q)", status, stage)
	}
}
