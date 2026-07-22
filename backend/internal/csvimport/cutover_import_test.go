package csvimport

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestApplyCutoverWithBeginImportsAndVerifiesAllTables(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	tx := &fakeCutoverTransaction{}

	result, err := applyCutoverWithBegin(
		context.Background(),
		func(context.Context) (cutoverTransaction, error) { return tx, nil },
		bundle,
		CutoverSeedIDs{ClinicID: 1, AnimalSpeciesID: 2, ExamTypeID: 3, TrimmingReservationTypeID: 4},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !tx.committed {
		t.Fatal("transaction was not committed")
	}
	if len(result.Counts) != len(CutoverTableSpecs()) {
		t.Fatalf("imported table counts = %d, want %d", len(result.Counts), len(CutoverTableSpecs()))
	}
	for table, count := range result.Counts {
		if count != 1 {
			t.Fatalf("table %s count = %d, want 1", table, count)
		}
	}
	if tx.copyCalls != len(CutoverTableSpecs()) || tx.setvalCalls != len(CutoverTableSpecs()) {
		t.Fatalf("copy calls=%d setval calls=%d, want %d each", tx.copyCalls, tx.setvalCalls, len(CutoverTableSpecs()))
	}
}

func TestApplyCutoverUsesForceNotNullForDeclaredTextColumns(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	tx := &fakeCutoverTransaction{}

	_, err := applyCutoverWithBegin(
		context.Background(),
		func(context.Context) (cutoverTransaction, error) { return tx, nil },
		bundle,
		CutoverSeedIDs{ClinicID: 1, AnimalSpeciesID: 2, ExamTypeID: 3, TrimmingReservationTypeID: 4},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(tx.copySQLs) != len(CutoverTableSpecs()) {
		t.Fatalf("copy SQL count = %d, want %d", len(tx.copySQLs), len(CutoverTableSpecs()))
	}
	if !strings.Contains(tx.copySQLs[0], `FORCE_NOT_NULL ("name", "license_number")`) {
		t.Fatalf("staffs COPY does not preserve required empty text: %s", tx.copySQLs[0])
	}
	if strings.Contains(tx.copySQLs[9], "FORCE_NOT_NULL") {
		t.Fatalf("appointments COPY unexpectedly forces nullable values: %s", tx.copySQLs[9])
	}
}

func TestApplyCutoverWithBeginFailsClosedAcrossTransactionStages(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	seeds := CutoverSeedIDs{ClinicID: 1, AnimalSpeciesID: 2, ExamTypeID: 3, TrimmingReservationTypeID: 4}
	tests := []struct {
		name        string
		tx          *fakeCutoverTransaction
		seeds       CutoverSeedIDs
		beginErr    error
		want        string
		wantUnknown bool
	}{
		{name: "begin", tx: &fakeCutoverTransaction{}, seeds: seeds, beginErr: errors.New("begin failed"), want: "begin cutover"},
		{name: "lock timeout setup", tx: &fakeCutoverTransaction{execErrorContains: "SET LOCAL"}, seeds: seeds, want: "lock timeout"},
		{name: "advisory lock", tx: &fakeCutoverTransaction{execErrorContains: "pg_advisory"}, seeds: seeds, want: "acquire cutover lock"},
		{name: "table lock", tx: &fakeCutoverTransaction{execErrorContains: "LOCK TABLE"}, seeds: seeds, want: "lock cutover tables"},
		{name: "target seed", tx: &fakeCutoverTransaction{}, seeds: CutoverSeedIDs{}, want: "explicit target seed"},
		{name: "copy", tx: &fakeCutoverTransaction{copyError: errors.New("private-owner-name")}, seeds: seeds, want: "target database rejected"},
		{name: "row verification", tx: &fakeCutoverTransaction{countMismatch: true}, seeds: seeds, want: "row count"},
		{name: "sequence advance", tx: &fakeCutoverTransaction{execErrorContains: "setval"}, seeds: seeds, want: "advance sequence"},
		{name: "sequence verification", tx: &fakeCutoverTransaction{sequenceBelowFloor: true}, seeds: seeds, want: "below application floor"},
		{name: "commit rollback", tx: &fakeCutoverTransaction{commitError: pgx.ErrTxCommitRollback}, seeds: seeds, want: "transaction rolled back"},
		{name: "commit unknown", tx: &fakeCutoverTransaction{commitError: errors.New("private-owner-name")}, seeds: seeds, want: "read-only verify", wantUnknown: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := applyCutoverWithBegin(
				context.Background(),
				func(context.Context) (cutoverTransaction, error) {
					if tt.beginErr != nil {
						return nil, tt.beginErr
					}
					return tt.tx, nil
				},
				bundle,
				tt.seeds,
			)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
			if strings.Contains(err.Error(), "private-owner-name") {
				t.Fatalf("error leaked private value: %v", err)
			}
			if errors.Is(err, ErrCutoverCommitOutcomeUnknown) != tt.wantUnknown {
				t.Fatalf("unknown commit classification = %v", errors.Is(err, ErrCutoverCommitOutcomeUnknown))
			}
			if tt.name == "begin" && !errors.Is(err, ErrCutoverTransactionNotStarted) {
				t.Fatalf("begin failure classification = %v", err)
			}
		})
	}
}

func validCutoverBundleForApply(t *testing.T) CutoverBundle {
	t.Helper()
	dir, digest := writeCutoverFixture(t, nil)
	bundle, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
		ManifestSHA256: digest,
		ClinicCode:     "hachioji",
		ClinicOrdinal:  1,
		RunID:          "run-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

type fakeCutoverTransaction struct {
	committed          bool
	rolledBack         bool
	copyCalls          int
	copySQLs           []string
	setvalCalls        int
	execErrorContains  string
	copyError          error
	commitError        error
	countMismatch      bool
	sequenceBelowFloor bool
}

func (tx *fakeCutoverTransaction) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	switch {
	case strings.Contains(query, "clinic_id <> $3"):
		return staticRow{values: []any{int64(0)}}
	case strings.Contains(query, "count(*)"):
		if tx.countMismatch {
			return staticRow{values: []any{int64(0)}}
		}
		return staticRow{values: []any{int64(1)}}
	case strings.Contains(query, "SELECT last_value FROM"):
		return staticRow{values: []any{int64(1)}}
	case strings.Contains(query, "SELECT COALESCE(max(id), 0)"):
		return staticRow{values: []any{int64(1_000_001)}}
	case strings.Contains(query, "last_value, is_called") && tx.sequenceBelowFloor:
		return staticRow{values: []any{int64(1), false}}
	default:
		return validTargetQuerier{}.QueryRow(ctx, query, args...)
	}
}

func (tx *fakeCutoverTransaction) Exec(_ context.Context, query string, _ ...any) (pgconn.CommandTag, error) {
	if tx.execErrorContains != "" && strings.Contains(query, tx.execErrorContains) {
		return pgconn.CommandTag{}, errors.New("forced execution failure")
	}
	if strings.Contains(query, "setval") {
		tx.setvalCalls++
	}
	return pgconn.NewCommandTag("SELECT 1"), nil
}

func (tx *fakeCutoverTransaction) CopyFrom(_ context.Context, reader io.Reader, copySQL string) (pgconn.CommandTag, error) {
	if tx.copyError != nil {
		return pgconn.CommandTag{}, tx.copyError
	}
	records, err := csv.NewReader(reader).ReadAll()
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	rows := len(records) - 1
	tx.copyCalls++
	tx.copySQLs = append(tx.copySQLs, copySQL)
	return pgconn.NewCommandTag(fmt.Sprintf("COPY %d", rows)), nil
}

func (tx *fakeCutoverTransaction) Commit(context.Context) error {
	if tx.commitError != nil {
		return tx.commitError
	}
	tx.committed = true
	return nil
}

func (tx *fakeCutoverTransaction) Rollback(context.Context) error {
	tx.rolledBack = true
	return nil
}

func TestTransformCutoverCSVResolvesOnlyDeclaredPlaceholders(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pets.csv")
	source := "id,clinic_id,owner_id,animal_species_id,name\n1000001,{{CLINIC_ID}},300001,{{FALLBACK_ANIMAL_SPECIES_ID}},Pochi\n"
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(source))
	var output bytes.Buffer
	count, err := transformCutoverCSV(context.Background(), path, &output, CutoverSeedIDs{
		ClinicID: 11, AnimalSpeciesID: 22, ExamTypeID: 33, TrimmingReservationTypeID: 44,
	}, hex.EncodeToString(sum[:]))
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
	if got := output.String(); got != "id,clinic_id,owner_id,animal_species_id,name\n1000001,11,300001,22,Pochi\n" {
		t.Fatalf("transformed CSV = %q", got)
	}
}

