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
		validCutoverSeeds(),
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
		validCutoverSeeds(),
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
	var paymentsCopySQL string
	for _, copySQL := range tx.copySQLs {
		if strings.Contains(copySQL, `COPY "payments"`) {
			paymentsCopySQL = copySQL
			break
		}
	}
	if !strings.Contains(paymentsCopySQL, `FORCE_NOT_NULL ("insurance_name")`) {
		t.Fatalf("payments COPY does not preserve required empty text: %s", paymentsCopySQL)
	}
}

func TestApplyCutoverWithBeginFailsClosedAcrossTransactionStages(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	seeds := validCutoverSeeds()
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
		{name: "payment graph", tx: &fakeCutoverTransaction{paymentGraphMismatch: true}, seeds: seeds, want: "payment graph"},
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
			if tt.name == "payment graph" && (tt.tx.committed || !tt.tx.rolledBack) {
				t.Fatalf("payment graph failure transaction state committed=%v rolledBack=%v", tt.tx.committed, tt.tx.rolledBack)
			}
		})
	}
}

func TestApplyCutoverRollsBackWhenPaymentSplitCopyFails(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	tx := &fakeCutoverTransaction{copyErrorContains: `"payment_splits"`}

	_, err := applyCutoverWithBegin(
		context.Background(),
		func(context.Context) (cutoverTransaction, error) { return tx, nil },
		bundle,
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), "payment_splits") {
		t.Fatalf("error = %v, want payment_splits COPY rejection", err)
	}
	if tx.committed || !tx.rolledBack {
		t.Fatalf("transaction state committed=%v rolledBack=%v", tx.committed, tx.rolledBack)
	}
	if tx.copyCalls != 14 {
		t.Fatalf("successful COPY calls before payment_splits = %d, want 14", tx.copyCalls)
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
	committed            bool
	rolledBack           bool
	copyCalls            int
	copySQLs             []string
	setvalCalls          int
	execErrorContains    string
	copyErrorContains    string
	copyError            error
	commitError          error
	countMismatch        bool
	sequenceBelowFloor   bool
	paymentGraphMismatch bool
}

