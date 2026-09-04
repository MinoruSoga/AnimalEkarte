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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/animal-ekarte/backend/internal/csvimport"
	"github.com/animal-ekarte/backend/internal/dbconn"
)

var phiLikeSubstrings = []string{
	"佐藤", "鈴木", "高橋", "太郎", "花子", "ポチ", "タマ", "山田",
	"owner_name", "pet_name", "staff_name",
	"090-", "@gmail.com",
}

func TestParseOptions(t *testing.T) {
	sourceDir := t.TempDir()
	args := []string{
		"apply",
		"--source-dir", sourceDir,
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
	if opt.command != "apply" || opt.sourceDir != sourceDir || opt.clinicOrdinal != 1 ||
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
		{"local rehearsal flag is not the STG switch", []string{"preflight", "--allow-local-rehearsal"}, "flag provided but not defined"},
		{"import rejects local rehearsal flag", []string{"import", "--allow-local-rehearsal"}, "flag provided but not defined"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseOptions(test.args)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestRunWithDependenciesEnvGateFailsClosedBeforePreflightAndOpenTarget(t *testing.T) {
	t.Setenv("DB_NAME", "animalekarte")

	tests := []struct {
		name     string
		appEnv   string
		sentinel string
		extraEnv map[string]string
		want     string
	}{
		{name: "production", appEnv: "production", sentinel: allowRehearsalSentinel, want: "production"},
		{name: "production mixed case", appEnv: " Production ", sentinel: allowRehearsalSentinel, want: "production"},
		{name: "development", appEnv: "development", sentinel: allowRehearsalSentinel, want: "staging"},
		{name: "local", appEnv: "local", sentinel: allowRehearsalSentinel, want: "staging"},
		{name: "empty APP_ENV", appEnv: "", sentinel: allowRehearsalSentinel, want: "staging"},
		{name: "REHEARSAL_ONLY refused without sentinel", appEnv: "staging", sentinel: "", want: "REHEARSAL_ONLY"},
		{
			name:     "CSV_IMPORT_ALLOW_LOCAL_REHEARSAL is not the STG switch",
			appEnv:   "staging",
			sentinel: "",
			extraEnv: map[string]string{"CSV_IMPORT_ALLOW_LOCAL_REHEARSAL": "1"},
			want:     "REHEARSAL_ONLY",
		},
		{name: "wrong sentinel", appEnv: "staging", sentinel: "1", want: allowRehearsalSentinel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ENV", tt.appEnv)
			t.Setenv(allowRehearsalEnv, tt.sentinel)
			t.Setenv("CSV_IMPORT_ALLOW_LOCAL_REHEARSAL", "")
			for key, value := range tt.extraEnv {
				t.Setenv(key, value)
			}

			target := &fakeCutoverTarget{}
			deps := testRunDependencies(t, testCLIBundle(), target)
			deps.preflightBundle = func(string, csvimport.ExpectedCutoverSource) (csvimport.CutoverBundle, error) {
				t.Fatal("preflightBundle must not be called when env gate fails")
				return csvimport.CutoverBundle{}, errors.New("preflightBundle must not be called")
			}
			deps.openTarget = func(context.Context, *pgxpool.Config) (cutoverTarget, error) {
				t.Fatal("openTarget must not be called when env gate fails")
				return nil, errors.New("openTarget must not be called")
			}

			err := runWithDependencies(context.Background(), testCLIArgs("preflight", t.TempDir()), slog.Default(), deps)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
			if target.pinged || target.preflightCalls != 0 {
				t.Fatalf("target was used: pinged=%v preflight=%d", target.pinged, target.preflightCalls)
			}
		})
	}
}

func TestParseOptionsImportCommand(t *testing.T) {
	opt, err := parseOptions([]string{
		"import",
		"--source-dir", t.TempDir(),
		"--expected-manifest-sha256", strings.Repeat("a", 64),
		"--clinic-code", "jouto",
		"--clinic-ordinal", "2",
		"--run-id", "run-1",
		"--clinic-id", "2",
		"--confirm-target-write",
		"--confirm-backup-ready",
		"--confirm-target-host", "db",
		"--confirm-target-database", "animalekarte",
		"--report-path", "/migration-reports/import.json",
	})
	if err != nil {
		t.Fatal(err)
	}
	if opt.command != "import" || opt.clinicID != 2 || opt.animalSpeciesID != 0 ||
		!opt.confirmTargetWrite || !opt.confirmBackupReady {
		t.Fatalf("parsed import options = %#v", opt)
	}
}

func TestRunWithDependenciesAllCommandsRequireTargetConfirmationsBeforePreflight(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")
	for _, command := range []string{"preflight", "verify"} {
		t.Run(command, func(t *testing.T) {
			deps := testRunDependencies(t, testCLIBundle(), &fakeCutoverTarget{})
			deps.preflightBundle = func(string, csvimport.ExpectedCutoverSource) (csvimport.CutoverBundle, error) {
				t.Fatal("source preflight must not run before exact target confirmations")
				return csvimport.CutoverBundle{}, nil
			}
			args := testCLIArgs(command, t.TempDir())
			for i := 0; i < len(args); i++ {
				if args[i] == "--confirm-target-host" {
					args = append(args[:i], args[i+2:]...)
					break
				}
			}
			if err := runWithDependencies(context.Background(), args, slog.Default(), deps); err == nil || !strings.Contains(err.Error(), "host confirmation") {
				t.Fatalf("error = %v, want host confirmation rejection", err)
			}
		})
	}
}

func TestRunWithDependenciesImportSequencesPreflightApplyVerify(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	root := filepath.Join(t.TempDir(), "reports")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	reportPath := filepath.Join(root, "import.json")
	target := &fakeCutoverTarget{applyResult: csvimport.CutoverResult{
		CompletedAt: time.Date(2026, 9, 1, 1, 2, 3, 0, time.UTC),
		ClinicCode:  "hachioji",
		RunID:       "run-1",
		Counts:      map[string]int64{"owners": 2},
	}}
	deps := testRunDependencies(t, testCLIBundle(), target)
	deps.reportRoot = root
	args := importCLIArgs(t.TempDir(), reportPath)
	if err := runWithDependencies(context.Background(), args, slog.Default(), deps); err != nil {
		t.Fatal(err)
	}
	if target.preflightCalls != 1 || target.applyCalls != 1 || target.verifyCalls != 1 {
		t.Fatalf("sequence preflight=%d apply=%d verify=%d", target.preflightCalls, target.applyCalls, target.verifyCalls)
	}
	if target.lastSeeds != (csvimport.CutoverSeedIDs{
		ClinicID: 1, AnimalSpeciesID: 2, ExamTypeID: 3,
		TrimmingReservationTypeID: 4, CashPaymentMethodID: 5, CreditCardPaymentMethodID: 6,
	}) {
		t.Fatalf("import seeds = %#v", target.lastSeeds)
	}
	report := readAuditReport(t, reportPath)
	if report.Status != "PASS" || report.Lane != auditLane {
		t.Fatalf("report = %#v", report)
	}
}

func TestRunWithDependenciesImportStopsBeforeApplyWhenPreflightFails(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	root := filepath.Join(t.TempDir(), "reports")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	target := &fakeCutoverTarget{preflightError: errors.New("band occupied")}
	deps := testRunDependencies(t, testCLIBundle(), target)
	deps.reportRoot = root
	err := runWithDependencies(context.Background(), importCLIArgs(t.TempDir(), filepath.Join(root, "import.json")), slog.Default(), deps)
	if err == nil || !strings.Contains(err.Error(), "preflight") {
		t.Fatalf("error = %v, want target preflight failure", err)
	}
	if target.applyCalls != 0 || target.verifyCalls != 0 {
		t.Fatalf("apply=%d verify=%d, want no writes after failed preflight", target.applyCalls, target.verifyCalls)
	}
}

func TestRunWithDependenciesImportSkipsVerifyOnRolledBackApply(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	root := filepath.Join(t.TempDir(), "reports")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	target := &fakeCutoverTarget{applyError: errors.New("copy failed")}
	deps := testRunDependencies(t, testCLIBundle(), target)
	deps.reportRoot = root
	err := runWithDependencies(context.Background(), importCLIArgs(t.TempDir(), filepath.Join(root, "import.json")), slog.Default(), deps)
	if err == nil || !strings.Contains(err.Error(), "rolled back") {
		t.Fatalf("error = %v, want rolled-back apply", err)
	}
	if target.verifyCalls != 0 {
		t.Fatalf("verify=%d, want skipped after rollback", target.verifyCalls)
	}
}

func TestRunWithDependenciesImportVerifiesAfterUnknownCommit(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	root := filepath.Join(t.TempDir(), "reports")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	target := &fakeCutoverTarget{applyError: fmt.Errorf("wrapped: %w", csvimport.ErrCutoverCommitOutcomeUnknown)}
	deps := testRunDependencies(t, testCLIBundle(), target)
	deps.reportRoot = root
	err := runWithDependencies(context.Background(), importCLIArgs(t.TempDir(), filepath.Join(root, "import.json")), slog.Default(), deps)
	if err == nil || !strings.Contains(err.Error(), "read-only verify") {
		t.Fatalf("error = %v, want unknown-commit guidance", err)
	}
	if target.verifyCalls != 1 {
		t.Fatalf("verify=%d, want 1 after unknown commit", target.verifyCalls)
	}
}

func TestRunWithDependenciesImportUnknownCommitVerifyFailure(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	root := filepath.Join(t.TempDir(), "reports")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	target := &fakeCutoverTarget{
		applyError:  fmt.Errorf("wrapped: %w", csvimport.ErrCutoverCommitOutcomeUnknown),
		verifyError: errors.New("counts mismatch"),
	}
	deps := testRunDependencies(t, testCLIBundle(), target)
	deps.reportRoot = root
	err := runWithDependencies(context.Background(), importCLIArgs(t.TempDir(), filepath.Join(root, "import.json")), slog.Default(), deps)
	if err == nil || !strings.Contains(err.Error(), "verify also failed") {
		t.Fatalf("error = %v, want combined unknown-commit and verify failure", err)
	}
	if target.verifyCalls != 1 {
		t.Fatalf("verify=%d", target.verifyCalls)
	}
	if !errors.Is(err, csvimport.ErrCutoverCommitOutcomeUnknown) {
		t.Fatalf("error = %v, want ErrCutoverCommitOutcomeUnknown in chain", err)
	}
}

func TestRunWithDependenciesImportRejectsOmittedExplicitSeedIDsBeforeOpeningTarget(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	target := &fakeCutoverTarget{}
	deps := testRunDependencies(t, testCLIBundle(), target)
	deps.openTarget = func(context.Context, *pgxpool.Config) (cutoverTarget, error) {
		t.Fatal("openTarget must not be called when explicit seed IDs are omitted")
		return nil, errors.New("openTarget must not be called")
	}
	args := stripSeedIDFlags(importCLIArgs(t.TempDir(), "/migration-reports/import.json"))
	err := runWithDependencies(context.Background(), args, slog.Default(), deps)
	if err == nil || !strings.Contains(err.Error(), "six explicit seed IDs") {
		t.Fatalf("error = %v, want omitted explicit seed ID rejection", err)
	}
	if target.pinged || target.applyCalls != 0 {
		t.Fatalf("target was used: pinged=%v apply=%d", target.pinged, target.applyCalls)
	}
}

func TestRunWithDependenciesImportFailsWhenVerifyFailsAfterApply(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	root := filepath.Join(t.TempDir(), "reports")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	target := &fakeCutoverTarget{
		applyResult: csvimport.CutoverResult{ClinicCode: "hachioji", RunID: "run-1"},
		verifyError: errors.New("counts mismatch"),
	}
	deps := testRunDependencies(t, testCLIBundle(), target)
	deps.reportRoot = root
	reportPath := filepath.Join(root, "import.json")
	err := runWithDependencies(context.Background(), importCLIArgs(t.TempDir(), reportPath), slog.Default(), deps)
	if err == nil || !strings.Contains(err.Error(), "verification failed") {
		t.Fatalf("error = %v, want verify failure after apply", err)
	}
	if target.applyCalls != 1 || target.verifyCalls != 1 {
		t.Fatalf("apply=%d verify=%d", target.applyCalls, target.verifyCalls)
	}
	report := readAuditReport(t, reportPath)
	if report.Status != "FAILED_POST_COMMIT_VERIFY" || report.FailureStage != "verify" {
		t.Fatalf("report = %#v, want post-commit verify failure", report)
	}
}

func TestRunWithDependenciesImportRejectsClinicIdentityMismatch(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	target := &fakeCutoverTarget{}
	deps := testRunDependencies(t, testCLIBundle(), target)
	deps.openTarget = func(context.Context, *pgxpool.Config) (cutoverTarget, error) {
		t.Fatal("openTarget must not be called when clinic identity fails")
		return nil, errors.New("openTarget must not be called")
	}
	args := importCLIArgs(t.TempDir(), "/migration-reports/import.json")
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--clinic-id" {
			args[i+1] = "2"
		}
	}
	err := runWithDependencies(context.Background(), args, slog.Default(), deps)
	if err == nil || !strings.Contains(err.Error(), "clinic-id must match") {
		t.Fatalf("error = %v, want clinic identity rejection", err)
	}
}

