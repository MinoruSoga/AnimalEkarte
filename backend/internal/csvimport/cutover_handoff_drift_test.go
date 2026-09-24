package csvimport

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Regression coverage for MIG-19 / BUG-LOCAL-HANDOFF-CSV-CONTRACT: rehearsal-grade
// handoff bundles tolerate bounded PK/FK, column-order, and schema-contract drift
// while formal/trusted artifacts stay fail-closed. Acceptance mapping:
//
//	A1  sequence-less target tables + explicit-id validation by header name
//	A2  manifest tables[] order normalized to parent-before-child spec order
//	A3  dangling-FK diagnostics (table/column/row/parent, no cell values)
//	A4  header-name column mapping (permutation policy, anomaly rejection,
//	    COPY payload re-emitted in spec order)
//	A5  csvContractSha256 / predicate / placeholder inventory drift notes

func localRehearsalExpectedSource(digest string) ExpectedCutoverSource {
	return ExpectedCutoverSource{
		ManifestSHA256: digest,
		ClinicCode:     "hachioji",
		ClinicOrdinal:  1,
		RunID:          "run-1",
		Provenance:     CutoverProvenanceContract{Mode: CutoverProvenanceLocalRehearsal},
	}
}

func formalExpectedSource(digest string) ExpectedCutoverSource {
	return ExpectedCutoverSource{
		ManifestSHA256: digest,
		ClinicCode:     "hachioji",
		ClinicOrdinal:  1,
		RunID:          "run-1",
	}
}

func rehearsalOnlyManifestForTargetTests() CutoverManifest {
	manifest := cutoverManifestForTargetTests()
	manifest.Status = "REHEARSAL_ONLY"
	manifest.HandoffEligibility = "REHEARSAL_ONLY"
	return manifest
}

func cutoverSpecByName(t *testing.T, name string) CutoverTableSpec {
	t.Helper()
	for _, spec := range CutoverTableSpecs() {
		if spec.Name == name {
			return spec
		}
	}
	t.Fatalf("no cutover spec for table %s", name)
	return CutoverTableSpec{}
}

func driftTestIDBand() CutoverIDBand {
	return CutoverIDBand{
		Base:               0,
		EndExclusive:       10_000_000,
		NonOwnerIDOffset:   1_000_000,
		OwnerFloor:         300_000,
		ApplicationIDFloor: applicationIDFloor,
	}
}

// writeDriftTestCSV writes one owner-only CSV and returns its path plus the
// manifest entry bound to the exact written bytes.
func writeDriftTestCSV(t *testing.T, spec CutoverTableSpec, header []string, rows [][]string) (string, CutoverManifestTable) {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, spec.Name+".csv")
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(header); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			t.Fatal(err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(buf.Bytes())
	return path, CutoverManifestTable{
		Table:    spec.Name,
		File:     spec.Name + ".csv",
		RowCount: int64(len(rows)),
		SHA256:   hex.EncodeToString(sum[:]),
	}
}

// permuteRow reorders a spec-order row to match a permuted header so fixtures
// keep every value bound to its column name.
func permuteRow(t *testing.T, specColumns, header, specRow []string) []string {
	t.Helper()
	row := make([]string, len(header))
	for i, name := range header {
		idx := columnIndex(specColumns, name)
		if idx < 0 {
			t.Fatalf("header names non-contract column %q", name)
		}
		row[i] = specRow[idx]
	}
	return row
}

func reversedColumns(columns []string) []string {
	reversed := append([]string(nil), columns...)
	slices.Reverse(reversed)
	return reversed
}

// validPetsRow is a spec-order row that satisfies every scalar gate: band
// ranges, clinic/species placeholders, and non-integer-free band columns.
func validPetsRow() []string {
	var spec CutoverTableSpec
	for _, candidate := range CutoverTableSpecs() {
		if candidate.Name == "pets" {
			spec = candidate
			break
		}
	}
	row := make([]string, len(spec.Columns))
	for i, column := range spec.Columns {
		switch column {
		case "id":
			row[i] = "1000001"
		case "clinic_id":
			row[i] = "{{CLINIC_ID}}"
		case "owner_id":
			row[i] = "300001"
		case "animal_species_id":
			row[i] = "{{FALLBACK_ANIMAL_SPECIES_ID}}"
		}
	}
	return row
}