func (tx *fakeCutoverTransaction) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	switch {
	case strings.Contains(query, "required animal_species master"):
		return staticRow{values: []any{int64(0), int64(0), int64(0), int64(6)}}
	case strings.Contains(query, "clinic_seed AS MATERIALIZED"):
		return validTargetQuerier{}.QueryRow(ctx, query, args...)
	case strings.Contains(query, "FROM payment_splits split") && tx.paymentGraphMismatch:
		return staticRow{values: []any{int64(1)}}
	case strings.Contains(query, "clinic_id <> $3"):
		return staticRow{values: []any{int64(0)}}
	case strings.Contains(query, "FROM payment_splits split"):
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
	if tx.copyErrorContains != "" && strings.Contains(copySQL, tx.copyErrorContains) {
		return pgconn.CommandTag{}, errors.New("forced payment copy failure")
	}
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
		CashPaymentMethodID: 55, CreditCardPaymentMethodID: 66,
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

func TestTransformCutoverCSVResolvesPaymentMethodPlaceholders(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "payments.csv")
	source := "id,method,payment_method_id\n1000001,cash,{{PAYMENT_METHOD_CASH_ID}}\n1000002,credit_card,{{PAYMENT_METHOD_CREDIT_CARD_ID}}\n"
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(source))
	var output bytes.Buffer
	count, err := transformCutoverCSV(context.Background(), path, &output, CutoverSeedIDs{
		ClinicID: 11, AnimalSpeciesID: 22, ExamTypeID: 33, TrimmingReservationTypeID: 44,
		CashPaymentMethodID: 55, CreditCardPaymentMethodID: 66,
	}, hex.EncodeToString(sum[:]))
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
	want := "id,method,payment_method_id\n1000001,cash,55\n1000002,credit_card,66\n"
	if got := output.String(); got != want {
		t.Fatalf("transformed CSV = %q, want %q", got, want)
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
	seeds := validCutoverSeeds()
	valid := cutoverSeedFacts{
		ClinicExists: true, SpeciesActive: true,
		ExamTypeClinicID: 1, ExamTypeName: "検査", ExamTypeActive: true,
		ReservationTypeClinicID: 1, ReservationTypeCategory: "trimming", ReservationTypeActive: true,
		CashMethodClinicID: 1, CashMethodSystemKey: "cash", CashMethodActive: true, CashMethodMatchCount: 1,
		CreditCardMethodClinicID: 1, CreditCardMethodSystemKey: "credit_card",
		CreditCardMethodActive: true, CreditCardMethodMatchCount: 1,
	}
	if err := validateCutoverSeedFacts(seeds, valid); err != nil {
		t.Fatalf("valid facts rejected: %v", err)
	}

	invalid := valid
	invalid.ReservationTypeClinicID = 9
	if err := validateCutoverSeedFacts(seeds, invalid); err == nil || !strings.Contains(err.Error(), "reservation type") {
		t.Fatalf("cross-clinic reservation type error = %v", err)
	}

	invalid = valid
	invalid.CashMethodMatchCount = 2
	if err := validateCutoverSeedFacts(seeds, invalid); err == nil || !strings.Contains(err.Error(), "unique cash") {
		t.Fatalf("duplicate cash payment method error = %v", err)
	}

	invalid = valid
	invalid.CreditCardMethodClinicID = 9
	if err := validateCutoverSeedFacts(seeds, invalid); err == nil || !strings.Contains(err.Error(), "credit-card payment method") {
		t.Fatalf("cross-clinic credit-card payment method error = %v", err)
	}

	invalid = valid
	invalid.SpeciesActive = false
	if err := validateCutoverSeedFacts(seeds, invalid); err == nil || !strings.Contains(err.Error(), "fallback animal species") {
		t.Fatalf("fallback animal species error = %v", err)
	}
}

