package csvimport

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Codex 4266290429: DESIGN 3.2 keeps the legacy sort order, so every
// candidate medicines row must carry an explicit signed 32-bit decimal
// sort_order. Empty values and out-of-int32 integers fail source preflight
// with a table/column/row diagnostic; negatives stay legal and no zero fill
// or NULL coercion happens.
func TestMedicineFinalSortOrderInt32(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	sortCol := columnIndex(spec.Columns, "sort_order")

	for _, rejected := range []string{"", " ", "abc", "1.5", "2147483648", "-2147483649"} {
		t.Run("rejects "+rejected, func(t *testing.T) {
			dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
				f.rows["medicines"][0][sortCol] = rejected
			})
			_, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
			if err == nil {
				t.Fatalf("PreflightCutoverBundle accepted sort_order %q", rejected)
			}
			if !strings.Contains(err.Error(), "sort_order") {
				t.Fatalf("PreflightCutoverBundle error = %v, want sort_order diagnostic", err)
			}
			if len(rejected) > 1 && strings.Contains(err.Error(), rejected) {
				t.Fatal("error leaks the rejected cell value")
			}
		})
	}

	for _, accepted := range []string{"0", "2147483647", "-2147483648"} {
		t.Run("accepts "+accepted, func(t *testing.T) {
			dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
				f.rows["medicines"][0][sortCol] = accepted
			})
			if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err != nil {
				t.Fatalf("PreflightCutoverBundle error = %v, want sort_order %q accepted", err, accepted)
			}
		})
	}
}

// Codex 4266290426: a candidate medicines name that already exists in the
// same target clinic (deleted_at IS NULL, any ID band, inactive included)
// would violate the partial unique index after staff/procedures committed.
// Under user decision A the table-per-commit apply must detect this BEFORE
// the first Begin/COPY by re-reading the bundle's medicines CSV (row count +
// SHA-256 bound to the exact opened bytes).
type medicineNameProbeSession struct {
	candidateFreshApplySession
	probeSQL    string
	probeClinic int64
	probeNames  []string
	probes      int
	exists      bool
	probeErr    error
}

func (s *medicineNameProbeSession) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "public.medicines") && strings.Contains(query, "name = ANY") {
		s.probes++
		s.probeSQL = query
		if len(args) == 2 {
			s.probeClinic, _ = args[0].(int64)
			s.probeNames, _ = args[1].([]string)
		}
		if s.probeErr != nil {
			return staticRow{err: s.probeErr}
		}
		return staticRow{values: []any{s.exists}}
	}
	return s.candidateFreshApplySession.QueryRow(ctx, query, args...)
}

func TestMedicineFinalNameCollisionRejectedBeforeBegin(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	session := &medicineNameProbeSession{exists: true}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("applyCutoverCommittingEachTable() error = %v, want target name collision rejection", err)
	}
	if session.probes == 0 {
		t.Fatal("apply never probed candidate medicines names against the target")
	}
	if len(session.txs) != 0 {
		t.Fatalf("transactions = %d, want 0 before denial", len(session.txs))
	}
}

func TestMedicineFinalNameProbeBindsClinicAndExactNames(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	session := &medicineNameProbeSession{}
	if _, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead); err != nil {
		t.Fatal(err)
	}
	seeds := validCutoverSeeds()
	if session.probeClinic != seeds.ClinicID {
		t.Fatalf("probe clinic = %d, want %d", session.probeClinic, seeds.ClinicID)
	}
	if len(session.probeNames) != 1 || session.probeNames[0] != "candidate-medicine" {
		t.Fatalf("probe names = %v, want the exact verbatim source names", session.probeNames)
	}
	if !strings.Contains(session.probeSQL, "deleted_at IS NULL") {
		t.Fatal("probe query must exclude soft-deleted rows only")
	}
	if strings.Contains(session.probeSQL, "is_active") {
		t.Fatal("probe query must NOT filter on is_active: inactive rows also collide")
	}
}

func TestMedicineFinalNameProbeQueryErrorFailsBeforeBegin(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	session := &medicineNameProbeSession{probeErr: errors.New("catalog lookup failed")}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !errors.Is(err, session.probeErr) {
		t.Fatalf("applyCutoverCommittingEachTable() error = %v, want preserved probe cause", err)
	}
	if strings.Contains(err.Error(), "catalog lookup failed") {
		t.Fatalf("diagnostic leaks driver detail: %v", err)
	}
	if len(session.txs) != 0 {
		t.Fatalf("transactions = %d, want 0 before denial", len(session.txs))
	}
}

func TestMedicineFinalNameProbeAcceptsDistinctNames(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	// Same name in another clinic or soft-deleted is out of scope for the
	// probe; only an undeleted same-clinic match collides.
	session := &medicineNameProbeSession{}
	if _, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead); err != nil {
		t.Fatal(err)
	}
	wantTables := len(CutoverTableSpecs(CutoverProfileMedicinesCandidate))
	copyCalls := 0
	for _, tx := range session.txs {
		copyCalls += tx.copyCalls
	}
	if copyCalls != wantTables {
		t.Fatalf("copy calls = %d, want %d", copyCalls, wantTables)
	}
}