// A1: rehearsal-grade targets may have no serial id sequence; explicit CSV ids
// are authoritative and a stable drift note is recorded per table. Formal
// targets keep the strict requirement.
func TestCutoverSequencesToleratedOnlyForRehearsalOnlyArtifact(t *testing.T) {
	ctx := context.Background()
	specs := CutoverTableSpecs()

	notes, err := validateCutoverSequences(ctx, &fakeCutoverTransaction{missingSequence: true}, true)
	if err != nil {
		t.Fatalf("validateCutoverSequences(tolerant) error = %v", err)
	}
	if len(notes) != len(specs) {
		t.Fatalf("drift notes = %d, want %d (one per sequence-less table)", len(notes), len(specs))
	}
	for i, spec := range specs {
		want := "target id sequence missing: " + spec.Name
		if notes[i] != want {
			t.Fatalf("notes[%d] = %q, want %q", i, notes[i], want)
		}
	}

	if _, err := validateCutoverSequences(ctx, &fakeCutoverTransaction{missingSequence: true}, false); err == nil ||
		!strings.Contains(err.Error(), "serial id sequence") {
		t.Fatalf("validateCutoverSequences(strict) error = %v, want serial sequence rejection", err)
	}
}

func TestPreflightCutoverTargetSequenceToleranceIsArtifactKeyed(t *testing.T) {
	ctx := context.Background()

	if err := PreflightCutoverTarget(ctx, &fakeCutoverTransaction{missingSequence: true}, rehearsalOnlyManifestForTargetTests(), validCutoverSeeds()); err != nil {
		t.Fatalf("PreflightCutoverTarget(REHEARSAL_ONLY) error = %v", err)
	}

	trusted := cutoverManifestForTargetTests()
	trusted.Status = "PASS"
	trusted.HandoffEligibility = "TRUSTED_CANDIDATE"
	if err := PreflightCutoverTarget(ctx, &fakeCutoverTransaction{missingSequence: true}, trusted, validCutoverSeeds()); err == nil ||
		!strings.Contains(err.Error(), "serial id sequence") {
		t.Fatalf("PreflightCutoverTarget(TRUSTED_CANDIDATE) error = %v, want serial sequence rejection", err)
	}
}

func TestAdvanceAndVerifyCutoverSequencesSkipMissingForRehearsalOnly(t *testing.T) {
	ctx := context.Background()
	specs := CutoverTableSpecs()

	advanceTx := &fakeCutoverTransaction{missingSequence: true}
	notes, err := advanceCutoverSequences(ctx, advanceTx, true)
	if err != nil {
		t.Fatalf("advanceCutoverSequences(tolerant) error = %v", err)
	}
	if len(notes) != len(specs) {
		t.Fatalf("advance drift notes = %d, want %d", len(notes), len(specs))
	}
	if advanceTx.setvalCalls != 0 {
		t.Fatalf("setval calls = %d, want 0 when sequences are absent", advanceTx.setvalCalls)
	}

	verifyNotes, err := verifyCutoverSequences(ctx, &fakeCutoverTransaction{missingSequence: true}, true)
	if err != nil {
		t.Fatalf("verifyCutoverSequences(tolerant) error = %v", err)
	}
	if len(verifyNotes) != len(specs) {
		t.Fatalf("verify drift notes = %d, want %d", len(verifyNotes), len(specs))
	}

	if _, err := advanceCutoverSequences(ctx, &fakeCutoverTransaction{missingSequence: true}, false); err == nil ||
		!strings.Contains(err.Error(), "resolve sequence") {
		t.Fatalf("advanceCutoverSequences(strict) error = %v, want sequence rejection", err)
	}
	if _, err := verifyCutoverSequences(ctx, &fakeCutoverTransaction{missingSequence: true}, false); err == nil ||
		!strings.Contains(err.Error(), "resolve sequence") {
		t.Fatalf("verifyCutoverSequences(strict) error = %v, want sequence rejection", err)
	}
}