func TestRequiredAnimalSpeciesRowsMatchSeedContract(t *testing.T) {
	want := []requiredAnimalSpeciesRow{
		{1, "犬"}, {2, "猫"}, {3, "鳥"}, {4, "うさぎ"}, {5, "ハムスター"}, {6, "その他"},
	}
	got := requiredAnimalSpeciesRows()
	if len(got) != len(want) {
		t.Fatalf("requiredAnimalSpeciesRows() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("requiredAnimalSpeciesRows()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestValidateRequiredAnimalSpeciesFacts(t *testing.T) {
	tests := []struct {
		name  string
		facts requiredAnimalSpeciesFacts
		want  string
	}{
		{name: "all six exact active", facts: requiredAnimalSpeciesFacts{ExactActiveMatchCount: 6}},
		{name: "missing", facts: requiredAnimalSpeciesFacts{MissingCount: 1, ExactActiveMatchCount: 5}, want: "missing"},
		{name: "inactive", facts: requiredAnimalSpeciesFacts{InactiveCount: 1, ExactActiveMatchCount: 5}, want: "inactive"},
		{name: "renamed", facts: requiredAnimalSpeciesFacts{RenamedCount: 1, ExactActiveMatchCount: 5}, want: "unexpected names"},
		{name: "inconsistent match count", facts: requiredAnimalSpeciesFacts{ExactActiveMatchCount: 5}, want: "incomplete or inconsistent"},
		{name: "zero matches", facts: requiredAnimalSpeciesFacts{}, want: "incomplete or inconsistent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredAnimalSpeciesFacts(tt.facts)
			if tt.want == "" {
				if err != nil {
					t.Fatalf("validateRequiredAnimalSpeciesFacts() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("validateRequiredAnimalSpeciesFacts() error = %v, want %q", err, tt.want)
			}
			if strings.Contains(err.Error(), "private-owner-name") {
				t.Fatalf("sensitive payload leaked: %v", err)
			}
		})
	}
}

func TestRequiredAnimalSpeciesQueryIsParameterizedAndShareLocked(t *testing.T) {
	for _, fragment := range []string{
		"required animal_species master",
		"FOR SHARE",
		"AS MATERIALIZED",
		"animal_species",
		"IS DISTINCT FROM",
		"$1::bigint",
		"$12::text",
	} {
		if !strings.Contains(requiredAnimalSpeciesQuery, fragment) {
			t.Fatalf("required animal species query missing %q", fragment)
		}
	}
	for _, forbidden := range []string{"private-owner-name", "password", "postgres://"} {
		if strings.Contains(requiredAnimalSpeciesQuery, forbidden) {
			t.Fatalf("required animal species query contains forbidden fragment %q", forbidden)
		}
	}
}

func TestPreflightRejectsMissingRequiredAnimalSpecies(t *testing.T) {
	err := PreflightCutoverTarget(
		context.Background(),
		missingRequiredAnimalSpeciesTargetQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("PreflightCutoverTarget() error = %v, want missing species rejection", err)
	}
}

func TestPreflightRejectsInactiveRequiredAnimalSpecies(t *testing.T) {
	err := PreflightCutoverTarget(
		context.Background(),
		inactiveRequiredAnimalSpeciesTargetQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), "inactive") {
		t.Fatalf("PreflightCutoverTarget() error = %v, want inactive species rejection", err)
	}
}

func TestPreflightRejectsRenamedRequiredAnimalSpecies(t *testing.T) {
	err := PreflightCutoverTarget(
		context.Background(),
		renamedRequiredAnimalSpeciesTargetQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), "unexpected names") {
		t.Fatalf("PreflightCutoverTarget() error = %v, want renamed species rejection", err)
	}
}

func TestRequiredAnimalSpeciesRejectionErrorsDoNotLeakPrivatePayload(t *testing.T) {
	for _, tc := range []struct {
		name    string
		querier cutoverQuerier
		want    string
	}{
		{name: "missing", querier: missingRequiredAnimalSpeciesTargetQuerier{}, want: "missing"},
		{name: "inactive", querier: inactiveRequiredAnimalSpeciesTargetQuerier{}, want: "inactive"},
		{name: "renamed", querier: renamedRequiredAnimalSpeciesTargetQuerier{}, want: "unexpected names"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := PreflightCutoverTarget(
				context.Background(),
				tc.querier,
				cutoverManifestForTargetTests(),
				validCutoverSeeds(),
			)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("PreflightCutoverTarget() error = %v, want %q", err, tc.want)
			}
			if strings.Contains(err.Error(), "private-owner-name") ||
				strings.Contains(err.Error(), "password") ||
				strings.Contains(err.Error(), "postgres://") {
				t.Fatalf("rejection error leaked sensitive payload: %v", err)
			}
		})
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
	seeds := validCutoverSeeds()
	target := validTargetQuerier{}
	if err := PreflightCutoverTarget(context.Background(), target, manifest, seeds); err != nil {
		t.Fatalf("PreflightCutoverTarget() error = %v", err)
	}
	if err := VerifyCutover(context.Background(), target, manifest, seeds); err != nil {
		t.Fatalf("VerifyCutover() error = %v", err)
	}
	// Verify is idempotent/read-only: a second call against the same snapshot must pass.
	if err := VerifyCutover(context.Background(), target, manifest, seeds); err != nil {
		t.Fatalf("second VerifyCutover() error = %v (must be idempotent)", err)
	}
}

// TestPreflightRejectsOccupiedClinicBand proves apply rerun cannot double-create:
// a non-empty clinic band fails closed at preflight (Issue #250 idempotency).
func TestPreflightRejectsOccupiedClinicBand(t *testing.T) {
	err := PreflightCutoverTarget(
		context.Background(),
		occupiedBandTargetQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("PreflightCutoverTarget() error = %v, want %s", err, CutoverRefBandOccupied)
	}
	if !strings.Contains(err.Error(), "already occupied") {
		t.Fatalf("PreflightCutoverTarget() error = %v, want occupied-band message", err)
	}
	// Non-PHI: no private payloads or numeric clinic IDs from the hostile fixture.
	for _, leaked := range []string{"private-owner-name", "private-clinic", "99999"} {
		if strings.Contains(err.Error(), leaked) {
			t.Fatalf("occupied-band error leaked %q: %v", leaked, err)
		}
	}
}

// TestVerifyRejectsCrossClinicAssignment proves clinic isolation on every
// clinic_id-bearing table (Issue #250 owner/pet/clinic/staff boundary).
func TestVerifyRejectsCrossClinicAssignment(t *testing.T) {
	err := VerifyCutover(
		context.Background(),
		crossClinicTargetQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), CutoverRefClinicIsolation) {
		t.Fatalf("VerifyCutover() error = %v, want %s", err, CutoverRefClinicIsolation)
	}
	if !strings.Contains(err.Error(), "another clinic") {
		t.Fatalf("VerifyCutover() error = %v, want cross-clinic message", err)
	}
	for _, leaked := range []string{"private-owner-name", "private-pet-name", "clinic-9-name"} {
		if strings.Contains(err.Error(), leaked) {
			t.Fatalf("clinic-isolation error leaked %q: %v", leaked, err)
		}
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
	seeds := validCutoverSeeds()
	target := lowSequenceTargetQuerier{}
	if err := PreflightCutoverTarget(context.Background(), target, manifest, seeds); err != nil {
		t.Fatalf("preflight rejected an existing sequence before advance: %v", err)
	}
	if err := VerifyCutover(context.Background(), target, manifest, seeds); err == nil || !strings.Contains(err.Error(), "below application floor") {
		t.Fatalf("VerifyCutover() error = %v, want sequence floor rejection", err)
	}
}

func TestVerifyRequiresSequenceToExceedCurrentMaxID(t *testing.T) {
	err := VerifyCutover(
		context.Background(),
		staleSequenceTargetQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), "does not exceed") {
		t.Fatalf("VerifyCutover() error = %v, want max-id rejection", err)
	}
}