func TestRunWithDependenciesImportRejectsMixedSeedFlags(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	target := &fakeCutoverTarget{}
	deps := testRunDependencies(t, testCLIBundle(), target)
	args := testCLIArgs("import", t.TempDir())
	args = append(args,
		"--confirm-target-write",
		"--confirm-backup-ready",
		"--report-path", "/migration-reports/import.json",
		"--fallback-animal-species-id", "2",
	)
	// testCLIArgs already sets all five IDs; zero three of them after the extra flag to create a mix.
	args = zeroNamedIntFlag(args, "--fallback-exam-type-id")
	args = zeroNamedIntFlag(args, "--trimming-reservation-type-id")
	args = zeroNamedIntFlag(args, "--cash-payment-method-id")
	args = zeroNamedIntFlag(args, "--credit-card-payment-method-id")
	err := runWithDependencies(context.Background(), args, slog.Default(), deps)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "seed") {
		t.Fatalf("error = %v, want mixed seed flag rejection", err)
	}
	if target.applyCalls != 0 {
		t.Fatalf("apply=%d after mixed seed flags", target.applyCalls)
	}
}

func TestValidateImportClinicIdentity(t *testing.T) {
	tests := []struct {
		name    string
		opt     options
		wantErr string
	}{
		{name: "hachioji", opt: options{command: "import", clinicCode: "hachioji", clinicOrdinal: 1, clinicID: 1}},
		{name: "jouto", opt: options{command: "import", clinicCode: "jouto", clinicOrdinal: 2, clinicID: 2}},
		{name: "shikishima", opt: options{command: "import", clinicCode: "shikishima", clinicOrdinal: 3, clinicID: 3}},
		{name: "hakobuneco", opt: options{command: "import", clinicCode: "hakobuneco", clinicOrdinal: 4, clinicID: 4}},
		{name: "preflight skips identity", opt: options{command: "preflight", clinicCode: "other", clinicOrdinal: 9, clinicID: 9}},
		{
			name:    "clinic-id mismatch",
			opt:     options{command: "import", clinicCode: "hachioji", clinicOrdinal: 1, clinicID: 2},
			wantErr: "clinic-id must match",
		},
		{
			name:    "jouto ordinal mismatch",
			opt:     options{command: "import", clinicCode: "jouto", clinicOrdinal: 1, clinicID: 2},
			wantErr: "clinic-code/ordinal",
		},
		{
			name:    "unknown clinic",
			opt:     options{command: "import", clinicCode: "other", clinicOrdinal: 9, clinicID: 9},
			wantErr: "clinic-code/ordinal",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImportClinicIdentity(tt.opt)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("valid identity rejected: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want %s", err, tt.wantErr)
			}
		})
	}
}