func TestTransformCutoverCSVRejectsDigestChangeWithoutLeakingValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "owners.csv")
	if err := os.WriteFile(path, []byte("id,name\n300001,private-owner-name\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	_, err := transformCutoverCSV(context.Background(), path, &output, CutoverSeedIDs{}, strings.Repeat("0", 64))
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "sha256") {
		t.Fatalf("error = %v, want sha256 rejection", err)
	}
	if strings.Contains(err.Error(), "private-owner-name") {
		t.Fatalf("error leaked CSV value: %v", err)
	}
}

func TestCutoverCopyResultErrorNeverLeaksDatabaseValue(t *testing.T) {
	err := cutoverCopyResultError("owners", errors.New("private-owner-name"), nil)
	if err == nil || !strings.Contains(err.Error(), "target database rejected") {
		t.Fatalf("error = %v, want sanitized target rejection", err)
	}
	if strings.Contains(err.Error(), "private-owner-name") {
		t.Fatalf("database value leaked: %v", err)
	}
}

func TestCutoverCopyResultErrorPrioritizesSanitizedDatabaseFailure(t *testing.T) {
	err := cutoverCopyResultError(
		"owners",
		errors.New("private-owner-name"),
		errors.New("io: read/write on closed pipe"),
	)
	if err == nil || !strings.Contains(err.Error(), "target database rejected") {
		t.Fatalf("error = %v, want sanitized target rejection", err)
	}
	if strings.Contains(err.Error(), "private-owner-name") {
		t.Fatalf("database value leaked: %v", err)
	}
}

