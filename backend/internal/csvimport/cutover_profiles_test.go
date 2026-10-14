package csvimport

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestCutoverProfileLegacyDefault(t *testing.T) {
	if got := len(CutoverTableSpecs()); got != 21 {
		t.Fatalf("default tables=%d, want preserved legacy21", got)
	}
	if _, found := CutoverPlaceholderColumns()["medicines.clinic_id"]; found {
		t.Fatal("legacy placeholder inventory contains candidate medicines")
	}
}

func TestCutoverProfileLegacyFormalSource(t *testing.T) {
	dir, _ := writeCutoverFixture(t, func(f *fixtureBundle) {
		setReferenceFixtureCell(f, "estimate_items", "medicine_id", "")
	})
	path := filepath.Join(dir, cutoverManifestName)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := decodeCutoverManifest(data)
	if err != nil {
		t.Fatal(err)
	}
	tables := make([]CutoverManifestTable, 0, 21)
	for _, table := range manifest.Tables {
		if table.Table != "medicines" {
			tables = append(tables, table)
		}
	}
	manifest.Tables = tables
	delete(manifest.PlaceholderColumns, "medicines.clinic_id")
	if err := os.Remove(filepath.Join(dir, "medicines.csv")); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	data, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := PreflightCutoverBundle(dir, formalExpectedSource(sha256Hex(data))); err != nil {
		t.Fatalf("preserved pinned legacy21 formal source rejected: %v", err)
	}
}

func TestCutoverProfileCandidatePublicGuardsBeforeQuery(t *testing.T) {
	dir, _ := writeCutoverFixture(t, nil, CutoverProfileMedicinesCandidate)
	data, err := os.ReadFile(filepath.Join(dir, cutoverManifestName))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := decodeCutoverManifest(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"target", "resume-target", "verify"} {
		t.Run(name, func(t *testing.T) {
			q := &cutoverReviewDeniedSession{}
			var err error
			switch name {
			case "target":
				err = PreflightCutoverTarget(context.Background(), q, manifest, validCutoverSeeds())
			case "resume-target":
				err = PreflightCutoverTargetAllowingResume(context.Background(), q, manifest, validCutoverSeeds())
			case "verify":
				err = VerifyCutoverWithProvenance(context.Background(), q, manifest, validCutoverSeeds(), CutoverProvenanceContract{Mode: CutoverProvenanceFormal})
			}
			if q.calls != 0 || err == nil || (!strings.Contains(err.Error(), "activation") && !strings.Contains(err.Error(), "profile")) {
				t.Fatalf("queries=%d err=%v, want candidate rejected before any DB query", q.calls, err)
			}
		})
	}
}

func TestCutoverProfileQuantityRequiresExplicitValue(t *testing.T) {
	if err := validateCutoverMedicineQuantity("", 2); err == nil {
		t.Fatal("empty medicine quantity accepted despite non-null application contract")
	}
	for _, value := range []string{"0", "1.0", "0.25", "99999999.99"} {
		if err := validateCutoverMedicineQuantity(value, 2); err != nil {
			t.Fatalf("explicit quantity rejected: %v", err)
		}
	}
}

func TestCutoverProfileCandidateInventoryFKRequired(t *testing.T) {
	for _, fk := range cutoverRequiredForeignKeys(CutoverProfileMedicinesCandidate) {
		if fk.childTable == "medicines" && fk.childColumn == "inventory_id" && fk.parentTable == "inventory_items" && fk.parentColumn == "id" {
			return
		}
	}
	t.Fatal("candidate medicine inventory target FK is not required")
}

func writeMedicinesCandidateFixture(t *testing.T, mutate func(*fixtureBundle)) (string, string) {
	t.Helper()
	return writeCutoverFixture(t, func(f *fixtureBundle) {
		mutateRehearsalOnlyHandoff(f)
		if mutate != nil {
			mutate(f)
		}
	}, CutoverProfileMedicinesCandidate)
}
func candidateExpectedSource(digest string) ExpectedCutoverSource {
	expected := localRehearsalExpectedSource(digest)
	expected.Provenance.Profile = CutoverProfileMedicinesCandidate
	return expected
}

func TestCutoverProfileSelectionFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		name              string
		fixture, selected CutoverProfile
		trusted           bool
		want              string
	}{
		{"unknown", CutoverProfileLegacy21, "unknown", false, "unknown cutover profile"},
		{"legacy cannot consume candidate", CutoverProfileMedicinesCandidate, CutoverProfileLegacy21, false, "profile"},
		{"candidate cannot consume legacy", CutoverProfileLegacy21, CutoverProfileMedicinesCandidate, false, "profile"},
		{"trusted candidate cannot use rehearsal", CutoverProfileMedicinesCandidate, CutoverProfileMedicinesCandidate, true, "activation"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				if !tc.trusted {
					mutateRehearsalOnlyHandoff(f)
				}
			}, tc.fixture)
			expected := localRehearsalExpectedSource(digest)
			expected.Provenance.Profile = tc.selected
			if _, err := PreflightCutoverBundle(dir, expected); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err=%v, want %s", err, tc.want)
			}
		})
	}
}