// A1 (id validation): the primary id column is located by header name even
// when a rehearsal producer emits it away from position zero. Duplicate,
// non-integer, and out-of-band ids each fail closed.
func TestCutoverCSVPrimaryIDValidatedByHeaderName(t *testing.T) {
	spec := cutoverSpecByName(t, "pets")
	header := reversedColumns(spec.Columns) // id lands at the last position, not first
	specRow := validPetsRow()
	secondRow := append([]string(nil), specRow...)
	secondRow[columnIndex(spec.Columns, "id")] = "1000002"
	band := driftTestIDBand()

	tests := []struct {
		name    string
		rows    [][]string
		wantErr string
	}{
		{
			name:    "distinct ids pass under permutation",
			rows:    [][]string{permuteRow(t, spec.Columns, header, specRow), permuteRow(t, spec.Columns, header, secondRow)},
			wantErr: "",
		},
		{
			name: "duplicate id rejected",
			rows: [][]string{
				permuteRow(t, spec.Columns, header, specRow),
				permuteRow(t, spec.Columns, header, specRow),
			},
			wantErr: "column id row 3: duplicate primary ID",
		},
		{
			name: "non-integer id rejected",
			rows: [][]string{func() []string {
				row := permuteRow(t, spec.Columns, header, specRow)
				row[columnIndex(header, "id")] = "not-an-integer"
				return row
			}()},
			wantErr: "column id row 2: primary ID must be an integer",
		},
		{
			name: "id below band rejected",
			rows: [][]string{func() []string {
				row := permuteRow(t, spec.Columns, header, specRow)
				row[columnIndex(header, "id")] = "42"
				return row
			}()},
			wantErr: "column id row 2: id is outside the clinic band",
		},
		{
			name: "id above band rejected",
			rows: [][]string{func() []string {
				row := permuteRow(t, spec.Columns, header, specRow)
				row[columnIndex(header, "id")] = "250000000"
				return row
			}()},
			wantErr: "column id row 2: id is outside the clinic band",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, table := writeDriftTestCSV(t, spec, header, tt.rows)
			_, err := validateCutoverCSV(path, spec, table, band, true)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateCutoverCSV error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateCutoverCSV error = %v, want text %q", err, tt.wantErr)
			}
			// The error must name the column and row, never echo a cell value.
			for _, row := range tt.rows {
				for _, value := range row {
					if value != "" && strings.Contains(err.Error(), value) && !strings.Contains(tt.wantErr, value) {
						t.Fatalf("error %q leaks cell value %q", err.Error(), value)
					}
				}
			}
		})
	}
}

// A2: a rehearsal producer may emit manifest tables[] in any order; preflight
// normalizes to the immutable parent-before-child spec order and records the
// drift. Formal bundles keep exact-order enforcement, and unknown, duplicate,
// or non-set-equal tables always fail closed.
func TestPreflightRehearsalNormalizesShuffledManifestTables(t *testing.T) {
	dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
		mutateRehearsalOnlyHandoff(f)
		slices.Reverse(f.manifest.Tables)
	})

	bundle, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
	if err != nil {
		t.Fatalf("PreflightCutoverBundle(shuffled rehearsal manifest) error = %v", err)
	}
	for i, spec := range CutoverTableSpecs() {
		if bundle.Manifest.Tables[i].Table != spec.Name {
			t.Fatalf("manifest tables[%d] = %s, want spec order %s (parents before children)", i, bundle.Manifest.Tables[i].Table, spec.Name)
		}
	}
	if !slices.Contains(bundle.ToleratedDrift, "manifest.tables order") {
		t.Fatalf("ToleratedDrift = %v, want manifest.tables order note", bundle.ToleratedDrift)
	}
}