func TestPreflightRejectsMissingValidatedForeignKey(t *testing.T) {
	err := PreflightCutoverTarget(
		context.Background(),
		missingForeignKeyTargetQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), "validated foreign key") {
		t.Fatalf("PreflightCutoverTarget() error = %v, want foreign-key rejection", err)
	}
}

func TestPreflightRejectsMissingValidatedCompositeForeignKey(t *testing.T) {
	err := PreflightCutoverTarget(
		context.Background(),
		missingCompositeForeignKeyTargetQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), "validated composite foreign key") {
		t.Fatalf("PreflightCutoverTarget() error = %v, want composite foreign-key rejection", err)
	}
	if !strings.Contains(err.Error(), "payments(") {
		t.Fatalf("PreflightCutoverTarget() error = %v, want payments composite naming", err)
	}
}

func TestCutoverRequiredForeignKeysIncludePaymentContract(t *testing.T) {
	required := map[string]bool{
		"payments.clinic_id->clinics.id":                       false,
		"payments.billing_id->billings.id":                     false,
		"payments.payment_method_id->payment_methods.id":       false,
		"payments.paid_by->staffs.id":                          false,
		"payment_splits.clinic_id->clinics.id":                 false,
		"payment_splits.billing_id->billings.id":               false,
		"payment_splits.payment_method_id->payment_methods.id": false,
		"payment_splits.paid_by->staffs.id":                    false,
		"appointments.owner_id->owners.id":                     false,
	}
	for _, foreignKey := range cutoverRequiredForeignKeys() {
		key := foreignKey.childTable + "." + foreignKey.childColumn + "->" +
			foreignKey.parentTable + "." + foreignKey.parentColumn
		if _, ok := required[key]; ok {
			required[key] = true
		}
	}
	for key, found := range required {
		if !found {
			t.Errorf("missing required payment foreign key %s", key)
		}
	}
}