func TestCutoverProfileExplicitCandidateFormalGuards(t *testing.T) {
	dir, digest := writeMedicinesCandidateFixture(t, nil)
	bundle, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
	if err != nil {
		t.Fatal(err)
	}
	formal := CutoverProvenanceContract{Mode: CutoverProvenanceFormal, Profile: CutoverProfileMedicinesCandidate}
	for _, name := range []string{"preflight", "resume", "verify", "apply", "apply-resume"} {
		t.Run(name, func(t *testing.T) {
			q := &cutoverReviewDeniedSession{}
			switch name {
			case "preflight":
				err = PreflightCutoverTargetWithProvenance(context.Background(), q, bundle.Manifest, validCutoverSeeds(), formal)
			case "resume":
				err = PreflightCutoverTargetAllowingResumeWithProvenance(context.Background(), q, bundle.Manifest, validCutoverSeeds(), formal)
			case "verify":
				err = VerifyCutoverWithProvenance(context.Background(), q, bundle.Manifest, validCutoverSeeds(), formal)
			case "apply":
				b := bundle
				b.Provenance = formal
				_, err = ApplyCutover(context.Background(), nil, b, validCutoverSeeds())
			case "apply-resume":
				b := bundle
				b.Provenance = formal
				_, err = ApplyCutoverCommittingEachTable(context.Background(), nil, b, validCutoverSeeds(), pgx.RepeatableRead)
			}
			if q.calls != 0 || err == nil || !strings.Contains(err.Error(), "activation") {
				t.Fatalf("calls=%d err=%v", q.calls, err)
			}
		})
	}
}

type profileSchemaTarget struct {
	dataType         string
	precision, scale int64
	missingInventory bool
	medicineCalls    int
}

func (q *profileSchemaTarget) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "pg_constraint") {
		valid := !(q.missingInventory && args[0] == "medicines" && args[1] == "inventory_id" && strings.Contains(query, "convalidated"))
		return staticRow{values: []any{valid}}
	}
	if len(args) == 2 && args[0] == "medicines" {
		q.medicineCalls++
		if args[1] == "default_quantity" {
			if strings.Contains(query, "numeric_precision") {
				return staticRow{values: []any{q.precision, q.scale}}
			}
			return staticRow{values: []any{q.dataType}}
		}
		if args[1] == "tax_rate" {
			// Canonical 001 schema mirrors the fixture: numeric with the
			// same precision/scale as the quantity column.
			if strings.Contains(query, "numeric_precision") {
				return staticRow{values: []any{q.precision, q.scale}}
			}
			return staticRow{values: []any{q.dataType}}
		}
	}
	return staticRow{values: []any{"bigint"}}
}

func TestCutoverProfileCandidateTargetPrecision(t *testing.T) {
	for _, tc := range []struct {
		name, typ        string
		precision, scale int64
		valid            bool
	}{
		{"exact", "numeric", 10, 2, true}, {"rounding", "numeric", 10, 1, false},
		{"overflow", "numeric", 9, 2, false}, {"wrong type", "double precision", 10, 2, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := &profileSchemaTarget{dataType: tc.typ, precision: tc.precision, scale: tc.scale}
			err := validateCutoverColumnTypes(context.Background(), q, CutoverProfileMedicinesCandidate)
			if (err == nil) != tc.valid {
				t.Fatalf("err=%v valid=%v", err, tc.valid)
			}
		})
	}
	legacy := &profileSchemaTarget{}
	if err := validateCutoverColumnTypes(context.Background(), legacy); err != nil || legacy.medicineCalls != 0 {
		t.Fatalf("legacy depends on medicine schema: %v calls=%d", err, legacy.medicineCalls)
	}
}

func TestCutoverProfileCandidateInventoryConstraintMissingOrUnvalidated(t *testing.T) {
	q := &profileSchemaTarget{missingInventory: true}
	if err := validateCutoverForeignKeys(context.Background(), q, CutoverProfileMedicinesCandidate); err == nil || !strings.Contains(err.Error(), "medicines.inventory_id") {
		t.Fatalf("missing/unvalidated inventory FK accepted: %v", err)
	}
	if err := validateCutoverForeignKeys(context.Background(), q); err != nil {
		t.Fatalf("legacy requires medicine-owned FK: %v", err)
	}
}

func TestCutoverProfileBoundCandidateCannotDowngradeToLegacy(t *testing.T) {
	dir, digest := writeMedicinesCandidateFixture(t, nil)
	bundle, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateCutoverProfile(bundle.Manifest, CutoverProvenanceContract{Mode: CutoverProvenanceLocalRehearsal}); err == nil || !strings.Contains(err.Error(), "binding mismatch") {
		t.Fatalf("candidate downgrade err=%v", err)
	}
	coverage := CutoverMappingCoverage(CutoverProfileMedicinesCandidate)
	if len(coverage) != 22 {
		t.Fatalf("candidate mapping tables=%d", len(coverage))
	}
	for _, entry := range coverage {
		if entry.ConsumerStatus != "candidate_22_rehearsal_only" {
			t.Fatalf("candidate claims formal status: %s", entry.ConsumerStatus)
		}
	}
}
