package csvimport

import (
	"strings"
	"testing"
)

// P2 fix 1 (EMR-342 DESIGN 3.2): the producer classifies blank source names
// as needs_review and removes them, so a medicines row with a blank or
// whitespace-only name must fail consumer preflight. Names are never
// normalized and rejected cell values never appear in errors.
func TestCutoverMedicinesBlankNameRejected(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	nameCol := columnIndex(spec.Columns, "name")
	for _, tc := range []struct{ label, value string }{
		{"empty", ""},
		{"ascii-space", " "},
		{"tab", "\t"},
		{"ideographic-space", "　"},
	} {
		t.Run(tc.label, func(t *testing.T) {
			dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
				f.rows["medicines"][0][nameCol] = tc.value
			})
			_, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
			if err == nil {
				t.Fatal("PreflightCutoverBundle accepted a blank medicines name")
			}
			if !strings.Contains(err.Error(), "name is required") {
				t.Fatalf("PreflightCutoverBundle error = %v, want blank name rejection", err)
			}
			// A single ASCII space appears in any prose message, so the
			// no-leak guard only applies to the multi-byte payload.
			if tc.value == "　" && strings.Contains(err.Error(), tc.value) {
				t.Fatal("error leaks the rejected cell value")
			}
		})
	}
}

func TestCutoverMedicinesNameNotNormalized(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	nameCol := columnIndex(spec.Columns, "name")
	dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
		f.rows["medicines"][0][nameCol] = "  padded  name  "
	})
	if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err != nil {
		t.Fatalf("PreflightCutoverBundle error = %v, want name accepted without normalization", err)
	}
}

// P2 fix 2 (EMR-342 DESIGN 3.4): a candidate22 medicine rewrite estimate row
// carries only medicine_id; its category must be medicine and the
// consultation/procedure/merchandise references must all be empty. A medicine
// category also requires medicine_id. Non-medicine rows are unchanged.
func TestCutoverEstimateMedicineRewriteJointInvariant(t *testing.T) {
	estimateSpec := cutoverSpecByName(t, "estimate_items")
	categoryCol := columnIndex(estimateSpec.Columns, "category")
	medicineCol := columnIndex(estimateSpec.Columns, "medicine_id")
	consultCol := columnIndex(estimateSpec.Columns, "consultation_id")
	procedureCol := columnIndex(estimateSpec.Columns, "procedure_id")
	merchandiseCol := columnIndex(estimateSpec.Columns, "merchandise_item_id")
	// Distinct existing master IDs avoid the legacy cross-master overlap guard.
	medicineID := "1000002"

	set := func(f *fixtureBundle, category, medicineIDValue, consultation, procedure, merchandise string) {
		f.rows["medicines"][0][columnIndex(cutoverSpecByName(t, "medicines").Columns, "id")] = medicineID
		f.rows["merchandise_items"][0][columnIndex(cutoverSpecByName(t, "merchandise_items").Columns, "id")] = "1000003"
		row := f.rows["estimate_items"][0]
		row[categoryCol] = category
		row[medicineCol] = medicineIDValue
		row[consultCol] = consultation
		row[procedureCol] = procedure
		row[merchandiseCol] = merchandise
	}

	t.Run("valid medicine rewrite accepted", func(t *testing.T) {
		dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
			set(f, "medicine", medicineID, "", "", "")
		})
		if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err != nil {
			t.Fatalf("PreflightCutoverBundle error = %v, want contract-consistent medicine rewrite accepted", err)
		}
	})

	rejections := []struct {
		name     string
		mutate   func(*fixtureBundle)
		wantText string
		leaked   string
	}{
		{
			name: "populated medicine id with non-medicine category",
			mutate: func(f *fixtureBundle) {
				set(f, "procedure", medicineID, "", "", "")
			},
			wantText: "category medicine",
			leaked:   medicineID,
		},
		{
			name: "medicine rewrite with procedure ref",
			mutate: func(f *fixtureBundle) {
				set(f, "medicine", medicineID, "", "1000001", "")
			},
			wantText: "procedure_id",
		},
		{
			name: "medicine rewrite with merchandise ref",
			mutate: func(f *fixtureBundle) {
				set(f, "medicine", medicineID, "", "", "1000003")
			},
			wantText: "merchandise_item_id",
		},
		{
			name: "medicine rewrite with consultation ref",
			mutate: func(f *fixtureBundle) {
				set(f, "medicine", medicineID, "1000001", "", "")
			},
			wantText: "consultation_id",
		},
		{
			name: "medicine category without medicine id",
			mutate: func(f *fixtureBundle) {
				set(f, "medicine", "", "", "", "")
			},
			wantText: "medicine_id",
		},
	}
	for _, tc := range rejections {
		t.Run(tc.name, func(t *testing.T) {
			dir, digest := writeMedicinesCandidateFixture(t, tc.mutate)
			_, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
			if err == nil {
				t.Fatal("PreflightCutoverBundle accepted a contract-violating medicine rewrite row")
			}
			if !strings.Contains(err.Error(), tc.wantText) {
				t.Fatalf("PreflightCutoverBundle error = %v, want %q", err, tc.wantText)
			}
			if tc.leaked != "" && strings.Contains(err.Error(), tc.leaked) {
				t.Fatal("error leaks the rejected cell value")
			}
		})
	}

	t.Run("valid non-medicine row accepted", func(t *testing.T) {
		dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
			set(f, "procedure", "", "", "1000001", "")
		})
		if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err != nil {
			t.Fatalf("PreflightCutoverBundle error = %v, want non-medicine estimate row accepted", err)
		}
	})
}

// Consultations are never CSV parents, so the consultation_id leg of the
// joint invariant cannot be exercised through a fixture bundle. Mirror the
// medicines lexical regression tests and call the row validator directly.
func TestCutoverEstimateMedicineRewriteRowLevelChecks(t *testing.T) {
	spec := cutoverSpecByName(t, "estimate_items")
	header := spec.Columns
	indexes := make(map[string]int, len(header))
	for i, column := range header {
		indexes[column] = i
	}
	row := make([]string, len(header))
	row[indexes["id"]] = "1000001"
	row[indexes["estimate_id"]] = "1000001"
	row[indexes["category"]] = "medicine"
	row[indexes["medicine_id"]] = "1000001"
	row[indexes["consultation_id"]] = "1000001"
	band := CutoverIDBand{Base: 0, EndExclusive: 10_000_000, NonOwnerIDOffset: 1_000_000, OwnerFloor: 300_000, ApplicationIDFloor: 1_000_000_000}
	if err := validateCutoverRow(spec, header, row, indexes, band, 2, CutoverProfileMedicinesCandidate); err == nil || !strings.Contains(err.Error(), "consultation_id") {
		t.Fatalf("validateCutoverRow error = %v, want consultation dual-reference rejection", err)
	}
}
