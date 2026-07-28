package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/animal-ekarte/backend/internal/csvimport"
	"github.com/animal-ekarte/backend/internal/dbconn"
)

func TestRunWithDependenciesWritesAggregateReceipt(t *testing.T) {
	t.Setenv("DB_NAME", "animalekarte_f8_g4_run_1")
	var output bytes.Buffer
	target := &fakeFailureTarget{}
	deps := validRunDependencies(target)
	deps.runRehearsal = func(
		context.Context,
		failureTarget,
		csvimport.SyntheticFailureInput,
	) (csvimport.SyntheticFailureResult, error) {
		return validFailureResult(), nil
	}

	if err := runWithDependencies(context.Background(), validArgs(), &output, deps); err != nil {
		t.Fatal(err)
	}
	if !target.pinged || !target.closed {
		t.Fatalf("target lifecycle pinged=%v closed=%v", target.pinged, target.closed)
	}
	var receipt executionReceipt
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Status != "FAILED_DATA_ROLLED_BACK" ||
		receipt.ExecutionMode != "SYNTHETIC_FAILURE_REHEARSAL" ||
		receipt.TransactionEvidenceSHA256 != strings.Repeat("e", 64) ||
		receipt.StartedAt != "2026-07-27T01:02:03.000Z" ||
		receipt.BandRowCountBefore != 0 ||
		receipt.BandRowCountAfter != 0 {
		t.Fatalf("receipt = %#v", receipt)
	}
}

func TestEvidenceTimestampMatchesJavaScriptCanonicalUTCFormat(t *testing.T) {
	value := time.Date(2026, 7, 27, 1, 2, 3, 123456789, time.FixedZone("JST", 9*60*60))
	if got := evidenceTimestamp(value); got != "2026-07-26T16:02:03.123Z" {
		t.Fatalf("evidenceTimestamp() = %q", got)
	}
}

func TestRunWithDependenciesRejectsUnsafeTargetBeforeOpeningIt(t *testing.T) {
	tests := []struct {
		name   string
		dbName string
		args   []string
		conn   dbconn.ConnParams
		want   string
	}{
		{
			name: "remote host", dbName: "animalekarte_f8_g4_run_1",
			args: validArgs(),
			conn: dbconn.ConnParams{Host: "remote.example", Port: "5432", SSLMode: "disable"},
			want: "DB_HOST=db",
		},
		{
			name: "shared database", dbName: "animalekarte",
			args: validArgs(),
			conn: dbconn.ConnParams{Host: "db", Port: "5432", SSLMode: "disable"},
			want: "disposable F8 G4",
		},
		{
			name: "target confirmation", dbName: "animalekarte_f8_g4_other",
			args: validArgs(),
			conn: dbconn.ConnParams{Host: "db", Port: "5432", SSLMode: "disable"},
			want: "confirmation",
		},
		{
			name: "SSL mode", dbName: "animalekarte_f8_g4_run_1",
			args: validArgs(),
			conn: dbconn.ConnParams{Host: "db", Port: "5432", SSLMode: "require"},
			want: "DB_SSL_MODE=disable",
		},
		{
			name: "missing sentinel", dbName: "animalekarte_f8_g4_run_1",
			args: withoutArgs(validArgs(), "--confirm-disposable-rehearsal", "F8_G4_DISPOSABLE_ONLY"),
			conn: dbconn.ConnParams{Host: "db", Port: "5432", SSLMode: "disable"},
			want: "disposable confirmation",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("DB_NAME", test.dbName)
			target := &fakeFailureTarget{}
			deps := validRunDependencies(target)
			deps.fromEnv = func() (dbconn.ConnParams, error) {
				return test.conn, nil
			}
			err := runWithDependencies(context.Background(), test.args, &bytes.Buffer{}, deps)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
			if target.opened {
				t.Fatal("unsafe target was opened")
			}
		})
	}
}