func TestValidateCutoverSeedFacts(t *testing.T) {
	seeds := CutoverSeedIDs{ClinicID: 1, AnimalSpeciesID: 2, ExamTypeID: 3, TrimmingReservationTypeID: 4}
	valid := cutoverSeedFacts{
		ClinicExists: true, SpeciesActive: true,
		ExamTypeClinicID: 1, ExamTypeName: "検査", ExamTypeActive: true,
		ReservationTypeClinicID: 1, ReservationTypeCategory: "trimming", ReservationTypeActive: true,
	}
	if err := validateCutoverSeedFacts(seeds, valid); err != nil {
		t.Fatalf("valid facts rejected: %v", err)
	}

	invalid := valid
	invalid.ReservationTypeClinicID = 9
	if err := validateCutoverSeedFacts(seeds, invalid); err == nil || !strings.Contains(err.Error(), "reservation type") {
		t.Fatalf("cross-clinic reservation type error = %v", err)
	}
}

func TestPreflightAndVerifyCutoverTarget(t *testing.T) {
	manifest := CutoverManifest{
		IDBand: CutoverIDBand{
			Base: 0, EndExclusive: 10_000_000, NonOwnerIDOffset: 1_000_000,
			OwnerFloor: 300_000, ApplicationIDFloor: applicationIDFloor,
		},
	}
	for _, spec := range CutoverTableSpecs() {
		manifest.Tables = append(manifest.Tables, CutoverManifestTable{Table: spec.Name, File: spec.Name + ".csv"})
	}
	seeds := CutoverSeedIDs{ClinicID: 1, AnimalSpeciesID: 2, ExamTypeID: 3, TrimmingReservationTypeID: 4}
	target := validTargetQuerier{}
	if err := PreflightCutoverTarget(context.Background(), target, manifest, seeds); err != nil {
		t.Fatalf("PreflightCutoverTarget() error = %v", err)
	}
	if err := VerifyCutover(context.Background(), target, manifest, seeds); err != nil {
		t.Fatalf("VerifyCutover() error = %v", err)
	}
}