func TestRunWithDependenciesImportRequiresApplyConfirmations(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	target := &fakeCutoverTarget{}
	deps := testRunDependencies(t, testCLIBundle(), target)
	deps.openTarget = func(context.Context, *pgxpool.Config) (cutoverTarget, error) {
		t.Fatal("openTarget must not be called when import confirmations fail")
		return nil, errors.New("openTarget must not be called")
	}
	err := runWithDependencies(context.Background(), testCLIArgs("import", t.TempDir()), slog.Default(), deps)
	if err == nil || !strings.Contains(err.Error(), "confirm-target-write") {
		t.Fatalf("error = %v, want write confirmation", err)
	}
}

func TestRunWithDependenciesApplyMissingConfirmationsDoesNotOpenTarget(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	tests := []struct {
		name  string
		extra []string
		want  string
	}{
		{name: "missing write acknowledgement", extra: nil, want: "confirm-target-write"},
		{
			name:  "missing backup acknowledgement",
			extra: []string{"--confirm-target-write"},
			want:  "confirm-backup-ready",
		},
		{
			name: "host mismatch",
			extra: []string{
				"--confirm-target-write",
				"--confirm-backup-ready",
				"--confirm-target-host", "other",
				"--confirm-target-database", "animalekarte",
				"--report-path", "/migration-reports/apply.json",
			},
			want: "host",
		},
		{
			name: "database mismatch",
			extra: []string{
				"--confirm-target-write",
				"--confirm-backup-ready",
				"--confirm-target-host", "db",
				"--confirm-target-database", "other",
				"--report-path", "/migration-reports/apply.json",
			},
			want: "database",
		},
		{
			name: "relative report path",
			extra: []string{
				"--confirm-target-write",
				"--confirm-backup-ready",
				"--confirm-target-host", "db",
				"--confirm-target-database", "animalekarte",
				"--report-path", "relative.json",
			},
			want: "absolute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := &fakeCutoverTarget{}
			deps := testRunDependencies(t, testCLIBundle(), target)
			deps.openTarget = func(context.Context, *pgxpool.Config) (cutoverTarget, error) {
				t.Fatal("openTarget must not be called when apply confirmations fail")
				return nil, errors.New("openTarget must not be called")
			}
			args := append(testCLIArgs("apply", t.TempDir()), tt.extra...)
			err := runWithDependencies(context.Background(), args, slog.Default(), deps)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
			if target.pinged || target.applyCalls != 0 {
				t.Fatalf("target was used: pinged=%v apply=%d", target.pinged, target.applyCalls)
			}
		})
	}
}

