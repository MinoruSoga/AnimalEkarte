package csvimport

import (
	"bytes"
	"context"
	"encoding/csv"
	"strconv"
	"strings"
	"testing"
)

// RED coverage for EMR-345: negative prices are rejected for exactly the
// four master tables. Only nullable master prices allow empty; zero stays allowed. Historical negative
// values in billing_items and estimate_items (refunds/discounts) stay
// allowed. Errors name table, column, and row — never the cell value.

var cutoverMasterPriceTargets = []struct {
	table  string
	column string
}{
	{"procedures", "price"},
	{"medicines", "price"},
	{"merchandise_items", "unit_price"},
	{"vaccines", "price"},
}

func TestCutoverMasterPriceRejectsNegative(t *testing.T) {
	for _, tc := range cutoverMasterPriceTargets {
		t.Run(tc.table+"."+tc.column, func(t *testing.T) {
			spec := cutoverSpecByName(t, tc.table)
			column := columnIndex(spec.Columns, tc.column)
			dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
				f.rows[tc.table][0][column] = "-100"
			})
			_, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
			if err == nil {
				t.Fatal("negative master price was accepted")
			}
			for _, want := range []string{"table " + tc.table, "column " + tc.column, "row 2", "negative"} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error = %q, want detail %q", err.Error(), want)
				}
			}
			if strings.Contains(err.Error(), "-100") {
				t.Fatal("error leaks the source cell value")
			}
		})
	}
}

func TestCutoverMasterPriceAcceptsZeroEmptyAndPositive(t *testing.T) {
	for _, tc := range cutoverMasterPriceTargets {
		for _, value := range []string{"0", "", "500"} {
			if tc.table == "merchandise_items" && value == "" {
				continue
			}
			name := tc.table + "." + tc.column + "/" + value
			if value == "" {
				name += "empty"
			}
			t.Run(name, func(t *testing.T) {
				spec := cutoverSpecByName(t, tc.table)
				column := columnIndex(spec.Columns, tc.column)
				dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
					f.rows[tc.table][0][column] = value
				})
				if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err != nil {
					t.Fatalf("PreflightCutoverBundle error = %v, want %s=%q accepted", err, tc.column, value)
				}
			})
		}
	}
}

func TestCutoverMasterPriceRejectsMalformed(t *testing.T) {
	for _, tc := range cutoverMasterPriceTargets {
		t.Run(tc.table+"."+tc.column, func(t *testing.T) {
			spec := cutoverSpecByName(t, tc.table)
			column := columnIndex(spec.Columns, tc.column)
			dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
				f.rows[tc.table][0][column] = "not-an-integer"
			})
			_, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
			if err == nil || !strings.Contains(err.Error(), "must be an integer") {
				t.Fatalf("PreflightCutoverBundle error = %v, want malformed integer rejection", err)
			}
			if strings.Contains(err.Error(), "not-an-integer") {
				t.Fatal("error leaks the source cell value")
			}
		})
	}
}

// The COPY transform is anonymous: it substitutes only the declared seed
// placeholders and emits every other cell verbatim. This guard proves an
// empty price stays empty (no DB-default claim), an explicit 0 stays 0 (no
// truncation or zero-fill), and default_quantity raw decimal bytes pass
// through unchanged.
func TestCutoverMasterPriceCopyTransformPreservesCells(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	priceIdx := columnIndex(spec.Columns, "price")
	quantityIdx := columnIndex(spec.Columns, "default_quantity")
	clinicIdx := columnIndex(spec.Columns, "clinic_id")
	makeRow := func(id, price, quantity string) []string {
		row := make([]string, len(spec.Columns))
		for i, column := range spec.Columns {
			switch column {
			case "id":
				row[i] = id
			case "clinic_id":
				row[i] = "{{CLINIC_ID}}"
			case "price":
				row[i] = price
			case "default_quantity":
				row[i] = quantity
			}
		}
		return row
	}
	path, table := writeDriftTestCSV(t, spec, spec.Columns, [][]string{
		makeRow("1000001", "0", "1.0"),
		makeRow("1000002", "", "0.5"),
	})
	var out bytes.Buffer
	seeds := validCutoverSeeds()
	records, err := transformCutoverCSV(context.Background(), path, spec, false, &out, seeds, table.SHA256)
	if err != nil {
		t.Fatalf("transformCutoverCSV error = %v", err)
	}
	if records != 2 {
		t.Fatalf("transformed records = %d, want 2", records)
	}
	rows, err := csv.NewReader(&out).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if got := rows[1][priceIdx]; got != "0" {
		t.Fatalf("explicit zero price transformed to %q, want verbatim 0", got)
	}
	if got := rows[2][priceIdx]; got != "" {
		t.Fatalf("empty price transformed to %q, want empty", got)
	}
	if got := rows[2][quantityIdx]; got != "0.5" {
		t.Fatalf("default_quantity transformed to %q, want verbatim 0.5", got)
	}
	wantClinic := strconv.FormatInt(seeds.ClinicID, 10)
	if got := rows[1][clinicIdx]; got != wantClinic {
		t.Fatalf("clinic placeholder transformed to %q, want bound seed %q", got, wantClinic)
	}
}

func TestCutoverMasterPriceAllowsHistoricalNegativeBillingAndEstimate(t *testing.T) {
	for _, tc := range []struct {
		table  string
		column string
	}{
		{"billing_items", "unit_price"},
		{"estimate_items", "unit_price"},
	} {
		t.Run(tc.table+"."+tc.column, func(t *testing.T) {
			spec := cutoverSpecByName(t, tc.table)
			column := columnIndex(spec.Columns, tc.column)
			dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
				f.rows[tc.table][0][column] = "-100"
			})
			if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err != nil {
				t.Fatalf("PreflightCutoverBundle error = %v, want historical negative %s accepted", err, tc.table)
			}
		})
	}
}