func TestPreflightAllowsLowSequenceButVerifyRequiresApplicationFloor(t *testing.T) {
	manifest := CutoverManifest{
		IDBand: CutoverIDBand{
			Base: 0, EndExclusive: 10_000_000, NonOwnerIDOffset: 1_000_000,
			OwnerFloor: 300_000, ApplicationIDFloor: applicationIDFloor,
		},
	}
	for _, spec := range CutoverTableSpecs() {
		manifest.Tables = append(manifest.Tables, CutoverManifestTable{Table: spec.Name, File: spec.Name + ".csv"})
	}
	seeds := CutoverSeedIDs{ClinicID: 1, AnimalSpeciesID: 2, ExamTypeID: 3, TrimmingReservationTypeID: 4}
	target := lowSequenceTargetQuerier{}
	if err := PreflightCutoverTarget(context.Background(), target, manifest, seeds); err != nil {
		t.Fatalf("preflight rejected an existing sequence before advance: %v", err)
	}
	if err := VerifyCutover(context.Background(), target, manifest, seeds); err == nil || !strings.Contains(err.Error(), "below application floor") {
		t.Fatalf("VerifyCutover() error = %v, want sequence floor rejection", err)
	}
}

func TestPreflightRejectsMissingValidatedForeignKey(t *testing.T) {
	manifest := CutoverManifest{
		IDBand: CutoverIDBand{
			Base: 0, EndExclusive: 10_000_000, NonOwnerIDOffset: 1_000_000,
			OwnerFloor: 300_000, ApplicationIDFloor: applicationIDFloor,
		},
	}
	for _, spec := range CutoverTableSpecs() {
		manifest.Tables = append(manifest.Tables, CutoverManifestTable{Table: spec.Name, File: spec.Name + ".csv"})
	}
	seeds := CutoverSeedIDs{ClinicID: 1, AnimalSpeciesID: 2, ExamTypeID: 3, TrimmingReservationTypeID: 4}
	err := PreflightCutoverTarget(context.Background(), missingForeignKeyTargetQuerier{}, manifest, seeds)
	if err == nil || !strings.Contains(err.Error(), "validated foreign key") {
		t.Fatalf("PreflightCutoverTarget() error = %v, want foreign-key rejection", err)
	}
}

type validTargetQuerier struct{}

func (validTargetQuerier) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	switch {
	case strings.Contains(query, "WITH\n  clinic_seed"):
		return staticRow{values: []any{true, true, int64(1), "検査", true, int64(1), "trimming", true}}
	case strings.Contains(query, "information_schema.columns"):
		return staticRow{values: []any{"bigint"}}
	case strings.Contains(query, "FROM pg_constraint"):
		return staticRow{values: []any{true}}
	case strings.Contains(query, "pg_get_serial_sequence"):
		return staticRow{values: []any{"public.fixture_id_seq"}}
	case strings.Contains(query, "last_value, is_called"):
		return staticRow{values: []any{applicationIDFloor, false}}
	case strings.Contains(query, "clinic_id <> $3"):
		return staticRow{values: []any{int64(0)}}
	case strings.Contains(query, "count(*)"):
		return staticRow{values: []any{int64(0)}}
	case strings.Contains(query, "SELECT EXISTS"):
		return staticRow{values: []any{false}}
	default:
		return staticRow{err: errUnexpectedQuery}
	}
}

type missingForeignKeyTargetQuerier struct{}

func (missingForeignKeyTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "FROM pg_constraint") {
		return staticRow{values: []any{false}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

type lowSequenceTargetQuerier struct{}

func (lowSequenceTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "last_value, is_called") {
		return staticRow{values: []any{int64(1), false}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

var errUnexpectedQuery = &unexpectedQueryError{}

type unexpectedQueryError struct{}

func (*unexpectedQueryError) Error() string { return "unexpected query" }

type staticRow struct {
	values []any
	err    error
}

func (row staticRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != len(row.values) {
		return errUnexpectedQuery
	}
	for i, value := range row.values {
		switch destination := destinations[i].(type) {
		case *bool:
			*destination = value.(bool)
		case *int64:
			*destination = value.(int64)
		case *string:
			*destination = value.(string)
		case **string:
			copy := value.(string)
			*destination = &copy
		default:
			return errUnexpectedQuery
		}
	}
	return nil
}