func TestPreflightTableOrderIsArtifactKeyed(t *testing.T) {
	// REHEARSAL_ONLY under staging rehearsal: same tolerance as local.
	dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
		mutateRehearsalOnlyHandoff(f)
		slices.Reverse(f.manifest.Tables)
	})
	bundle, err := PreflightCutoverBundle(dir, stagingExpectedCutoverSource(digest))
	if err != nil {
		t.Fatalf("PreflightCutoverBundle(staging rehearsal-only shuffled) error = %v", err)
	}
	if !slices.Contains(bundle.ToleratedDrift, "manifest.tables order") {
		t.Fatalf("ToleratedDrift = %v, want manifest.tables order note", bundle.ToleratedDrift)
	}

	// Verified PASS artifacts stay strict under every mode.
	for name, expected := range map[string]func(string) ExpectedCutoverSource{
		"formal":            formalExpectedSource,
		"staging rehearsal": stagingExpectedCutoverSource,
		"local rehearsal":   localRehearsalExpectedSource,
	} {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			slices.Reverse(f.manifest.Tables)
		})
		if _, err := PreflightCutoverBundle(dir, expected(digest)); err == nil ||
			!strings.Contains(err.Error(), "table order mismatch") {
			t.Fatalf("%s: PreflightCutoverBundle error = %v, want table order mismatch", name, err)
		}
	}
}

func TestPreflightRehearsalRejectsManifestTableAnomalies(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*fixtureBundle)
		wantErr string
	}{
		{
			name: "duplicate table entry",
			mutate: func(f *fixtureBundle) {
				f.manifest.Tables[0].Table = "pets"
			},
			wantErr: "more than once",
		},
		{
			name: "unknown table entry",
			mutate: func(f *fixtureBundle) {
				f.manifest.Tables[0].Table = "bogus_table"
			},
			wantErr: "unknown table",
		},
		{
			name: "missing table entry",
			mutate: func(f *fixtureBundle) {
				f.manifest.Tables = f.manifest.Tables[:len(f.manifest.Tables)-1]
			},
			wantErr: "want 21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				mutateRehearsalOnlyHandoff(f)
				tt.mutate(f)
			})
			_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("PreflightCutoverBundle error = %v, want text %q", err, tt.wantErr)
			}
		})
	}
}

// A3: dangling foreign keys fail with table, column, CSV row, and parent table
// — never the source cell value. Self-parent misses name the declaring row.
func TestPreflightDanglingForeignKeyDiagnostics(t *testing.T) {
	petsSpec := cutoverSpecByName(t, "pets")
	proceduresSpec := cutoverSpecByName(t, "procedures")

	t.Run("pet references a missing owner", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			f.rows["pets"][0][columnIndex(petsSpec.Columns, "owner_id")] = "300999"
		})
		_, err := PreflightCutoverBundle(dir, formalExpectedSource(digest))
		if err == nil {
			t.Fatal("PreflightCutoverBundle succeeded, want dangling owner rejection")
		}
		for _, want := range []string{"table pets", "column owner_id", "row 2", "owners"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error = %q, want detail %q", err.Error(), want)
			}
		}
		if strings.Contains(err.Error(), "300999") {
			t.Fatalf("error leaks the source cell value: %q", err.Error())
		}
	})

	t.Run("procedure self-parent miss names the declaring row", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			f.rows["procedures"][0][columnIndex(proceduresSpec.Columns, "parent_id")] = "1000002"
		})
		_, err := PreflightCutoverBundle(dir, formalExpectedSource(digest))
		if err == nil {
			t.Fatal("PreflightCutoverBundle succeeded, want self-parent rejection")
		}
		for _, want := range []string{"table procedures", "column parent_id", "row 2", "procedures"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error = %q, want detail %q", err.Error(), want)
			}
		}
		if strings.Contains(err.Error(), "1000002") {
			t.Fatalf("error leaks the source cell value: %q", err.Error())
		}
	})
}

