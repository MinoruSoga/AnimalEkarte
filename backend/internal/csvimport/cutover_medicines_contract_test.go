package csvimport

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// RED coverage for EMR-343: the medicines master table joins the cutover
// contract as table 22, between procedures and merchandise_items, and
// estimate_items.medicine_id becomes a CSV-parent reference.

func TestCutoverMedicinesTableSpecSchemaAndOrder(t *testing.T) {
	specs := CutoverTableSpecs()
	if len(specs) != 22 {
		t.Fatalf("table count = %d, want 22", len(specs))
	}
	if specs[1].Name != "procedures" || specs[2].Name != "medicines" || specs[3].Name != "merchandise_items" {
		t.Fatalf("master order = %s, %s, %s; want procedures, medicines, merchandise_items",
			specs[1].Name, specs[2].Name, specs[3].Name)
	}
	spec := cutoverSpecByName(t, "medicines")
	wantColumns := []string{
		"id", "clinic_id", "name", "price", "is_active", "description",
		"parent_id", "dosage_form", "medicine_unit", "inventory_id",
		"default_quantity", "tax_type", "tax_rate", "sort_order", "is_non_insurance",
	}
	if !slices.Equal(spec.Columns, wantColumns) {
		t.Fatalf("medicines columns = %v, want %v", spec.Columns, wantColumns)
	}
	if !slices.Equal(spec.BandColumns, []string{"id", "parent_id"}) {
		t.Fatalf("medicines band columns = %v, want [id parent_id]", spec.BandColumns)
	}
	if !slices.Equal(spec.ForceNotNullColumns, []string{"name", "description"}) {
		t.Fatalf("medicines force-not-null columns = %v, want [name description]", spec.ForceNotNullColumns)
	}
}

func TestCutoverMedicinesPlaceholderInventory(t *testing.T) {
	placeholders := CutoverPlaceholderColumns()
	if got := placeholders["medicines.clinic_id"]; got != "{{CLINIC_ID}}" {
		t.Fatalf("placeholder for medicines.clinic_id = %q, want {{CLINIC_ID}}", got)
	}
}

func TestCutoverMedicinesReferenceClassification(t *testing.T) {
	references := cutoverCSVReferences()
	want := []cutoverCSVReference{
		{"clinic_id", "clinics", false, cutoverClinicSeed},
		{"parent_id", "medicines", true, cutoverCSVParent},
		{"inventory_id", "inventory_items", true, cutoverUnavailableParent},
	}
	if got := references["medicines"]; !reflect.DeepEqual(got, want) {
		t.Fatalf("medicines references = %+v, want %+v", got, want)
	}
	found := false
	for _, ref := range references["estimate_items"] {
		if ref.column == "medicine_id" {
			found = true
			if ref.kind != cutoverCSVParent || ref.parent != "medicines" || !ref.nullable {
				t.Fatalf("estimate_items.medicine_id = %+v, want nullable CSV parent medicines", ref)
			}
		}
	}
	if !found {
		t.Fatal("estimate_items.medicine_id missing from reference inventory")
	}
}

func TestCutoverMedicinesRequiredForeignKeys(t *testing.T) {
	required := map[[2]string][2]string{
		{"medicines", "clinic_id"}: {"clinics", "id"},
		{"medicines", "parent_id"}: {"medicines", "id"},
	}
	for _, fk := range cutoverRequiredForeignKeys() {
		key := [2]string{fk.childTable, fk.childColumn}
		if want, ok := required[key]; ok && want == [2]string{fk.parentTable, fk.parentColumn} {
			delete(required, key)
		}
	}
	for key := range required {
		t.Errorf("cutoverRequiredForeignKeys is missing %s.%s", key[0], key[1])
	}
}

func TestCutoverMedicinesCopyContract(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	copySQL := cutoverCopySQL(spec, pgx.Identifier{spec.Name})
	if !strings.HasPrefix(copySQL, `COPY "medicines" (`) {
		t.Fatalf("direct COPY SQL = %s", copySQL)
	}
	if !strings.Contains(copySQL, `FORCE_NOT_NULL ("name", "description")`) {
		t.Fatalf("COPY does not preserve required empty text: %s", copySQL)
	}
	// parent_id is a self reference: apply must stage and insert by depth
	// because the FK is not deferrable, same as procedures and vaccines.
	column, ok := cutoverSelfReferenceColumn("medicines")
	if !ok || column != "parent_id" {
		t.Fatalf("self reference column = %q, %v; want parent_id", column, ok)
	}
}