func TestCutoverRequiredCompositeForeignKeysIncludePaymentClinicAxis(t *testing.T) {
	required := map[string]bool{
		"payments(billing_id, clinic_id)->billings(id, clinic_id)":               false,
		"payments(payment_method_id, clinic_id)->payment_methods(id, clinic_id)": false,
	}
	for _, foreignKey := range cutoverRequiredCompositeForeignKeys() {
		key := foreignKey.childTable + "(" + strings.Join(foreignKey.childColumns, ", ") + ")->" +
			foreignKey.parentTable + "(" + strings.Join(foreignKey.parentColumns, ", ") + ")"
		if _, ok := required[key]; ok {
			required[key] = true
		}
	}
	for key, found := range required {
		if !found {
			t.Errorf("missing required composite foreign key %s", key)
		}
	}
	for _, fragment := range []string{
		"array_agg(a.attname ORDER BY cols.ordinality)",
		"c.convalidated = true",
		"c.conenforced = true",
		"unnest(c.conkey) WITH ORDINALITY",
		"unnest(c.confkey) WITH ORDINALITY",
	} {
		if !strings.Contains(cutoverCompositeForeignKeyQuery, fragment) {
			t.Errorf("composite foreign-key query is missing %q", fragment)
		}
	}
}

func TestPreflightRejectsMissingRequiredTargetColumn(t *testing.T) {
	err := PreflightCutoverTarget(
		context.Background(),
		missingRequiredTargetColumnQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), "staffs.missing_required_column") {
		t.Fatalf("PreflightCutoverTarget() error = %v, want staffs.missing_required_column", err)
	}
}

func TestPreflightAcceptsNullableOrDefaultedTargetColumnsOutsideSpec(t *testing.T) {
	// validTargetQuerier returns no NOT NULL/no-default columns, modeling a
	// target where only nullable or defaulted columns exist outside the CSV.
	err := PreflightCutoverTarget(
		context.Background(),
		validTargetQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err != nil {
		t.Fatalf("PreflightCutoverTarget() error = %v, want success for nullable/defaulted extras", err)
	}
}

func TestRequiredTargetColumnsQueryExcludesIdentityGeneratedAndDefaults(t *testing.T) {
	for _, fragment := range []string{
		"is_nullable = 'NO'",
		"column_default IS NULL",
		"COALESCE(is_identity, 'NO') = 'NO'",
		"COALESCE(is_generated, 'NEVER') = 'NEVER'",
	} {
		if !strings.Contains(cutoverRequiredTargetColumnsQuery, fragment) {
			t.Errorf("required target columns query is missing %q", fragment)
		}
	}
}

func TestPreflightRejectsMissingPaymentBillingUniqueConstraint(t *testing.T) {
	err := PreflightCutoverTarget(
		context.Background(),
		missingPaymentUniqueTargetQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), "payments.billing_id") {
		t.Fatalf("PreflightCutoverTarget() error = %v, want payments.billing_id uniqueness rejection", err)
	}
}

func TestPreflightRejectsMissingPaymentMethodSystemKeyUniqueIndex(t *testing.T) {
	err := PreflightCutoverTarget(
		context.Background(),
		missingPaymentMethodUniqueTargetQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), "payment_methods") {
		t.Fatalf("PreflightCutoverTarget() error = %v, want payment_methods uniqueness rejection", err)
	}
}

func TestPreflightRejectsMissingPaymentSplitsBillingIndex(t *testing.T) {
	err := PreflightCutoverTarget(
		context.Background(),
		missingPaymentSplitsBillingIndexTargetQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), "payment_splits.billing_id") {
		t.Fatalf("PreflightCutoverTarget() error = %v, want payment_splits index rejection", err)
	}
}