// A4: CSV columns map to contract columns by header name. Rehearsal bundles
// accept permutations; duplicate, unknown, or missing columns fail closed;
// formal bundles require the exact order.
func TestCutoverCSVHeaderPermutationPolicy(t *testing.T) {
	spec := cutoverSpecByName(t, "pets")
	specRow := validPetsRow()
	band := driftTestIDBand()

	permutedHeader := reversedColumns(spec.Columns)
	permutedRow := permuteRow(t, spec.Columns, permutedHeader, specRow)

	duplicateHeader := append([]string(nil), spec.Columns...)
	duplicateHeader[3] = "id"
	unknownHeader := append(append([]string(nil), spec.Columns...), "bogus_column")
	missingHeader := append([]string(nil), spec.Columns[:len(spec.Columns)-1]...)

	tests := []struct {
		name             string
		header           []string
		rows             [][]string
		allowPermutation bool
		wantErr          string
		wantPermuted     bool
	}{
		{
			name:             "rehearsal accepts a full permutation",
			header:           permutedHeader,
			rows:             [][]string{permutedRow},
			allowPermutation: true,
			wantPermuted:     true,
		},
		{
			name:             "formal rejects a full permutation",
			header:           permutedHeader,
			rows:             [][]string{permutedRow},
			allowPermutation: false,
			wantErr:          "does not match exact column order",
		},
		{
			name:             "duplicate header rejected for rehearsal",
			header:           duplicateHeader,
			rows:             [][]string{specRow},
			allowPermutation: true,
			wantErr:          "duplicate column id",
		},
		{
			name:             "duplicate header rejected for formal",
			header:           duplicateHeader,
			rows:             [][]string{specRow},
			allowPermutation: false,
			wantErr:          "does not match exact column order",
		},
		{
			name:   "unknown header rejected for rehearsal",
			header: unknownHeader,
			rows: [][]string{
				append(append([]string(nil), specRow...), "x"),
			},
			allowPermutation: true,
			wantErr:          "unknown column bogus_column",
		},
		{
			name:   "missing header rejected for rehearsal",
			header: missingHeader,
			rows: [][]string{
				specRow[:len(specRow)-1],
			},
			allowPermutation: true,
			wantErr:          "missing column",
		},
		{
			name:             "exact order passes under both policies",
			header:           spec.Columns,
			rows:             [][]string{specRow},
			allowPermutation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, table := writeDriftTestCSV(t, spec, tt.header, tt.rows)
			permuted, err := validateCutoverCSV(path, spec, table, band, tt.allowPermutation)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateCutoverCSV error = %v", err)
				}
				if permuted != tt.wantPermuted {
					t.Fatalf("permuted = %v, want %v", permuted, tt.wantPermuted)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateCutoverCSV error = %v, want text %q", err, tt.wantErr)
			}
		})
	}
}

// A4 (COPY transform): a permuted rehearsal CSV is re-emitted in spec column
// order with placeholders resolved, so the COPY stream matches the target
// table definition.
func TestTransformCutoverCSVReEmitsSpecColumnOrder(t *testing.T) {
	ctx := context.Background()
	spec := cutoverSpecByName(t, "pets")
	specRow := validPetsRow()
	header := reversedColumns(spec.Columns)
	path, table := writeDriftTestCSV(t, spec, header, [][]string{permuteRow(t, spec.Columns, header, specRow)})

	var output bytes.Buffer
	count, err := transformCutoverCSV(ctx, path, spec, true, &output, validCutoverSeeds(), table.SHA256)
	if err != nil {
		t.Fatalf("transformCutoverCSV error = %v", err)
	}
	if count != 1 {
		t.Fatalf("transformed row count = %d, want 1", count)
	}
	records, err := csv.NewReader(&output).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("transformed CSV has %d records, want header + 1 row", len(records))
	}
	if !slices.Equal(records[0], spec.Columns) {
		t.Fatalf("transformed header = %v, want spec column order %v", records[0], spec.Columns)
	}
	want := make([]string, len(spec.Columns))
	for i, column := range spec.Columns {
		switch column {
		case "id":
			want[i] = "1000001"
		case "clinic_id":
			want[i] = "1" // seeds.ClinicID
		case "owner_id":
			want[i] = "300001"
		case "animal_species_id":
			want[i] = "2" // seeds.AnimalSpeciesID
		}
	}
	if !slices.Equal(records[1], want) {
		t.Fatalf("transformed row = %v, want %v", records[1], want)
	}
}

