package csvimport

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestPreflightCutoverBundleRejectsMissingEstimatePet(t *testing.T) {
	dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
		for _, spec := range CutoverTableSpecs() {
			if spec.Name == "estimates" {
				f.rows[spec.Name][0][columnIndex(spec.Columns, "pet_id")] = "1000084"
			}
			if spec.Name == "estimate_items" {
				f.rows[spec.Name][0][columnIndex(spec.Columns, "consultation_id")] = ""
				f.rows[spec.Name][0][columnIndex(spec.Columns, "medicine_id")] = ""
			}
		}
	})
	_, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
		ManifestSHA256: digest,
		ClinicCode:     "hachioji",
		ClinicOrdinal:  1,
		RunID:          "run-1",
	})
	if err == nil || !strings.Contains(err.Error(), "estimates") || !strings.Contains(err.Error(), "pet_id") {
		t.Fatalf("PreflightCutoverBundle() error = %v, want missing estimates.pet_id rejection", err)
	}
	if strings.Contains(err.Error(), "1000084") || strings.Contains(err.Error(), dir) {
		t.Fatal("source reference error exposes an identifier or path")
	}
}

func TestPreflightCutoverBundleRejectsEveryMissingCSVParent(t *testing.T) {
	for table, refs := range cutoverCSVReferences() {
		for _, ref := range refs {
			if ref.kind != cutoverCSVParent {
				continue
			}
			t.Run(table+"/"+ref.column, func(t *testing.T) {
				dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
					setReferenceFixtureCell(f, table, ref.column, "1000084")
				})
				_, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
				if err == nil {
					t.Fatal("missing in-band parent was accepted")
				}
				if strings.Contains(err.Error(), "1000084") || strings.Contains(err.Error(), dir) {
					t.Fatal("reference error exposes source values or path")
				}
			})
		}
	}
}

func TestPreflightCutoverBundleRejectsEmptyRequiredReferences(t *testing.T) {
	for table, refs := range cutoverCSVReferences() {
		for _, ref := range refs {
			if ref.nullable {
				continue
			}
			t.Run(table+"/"+ref.column, func(t *testing.T) {
				dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
					setReferenceFixtureCell(f, table, ref.column, "")
				})
				if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil {
					t.Fatal("empty required reference was accepted")
				}
			})
		}
	}
}

func TestPreflightCutoverBundleAllowsNullableReferences(t *testing.T) {
	dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
		for table, refs := range cutoverCSVReferences() {
			for _, ref := range refs {
				if ref.nullable {
					setReferenceFixtureCell(f, table, ref.column, "")
				}
			}
		}
	})
	if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
		t.Fatalf("nullable references rejected: %v", err)
	}
}

func TestPreflightCutoverBundleRejectsNonImportedParents(t *testing.T) {
	for _, tc := range []struct{ table, column string }{
		{"estimate_items", "consultation_id"},
		{"estimate_items", "medicine_id"},
		{"vaccines", "inventory_id"},
	} {
		for _, value := range []string{"0", "1000001"} {
			t.Run(tc.table+"/"+tc.column+"/"+value, func(t *testing.T) {
				dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
					setReferenceFixtureCell(f, tc.table, tc.column, value)
				})
				if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil {
					t.Fatal("non-imported parent was accepted")
				}
			})
		}
	}
}

func TestPreflightCutoverBundleRejectsUnpinnedSeeds(t *testing.T) {
	for table, refs := range cutoverCSVReferences() {
		for _, ref := range refs {
			if ref.kind != cutoverClinicSeed && ref.kind != cutoverPlaceholderSeed && ref.kind != cutoverSpeciesSeed {
				continue
			}
			t.Run(table+"/"+ref.column, func(t *testing.T) {
				dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
					setReferenceFixtureCell(f, table, ref.column, "1000084")
				})
				if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil {
					t.Fatal("unpinned numeric seed was accepted")
				}
			})
		}
	}
}

func TestPreflightCutoverBundleAcceptsPinnedSpecies(t *testing.T) {
	for species := 1; species <= 6; species++ {
		t.Run(strconv.Itoa(species), func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				setReferenceFixtureCell(f, "pets", "animal_species_id", strconv.Itoa(species))
			})
			if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
				t.Fatalf("pinned species rejected: %v", err)
			}
		})
	}
}

func TestPreflightCutoverBundleAcceptsForwardSelfReferences(t *testing.T) {
	dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
		for i, spec := range CutoverTableSpecs() {
			if spec.Name != "procedures" && spec.Name != "vaccines" {
				continue
			}
			setReferenceFixtureCell(f, spec.Name, "parent_id", "1000002")
			parent := append([]string(nil), f.rows[spec.Name][0]...)
			parent[columnIndex(spec.Columns, "id")] = "1000002"
			parent[columnIndex(spec.Columns, "parent_id")] = ""
			f.rows[spec.Name] = append(f.rows[spec.Name], parent)
			f.manifest.Tables[i].RowCount++
		}
	})
	if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
		t.Fatalf("forward self reference rejected: %v", err)
	}
}

func TestPreflightCutoverBundleUsesNumericReferenceIdentity(t *testing.T) {
	dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
		for table, refs := range cutoverCSVReferences() {
			for _, ref := range refs {
				if ref.kind == cutoverCSVParent {
					value := "+001000001"
					if ref.parent == "owners" {
						value = "+00300001"
					}
					setReferenceFixtureCell(f, table, ref.column, value)
				}
			}
		}
	})
	if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err != nil {
		t.Fatalf("numeric-equivalent references rejected: %v", err)
	}
}