func TestVerifyRejectsPaymentSplitWithoutPaymentParent(t *testing.T) {
	err := VerifyCutover(
		context.Background(),
		missingPaymentParentTargetQuerier{},
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	)
	if err == nil || !strings.Contains(err.Error(), "payment parent") {
		t.Fatalf("VerifyCutover() error = %v, want payment parent rejection", err)
	}
}

func TestPaymentGraphVerificationQueryFailsClosedForNullsAndOutsideBandSplits(t *testing.T) {
	required := []string{
		"LEFT JOIN billings billing",
		"split_rows AS MATERIALIZED",
		"split_summaries AS MATERIALIZED",
		"JOIN payment_rows payment ON payment.billing_id = split.billing_id",
		"payment.deleted_at IS NOT NULL",
		"payment.parent_billing_id < $1",
		"payment.billing_deleted_at IS NOT NULL",
		"payment.clinic_id IS NULL",
		"payment.clinic_id IS DISTINCT FROM $3",
		"payment.clinic_id IS DISTINCT FROM payment.billing_clinic_id",
		"payment.billing_clinic_id IS DISTINCT FROM $3",
		"payment.billing_total_amount IS DISTINCT FROM payment.total_amount",
		"payment.subtotal IS NULL",
		"payment.tax_total IS NULL",
		"payment.total_amount IS NULL",
		"payment.insurance_ratio IS NULL",
		"payment.insurance_ratio > 1",
		"payment.insurance_amount IS NULL",
		"payment.insurance_amount < 0",
		"payment.discount_amount IS NULL",
		"payment.billing_amount IS NULL",
		"payment.billing_amount = 0",
		"payment.received_amount IS NULL",
		"payment.change_amount IS NULL",
		"payment.created_at IS NULL",
		"payment.paid_by < $1",
		"NOT COALESCE(",
		"bool_and(COALESCE(",
		"split.id >= $1",
		"split.paid_by >= $1",
		"split.created_at IS NOT NULL",
		"split.received_amount IS NOT NULL",
		"split.change_amount IS NOT NULL",
		"payment.method IS DISTINCT FROM",
		"completed_billing_violations AS",
		"billing.status = 'completed'",
		"billing.status IS NULL",
		"billing.status NOT IN ('waiting', 'completed', 'cancelled', 'pending')",
		"payment.billing_status IS DISTINCT FROM 'completed'",
		"payment.billing_completed_at IS DISTINCT FROM payment.created_at",
		"COALESCE(billing.total_amount, 0) <> 0",
	}
	for _, fragment := range required {
		if !strings.Contains(verifyCutoverPaymentGraphQuery, fragment) {
			t.Errorf("payment verification query is missing fail-closed fragment %q", fragment)
		}
	}

	splitRowsStart := strings.Index(verifyCutoverPaymentGraphQuery, "split_rows AS MATERIALIZED")
	orphanStart := strings.Index(verifyCutoverPaymentGraphQuery, "orphan_split_violations AS")
	if splitRowsStart < 0 || orphanStart <= splitRowsStart {
		t.Fatal("payment verification query does not contain the expected graph sections")
	}
	splitAggregationQuery := verifyCutoverPaymentGraphQuery[splitRowsStart:orphanStart]
	if !strings.Contains(splitAggregationQuery, "split.id >= $1") ||
		!strings.Contains(splitAggregationQuery, "split.id < $2") {
		t.Fatal("parent split aggregation must reject child rows outside the cutover ID band")
	}
	if strings.Contains(splitAggregationQuery, "JOIN LATERAL") {
		t.Fatal("payment split aggregation must be set-based rather than correlated per payment")
	}
}

