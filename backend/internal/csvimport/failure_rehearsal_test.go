package csvimport

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestRunSyntheticFailureRehearsalProvesRollback(t *testing.T) {
	target := &fakeFailureRehearsalTarget{}
	input := validSyntheticFailureInput()

	result, err := runSyntheticFailureRehearsal(
		context.Background(),
		target,
		func(context.Context) (failureRehearsalTx, error) {
			target.tx = &fakeFailureRehearsalTx{}
			return target.tx, nil
		},
		input,
	)
	if err != nil {
		t.Fatal(err)
	}
	if target.tx == nil ||
		target.tx.copyCalls != 1 ||
		!target.tx.injectedFailure ||
		!target.tx.rolledBack ||
		target.tx.committed {
		t.Fatalf("transaction evidence = %#v", target.tx)
	}
	if result.BandRowCountBefore != 0 || result.BandRowCountAfter != 0 {
		t.Fatalf("band counts before=%d after=%d", result.BandRowCountBefore, result.BandRowCountAfter)
	}
	if result.BeforeBandCountsSHA256 == "" ||
		result.BeforeBandCountsSHA256 != result.AfterBandCountsSHA256 ||
		result.TransactionEvidenceSHA256 == "" {
		t.Fatalf("digest evidence = %#v", result)
	}
	if result.FailureInjectedAt.Before(result.StartedAt) ||
		result.CompletedAt.Before(result.FailureInjectedAt) {
		t.Fatalf("timestamps are not chronological: %#v", result)
	}
}

func TestRunSyntheticFailureRehearsalFailsClosed(t *testing.T) {
	tests := []struct {
		name   string
		target *fakeFailureRehearsalTarget
		tx     *fakeFailureRehearsalTx
		begin  error
		want   string
	}{
		{
			name:   "occupied band before transaction",
			target: &fakeFailureRehearsalTarget{outsideBandCount: 1},
			tx:     &fakeFailureRehearsalTx{},
			want:   "empty before",
		},
		{
			name:   "transaction begin",
			target: &fakeFailureRehearsalTarget{},
			tx:     &fakeFailureRehearsalTx{},
			begin:  errors.New("private driver detail"),
			want:   "begin synthetic failure transaction",
		},
		{
			name:   "synthetic copy",
			target: &fakeFailureRehearsalTarget{},
			tx:     &fakeFailureRehearsalTx{copyError: errors.New("private row value")},
			want:   "copy synthetic owner",
		},
		{
			name:   "wrong failure kind",
			target: &fakeFailureRehearsalTarget{},
			tx:     &fakeFailureRehearsalTx{injectionError: errors.New("private database detail")},
			want:   "required foreign-key violation",
		},
		{
			name:   "rollback",
			target: &fakeFailureRehearsalTarget{},
			tx:     &fakeFailureRehearsalTx{rollbackError: errors.New("private rollback detail")},
			want:   "rollback synthetic failure transaction",
		},
		{
			name:   "residue after rollback",
			target: &fakeFailureRehearsalTarget{outsideCountsAfter: 1},
			tx:     &fakeFailureRehearsalTx{},
			want:   "rollback changed the target band",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			target := test.target
			_, err := runSyntheticFailureRehearsal(
				context.Background(),
				target,
				func(context.Context) (failureRehearsalTx, error) {
					target.tx = test.tx
					return test.tx, test.begin
				},
				validSyntheticFailureInput(),
			)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
			for _, leaked := range []string{"private driver detail", "private row value", "private database detail", "private rollback detail"} {
				if strings.Contains(err.Error(), leaked) {
					t.Fatalf("error leaked sensitive detail %q: %v", leaked, err)
				}
			}
		})
	}
}

