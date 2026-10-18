package csvimport

import (
	"fmt"
	"strings"
	"testing"
)

func TestPreflightCutoverDDLScalarMapCoverage(t *testing.T) {
	seen := map[string]struct{}{}
	allCols := map[string]map[string]struct{}{}
	for _, profile := range []CutoverProfile{CutoverProfileLegacy21, CutoverProfileMedicinesCandidate} {
		for _, spec := range CutoverTableSpecs(profile) {
			cols, ok := allCols[spec.Name]
			if !ok {
				cols = map[string]struct{}{}
				allCols[spec.Name] = cols
			}
			for _, c := range spec.Columns {
				cols[c] = struct{}{}
			}
		}
	}
	for table, rules := range cutoverDDLScalarColumns {
		cols, ok := allCols[table]
		if !ok {
			t.Fatalf("scalar map references unknown table %s", table)
		}
		for _, rule := range rules {
			key := table + "." + rule.column
			if _, dup := seen[key]; dup {
				t.Fatalf("duplicate scalar rule %s", key)
			}
			seen[key] = struct{}{}
			if _, ok := cols[rule.column]; !ok {
				t.Fatalf("scalar rule %s references missing column", key)
			}
		}
	}
	if len(seen) != 137 {
		t.Fatalf("scalar map covers %d columns, want 137", len(seen))
	}
	if len(cutoverDDLScalarColumns) != 22 {
		t.Fatalf("scalar map covers %d tables, want 22", len(cutoverDDLScalarColumns))
	}
	legacy := map[string]struct{}{}
	for _, spec := range CutoverTableSpecs(CutoverProfileLegacy21) {
		legacy[spec.Name] = struct{}{}
	}
	candidate := map[string]struct{}{}
	for _, spec := range CutoverTableSpecs(CutoverProfileMedicinesCandidate) {
		candidate[spec.Name] = struct{}{}
	}
	for table := range cutoverDDLScalarColumns {
		if _, ok := candidate[table]; !ok {
			t.Fatalf("candidate specs missing scalar table %s", table)
		}
		if table != "medicines" {
			if _, ok := legacy[table]; !ok {
				t.Fatalf("legacy specs missing scalar table %s", table)
			}
		} else if _, ok := legacy[table]; ok {
			t.Fatal("legacy specs unexpectedly include medicines")
		}
	}
}

func TestPreflightCutoverDDLScalarIntegers(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rule  cutoverDDLScalarRule
		value string
		valid bool
	}{
		{"int32-max", cutoverDDLScalarRule{kind: cutoverDDLInt32}, "2147483647", true},
		{"int32-min", cutoverDDLScalarRule{kind: cutoverDDLInt32}, "-2147483648", true},
		{"int32-signed-pad", cutoverDDLScalarRule{kind: cutoverDDLInt32}, "+00042", true},
		{"int32-space", cutoverDDLScalarRule{kind: cutoverDDLInt32}, " 1 ", true},
		{"int32-over", cutoverDDLScalarRule{kind: cutoverDDLInt32}, "2147483648", false},
		{"int32-under", cutoverDDLScalarRule{kind: cutoverDDLInt32}, "-2147483649", false},
		{"int32-decimal", cutoverDDLScalarRule{kind: cutoverDDLInt32}, "1.0", false},
		{"int32-exp", cutoverDDLScalarRule{kind: cutoverDDLInt32}, "1e0", false},
		{"int32-nan", cutoverDDLScalarRule{kind: cutoverDDLInt32}, "NaN", false},
		{"int32-inf", cutoverDDLScalarRule{kind: cutoverDDLInt32}, "Infinity", false},
		{"int32-blank-nn", cutoverDDLScalarRule{kind: cutoverDDLInt32}, "", false},
		{"int32-blank-null", cutoverDDLScalarRule{kind: cutoverDDLInt32, nullable: true}, "", true},
		{"int32-signed-neg", cutoverDDLScalarRule{kind: cutoverDDLInt32}, "-42", true},
		{"int64-max", cutoverDDLScalarRule{kind: cutoverDDLInt64}, "9223372036854775807", true},
		{"int64-min", cutoverDDLScalarRule{kind: cutoverDDLInt64}, "-9223372036854775808", true},
		{"int64-over", cutoverDDLScalarRule{kind: cutoverDDLInt64}, "9223372036854775808", false},
		{"int64-under", cutoverDDLScalarRule{kind: cutoverDDLInt64}, "-9223372036854775809", false},
		{"int64-neg-price", cutoverDDLScalarRule{kind: cutoverDDLInt64}, "-500", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCutoverDDLScalarValue("billing_items", tc.rule, tc.value, 2)
			if tc.valid && err != nil {
				t.Fatalf("rejected: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("accepted invalid")
			}
			if err != nil && strings.Contains(err.Error(), tc.value) && len(tc.value) >= 2 {
				t.Fatalf("diagnostic echoes value: %v", err)
			}
		})
	}
}

