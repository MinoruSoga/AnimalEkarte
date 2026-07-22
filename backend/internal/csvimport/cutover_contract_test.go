package csvimport

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestPreflightCutoverBundleAcceptsExactContract(t *testing.T) {
	dir, manifestDigest := writeCutoverFixture(t, nil)

	bundle, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
		ManifestSHA256: manifestDigest,
		ClinicCode:     "hachioji",
		ClinicOrdinal:  1,
		RunID:          "run-1",
	})
	if err != nil {
		t.Fatalf("PreflightCutoverBundle() error = %v", err)
	}
	if len(bundle.Manifest.Tables) != 19 {
		t.Fatalf("table count = %d, want 19", len(bundle.Manifest.Tables))
	}
	if bundle.Manifest.IDBand.ApplicationIDFloor != 1_000_000_000 {
		t.Fatalf("application floor = %d", bundle.Manifest.IDBand.ApplicationIDFloor)
	}
}

func TestCutoverTableSpecsDeclareNonNullableTextColumns(t *testing.T) {
	expected := map[string][]string{
		"staffs":                       {"name", "license_number"},
		"procedures":                   {"name", "description"},
		"merchandise_items":            {"name"},
		"owners":                       {"name", "company", "postal_code", "address1", "address2", "home_postal_code", "home_address1", "home_address2", "phone", "company_phone", "email", "remarks"},
		"pets":                         {"pet_number", "name", "breed", "color", "food", "remarks"},
		"medical_records":              {"record_no"},
		"inquiries":                    {"chief_complaint", "owner_observations", "history", "notes", "allergy_info", "current_medications"},
		"clinical_plans":               {"physical_exam", "diagnosis_details", "treatment_policy"},
		"vital_records":                {"notes"},
		"appointments":                 nil,
		"appointment_trimming_details": {"remarks"},
		"billings":                     nil,
		"billing_items":                {"name"},
		"estimates":                    {"estimate_no", "title", "comment", "notes"},
		"estimate_items":               {"name"},
		"exams":                        {"result_summary"},
		"exam_results":                 {"name", "inspection_value", "normal_value"},
		"vaccines":                     {"name", "description", "interval"},
		"vaccinations":                 {"supplemental", "lot1", "lot2", "lot3", "lot4", "remarks"},
	}

	for _, spec := range CutoverTableSpecs() {
		if !slices.Equal(spec.ForceNotNullColumns, expected[spec.Name]) {
			t.Errorf("%s force-not-null columns = %v, want %v", spec.Name, spec.ForceNotNullColumns, expected[spec.Name])
		}
	}
}

func TestPreflightCutoverBundleFailsClosed(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*fixtureBundle)
		wantErr string
	}{
		{
			name: "manifest table order drift",
			mutate: func(f *fixtureBundle) {
				f.manifest.Tables[0], f.manifest.Tables[1] = f.manifest.Tables[1], f.manifest.Tables[0]
			},
			wantErr: "table order",
		},
		{
			name: "CSV digest mismatch",
			mutate: func(f *fixtureBundle) {
				f.afterManifest = func(dir string) {
					file := filepath.Join(dir, "owners.csv")
					contents, _ := os.ReadFile(file)
					_ = os.WriteFile(file, append(contents, '\n'), 0o600)
				}
			},
			wantErr: "sha256",
		},
		{
			name: "row count mismatch",
			mutate: func(f *fixtureBundle) {
				f.manifest.Tables[0].RowCount = 2
			},
			wantErr: "row count",
		},
		{
			name: "unsafe manifest filename",
			mutate: func(f *fixtureBundle) {
				f.manifest.Tables[0].File = "../staffs.csv"
			},
			wantErr: "filename",
		},
		{
			name: "unknown placeholder",
			mutate: func(f *fixtureBundle) {
				f.rows["owners"][0][2] = "{{UNKNOWN_ID}}"
			},
			wantErr: "placeholder",
		},
		{
			name: "ID outside clinic band",
			mutate: func(f *fixtureBundle) {
				f.rows["staffs"][0][0] = "10000000"
			},
			wantErr: "clinic band",
		},
		{
			name: "duplicate primary ID",
			mutate: func(f *fixtureBundle) {
				f.rows["owners"] = append(f.rows["owners"], append([]string(nil), f.rows["owners"][0]...))
				f.manifest.Tables[3].RowCount = 2
			},
			wantErr: "duplicate",
		},
		{
			name: "unexpected file",
			mutate: func(f *fixtureBundle) {
				f.afterManifest = func(dir string) {
					_ = os.WriteFile(filepath.Join(dir, "unexpected.csv"), []byte("id\n"), 0o600)
				}
			},
			wantErr: "unexpected file",
		},
		{
			name: "group readable source",
			mutate: func(f *fixtureBundle) {
				f.afterManifest = func(dir string) {
					_ = os.Chmod(filepath.Join(dir, "owners.csv"), 0o640)
				}
			},
			wantErr: "owner-only",
		},
		{
			name: "oversized CSV",
			mutate: func(f *fixtureBundle) {
				f.afterManifest = func(dir string) {
					_ = os.Truncate(filepath.Join(dir, "owners.csv"), maxCutoverCSVBytes+1)
				}
			},
			wantErr: "size limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, tt.mutate)
			_, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
				ManifestSHA256: digest,
				ClinicCode:     "hachioji",
				ClinicOrdinal:  1,
				RunID:          "run-1",
			})
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantErr)) {
				t.Fatalf("error = %v, want text %q", err, tt.wantErr)
			}
		})
	}
}