func TestCutoverMedicinesDefaultQuantityAndDosageForm(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	dosageFormIdx := columnIndex(spec.Columns, "dosage_form")
	defaultQuantityIdx := columnIndex(spec.Columns, "default_quantity")

	for _, tc := range []struct {
		name            string
		dosageForm      string
		defaultQuantity string
	}{
		{"empty", "", ""},
		{"tablet", "tablet", "1"},
		{"liquid", "liquid", "0.5"},
		{"injection", "injection", "1"},
		{"topical", "topical", "1"},
		{"powder", "powder", "2.5"},
	} {
		t.Run("accepts "+tc.name, func(t *testing.T) {
			dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
				f.rows["medicines"][0][dosageFormIdx] = tc.dosageForm
				f.rows["medicines"][0][defaultQuantityIdx] = tc.defaultQuantity
			})
			if _, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest)); err != nil {
				t.Fatalf("PreflightCutoverBundle error = %v, want dosage_form=%q default_quantity=%q accepted",
					err, tc.dosageForm, tc.defaultQuantity)
			}
		})
	}

	t.Run("rejects unknown dosage form", func(t *testing.T) {
		dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
			f.rows["medicines"][0][dosageFormIdx] = "bogus-form"
		})
		_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
		if err == nil || !strings.Contains(err.Error(), "dosage_form") {
			t.Fatalf("PreflightCutoverBundle error = %v, want dosage_form rejection", err)
		}
		if strings.Contains(err.Error(), "bogus-form") {
			t.Fatal("error leaks the source cell value")
		}
	})

	t.Run("rejects non-numeric default quantity", func(t *testing.T) {
		dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
			f.rows["medicines"][0][defaultQuantityIdx] = "not-a-number"
		})
		_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
		if err == nil || !strings.Contains(err.Error(), "default_quantity") {
			t.Fatalf("PreflightCutoverBundle error = %v, want default_quantity rejection", err)
		}
		if strings.Contains(err.Error(), "not-a-number") {
			t.Fatal("error leaks the source cell value")
		}
	})
}

// CSV-655-R1: medicine_unit is the target DB enum (per_tablet, per_ml,
// per_dose, per_gram). A nonempty unknown label must fail source preflight
// before any database write; empty stays NULL.
func TestCutoverMedicinesUnitEnum(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	unitIdx := columnIndex(spec.Columns, "medicine_unit")

	for _, value := range []string{"", "per_tablet", "per_ml", "per_dose", "per_gram"} {
		name := value
		if name == "" {
			name = "empty stays NULL"
		}
		t.Run("accepts "+name, func(t *testing.T) {
			dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
				f.rows["medicines"][0][unitIdx] = value
			})
			if _, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest)); err != nil {
				t.Fatalf("PreflightCutoverBundle error = %v, want medicine_unit=%q accepted", err, value)
			}
		})
	}

	t.Run("rejects unknown unit in the full rehearsal bundle", func(t *testing.T) {
		dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
			f.rows["medicines"][0][unitIdx] = "per_pill"
		})
		_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
		if err == nil || !strings.Contains(err.Error(), "table medicines") || !strings.Contains(err.Error(), "column medicine_unit") || !strings.Contains(err.Error(), "row 2") {
			t.Fatalf("PreflightCutoverBundle error = %v, want medicine_unit rejection naming table, column, row", err)
		}
		if strings.Contains(err.Error(), "per_pill") {
			t.Fatal("error leaks the source cell value")
		}
	})
}

// A legacy 21-table producer bundle (no medicines entry, no medicines.csv)
// must fail closed — including under local-rehearsal drift tolerance, where
// the manifest tables count check runs before drift is considered.
func TestCutoverMedicinesLegacy21BundleRejected(t *testing.T) {
	dropMedicines := func(f *fixtureBundle) {
		kept := f.manifest.Tables[:0]
		for _, table := range f.manifest.Tables {
			if table.Table != "medicines" {
				kept = append(kept, table)
			}
		}
		f.manifest.Tables = kept
		delete(f.rows, "medicines")
	}
	t.Run("formal shape rejected", func(t *testing.T) {
		dir, digest := writeCutoverFixture(t, dropMedicines)
		_, err := PreflightCutoverBundle(dir, formalExpectedSource(digest))
		if err == nil || !strings.Contains(err.Error(), "want 22") {
			t.Fatalf("PreflightCutoverBundle error = %v, want 22-table count rejection", err)
		}
	})
	t.Run("local rehearsal", func(t *testing.T) {
		dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
			mutateRehearsalOnlyHandoff(f)
			dropMedicines(f)
		})
		_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
		if err == nil || !strings.Contains(err.Error(), "want 22") {
			t.Fatalf("PreflightCutoverBundle error = %v, want 22-table count rejection even in rehearsal", err)
		}
	})
}