func TestPreflightCutoverDDLScalarNumericBounds(t *testing.T) {
	numeric := func(p, s int) cutoverDDLScalarRule {
		return cutoverDDLScalarRule{kind: cutoverDDLNumeric, precision: p, scale: s}
	}
	for _, tc := range []struct {
		name  string
		rule  cutoverDDLScalarRule
		value string
		valid bool
	}{
		{"p5s2-pos-max", numeric(5, 2), "999.994", true},
		{"p5s2-neg-max", numeric(5, 2), "-999.994", true},
		{"p5s2-pos-over", numeric(5, 2), "999.995", false},
		{"p5s2-neg-over", numeric(5, 2), "-999.995", false},
		{"p3s2-max", numeric(3, 2), "9.994", true},
		{"p3s2-over", numeric(3, 2), "9.995", false},
		{"p6s2-max", numeric(6, 2), "9999.994", true},
		{"p6s2-over", numeric(6, 2), "9999.995", false},
		{"p10s1-max", numeric(10, 1), "999999999.94", true},
		{"p10s1-over", numeric(10, 1), "999999999.95", false},
		{"p10s2-max", numeric(10, 2), "99999999.994", true},
		{"p10s2-over", numeric(10, 2), "99999999.995", false},
		{"p5s2-exp-max", numeric(5, 2), "9.99994e2", true},
		{"p5s2-exp-over", numeric(5, 2), "9.99995e2", false},
		{"p-nan", numeric(5, 2), "NaN", true},
		{"p-pos-inf", numeric(5, 2), "Infinity", false},
		{"p-neg-inf", numeric(5, 2), "-Infinity", false},
		{"u-exp-max", numeric(0, 0), "1e131071", true},
		{"u-exp-over", numeric(0, 0), "1e131072", false},
		{"u-exp-neg-over", numeric(0, 0), "-1e131072", false},
		{"u-frac-max", numeric(0, 0), "1e-16383", true},
		{"u-frac-over", numeric(0, 0), "1e-16384", false},
		{"u-zero-frac-max", numeric(0, 0), "0e-16383", true},
		{"u-zero-frac-over", numeric(0, 0), "0e-16384", false},
		{"u-zero-exp", numeric(0, 0), "0e131072", true},
		{"u-pos-inf", numeric(0, 0), "Infinity", true},
		{"u-neg-inf", numeric(0, 0), "-Infinity", true},
		{"u-nan", numeric(0, 0), "NaN", true},
		{"u-junk", numeric(0, 0), "not-a-number-SOURCE_SECRET", false},
		{"u-ws-only", numeric(0, 0), "   ", false},
		{"u-unicode-ws", numeric(0, 0), " 1 ", false},
		{"u-trailing-zeros", numeric(0, 0), "1." + strings.Repeat("0", 16384), false},
		{"u-trailing-zeros-ok", numeric(0, 0), "1." + strings.Repeat("0", 16383), true},
		{"u-blank-null", cutoverDDLScalarRule{kind: cutoverDDLNumeric, nullable: true}, "", true},
		{"u-blank-nn", numeric(0, 0), "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCutoverDDLScalarValue("estimate_items", tc.rule, tc.value, 2)
			if tc.valid && err != nil {
				t.Fatalf("rejected: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("accepted invalid")
			}
			if !tc.valid && err != nil && len(tc.value) >= 2 && len(tc.value) <= 64 && strings.Contains(err.Error(), tc.value) {
				t.Fatalf("diagnostic echoes value: %v", err)
			}
		})
	}
}

func TestPreflightCutoverDDLScalarRulesReject(t *testing.T) {
	for table, rules := range cutoverDDLScalarColumns {
		for _, rule := range rules {
			t.Run(table+"."+rule.column, func(t *testing.T) {
				err := validateCutoverDDLScalarValue(table, rule, "not-a-number-SOURCE_SECRET", 2)
				coord := fmt.Sprintf("table %s column %s row 2:", table, rule.column)
				if err == nil || !strings.Contains(err.Error(), coord) {
					t.Fatalf("invalid sentinel accepted: %v", err)
				}
				if err != nil && strings.Contains(err.Error(), "SOURCE_SECRET") {
					t.Fatalf("diagnostic echoes source value: %v", err)
				}
				emptyErr := validateCutoverDDLScalarValue(table, rule, "", 2)
				if rule.nullable && emptyErr != nil {
					t.Fatalf("nullable empty rejected: %v", emptyErr)
				}
				if !rule.nullable && (emptyErr == nil || !strings.Contains(emptyErr.Error(), coord)) {
					t.Fatalf("NN empty accepted: %v", emptyErr)
				}
			})
		}
	}
}