func TestVerifyCutoverPaymentGraphPreservesQueryErrorIdentity(t *testing.T) {
	queryErr := errors.New("query failed")
	manifest := cutoverManifestForTargetTests()
	err := verifyCutoverPaymentGraph(
		context.Background(),
		errorTargetQuerier{err: queryErr},
		&manifest,
		validCutoverSeeds(),
	)
	if !errors.Is(err, queryErr) {
		t.Fatalf("verifyCutoverPaymentGraph() error = %v, want wrapped query error", err)
	}
}

func TestUniqueIndexVerificationQueriesRejectWeakerExpressionAndPartialIndexes(t *testing.T) {
	for _, fragment := range []string{
		"index_definition.indexprs IS NULL",
		"index_definition.indnkeyatts = 1",
		"index_definition.indnatts = 1",
	} {
		if !strings.Contains(paymentsBillingUniqueIndexQuery, fragment) {
			t.Errorf("payments index query is missing %q", fragment)
		}
	}
	for _, fragment := range []string{
		"index_definition.indexprs IS NULL",
		"index_definition.indnkeyatts = 2",
		"index_definition.indnatts = 2",
		"'system_keyisnotnullanddeleted_atisnull'",
		"'deleted_atisnullandsystem_keyisnotnull'",
	} {
		if !strings.Contains(paymentMethodSystemKeyUniqueIndexQuery, fragment) {
			t.Errorf("payment method index query is missing %q", fragment)
		}
	}
	if strings.Contains(paymentMethodSystemKeyUniqueIndexQuery, "LIKE '%") {
		t.Fatal("payment method index predicate must not accept a substring match")
	}
	for _, fragment := range []string{
		"target_table.relname = 'payment_splits'",
		"index_definition.indpred IS NULL",
		"key_column.ordinality = 1",
		") = 'billing_id'",
	} {
		if !strings.Contains(paymentSplitsBillingIndexQuery, fragment) {
			t.Errorf("payment_splits index query is missing %q", fragment)
		}
	}
}

func cutoverManifestForTargetTests() CutoverManifest {
	manifest := CutoverManifest{
		IDBand: CutoverIDBand{
			Base: 0, EndExclusive: 10_000_000, NonOwnerIDOffset: 1_000_000,
			OwnerFloor: 300_000, ApplicationIDFloor: applicationIDFloor,
		},
	}
	for _, spec := range CutoverTableSpecs() {
		manifest.Tables = append(manifest.Tables, CutoverManifestTable{Table: spec.Name, File: spec.Name + ".csv"})
	}
	return manifest
}

func validCutoverSeeds() CutoverSeedIDs {
	return CutoverSeedIDs{
		ClinicID: 1, AnimalSpeciesID: 2, ExamTypeID: 3, TrimmingReservationTypeID: 4,
		CashPaymentMethodID: 5, CreditCardPaymentMethodID: 6,
	}
}

type validTargetQuerier struct{}