func TestMedicineFinalSourceChangedFailsBeforeBegin(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	medicinesPath := filepath.Join(bundle.SourceDir, "medicines.csv")
	contents, err := os.ReadFile(medicinesPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(medicinesPath, append(contents, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	session := &medicineNameProbeSession{}
	_, err = applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || (!strings.Contains(err.Error(), "sha256") && !strings.Contains(err.Error(), "row count")) {
		t.Fatalf("applyCutoverCommittingEachTable() error = %v, want changed-source fail-closed", err)
	}
	if len(session.txs) != 0 {
		t.Fatalf("transactions = %d, want 0 before denial", len(session.txs))
	}
}

// Codex 4266290426 follow-up: the pre-Begin re-read must surface csv parse
// and io failures instead of silently stopping at the first bad record. The
// manifest row count and digest intentionally match the malformed bytes so
// the metadata checks cannot mask the parse failure; the apply must be
// denied before any Begin/COPY.
func TestMedicineFinalMalformedTrailerRejectsBeforeBegin(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	medicinesPath := filepath.Join(bundle.SourceDir, "medicines.csv")
	contents, err := os.ReadFile(medicinesPath)
	if err != nil {
		t.Fatal(err)
	}
	modified := append(append([]byte(nil), contents...), []byte("\n1000009,unterminated-name")...)
	if err := os.WriteFile(medicinesPath, modified, 0o600); err != nil {
		t.Fatal(err)
	}
	for i := range bundle.Manifest.Tables {
		if bundle.Manifest.Tables[i].Table == "medicines" {
			bundle.Manifest.Tables[i].SHA256 = sha256Hex(modified)
		}
	}
	session := &medicineNameProbeSession{}
	_, err = applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil {
		t.Fatal("applyCutoverCommittingEachTable accepted a malformed medicines csv trailer")
	}
	if len(session.txs) != 0 {
		t.Fatalf("transactions = %d, want 0 before denial", len(session.txs))
	}
	if session.probes != 0 {
		t.Fatalf("name probes = %d, want 0 before denial", session.probes)
	}
}

// Existing validateCutoverCSV accepts either hex case via strings.EqualFold;
// the pre-Begin re-read must stay consistent and accept an uppercase digest
// without failing the whole candidate apply.
func TestMedicineFinalUppercaseDigestAccepted(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	for i := range bundle.Manifest.Tables {
		if bundle.Manifest.Tables[i].Table == "medicines" {
			bundle.Manifest.Tables[i].SHA256 = strings.ToUpper(bundle.Manifest.Tables[i].SHA256)
		}
	}
	session := &candidateFreshApplySession{}
	result, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err != nil {
		t.Fatal(err)
	}
	want := len(CutoverTableSpecs(CutoverProfileMedicinesCandidate))
	copyCalls := 0
	for _, tx := range session.txs {
		if !tx.committed {
			t.Fatal("transaction was not committed")
		}
		copyCalls += tx.copyCalls
	}
	if copyCalls != want {
		t.Fatalf("copy calls = %d, want %d", copyCalls, want)
	}
	if result.Counts["medicines"] != 1 {
		t.Fatalf("medicines count = %d, want 1", result.Counts["medicines"])
	}
}

func TestMedicineFinalLegacyNeverProbesNames(t *testing.T) {
	bundle := validCutoverBundleForApply(t)
	session := &medicineNameProbeSession{}
	if _, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead); err != nil {
		t.Fatal(err)
	}
	if session.probes != 0 {
		t.Fatalf("legacy apply issued %d medicines name probes, want 0", session.probes)
	}
}

func TestMedicineFinalNameHeaderPermutationAcceptedInRehearsal(t *testing.T) {
	permuted := []string{"name", "id", "clinic_id", "price", "is_active", "description", "parent_id", "dosage_form", "medicine_unit", "inventory_id", "default_quantity", "tax_type", "tax_rate", "sort_order", "is_non_insurance"}
	dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
		mutateRehearsalOnlyHandoff(f)
		if f.headerOrder == nil {
			f.headerOrder = map[string][]string{}
		}
		f.headerOrder["medicines"] = permuted
	}, CutoverProfileMedicinesCandidate)
	bundle, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
	if err != nil {
		t.Fatal(err)
	}
	session := &medicineNameProbeSession{}
	if _, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead); err != nil {
		t.Fatal(err)
	}
	if len(session.probeNames) != 1 || session.probeNames[0] != "candidate-medicine" {
		t.Fatalf("probe names = %v, want header-permutation-resolved names", session.probeNames)
	}
}