func TestPreflightCutoverDDLScalarInteger64(t *testing.T) {
	rule := cutoverDDLScalarRule{column: "id", kind: cutoverDDLInt64}
	for _, tc := range []struct {
		name  string
		value string
		valid bool
	}{
		{"max", "9223372036854775807", true},
		{"min", "-9223372036854775808", true},
		{"signpad", "+00042", true},
		{"space", " 1 ", true},
		{"over", "9223372036854775808", false},
		{"under", "-9223372036854775809", false},
		{"decimal", "1.0", false},
		{"exponent", "1e0", false},
		{"nan", "NaN", false},
		{"inf", "Infinity", false},
		{"blank", "", false},
		{"junk", "not-a-number", false},
		{"unicode-ws", " 1 ", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCutoverDDLScalarValue("staffs", rule, tc.value, 2)
			if tc.valid && err != nil {
				t.Fatalf("rejected: %v", err)
			}
			if !tc.valid {
				if err == nil || !strings.Contains(err.Error(), "table staffs column id row 2:") {
					t.Fatalf("accepted invalid: %v", err)
				}
				if len(tc.value) >= 4 && strings.Contains(err.Error(), tc.value) {
					t.Fatalf("diagnostic echoes value: %v", err)
				}
			}
		})
	}
}

func TestPreflightCutoverDDLScalarNumericSymmetry(t *testing.T) {
	numeric := func(p, s int) cutoverDDLScalarRule {
		return cutoverDDLScalarRule{column: "tax_rate", kind: cutoverDDLNumeric, precision: p, scale: s}
	}
	for _, tc := range []struct {
		name  string
		rule  cutoverDDLScalarRule
		value string
		valid bool
	}{
		{"p3s2-neg-max", numeric(3, 2), "-9.994", true},
		{"p3s2-neg-over", numeric(3, 2), "-9.995", false},
		{"p6s2-neg-max", numeric(6, 2), "-9999.994", true},
		{"p6s2-neg-over", numeric(6, 2), "-9999.995", false},
		{"p10s1-neg-max", numeric(10, 1), "-999999999.94", true},
		{"p10s1-neg-over", numeric(10, 1), "-999999999.95", false},
		{"p10s2-neg-max", numeric(10, 2), "-99999999.994", true},
		{"p10s2-neg-over", numeric(10, 2), "-99999999.995", false},
		{"p5s2-leadzero-ok", numeric(5, 2), "000999.994", true},
		{"p5s2-leadzero-over", numeric(5, 2), "000999.995", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCutoverDDLScalarValue("estimate_items", tc.rule, tc.value, 2)
			if tc.valid && err != nil {
				t.Fatalf("rejected: %v", err)
			}
			if !tc.valid && (err == nil || !strings.Contains(err.Error(), "table estimate_items column tax_rate row 2:")) {
				t.Fatalf("accepted invalid: %v", err)
			}
		})
	}
}

func TestPreflightCutoverDDLScalarPlaceholders(t *testing.T) {
	id := func(col string) cutoverDDLScalarRule {
		return cutoverDDLScalarRule{column: col, kind: cutoverDDLInt64}
	}
	for _, tc := range []struct {
		name  string
		table string
		rule  cutoverDDLScalarRule
		value string
		valid bool
	}{
		{"clinic-token", "staffs", id("clinic_id"), "{{CLINIC_ID}}", true},
		{"reservation-token", "appointments", id("reservation_type_id"), "{{TRIMMING_RESERVATION_TYPE_ID}}", true},
		{"species-token", "pets", id("animal_species_id"), "{{FALLBACK_ANIMAL_SPECIES_ID}}", true},
		{"paymethod-token", "payments", id("payment_method_id"), "{{PAYMENT_METHOD_CASH_ID}}", true},
		{"examtype-token", "exams", id("exam_type_id"), "{{FALLBACK_EXAM_TYPE_ID}}", true},
		{"forged-token", "staffs", id("clinic_id"), "{{FORGED_TOKEN}}", false},
		{"misplaced-token", "owners", cutoverDDLScalarRule{column: "discount_rate", kind: cutoverDDLNumeric, precision: 5, scale: 2}, "{{CLINIC_ID}}", false},
		{"misplaced-token2", "procedures", cutoverDDLScalarRule{column: "duration", kind: cutoverDDLInt32, nullable: true}, "{{CLINIC_ID}}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCutoverDDLScalarValue(tc.table, tc.rule, tc.value, 2)
			if tc.valid && err != nil {
				t.Fatalf("legitimate token rejected: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("forged/misplaced token accepted")
			}
		})
	}
}

func TestPreflightCutoverDDLScalarUnknownKindFailsClosed(t *testing.T) {
	err := validateCutoverDDLScalarValue("owners", cutoverDDLScalarRule{column: "discount_rate", kind: cutoverDDLScalarKind(255)}, "123", 2)
	if err == nil || !strings.Contains(err.Error(), "table owners column discount_rate row 2:") {
		t.Fatalf("unknown scalar kind accepted: %v", err)
	}
}