func (validTargetQuerier) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	switch {
	case strings.Contains(query, "required animal_species master"):
		return staticRow{values: []any{int64(0), int64(0), int64(0), int64(6)}}
	case strings.Contains(query, "clinic_seed AS MATERIALIZED"):
		return staticRow{values: []any{
			true, true, int64(1), "検査", true, int64(1), "trimming", true,
			int64(1), "cash", true, int64(1),
			int64(1), "credit_card", true, int64(1),
		}}
	case strings.Contains(query, "information_schema.columns") && strings.Contains(query, "is_nullable"):
		// No NOT NULL / no-default / non-identity / non-generated columns outside
		// the CSV contract: serial/identity ids and defaulted columns are filtered
		// by the query itself, so unit tests model an empty required set.
		return staticRow{values: []any{[]string{}}}
	case strings.Contains(query, "information_schema.columns"):
		return staticRow{values: []any{"bigint"}}
	case strings.Contains(query, "FROM pg_constraint"):
		return staticRow{values: []any{true}}
	case strings.Contains(query, "FROM pg_index"):
		return staticRow{values: []any{true}}
	case strings.Contains(query, "pg_get_serial_sequence"):
		return staticRow{values: []any{"public.fixture_id_seq"}}
	case strings.Contains(query, "last_value, is_called"):
		return staticRow{values: []any{applicationIDFloor, false}}
	case strings.Contains(query, "SELECT COALESCE(max(id), 0)"):
		return staticRow{values: []any{int64(0)}}
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

type missingRequiredAnimalSpeciesTargetQuerier struct{}

func (missingRequiredAnimalSpeciesTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "required animal_species master") {
		return staticRow{values: []any{int64(1), int64(0), int64(0), int64(5)}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

type inactiveRequiredAnimalSpeciesTargetQuerier struct{}

func (inactiveRequiredAnimalSpeciesTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "required animal_species master") {
		return staticRow{values: []any{int64(0), int64(1), int64(0), int64(5)}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

type renamedRequiredAnimalSpeciesTargetQuerier struct{}

func (renamedRequiredAnimalSpeciesTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "required animal_species master") {
		return staticRow{values: []any{int64(0), int64(0), int64(1), int64(5)}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

type errorTargetQuerier struct {
	err error
}

func (q errorTargetQuerier) QueryRow(context.Context, string, ...any) pgx.Row {
	return staticRow{err: q.err}
}

type missingForeignKeyTargetQuerier struct{}

func (missingForeignKeyTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "FROM pg_constraint") {
		return staticRow{values: []any{false}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

type missingCompositeForeignKeyTargetQuerier struct{}

func (missingCompositeForeignKeyTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	// Single-column FK preflight keeps array_length=1; composite uses ordered
	// array_agg over conkey/confkey and must fail independently when missing.
	if strings.Contains(query, "FROM pg_constraint") && strings.Contains(query, "array_agg") {
		return staticRow{values: []any{false}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

type missingRequiredTargetColumnQuerier struct{}

func (missingRequiredTargetColumnQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "information_schema.columns") && strings.Contains(query, "is_nullable") {
		return staticRow{values: []any{[]string{"missing_required_column"}}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

type missingPaymentUniqueTargetQuerier struct{}

func (missingPaymentUniqueTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "FROM pg_index") {
		return staticRow{values: []any{false}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

type missingPaymentMethodUniqueTargetQuerier struct{}

func (missingPaymentMethodUniqueTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "target_table.relname = 'payment_methods'") {
		return staticRow{values: []any{false}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

type missingPaymentSplitsBillingIndexTargetQuerier struct{}

func (missingPaymentSplitsBillingIndexTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "target_table.relname = 'payment_splits'") {
		return staticRow{values: []any{false}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

type missingPaymentParentTargetQuerier struct{}

func (missingPaymentParentTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "FROM payment_splits split") {
		return staticRow{values: []any{int64(1)}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

// occupiedBandTargetQuerier models a target whose clinic band already holds rows,
// so a second apply/preflight must fail closed (no silent double-create).
type occupiedBandTargetQuerier struct{}

func (occupiedBandTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "SELECT EXISTS") {
		return staticRow{values: []any{true}}
	}
	return validTargetQuerier{}.QueryRow(ctx, query, args...)
}

// crossClinicTargetQuerier reports band rows whose clinic_id does not match the
// operator-supplied seed clinic (cross-tenant assignment).
type crossClinicTargetQuerier struct{}

func (crossClinicTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "clinic_id <> $3") {
		return staticRow{values: []any{int64(1)}}
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

type staleSequenceTargetQuerier struct{}

func (staleSequenceTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "last_value, is_called") {
		return staticRow{values: []any{applicationIDFloor, false}}
	}
	if strings.Contains(query, "SELECT COALESCE(max(id), 0)") {
		return staticRow{values: []any{applicationIDFloor}}
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
		case *[]string:
			if value == nil {
				*destination = nil
				continue
			}
			*destination = append([]string(nil), value.([]string)...)
		default:
			return errUnexpectedQuery
		}
	}
	return nil
}