func TestCutoverMedicinesMissingCSV(t *testing.T) {
	dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
		f.afterManifest = func(dir string) {
			_ = os.Remove(filepath.Join(dir, "medicines.csv"))
		}
	})
	_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
	if err == nil || !strings.Contains(err.Error(), "medicines") {
		t.Fatalf("PreflightCutoverBundle error = %v, want missing medicines.csv rejection", err)
	}
}

func TestCutoverMedicinesEstimateItemReference(t *testing.T) {
	estimateItemsSpec := cutoverSpecByName(t, "estimate_items")
	medicinesSpec := cutoverSpecByName(t, "medicines")
	medicineCol := columnIndex(estimateItemsSpec.Columns, "medicine_id")
	medicinesIDCol := columnIndex(medicinesSpec.Columns, "id")

	t.Run("resolves a medicines CSV parent", func(t *testing.T) {
		dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
			f.rows["estimate_items"][0][medicineCol] = f.rows["medicines"][0][medicinesIDCol]
		})
		if _, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest)); err != nil {
			t.Fatalf("PreflightCutoverBundle error = %v, want medicines parent accepted", err)
		}
	})

	t.Run("rejects a missing medicines parent", func(t *testing.T) {
		dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
			f.rows["estimate_items"][0][medicineCol] = "1000002"
		})
		_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
		if err == nil || !strings.Contains(err.Error(), "reference to medicines") {
			t.Fatalf("PreflightCutoverBundle error = %v, want missing medicines parent rejection", err)
		}
		if strings.Contains(err.Error(), "1000002") {
			t.Fatal("error leaks the source cell value")
		}
	})

	t.Run("empty medicine_id stays nullable", func(t *testing.T) {
		dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
			f.rows["estimate_items"][0][medicineCol] = ""
		})
		if _, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest)); err != nil {
			t.Fatalf("PreflightCutoverBundle error = %v, want empty medicine_id accepted", err)
		}
	})
}

// inventory_id is a non-imported parent: the medicines inventory link must
// stay empty, exactly like vaccines.inventory_id.
func TestCutoverMedicinesInventoryMustBeEmpty(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	inventoryCol := columnIndex(spec.Columns, "inventory_id")
	dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
		f.rows["medicines"][0][inventoryCol] = "1000001"
	})
	_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
	if err == nil || !strings.Contains(err.Error(), "inventory") {
		t.Fatalf("PreflightCutoverBundle error = %v, want non-empty inventory_id rejection", err)
	}
	if strings.Contains(err.Error(), "1000001") {
		t.Fatal("error leaks the source cell value")
	}
}

func TestCutoverMedicinesClinicMismatch(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	clinicCol := columnIndex(spec.Columns, "clinic_id")
	dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
		f.rows["medicines"][0][clinicCol] = "1000001"
	})
	_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
	if err == nil || !strings.Contains(err.Error(), "clinic_id") {
		t.Fatalf("PreflightCutoverBundle error = %v, want clinic placeholder rejection", err)
	}
	if strings.Contains(err.Error(), "1000001") {
		t.Fatal("error leaks the source cell value")
	}
}

// V7: medicine names must be unique inside the clinic bundle, like the
// producer-side procedure name guard.
func TestCutoverMedicinesDuplicateNames(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	idCol := columnIndex(spec.Columns, "id")
	nameCol := columnIndex(spec.Columns, "name")
	dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
		base := f.rows["medicines"][0]
		base[nameCol] = "duplicated-name"
		second := append([]string(nil), base...)
		second[idCol] = "1000002"
		f.rows["medicines"] = append(f.rows["medicines"], second)
		for i := range f.manifest.Tables {
			if f.manifest.Tables[i].Table == "medicines" {
				f.manifest.Tables[i].RowCount++
			}
		}
	})
	_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
	if err == nil || !strings.Contains(err.Error(), "duplicate") || !strings.Contains(err.Error(), "name") {
		t.Fatalf("PreflightCutoverBundle error = %v, want duplicate medicines name rejection", err)
	}
	if strings.Contains(err.Error(), "duplicated-name") {
		t.Fatal("error leaks the source cell value")
	}
}