func TestReadOwnerOnlyRegularFileRejectsOversizedManifest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.json")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(maxCutoverManifestBytes + 1); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := readOwnerOnlyRegularFile(path); err == nil || !strings.Contains(err.Error(), "size limit") {
		t.Fatalf("readOwnerOnlyRegularFile() error = %v, want size-limit rejection", err)
	}
}

func TestPreflightCutoverBundleRejectsSymlinkedCSV(t *testing.T) {
	dir, digest := writeCutoverFixture(t, nil)
	owners := filepath.Join(dir, "owners.csv")
	target := filepath.Join(dir, "owners-real.csv")
	if err := os.Rename(owners, target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, owners); err != nil {
		t.Fatal(err)
	}

	_, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
		ManifestSHA256: digest,
		ClinicCode:     "hachioji",
		ClinicOrdinal:  1,
		RunID:          "run-1",
	})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "symbolic link") {
		t.Fatalf("error = %v, want symbolic link rejection", err)
	}
}

type fixtureBundle struct {
	manifest      CutoverManifest
	rows          map[string][][]string
	afterManifest func(string)
}

func writeCutoverFixture(t *testing.T, mutate func(*fixtureBundle)) (string, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	f := fixtureBundle{
		manifest: CutoverManifest{
			GeneratedAt:               "2026-07-22T00:00:00Z",
			Status:                    "PASS",
			SourceLayer:               "animalekarte_stage",
			SourceRunID:               "run-1",
			ClinicCode:                "hachioji",
			ClinicOrdinal:             1,
			ClinicBandBase:            0,
			ClinicBandEndExclusive:    10_000_000,
			StageIDOffset:             1_000_000,
			IDBand:                    CutoverIDBand{Base: 0, EndExclusive: 10_000_000, NonOwnerIDOffset: 1_000_000, OwnerFloor: 300_000, ApplicationIDFloor: 1_000_000_000},
			OutputDir:                 "sensitive-local/animalekarte-csv-export/hachioji/run-1",
			Format:                    "csv-with-header",
			ImportablePredicate:       "mapping_status IN ('confirmed','inferred'), plus per-table CSV FK-parent eligibility guards (see scripts/lib/stage-csv-columns.mjs)",
			PlaceholderColumns:        CutoverPlaceholderColumns(),
			PlaceholderResolutionNote: "source placeholders",
		},
		rows: map[string][][]string{},
	}

	for _, spec := range CutoverTableSpecs() {
		row := make([]string, len(spec.Columns))
		for i, column := range spec.Columns {
			switch column {
			case "id":
				if spec.Name == "owners" {
					row[i] = "300001"
				} else {
					row[i] = "1000001"
				}
			case "clinic_id":
				row[i] = "{{CLINIC_ID}}"
			case "animal_species_id":
				row[i] = "{{FALLBACK_ANIMAL_SPECIES_ID}}"
			case "exam_type_id":
				row[i] = "{{FALLBACK_EXAM_TYPE_ID}}"
			case "reservation_type_id":
				row[i] = "{{TRIMMING_RESERVATION_TYPE_ID}}"
			}
		}
		for _, column := range spec.BandColumns {
			idx := columnIndex(spec.Columns, column)
			if idx < 0 || row[idx] != "" || column == "clinic_id" {
				continue
			}
			if column == "owner_id" {
				row[idx] = "300001"
			} else {
				row[idx] = "1000001"
			}
		}
		f.rows[spec.Name] = [][]string{row}
		f.manifest.Tables = append(f.manifest.Tables, CutoverManifestTable{
			Table: spec.Name, File: spec.Name + ".csv", RowCount: 1,
		})
	}
	if mutate != nil {
		mutate(&f)
	}

	for i, spec := range CutoverTableSpecs() {
		fileName := spec.Name + ".csv"
		file := filepath.Join(dir, fileName)
		out, err := os.OpenFile(file, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			t.Fatal(err)
		}
		w := csv.NewWriter(out)
		if err := w.Write(spec.Columns); err != nil {
			t.Fatal(err)
		}
		for _, row := range f.rows[spec.Name] {
			if err := w.Write(row); err != nil {
				t.Fatal(err)
			}
		}
		w.Flush()
		if err := w.Error(); err != nil {
			t.Fatal(err)
		}
		if err := out.Close(); err != nil {
			t.Fatal(err)
		}
		contents, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(contents)
		f.manifest.Tables[i].SHA256 = hex.EncodeToString(sum[:])
	}

	manifestBytes, err := json.MarshalIndent(f.manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	manifestBytes = append(manifestBytes, '\n')
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), manifestBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if f.afterManifest != nil {
		f.afterManifest(dir)
	}
	sum := sha256.Sum256(manifestBytes)
	return dir, hex.EncodeToString(sum[:])
}

func columnIndex(columns []string, name string) int {
	for i, column := range columns {
		if column == name {
			return i
		}
	}
	return -1
}