func TestPreflightCutoverBundleRejectsNumericDuplicateParents(t *testing.T) {
	dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
		parent := append([]string(nil), f.rows["pets"][0]...)
		parent[0] = "+001000001"
		f.rows["pets"] = append(f.rows["pets"], parent)
		f.manifest.Tables[4].RowCount++
	})
	if _, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest)); err == nil || !strings.Contains(err.Error(), "duplicate primary ID") {
		t.Fatalf("numeric duplicate error = %v", err)
	}
}

func TestCutoverReferenceInventoryCoversAllTargetForeignKeys(t *testing.T) {
	references := cutoverCSVReferences()
	if err := validateCutoverReferenceInventory(CutoverTableSpecs(), references); err != nil {
		t.Fatal(err)
	}
	for _, target := range cutoverRequiredForeignKeys() {
		found := false
		for _, ref := range references[target.childTable] {
			if ref.column == target.childColumn && ref.parent == target.parentTable && target.parentColumn == "id" {
				found = true
			}
		}
		if !found {
			t.Errorf("target FK is not classified: %s.%s", target.childTable, target.childColumn)
		}
	}
	for _, target := range cutoverRequiredCompositeForeignKeys() {
		for i, column := range target.childColumns {
			found := false
			for _, ref := range references[target.childTable] {
				if ref.column == column && (column == "clinic_id" && ref.kind == cutoverClinicSeed || ref.parent == target.parentTable && target.parentColumns[i] == "id") {
					found = true
				}
			}
			if !found {
				t.Errorf("target composite FK is not classified: %s.%s", target.childTable, column)
			}
		}
	}
}

func TestCutoverReferenceInventoryRejectsContractDrift(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(map[string][]cutoverCSVReference)
	}{
		{"missing table", func(refs map[string][]cutoverCSVReference) { delete(refs, "pets") }},
		{"renamed table", func(refs map[string][]cutoverCSVReference) { refs["unknown"] = refs["pets"]; delete(refs, "pets") }},
		{"missing column", func(refs map[string][]cutoverCSVReference) { refs["estimates"] = refs["estimates"][:4] }},
		{"duplicate column", func(refs map[string][]cutoverCSVReference) { refs["pets"] = append(refs["pets"], refs["pets"][0]) }},
		{"unknown column", func(refs map[string][]cutoverCSVReference) { refs["pets"][0].column = "unknown" }},
		{"unknown parent", func(refs map[string][]cutoverCSVReference) { refs["pets"][1].parent = "unknown" }},
		{"empty parent", func(refs map[string][]cutoverCSVReference) { refs["pets"][1].parent = "" }},
		{"unknown kind", func(refs map[string][]cutoverCSVReference) { refs["pets"][1].kind = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			refs := cutoverCSVReferences()
			tc.mutate(refs)
			if err := validateCutoverReferenceInventory(CutoverTableSpecs(), refs); err == nil {
				t.Fatal("inventory drift was accepted")
			}
		})
	}
}

func TestCutoverReferenceGraphRehashesOpenedBytes(t *testing.T) {
	dir, digest := writeCutoverFixture(t, nil)
	bundle, err := PreflightCutoverBundle(dir, referenceFixtureSource(digest))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "staffs.csv")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// Change only an unused text field: parent IDs still resolve. The graph
	// must reject the new bytes instead of trusting an earlier path check.
	changed := strings.Replace(string(contents), "1000001,{{CLINIC_ID}},", "1000001,{{CLINIC_ID}},private-value", 1)
	if err := os.WriteFile(path, []byte(changed), 0o600); err != nil {
		t.Fatal(err)
	}
	err = validateCutoverReferenceGraph(dir, bundle.Manifest)
	if err == nil || !strings.Contains(err.Error(), "digest or row count changed") || strings.Contains(err.Error(), "private-value") || strings.Contains(err.Error(), dir) {
		t.Fatalf("changed file error = %v", err)
	}
}

func TestCutoverReferenceIDMatchesSignedIntegerDomain(t *testing.T) {
	for _, value := range []string{"", "0", "-1", " 1000001", "1000001 ", "1.0", "18446744073709551615", "9223372036854775808"} {
		if _, ok := cutoverReferenceID(value); ok {
			t.Errorf("invalid ID spelling accepted: %q", value)
		}
	}
	for _, value := range []string{"1000001", "+1000001", "001000001"} {
		if got, ok := cutoverReferenceID(value); !ok || got != 1000001 {
			t.Errorf("numeric ID = %d, %v", got, ok)
		}
	}
}

func TestCutoverReferenceInventoryIsReturnedIndependently(t *testing.T) {
	first := cutoverCSVReferences()
	second := cutoverCSVReferences()
	first["pets"][0].parent = "changed"
	if reflect.DeepEqual(first, second) || !reflect.DeepEqual(second, cutoverCSVReferences()) {
		t.Fatal("reference inventory shares mutable backing data")
	}
}

func referenceFixtureSource(digest string) ExpectedCutoverSource {
	return ExpectedCutoverSource{ManifestSHA256: digest, ClinicCode: "hachioji", ClinicOrdinal: 1, RunID: "run-1"}
}

func setReferenceFixtureCell(f *fixtureBundle, table, column, value string) {
	for _, spec := range CutoverTableSpecs() {
		if spec.Name == table {
			f.rows[table][0][columnIndex(spec.Columns, column)] = value
			return
		}
	}
}