// A5: rehearsal-grade bundles tolerate csvContractSha256 and
// predicate/placeholder-inventory drift with recorded notes; formal and
// trusted artifacts reject the same drift.
func TestPreflightContractDriftIsArtifactKeyed(t *testing.T) {
	driftedManifest := func(f *fixtureBundle) {
		f.manifest.CSVContractSHA256 = strings.Repeat("9", 64)
		f.manifest.ImportablePredicate = "mapping_status IN ('confirmed')"
		f.manifest.PlaceholderColumns = map[string]string{"staffs.clinic_id": "{{CLINIC_ID}}"}
	}

	t.Run("rehearsal-only local handoff tolerates drift with notes", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			mutateRehearsalOnlyHandoff(f)
			driftedManifest(f)
		})
		bundle, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
		if err != nil {
			t.Fatalf("PreflightCutoverBundle error = %v", err)
		}
		for _, want := range []string{"manifest.csvContractSha256", "manifest.importablePredicate", "manifest.placeholderColumns"} {
			if !slices.Contains(bundle.ToleratedDrift, want) {
				t.Fatalf("ToleratedDrift = %v, want note %q", bundle.ToleratedDrift, want)
			}
		}
	})

	t.Run("rehearsal-only staging handoff tolerates drift", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			mutateRehearsalOnlyHandoff(f)
			driftedManifest(f)
		})
		if _, err := PreflightCutoverBundle(dir, stagingExpectedCutoverSource(digest)); err != nil {
			t.Fatalf("PreflightCutoverBundle(staging rehearsal) error = %v", err)
		}
	})

	t.Run("formal mode rejects digest drift", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			driftedManifest(f)
		})
		_, err := PreflightCutoverBundle(dir, formalExpectedSource(digest))
		if err == nil || !strings.Contains(err.Error(), "mapping contract binding is invalid") {
			t.Fatalf("PreflightCutoverBundle error = %v, want mapping contract rejection", err)
		}
	})

	// This is the bug.md failure: a TRUSTED_CANDIDATE artifact must never be
	// downgraded by rehearsal-mode flags — digest drift stays fail-closed.
	t.Run("trusted artifact under local rehearsal rejects digest drift", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			f.manifest.CSVContractSHA256 = strings.Repeat("9", 64)
		})
		_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
		if err == nil || !strings.Contains(err.Error(), "CSV contract digest is invalid") {
			t.Fatalf("PreflightCutoverBundle error = %v, want CSV contract digest rejection", err)
		}
	})

	t.Run("verified staging PASS bundle rejects digest drift", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			f.manifest.CSVContractSHA256 = strings.Repeat("9", 64)
		})
		_, err := PreflightCutoverBundle(dir, stagingExpectedCutoverSource(digest))
		if err == nil || !strings.Contains(err.Error(), "mapping contract binding is invalid") {
			t.Fatalf("PreflightCutoverBundle error = %v, want mapping contract rejection", err)
		}
	})
}

// A5 (structural gates stay authoritative): tolerated contract drift does not
// relax per-file sha256, row counts, or the reference graph.
func TestPreflightRehearsalDriftKeepsStructuralGates(t *testing.T) {
	t.Run("row count mismatch still fails", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			mutateRehearsalOnlyHandoff(f)
			for i := range f.manifest.Tables {
				if f.manifest.Tables[i].Table == "pets" {
					f.manifest.Tables[i].RowCount = 99
				}
			}
		})
		_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
		if err == nil || !strings.Contains(err.Error(), "row count mismatch") {
			t.Fatalf("PreflightCutoverBundle error = %v, want row count rejection", err)
		}
	})

	t.Run("dangling FK still fails under rehearsal drift", func(t *testing.T) {
		petsSpec := cutoverSpecByName(t, "pets")
		dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
			mutateRehearsalOnlyHandoff(f)
			f.manifest.CSVContractSHA256 = strings.Repeat("9", 64)
			f.rows["pets"][0][columnIndex(petsSpec.Columns, "owner_id")] = "300999"
		})
		_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
		if err == nil || !strings.Contains(err.Error(), "owner_id") {
			t.Fatalf("PreflightCutoverBundle error = %v, want dangling owner rejection", err)
		}
		if strings.Contains(err.Error(), "300999") {
			t.Fatalf("error leaks the source cell value: %q", err.Error())
		}
	})
}
