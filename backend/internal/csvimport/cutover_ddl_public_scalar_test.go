package csvimport

import (
	"fmt"
	"strings"
	"testing"
)

// TestPreflightCutoverDDLTemporalRequired rejects blank values in the ten required temporal columns under both profiles.
func TestPreflightCutoverDDLTemporalRequired(t *testing.T) {
	cells := []struct{ table, column string }{
		{"appointments", "start_time"},
		{"appointments", "end_time"},
		{"medical_records", "date"},
		{"vital_records", "recorded_at"},
		{"billings", "scheduled_date"},
		{"exams", "date"},
		{"vaccinations", "date"},
		{"estimates", "created_at"},
		{"payments", "created_at"},
		{"payment_splits", "created_at"},
	}
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		for _, cell := range cells {
			for _, value := range []string{"", "  "} {
				t.Run(string(profile)+"/"+cell.table+"."+cell.column, func(t *testing.T) {
					dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
						setFixtureCell(t, f, cell.table, cell.column, value, profile)
					})
					_, err := PreflightCutoverBundle(dir, source)
					coord := fmt.Sprintf("table %s column %s row 2:", cell.table, cell.column)
					if err == nil || !strings.Contains(err.Error(), coord) || !strings.Contains(err.Error(), "must not be empty") {
						t.Fatalf("empty/blank %s.%s accepted under %s: %v", cell.table, cell.column, profile, err)
					}
				})
			}
		}
	}
}

// TestPreflightCutoverDDLOptionalTemporalAllowed proves optional temporal columns accept empty values under both profiles.
func TestPreflightCutoverDDLOptionalTemporalAllowed(t *testing.T) {
	cells := []struct{ table, column string }{
		{"owners", "birth_date"},
		{"pets", "birth_date"},
		{"pets", "neutered_date"},
		{"pets", "deceased_at"},
		{"estimates", "valid_until"},
		{"vaccinations", "next_date"},
	}
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		for _, cell := range cells {
			t.Run(string(profile)+"/"+cell.table+"."+cell.column, func(t *testing.T) {
				dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
					setFixtureCell(t, f, cell.table, cell.column, "", profile)
				})
				if _, err := PreflightCutoverBundle(dir, source); err != nil {
					t.Fatalf("empty optional %s.%s rejected under %s: %v", cell.table, cell.column, profile, err)
				}
			})
		}
	}
}

// TestPreflightCutoverDDLScalarPublicFields exercises scalar validation end-to-end on representative public numeric fields.
func TestPreflightCutoverDDLScalarPublicFields(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		for _, tc := range []struct {
			name   string
			table  string
			column string
			value  string
			valid  bool
		}{
			{"pet-weight-ok", "pets", "weight", "12.34", true},
			{"pet-weight-over", "pets", "weight", "12345.678", false},
			{"owner-discount-ok", "owners", "discount_rate", "0.12", true},
			{"owner-discount-over", "owners", "discount_rate", "999.999", false},
			{"proc-taxrate-ok", "procedures", "tax_rate", "0.10", true},
			{"proc-duration-int", "procedures", "duration", "30", true},
			{"proc-duration-dec", "procedures", "duration", "30.5", false},
			{"vaccine-sort-int", "vaccines", "sort_order", "5", true},
			{"vaccine-sort-junk", "vaccines", "sort_order", "not-a-number-SOURCE_SECRET", false},
			{"merch-taxrate-ok", "merchandise_items", "tax_rate", "0.10", true},
			{"merch-taxrate-junk", "merchandise_items", "tax_rate", "abc", false},
			{"billing-unit-price-neg", "billing_items", "unit_price", "-100", true},
			{"billing-unit-price-dec", "billing_items", "unit_price", "10.5", false},
			{"payment-ratio-ok", "payments", "insurance_ratio", "0.50", true},
			{"estimate-rate-ok", "estimate_items", "discount_rate", "0.50", true},
		} {
			t.Run(string(profile)+"/"+tc.name, func(t *testing.T) {
				dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
					setFixtureCell(t, f, tc.table, tc.column, tc.value, profile)
				})
				_, err := PreflightCutoverBundle(dir, source)
				if tc.valid && err != nil {
					t.Fatalf("%s.%s=%q rejected under %s: %v", tc.table, tc.column, tc.value, profile, err)
				}
				if !tc.valid {
					coord := fmt.Sprintf("table %s column %s row 2:", tc.table, tc.column)
					if err == nil || !strings.Contains(err.Error(), coord) {
						t.Fatalf("%s.%s=%q accepted under %s: %v", tc.table, tc.column, tc.value, profile, err)
					}
					if len(tc.value) >= 4 && strings.Contains(err.Error(), tc.value) {
						t.Fatalf("diagnostic echoes source value: %v", err)
					}
				}
			})
		}
	}
}