func TestParseOptionsRejectsVariableInjectionSurface(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{"missing command", nil, "command is required"},
		{"wrong command", append([]string{"apply"}, validArgs()[1:]...), "unsupported command"},
		{"operator SQL", append(validArgs(), "--injection-sql", "DELETE FROM owners"), "flag provided but not defined"},
		{"operator source", append(validArgs(), "--source-dir", "/private/data"), "flag provided but not defined"},
		{"wrong ordinal", replaceArg(validArgs(), "--clinic-ordinal", "2"), "clinic ordinal 1"},
		{"invalid release", replaceArg(validArgs(), "--target-release-commit", "short"), "release commit"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseOptions(test.args)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestRunWithDependenciesPreservesErrorChain(t *testing.T) {
	t.Setenv("DB_NAME", "animalekarte_f8_g4_run_1")
	for _, test := range []struct {
		name   string
		mutate func(*runDependencies, *fakeFailureTarget)
		want   []string
	}{
		{
			name: "ping",
			mutate: func(_ *runDependencies, target *fakeFailureTarget) {
				target.pingError = errors.New("secret database detail")
			},
			want: []string{"ping disposable target database", "secret database detail"},
		},
		{
			name: "rehearsal",
			mutate: func(deps *runDependencies, _ *fakeFailureTarget) {
				deps.runRehearsal = func(
					context.Context,
					failureTarget,
					csvimport.SyntheticFailureInput,
				) (csvimport.SyntheticFailureResult, error) {
					return csvimport.SyntheticFailureResult{}, errors.New("synthetic failure rollback changed the target band")
				}
			},
			want: []string{"rollback changed"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			target := &fakeFailureTarget{}
			deps := validRunDependencies(target)
			test.mutate(&deps, target)
			err := runWithDependencies(context.Background(), validArgs(), &bytes.Buffer{}, deps)
			if err == nil {
				t.Fatal("expected error")
			}
			for _, want := range test.want {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error = %v, want substring %q", err, want)
				}
			}
		})
	}
}

func TestRunWithDependenciesFailsClosedBeforeReceipt(t *testing.T) {
	t.Setenv("DB_NAME", "animalekarte_f8_g4_run_1")
	tests := []struct {
		name   string
		mutate func(*runDependencies)
		want   []string
	}{
		{
			name: "timezone",
			mutate: func(deps *runDependencies) {
				deps.configureTimeZone = func() error { return errors.New("private timezone detail") }
			},
			want: []string{"timezone configuration failed", "private timezone detail"},
		},
		{
			name: "environment",
			mutate: func(deps *runDependencies) {
				deps.fromEnv = func() (dbconn.ConnParams, error) {
					return dbconn.ConnParams{}, errors.New("missing database environment")
				}
			},
			want: []string{"missing database environment"},
		},
		{
			name: "open",
			mutate: func(deps *runDependencies) {
				deps.openTarget = func(context.Context, *pgxpool.Config) (failureTarget, error) {
					return nil, errors.New("private open detail")
				}
			},
			want: []string{"open disposable target database", "private open detail"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			deps := validRunDependencies(&fakeFailureTarget{})
			test.mutate(&deps)
			err := runWithDependencies(context.Background(), validArgs(), &bytes.Buffer{}, deps)
			if err == nil {
				t.Fatal("expected error")
			}
			for _, want := range test.want {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error = %v, want substring %q", err, want)
				}
			}
		})
	}
}

func TestProductionDependencyWiringIsFailClosed(t *testing.T) {
	deps := productionRunDependencies()
	if deps.configureTimeZone == nil || deps.fromEnv == nil || deps.openTarget == nil || deps.runRehearsal == nil {
		t.Fatal("production dependency is missing")
	}
	config, err := pgxpool.ParseConfig("postgres://runner:secret@127.0.0.1:1/disposable?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	target, err := deps.openTarget(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if err := target.Ping(context.Background()); err == nil {
		t.Fatal("unexpectedly connected to the closed test port")
	}
	target.Close()
	if _, err := deps.runRehearsal(
		context.Background(),
		&fakeFailureTarget{},
		csvimport.SyntheticFailureInput{},
	); err == nil || !strings.Contains(err.Error(), "invalid disposable target") {
		t.Fatalf("unexpected target error = %v", err)
	}
	if err := run(context.Background(), nil, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "command is required") {
		t.Fatalf("run error = %v", err)
	}
}