func TestValidateSyntheticFailureInput(t *testing.T) {
	valid := validSyntheticFailureInput()
	if err := validateSyntheticFailureInput(valid); err != nil {
		t.Fatalf("valid input rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*SyntheticFailureInput)
		want   string
	}{
		{"clinic ordinal", func(input *SyntheticFailureInput) { input.ClinicOrdinal = 2 }, "clinic ordinal 1"},
		{"run id", func(input *SyntheticFailureInput) { input.RunID = "../escape" }, "run ID"},
		{"release", func(input *SyntheticFailureInput) { input.TargetReleaseCommit = "short" }, "release commit"},
		{"database identity", func(input *SyntheticFailureInput) { input.TargetDatabaseIdentitySHA256 = "invalid" }, "database identity"},
		{"fixture digest", func(input *SyntheticFailureInput) { input.FixtureManifestSHA256 = "invalid" }, "fixture manifest"},
		{"runtime digest", func(input *SyntheticFailureInput) { input.FailureRuntimeReportSHA256 = "invalid" }, "runtime report"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := valid
			test.mutate(&input)
			err := validateSyntheticFailureInput(input)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func validSyntheticFailureInput() SyntheticFailureInput {
	return SyntheticFailureInput{
		ClinicCode:                   "hachioji",
		ClinicOrdinal:                1,
		RunID:                        "hachioji-f8-failure-20260727",
		TargetReleaseCommit:          strings.Repeat("a", 40),
		TargetDatabaseIdentitySHA256: strings.Repeat("b", 64),
		FixtureManifestSHA256:        strings.Repeat("c", 64),
		FailureRuntimeReportSHA256:   strings.Repeat("d", 64),
		Seeds:                        validCutoverSeeds(),
	}
}

type fakeFailureRehearsalTarget struct {
	tx                 *fakeFailureRehearsalTx
	outsideBandCount   int64
	outsideCountsAfter int64
	countCalls         int
}

func (target *fakeFailureRehearsalTarget) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "clinic_seed AS MATERIALIZED") {
		return validTargetQuerier{}.QueryRow(ctx, query, args...)
	}
	if strings.Contains(query, "count(*)") {
		target.countCalls++
		if target.outsideBandCount != 0 && target.countCalls == 1 {
			return staticRow{values: []any{target.outsideBandCount}}
		}
		if target.outsideCountsAfter != 0 && target.countCalls > len(CutoverTableSpecs()) {
			return staticRow{values: []any{target.outsideCountsAfter}}
		}
		return staticRow{values: []any{int64(0)}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

type fakeFailureRehearsalTx struct {
	copyCalls       int
	injectedFailure bool
	rolledBack      bool
	committed       bool
	copyError       error
	injectionError  error
	rollbackError   error
}

func (tx *fakeFailureRehearsalTx) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "clinic_seed AS MATERIALIZED") {
		return validTargetQuerier{}.QueryRow(ctx, query, args...)
	}
	if strings.Contains(query, "clinic_id <> $3") ||
		strings.Contains(query, "FROM payment_splits split") {
		return staticRow{values: []any{int64(0)}}
	}
	if strings.Contains(query, "count(*)") {
		return staticRow{values: []any{int64(1)}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

func (tx *fakeFailureRehearsalTx) Exec(_ context.Context, query string, _ ...any) (pgconn.CommandTag, error) {
	if strings.Contains(query, "INSERT INTO pets") {
		tx.injectedFailure = true
		if tx.injectionError != nil {
			return pgconn.CommandTag{}, tx.injectionError
		}
		return pgconn.CommandTag{}, &pgconn.PgError{Code: "23503"}
	}
	return pgconn.NewCommandTag("SELECT 1"), nil
}

func (tx *fakeFailureRehearsalTx) CopyFrom(
	_ context.Context,
	reader io.Reader,
	copySQL string,
) (pgconn.CommandTag, error) {
	if tx.copyError != nil {
		return pgconn.CommandTag{}, tx.copyError
	}
	if !strings.HasPrefix(copySQL, "COPY ") {
		return pgconn.CommandTag{}, errors.New("unexpected synthetic copy contract")
	}
	contents, err := io.ReadAll(reader)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	if len(contents) == 0 {
		return pgconn.CommandTag{}, errors.New("empty synthetic copy")
	}
	tx.copyCalls++
	return pgconn.NewCommandTag("COPY 1"), nil
}

func (tx *fakeFailureRehearsalTx) Rollback(context.Context) error {
	tx.rolledBack = true
	return tx.rollbackError
}

func (tx *fakeFailureRehearsalTx) Commit(context.Context) error {
	tx.committed = true
	return errors.New("synthetic failure transaction must not commit")
}