// TestPreflightCutoverDDLScalarPublicFieldsExtra covers additional public scalar fields (sort_order, tax_rate, insurance_ratio, unit_price) with valid and invalid literals.
func TestPreflightCutoverDDLScalarPublicFieldsExtra(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		for _, tc := range []struct {
			name   string
			table  string
			column string
			value  string
			valid  bool
		}{
			{"proc-sort-ok", "procedures", "sort_order", "5", true},
			{"proc-sort-junk", "procedures", "sort_order", "xyz", false},
			{"proc-taxrate-junk", "procedures", "tax_rate", "abc", false},
			{"proc-taxrate-big", "procedures", "tax_rate", "1e131072", false},
			{"merch-sort-ok", "merchandise_items", "sort_order", "7", true},
			{"merch-sort-junk", "merchandise_items", "sort_order", "1.5", false},
			{"est-subtotal-dec", "estimates", "subtotal", "1.5", false},
			{"est-subtotal-ok", "estimates", "subtotal", "-50", true},
			{"est-taxtotal-dec", "estimates", "tax_total", "x", false},
			{"est-total-dec", "estimates", "total_amount", "9e999", false},
			{"eitem-taxrate-ok", "estimate_items", "tax_rate", "0.10", true},
			{"eitem-taxrate-over", "estimate_items", "tax_rate", "9.995", false},
			{"eitem-discount-ok", "estimate_items", "discount_rate", "0.50", true},
			{"eitem-discount-over", "estimate_items", "discount_rate", "999.995", false},
			{"pay-ratio-ok", "payments", "insurance_ratio", "0.50", true},
			{"pay-ratio-over", "payments", "insurance_ratio", "9.995", false},
			{"examres-sort-ok", "exam_results", "sort_order", "3", true},
			{"examres-sort-junk", "exam_results", "sort_order", "n/a", false},
			{"bi-price-neg", "billing_items", "unit_price", "-100", true},
			{"bi-price-pos", "billing_items", "unit_price", "2500", true},
			{"bi-price-range", "billing_items", "unit_price", "9223372036854775808", false},
			{"ei-price-neg", "estimate_items", "unit_price", "-200", true},
			{"ei-price-pos", "estimate_items", "unit_price", "700", true},
			{"ei-price-range", "estimate_items", "unit_price", "-9223372036854775809", false},
		} {
			t.Run(string(profile)+"/"+tc.name, func(t *testing.T) {
				dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
					setFixtureCell(t, f, tc.table, tc.column, tc.value, profile)
				})
				_, err := PreflightCutoverBundle(dir, source)
				if tc.valid && err != nil {
					t.Fatalf("%s.%s rejected under %s: %v", tc.table, tc.column, profile, err)
				}
				if !tc.valid {
					coord := fmt.Sprintf("table %s column %s row 2:", tc.table, tc.column)
					if err == nil || !strings.Contains(err.Error(), coord) {
						t.Fatalf("%s.%s accepted under %s: %v", tc.table, tc.column, profile, err)
					}
					if len(tc.value) >= 4 && strings.Contains(err.Error(), tc.value) {
						t.Fatalf("diagnostic echoes source value: %v", err)
					}
				}
			})
		}
	}
}

// TestPreflightCutoverBillingPendingCompletedAtAllowed proves a waiting billing row may leave completed_at empty under both profiles.
func TestPreflightCutoverBillingPendingCompletedAtAllowed(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		t.Run(string(profile), func(t *testing.T) {
			dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
				addFixtureRow(t, f, "billings", "1000002", map[string]string{
					"status":            "waiting",
					"medical_record_id": "",
					"completed_at":      "",
					"total_amount":      "0",
				}, profile)
			})
			if _, err := PreflightCutoverBundle(dir, source); err != nil {
				t.Fatalf("pending billing with empty completed_at rejected under %s: %v", profile, err)
			}
		})
	}
}

// TestPreflightCutoverDDLScalarOptionalDiscountAllowed proves estimate_items.discount_rate may be empty under both profiles.
func TestPreflightCutoverDDLScalarOptionalDiscountAllowed(t *testing.T) {
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		t.Run(string(profile), func(t *testing.T) {
			dir, source := lateReviewFixture(t, profile, func(f *fixtureBundle) {
				setFixtureCell(t, f, "estimate_items", "discount_rate", "", profile)
			})
			if _, err := PreflightCutoverBundle(dir, source); err != nil {
				t.Fatalf("empty optional estimate_items.discount_rate rejected under %s: %v", profile, err)
			}
		})
	}
}