func TestBuildDisposablePoolConfigRejectsInvalidPortAndClearsFallbacks(t *testing.T) {
	for _, port := range []string{"", "0", "65536", "invalid"} {
		if _, err := buildDisposablePoolConfig(
			dbconn.ConnParams{Host: "db", Port: port, SSLMode: "disable"},
			"animalekarte_f8_g4_run_1",
		); err == nil {
			t.Fatalf("invalid port %q was accepted", port)
		}
	}
	t.Setenv("PGHOST", "remote.example")
	t.Setenv("PGOPTIONS", "-c search_path=attacker")
	config, err := buildDisposablePoolConfig(
		dbconn.ConnParams{
			Host: "db", Port: "5432", User: "runner", Password: "secret", SSLMode: "disable",
		},
		"animalekarte_f8_g4_run_1",
	)
	if err != nil {
		t.Fatal(err)
	}
	if config.ConnConfig.Host != "db" ||
		len(config.ConnConfig.Fallbacks) != 0 ||
		len(config.ConnConfig.RuntimeParams) != 1 {
		t.Fatalf("effective config = %#v", config.ConnConfig)
	}
}

func validArgs() []string {
	return []string{
		"run",
		"--clinic-code", "hachioji",
		"--clinic-ordinal", "1",
		"--run-id", "hachioji-f8-failure-20260727",
		"--target-release-commit", strings.Repeat("a", 40),
		"--target-database-identity-sha256", strings.Repeat("b", 64),
		"--fixture-manifest-sha256", strings.Repeat("c", 64),
		"--failure-runtime-report-sha256", strings.Repeat("d", 64),
		"--clinic-id", "1",
		"--fallback-animal-species-id", "2",
		"--fallback-exam-type-id", "3",
		"--trimming-reservation-type-id", "4",
		"--cash-payment-method-id", "5",
		"--credit-card-payment-method-id", "6",
		"--confirm-target-database", "animalekarte_f8_g4_run_1",
		"--confirm-disposable-rehearsal", "F8_G4_DISPOSABLE_ONLY",
	}
}

func validFailureResult() csvimport.SyntheticFailureResult {
	startedAt := time.Date(2026, 7, 27, 1, 2, 3, 0, time.UTC)
	counts := csvimport.BandCountsEvidence{
		SchemaVersion: 1,
		TableCount:    21,
		Tables:        []csvimport.BandCount{},
		TotalRowCount: 0,
	}
	return csvimport.SyntheticFailureResult{
		StartedAt:                 startedAt,
		FailureInjectedAt:         startedAt.Add(time.Second),
		CompletedAt:               startedAt.Add(2 * time.Second),
		BeforeBandCounts:          counts,
		AfterBandCounts:           counts,
		TransactionEvidence:       csvimport.TransactionEvidence{},
		BeforeBandCountsSHA256:    strings.Repeat("f", 64),
		AfterBandCountsSHA256:     strings.Repeat("f", 64),
		TransactionEvidenceSHA256: strings.Repeat("e", 64),
	}
}

type fakeFailureTarget struct {
	opened    bool
	pinged    bool
	closed    bool
	pingError error
}

func (target *fakeFailureTarget) Ping(context.Context) error {
	target.pinged = true
	return target.pingError
}

func (target *fakeFailureTarget) Close() {
	target.closed = true
}

func validRunDependencies(target *fakeFailureTarget) runDependencies {
	return runDependencies{
		configureTimeZone: func() error { return nil },
		fromEnv: func() (dbconn.ConnParams, error) {
			return dbconn.ConnParams{
				Host: "db", Port: "5432", User: "runner", Password: "secret", SSLMode: "disable",
			}, nil
		},
		openTarget: func(context.Context, *pgxpool.Config) (failureTarget, error) {
			target.opened = true
			return target, nil
		},
		runRehearsal: func(
			context.Context,
			failureTarget,
			csvimport.SyntheticFailureInput,
		) (csvimport.SyntheticFailureResult, error) {
			return validFailureResult(), nil
		},
	}
}

func withoutArgs(args []string, pair ...string) []string {
	result := append([]string(nil), args...)
	for index := 0; index < len(result)-1; index++ {
		if result[index] == pair[0] && result[index+1] == pair[1] {
			return append(result[:index], result[index+2:]...)
		}
	}
	return result
}

func replaceArg(args []string, flagName, value string) []string {
	result := append([]string(nil), args...)
	for index := 0; index < len(result)-1; index++ {
		if result[index] == flagName {
			result[index+1] = value
			return result
		}
	}
	return result
}