func TestRunWithDependenciesStagingSentinelRunsSourcePreflight(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	sourceDir := t.TempDir()
	if strings.Contains(sourceDir, "_old_db_handoff") {
		t.Fatal("test source must not use _old_db_handoff")
	}

	var gotDir string
	var gotExpected csvimport.ExpectedCutoverSource
	target := &fakeCutoverTarget{}
	deps := testRunDependencies(t, testCLIBundle(), target)
	deps.preflightBundle = func(dir string, expected csvimport.ExpectedCutoverSource) (csvimport.CutoverBundle, error) {
		gotDir = dir
		gotExpected = expected
		return testCLIBundle(), nil
	}

	if err := runWithDependencies(context.Background(), testCLIArgs("preflight", sourceDir), slog.Default(), deps); err != nil {
		t.Fatal(err)
	}
	if gotExpected.Provenance.Mode != csvimport.CutoverProvenanceStagingRehearsal ||
		gotExpected.Provenance.Target.Environment != "staging" || gotExpected.Provenance.Target.ClinicID != 1 {
		t.Fatalf("source preflight staging provenance = %+v", gotExpected.Provenance)
	}
	if gotDir != sourceDir {
		t.Fatalf("source dir = %q, want temp dir %q", gotDir, sourceDir)
	}
	if strings.Contains(gotDir, "_old_db_handoff") {
		t.Fatalf("source dir used _old_db_handoff: %s", gotDir)
	}
	if !target.pinged || !target.closed || target.preflightCalls != 1 {
		t.Fatalf("target lifecycle: pinged=%v closed=%v preflight=%d", target.pinged, target.closed, target.preflightCalls)
	}
}

