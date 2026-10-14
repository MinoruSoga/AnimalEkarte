package csvimport

import (
	"strings"
	"testing"
)

// EMR-342 DESIGN 3.2 maps 15 medicines fields. Guard inventory audited as a
// whole, not piecemeal: id + clinic_id (band/clinic seed), name (unique and
// non-blank), price (non-negative int64 source, bigint target),
// is_active/is_non_insurance (pinned t-true/f-false), description
// (force-not-null), dosage_form (enum), medicine_unit (empty), inventory_id
// (unavailable parent), sort_order (source integer), and the three fixed-mapping
// contracts exercised here: default_quantity fixed at 1 (numeric
// equivalence), parent_id unavailable (no legacy source), and the tax pair
// excluded/exempt + 0.10 under the current producer contract.

// 4264495015: producer default_quantity is the fixed value 1.0. Storage-
// equivalent decimal spellings are accepted without binary floating point;
// zero and any other representable value fail the public bundle contract.
// The lexical numeric(10,2) helper keeps its own semantics and tests.
func TestCutoverMedicinesDefaultQuantityFixedToOne(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	quantityCol := columnIndex(spec.Columns, "default_quantity")

	for _, accepted := range []string{"1", "1.0", "1.00", "1.0000", "+01.00"} {
		t.Run("accepts "+accepted, func(t *testing.T) {
			dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
				f.rows["medicines"][0][quantityCol] = accepted
			})
			if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err != nil {
				t.Fatalf("PreflightCutoverBundle error = %v, want %q accepted as one", err, accepted)
			}
		})
	}

	for _, rejected := range []string{"0", "1.23", "0.5", "2.5", "-1", "1e0"} {
		t.Run("rejects "+rejected, func(t *testing.T) {
			dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
				f.rows["medicines"][0][quantityCol] = rejected
			})
			_, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
			if err == nil || !strings.Contains(err.Error(), "default_quantity") {
				t.Fatalf("PreflightCutoverBundle error = %v, want fixed default_quantity rejection", err)
			}
			if len(rejected) > 2 && strings.Contains(err.Error(), rejected) {
				t.Fatal("error leaks the rejected cell value")
			}
		})
	}
}

// 4264495027: the tax pair follows the live producer contract — MIG-38 maps
// blank/code 1 to excluded and 2/3 to exempt, and MIG-34 reuses the MIG-27
// formula so both producer states carry tax_rate 0.10 (030_stage.sql CONSTANT,
// asserted by test-procedure-master-pricing.mjs). The consumer accepts the
// two producer states only; included, unknown, empty, and any rate not
// exactly 0.10 fail closed. The 0.10 comparison is exact decimal digit
// inspection, never a binary float.
func TestCutoverMedicinesTaxPairPinned(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	taxTypeCol := columnIndex(spec.Columns, "tax_type")
	taxRateCol := columnIndex(spec.Columns, "tax_rate")

	for _, acceptedRate := range []string{"0.1", "0.10", "0.1000"} {
		t.Run("accepts excluded "+acceptedRate, func(t *testing.T) {
			dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
				f.rows["medicines"][0][taxTypeCol] = "excluded"
				f.rows["medicines"][0][taxRateCol] = acceptedRate
			})
			if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err != nil {
				t.Fatalf("PreflightCutoverBundle error = %v, want excluded/%q accepted", err, acceptedRate)
			}
		})
		t.Run("accepts exempt "+acceptedRate, func(t *testing.T) {
			dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
				f.rows["medicines"][0][taxTypeCol] = "exempt"
				f.rows["medicines"][0][taxRateCol] = acceptedRate
			})
			if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err != nil {
				t.Fatalf("PreflightCutoverBundle error = %v, want exempt/%q accepted", err, acceptedRate)
			}
		})
	}

	rejections := []struct{ label, taxType, taxRate string }{
		{"exempt with zero rate", "exempt", "0"},
		{"included with 0.10", "included", "0.10"},
		{"excluded with 0.08", "excluded", "0.08"},
		{"empty tax type", "", "0.10"},
		{"empty tax rate", "excluded", ""},
		{"whole rate", "excluded", "1"},
		{"negative rate", "excluded", "-0.10"},
		{"exponent rate", "excluded", "1e-1"},
		{"zero rate", "excluded", "0"},
	}
	for _, tc := range rejections {
		t.Run("rejects "+tc.label, func(t *testing.T) {
			dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
				f.rows["medicines"][0][taxTypeCol] = tc.taxType
				f.rows["medicines"][0][taxRateCol] = tc.taxRate
			})
			_, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
			if err == nil || (!strings.Contains(err.Error(), "tax_type") && !strings.Contains(err.Error(), "tax_rate")) {
				t.Fatalf("PreflightCutoverBundle error = %v, want pinned tax pair rejection", err)
			}
			// The pinned-value hint contains "0.10", so negative/exponent
			// spellings embedding it are exempt from the no-leak guard.
			if tc.taxType != "" && strings.Contains(err.Error(), tc.taxType) {
				t.Fatal("error leaks the rejected tax_type cell value")
			}
			if len(tc.taxRate) > 1 && !strings.Contains(tc.taxRate, "0.10") && strings.Contains(err.Error(), tc.taxRate) {
				t.Fatal("error leaks the rejected tax_rate cell value")
			}
		})
	}
}

// 4264495021: the old database has no medicines parent source, so
// medicines.parent_id is unavailable: empty stays NULL and a valid band ID
// is rejected at source preflight. The target FK/self-reference plumbing is
// preserved for the migration/application schema.
func TestCutoverMedicinesParentIDUnavailable(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	parentCol := columnIndex(spec.Columns, "parent_id")

	t.Run("empty parent_id accepted", func(t *testing.T) {
		dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
			f.rows["medicines"][0][parentCol] = ""
		})
		if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err != nil {
			t.Fatalf("PreflightCutoverBundle error = %v, want empty parent_id accepted", err)
		}
	})

	t.Run("own parent id rejected", func(t *testing.T) {
		dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
			f.rows["medicines"][0][parentCol] = "1000002"
		})
		_, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
		if err == nil || !strings.Contains(err.Error(), "parent_id") {
			t.Fatalf("PreflightCutoverBundle error = %v, want parent_id unavailable rejection", err)
		}
		if strings.Contains(err.Error(), "1000002") {
			t.Fatal("error leaks the source cell value")
		}
	})

	// The classification gap the unavailable-parent contract closes: today a
	// resolvable non-self parent is still accepted even though the legacy
	// database has no medicines parent source.
	t.Run("valid different parent rejected", func(t *testing.T) {
		dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
			base := f.rows["medicines"][0]
			second := append([]string(nil), base...)
			second[columnIndex(spec.Columns, "id")] = "1000004"
			second[columnIndex(spec.Columns, "name")] = "second-medicine"
			second[parentCol] = "1000002"
			f.rows["medicines"] = append(f.rows["medicines"], second)
			for i := range f.manifest.Tables {
				if f.manifest.Tables[i].Table == "medicines" {
					f.manifest.Tables[i].RowCount++
				}
			}
		})
		_, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
		if err == nil || !strings.Contains(err.Error(), "parent_id") {
			t.Fatalf("PreflightCutoverBundle error = %v, want parent_id unavailable rejection", err)
		}
		if strings.Contains(err.Error(), "1000002") {
			t.Fatal("error leaks the source cell value")
		}
	})
}