func TestRunWithDependenciesApplyReportHasLane(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	root := filepath.Join(t.TempDir(), "reports")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	reportPath := filepath.Join(root, "apply.json")
	target := &fakeCutoverTarget{applyResult: csvimport.CutoverResult{
		CompletedAt: time.Date(2026, 7, 22, 1, 2, 3, 0, time.UTC),
		ClinicCode:  "hachioji",
		RunID:       "run-1",
		Counts:      map[string]int64{"owners": 2, "pets": 3, "staffs": 1},
	}}
	deps := testRunDependencies(t, testCLIBundle(), target)
	deps.reportRoot = root
	args := append(testCLIArgs("apply", t.TempDir()),
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
	if target.applyCalls != 1 {
		t.Fatalf("apply calls=%d", target.applyCalls)
	}
	if !strings.Contains(logs.String(), "committing each table") {
		t.Fatalf("apply logs must record per-table commits: %q", logs.String())
	}

	raw, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	assertNoPHI(t, raw)
	if !strings.Contains(string(raw), `"lane": "stg-uat-rehearsal"`) {
		t.Fatalf("report JSON missing lane stg-uat-rehearsal: %s", raw)
	}
	report := readAuditReport(t, reportPath)
	if report.Lane != auditLane || report.Status != "PASS" || report.Counts["owners"] != 2 {
		t.Fatalf("report = %#v", report)
	}
	if report.SeedIDs != target.lastSeeds ||
		report.SeedIDs.CashPaymentMethodID != 5 ||
		report.SeedIDs.CreditCardPaymentMethodID != 6 {
		t.Fatalf("report seed IDs = %#v, target seeds = %#v", report.SeedIDs, target.lastSeeds)
	}
}

func TestRunWithDependenciesRemoteHostUsesFakeOpenTarget(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	opened := false
	var openedHost string
	target := &fakeCutoverTarget{}
	deps := testRunDependencies(t, testCLIBundle(), target)
	deps.fromEnv = func() (dbconn.ConnParams, error) {
		return dbconn.ConnParams{
			Host: "example.invalid", Port: "5432", User: "user", Password: "secret", SSLMode: "require",
		}, nil
	}
	deps.openTarget = func(_ context.Context, cfg *pgxpool.Config) (cutoverTarget, error) {
		opened = true
		if cfg != nil {
			openedHost = cfg.ConnConfig.Host
		}
		return target, nil
	}

	args := testCLIArgs("preflight", t.TempDir())
	for i := range args {
		if args[i] == "db" && i > 0 && args[i-1] == "--confirm-target-host" {
			args[i] = "example.invalid"
		}
	}
	if err := runWithDependencies(context.Background(), args, slog.Default(), deps); err != nil {
		t.Fatal(err)
	}
	if !opened {
		t.Fatal("expected fake openTarget after remote host string gate passed")
	}
	if openedHost != "example.invalid" {
		t.Fatalf("opened host = %q, want example.invalid (string gate only, no real session)", openedHost)
	}
	if !target.pinged {
		t.Fatal("expected fake target ping, not a real remote session")
	}
}

func TestRequireRemoteHostAllowedIsStringGate(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		sentinel string
		wantErr  bool
		want     string
	}{
		{name: "local db", host: "db", sentinel: "", wantErr: false},
		{name: "localhost", host: "localhost", sentinel: "", wantErr: false},
		{name: "example.invalid without sentinel", host: "example.invalid", sentinel: "", wantErr: true, want: "non-local"},
		{name: "example.invalid names sentinel", host: "example.invalid", sentinel: "", wantErr: true, want: allowRehearsalEnv + "=" + allowRehearsalSentinel},
		{name: "example.invalid with sentinel", host: "example.invalid", sentinel: allowRehearsalSentinel, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(allowRehearsalEnv, tt.sentinel)
			err := requireRemoteHostAllowed(tt.host)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.want) {
					t.Fatalf("error = %v, want %q", err, tt.want)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAuditReportJSONHasLaneAndForbidsPHI(t *testing.T) {
	completed := time.Date(2026, 8, 25, 1, 2, 3, 0, time.UTC)
	report := auditReport{
		Status:         "PASS",
		Lane:           auditLane,
		StartedAt:      time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC),
		CompletedAt:    &completed,
		ManifestSHA256: strings.Repeat("a", 64),
		ClinicCode:     "hachioji",
		ClinicOrdinal:  1,
		RunID:          "run-1",
		TargetHost:     "db",
		TargetDatabase: "animalekarte",
		SeedIDs: csvimport.CutoverSeedIDs{
			ClinicID:                  1,
			AnimalSpeciesID:           2,
			ExamTypeID:                3,
			TrimmingReservationTypeID: 4,
			CashPaymentMethodID:       5,
			CreditCardPaymentMethodID: 6,
		},
		IDBand: csvimport.CutoverIDBand{Base: 0, EndExclusive: 10_000_000},
		Counts: map[string]int64{"owners": 2, "pets": 3, "staffs": 1},
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	assertNoPHI(t, raw)
	if !strings.Contains(string(raw), `"lane":"stg-uat-rehearsal"`) {
		t.Fatalf("marshaled report missing stable lane key: %s", raw)
	}
	if !strings.Contains(string(raw), `"clinicId":1`) || !strings.Contains(string(raw), `"owners":2`) {
		t.Fatalf("marshaled report missing seed IDs or aggregate counts: %s", raw)
	}
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
	if err := validateApplyConfirmations(valid); err != nil {
		t.Fatalf("valid confirmations rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*options)
		want   string
	}{
		{"write acknowledgement", func(o *options) { o.confirmTargetWrite = false }, "write"},
		{"backup acknowledgement", func(o *options) { o.confirmBackupReady = false }, "backup"},
		{"audit report", func(o *options) { o.reportPath = "" }, "report"},
		{"relative report", func(o *options) { o.reportPath = "report.json" }, "absolute"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := valid
			tt.mutate(&candidate)
			err := validateApplyConfirmations(candidate)
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
	file, err := createAuditReport(path, root, auditReport{Status: "STARTED", Lane: auditLane})
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
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	assertNoPHI(t, raw)
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

func TestBuildTargetPoolConfig(t *testing.T) {
	t.Setenv("PGHOST", "db,evil.invalid")
	t.Setenv("PGPORT", "5432,5432")
	t.Setenv("PGOPTIONS", "-c search_path=attacker")

	local := dbconn.ConnParams{
		Host:     "db",
		Port:     "5432",
		User:     "operator host=evil.invalid",
		Password: "secret host=evil.invalid",
		SSLMode:  "disable",
	}
	config, err := buildTargetPoolConfig(local, "animalekarte host=evil.invalid")
	if err != nil {
		t.Fatal(err)
	}
	if config.ConnConfig.Host != "db" || config.ConnConfig.Password != local.Password || config.ConnConfig.Database != "animalekarte host=evil.invalid" {
		t.Fatalf("structured config changed identities: host=%q database=%q", config.ConnConfig.Host, config.ConnConfig.Database)
	}
	if len(config.ConnConfig.Fallbacks) != 0 {
		t.Fatalf("environment-derived fallbacks retained: %d", len(config.ConnConfig.Fallbacks))
	}
	if config.ConnConfig.RuntimeParams["TimeZone"] != dbconn.JapanTimeZone {
		t.Fatalf("timezone = %#v", config.ConnConfig.RuntimeParams)
	}

	remote, err := buildTargetPoolConfig(dbconn.ConnParams{
		Host: "example.invalid", Port: "5432", User: "user", Password: "secret host=evil.invalid", SSLMode: "require",
	}, "animalekarte")
	if err != nil {
		t.Fatal(err)
	}
	if remote.ConnConfig.Host != "example.invalid" {
		t.Fatalf("remote host = %q, want example.invalid (string assignment, no session)", remote.ConnConfig.Host)
	}
	if remote.ConnConfig.Password != "secret host=evil.invalid" {
		t.Fatal("password was not assigned structurally")
	}

	tests := []struct {
		name string
		conn dbconn.ConnParams
		db   string
		want string
	}{
		{name: "local ssl required", conn: dbconn.ConnParams{Host: "db", Port: "5432", SSLMode: "require"}, db: "animalekarte", want: "disable"},
		{name: "invalid port", conn: dbconn.ConnParams{Host: "db", Port: "bad", SSLMode: "disable"}, db: "animalekarte", want: "DB_PORT"},
		{name: "zero port", conn: dbconn.ConnParams{Host: "db", Port: "0", SSLMode: "disable"}, db: "animalekarte", want: "DB_PORT"},
		{name: "remote ssl not whitelisted", conn: dbconn.ConnParams{Host: "example.invalid", Port: "5432", SSLMode: "prefer"}, db: "animalekarte", want: "DB_SSL_MODE"},
		{name: "remote ssl disable", conn: dbconn.ConnParams{Host: "example.invalid", Port: "5432", SSLMode: "disable"}, db: "animalekarte", want: "DB_SSL_MODE"},
		{name: "remote ssl injection", conn: dbconn.ConnParams{Host: "example.invalid", Port: "5432", SSLMode: "require host=evil.invalid"}, db: "animalekarte", want: "DB_SSL_MODE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildTargetPoolConfig(tt.conn, tt.db)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}

	if config.ConnConfig.TLSConfig != nil {
		t.Fatal("local disable must not enable TLS")
	}
	if config.AfterConnect == nil {
		t.Fatal("AfterConnect must enable app.bypass_rls so PlanetScale user-defined roles can see clinic seeds")
	}

	for _, mode := range []string{"require", "verify-ca", "verify-full"} {
		t.Run("remote ssl "+mode, func(t *testing.T) {
			cfg, err := buildTargetPoolConfig(dbconn.ConnParams{
				Host: "example.invalid", Port: "5432", User: "user", Password: "secret", SSLMode: mode, SSLRootCert: "system",
			}, "animalekarte")
			if err != nil {
				t.Fatal(err)
			}
			if cfg.ConnConfig.Host != "example.invalid" {
				t.Fatalf("host = %q", cfg.ConnConfig.Host)
			}
			if cfg.ConnConfig.TLSConfig == nil {
				t.Fatal("remote TLSConfig is nil; PlanetScale would reject plaintext with SSL/TLS required")
			}
			if cfg.ConnConfig.TLSConfig.ServerName != "example.invalid" {
				t.Fatalf("TLSConfig.ServerName = %q", cfg.ConnConfig.TLSConfig.ServerName)
			}
			if cfg.ConnConfig.DialFunc == nil {
				t.Fatal("remote DialFunc must set TCP keepalives for long STG UAT imports")
			}
		})
	}
}

func TestApplyIsolationLevelIsRepeatableRead(t *testing.T) {
	if applyIsolationLevel() != pgx.RepeatableRead {
		t.Fatalf("apply isolation = %q, want RepeatableRead", applyIsolationLevel())
	}
	if applyIsolationLevel() == pgx.Serializable {
		t.Fatal("STG UAT apply must not use Serializable")
	}
}

func TestCutoverApplyFailureClassification(t *testing.T) {
	status, stage := cutoverApplyFailureClassification(errors.New("pre-commit failure"))
	if status != "FAILED_TABLE_ROLLED_BACK" || stage != "table" {
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

func TestRunWithDependenciesRecordsApplyFailures(t *testing.T) {
	withStagingRehearsalEnv(t)
	t.Setenv("DB_NAME", "animalekarte")

	for _, test := range []struct {
		name       string
		applyError error
		wantStatus string
		wantError  string
	}{
		{"rolled back", errors.New("copy failed"), "FAILED_TABLE_ROLLED_BACK", "uncommitted table rolled back"},
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
			args := append(testCLIArgs("apply", t.TempDir()),
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
			raw, readErr := os.ReadFile(reportPath)
			if readErr != nil {
				t.Fatal(readErr)
			}
			assertNoPHI(t, raw)
			got := readAuditReport(t, reportPath)
			if got.Status != test.wantStatus || got.Lane != auditLane {
				t.Fatalf("status = %q lane = %q, want status %q lane %q", got.Status, got.Lane, test.wantStatus, auditLane)
			}
		})
	}
}

func withStagingRehearsalEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "staging")
	t.Setenv(allowRehearsalEnv, allowRehearsalSentinel)
	t.Setenv("CSV_IMPORT_ALLOW_LOCAL_REHEARSAL", "")
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

func testCLIArgs(command, sourceDir string) []string {
	return []string{
		command,
		"--source-dir", sourceDir,
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
		"--confirm-target-host", "db",
		"--confirm-target-database", "animalekarte",
	}
}

func importCLIArgs(sourceDir, reportPath string) []string {
	return append(testCLIArgs("import", sourceDir),
		"--confirm-target-write",
		"--confirm-backup-ready",
		"--report-path", reportPath,
	)
}

func stripSeedIDFlags(args []string) []string {
	names := map[string]struct{}{
		"--fallback-animal-species-id":    {},
		"--fallback-exam-type-id":         {},
		"--trimming-reservation-type-id":  {},
		"--cash-payment-method-id":        {},
		"--credit-card-payment-method-id": {},
	}
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if _, ok := names[args[i]]; ok {
			i++
			continue
		}
		out = append(out, args[i])
	}
	return out
}

func zeroNamedIntFlag(args []string, name string) []string {
	out := append([]string{}, args...)
	for i := 0; i < len(out)-1; i++ {
		if out[i] == name {
			out[i+1] = "0"
		}
	}
	return out
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

func assertNoPHI(t *testing.T, raw []byte) {
	t.Helper()
	body := string(raw)
	for _, fragment := range phiLikeSubstrings {
		if strings.Contains(body, fragment) {
			t.Fatalf("audit JSON contains PHI-like substring %q: %s", fragment, body)
		}
	}
}
